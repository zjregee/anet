package anet

/*
#cgo LDFLAGS: -luring
#include <liburing.h>
*/
import "C"

import (
	"sync"
	"errors"
	"unsafe"
	"syscall"

	"github.com/google/uuid"
)

const (
	DEFAULT_RING_SIZE = 1024
)

func newDefaultRing() (Ring, error) {
	ring := &defaultRing{}
	C.io_uring_queue_init(DEFAULT_RING_SIZE, &ring.ring, 0)
	ring.id = uuid.New().String()[:8]
	ring.opcache = sync.Pool{
		New: func() interface{} {
			return &FDOperator{}
		},
	}
	return ring, nil
}

func encodeUserData(event RingEvent, fd int) uint64 {
	if fd > (1 << 56) {
		panic("encodeUserData panicked: fd will be lost")
	}
	return uint64(fd) | (uint64(event) << 56)
}

func decodeUserData(data uint64) (RingEvent, int) {
	event := RingEvent(data >> 56)
	fd := int(data & 0xffffffffffffff)
	return event, fd
}

type defaultRing struct {
	mu      sync.Mutex
	id      string
	opmap   sync.Map
	opcache sync.Pool
	ring    C.struct_io_uring
}

func (r *defaultRing) Id() string {
	return r.id
}

func (r *defaultRing) Wait() error {
	var cqe *C.struct_io_uring_cqe
	for {
		C.io_uring_wait_cqe(&r.ring, &cqe)
		if cqe == nil {
			continue
		}
		C.io_uring_cqe_seen(&r.ring, cqe)

		userData := uint64(cqe.user_data)
		event, fd := decodeUserData(userData)
		operator := r.getOperator(fd)
		switch event {
		case RingPrepRead:
			if cqe.res < 0 {
				errno := syscall.Errno(-cqe.res)
				if errno == syscall.EAGAIN {
					log.Warnf("[ring %s] error occurred while waiting, wait 10ms to retry: %s", r.id, errno.Error())
					continue
				} else {
					operator.OnRead(int(cqe.res), errno)
				}
			} else {
				operator.OnRead(int(cqe.res), nil)
			}
		case RingPRepWrite:
			if cqe.res < 0 {
				errno := syscall.Errno(-cqe.res)
				if errno == syscall.EAGAIN {
					log.Warnf("[ring %s] error occurred while waiting, wait 10ms to retry: %s", r.id, errno.Error())
					continue
				} else {
					operator.OnWrite(int(cqe.res), errno)
				}
			} else {
				operator.OnWrite(int(cqe.res), nil)
			}
		default:
			log.Warnf("[ring %s] unsupported RingEvent", r.id)
		}
	}
}

func (r *defaultRing) Submit(operator *FDOperator, event RingEvent, eventData interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch event {
	case RingPrepRead:
		sqe := C.io_uring_get_sqe(&r.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		data := eventData.(PrepReadEventData)
		userData := encodeUserData(RingPrepRead, operator.FD)
		sqe.user_data = C.ulonglong(userData)
		C.io_uring_prep_read(sqe, C.int(operator.FD), unsafe.Pointer(&data.Data[0]), C.uint(data.Size), 0)
		if C.io_uring_submit(&r.ring) < 0 {
			return errors.New("failed to submit SQE")
		}
	case RingPRepWrite:
		sqe := C.io_uring_get_sqe(&r.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		data := eventData.(PrepWriteEventData)
		userData := encodeUserData(RingPRepWrite, operator.FD)
		sqe.user_data = C.ulonglong(userData)
		C.io_uring_prep_write(sqe, C.int(operator.FD), unsafe.Pointer(&data.Data[0]), C.uint(data.Size), 0)
		if C.io_uring_submit(&r.ring) < 0 {
			return errors.New("failed to submit SQE")
		}
	default:
		return errors.New("unsupported RingEvent")
	}
	return nil
}

func (r *defaultRing) Close() error {
	return nil
}

func (r *defaultRing) Alloc() *FDOperator {
	return r.opcache.Get().(*FDOperator)
}

func (r *defaultRing) Free(operator *FDOperator) {
	r.delOperator(operator.FD)
	operator.Reset()
	r.opcache.Put(operator)
}

func (r *defaultRing) Register(operator *FDOperator) {
	r.opmap.Store(operator.FD, operator)
}

func (r *defaultRing) getOperator(fd int) *FDOperator {
	operator, ok := r.opmap.Load(fd)
	if !ok {
		return nil
	}
	return operator.(*FDOperator)
}

func (r *defaultRing) delOperator(fd int) {
	r.opmap.Delete(fd)
}

func (r *defaultRing) onClose() {
	C.io_uring_queue_exit(&r.ring)
}
