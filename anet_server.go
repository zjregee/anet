package anet

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

func newServer(ln net.Listener, opts *options, onQuit func(err error)) *server {
	return &server{
		ln:     ln,
		opts:   opts,
		onQuit: onQuit,
	}
}

type server struct {
	operator    FDOperator
	ln          net.Listener
	opts        *options
	onQuit      func(err error)
	connections sync.Map
}

func (s *server) run() error {
	fd, err := s.ln.(*net.TCPListener).File()
	if err != nil {
		return errors.New("failed to create tcp listener")
	}
	s.operator = FDOperator{
		FD: int(fd.Fd()),
		OnAccept: s.onAccept,
	}
	s.operator.ring = ringmanager.pick()
	err = s.operator.control(RingPrepAccept)
	if err != nil {
		s.onQuit(err)
	}
	return err
}

func (s *server) onAccept() error {
	conn, err := s.ln.Accept()
	if err != nil {
		if strings.Contains(err.Error(), "closed") {
			s.operator.control(RingDetach)
			s.onQuit(err)
		}
		return err
	}
	if conn == nil {
		return nil
	}
	connection := &connection{}
	connection.init(conn, s.opts)
	fd, err := conn.(*net.IPConn).File()
	if err != nil {
		return errors.New("failed to create tcp connection")
	}
	connection.AddCloseCallback(func(connection Connection) error {
		s.connections.Delete(int(fd.Fd()))
		return nil
	})
	s.connections.Store(int(fd.Fd()), connection)
	connection.onConnect()
	return nil
}

func (s *server) close(ctx context.Context) error {
	s.operator.control(RingDetach)
	_ = s.ln.Close()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var hasConn bool
	for {
		hasConn = false
		s.connections.Range(func(key, value interface{}) bool {
			hasConn = true
			conn := value.(*connection)
			_ = conn.onClose()
			return true
		})
		if !hasConn {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}
