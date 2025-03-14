package socks

import (
	"context"
	"io"
	"net"
	"time"
)

// ProxyHandler is the interface for handling the proxy requests
type ProxyHandler interface {
	Init(net.Addr, Request) (io.ReadWriteCloser, *Error)
	ReadFromClient(context.Context, io.ReadCloser, io.WriteCloser) error
	ReadFromRemote(context.Context, io.ReadCloser, io.WriteCloser) error
	Close() error
}

// Proxy is the main struct
type Proxy struct {
	ServerAddr   string
	Done         chan struct{}
	Proxyhandler ProxyHandler
	Timeout      time.Duration
	Log          Logger
}

func NewSimpleProxy(addr string, handler ProxyHandler) *Proxy {
	return &Proxy{
		ServerAddr:   addr,
		Proxyhandler: handler,
		Timeout:      time.Second * 60,
		Log:          nil,
		Done:         nil,
	}
}

func NewProxy(addr string, handler ProxyHandler, done chan struct{}, timeout time.Duration, log Logger) *Proxy {
	return &Proxy{
		ServerAddr:   addr,
		Done:         done,
		Proxyhandler: handler,
		Timeout:      timeout,
		Log:          log,
	}
}

// Start is the main function to start a proxy
func (p *Proxy) Start() error {
	if p.Log == nil {
		p.Log = &NilLogger{} // allow not to set logger
	}

	listener, err := net.Listen("tcp", p.ServerAddr)
	if err != nil {
		return err
	}
	go p.run(listener)
	return nil
}

func (p *Proxy) run(listener net.Listener) {
	for {
		select {
		case <-p.Done:
			return
		default:
			connection, err := listener.Accept()
			if err == nil {
				go p.handle(connection)
			} else {
				p.Log.Errorf("Error accepting conn: %v", err)
			}
		}
	}
}

// Stop stops the proxy
func (p *Proxy) Stop() {
	p.Log.Warn("Stopping proxy")
	if p.Done == nil {
		return
	}
	close(p.Done)
	p.Done = nil
}
