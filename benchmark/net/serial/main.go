package main

import (
	"os"
	"net"
	"fmt"
	"time"
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
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			logger.Warnf("error occurred when read: %s",err.Error())
			return
		}
		_, err = writer.WriteString(message)
		if err != nil {
			logger.Warnf("error occurred when write: %s",err.Error())
			return
		}
		err = writer.Flush()
		if err != nil {
			logger.Warnf("error occurred when flush: %s",err.Error())
			return
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
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.FatalLevel)
	stopchan := make(chan interface{})
	runServer(port, stopchan, logger)
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
	fmt.Printf("the total time for net to execute %dk connections using 1 goroutines, with %d writes per connection, is: %d minutes and %d seconds\n", m / 1000, n, minutes, seconds)
}
