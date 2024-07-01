package anet

import (
	"runtime"
	"sync/atomic"
)

const (
	connecting int32 = iota
	processing
	flushing
	closing
)

const (
	none int32 = iota
	user
	ring
)

type locker struct {
	keychain [4]int32
}

func (c *connection) isClosedBy(who int32) bool {
	return atomic.LoadInt32(&c.keychain[closing]) == who
}

func (c *connection) lock(key int32) bool {
	result := atomic.CompareAndSwapInt32(&c.keychain[key], 0, 1)
	switch key {
	case connecting:
		if result {
			log.Tracef("[connection %s] try to lock connecting status successed", c.id)
		} else {
			log.Tracef("[connection %s] try to lock connecting status failed", c.id)
		}
	case processing:
		if result {
			log.Tracef("[connection %s] try to lock processing status successed", c.id)
		} else {
			log.Tracef("[connection %s] try to lock processing status failed", c.id)
		}
	case flushing:
		if result {
			log.Tracef("[connection %s] try to lock flushing status successed", c.id)
		} else {
			log.Tracef("[connection %s] try to lock flushing status failed", c.id)
		}
	default:
		panic("unreachable code")
	}
	return result
}

func (c *connection) unlock(key int32) {
	switch key {
	case connecting:
		log.Tracef("[connection %s] unlock connecting status", c.id)
	case processing:
		log.Tracef("[connection %s] unlock processing status", c.id)
	case flushing:
		log.Tracef("[connection %s] unlock flushing status", c.id)
	default:
		panic("unreachable code")
	}
	atomic.StoreInt32(&c.keychain[key], 0)
}

func (c *connection) force(key int32, value int32) {
	atomic.StoreInt32(&c.keychain[key], value)
}

func (c *connection) waitUnlock(key int32) {
	switch key {
	case connecting:
		log.Tracef("[connection %s] waitUnlock wait for connecting status unlock", c.id)
	case processing:
		log.Tracef("[connection %s] waitUnlock wait for processing status unlock", c.id)
	case flushing:
		log.Tracef("[connection %s] waitUnlock wait for flushing status unlock", c.id)
	default:
		panic("unreachable code")
	}
	for !atomic.CompareAndSwapInt32(&c.keychain[key], 0, 2) && atomic.LoadInt32(&c.keychain[key]) != 2 {
		runtime.Gosched()
	}
	log.Tracef("[connection %s] waitUnlock return", c.id)
}

func (c *connection) closedBy(who int32) bool {
	result := atomic.CompareAndSwapInt32(&c.keychain[closing], none, who)
	switch who {
	case user:
		if result {
			log.Tracef("[connection %s] closeBy user successed", c.id)
		} else {
			log.Tracef("[connection %s] closeBy user failed", c.id)
		}
	case ring:
		if result {
			log.Tracef("[connection %s] closeBy ring successed", c.id)
		} else {
			log.Tracef("[connection %s] closeBy ring failed", c.id)
		}
	default:
		panic("unreachable code")
	}
	return result
}
