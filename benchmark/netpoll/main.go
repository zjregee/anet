package main

import (
	"context"
	"flag"

	"github.com/cloudwego/netpoll"
)

func handleConnection(ctx context.Context, conn netpoll.Connection) error {
	defer conn.Close()

	reader := conn.Reader()
	writer := conn.Writer()

	for {
		message, err := reader.Until('\n')
		if err != nil {
			return err
		}
		_, err = writer.WriteBinary(message)
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

	listener, err := netpoll.CreateListener("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	eventLoop, err := netpoll.NewEventLoop(handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	_ = eventLoop.Serve(listener)
}
