package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/cloudwego/netpoll"
)

func runServer(port string, stopChan chan interface{}) {
	listener, err := netpoll.CreateListener("tcp", port)
	if err != nil {
		panic("can't failed here")
	}
	eventLoop, err := netpoll.NewEventLoop(handleConnection)
	if err != nil {
		panic("can't failed here")
	}
	go eventLoop.Serve(listener)

	go func() {
		<- stopChan
		eventLoop.Shutdown(context.Background())
		listener.Close()
	}()
}

func handleConnection(ctx context.Context, conn netpoll.Connection) error {
	defer conn.Close()

	reader := conn.Reader()
	writer := conn.Writer()

	for {
		message, err := reader.Until('\n')
		if err != nil {
			return err
		}
		_, err = writer.WriteBinary(message)
		if err != nil {
			return err
		}
		err = writer.Flush()
		if err != nil {
			return err
		}
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
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	m := 10000
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
	fmt.Printf("the total time for netpoll to execute %dk connections using 1 goroutines, with %d writes per connection, is: %d minutes and %d seconds\n", m / 1000, n, minutes, seconds)
}
