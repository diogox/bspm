package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/diogox/bspm/internal/grpc/bspm"
)

type Client bspm.BSPMClient

func NewClient() (Client, error) {
	timeout := 1 * time.Second

	conn, err := grpc.Dial(unixSocketPath,
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return bspm.NewBSPMClient(conn), nil
}
