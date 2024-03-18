package anet

type Reader interface {
	Next(n int) ([]byte, error)
	Book(n int) []byte
	BookAck(n int) error
	Len() int
}

type Writer interface {
	
}
