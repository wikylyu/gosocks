package main

import (
	"context"
	"io"
	"net"
	"time"

	socks "github.com/wikylyu/gosocks"
)

func main() {
	handler := &MyCustomHandler{
		Timeout: 1 * time.Second,
		PropA:   "A",
		PropB:   "B",
	}
	p := socks.Proxy{
		ServerAddr:   "127.0.0.1:8282",
		Proxyhandler: handler,
		Timeout:      1 * time.Second,
	}
	if err := p.Start(); err != nil {
		panic(err)
	}
	<-p.Done
}

type MyCustomHandler struct {
	Timeout time.Duration
	PropA   string
	PropB   string
}

func (s *MyCustomHandler) Init(addr net.Addr, request socks.Request) (io.ReadWriteCloser, *socks.Error) {
	conn, err := net.DialTimeout("tcp", request.GetDestinationString(), s.Timeout)
	if err != nil {
		return nil, socks.NewError(socks.RequestReplyHostUnreachable, err)
	}
	return conn, nil
}

func (s *MyCustomHandler) ReadFromRemote(ctx context.Context, remote io.ReadCloser, client io.WriteCloser) error {
	_, err := io.Copy(client, remote)
	if err != nil {
		return err
	}
	return nil
}

func (s *MyCustomHandler) ReadFromClient(ctx context.Context, client io.ReadCloser, remote io.WriteCloser) error {
	_, err := io.Copy(remote, client)
	if err != nil {
		return err
	}
	return nil
}

func (s *MyCustomHandler) Close() error {
	return nil
}
