package anet

import (
	"net"
	"time"
	"errors"
	"context"
	"sync/atomic"

	"github.com/google/uuid"
)

type connection struct {
	locker
	onEvent
	id               string
	fd               int
	opts             *options
	context          context.Context
	operator         *FDOperator
	waitReadSize     int32
	readTimeout      time.Duration
	writeTimeout     time.Duration
	readTrigger      chan error
	writeTrigger     chan error
	readLoopTrigger  chan error
	writeLoopTrigger chan error
	inputBuffer      ReadWriter
	outputBuffer     ReadWriter
	state            int32 // 0: not connected, 1: connected, 2: closed
}

var _ Reader     = &connection{}
var _ Writer     = &connection{}
var _ ReadWriter = &connection{}

func (c *connection) Seek(n int) ([]byte, error) {
	return c.inputBuffer.Seek(n)
}

func (c *connection) SeekAck(n int) error {
	return c.inputBuffer.SeekAck(n)
}

func (c *connection) ReadAll() ([]byte, error) {
	return c.inputBuffer.ReadAll()
}

func (c *connection) ReadUtil(n int) ([]byte, error) {
	if c.inputBuffer.Len() >= n {
		return c.inputBuffer.ReadBytes(n)
	}
	if !c.isActive() {
		// connection is closed, shoud not wait here
		return nil, Exception(ErrConnClosed, "when ReadUtil")
	}
	var err error
	if c.readTimeout != 0 {
		err = c.waitReadWithTimeout(n, c.readTimeout)
	} else {
		err = c.waitRead(n)
	}
	if err != nil {
		return nil, err
	}
	return c.inputBuffer.ReadBytes(n)
}

func (c *connection) ReadBytes(n int) ([]byte, error) {
	return c.inputBuffer.ReadBytes(n)
}

func (c *connection) ReadString(n int) (string, error) {
	return c.inputBuffer.ReadString(n)
}

func (c *connection) Len() int {
	return c.inputBuffer.Len()
}

func (c *connection) WriteBytes(data []byte, n int) error {
	err := c.outputBuffer.WriteBytes(data, n)
	c.triggerWriteLoop(nil)
	return err
}

func (c *connection) WriteString(data string, n int) error {
	err := c.outputBuffer.WriteString(data, n)
	c.triggerWriteLoop(nil)
	return err
}

func (c *connection) Flush() error {
	if !c.isActive() {
		// connection is closed, should not wait here
		return Exception(ErrConnClosed, "when flush")
	}
	c.triggerWriteLoop(nil)
	var err error
	if c.readTimeout != 0 {
		err = c.waitFlushWithTimeout(c.readTimeout)
	} else {
		err = c.waitFlush()
	}
	return err
}

func (c *connection) Book(n int) []byte {
	return c.outputBuffer.Book(n)
}

func (c *connection) BookAck(n int) error {
	err := c.outputBuffer.BookAck(n)
	c.triggerWriteLoop(nil)
	return err
}

func (c *connection) Reader() Reader {
	return c
}

func (c *connection) Writer() Writer {
	return c
}

func (c *connection) SetReadTimeout(timeout time.Duration) {
	c.readTimeout = timeout
}

func (c *connection) SetWriteTimeout(timeout time.Duration) {
	c.writeTimeout = timeout
}

func (c *connection) Id() string {
	return c.id
}

func (c *connection) Close() error {
	return c.onClose()
}

func (c *connection) init(conn net.Conn, opts *options) {
	c.opts = opts
	c.waitReadSize = 0
	c.readTimeout = 0
	c.writeTimeout = 0
	c.readTrigger = make(chan error)
	c.writeTrigger = make(chan error)
	c.readLoopTrigger = make(chan error)
	c.writeLoopTrigger = make(chan error)
	c.inputBuffer = newBytesBuffer(4096)
	c.outputBuffer = newBytesBuffer(4096)
	c.state = 0
	c.id = uuid.New().String()[:8]
	c.initNetFD(conn)
	c.initFDOperator()
	c.initFinalizer()
}

func (c *connection) initNetFD(conn net.Conn) {
	fd, err := conn.(*net.TCPConn).File()
	if err != nil {
		panic("can't panic here")
	}
	c.fd = int(fd.Fd())
}

func (c *connection) initFDOperator() {
	ring := RingManager.Pick()
	op := ring.Alloc()
	op.FD = c.fd
	op.OnRead = c.onRead
	op.OnWrite = c.onWrite
	op.Ring = ring
	c.operator = op
}

func (c *connection) initFinalizer() {
	c.AddCloseCallback(func(connection Connection) error {
		c.operator.Free()
		return nil
	})
}

func (c *connection) isActive() bool {
	return c.isClosedBy(none)
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

func (c *connection) triggerReadLoop(err error) {
	select {
	case c.readLoopTrigger <- err:
	default:
	}
}

func (c *connection) triggerWriteLoop(err error) {
	select {
	case c.writeLoopTrigger <- err:
	default:
	}
}

func (c *connection) waitRead(n int) error {
	if c.inputBuffer.Len() > n {
		return nil
	}
	atomic.StoreInt32(&c.waitReadSize, int32(n))
	defer atomic.StoreInt32(&c.waitReadSize, 0)
	for (c.inputBuffer.Len() < n) {
		err := <-c.readTrigger
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *connection) waitReadWithTimeout(n int, timeout time.Duration) error {
	if c.inputBuffer.Len() > n {
		return nil
	}
	atomic.StoreInt32(&c.waitReadSize, int32(n))
	defer atomic.StoreInt32(&c.waitReadSize, 0)
	timer := time.NewTimer(timeout)
	for (c.inputBuffer.Len() < n) {
		select {
		case err := <-c.readTrigger:
			if err != nil {
				return err
			}
		case <-timer.C:
			return errors.New("timeout")
		}
	}
	return nil
}

func (c *connection) waitFlush() error {
	if c.outputBuffer.Len() == 0 {
		return nil
	}
	for (c.outputBuffer.Len() > 0) {
		err := <- c.writeTrigger
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *connection) waitFlushWithTimeout(timeout time.Duration) error {
	if c.outputBuffer.Len() == 0 {
		return nil
	}
	timer := time.NewTimer(timeout)
	for (c.outputBuffer.Len() > 0) {
		select {
		case err := <-c.writeTrigger:
			if err != nil {
				return err
			}
		case <-timer.C:
			return Exception(ErrWriteTimeout, "")
		}
	}
	return nil
}
