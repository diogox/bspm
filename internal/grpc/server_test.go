package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	transparentmonocle "github.com/diogox/bspm/internal/feature/transparent_monocle"
	"github.com/diogox/bspm/internal/grpc"
	"github.com/diogox/bspm/internal/grpc/bspm"
	"github.com/diogox/bspm/internal/log"
)

func TestServer_MonocleModeToggle(t *testing.T) {
	t.Run("should toggle monocle mode", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := transparentmonocle.NewMockFeature(ctrl)
		mockService.EXPECT().
			ToggleCurrentDesktop().
			Return(nil)

		logger, err := log.New(zaptest.NewLogger(t), false)
		require.NoError(t, err)

		_, err = grpc.
			NewTestServer(logger, mockService).
			MonocleModeToggle(context.Background(), &empty.Empty{})
		assert.NoError(t, err)
	})
	t.Run("should return error when service returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("error")

		mockService := transparentmonocle.NewMockFeature(ctrl)
		mockService.EXPECT().
			ToggleCurrentDesktop().
			Return(expectedErr)

		logger, err := log.New(zaptest.NewLogger(t), false)
		require.NoError(t, err)

		_, err = grpc.
			NewTestServer(logger, mockService).
			MonocleModeToggle(context.Background(), &empty.Empty{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, expectedErr))
	})
}

func TestServer_MonocleModeCycle(t *testing.T) {
	t.Run("should cycle nodes monocle mode", func(t *testing.T) {
		t.Run("to next node", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := transparentmonocle.NewMockFeature(ctrl)
			mockService.EXPECT().
				FocusNextHiddenNode().
				Return(nil)

			logger, err := log.New(zaptest.NewLogger(t), false)
			require.NoError(t, err)

			_, err = grpc.
				NewTestServer(logger, mockService).
				MonocleModeCycle(context.Background(), &bspm.MonocleModeCycleRequest{
					CycleDirection: bspm.CycleDir_CYCLE_DIR_NEXT,
				})
			assert.NoError(t, err)
		})
		t.Run("to previous node", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := transparentmonocle.NewMockFeature(ctrl)
			mockService.EXPECT().
				FocusPreviousHiddenNode().
				Return(nil)

			logger, err := log.New(zaptest.NewLogger(t), false)
			require.NoError(t, err)

			_, err = grpc.
				NewTestServer(logger, mockService).
				MonocleModeCycle(context.Background(), &bspm.MonocleModeCycleRequest{
					CycleDirection: bspm.CycleDir_CYCLE_DIR_PREV,
				})
			assert.NoError(t, err)
		})
	})
	t.Run("should return error when service returns error", func(t *testing.T) {
		t.Run("when cycling to next node", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			expectedErr := errors.New("error")

			mockService := transparentmonocle.NewMockFeature(ctrl)
			mockService.EXPECT().
				FocusNextHiddenNode().
				Return(expectedErr)

			logger, err := log.New(zaptest.NewLogger(t), false)
			require.NoError(t, err)

			_, err = grpc.
				NewTestServer(logger, mockService).
				MonocleModeCycle(context.Background(), &bspm.MonocleModeCycleRequest{
					CycleDirection: bspm.CycleDir_CYCLE_DIR_NEXT,
				})
			require.Error(t, err)
			assert.True(t, errors.Is(err, expectedErr))
		})
		t.Run("when cycling to previous node", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			expectedErr := errors.New("error")

			mockService := transparentmonocle.NewMockFeature(ctrl)
			mockService.EXPECT().
				FocusPreviousHiddenNode().
				Return(expectedErr)

			logger, err := log.New(zaptest.NewLogger(t), false)
			require.NoError(t, err)

			_, err = grpc.
				NewTestServer(logger, mockService).
				MonocleModeCycle(context.Background(), &bspm.MonocleModeCycleRequest{
					CycleDirection: bspm.CycleDir_CYCLE_DIR_PREV,
				})
			require.Error(t, err)
			assert.True(t, errors.Is(err, expectedErr))
		})
	})
}

func TestServer_MonocleModeSubscribe(t *testing.T) {
	t.Run("should return subscription messages for monocle mode node count", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var (
			countCh = make(chan int, 1)
			count   = 2
		)

		var (
			mockService             = transparentmonocle.NewMockFeature(ctrl)
			mockGRPCSubscribeServer = grpc.NewMockBSPM_MonocleModeSubscribeServer(ctrl)
		)

		gomock.InOrder(
			mockService.EXPECT().
				SubscribeNodeCount().
				Return(countCh),
			mockGRPCSubscribeServer.EXPECT().
				Send(&bspm.MonocleModeSubscribeResponse{
					SubscriptionType: &bspm.MonocleModeSubscribeResponse_NodeCount{
						NodeCount: int32(count),
					},
				}).
				Do(func(interface{}) {
					// End test
					close(countCh)
				}).
				Return(nil),
		)

		countCh <- count

		logger, err := log.New(zaptest.NewLogger(t), false)
		require.NoError(t, err)

		err = grpc.
			NewTestServer(logger, mockService).
			MonocleModeSubscribe(&bspm.MonocleModeSubscribeRequest{
				Type: bspm.MonocleModeSubscriptionType_MONOCLE_MODE_SUBSCRIPTION_TYPE_NODE_COUNT,
			}, mockGRPCSubscribeServer)
		require.NoError(t, err)
	})
	t.Run("should return error when", func(t *testing.T) {
		t.Run("sending subscription message fails for monocle mode count nodes subscription", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				countCh     = make(chan int, 1)
				expectedErr = errors.New("error")
			)

			var (
				mockService             = transparentmonocle.NewMockFeature(ctrl)
				mockGRPCSubscribeServer = grpc.NewMockBSPM_MonocleModeSubscribeServer(ctrl)
			)

			gomock.InOrder(
				mockService.EXPECT().
					SubscribeNodeCount().
					Return(countCh),
				mockGRPCSubscribeServer.EXPECT().
					Send(gomock.Any()).
					Do(func(interface{}) {
						// End test
						close(countCh)
					}).
					Return(expectedErr),
			)

			countCh <- 1

			logger, err := log.New(zaptest.NewLogger(t), false)
			require.NoError(t, err)

			err = grpc.
				NewTestServer(logger, mockService).
				MonocleModeSubscribe(&bspm.MonocleModeSubscribeRequest{
					Type: bspm.MonocleModeSubscriptionType_MONOCLE_MODE_SUBSCRIPTION_TYPE_NODE_COUNT,
				}, mockGRPCSubscribeServer)
			require.Error(t, err)
			assert.True(t, errors.Is(err, expectedErr))
		})
		t.Run("subscription type is invalid", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				mockService             = transparentmonocle.NewMockFeature(ctrl)
				mockGRPCSubscribeServer = grpc.NewMockBSPM_MonocleModeSubscribeServer(ctrl)
			)

			logger, err := log.New(zaptest.NewLogger(t), false)
			require.NoError(t, err)

			err = grpc.
				NewTestServer(logger, mockService).
				MonocleModeSubscribe(&bspm.MonocleModeSubscribeRequest{
					Type: bspm.MonocleModeSubscriptionType_MONOCLE_MODE_SUBSCRIPTION_TYPE_INVALID,
				}, mockGRPCSubscribeServer)
			assert.Error(t, err)
		})
	})
}
