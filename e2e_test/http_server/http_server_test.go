package httpserver

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPServerSerial(t *testing.T) {
	port := ":8003"
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	m := 1000
	for i := 0; i < m; i++ {
		resp, err := http.Get("http://localhost" + port + "/test")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestHTTPServerConcurrent(t *testing.T) {
	port := ":8004"
	stopchan := make(chan interface{})
	runServer(port, stopchan)
	defer close(stopchan)

	c := 12
	m := 1000
	var wg sync.WaitGroup
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for i := 0; i < m; i++ {
				resp, err := http.Get("http://localhost" + port + "/test")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
			}
		}(i)
	}
	wg.Wait()
}
