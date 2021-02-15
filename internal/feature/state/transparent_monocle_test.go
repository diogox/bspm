package state

import (
	"errors"
	"sync"
	"testing"

	"github.com/diogox/bspc-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransparentMonocle_Get(t *testing.T) {
	t.Run("should get stored state", func(t *testing.T) {
		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			state          = TransparentMonocleState{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
		)

		sm := sync.Map{}
		sm.Store(desktopID, state)

		got, err := transparentMonocle{desktops: &sm}.Get(desktopID)
		require.NoError(t, err)

		assert.Equal(t, state, got)
	})
	t.Run("should return error when", func(t *testing.T) {
		t.Run("desktop id not found", func(t *testing.T) {
			const nonExistentDesktopID = bspc.ID(1)

			_, err := NewTransparentMonocle().Get(nonExistentDesktopID)
			require.Error(t, err)

			assert.True(t, errors.Is(err, ErrNotFound))
		})
		t.Run("saved state has invalid type", func(t *testing.T) {
			const desktopID = bspc.ID(1)

			sm := sync.Map{}
			sm.Store(desktopID, "invalid type")

			_, err := transparentMonocle{desktops: &sm}.Get(desktopID)
			require.Error(t, err)
			assert.Contains(t, "invalid state type", err.Error())
		})
	})
}

func TestTransparentMonocle_Set(t *testing.T) {
	t.Run("should set state", func(t *testing.T) {
		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			state          = TransparentMonocleState{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
		)

		sm := sync.Map{}
		transparentMonocle{desktops: &sm}.Set(desktopID, state)

		got, ok := sm.Load(desktopID)
		require.True(t, ok)

		assert.Equal(t, state, got)
	})
}

func TestTransparentMonocle_Delete(t *testing.T) {
	t.Run("should delete state", func(t *testing.T) {
		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			state          = TransparentMonocleState{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
		)

		sm := sync.Map{}
		sm.Store(desktopID, state)

		got, err := transparentMonocle{desktops: &sm}.Delete(desktopID)
		require.NoError(t, err)

		assert.Equal(t, state, got)
	})
	t.Run("should return error when", func(t *testing.T) {
		t.Run("desktop id not found", func(t *testing.T) {
			const nonExistentDesktopID = bspc.ID(1)

			_, err := NewTransparentMonocle().Delete(nonExistentDesktopID)
			require.Error(t, err)

			assert.True(t, errors.Is(err, ErrNotFound))
		})
		t.Run("saved state has invalid type", func(t *testing.T) {
			const desktopID = bspc.ID(1)

			sm := sync.Map{}
			sm.Store(desktopID, "invalid type")

			_, err := transparentMonocle{desktops: &sm}.Delete(desktopID)
			require.Error(t, err)

			assert.Contains(t, "invalid state type", err.Error())
		})
	})
}
