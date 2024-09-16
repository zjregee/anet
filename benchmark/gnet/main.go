package main

import (
	"flag"
	"fmt"

	"github.com/panjf2000/gnet/v2"
)

type echoServer struct {
	eng gnet.Engine

	gnet.BuiltinEventEngine
}

func (es *echoServer) OnBoot(eng gnet.Engine) gnet.Action {
	es.eng = eng
	return gnet.None
}

func (es *echoServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	_, _ = c.Write(buf)
	return gnet.None
}

func main() {
	var port string
	flag.StringVar(&port, "port", ":8000", "")
	flag.Parse()

	echo := &echoServer{}
	_ = gnet.Run(echo, fmt.Sprintf("tcp://%s", port))
}
