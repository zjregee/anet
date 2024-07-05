package anet

type LoadBalance interface {
	Pick() Ring
	Rebalance(rings []Ring)
}

type Ring interface {
	Id() string
	Wait() error
	Submit(operator *FDOperator, event RingEvent, eventData interface{}) error
	Alloc() *FDOperator
	Free(operator *FDOperator)
	Register(operator *FDOperator)
	Close() error
}

type PrepReadEventData struct {
	Size int
	Data []byte
}

type PrepWriteEventData struct {
	Size int
	Data []byte
}

type RingEvent int

const (
	RingPrepRead  RingEvent = 0x1
	RingPRepWrite RingEvent = 0x2
)
