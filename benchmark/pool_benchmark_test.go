package benchmark

import (
	"sync"
	"testing"
	"time"
	"anet/internal/pool"
)

const (
	RunTimes           = 1e6
	PoolCapacity       = 5e6
	BenchParam         = 10
	DefaultExpiredTime = 10 * time.Second
)

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := pool.NewPool(PoolCapacity, pool.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func BenchmarkMultiPool(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := pool.NewMultiPool(10, PoolCapacity / 10, pool.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func BenchmarkGoroutinesThroughput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go demoFunc()
		}
	}
}

func BenchmarkPoolThroughput(b *testing.B) {
	p, _ := pool.NewPool(PoolCapacity, pool.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(demoFunc)
		}
	}
}

func BenchmarkMultiPoolThroughput(b *testing.B) {
	p, _ := pool.NewMultiPool(10, PoolCapacity / 10, pool.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(demoFunc)
		}
	}
}
