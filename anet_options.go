package anet

import (
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	RingManager = newDefaultRingManager(4)
}

func SetLogger(logger *logrus.Logger) {
	log = logger
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
	onRequest    OnRequest
	readTimeout  time.Duration
	writeTimeout time.Duration
}
