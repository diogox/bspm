package ipc

import "net"

var (
	SetupInitListener = func(newFunc func() (*net.UnixListener, error)) {
		initListener = newFunc
	}
	SetupAcceptConn = func(newFunc func(*net.UnixListener) (*net.UnixConn, error)) {
		acceptConn = newFunc
	}
	SetupReadConn = func(newFunc func(*net.UnixConn, *[]byte) error) {
		readConn = newFunc
	}
	SetupCloseListener = func(newFunc func(*net.UnixListener) error) {
		closeListener = newFunc
	}
)
