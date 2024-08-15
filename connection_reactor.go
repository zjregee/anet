package anet

const (
	defaultReadSize = 1024
)

func (c *connection) onRead(n int, err error) {
	_ = c.inputBuffer.BookAck(n)
	c.readTrigger <- err
}

func (c *connection) onWrite(n int, err error) {
	_ = c.outputBuffer.SeekAck(n)
	c.writeTrigger <- err
}

func (c *connection) submitRead() {
	eventData := RingEventData{}
	eventData.Size = defaultReadSize
	eventData.Data = c.inputBuffer.Book(defaultReadSize)
	eventData.Event = RingPrepRead
	c.operator.Submit(eventData)
}

func (c *connection) submitWrite() {
	size := c.outputBuffer.Len()
	eventData := RingEventData{}
	eventData.Size = size
	eventData.Data, _ = c.outputBuffer.Seek(size)
	eventData.Event = RingPrepWrite
	c.operator.Submit(eventData)
}
