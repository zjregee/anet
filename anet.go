package anet

import (
	"context"
	"net"
	"sync"
)

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

type eventLoop struct {
	sync.Mutex
	opts *options
	svr  *server
	stop chan error
}

func (evl *eventLoop) Serve(ln net.Listener) error {
	evl.Lock()
	evl.svr = newServer(ln, evl.opts, evl.quit)
	evl.svr.run()
	evl.Unlock()
	return evl.waitQuit()
}

func (evl *eventLoop) Shutdown(ctx context.Context) error {
	evl.Lock()
	svr := evl.svr
	evl.svr = nil
	evl.Unlock()
	if svr == nil {
		return nil
	}
	evl.quit(nil)
	return svr.close(ctx)
}

func (evl *eventLoop) waitQuit() error {
	return <-evl.stop
}

func (evl *eventLoop) quit(err error) {
	select {
	case evl.stop <- err:
	default:
	}
}
