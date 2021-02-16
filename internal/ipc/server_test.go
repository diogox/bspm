package ipc_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspm/internal/ipc"
)

func TestNewServer(t *testing.T) {
	t.Run("should fail when initiating the listener returns an error", func(t *testing.T) {
		expectedErr := errors.New("error")

		ipc.SetupInitListener(func() (*net.UnixListener, error) {
			return nil, expectedErr
		})

		_, err := ipc.NewServer()
		require.Error(t, err)
		assert.True(t, errors.Is(err, expectedErr))
	})
}

func TestServer_Listen(t *testing.T) {
	t.Run("should start listening successfully", func(t *testing.T) {
		const expectedMsg = ipc.Message("ipc message")

		ipc.SetupInitListener(func() (*net.UnixListener, error) {
			return &net.UnixListener{}, nil
		})
		ipc.SetupAcceptConn(func(*net.UnixListener) (*net.UnixConn, error) {
			return &net.UnixConn{}, nil
		})
		ipc.SetupReadConn(func(_ *net.UnixConn, buffer *[]byte) error {
			*buffer = []byte(expectedMsg)
			return io.EOF
		})

		c, err := ipc.NewServer()
		require.NoError(t, err)

		msgCh, errCh := c.Listen()
		require.NoError(t, err)

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*1))
		defer cancel()

		select {
		case msg := <-msgCh:
			assert.Equal(t, expectedMsg, msg)
		case <-errCh:
			t.Fail()
		case <-ctx.Done():
			t.Fail()
		}
	})
	t.Run("should return error when", func(t *testing.T) {
		t.Run("accepting a connection returns an error", func(t *testing.T) {
			expectedErr := errors.New("error")

			ipc.SetupInitListener(func() (*net.UnixListener, error) {
				return &net.UnixListener{}, nil
			})
			ipc.SetupAcceptConn(func(*net.UnixListener) (*net.UnixConn, error) {
				return nil, expectedErr
			})

			c, err := ipc.NewServer()
			require.NoError(t, err)

			msgCh, errCh := c.Listen()
			require.NoError(t, err)

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*1))
			defer cancel()

			select {
			case <-msgCh:
				t.Fail()
			case err := <-errCh:
				assert.True(t, errors.Is(err, expectedErr))
			case <-ctx.Done():
				t.Fail()
			}
		})
		t.Run("reading returns an error", func(t *testing.T) {
			expectedErr := errors.New("error")

			ipc.SetupInitListener(func() (*net.UnixListener, error) {
				return &net.UnixListener{}, nil
			})
			ipc.SetupAcceptConn(func(*net.UnixListener) (*net.UnixConn, error) {
				return &net.UnixConn{}, nil
			})
			ipc.SetupReadConn(func(_ *net.UnixConn, buffer *[]byte) error {
				return expectedErr
			})

			c, err := ipc.NewServer()
			require.NoError(t, err)

			msgCh, errCh := c.Listen()
			require.NoError(t, err)

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*1))
			defer cancel()

			select {
			case <-msgCh:
				t.Fail()
			case err := <-errCh:
				assert.True(t, errors.Is(err, expectedErr))
			case <-ctx.Done():
				t.Fail()
			}
		})
	})
}

func TestServer_Close(t *testing.T) {
	t.Run("should successfully close the listener", func(t *testing.T) {
		ipc.SetupInitListener(func() (*net.UnixListener, error) {
			return &net.UnixListener{}, nil
		})
		ipc.SetupCloseListener(func(*net.UnixListener) error {
			return nil
		})

		c, err := ipc.NewServer()
		require.NoError(t, err)

		err = c.Close()
		assert.NoError(t, err)
	})
	t.Run("should fail when closing the listener returns an error", func(t *testing.T) {
		expectedErr := errors.New("error")

		ipc.SetupInitListener(func() (*net.UnixListener, error) {
			return &net.UnixListener{}, nil
		})
		ipc.SetupCloseListener(func(*net.UnixListener) error {
			return expectedErr
		})

		c, err := ipc.NewServer()
		require.NoError(t, err)

		err = c.Close()
		require.Error(t, err)

		assert.True(t, errors.Is(err, expectedErr))
	})
}
