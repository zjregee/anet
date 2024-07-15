package echoserver

import (
	"bufio"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjregee/anet"
)

func TestEchoServerSerial(t *testing.T) {
	port := ":8001"
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	m := 1000
	n := 100
	messageLength := 48

	for i := 0; i < m; i++ {
		conn, err := net.Dial("tcp", port)
		if err != nil {
			t.Fatalf("failed to connect to server: %v", err)
		}

		for j := 0; j < n; j++ {
			message := anet.GetRandomString(messageLength-1) + "\n"
			_, err = conn.Write([]byte(message))
			if err != nil {
				t.Fatalf("failed to send message: %v", err)
			}

			response, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				t.Fatalf("failed to read response: %v", err)
			}

			require.Equal(t, message, response)
		}

		conn.Close()
	}
}

func TestEchoServerConcurrent(t *testing.T) {
	port := ":8002"
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	c := 12
	m := 1000
	n := 100
	messageLength := 48

	errChan := make(chan error)

	var wg sync.WaitGroup
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < m; j++ {
				conn, err := net.Dial("tcp", port)
				if err != nil {
					select {
					case errChan <- errors.New("failed to connect to server"):
					default:
					}
					return
				}

				for k := 0; k < n; k++ {
					message := anet.GetRandomString(messageLength-1) + "\n"
					_, err = conn.Write([]byte(message))
					if err != nil {
						select {
						case errChan <- errors.New("failed to send message"):
						default:
						}
						return
					}

					response, err := bufio.NewReader(conn).ReadString('\n')
					if err != nil {
						select {
						case errChan <- errors.New("failed to read response"):
						default:
						}
						return
					}

					require.Equal(t, message, response)
				}

				conn.Close()
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	err := <-errChan
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}
}
