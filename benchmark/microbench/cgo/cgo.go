package cgo

/*
#cgo LDFLAGS: -L. -lsum
#include <stdint.h>
int64_t sum(int64_t n);
*/
import "C"

func cgosum(n int64) int64 {
	return int64(C.sum(C.int64_t(n)))
}
