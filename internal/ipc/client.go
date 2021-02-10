package ipc

import (
	"fmt"
	"net"
)

type (
	Client interface {
		Send(msg Message) error
		Close() error
	}

	client struct {
		conn *net.UnixConn
	}
)

func NewClient() (Client, error) {
	addr, err := net.ResolveUnixAddr("unixgram", serverUnixSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve unix address: %v", err)
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %v", err)
	}

	return client{
		conn: conn,
	}, nil
}

func (c client) Send(msg Message) error {
	// TODO: Use monkey testing to unit test this?
	if _, err := c.conn.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

func (c client) Close() error {
	return c.conn.Close()
}
