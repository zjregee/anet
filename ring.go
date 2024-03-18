package anet

type Ring interface {
	Wait() error
	Close() error
	Control(operator *FDOperator, event RingEvent, eventData interface{}) error
	Alloc() (operator *FDOperator)
	Free(operator *FDOperator)
}

type RingEvent int

const (
	RingPrepAccept  RingEvent = 0x1
	RingPrepConnect RingEvent = 0x2
	RingPrepRead    RingEvent = 0x3
	RingPrepWrite   RingEvent = 0x4
	RingDetach      RingEvent = 0x5
)
