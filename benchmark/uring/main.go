package main

/*
#cgo LDFLAGS: -luring
#include <liburing.h>
*/
import "C"

import (
	"os"
	"net"
	"fmt"
	"time"
	"bufio"
	"unsafe"
	"syscall"
	"math/rand"

	"github.com/sirupsen/logrus"
)

const (
	QUEUE_DEPTH = 1024
	BUFFER_SIZE = 1024
)

func runServer(port string, stopChan chan interface{}, logger *logrus.Logger) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	go func() {
		var ring C.struct_io_uring
		C.io_uring_queue_init(C.uint(QUEUE_DEPTH), &ring, 0)

		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warnf("error occurred when accept: %s", err.Error())
				continue
			}
			handleConnection(conn, &ring, logger)
		}
	}()

	go func() {
		<- stopChan
		listener.Close()
	}()
}

func handleConnection(conn net.Conn, ring *C.struct_io_uring, logger *logrus.Logger) {
	defer conn.Close()

	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		logger.Warnf("error orrured when get raw fd: %s",err.Error())
		return
	}
	defer file.Close()
	fd := file.Fd()

	buffer := make([]byte, BUFFER_SIZE)

	for {
		sqe := C.io_uring_get_sqe(ring)
		C.io_uring_prep_read(sqe, C.int(fd), unsafe.Pointer(&buffer[0]), C.uint(BUFFER_SIZE), 0)
		C.io_uring_submit(ring)

		var n int
		var cqe *C.struct_io_uring_cqe
		for {
			C.io_uring_wait_cqe(ring, &cqe)
			if cqe == nil {
				continue
			}
			if cqe.res < 0 {
				errno := syscall.Errno(-cqe.res)
				if errno == syscall.EAGAIN {
					C.io_uring_cqe_seen(ring, cqe)
					continue
				} else {
					logger.Warnf("error occurred when read: %s", errno.Error())
					C.io_uring_cqe_seen(ring, cqe)
					return
				}
			} else {
				n = int(cqe.res)
				break
			}
		}
		C.io_uring_cqe_seen(ring, cqe)

		if n == 0 {
			return
		}

		sqe = C.io_uring_get_sqe(ring)
		C.io_uring_prep_write(sqe, C.int(fd), unsafe.Pointer(&buffer[0]), C.uint(n), 0)
		C.io_uring_submit(ring)

		for {
			C.io_uring_wait_cqe(ring, &cqe)
			if cqe == nil {
				continue
			}
			if cqe.res < 0 {
				errno := syscall.Errno(-cqe.res)
				if errno == syscall.EAGAIN {
					C.io_uring_cqe_seen(ring, cqe)
					continue
				} else {
					logger.Warnf("error occurred when write: %s", errno.Error())
					C.io_uring_cqe_seen(ring, cqe)
					return
				}
			} else {
				break
			}
		}
		C.io_uring_cqe_seen(ring, cqe)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	port := ":8000"
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	stopchan := make(chan interface{})
	runServer(port, stopchan, logger)
	defer close(stopchan)

	m := 100000
	n := 100
	messageLength := 48

	start := time.Now()

	for i := 0; i < m; i++ {
		conn, err := net.Dial("tcp", port)
		if err != nil {
			fmt.Printf("failed to connect to server: %v\n", err)
			conn.Close()
			continue
		}

		for j := 0; j < n; j++ {
			message := randomString(messageLength) +  "\n"
			_, err = conn.Write([]byte(message))
			if err != nil {
				fmt.Printf("failed to send message: %v\n", err)
				break
			}

			response, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Printf("failed to read response: %v\n", err)
				break
			}

			if message != response {
				fmt.Printf("%v %v failed\n", i, j)
				fmt.Printf("expect: %s\n", message)
				fmt.Printf("actual: %s\n", response)
				break
			}
		}

		conn.Close()
	}

	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
    seconds := int(elapsed.Seconds()) % 60
	fmt.Printf("the total time for uring to execute %dk connections, with %d writes per connection, is: %d minutes %d seconds\n", m / 1000, n, minutes, seconds)
}
