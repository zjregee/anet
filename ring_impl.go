package anet

/*
#cgo LDFLAGS: -luring
#include <liburing.h>
#include <sys/socket.h>
*/
import "C"

import (
	"net"
	"unsafe"
	"errors"
)

const (
	DEFAULTRINGSIZE = 1024
)

func openDefaultPoll() (*defaultRing, error) {
	var ring = new(defaultRing)
	C.io_uring_queue_init(DEFAULTRINGSIZE, &ring.ring, 0)
	return ring, nil
}

func encodeUserData(event RingEvent, fd int) uint64 {
	return uint64(fd) | (uint64(event) << 56)
}

func decodeUserData(data uint64) (RingEvent, int) {
	event := RingEvent(data >> 56)
	fd := int(data & 0xffffffff)
	return event, fd
}

type defaultRing struct {
	ring    C.struct_io_uring
	opcache *operatorCache
}

func (p *defaultRing) detach(operator *FDOperator) {
	
}

func (p *defaultRing) handler() bool {
	return true
}

func (p *defaultRing) Wait() error {
	for {
		var cqes []*C.struct_io_uring_cqe
		for {
			var cqe *C.struct_io_uring_cqe
			if C.io_uring_wait_cqe(&p.ring, &cqe) < 0 {
				return errors.New("failed to peed for CQE")
			}
			if cqe == nil {
				break
			}
			cqes = append(cqes, cqe)
		}
		for _, cqe := range cqes {
			// if (cqe.flags & C.IORING_CQE_F_READ) != 0 {

			// }

			C.io_uring_cqe_seen(&p.ring, cqe)
		}
	}
}

type PrepAcceptEventData struct {

}

type PrepReadEventData struct {

}

type PrepWriteEventData struct {

}

func (p *defaultRing) Control(operator *FDOperator, event RingEvent, eventData interface{}) error {
	switch event {
	case RingPrepAccept:
		listener, err := net.Listen("tcp", "127.0.0.1:8080")
		if err != nil {
			return errors.New("failed to create listener")
		}
		fd, err := listener.(*net.TCPListener).File()
		if err != nil {
			return errors.New("failed to create tcp listener")
		}
		c_fd := C.int(fd.Fd())
		var sockaddr C.struct_sockaddr
		var addrlen C.socklen_t = C.sizeof_struct_sockaddr
		sqe := C.io_uring_get_sqe(&p.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		C.io_uring_prep_accept(sqe, c_fd, &sockaddr, &addrlen, 0)
		if (C.io_uring_submit(&p.ring) < 0) {
			return errors.New("failed to submit SQE")
		}
	case RingPrepRead:
		bufferSize := 1024
		buffer := make([]byte, bufferSize)
		sqe := C.io_uring_get_sqe(&p.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		C.io_uring_prep_read(sqe, 0, unsafe.Pointer(&buffer[0]), C.uint(bufferSize), 0)
		if (C.io_uring_submit(&p.ring) < 0) {
			return errors.New("failed to submit SQE")
		}
	case RingPrepWrite:
		dataSize := 1024
		data := []byte{}
		sqe := C.io_uring_get_sqe(&p.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		C.io_uring_prep_write(sqe, 0, unsafe.Pointer(&data[0]), C.uint(dataSize), 0)
		if C.io_uring_submit(&p.ring) < 0 {
			return errors.New("failed to submit SQE")
		}
	default:
		return errors.New("unsupported RingEvent")
	}
	return nil
}

func (p *defaultRing) Close() error {
	return nil
}

func (p *defaultRing) Alloc() (operator *FDOperator) {
	op := p.opcache.alloc()
	op.ring = p
	return op
}

func (p *defaultRing) Free(operator *FDOperator) {
	p.opcache.freeable(operator)
}
