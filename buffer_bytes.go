package anet

import (
	"errors"
)

func NewBytesBuffer(cap int) ReadWriter {
	return &bytesBuffer{
		buffer: make([]byte, 0, cap),
		start: 0,
		end: 0,
		cap: cap,
	}
}

type bytesBuffer struct {
	buffer []byte
	start  int
	end    int
	cap    int
}

var _ Reader     = &bytesBuffer{}
var _ Writer     = &bytesBuffer{}
var _ ReadWriter = &bytesBuffer{}

func (b *bytesBuffer) Seek(n int) ([]byte, error) {
	if b.end - b.start < n {
		return nil, errors.New("not enough data in buffer")
	}
	data := b.buffer[b.start:b.start + n]
	return data, nil
}

func (b *bytesBuffer) SeekAck(n int) error {
	if b.end - b.start < n {
		return errors.New("not enough data in buffer")
	}
	b.start += n
	return nil
}

func (b *bytesBuffer) ReadAll() ([]byte, error) {
	data := b.buffer[b.start:b.end]
	b.start = b.end
	return data, nil
}

func (b *bytesBuffer) ReadUtil(n int) ([]byte, error) {
	panic("unreachable code")
}

func (b *bytesBuffer) ReadBytes(n int) ([]byte, error) {
	if b.end - b.start < n {
		return nil, errors.New("not enough data in buffer")
	}
	data := b.buffer[b.start:b.start + n]
	b.start += n
	return data, nil
}

func (b *bytesBuffer) ReadString(n int) (string, error) {
	if b.end - b.start < n {
		return "", errors.New("not enough data in buffer")
	}
	data := b.buffer[b.start:b.start + n]
	b.start += n
	return string(data), nil
}

func (b *bytesBuffer) Len() int {
	return b.end - b.start
}

func (b *bytesBuffer) WriteBytes(data []byte, n int) error {
	if b.cap - b.end < n {
		b.increase()
	}
	copy(b.buffer[b.end:b.end + n], data)
	b.end += n
	return nil
}

func (b *bytesBuffer) WriteString(data string, n int) error {
	if b.cap - b.end < n {
		b.increase()
	}
	copy(b.buffer[b.end:b.end + n], []byte(data))
	b.end += n
	return nil
}

func (b *bytesBuffer) Flush() error {
	b.start = 0
	b.end = 0
	return nil
}

func (b *bytesBuffer) Book(n int) []byte {
	if b.cap - b.end < n {
		b.increase()
	}
	return b.buffer[b.end:b.end + n]
}

func (b *bytesBuffer) BookAck(n int) error {
	if b.cap - b.end < n {
		return errors.New("not enough space in buffer")
	}
	b.end += n
	return nil
}

// FIXME: We should guarantee increase size > required
func (b *bytesBuffer) increase() {
	b.cap *= 2
	newBuffer := make([]byte, b.end, b.cap)
	copy(newBuffer, b.buffer[b.start:b.end])
	b.buffer = newBuffer
	b.end -= b.start
	b.start = 0
}
