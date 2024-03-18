package pool

import "time"

type goWorker struct {
	pool     *goPool
	task     chan func()
	lastUsed time.Time
}

func (w *goWorker) run() {
	w.pool.addRunning(1)
	go func() {
		defer func() {
			w.pool.addRunning(-1)
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				}
			}
			w.pool.cond.Signal()
		}()
		for f := range w.task {
			if f == nil {
				return
			}
			f()
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}

func (w *goWorker) finish() {
	w.task <- nil
}

func (w *goWorker) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorker) updateLastUsedTime(time time.Time) {
	w.lastUsed = time
}

func (w *goWorker) inputFunc(fn func()) {
	w.task <- fn
}
