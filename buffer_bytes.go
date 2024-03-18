package anet

import "errors"

func newBytesBuffer(cap int) *BytesBuffer {
	return &BytesBuffer{
		buffer: make([]byte, 0, cap),
		start: 0,
		end: 0,
		cap: cap,
	}
}

type BytesBuffer struct {
	buffer []byte
	start  int
	end    int
	cap    int
}

var _ Reader = &BytesBuffer{}
var _ Writer = &BytesBuffer{}

func (b *BytesBuffer) Next(n int) ([]byte, error) {
	if n > b.end - b.start {
		return nil, errors.New("not enough bytes in buffer")
	}
	data := b.buffer[b.start:b.start + n]
	b.start += n
	return data, nil
}

func (b *BytesBuffer) Book(n int) []byte {
	if b.cap - b.end < n {
		b.cap *= 2
		newBuffer := make([]byte, b.end, b.cap)
		copy(newBuffer, b.buffer)
		b.buffer = newBuffer
	}
	return b.buffer[b.end:b.end + n]
}

func (b *BytesBuffer) BookAck(n int) error {
	if b.cap - b.end < n {
		return errors.New("not enough space in buffer")
	}
	b.end += n
	return nil
}

func (b *BytesBuffer) Len() int {
	return b.end - b.start
}
