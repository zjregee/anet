package anet

type FDOperator struct {
	FD      int
	OnRead  func(n int) error
	OnWrite func(n int) error
	ring    Ring
}

func (op *FDOperator) submit(evnet RingEvent, eventData interface{}) error {
	return op.ring.Submit(op, evnet, eventData)
}

func (op *FDOperator) free() {
	op.ring.Free(op)
}

func (op *FDOperator) reset() {
	op.FD = 0
	op.OnRead = nil
	op.OnWrite = nil
	op.ring = nil	
}
