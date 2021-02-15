package state_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/diogox/bspc-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspm/internal/feature/state"
)

func TestTransparentMonocle_Get(t *testing.T) {
	t.Run("should get stored state", func(t *testing.T) {
		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			st             = state.TransparentMonocleState{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
		)

		sm := sync.Map{}
		sm.Store(desktopID, st)

		got, err := state.NewTransparentMonocle().WithMap(&sm).Get(desktopID)
		require.NoError(t, err)

		assert.Equal(t, st, got)
	})
	t.Run("should return error when", func(t *testing.T) {
		t.Run("desktop id not found", func(t *testing.T) {
			const nonExistentDesktopID = bspc.ID(1)

			_, err := state.NewTransparentMonocle().Get(nonExistentDesktopID)
			require.Error(t, err)

			assert.True(t, errors.Is(err, state.ErrNotFound))
		})
		t.Run("saved state has invalid type", func(t *testing.T) {
			const desktopID = bspc.ID(1)

			sm := sync.Map{}
			sm.Store(desktopID, "invalid type")

			_, err := state.NewTransparentMonocle().WithMap(&sm).Get(desktopID)
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
			st             = state.TransparentMonocleState{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
		)

		sm := sync.Map{}
		state.NewTransparentMonocle().WithMap(&sm).Set(desktopID, st)

		got, ok := sm.Load(desktopID)
		require.True(t, ok)

		assert.Equal(t, st, got)
	})
}

func TestTransparentMonocle_Delete(t *testing.T) {
	t.Run("should delete state", func(t *testing.T) {
		var (
			desktopID      = bspc.ID(1)
			selectedNodeID = bspc.ID(2)
			st             = state.TransparentMonocleState{
				SelectedNodeID: &selectedNodeID,
				HiddenNodeIDs:  []bspc.ID{3},
			}
		)

		sm := sync.Map{}
		sm.Store(desktopID, st)

		got, err := state.NewTransparentMonocle().WithMap(&sm).Delete(desktopID)
		require.NoError(t, err)

		assert.Equal(t, st, got)
	})
	t.Run("should return error when", func(t *testing.T) {
		t.Run("desktop id not found", func(t *testing.T) {
			const nonExistentDesktopID = bspc.ID(1)

			_, err := state.NewTransparentMonocle().Delete(nonExistentDesktopID)
			require.Error(t, err)

			assert.True(t, errors.Is(err, state.ErrNotFound))
		})
		t.Run("saved state has invalid type", func(t *testing.T) {
			const desktopID = bspc.ID(1)

			sm := sync.Map{}
			sm.Store(desktopID, "invalid type")

			_, err := state.NewTransparentMonocle().WithMap(&sm).Delete(desktopID)
			require.Error(t, err)

			assert.Contains(t, "invalid state type", err.Error())
		})
	})
}
