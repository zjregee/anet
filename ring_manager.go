package anet

import (
	"errors"
	"runtime"
)

func setNumLoops(numLoops int) error {
	return ringmanager.setNumLoops(numLoops)
}

var ringmanager *manager

func init() {
	loops := runtime.GOMAXPROCS(0) / 20 + 1
	ringmanager = &manager{}
	ringmanager.setNumLoops(loops)
}

type manager struct {
	numLoops int
	balance  loadbalance
	rings    []Ring
}

func (m *manager) setNumLoops(numLoops int) error {
	if numLoops < 1 {
		return errors.New("set invalid numLoops")
	}
	if numLoops < m.numLoops {
		var err error
		rings := make([]Ring, numLoops)
		for index := 0; index < m.numLoops; index++ {
			if index < numLoops {
				rings[index] = m.rings[index]
			} else {
				err = m.rings[index].Close()
			}
		}
		m.numLoops = numLoops
		m.rings = rings
		m.balance.rebalance(m.rings)
		return err
	}
	m.numLoops = numLoops
	return m.run()
}

func (m *manager) pick() Ring {
	return m.balance.pick()
}

func (m *manager) run() error {
	for index := len(m.rings); index < m.numLoops; index++ {
		ring, err := openDefaultPoll()
		if err != nil {
			_ = m.close()
			return err
		}
		m.rings = append(m.rings, ring)
		go ring.Wait()
	}
	m.balance.rebalance(m.rings)
	return nil
}

func (m *manager) reset() error {
	for _, ring := range m.rings {
		_ = ring.Close()
	}
	m.rings = nil
	return m.run()
}

func (m *manager) close() error {
	var err error
	for _, ring := range m.rings {
		err = ring.Close()
	}
	m.numLoops = 0
	m.balance = nil
	m.rings = nil
	return err
}
