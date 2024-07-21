package main

import (
	"flag"
	"net"
)

func handleConnection(conn net.Conn) {
	connection := connection{}
	connection.init(conn)
	connection.run()
}

func main() {
	var port string
	flag.StringVar(&port, "port", ":8000", "")
	flag.Parse()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic("shouldn't failed here")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}
