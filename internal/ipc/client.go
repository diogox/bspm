package ipc

import (
	"fmt"
	"net"
)

var (
	initConn = func() (*net.UnixConn, error) {
		addr, err := net.ResolveUnixAddr("unixgram", serverUnixSocket)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve unix address: %v", err)
		}

		conn, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to socket: %v", err)
		}

		return conn, nil
	}
	writeConn = func(conn *net.UnixConn, msg Message) error {
		if _, err := conn.Write([]byte(msg)); err != nil {
			return err
		}

		return nil
	}
	closeConn = func(conn *net.UnixConn) error {
		return conn.Close()
	}
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
	conn, err := initConn()
	if err != nil {
		return nil, err
	}

	return client{
		conn: conn,
	}, nil
}

func (c client) Send(msg Message) error {
	if err := writeConn(c.conn, msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (c client) Close() error {
	return closeConn(c.conn)
}
