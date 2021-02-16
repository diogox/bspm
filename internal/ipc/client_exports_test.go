package ipc

import (
	"net"
)

var (
	SetupInitConn = func(newFunc func() (*net.UnixConn, error)) {
		initConn = newFunc
	}
	SetupWriteConn = func(newFunc func(conn *net.UnixConn, message Message) error) {
		writeConn = newFunc
	}
	SetupCloseConn = func(newFunc func(conn *net.UnixConn) error) {
		closeConn = newFunc
	}
)
