package socks

// https://tools.ietf.org/html/rfc1928

import (
	"context"
	"fmt"
	"io"
	"net"
)

func (p *Proxy) handle(conn net.Conn) {
	defer conn.Close()
	defer func() {
		p.Log.Debug("client connection closed")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Log.Debugf("got connection from %s", conn.RemoteAddr().String())

	if err := p.socks(ctx, conn); err != nil {
		// send error reply
		p.Log.Debugf("socks error: %v", err.Err)
	}
}

func (p *Proxy) socks(ctx context.Context, conn net.Conn) *Error {
	defer conn.Close()
	defer func() {
		if err := p.Proxyhandler.Close(); err != nil {
			p.Log.Errorf("error on cleanup: %v", err)
		}
	}()

	if err := p.handleConnect(ctx, conn); err != nil {
		return err
	}

	request, err := p.handleRequest(ctx, conn)
	if err != nil {
		return err
	}

	// Should we assume connection succeed here?
	remote, err := p.Proxyhandler.Init(conn.RemoteAddr(), *request)
	if err != nil {
		p.socksErrorReply(ctx, conn, err.Reason)
		p.Log.Warnf("Connecting to %s failed: %v", request.GetDestinationString(), err)
		return err
	}
	defer remote.Close()
	p.Log.Infof("Connection established %s - %s", conn.RemoteAddr().String(), request.GetDestinationString())

	err = p.handleRequestReply(ctx, conn, request)
	if err != nil {
		p.socksErrorReply(ctx, conn, err.Reason)
		return err
	}

	p.Log.Debug("beginning of data copy")

	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer conn.Close()
		defer remote.Close()
		for {
			if err := p.Proxyhandler.ReadFromClient(ctx2, conn, remote); err != nil {
				break
			}
		}
	}()

	for {
		if err := p.Proxyhandler.ReadFromRemote(ctx2, remote, conn); err != nil {
			break
		}
	}

	p.Log.Infof("Connection closed %s - %s", conn.RemoteAddr().String(), request.GetDestinationString())

	return nil
}

func (p *Proxy) socksErrorReply(ctx context.Context, conn io.ReadWriteCloser, reason RequestReplyReason) error {
	// send error reply
	repl, err := requestReply(nil, reason)
	if err != nil {
		return err
	}
	err = connectionWrite(ctx, conn, repl, p.Timeout)
	if err != nil {
		return err
	}

	return nil
}

func (p *Proxy) handleConnect(ctx context.Context, conn io.ReadWriteCloser) *Error {
	buf, err := connectionRead(ctx, conn, p.Timeout)
	if err != nil {
		return &Error{Reason: RequestReplyConnectionRefused, Err: err}
	}
	header, err := parseHeader(buf)
	if err != nil {
		return &Error{Reason: RequestReplyConnectionRefused, Err: err}
	}
	switch header.Version {
	case Version4:
		return &Error{Reason: RequestReplyCommandNotSupported, Err: fmt.Errorf("socks4 not yet implemented")}
	case Version5:
	default:
		return &Error{Reason: RequestReplyCommandNotSupported, Err: fmt.Errorf("version %#x not yet implemented", byte(header.Version))}
	}

	methodSupported := false
	for _, x := range header.Methods {
		if x == MethodNoAuthRequired {
			methodSupported = true
			break
		}
	}
	if !methodSupported {
		return &Error{Reason: RequestReplyMethodNotSupported, Err: fmt.Errorf("we currently only support no authentication")}
	}
	reply := make([]byte, 2)
	reply[0] = byte(Version5)
	reply[1] = byte(MethodNoAuthRequired)
	err = connectionWrite(ctx, conn, reply, p.Timeout)
	if err != nil {
		return &Error{Reason: RequestReplyGeneralFailure, Err: fmt.Errorf("could not send connect reply: %w", err)}
	}
	return nil
}

func (p *Proxy) handleRequest(ctx context.Context, conn io.ReadWriteCloser) (*Request, *Error) {
	buf, err := connectionRead(ctx, conn, p.Timeout)
	if err != nil {
		return nil, &Error{Reason: RequestReplyGeneralFailure, Err: fmt.Errorf("error on ConnectionRead: %w", err)}
	}
	request, err2 := parseRequest(buf)
	if err2 != nil {
		return nil, err2
	}
	return request, nil
}

func (p *Proxy) handleRequestReply(ctx context.Context, conn io.ReadWriteCloser, request *Request) *Error {
	repl, err := requestReply(request, RequestReplySucceeded)
	if err != nil {
		return &Error{Reason: RequestReplyGeneralFailure, Err: fmt.Errorf("error on requestReply: %w", err)}
	}
	err = connectionWrite(ctx, conn, repl, p.Timeout)
	if err != nil {
		return &Error{Reason: RequestReplyGeneralFailure, Err: fmt.Errorf("error on RequestResponse: %w", err)}
	}

	return nil
}
