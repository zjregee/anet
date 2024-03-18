package anet

type CloseCallback func(connection Connection) error

type Connection interface {
	SetOnRequest(on OnRequest) error
	AddCloseCallback(callback CloseCallback) error
}