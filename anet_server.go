package anet

import (
	"net"
	"sync"
	"time"
	"context"
	"strings"
)

func newServer(ln net.Listener, opts *options, onQuit func(err error)) *server {
	return &server{
		ln: ln,
		opts: opts,
		onQuit: onQuit,
	}
}

type server struct {
	ln          net.Listener
	opts        *options
	connections sync.Map
	onQuit      func(err error)
}

func (s *server) run() {
	for {
		log.Info("waiting for connection accept")
		conn, err := s.ln.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				log.Error("connection accepted with closed, server quit")
				s.onQuit(err)
				return
			}
			log.Warn("onnection accepted with error, wait 10 ms for retry")
			time.Sleep(10 * time.Millisecond)
			continue
		}
		log.Info("connection accepted")
		go s.onAccept(conn)
	}
}

func (s *server) onAccept(conn net.Conn) {
	fd, err := conn.(*net.TCPConn).File()
	if err != nil {
		log.Error("can't get connection's fd")
		return
	}
	connection := &connection{}
	connection.init(conn, s.opts)
	connection.AddCloseCallback(func(connection Connection) error {
		s.connections.Delete(fd.Fd())
		return nil
	})
	s.connections.Store(fd.Fd(), connection)
	connection.onPrepare()
}

func (s *server) close(ctx context.Context) error {
	err := s.ln.Close()
	if err != nil {
		log.Error("error occured while close listener")
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var hasConn bool
	for {
		hasConn = false
		s.connections.Range(func(key, value interface{}) bool {
			hasConn = true
			connection := value.(*connection)
			err := connection.Close()
			if err != nil {
				log.Error("error occured while close connection")
			}
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
