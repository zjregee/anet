package anet

import (
	"errors"
	"sync/atomic"
)

func newDefaultRingManager(n int) *manager {
	ringmanager := &manager{}
	ringmanager.numLoops = n
	ringmanager.balance = &roundRobinLB{}
	err := ringmanager.Run()
	if err != nil {
		panic("should't failed here")
	}
	return ringmanager
}

type roundRobinLB struct {
	rings      []Ring
	lastpicked int32
}

func (b *roundRobinLB) Pick() (ring Ring) {
	if len(b.rings) == 0 {
		return nil
	}
	index := int(atomic.AddInt32(&b.lastpicked, 1)) % len(b.rings)
	return b.rings[index]
}

func (b *roundRobinLB) Rebalance(rings []Ring) {
	b.rings = rings
	b.lastpicked = 0
}

var RingManager *manager

type manager struct {
	numLoops int
	rings    []Ring
	balance  LoadBalance
}

func (m *manager) Pick() Ring {
	return m.balance.Pick()
}

func (m *manager) Run() error {
	var errs []error
	var rings []Ring
	for index := 0; index < m.numLoops; index++ {
		ring, err := newDefaultRing()
		if err != nil {
			errs = append(errs, err)
			log.Warnf("error occurred while open ring")
		} else {
			go func() {
				_ = ring.Wait()
			}()
			rings = append(rings, ring)
		}
	}
	m.rings = rings
	m.balance.Rebalance(m.rings)
	err := errors.Join(errs...)
	return err
}
