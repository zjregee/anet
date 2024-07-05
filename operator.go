package anet

type FDOperator struct {
	FD      int
	OnRead  func(n int, err error)
	OnWrite func(n int, err error)
	Ring    Ring
}

func (op *FDOperator) Submit(eventData RingEventData) {
	eventData.Operator = op
	op.Ring.Submit(eventData)
}

func (op *FDOperator) Register() {
	op.Ring.Register(op)
}

func (op *FDOperator) Free() {
	op.Ring.Free(op)
}

func (op *FDOperator) Reset() {
	op.FD = 0
	op.OnRead = nil
	op.OnWrite = nil
	op.Ring = nil	
}
