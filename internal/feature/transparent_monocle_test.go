package feature_test

import (
	"testing"

	"github.com/diogox/bspc-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/diogox/bspm/internal/bspwm"
	bspwmevent "github.com/diogox/bspm/internal/bspwm/event"
	"github.com/diogox/bspm/internal/feature"
	"github.com/diogox/bspm/internal/feature/state"
	"github.com/diogox/bspm/internal/log"
)

// TODO: Finish testing the other methods.

func TestNewTransparentMonocle(t *testing.T) {
	t.Run("should return a transparent monocle workflow manager successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var (
			mockEventManager = bspwmevent.NewMockManager(ctrl)
			mockService      = bspwm.NewMockService(ctrl)
			mockState        = state.NewMockTransparentMonocle(ctrl)
		)

		mockService.EXPECT().
			Events().
			Return(mockEventManager).
			Times(5)
		mockEventManager.EXPECT().
			On(bspc.EventTypeNodeAdd, gomock.Any())
		mockEventManager.EXPECT().
			On(bspc.EventTypeNodeRemove, gomock.Any())
		mockEventManager.EXPECT().
			On(bspc.EventTypeNodeTransfer, gomock.Any())
		mockEventManager.EXPECT().
			On(bspc.EventTypeNodeSwap, gomock.Any())
		mockEventManager.EXPECT().
			Start().
			Return(nil, nil)

		logger, err := log.New(zaptest.NewLogger(t), false)
		require.NoError(t, err)

		_, _, err = feature.StartTransparentMonocle(logger, mockState, mockService)
		assert.NoError(t, err)
	})
}
