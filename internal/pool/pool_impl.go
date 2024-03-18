package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
	syncx "anet/internal/sync"
)

const (
	nowTimeUpdateInterval = 500 * time.Millisecond
)

func newGoPool(size int, options ...Option) (*goPool, error) {
	if size <= 0 {
		size = -1
	}
	opts := loadOptions(options...)
	if !opts.DisablePurge {
		if expiry := opts.ExpiryDuration; expiry < 0 {
			return nil, errors.New("invalid expiry for pool")
		} else if expiry == 0 {
			opts.ExpiryDuration = DefaultCleanIntervalTime
		}
	}
	p := &goPool{
		capacity: int32(size),
		lock: syncx.NewSpinLock(),
		options: opts,
	}
	p.workerCache.New = func() interface{} {
		return &goWorker{
			pool: p,
			task: make(chan func(), workerChanCapacity),
		}
	}
	p.workers = newWorkerStack(0)
	p.cond = sync.NewCond(p.lock)
	p.goPurge()
	p.goTicktock()
	return p, nil
}

type goPool struct {
	capacity     int32
	running      int32
	waiting      int32
	cond         *sync.Cond
	lock         sync.Locker
	workers      workerQueue
	workerCache  sync.Pool
	state        int32
	now          atomic.Value
	options      *Options
	purgeDone    int32
	stopPurge    context.CancelFunc
	ticktockDone int32
	stopTicktock context.CancelFunc
}

func (p *goPool) Capacity() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *goPool) Waiting() int {
	return int(atomic.LoadInt32(&p.waiting))
}

func (p *goPool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *goPool) Free() int {
	c := p.Capacity()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

func (p *goPool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == CLOSED
}

func (p *goPool) Submit(task func()) error {
	if p.IsClosed() {
		return errors.New("this pool hash been closed")
	}
	w, err := p.retrieveWorker()
	if w != nil {
		w.inputFunc(task)
	}
	return err
}

func (p *goPool) Release() error {
	if !atomic.CompareAndSwapInt32(&p.state, OPENED, CLOSED) {
		return errors.New("this pool has been closed")
	}
	if p.stopPurge != nil {
		p.stopPurge()
		p.stopPurge = nil
	}
	p.stopTicktock()
	p.stopTicktock = nil
	p.lock.Lock()
	p.workers.reset()
	p.lock.Unlock()
	p.cond.Broadcast()
	return nil
}

func (p *goPool) ReleaseTimeout(timeout time.Duration) error {
	if err := p.Release(); err != nil {
		return err
	}
	endTime := time.Now().Add(timeout)
	for time.Now().Before(endTime) {
		if p.Running() == 0 {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return errors.New("operation timed out")
}

func (p *goPool) Reboot() {
	if atomic.CompareAndSwapInt32(&p.state, CLOSED, OPENED) {
		atomic.StoreInt32(&p.purgeDone, 0)
		go p.goPurge()
		atomic.StoreInt32(&p.ticktockDone, 0)
		go p.goTicktock()
	}
}

func (p *goPool) nowTime() time.Time {
	return p.now.Load().(time.Time)
}

func (p *goPool) addRunning(delta int) {
	atomic.AddInt32(&p.running, int32(delta))
}

func (p *goPool) addWaiting(delta int) {
	atomic.AddInt32(&p.waiting, int32(delta))
}

func (p *goPool) goTicktock() {
	p.now.Store(time.Now())
	var ctx context.Context
	ctx, p.stopTicktock = context.WithCancel(context.Background())
	go p.ticktock(ctx)
}

func (p *goPool) goPurge() {
	if p.options.DisablePurge {
		return
	}
	var ctx context.Context
	ctx, p.stopPurge = context.WithCancel(context.Background())
	go p.purgeStaleWorkers(ctx)
}

func (p *goPool) ticktock(ctx context.Context) {
	ticker := time.NewTicker(nowTimeUpdateInterval)
	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.ticktockDone, 1)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if p.IsClosed() {
			break
		}
		p.now.Store(time.Now())
	}
}

func (p *goPool) purgeStaleWorkers(ctx context.Context) {
	ticker := time.NewTicker(p.options.ExpiryDuration)
	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.purgeDone, 1)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		p.lock.Lock()
		if p.IsClosed() {
			p.lock.Unlock()
			break
		}
		staleWorkers := p.workers.refresh(p.options.ExpiryDuration)
		p.lock.Unlock()
		for i := range staleWorkers {
			staleWorkers[i].finish()
			staleWorkers[i] = nil
		}
	}
}

func (p *goPool) retrieveWorker() (worker, error) {
	p.lock.Lock()
	if p.IsClosed() {
		return nil, errors.New("this pool hash been closed")
	}
	for {
		if w := p.workers.detach(); w != nil {
			p.lock.Unlock()
			return w, nil
		}
		if capacity := p.Capacity(); capacity == - 1 || capacity > p.Running() {
			p.lock.Unlock()
			w := p.workerCache.Get().(*goWorker)
			w.run()
			return w, nil
		}
		if p.options.NonBlocking || (p.options.MaxBlockingTasks != 0 && p.Waiting() >= p.options.MaxBlockingTasks) {
			p.lock.Unlock();
			return nil, errors.New("too many groutines blocked on submit or Nonblocking is set")
		}
		p.addWaiting(1)
		p.cond.Wait()
		p.addWaiting(-1)
		if p.IsClosed() {
			return nil, errors.New("this pool hash been closed")
		}
	}
}

func (p *goPool) revertWorker(worker *goWorker) bool {
	if capacity := p.Capacity(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		return false
	}
	worker.updateLastUsedTime(p.nowTime())
	p.lock.Lock()
	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}
	p.workers.insert(worker)
	p.cond.Signal()
	p.lock.Unlock()
	return true
}
