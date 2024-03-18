package anet

import "sync/atomic"

type FDOperator struct {
	FD        int
	OnAccept  func() error
	Inputs    func(vs [][]byte) (rs [][]byte)
	InputAck  func(n int) (err error)
	Outputs   func(vs [][]byte) (rs [][]byte)
	OutputAck func(n int) (err error)
	ring      Ring
	detached  int32
	next      *FDOperator
	index     int32
}

func (op *FDOperator) control(event RingEvent) error {
	if event == RingDetach && atomic.AddInt32(&op.detached, 1) > 1 {
		return nil
	}
	return op.ring.Control(op, event, nil)
}

func (op *FDOperator) free() {
	op.ring.Free(op)
}

func (op *FDOperator) reset() {
	op.FD = 0
	op.OnAccept = nil
	op.Inputs, op.InputAck = nil, nil
	op.Outputs, op.OutputAck = nil, nil
	op.ring = nil
	op.detached = 0
}
