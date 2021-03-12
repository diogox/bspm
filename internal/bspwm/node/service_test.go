package bspwmnode_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/diogox/bspc-go"
	"github.com/diogox/bspc-go/bspctest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspm/internal/bspwm/filter"
	bspwmnode "github.com/diogox/bspm/internal/bspwm/node"
)

func TestService_Get(t *testing.T) {
	const id = bspc.ID(3)

	t.Run("should get node by given filter", func(t *testing.T) {
		tt := []struct {
			name        string
			getFilter   filter.NodeFilter
			filterValue interface{}
		}{
			{
				name:        "with id filter",
				getFilter:   filter.NodeID(id),
				filterValue: id,
			},
			{
				name:        "with 'local biggest' filter",
				getFilter:   filter.NodeLocalBiggest,
				filterValue: filter.NodeLocalBiggest,
			},
			{
				name:        "with 'focused' filter",
				getFilter:   filter.NodeFocused,
				filterValue: filter.NodeFocused,
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				var (
					expectedCmd  = fmt.Sprintf("query -n %v -T", tc.filterValue)
					expectedNode = bspc.Node{
						ID:     id,
						Marked: true,
					}
				)

				mockClient := bspwmnode.NewMockClient(ctrl)
				mockClient.EXPECT().
					Query(expectedCmd, bspctest.QueryResponse(t, expectedNode)).
					Return(nil)

				got, err := bspwmnode.NewService(mockClient).Get(tc.getFilter)
				require.NoError(t, err)

				assert.Equal(t, expectedNode, got)
			})
		}
	})
	t.Run("should fail when bspc returns an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("error")

		mockClient := bspwmnode.NewMockClient(ctrl)
		mockClient.EXPECT().
			Query(gomock.Any(), gomock.Any()).
			Return(expectedErr)

		_, err := bspwmnode.NewService(mockClient).Get(filter.NodeFocused)
		require.Error(t, err)

		assert.True(t, errors.Is(err, expectedErr))
	})
}

func TestService_SetVisibility(t *testing.T) {
	t.Run("should set node visibility", func(t *testing.T) {
		tt := []struct {
			name      string
			isVisible bool
			cmdHidden string
		}{
			{
				name:      "to true",
				isVisible: true,
				cmdHidden: "off",
			},
			{
				name:      "to false",
				isVisible: false,
				cmdHidden: "on",
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				var (
					id          = bspc.ID(3)
					expectedCmd = fmt.Sprintf("node %d --flag hidden=%s", id, tc.cmdHidden)
				)

				mockClient := bspwmnode.NewMockClient(ctrl)
				mockClient.EXPECT().
					Query(expectedCmd, nil).
					Return(nil)

				err := bspwmnode.NewService(mockClient).SetVisibility(id, tc.isVisible)
				require.NoError(t, err)
			})
		}
	})

	t.Run("should fail when bspc returns an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("error")

		mockClient := bspwmnode.NewMockClient(ctrl)
		mockClient.EXPECT().
			Query(gomock.Any(), gomock.Any()).
			Return(expectedErr)

		err := bspwmnode.NewService(mockClient).SetVisibility(bspc.ID(3), true)
		require.Error(t, err)

		assert.True(t, errors.Is(err, expectedErr))
	})
}
