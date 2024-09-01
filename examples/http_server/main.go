package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/zjregee/anet"
	"github.com/zjregee/anet/ahttp"
)

func main() {
	logger := logrus.New()
	logFile := os.Getenv("ANET_RUNTIME_LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic("shouldn't failed here")
		}
		defer file.Close()
		logger.SetOutput(file)
	}
	logger.SetLevel(logrus.InfoLevel)
	anet.SetLogger(logger)

	server := ahttp.New()
	server.GET("/test", func(c *ahttp.Context) error {
		return c.NoContent(http.StatusOK)
	})
	server.POST("/test/name/:name/age/:age", func(c *ahttp.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("name: %s, age: %s", c.Get("name"), c.Get("age")))
	})
	server.POST("/test", func(c *ahttp.Context) error {
		data := struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{
			Name: "bob",
			Age:  18,
		}
		return c.JSON(http.StatusOK, data)
	})
	_ = server.Start(":8080")
}
