package tcpserver

import (
	"context"

	"github.com/zjregee/anet"
)

func runServer(port string, stopChan chan interface{}) {
	listener, err := anet.CreateListener("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	eventLoop, err := anet.NewEventLoop(handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	go func() {
		_ = eventLoop.Serve(listener)
	}()

	go func() {
		<-stopChan
		_ = eventLoop.Shutdown(context.Background())
		listener.Close()
	}()
}

func handleConnection(_ context.Context, connection anet.Connection) error {
	reader, writer := connection.Reader(), connection.Writer()

	for {
		data, err := reader.ReadUtil('\n')
		if err != nil {
			return err
		}
		err = writer.WriteBytes(data, len(data))
		if err != nil {
			return err
		}
		err = writer.Flush()
		if err != nil {
			return err
		}
	}
}
