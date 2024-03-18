package anet

import (
	"net"
	"sync/atomic"
)

type connection struct {
	locker
	onEvent
	operator     *FDOperator
	readTrigger  chan error
	writeTrigger chan error
	inputBuffer  []byte
	outputBuffer []byte
	waitReadSize int64
}

func (c *connection) init(conn net.Conn, opts *options) error {
	return nil
}

func (c *connection) triggerRead(err error) {
	select {
	case c.readTrigger <- err:
	default:
	}
}

func (c *connection) triggerWrite(err error) {
	select {
	case c.writeTrigger <- err:
	default:
	}
}

func (c *connection) waitRead(n int) (err error) {
	if n < len(c.inputBuffer) {
		return nil
	}
	atomic.StoreInt64(&c.waitReadSize, int64(n))
	defer atomic.StoreInt64(&c.waitReadSize, 0)
	for len(c.inputBuffer) < n {
		err = <-c.readTrigger
		if err != nil {
			return err
		}
	}
	return nil
}
