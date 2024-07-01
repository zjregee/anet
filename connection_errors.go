package anet

import (
	"fmt"
	"net"
	"syscall"
)

const (
	ErrConnClosed   = syscall.Errno(0x101)
	ErrReadTimeout  = syscall.Errno(0x102)
	ErrWriteTimeout = syscall.Errno(0x103)
	ErrEOF          = syscall.Errno(0x104)
)

const ErrnoMask = 0xFF

func Exception(err error, suffix string) error {
	errno, ok := err.(syscall.Errno)
	if !ok {
		if suffix == "" {
			return err
		}
		return fmt.Errorf("%s %s", err.Error(), suffix)
	}
	return &exception{
		errno: errno,
		suffix: suffix,
	}
}

var _ net.Error = &exception{}

type exception struct {
	errno  syscall.Errno
	suffix string
}

func (e *exception) Error() string {
	var s string
	if int(e.errno) & 0x100 != 0 {
		s = errnos[int(e.errno) & ErrnoMask]
	}
	if s == "" {
		s = e.errno.Error()
	}
	if e.suffix != "" {
		s += " "
		s += e.suffix
	}
	return s
}

func (e *exception) Timeout() bool {
	if e.errno == ErrReadTimeout || e.errno == ErrWriteTimeout {
		return true
	}
	return e.errno.Timeout()
}

func (e *exception) Temporary() bool {
	return e.errno.Temporary()
}

var errnos = [...]string{
	ErrnoMask & ErrConnClosed:   "connection has been closed",
	ErrnoMask & ErrReadTimeout:  "connection read timeout",
	ErrnoMask & ErrWriteTimeout: "connection write timeout",
	ErrnoMask & ErrEOF:          "EOF",
}
