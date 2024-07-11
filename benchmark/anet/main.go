package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"flag"
	"bufio"
	"context"

	"github.com/zjregee/anet"
	"github.com/zjregee/anet/benchmark/utils"
)

func runServer(port string, stopChan chan interface{}) {
	listener, err := anet.CreateListener("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	eventLoop, err := anet.NewEventLoop(handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	go eventLoop.Serve(listener)

	go func() {
		<- stopChan
		eventLoop.Shutdown(context.Background())
		listener.Close()
	}()
}

func handleConnection(_ context.Context, connection anet.Connection) error {
	reader, writer := connection.Reader(), connection.Writer()

	for {
		data, err := reader.ReadUtil(1);
		if err != nil {
			return err
		}
		err = writer.WriteBytes(data, len(data));
		if err != nil {
			return err
		}
		err = writer.Flush()
		if err != nil {
			return err
		}
	}
}

func main() {
	port := ":8000"
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	var (
		c int
		m int
		n int
		messageLength int
	)

	flag.IntVar(&c, "c", 12, "")
	flag.IntVar(&m, "m", 1000, "")
	flag.IntVar(&n, "n", 100, "")
	flag.IntVar(&messageLength, "len", 48, "")
    flag.Parse()

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < m; j++ {
				conn, err := net.Dial("tcp", port)
				if err != nil {
					fmt.Printf("failed to connect to server: %v\n", err)
					conn.Close()
					continue
				}

				for k := 0; k < n; k++ {
					message := utils.GetRandomString(messageLength - 1) +  "\n"
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
						fmt.Printf("%v %v %v failed\n", i, j, k)
						fmt.Printf("expect: %s\n", message)
						fmt.Printf("actual: %s\n", response)
						break
					}
				}

				conn.Close()
			}
		}(i)
	}
	wg.Wait()

	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
    seconds := int(elapsed.Seconds()) % 60
	milliseconds := int(elapsed.Milliseconds() % 1000)
	fmt.Printf("the total time for anet to execute %dk connections using %d goroutines, with %d writes per connection and %d bytes per write, is: %d min %d sec %d ms\n", c * m / 1000, c, n, messageLength, minutes, seconds, milliseconds)
}
