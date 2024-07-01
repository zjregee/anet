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

	"golang.org/x/sys/unix"
	"github.com/google/uuid"
)

const (
	DEFAULT_RING_SIZE = 1024
)

func newDefaultRing()  (Ring, error) {
	ring := &defaultRing{}
	C.io_uring_queue_init(DEFAULT_RING_SIZE, &ring.ring, 0)
	ring.id = uuid.New().String()[:8]
	ring.opcache = sync.Pool{
		New: func() interface{} {
			return &FDOperator{}
		},
	}
	// fd, err := unix.Eventfd(0, 0)
	// if err != nil {
	// 	return nil, err
	// }
	// err = ring.registerEventFD(fd)
	// if err != nil {
	// 	return nil, err
	// }
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
	id      string
	wop    	*FDOperator
	buf     []byte
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
		if C.io_uring_wait_cqe(&r.ring, &cqe) < 0 {
			log.Fatalf("[ring %s] quit since error occurred while waiting", r.id)
			r.onClose()
			return errors.New("failed to wait CQE")
		}
		if cqe.res < 0 {
			errno := syscall.Errno(-cqe.res)
			if errno == syscall.EAGAIN {
				log.Fatalf("[ring %s] error occurred while waiting, wait 10ms to retry", r.id)
				continue
			} else {
				log.Fatalf("[ring %s] quit since error occurred while waiting", r.id)
				r.onClose()
				return errors.New("failed to wait CQE")
			}
		}
		userData := uint64(cqe.user_data)
		event, fd := decodeUserData(userData)
		operator := r.getOperator(fd)
		switch event {
		case RingPrepRead:
			if r.wop != nil && operator.FD == r.wop.FD {
				log.Infof("[ring %s] quit since eventfd triggered", r.id)
				C.io_uring_cqe_seen(&r.ring, cqe)
				r.onClose()
				return nil
			}
			log.Infof("[ring %s] new event has completed, fd: %d, event: prep read", r.id, operator.FD)
			operator.OnRead(int(cqe.res))
		case RingPRepWrite:
			log.Infof("[ring %s] new event has completed, fd: %d, event: prep write", r.id, operator.FD)
			operator.OnWrite(int(cqe.res))
		default:
			log.Fatalf("[ring %s] quit since error occurred while waiting", r.id)
			return errors.New("unsupported RingEvent")
		}
		C.io_uring_cqe_seen(&r.ring, cqe)
	}
}

func (r *defaultRing) Submit(operator *FDOperator, event RingEvent, eventData interface{}) error {
	switch event {
	case RingPrepRead:
		log.Infof("[ring %s] new event has sumitted, fd: %d, event: prep read", r.id, operator.FD)
		sqe := C.io_uring_get_sqe(&r.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		data := eventData.(PrepReadEventData)
		userData := encodeUserData(RingPrepRead, operator.FD)
		sqe.user_data = C.ulonglong(userData)
		r.setOperator(operator.FD, operator)
		C.io_uring_prep_read(sqe, C.int(operator.FD), unsafe.Pointer(&data.data[0]), C.uint(data.size), 0)
		if C.io_uring_submit(&r.ring) < 0 {
			return errors.New("failed to submit SQE")
		}
	case RingPRepWrite:
		log.Infof("[ring %s] new event has sumitted, fd: %d, event: prep write", r.id, operator.FD)
		sqe := C.io_uring_get_sqe(&r.ring)
		if sqe == nil {
			return errors.New("failed to get SQE")
		}
		data := eventData.(PrepWriteEventData)
		userData := encodeUserData(RingPRepWrite, operator.FD)
		sqe.user_data = C.ulonglong(userData)
		r.setOperator(operator.FD, operator)
		C.io_uring_prep_write(sqe, C.int(operator.FD), unsafe.Pointer(&data.data[0]), C.uint(data.size), 0)
		if C.io_uring_submit(&r.ring) < 0 {
			return errors.New("failed to submit SQE")
		}
	default:
		return errors.New("unsupported RingEvent")
	}
	return nil
}

func (r *defaultRing) Close() error {
	if r.wop == nil {
		return nil
	}
	_, err := unix.Write(r.wop.FD, []byte("1"))
	return err
}

func (r *defaultRing) Alloc() *FDOperator {
	return r.opcache.Get().(*FDOperator)
}

func (r *defaultRing) Free(operator *FDOperator) {
	r.delOperator(operator.FD)
	operator.reset()
	r.opcache.Put(operator)
}

// will cased failure for now
func (r *defaultRing) registerEventFD(fd int) error {
	operator := &FDOperator{}
	operator.FD = fd
	r.wop = operator
	r.buf = make([]byte, 1)
	r.setOperator(operator.FD, operator)
	log.Infof("[ring %s] register eventfd", r.id)
	sqe := C.io_uring_get_sqe(&r.ring)
	if sqe == nil {
		return errors.New("failed to get SQE")
	}
	userData := encodeUserData(RingPrepRead, fd)
	sqe.user_data = C.ulonglong(userData)
	C.io_uring_prep_read(sqe, C.int(operator.FD), unsafe.Pointer(&r.buf[0]), 1, 0)
	if C.io_uring_submit(&r.ring) < 0 {
		return errors.New("failed to submit SQE")
	}
	return nil
}

func (r *defaultRing) getOperator(fd int) *FDOperator {
	operator, ok := r.opmap.Load(fd)
	if !ok {
		return nil
	}
	return operator.(*FDOperator)
}

func (r *defaultRing) setOperator(fd int, operator *FDOperator) {
	r.opmap.Store(fd, operator)
}

func (r *defaultRing) delOperator(fd int) {
	r.opmap.Delete(fd)
}

func (r *defaultRing) onClose() {
	if r.wop != nil {
		err := unix.Close(r.wop.FD)
		if err != nil {
			log.Fatalf("[ring %s] error occurred when close eventfd", r.id)
		}
	}
	C.io_uring_queue_exit(&r.ring)
}
