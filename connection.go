package anet

type CloseCallback func(connection Connection) error

type Connection interface {
	ID() string
	Reader() Reader
	Writer() Writer
	AddCloseCallback(callback CloseCallback)
	Close() error
}
