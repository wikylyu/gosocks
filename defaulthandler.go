package socks

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// DefaultHandler is the default socks5 implementation
type DefaultHandler struct {
	// Timeout defines the connect timeout to the destination
	Timeout time.Duration
}

// PreHandler is the default socks5 implementation
func (s DefaultHandler) Init(addr net.Addr, request Request) (io.ReadWriteCloser, *Error) {
	target := fmt.Sprintf("%s:%d", request.DestinationAddress, request.DestinationPort)
	remote, err := net.DialTimeout("tcp", target, s.Timeout)
	if err != nil {
		return nil, NewError(RequestReplyNetworkUnreachable, err)
	}
	return remote, nil
}

// CopyFromClientToRemote is the default socks5 implementation
func (s DefaultHandler) ReadFromClient(ctx context.Context, client io.ReadCloser, remote io.WriteCloser) error {
	if _, err := io.Copy(remote, client); err != nil {
		return err
	}
	return nil
}

// CopyFromRemoteToClient is the default socks5 implementation
func (s DefaultHandler) ReadFromRemote(ctx context.Context, remote io.ReadCloser, client io.WriteCloser) error {
	if _, err := io.Copy(client, remote); err != nil {
		return err
	}
	return nil
}

// Cleanup is the default socks5 implementation
func (s DefaultHandler) Close() error {
	return nil
}
