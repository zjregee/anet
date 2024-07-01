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
	Close() error
}

type PrepReadEventData struct {
	size int
	data []byte
}

type PrepWriteEventData struct {
	size int
	data []byte
}

type RingEvent int

const (
	RingPrepRead  RingEvent = 0x1
	RingPRepWrite RingEvent = 0x2
)
