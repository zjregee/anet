package echoserver

import (
	"os"
	"context"

	"github.com/zjregee/anet"
	"github.com/sirupsen/logrus"
)

func runServer(port string, stopChan chan interface{}) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.WarnLevel)
	anet.SetLogger(logger)

	listener, err := anet.CreateListener("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}
	eventLoop, err := anet.NewEventLoop(handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	eventLoop.ServeNonBlocking(listener)

	go func() {
		<- stopChan
		eventLoop.Shutdown()
	}()
}

func handleConnection(ctx context.Context, connection anet.Connection) error {
	reader, writer := connection.Reader(), connection.Writer()
	data, err := reader.ReadAll();
	if err != nil {
		return err
	}
	err = writer.WriteBytes(data, len(data));
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
