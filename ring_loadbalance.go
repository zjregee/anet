package anet

import "sync/atomic"

type loadbalance interface {
	pick() (ring Ring)
	rebalance(rings []Ring)
}

func newLoadbalance(rings []Ring) loadbalance {
	return &roundRobinLB{rings: rings, ringSize: len(rings)}
}

type roundRobinLB struct {
	rings    []Ring
	accepted uintptr
	ringSize int
}

func (b *roundRobinLB) pick() (ring Ring) {
	index := int(atomic.AddUintptr(&b.accepted, 1)) % b.ringSize
	return b.rings[index]
}

func (b *roundRobinLB) rebalance(rings []Ring) {
	b.rings, b.ringSize = rings, len(rings)
}
