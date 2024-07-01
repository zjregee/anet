package anet

import (
	"sync/atomic"
)

const (
	defaultReadSize = 1024
)

func (c *connection) onRead(n int) error {
	log.Infof("[conn %s] has read %d bytes of new data", c.id, n)
	err := c.inputBuffer.BookAck(n)
	if err != nil {
		return err
	}
	processed := c.onRequest()
	if !processed && n >= int(atomic.LoadInt32(&c.waitReadSize)) {
		// there is a goroutine waiting for enough data
		c.triggerRead(nil)
	}
	c.triggerReadLoop(nil)
	return nil
}

func (c *connection) onWrite(n int) error {
	log.Infof("[conn %s] has write %d bytes of new data", c.id, n)
	err := c.outputBuffer.SeekAck(n)
	if err != nil {
		return err
	}
	c.unlock(flushing)
	c.triggerWrite(nil)
	c.triggerWriteLoop(nil)
	return nil
}

func (c *connection) onHup() error {
	log.Infof("[conn %s] waiting for closing since peer close", c.id)
	c.force(flushing, 2)
	if c.closedBy(ring) {
		log.Infof("[conn %s] already closed", c.id)
		return nil
	}
	err := Exception(ErrConnClosed, "peer close")
	c.triggerRead(err)
	c.triggerWrite(err)
	c.triggerReadLoop(err)
	c.triggerWriteLoop(err)
	log.Infof("[conn %s] waiting for the close callback to execute", c.id)
	err = c.closeCallback()
	if err != nil {
		log.Infof("[conn %s] closed with error: %s", c.id, err.Error())
	} else {
		log.Infof("[conn %s] closed", c.id)
	}
	return err
}

func (c *connection) onClose() error {
	log.Infof("[conn %s] waiting for closing since self close", c.id)
	if c.closedBy(user) {
		log.Infof("[conn %s] already closed", c.id)
		return nil
	}
	// waiting flush finished first before closed
	c.waitUnlock(flushing)
	err := Exception(ErrConnClosed, "self close")
	c.triggerRead(err)
	c.triggerWrite(err)
	c.triggerReadLoop(err)
	c.triggerWriteLoop(err)
	log.Infof("[conn %s] waiting for the close callback to execute", c.id)
	err = c.closeCallback()
	if err != nil {
		log.Infof("[conn %s] closed with error: %s", c.id, err.Error())
	} else {
		log.Infof("[conn %s] closed", c.id)
	}
	return err
}

func (c *connection) readLoop() {
	for {
		log.Infof("[conn %s] sumbit prep read event", c.id)
		eventData := PrepReadEventData{}
		eventData.size = defaultReadSize
		eventData.data = c.inputBuffer.Book(defaultReadSize)
		c.operator.submit(RingPrepRead, eventData)
		err := <-c.readLoopTrigger
		if err != nil {
			log.Infof("[conn %s] read loop quit since connection closed", c.id)
			return
		}
	}
}

func (c *connection) writeLoop() {
	for {
		size := c.outputBuffer.Len()
		if size > 0 && c.lock(flushing) {
			log.Infof("[conn %s] sumbit prep write event", c.id)
			eventData := PrepWriteEventData{}
			eventData.size = size
			eventData.data, _ = c.outputBuffer.Seek(size)
			c.operator.submit(RingPRepWrite, eventData)
		}
		err := <-c.writeLoopTrigger
		if err != nil {
			log.Infof("[conn %s] read loop quit since connection closed", c.id)
			return
		}
	}
}
