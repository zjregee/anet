package main

import (
	"os"
	"net"

	"github.com/zjregee/anet"
	"github.com/sirupsen/logrus"
)

const BUFFER_SIZE = 1024

type connection struct {
	fd           int
	file         *os.File
	conn         net.Conn
	logger       *logrus.Logger
	readTrigger  chan error
	writeTrigger chan error
	size         int
	buffer       []byte
	operator     *anet.FDOperator
}

func (c *connection) init(conn net.Conn, logger *logrus.Logger) error {
	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		return err
	}
	c.fd = int(file.Fd())
	c.file = file
	c.conn = conn
	c.logger = logger
	c.readTrigger = make(chan error)
	c.writeTrigger = make(chan error)
	c.buffer = make([]byte, BUFFER_SIZE)
	c.size = 0
	return c.initFDOperator()
}

func (c *connection) initFDOperator() error {
	ring := anet.RingManager.Pick()
	op := ring.Alloc()
	op.FD = c.fd
	op.OnRead = c.onRead
	op.OnWrite = c.onWrite
	op.Ring = ring
	op.Register()
	c.operator = op
	return nil
}

func (c *connection) close() {
	c.operator.Free()
	c.file.Close()
	c.conn.Close()
}

func (c *connection) run() {
	defer c.close()

	for {
		readEventData := anet.RingEventData{}
		readEventData.Size = BUFFER_SIZE
		readEventData.Data = c.buffer
		readEventData.Event = anet.RingPrepRead
		c.operator.Submit(readEventData)
		err := <-c.readTrigger
		if err != nil {
			c.logger.Warnf("error occurred when read: %s", err.Error())
			return
		}

		if (c.size == 0) {
			return
		}

		writeEventData := anet.RingEventData{}
		writeEventData.Size = c.size
		writeEventData.Data = c.buffer
		writeEventData.Event = anet.RingPrepWrite
		c.operator.Submit(writeEventData)
		err = <-c.writeTrigger
		if err != nil {
			c.logger.Warnf("error occurred when write: %s", err.Error())
			return
		}
	}
}

func (c *connection) onRead(n int, err error) {
	c.size = n
	c.readTrigger <- err
}

func (c *connection) onWrite(n int, err error) {
	c.writeTrigger <- err
}
