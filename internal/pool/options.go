package pool

import "time"

type Option func(opts *Options)

type Options struct {
	ExpiryDuration   time.Duration
	MaxBlockingTasks int
	NonBlocking      bool
	PanicHandler     func(interface{})
	DisablePurge     bool
}

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(opts *Options) {
		opts.ExpiryDuration = expiryDuration
	}
}

func WithMaxBlockingTasks(maxBlockingTasks int) Option {
	return func(opts *Options) {
		opts.MaxBlockingTasks = maxBlockingTasks
	}
}

func WithNonblocking(nonblocking bool) Option {
	return func(opts *Options) {
		opts.NonBlocking = nonblocking
	}
}

func WithPanicHandler(panicHandler func(interface{})) Option {
	return func(opts *Options) {
		opts.PanicHandler = panicHandler
	}
}

func WithDisablePurge(disable bool) Option {
	return func(opts *Options) {
		opts.DisablePurge = disable
	}
}

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}
