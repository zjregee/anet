package pool

import (
	"math"
	"runtime"
	"time"
)

const (
	DefaultPoolSize          = math.MaxInt32
	DefaultCleanIntervalTime = time.Second
)

const (
	OPENED = iota
	CLOSED
)

var (
	defaultPool, _  = newGoPool(DefaultPoolSize)
	workerChanCapacity = func() int {
		if runtime.GOMAXPROCS(0) == 1 {
			return 0
		}
		return 1
	}()
)

func Submit(task func()) error {
	return defaultPool.Submit(task)
}

func Running() int {
	return defaultPool.Running()
}

func Free() int {
	return defaultPool.Free()
}

func Capacity() int {
	return defaultPool.Capacity()
}

func Release() error {
	return defaultPool.Release()
}

func ReleaseTimeout(timeout time.Duration) error {
	return defaultPool.ReleaseTimeout(timeout)
}

func Reboot() {
	defaultPool.Reboot()
}

func NewPool(size int, options ...Option) (Pool, error) {
	return newGoPool(size, options...)
}

func NewMultiPool(size, sizePerPool int, options ...Option) (Pool, error) {
	return newMultiGoPool(size, sizePerPool, options...)
}

type Pool interface {
	Capacity() int
	Waiting() int
	Running() int
	Free() int
	IsClosed() bool
	Submit(task func()) error
	Release() error
	ReleaseTimeout(timeout time.Duration) error
	Reboot()
}

type worker interface {
	run()
	finish()
	inputFunc(func())
	lastUsedTime() time.Time
	updateLastUsedTime(time time.Time)
}

type workerQueue interface {
	len() int
	isEmpty() bool
	insert(worker)
	detach() worker
	refresh(duration time.Duration) []worker
	reset()
}
