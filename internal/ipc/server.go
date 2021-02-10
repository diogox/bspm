package ipc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

const serverUnixSocket = "/tmp/bspm.socket"

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
	addr, err := net.ResolveUnixAddr("unixgram", serverUnixSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve unix address: %v", err)
	}

	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %v", err)
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
			conn, err := s.listener.AcceptUnix()
			if err != nil {
				errCh <- err
				continue
			}

			var msg []byte
			for buffer := make([]byte, maxBufferSize); ; buffer = make([]byte, maxBufferSize) {
				// TODO: Use Monkey patching to test this?
				if _, _, err := conn.ReadFromUnix(buffer); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					errCh <- fmt.Errorf("failed to receive response: %v", err)
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
	return s.listener.Close()
}
