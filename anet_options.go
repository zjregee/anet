package anet

import (
	"time"

	"github.com/sirupsen/logrus"
)

func SetLogger(logger *logrus.Logger) {
	setLogger(logger)
}

func SetLoggerLevel(level logrus.Level) {
	setLoggerLevel(level)
}

func WithOnConnect(onConnect OnConnect) Option {
	return Option{
		f: func(op *options) {
			op.onConnect = onConnect
		},
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return Option{
		f: func(op *options) {
			op.readTimeout = timeout
		},
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return Option{
		f: func(op *options) {
			op.writeTimeout = timeout
		},
	}
}

type Option struct {
	f func(*options)
}

type options struct {
	onConnect    OnConnect
	onRequest    OnRequest
	readTimeout  time.Duration
	writeTimeout time.Duration
}
