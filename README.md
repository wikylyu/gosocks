# GOSOCKS

Basic golang implementation of a socks5 proxy. This implementation is currently not feature complete and only supports the `CONNECT` command and no authentication.

This implemention also defines some handlers you can use to implement your own protocol behind this proxy server. This can be useful if you come a across a protocol that can be abused for proxy functionality and build a socks5 proxy around it.

The SOCKS protocol is defined in [rfc1928](https://tools.ietf.org/html/rfc1928)

## Documentation

[https://pkg.go.dev/github.com/firefart/gosocks](https://pkg.go.dev/github.com/firefart/gosocks)

## Handler Interface

```golang
type ProxyHandler interface {
	Init(net.Addr, Request) (io.ReadWriteCloser, *Error)
	ReadFromClient(context.Context, io.ReadCloser, io.WriteCloser) error
	ReadFromRemote(context.Context, io.ReadCloser, io.WriteCloser) error
	Close() error
}
```

### Init

Init is called before the copy operations and it should return a connection to the target that is ready to receive data.

### ReadFromClient

ReadFromClient is the method that handles the data copy from the client (you) to the remote connection. You can see the `DefaultHandler` for a sample implementation.

### ReadFromRemote

ReadFromRemote is the method that handles the data copy from the remote connection to the client (you). You can see the `DefaultHandler` for a sample implementation.

### Close

Close is called after the request finishes or errors out. Client and remote connections are closed before this method is called.

## Usage

### Default Usage

```golang
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
```

### Usage with custom handlers

```golang
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

```
