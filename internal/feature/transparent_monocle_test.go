package feature_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/diogox/bspc-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/diogox/bspm/internal/feature"
	"github.com/diogox/bspm/internal/feature/state"
	"github.com/diogox/bspm/internal/log"
)

// TODO: Finish testing the other methods.

func TestNewTransparentMonocle(t *testing.T) {
	// TODO: Set up some kind of deadline for the tests below. Failed tests will hang when a `continue` statement is reached.

	// TODO: Complete this: Each of the below has untested paths.

	t.Run("should return a transparent monocle workflow manager and", func(t *testing.T) {
		t.Run("handle added node events", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				desktopID      = bspc.ID(1)
				previousNodeID = bspc.ID(2)
				newNodeID      = bspc.ID(3)
			)

			var (
				evCh  = make(chan bspc.Event, 1)
				errCh = make(chan error)
			)

			var (
				notFloating   = bspc.StateTypeTiled
				queryResponse = bspc.Node{
					Client: bspc.NodeClient{
						State: notFloating,
					},
				}
			)

			var (
				mockClient = feature.NewMockClient(ctrl)
				mockState  = state.NewMockTransparentMonocle(ctrl)
			)

			gomock.InOrder(
				mockClient.EXPECT().
					SubscribeEvents(bspc.EventTypeNodeAdd, bspc.EventTypeNodeRemove).
					Return(evCh, errCh, nil),
				mockClient.EXPECT().
					Query(fmt.Sprintf("query -n %d -T", newNodeID), ResponseResolver(t, queryResponse)).
					Return(nil),
				mockState.EXPECT().
					Get(desktopID).
					Return(
						state.TransparentMonocleState{
							SelectedNodeID: &previousNodeID,
							HiddenNodeIDs:  nil,
						},
						nil,
					),
				mockClient.EXPECT().
					Query(fmt.Sprintf("node %d --flag hidden=on", previousNodeID), nil).
					Return(nil),
				mockState.EXPECT().
					Set(desktopID,
						state.TransparentMonocleState{
							SelectedNodeID: &newNodeID,
							HiddenNodeIDs:  []bspc.ID{previousNodeID},
						},
					),
			)

			logger, err := log.New(
				zaptest.NewLogger(
					t,
					zaptest.WrapOptions(
						zap.Hooks(
							func(e zapcore.Entry) error {
								require.NotEqual(t, zap.ErrorLevel, e.Level)
								return nil
							},
						),
					),
				),
				true,
			)
			require.NoError(t, err)

			_, cancel, err := feature.StartTransparentMonocle(logger, mockState, mockClient)
			require.NoError(t, err)

			const notUsedID = bspc.NilID

			evCh <- bspc.Event{
				Type: bspc.EventTypeNodeAdd,
				Payload: bspc.EventNodeAdd{
					MonitorID: notUsedID,
					DesktopID: desktopID,
					NodeID:    newNodeID,
					IPID:      notUsedID,
				},
			}

			time.Sleep(time.Millisecond * 100)
			cancel()
		})

		t.Run("handle node removal events", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				desktopID    = bspc.ID(1)
				nodeID       = bspc.ID(2)
				hiddenNodeID = bspc.ID(3)
			)

			var (
				evCh  = make(chan bspc.Event, 1)
				errCh = make(chan error)
			)

			var (
				mockClient = feature.NewMockClient(ctrl)
				mockState  = state.NewMockTransparentMonocle(ctrl)
			)

			gomock.InOrder(
				mockClient.EXPECT().
					SubscribeEvents(bspc.EventTypeNodeAdd, bspc.EventTypeNodeRemove).
					Return(evCh, errCh, nil),
				mockState.EXPECT().
					Get(desktopID).
					Return(
						state.TransparentMonocleState{
							SelectedNodeID: &nodeID,
							HiddenNodeIDs:  []bspc.ID{hiddenNodeID},
						},
						nil,
					),
				mockClient.EXPECT().
					Query(fmt.Sprintf("node %d --flag hidden=off", hiddenNodeID), nil).
					Return(nil),
				mockState.EXPECT().
					Set(desktopID,
						state.TransparentMonocleState{
							SelectedNodeID: &hiddenNodeID,
							HiddenNodeIDs:  []bspc.ID{},
						},
					),
			)

			logger, err := log.New(
				zaptest.NewLogger(
					t,
					zaptest.WrapOptions(
						zap.Hooks(
							func(e zapcore.Entry) error {
								require.NotEqual(t, zap.ErrorLevel, e.Level)
								return nil
							},
						),
					),
				),
				true,
			)
			require.NoError(t, err)

			_, cancel, err := feature.StartTransparentMonocle(logger, mockState, mockClient)
			require.NoError(t, err)

			const notUsedID = bspc.NilID

			evCh <- bspc.Event{
				Type: bspc.EventTypeNodeRemove,
				Payload: bspc.EventNodeRemove{
					MonitorID: notUsedID,
					DesktopID: desktopID,
					NodeID:    nodeID,
				},
			}

			time.Sleep(time.Millisecond * 100)
			cancel()
		})

		t.Run("should handle subscription errors", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				evCh  = make(chan bspc.Event)
				errCh = make(chan error)
			)

			var (
				mockClient = feature.NewMockClient(ctrl)
				mockState  = state.NewMockTransparentMonocle(ctrl)
			)

			mockClient.EXPECT().
				SubscribeEvents(bspc.EventTypeNodeAdd, bspc.EventTypeNodeRemove).
				Return(evCh, errCh, nil)

			zapLogger := zaptest.NewLogger(t)

			logger, err := log.New(zapLogger, true)
			require.NoError(t, err)

			_, cancel, err := feature.StartTransparentMonocle(logger, mockState, mockClient)
			require.NoError(t, err)

			zapLogger.WithOptions(
				zap.Hooks(
					func(e zapcore.Entry) error {
						assert.Equal(t, zap.ErrorLevel, e.Level)
						cancel()
						return nil
					},
				),
			)

			errCh <- errors.New("error")

			time.Sleep(time.Millisecond * 100)
			cancel()
		})
	})
}

type resResolverMatcher struct {
	t   *testing.T
	res interface{}
}

// TODO: Add the below to a `bspctest` package inside the `bspc-go` repo, akin to `zap`'s `zaptest`.
func ResponseResolver(t *testing.T, res interface{}) *resResolverMatcher {
	return &resResolverMatcher{
		t:   t,
		res: res,
	}
}

func (m *resResolverMatcher) String() string {
	bb, err := json.Marshal(m.res)
	require.NoError(m.t, err)

	return string(bb)
}

func (m *resResolverMatcher) Matches(x interface{}) bool {
	resolver, ok := x.(bspc.QueryResponseResolver)
	if !ok {
		return false
	}

	bb, err := json.Marshal(m.res)
	require.NoError(m.t, err)

	// Running this populates the variable passed into it by reference.
	err = resolver(bb)
	require.NoError(m.t, err)

	return true
}
