package echoserver

import (
	"net"
	"time"
	"bufio"
	"testing"
	"math/rand"

	"github.com/stretchr/testify/require"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func TestEchoServerSerial(t *testing.T) {
	port := ":8000"
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	n := 1
	m := 1
	messageLength := 48
	for i := 0; i < n ; i++ {
		conn, err := net.Dial("tcp", port)
		if err != nil {
			t.Fatalf("failed to connect to server: %v", err)
		}
	
		for j := 0; j < m; j++ {
			message := randomString(messageLength) +  "\n"
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
