package cgo

import (
	"testing"

	"github.com/ebitengine/purego"
)

func sum(n int64) int64 {
	var result int64
	result = 0
	for i := int64(0); i < n; i++ {
		result += i
	}
	return result
}

func BenchmarkGoSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = sum(100)
	}
}

func BenchmarkCgoSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = cgosum(100)
	}
}

func BenchmarkPuregoSum(b *testing.B) {
	libc, err := purego.Dlopen("./libsum.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic("shouldn't failed here")
	}
	defer purego.Dlclose(libc)
	var puregosum func(int64) int64
	purego.RegisterLibFunc(&puregosum, libc, "sum")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = puregosum(100)
	}
}
