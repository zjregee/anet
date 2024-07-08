package anet

import (
	"net"
	"sync"
	"context"
	// "runtime"
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
		opts: opts,
		stop: make(chan error, 1),
	}, nil
}

type EventLoop interface {
	Serve(ln net.Listener) error
	ServeNonBlocking(ln net.Listener) error
	Shutdown() error
	ShutdownWithContext(ctx context.Context) error
}

type OnRequest func(ctx context.Context, connection Connection) error

type OnConnect func(ctx context.Context, connection Connection) context.Context

func init() {
	log = newDefaultLogger()
	// n := (runtime.GOMAXPROCS(0) - 1) / 20 + 1
	RingManager = newDefaultRingManager(4)
}

type eventLoop struct {
	sync.Mutex
	opts *options
	svr  *server
	stop chan error
}

func (evl *eventLoop) Serve(ln net.Listener) error {
	evl.Lock()
	evl.svr = newServer(ln, evl.opts, func(err error) {
		evl.Lock()
		if evl.svr == nil {
			evl.Unlock()
			return
		}
		evl.svr = nil
		evl.Unlock()
		select {
		case evl.stop <- err:
		default:
		}
	})
	go evl.svr.run()
	evl.Unlock()

	err := <-evl.stop
	if err != nil {
		log.Error("[EventLoop] error eccoured while serving")
	}
	return err
}

func (evl *eventLoop) ServeNonBlocking(ln net.Listener) error {
	evl.Lock()
	evl.svr = newServer(ln, evl.opts, func(err error) {
		evl.Lock()
		if evl.svr == nil {
			evl.Unlock()
			return
		}
		evl.svr = nil
		evl.Unlock()
		select {
		case evl.stop <- err:
		default:
		}
	})
	go evl.svr.run()
	evl.Unlock()
	return nil
}

func (evl *eventLoop) Shutdown() error {
	return evl.ShutdownWithContext(context.Background())
}

func (evl *eventLoop) ShutdownWithContext(ctx context.Context) error {
	evl.Lock()
	if evl.svr == nil {
		evl.Unlock()
		return nil
	}
	svr := evl.svr
	evl.svr = nil
	evl.Unlock()

	log.Info("[EventLoop] closed by user")
	select {
	case evl.stop <- nil:
	default:
	}
	err := svr.close(ctx)
	if err != nil {
		log.Error("[EventLoop] error occurred while closing")
	}
	return err
}
