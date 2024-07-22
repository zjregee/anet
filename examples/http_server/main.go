package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/zjregee/anet"
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

	listener, err := anet.CreateListener("tcp", ":8080")
	if err != nil {
		panic("shouldn't failed here")
	}

	eventLoop, err := anet.NewEventLoop(handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	_ = eventLoop.Serve(listener)
}

func handleConnection(_ context.Context, connection anet.Connection) error {
	reader := bufio.NewReader(connection)
	writer := connection.Writer()

	for {
		request, err := http.ReadRequest(reader)
		if err != nil {
			return err
		}
		fmt.Printf("request: %s\n", request.URL)
		fmt.Printf("request: %s\n", request.Method)

		response := http.Response{
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Body:       nil,
		}
		response.Header.Set("Content-Type", "text/plain")

		err = response.Write(connection)
		if err != nil {
			return err
		}

		err = writer.Flush()
		if err != nil {
			return err
		}
	}
}
