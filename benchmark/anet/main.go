package main

import (
	"context"
	"flag"

	"github.com/zjregee/anet"
)

func handleConnection(_ context.Context, connection anet.Connection) error {
	reader, writer := connection.Reader(), connection.Writer()

	for {
		data, err := reader.ReadUtil('\n')
		if err != nil {
			return err
		}
		reader.Release()
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

func main() {
	var port string
	flag.StringVar(&port, "port", ":8000", "")
	flag.Parse()

	listener, err := anet.CreateListener("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	eventLoop, err := anet.NewEventLoop(handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	_ = eventLoop.Serve(listener)
}
