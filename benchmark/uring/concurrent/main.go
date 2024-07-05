package main

import (
	"os"
	"net"
	"fmt"
	"time"
	"sync"
	"bufio"
	"math/rand"

	"github.com/sirupsen/logrus"
)

func runServer(port string, stopChan chan interface{}, logger *logrus.Logger) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warnf("error occurred when accept: %s", err.Error())
				continue
			}
			handleConnection(conn, logger)
		}
	}()

	go func() {
		<- stopChan
		listener.Close()
	}()
}

func handleConnection(conn net.Conn, logger *logrus.Logger) {
	connection := connection{}
	connection.init(conn, logger)
	go connection.run()
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

	c := 12
	m := 10000
	n := 100
	messageLength := 48

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < m; j++ {
				conn, err := net.Dial("tcp", port)
				if err != nil {
					fmt.Printf("failed to connect to server: %v\n", err)
					conn.Close()
					continue
				}

				for k := 0; k < n; k++ {
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
						fmt.Printf("%v %v %v failed\n", i, j, k)
						fmt.Printf("expect: %s\n", message)
						fmt.Printf("actual: %s\n", response)
						break
					}
				}

				conn.Close()
			}
		}()
	}
	wg.Wait()

	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
    seconds := int(elapsed.Seconds()) % 60
	fmt.Printf("the total time for uring to execute %dk connections, with %d writes per connection, is: %d minutes %d seconds\n", m / 1000, n, minutes, seconds)
}
