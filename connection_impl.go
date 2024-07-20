package anet

import (
	"context"
	"errors"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type connection struct {
	id                string
	fd                int
	file              *os.File
	conn              net.Conn
	context           context.Context
	operator          *FDOperator
	waitReadSize      int32
	readTimeout       time.Duration
	writeTimeout      time.Duration
	readTrigger       chan error
	writeTrigger      chan error
	inputBuffer       ReadWriter
	outputBuffer      ReadWriter
	onRequestCallback OnRequest
	state             int32 // 0: connected, 1: closed
}

var _ Reader = &connection{}
var _ Writer = &connection{}
var _ ReadWriter = &connection{}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) Seek(n int) ([]byte, error) {
	return c.inputBuffer.Seek(n)
}

func (c *connection) SeekAck(n int) error {
	return c.inputBuffer.SeekAck(n)
}

func (c *connection) SeekAll() ([]byte, error) {
	return c.inputBuffer.SeekAll()
}

func (c *connection) ReadAll() ([]byte, error) {
	return c.inputBuffer.ReadAll()
}

func (c *connection) ReadUtil(delim byte) ([]byte, error) {
	return c.waitReadUntil(delim)
}

func (c *connection) ReadBytes(n int) ([]byte, error) {
	if c.inputBuffer.Len() >= n {
		return c.inputBuffer.ReadBytes(n)
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

func (c *connection) ReadString(n int) (string, error) {
	if c.inputBuffer.Len() >= n {
		return c.inputBuffer.ReadString(n)
	}
	var err error
	if c.readTimeout != 0 {
		err = c.waitReadWithTimeout(n, c.readTimeout)
	} else {
		err = c.waitRead(n)
	}
	if err != nil {
		return "", err
	}
	return c.inputBuffer.ReadString(n)
}

func (c *connection) Len() int {
	return c.inputBuffer.Len()
}

func (c *connection) Release() {
	c.inputBuffer.Release()
}

func (c *connection) WriteBytes(data []byte, n int) error {
	err := c.outputBuffer.WriteBytes(data, n)
	return err
}

func (c *connection) WriteString(data string, n int) error {
	err := c.outputBuffer.WriteString(data, n)
	return err
}

func (c *connection) Flush() error {
	var err error
	if c.writeTimeout != 0 {
		err = c.waitFlushWithTimeout(c.writeTimeout)
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
	return err
}

func (c *connection) Reader() Reader {
	return c
}

func (c *connection) Writer() Writer {
	return c
}

func (c *connection) AddCloseCallback(callback CloseCallback) {}

func (c *connection) Close() error {
	c.operator.Free()
	c.file.Close()
	c.conn.Close()
	return nil
}

func (c *connection) init(conn net.Conn, opts *options) {
	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		panic("shouldn't failed here")
	}
	c.fd = int(file.Fd())
	c.file = file
	c.conn = conn

	ring := RingManager.Pick()
	op := ring.Alloc()
	op.FD = c.fd
	op.OnRead = c.onRead
	op.OnWrite = c.onWrite
	op.Ring = ring
	op.Register()
	c.operator = op

	c.state = 0
	c.waitReadSize = 0
	c.readTimeout = 0
	c.writeTimeout = 0
	c.id = uuid.New().String()[:8]
	c.context = context.Background()
	c.readTrigger = make(chan error)
	c.writeTrigger = make(chan error)
	c.inputBuffer = NewBytesBuffer(4096)
	c.outputBuffer = NewBytesBuffer(4096)
	c.onRequestCallback = opts.onRequest
}

func (c *connection) run() {
	defer c.Close()

	for {
		err := c.onRequestCallback(c.context, c)
		if err != nil {
			return
		}
	}
}

func (c *connection) waitRead(n int) error {
	if c.inputBuffer.Len() >= n {
		return nil
	}
	for c.inputBuffer.Len() < n {
		c.submitRead()
		err := <-c.readTrigger
		if err != nil {
			return err
		}
		if c.inputBuffer.Len() == 0 {
			return errors.New("EOF")
		}
	}
	return nil
}

func (c *connection) waitReadUntil(delim byte) ([]byte, error) {
	for {
		if c.inputBuffer.Len() > 0 {
			data, err := c.inputBuffer.SeekAll()
			if err != nil {
				return nil, err
			}
			index := -1
			for i, b := range data {
				if b == delim {
					index = i
					break
				}
			}
			if index != -1 {
				c.inputBuffer.SeekAck(index + 1)
				return data[:index+1], nil
			}
		}
		c.submitRead()
		err := <-c.readTrigger
		if err != nil {
			return nil, err
		}
		if c.inputBuffer.Len() == 0 {
			return nil, errors.New("EOF")
		}
	}
}

func (c *connection) waitFlush() error {
	if c.outputBuffer.Len() == 0 {
		return nil
	}
	for c.outputBuffer.Len() > 0 {
		c.submitWrite()
		err := <-c.writeTrigger
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *connection) waitReadWithTimeout(n int, timeout time.Duration) error {
	if c.inputBuffer.Len() >= n {
		return nil
	}
	atomic.StoreInt32(&c.waitReadSize, int32(n))
	defer atomic.StoreInt32(&c.waitReadSize, 0)
	timer := time.NewTimer(timeout)
	for c.inputBuffer.Len() < n {
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

func (c *connection) waitFlushWithTimeout(timeout time.Duration) error {
	if c.outputBuffer.Len() == 0 {
		return nil
	}
	timer := time.NewTimer(timeout)
	for c.outputBuffer.Len() > 0 {
		select {
		case err := <-c.writeTrigger:
			if err != nil {
				return err
			}
		case <-timer.C:
			return errors.New("timeout")
		}
	}
	return nil
}
