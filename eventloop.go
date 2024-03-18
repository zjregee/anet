package anet

import (
	"net"
	"context"
)

type EventLoop interface {
	Serve(ln net.Listener) error
	Shutdown(ctx context.Context) error
}

type OnPrepare func(connection Connection) context.Context

type OnConnect func(ctx context.Context, connection Connection) context.Context

type OnRequest func(ctx context.Context, connection Connection) error
