package bspwmdesktop_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspc-go"
	"github.com/diogox/bspc-go/bspctest"

	bspwmdesktop "github.com/diogox/bspm/internal/bspwm/desktop"
	"github.com/diogox/bspm/internal/bspwm/filter"
)

func TestService_Get(t *testing.T) {
	buildQuery := func(filter interface{}) string {
		return fmt.Sprintf("query -d %s -T", filter)
	}

	t.Run("should return focused node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		want := bspc.Desktop{
			Name: "desktop-name",
			ID:   bspc.ID(1),
		}

		mockClient := bspwmdesktop.NewMockClient(ctrl)
		mockClient.EXPECT().
			Query(buildQuery(filter.DesktopFocused), bspctest.QueryResponse(t, want)).
			Return(nil)

		s := bspwmdesktop.NewService(mockClient)

		got, err := s.Get(filter.DesktopFocused)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
	t.Run("should return error when bspc returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("error")

		mockClient := bspwmdesktop.NewMockClient(ctrl)
		mockClient.EXPECT().
			Query(gomock.Any(), gomock.Any()).
			Return(expectedErr)

		s := bspwmdesktop.NewService(mockClient)

		_, err := s.Get(filter.DesktopFocused)
		require.Error(t, err)

		assert.True(t, errors.Is(err, expectedErr))
	})
}
