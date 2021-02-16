package ipc_test

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspm/internal/ipc"
)

func TestNewClient(t *testing.T) {
	t.Run("should fail when initiating the connection returns an error", func(t *testing.T) {
		expectedErr := errors.New("error")

		ipc.SetupInitConn(func() (*net.UnixConn, error) {
			return nil, expectedErr
		})

		_, err := ipc.NewClient()
		require.Error(t, err)
		assert.True(t, errors.Is(err, expectedErr))
	})
}

func TestClient_Send(t *testing.T) {
	t.Run("should send message successfully", func(t *testing.T) {
		const expectedMsg = ipc.Message("ipc message")

		ipc.SetupInitConn(func() (*net.UnixConn, error) {
			return &net.UnixConn{}, nil
		})
		ipc.SetupWriteConn(func(_ *net.UnixConn, message ipc.Message) error {
			assert.Equal(t, expectedMsg, message)
			return nil
		})

		c, err := ipc.NewClient()
		require.NoError(t, err)

		err = c.Send(expectedMsg)
		require.NoError(t, err)
	})
	t.Run("should return error when sending message fails", func(t *testing.T) {
		expectedErr := errors.New("error")

		ipc.SetupWriteConn(func(_ *net.UnixConn, message ipc.Message) error {
			return expectedErr
		})

		c, err := ipc.NewClient()
		require.NoError(t, err)

		err = c.Send("ipc message")
		require.Error(t, err)

		assert.True(t, errors.Is(err, expectedErr))
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("should close connection successfully", func(t *testing.T) {
		ipc.SetupCloseConn(func(*net.UnixConn) error {
			return nil
		})

		c, err := ipc.NewClient()
		require.NoError(t, err)

		err = c.Close()
		assert.NoError(t, err)
	})
	t.Run("should return error when closing the connection fails", func(t *testing.T) {
		expectedErr := errors.New("error")

		ipc.SetupCloseConn(func(*net.UnixConn) error {
			return expectedErr
		})

		c, err := ipc.NewClient()
		require.NoError(t, err)

		err = c.Close()
		require.Error(t, err)

		assert.True(t, errors.Is(err, expectedErr))
	})
}
