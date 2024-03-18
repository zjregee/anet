package pool

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"sync/atomic"
)

func newMultiGoPool(size, sizePerPool int, options ...Option) (*multiGoPool, error) {
	pools := make([]*goPool, size)
	for i := 0; i < size; i++ {
		pool, err := newGoPool(sizePerPool, options...)
		if err != nil {
			return nil, err
		}
		pools[i] = pool
	}
	return &multiGoPool{pools: pools}, nil
}

type multiGoPool struct {
	pools []*goPool
	index uint32
	state int32
}

func (mp *multiGoPool) Capacity() int {
	var capacity int
	for _, pool := range mp.pools {
		capacity += pool.Capacity()
	}
	return capacity
}

func (mp *multiGoPool) Waiting() int {
	var waiting int
	for _, pool := range mp.pools {
		waiting += pool.Waiting()
	}
	return waiting
}

func (mp *multiGoPool) Running() int {
	var running int
	for _, pool := range mp.pools {
		running += pool.Running()
	}
	return running
}

func (mp *multiGoPool) Free() int {
	var free int
	for _, pool := range mp.pools {
		free += pool.Free()
	}
	return free
}

func (mp *multiGoPool) IsClosed() bool {
	return atomic.LoadInt32(&mp.state) == CLOSED
}

func (mp *multiGoPool) Submit(task func()) error {
	if mp.IsClosed() {
		return errors.New("this pool hash been closed")
	}
	return mp.pools[mp.next()].Submit(task)
}

func (mp *multiGoPool) Release() error {
	if !atomic.CompareAndSwapInt32(&mp.state, OPENED, CLOSED) {
		return errors.New("this pool has been closed")
	}
	var errStr strings.Builder
	for i, pool := range mp.pools {
		if err := pool.Release(); err != nil {
			errStr.WriteString(fmt.Sprintf("pool %d: %v\n", i, err))
			if i < len(mp.pools)-1 {
				errStr.WriteString(" | ")
			}
		}
	}
	if errStr.Len() == 0 {
		return nil
	}
	return errors.New(errStr.String())
}

func (mp *multiGoPool) ReleaseTimeout(timeout time.Duration) error {
	if !atomic.CompareAndSwapInt32(&mp.state, OPENED, CLOSED) {
		return errors.New("this pool has been closed")
	}
	var errStr strings.Builder
	for i, pool := range mp.pools {
		if err := pool.ReleaseTimeout(timeout); err != nil {
			errStr.WriteString(fmt.Sprintf("pool %d: %v\n", i, err))
			if i < len(mp.pools)-1 {
				errStr.WriteString(" | ")
			}
		}
	}
	if errStr.Len() == 0 {
		return nil
	}
	return errors.New(errStr.String())
}

func (mp *multiGoPool) Reboot() {
	if atomic.CompareAndSwapInt32(&mp.state, CLOSED, OPENED) {
		atomic.StoreUint32(&mp.index, 0)
		for _, pool := range mp.pools {
			pool.Reboot()
		}
	}
}

func (mp *multiGoPool) next() int {
	idx := int((atomic.AddUint32(&mp.index, 1) - 1) % uint32(len(mp.pools)))
	if idx == -1 {
		idx = 0
	}
	return idx
}
