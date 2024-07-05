package anet

type LoadBalance interface {
	Pick() Ring
	Rebalance(rings []Ring)
}

type Ring interface {
	Id() string
	Wait() error
	Submit(eventData RingEventData)
	Alloc() *FDOperator
	Free(operator *FDOperator)
	Register(operator *FDOperator)
	Close() error
}

type RingEvent int

const (
	RingPrepRead  RingEvent = 0x1
	RingPrepWrite RingEvent = 0x2
)

type RingEventData struct {
	Size     int
	Data     []byte
	Event    RingEvent
	Operator *FDOperator
}
