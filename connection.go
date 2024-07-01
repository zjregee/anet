package anet

import (
	"time"
)

type CloseCallback func(connection Connection) error

type Connection interface {
	Reader() Reader
	Writer() Writer
	SetReadTimeout(timeout time.Duration)
	SetWriteTimeout(timeout time.Duration)
	SetOnRequest(onRequest OnRequest)
	AddCloseCallback(callback CloseCallback)
	Close() error
}
