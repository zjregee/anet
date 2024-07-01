package anet

type Reader interface {
	Seek(n int) ([]byte, error)
	SeekAck(n int) error
	ReadAll() ([]byte, error)
	ReadUtil(n int) ([]byte, error)
	ReadBytes(n int) ([]byte, error)
	ReadString(n int) (string, error)
	Len() int
}

type Writer interface {
	Book(n int) []byte
	BookAck(n int) error
	WriteBytes(data []byte, n int) error
	WriteString(data string, n int) error
	Flush() error
}

type ReadWriter interface {
	Reader
	Writer
}
