package anet

import "sync/atomic"

const (
	processing int32 = iota
	flushing
	closing
)

const (
	none int32 = iota
	user
	ring
)

type locker struct {
	keychain [3]int32
}

func (l *locker) lock(k int32) bool {
	return atomic.CompareAndSwapInt32(&l.keychain[k], 0, 1)
}

func (l *locker) unlock(k int32) {
	atomic.StoreInt32(&l.keychain[k], 0)
}

func (l *locker) status(k int32) int32 {
	return atomic.LoadInt32(&l.keychain[k])
}

func (l *locker) closeBy(w int32) bool {
	return atomic.CompareAndSwapInt32(&l.keychain[closing], none, w)
}

func (l *locker) isCloseBy(w int32) bool {
	return atomic.LoadInt32(&l.keychain[closing]) == w
}
