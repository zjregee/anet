package anet

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
)

func CreateListener(network, addr string) (net.Listener, error) {
	return net.Listen(network, addr)
}

func NewEventLoop(onRequest OnRequest, ops ...Option) (EventLoop, error) {
	opts := &options{
		onRequest: onRequest,
	}
	for _, do := range ops {
		do.f(opts)
	}
	return &eventLoop{
		id:   uuid.New().String()[:8],
		opts: opts,
	}, nil
}

type EventLoop interface {
	Serve(ln net.Listener) error
	Shutdown(ctx context.Context) error
}

type OnRequest func(ctx context.Context, connection Connection) error

type eventLoop struct {
	id   string
	opts *options
	ln   net.Listener
}

func (evl *eventLoop) Serve(ln net.Listener) error {
	evl.ln = ln
	for {
		conn, err := evl.ln.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				log.Warnf("[eventloop %s] eventloop quit since listener closed", evl.id)
				return nil
			}
			log.Warnf("[evetloop %s] listener accepted with error, wait 10 ms for retry", evl.id)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		go evl.onAccept(conn)
	}
}

func (evl *eventLoop) onAccept(conn net.Conn) {
	connection := &connection{}
	connection.init(conn, evl.opts)
	connection.run()
}

func (evl *eventLoop) Shutdown(_ context.Context) error {
	return evl.ln.Close()
}
