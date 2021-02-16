package ipc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

const serverUnixSocket = "/tmp/bspm.socket"

var (
	initListener = func() (*net.UnixListener, error) {
		addr, err := net.ResolveUnixAddr("unixgram", serverUnixSocket)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve unix address: %v", err)
		}

		l, err := net.ListenUnix("unix", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to socket: %v", err)
		}

		return l, nil
	}
	acceptConn = func(l *net.UnixListener) (*net.UnixConn, error) {
		return l.AcceptUnix()
	}
	readConn = func(conn *net.UnixConn, buffer *[]byte) error {
		_, _, err := conn.ReadFromUnix(*buffer)
		return err
	}
	closeListener = func(l *net.UnixListener) error {
		return l.Close()
	}
)

type (
	Server interface {
		Listen() (chan Message, chan error)
		Close() error
	}

	server struct {
		listener *net.UnixListener
	}
)

func NewServer() (Server, error) {
	l, err := initListener()
	if err != nil {
		return nil, err
	}

	return server{
		listener: l,
	}, nil
}

func (s server) Listen() (chan Message, chan error) {
	var (
		msgCh = make(chan Message)
		errCh = make(chan error)
	)

	go func() {
		for {
			conn, err := acceptConn(s.listener)
			if err != nil {
				errCh <- err
				continue
			}

			var msg []byte
			for buffer := make([]byte, maxBufferSize); ; buffer = make([]byte, maxBufferSize) {
				if err := readConn(conn, &buffer); err != nil {
					if errors.Is(err, io.EOF) {
						msg = append(msg, buffer...)
						break
					}

					errCh <- fmt.Errorf("failed to receive response: %w", err)
					break
				}

				msg = append(msg, buffer...)
			}

			msgCh <- Message(bytes.Trim(msg, "\x00"))
		}
	}()

	return msgCh, errCh
}

func (s server) Close() error {
	return closeListener(s.listener)
}
