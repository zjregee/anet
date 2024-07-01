package main

/*
#cgo LDFLAGS: -luring
#include <liburing.h>
*/
import "C"

import (
	"fmt"
	"net"
	"os"
	"unsafe"
	"syscall"
)

const (
	QUEUE_DEPTH = 128
	BUFFER_SIZE = 1024
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on :8080")
	var ring C.struct_io_uring
	C.io_uring_queue_init(C.uint(QUEUE_DEPTH), &ring, 0)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn, &ring)
	}
}

func handleConnection(conn net.Conn, ring *C.struct_io_uring) {
	defer conn.Close()
	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		fmt.Println("Error converting connection: ",err.Error())
		return
	}
	fd := file.Fd()
	buffer := make([]byte, BUFFER_SIZE)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading connection: ",err.Error())
			return
		}
		sqe := C.io_uring_get_sqe(ring)
		C.io_uring_prep_write(sqe, C.int(fd), unsafe.Pointer(&buffer[0]), C.uint(n), 0)
		C.io_uring_submit(ring)
		var cqe *C.struct_io_uring_cqe
		for {
			C.io_uring_wait_cqe(ring, &cqe)
			if cqe.res < 0 {
				errno := syscall.Errno(-cqe.res)
				if errno == syscall.EAGAIN {
					continue
				} else {
					fmt.Println("Error writing connection: ", errno.Error())
					return
				}
			} else {
				break
			}
		}
		C.io_uring_cqe_seen(ring, cqe)
	}
}
