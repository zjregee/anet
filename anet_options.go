package anet

func SetNumLoops(numLoops int) error {
	return setNumLoops(numLoops)
}

func WithOnPrepare(onPrepare OnPrepare) Option {
	return Option{func(op *options) {
		op.onPrepare = onPrepare
	}}
}

func WithOnConnect(onConnect OnConnect) Option {
	return Option{func(op *options) {
		op.onConnect = onConnect
	}}
}

type Option struct {
	f func(*options)
}

type options struct {
	onPrepare OnPrepare
	onConnect OnConnect
	onRequest OnRequest
}
