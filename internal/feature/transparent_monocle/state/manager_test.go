package state_test

import (
	"testing"

	"github.com/diogox/bspc-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspm/internal/feature/transparent_monocle/state"
	"github.com/diogox/bspm/internal/feature/transparent_monocle/topic"
	"github.com/diogox/bspm/internal/subscription"
)

func TestTransparentMonocle_Get(t *testing.T) {
	t.Run("should get stored state", func(t *testing.T) {
		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			st             = state.State{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
			initial = map[bspc.ID]state.State{desktopID: st}
		)

		got, ok := state.NewTransparentMonocle(nil).WithState(initial).Get(desktopID)
		require.True(t, ok)

		assert.Equal(t, st, got)
	})
	t.Run("should return false when desktop id not found", func(t *testing.T) {
		const nonExistentDesktopID = bspc.ID(1)

		_, ok := state.NewTransparentMonocle(nil).Get(nonExistentDesktopID)
		assert.False(t, ok)
	})
}

func TestTransparentMonocle_Set(t *testing.T) {
	t.Run("should set state", func(t *testing.T) {
		t.Run("and publish 'monocle enabled' topic", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				desktopID      = bspc.ID(1)
				selectedNodeID = bspc.ID(2)
				st             = state.State{
					SelectedNodeID: &selectedNodeID,
					HiddenNodeIDs:  []bspc.ID{3},
				}
				initial = map[bspc.ID]state.State{}
			)

			mockSubscriptions := subscription.NewMockManager(ctrl)
			mockSubscriptions.EXPECT().Publish(topic.MonocleEnabled, st)

			state.NewTransparentMonocle(mockSubscriptions).WithState(initial).Set(desktopID, st)

			got, ok := initial[desktopID]
			require.True(t, ok)

			assert.Equal(t, st, got)
		})
		t.Run("and publish 'monocle state changed' topic", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				desktopID      = bspc.ID(1)
				selectedNodeID = bspc.ID(2)
				st             = state.State{
					SelectedNodeID: &selectedNodeID,
					HiddenNodeIDs:  []bspc.ID{3},
				}
				initial = map[bspc.ID]state.State{desktopID: st}
			)

			mockSubscriptions := subscription.NewMockManager(ctrl)
			mockSubscriptions.EXPECT().Publish(topic.MonocleStateChanged, st)

			state.NewTransparentMonocle(mockSubscriptions).WithState(initial).Set(desktopID, st)

			got, ok := initial[desktopID]
			require.True(t, ok)

			assert.Equal(t, st, got)
		})
	})
}

func TestTransparentMonocle_Delete(t *testing.T) {
	t.Run("should delete state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			st             = state.State{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
			initial = map[bspc.ID]state.State{desktopID: st}
		)

		mockSubscriptions := subscription.NewMockManager(ctrl)
		mockSubscriptions.EXPECT().Publish(topic.MonocleDisabled, st)

		state.NewTransparentMonocle(mockSubscriptions).WithState(initial).Delete(desktopID)
	})
}
