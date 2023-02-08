package main

import (
	"time"

	socks "github.com/wikylyu/gosocks"
)

func main() {
	handler := socks.DefaultHandler{
		Timeout: 1 * time.Second,
	}
	listen := "127.0.0.1:8181"
	p := socks.Proxy{
		ServerAddr:   listen,
		Proxyhandler: handler,
		Timeout:      1 * time.Second,
	}
	if err := p.Start(); err != nil {
		panic(err)
	}
	<-p.Done
}
