package anet

type Reader interface {
	Seek(n int) ([]byte, error)
	SeekAck(n int) error
	SeekAll() ([]byte, error)
	ReadAll() ([]byte, error)
	ReadUtil(delim byte) ([]byte, error)
	ReadBytes(n int) ([]byte, error)
	ReadString(n int) (string, error)
	Len() int
	Release()
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
