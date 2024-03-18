package anet

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

func newOperatorCache() *operatorCache {
	return &operatorCache{
		cache: make([]*FDOperator, 0, 1024),
		freelist: make([]int32, 0, 1024),
	}
}

type operatorCache struct {
	locked     int32
	first      *FDOperator
	cache      []*FDOperator
	freelist   []int32
	freelocked int32
}

func (c *operatorCache) alloc() *FDOperator {
	lock(&c.locked)
	if c.first == nil {
		const opSize = unsafe.Sizeof(FDOperator{})
		n := 4096 / opSize
		if n == 0 {
			n = 1
		}
		index := int32(len(c.cache))
		for i := uintptr(0); i < n; i++ {
			pd := &FDOperator{index: index}
			c.cache = append(c.cache, pd)
			pd.next = c.first
			c.first = pd
			index++
		}
	}
	op := c.first
	c.first = op.next
	unlock(&c.locked)
	return op
}

func (c *operatorCache) freeable(op *FDOperator) {
	lock(&c.freelocked)
	c.freelist = append(c.freelist, op.index)
	unlock(&c.freelocked)
}

func (c *operatorCache) free() {
	lock(&c.freelocked)
	if len(c.freelist) == 0 {
		unlock(&c.freelocked)
		return
	}
	lock(&c.locked)
	for _, index := range c.freelist {
		op := c.cache[index]
		op.next = c.first
		c.first = op
	}
	c.freelist = c.freelist[:0]
	unlock(&c.locked)
	unlock(&c.freelocked)
}

func lock(locked *int32) {
	for !atomic.CompareAndSwapInt32(locked, 0, 1) {
		runtime.Gosched()
	}
}

func unlock(locked *int32) {
	atomic.StoreInt32(locked, 0)
}
