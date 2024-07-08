package main

import (
	"os"
	"fmt"
	"context"

	"github.com/zjregee/anet"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logFile := os.Getenv("ANET_RUNTIME_LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0644)
		if err != nil {
			panic("can't failed here")
		}
		defer file.Close()
		logger.SetOutput(file)
	}
	anet.SetLogger(logger)
	anet.SetLoggerLevel(logrus.InfoLevel)

	listener, err := anet.CreateListener("tcp", ":8000")
	if err != nil {
		panic("can't failed here")
	}
	eventLoop, err := anet.NewEventLoop(handler)
	if err != nil {
		panic("can't failed here")
	}
	_ = eventLoop.Serve(listener)
}

func handler(ctx context.Context, connection anet.Connection) error {
	reader, writer := connection.Reader(), connection.Writer()
	defer connection.Close()
	data, err := reader.ReadAll();
	if err != nil {
		fmt.Println("error occurred while reading connection: {}", err.Error())
	}
	err = writer.WriteBytes(data, len(data));
	if err != nil {
		fmt.Println("error occurred while writing connection: {}", err.Error())
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("error occurred while flushing connection: {}", err.Error())
	}
	return nil
}
