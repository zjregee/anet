package httpserver

import (
	"net/http"

	"github.com/zjregee/anet/ahttp"
)

func runServer(port string, stopChan chan interface{}) {
	server := ahttp.New()
	server.Add("/test", func(c *ahttp.Context) error {
		return c.NoContent(http.StatusOK)
	})

	go func() {
		_ = server.Start(port)
	}()

	go func() {
		<-stopChan
		_ = server.Shutdown()
	}()
}
