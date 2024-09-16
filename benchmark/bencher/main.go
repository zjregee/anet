package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/caio/go-tdigest/v4"
	"github.com/zjregee/anet"
)

func main() {
	var (
		c             int
		m             int
		messageLength int
		port          string
		name          string
	)

	flag.IntVar(&c, "c", 12, "")
	flag.IntVar(&m, "m", 1000000, "")
	flag.IntVar(&messageLength, "len", 1024, "")
	flag.StringVar(&port, "port", ":8000", "")
	flag.StringVar(&name, "name", "anet", "")
	flag.Parse()

	count := 0
	t, _ := tdigest.New()
	var mu sync.Mutex
	var wg sync.WaitGroup
	message := anet.GetRandomString(messageLength-1) + "\n"

	start := time.Now()
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", port)
			defer func() {
				err := conn.Close()
				if err != nil {
					fmt.Printf("failed to close connection: %v\n", err)
				}
			}()
			if err != nil {
				fmt.Printf("failed to connect to server: %v\n", err)
				return
			}
			for {
				start := time.Now()
				_, err = conn.Write([]byte(message))
				if err != nil {
					fmt.Printf("failed to send message: %v\n", err)
					return
				}
				response, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Printf("failed to read response: %v\n", err)
					return
				}
				elapsed := time.Since(start)
				if response != message {
					fmt.Printf("expect: %s\n", message)
					fmt.Printf("actual: %s\n", response)
					return
				}
				mu.Lock()
				count += 1
				_ = t.Add(float64(elapsed.Microseconds()))
				if count > m {
					mu.Unlock()
					return
				}
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	milliseconds := int(elapsed.Milliseconds() % 1000)
	fmt.Printf("the total time for %s to execute %dk connections using %d goroutines, with %d bytes per write, is: %d min %d sec %d ms\n", name, count/1000, c, messageLength, minutes, seconds, milliseconds)
	qps := float64(count) / elapsed.Seconds()
	fmt.Printf("qps: %f\n", qps)
	fmt.Printf("average latency: %f us\n", t.Quantile(0.5))
	fmt.Printf("99th percentile latency: %f us\n", t.Quantile(0.99))
}
