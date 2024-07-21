package anet

/*
#cgo LDFLAGS: -luring
#include <liburing.h>
*/
import "C"

import (
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

const (
	DEFAULT_RING_SIZE  = 1024
	DEFAULT_BATCH_SIZE = 32
)

func newDefaultRing() (Ring, error) {
	ring := &defaultRing{}
	C.io_uring_queue_init(DEFAULT_RING_SIZE, &ring.ring, 0)
	ring.id = uuid.New().String()[:8]
	ring.ch = make(chan RingEventData, 4)
	ring.opcache = sync.Pool{
		New: func() interface{} {
			return &FDOperator{}
		},
	}
	ring.num = 0
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
	opmap   sync.Map
	opcache sync.Pool
	ring    C.struct_io_uring
	ch      chan RingEventData
	num     int
	mu      sync.Mutex
}

func (r *defaultRing) Id() string {
	return r.id
}

func (r *defaultRing) Wait() error {
	go r.submitLoop()
	go r.loop()

	var cqe *C.struct_io_uring_cqe
	cqes := make([]*C.struct_io_uring_cqe, DEFAULT_BATCH_SIZE)
	for {
		C.io_uring_wait_cqe(&r.ring, &cqe)
		if cqe == nil {
			continue
		}
		r.handleEvent(cqe)
		count := C.io_uring_peek_batch_cqe(&r.ring, &cqes[0], DEFAULT_BATCH_SIZE)
		for i := 0; i < int(count); i++ {
			r.handleEvent(cqes[i])
		}
	}
}

func (r *defaultRing) Submit(eventData RingEventData) {
	r.ch <- eventData
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

func (r *defaultRing) submitLoop() {
	for eventData := range r.ch {
		r.mu.Lock()
		sqe := C.io_uring_get_sqe(&r.ring)
		if sqe == nil {
			panic("should't failed here")
		}
		switch eventData.Event {
		case RingPrepRead:
			userData := encodeUserData(RingPrepRead, eventData.Operator.FD)
			sqe.user_data = C.ulonglong(userData)
			C.io_uring_prep_read(sqe, C.int(eventData.Operator.FD), unsafe.Pointer(&eventData.Data[0]), C.uint(eventData.Size), 0)
		case RingPrepWrite:
			userData := encodeUserData(RingPrepWrite, eventData.Operator.FD)
			sqe.user_data = C.ulonglong(userData)
			C.io_uring_prep_write(sqe, C.int(eventData.Operator.FD), unsafe.Pointer(&eventData.Data[0]), C.uint(eventData.Size), 0)
		default:
			panic("should't failed here")
		}
		r.num += 1
		if r.num >= 10 {
			C.io_uring_submit(&r.ring)
			r.num = 0
		}
		r.mu.Unlock()
	}
}

func (r *defaultRing) loop() {
	for {
		time.Sleep(time.Microsecond)
		r.mu.Lock()
		if r.num > 0 {
			C.io_uring_submit(&r.ring)
			r.num = 0
		}
		r.mu.Unlock()
	}
}

func (r *defaultRing) handleEvent(cqe *C.struct_io_uring_cqe) {
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
				time.Sleep(time.Millisecond * 10)
			} else {
				operator.OnRead(int(cqe.res), errno)
			}
		} else {
			operator.OnRead(int(cqe.res), nil)
		}
	case RingPrepWrite:
		if cqe.res < 0 {
			errno := syscall.Errno(-cqe.res)
			if errno == syscall.EAGAIN {
				log.Warnf("[ring %s] error occurred while waiting, wait 10ms to retry: %s", r.id, errno.Error())
				time.Sleep(time.Millisecond * 10)
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
