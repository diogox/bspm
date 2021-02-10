//go:generate mockgen -package feature -destination ./bspc_mock.go github.com/diogox/bspc-go Client
//go:generate mockgen -package feature -destination ./transparent_monocle_mock.go -self_package github.com/diogox/bspm/internal/feature github.com/diogox/bspm/internal/feature TransparentMonocle

package feature

import (
	"errors"
	"fmt"

	"github.com/diogox/bspc-go"
	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/feature/state"
	"github.com/diogox/bspm/internal/log"
)

type (
	TransparentMonocle interface {
		ToggleCurrentDesktop() error
		FocusPreviousHiddenNode() error
		FocusNextHiddenNode() error
	}

	transparentMonocle struct {
		logger     *log.Logger
		bspcClient bspc.Client
		desktops   state.TransparentMonocle
	}
)

func StartTransparentMonocle(
	logger *log.Logger,
	desktops state.TransparentMonocle,
	client bspc.Client,
) (TransparentMonocle, func(), error) {
	evCh, errCh, err := client.SubscribeEvents(bspc.EventTypeNodeAdd, bspc.EventTypeNodeRemove)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to events: %v", err)
	}

	cancelCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-cancelCh:
				logger.Info("closing transparent monocle events subscription")
				return

			case err, ok := <-errCh:
				if !ok {
					logger.Error("error channel closed unexpectedly", zap.Error(err))
					return
				}

				logger.Error("error received from transparent monocle events subscription", zap.Error(err))

			case ev, ok := <-evCh:
				if !ok {
					logger.Error("event channel closed unexpectedly", zap.Error(err))
					return
				}

				switch ev.Type {
				case bspc.EventTypeNodeAdd:
					payload, ok := ev.Payload.(bspc.EventNodeAdd)
					if !ok {
						logger.Error("failed to type cast event into specified event type",
							zap.String("event_type", string(bspc.EventTypeNodeAdd)),
							zap.Any("event_payload", ev.Payload),
						)
						continue
					}

					var addedNode bspc.Node
					if err := client.Query(fmt.Sprintf("query -n %d -T", payload.NodeID), bspc.ToStruct(&addedNode)); err != nil {
						logger.Error("failed to retrieve info on added node",
							zap.Uint("node_id", uint(payload.NodeID)),
							zap.Error(err),
						)
						continue
					}

					if addedNode.Client.State == bspc.StateTypeFloating {
						// It's a floating window. Ignore it
						continue
					}

					st, err := desktops.Get(payload.DesktopID)
					if err != nil {
						if !errors.Is(err, state.ErrNotFound) {
							logger.Error("failed to get desktop state",
								zap.Uint("desktop_id", uint(payload.DesktopID)),
								zap.Error(err),
							)
						}

						continue
					}

					newHiddenNodeIDs := st.HiddenNodeIDs
					if st.SelectedNodeID != nil {
						if err := client.Query(fmt.Sprintf("node %d --flag hidden=on", *st.SelectedNodeID), nil); err != nil {
							logger.Error("failed to hide previously focused node",
								zap.Uint("node_id", uint(*st.SelectedNodeID)),
								zap.Error(err),
							)
							continue
						}

						newHiddenNodeIDs = append(newHiddenNodeIDs, *st.SelectedNodeID)
					}

					desktops.Set(payload.DesktopID, state.TransparentMonocleState{
						SelectedNodeID: &payload.NodeID,
						HiddenNodeIDs:  newHiddenNodeIDs,
					})
				case bspc.EventTypeNodeRemove:
					payload, ok := ev.Payload.(bspc.EventNodeRemove)
					if !ok {
						logger.Error("failed to type cast event into specified event type",
							zap.String("event_type", string(bspc.EventTypeNodeRemove)),
							zap.Any("event_payload", ev.Payload),
						)
					}

					st, err := desktops.Get(payload.DesktopID)
					if err != nil {
						if !errors.Is(err, state.ErrNotFound) {
							logger.Error("failed to get desktop state",
								zap.Uint("desktop_id", uint(payload.DesktopID)),
								zap.Error(err),
							)
						}

						continue
					}

					if st.SelectedNodeID == nil || *st.SelectedNodeID != payload.NodeID {
						logger.Info("Ignoring non-selected node removal",
							zap.Uint("desktop_id", uint(payload.DesktopID)),
							zap.Error(err),
						)

						// It was most likely a floating window. Ignore it.
						continue
					}

					var (
						newSelectedNodeID *bspc.ID
						newHiddenNodeIDs  []bspc.ID
					)

					if len(st.HiddenNodeIDs) != 0 {
						newSelectedNodeID = &st.HiddenNodeIDs[len(st.HiddenNodeIDs)-1]

						if err := client.Query(fmt.Sprintf("node %d --flag hidden=off", *newSelectedNodeID), nil); err != nil {
							logger.Error("failed to show newly focused node",
								zap.Uint("node_id", uint(*st.SelectedNodeID)),
								zap.Error(err),
							)
							continue
						}

						newHiddenNodeIDs = removeFromSlice(st.HiddenNodeIDs, *newSelectedNodeID)
					}

					desktops.Set(payload.DesktopID, state.TransparentMonocleState{
						SelectedNodeID: newSelectedNodeID,
						HiddenNodeIDs:  newHiddenNodeIDs,
					})
				}
			}
		}
	}()

	cancelFunc := func() { cancelCh <- struct{}{} }

	return &transparentMonocle{
		logger:     logger,
		bspcClient: client,
		desktops:   desktops,
	}, cancelFunc, nil
}

func (tm transparentMonocle) ToggleCurrentDesktop() error {
	var desktop bspc.Desktop
	if err := tm.bspcClient.Query("query -d focused -T", bspc.ToStruct(&desktop)); err != nil {
		return fmt.Errorf("failed to get current desktop state: %v", err)
	}

	st, err := tm.desktops.Delete(desktop.ID)
	if err != nil {
		if !errors.Is(err, state.ErrNotFound) {
			return fmt.Errorf("failed to retrieve and delete state: %v", err)
		}

		return tm.enableMode(desktop)
	}

	return tm.disableMode(st)
}

func (tm transparentMonocle) enableMode(desktop bspc.Desktop) error {
	if err := tm.bspcClient.Query("desktop focused -l monocle", nil); err != nil {
		return fmt.Errorf("failed to set current desktop monocle layout: %v", err)
	}

	var (
		selectedNodeID *bspc.ID
		hiddenNodeIDs  []bspc.ID
	)

	if focused := desktop.FocusedNodeID; focused != bspc.NilID {
		allNodes := getVisibleLeafNodes(desktop.Root)

		selectedNodeID = &focused
		if n := allNodes[focused]; n.Client.State == bspc.StateTypeFloating {
			// If the focused node when monocle mode is activated is a floating node,
			// we'll just use the biggest node as the main one.
			var biggestNode bspc.Node
			if err := tm.bspcClient.Query("query -n biggest.local -T", bspc.ToStruct(&biggestNode)); err != nil {
				return fmt.Errorf("failed to query biggest node in current desktop: %v", err)
			}

			selectedNodeID = &biggestNode.ID
		}

		for id, n := range allNodes {
			if id == *selectedNodeID {
				continue
			}

			if n.Client.State == bspc.StateTypeFloating {
				continue
			}

			if err := tm.bspcClient.Query(fmt.Sprintf("node %d --flag hidden=on", id), nil); err != nil {
				return fmt.Errorf("failed to hide %d node: %v", id, err)
			}

			hiddenNodeIDs = append(hiddenNodeIDs, id)
		}
	}

	tm.desktops.Set(desktop.ID, state.TransparentMonocleState{
		SelectedNodeID: selectedNodeID,
		HiddenNodeIDs:  hiddenNodeIDs,
	})

	return nil
}

func (tm transparentMonocle) disableMode(st state.TransparentMonocleState) error {
	for _, n := range st.HiddenNodeIDs {
		if err := tm.bspcClient.Query(fmt.Sprintf("node %d --flag hidden=off", n), nil); err != nil {
			return fmt.Errorf("failed to un-hide %d node: %v", n, err)
		}
	}

	if err := tm.bspcClient.Query("desktop focused -l tiled", nil); err != nil {
		return fmt.Errorf("failed to set current desktop tiled layout: %v", err)
	}

	return nil
}

func (tm transparentMonocle) FocusPreviousHiddenNode() error {
	var desktop bspc.Desktop
	if err := tm.bspcClient.Query("query -d focused -T", bspc.ToStruct(&desktop)); err != nil {
		return fmt.Errorf("failed to get current desktop state: %v", err)
	}

	st, err := tm.desktops.Get(desktop.ID)
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return nil
		}

		return fmt.Errorf("failed to get desktop state: %v", err)
	}

	if st.SelectedNodeID == nil || len(st.HiddenNodeIDs) == 0 {
		// There are no nodes in the current desktop
		return nil
	}

	nextNodeID := st.HiddenNodeIDs[len(st.HiddenNodeIDs)-1]
	if err := tm.bspcClient.Query(fmt.Sprintf("node %d --flag hidden=off", nextNodeID), nil); err != nil {
		return fmt.Errorf("failed to un-hide %d node: %v", nextNodeID, err)
	}

	if err := tm.bspcClient.Query(fmt.Sprintf("node %d --flag hidden=on", *st.SelectedNodeID), nil); err != nil {
		return fmt.Errorf("failed to hide %d node: %v", st.SelectedNodeID, err)
	}

	tm.desktops.Set(desktop.ID, state.TransparentMonocleState{
		SelectedNodeID: &nextNodeID,
		HiddenNodeIDs:  append([]bspc.ID{*st.SelectedNodeID}, removeFromSlice(st.HiddenNodeIDs, nextNodeID)...),
	})

	return nil
}

func (tm transparentMonocle) FocusNextHiddenNode() error {
	var desktop bspc.Desktop
	if err := tm.bspcClient.Query("query -d focused -T", bspc.ToStruct(&desktop)); err != nil {
		return fmt.Errorf("failed to get current desktop state: %v", err)
	}

	st, err := tm.desktops.Get(desktop.ID)
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return nil
		}

		return fmt.Errorf("failed to get desktop state: %v", err)
	}

	if st.SelectedNodeID == nil || len(st.HiddenNodeIDs) == 0 {
		// There are no nodes in the current desktop
		return nil
	}

	nextNodeID := st.HiddenNodeIDs[0]
	if err := tm.bspcClient.Query(fmt.Sprintf("node %d --flag hidden=off", nextNodeID), nil); err != nil {
		return fmt.Errorf("failed to un-hide %d node: %v", nextNodeID, err)
	}

	if err := tm.bspcClient.Query(fmt.Sprintf("node %d --flag hidden=on", *st.SelectedNodeID), nil); err != nil {
		return fmt.Errorf("failed to hide %d node: %v", st.SelectedNodeID, err)
	}

	tm.desktops.Set(desktop.ID, state.TransparentMonocleState{
		SelectedNodeID: &nextNodeID,
		HiddenNodeIDs:  append(removeFromSlice(st.HiddenNodeIDs, nextNodeID), *st.SelectedNodeID),
	})

	return nil
}

func removeFromSlice(slice []bspc.ID, toRemove bspc.ID) []bspc.ID {
	ss := make([]bspc.ID, 0, len(slice)-1)
	for _, id := range slice {
		if id == toRemove {
			continue
		}

		ss = append(ss, id)
	}

	return ss
}

// This retrieves only the nodes that correspond to actual windows.
// Some nodes only serve to split the screen to hold other nodes.
// They don't represent windows.
// TODO: I made this return type into a map for simplicity in the code above. Could there be a performance hit, though?
//  If so, revert it.
func getVisibleLeafNodes(node bspc.Node) map[bspc.ID]bspc.Node {
	nodes := make(map[bspc.ID]bspc.Node)

	if node.FirstChild == nil && node.SecondChild == nil {
		nodes[node.ID] = node
	}

	if node.FirstChild != nil {
		for k, v := range getVisibleLeafNodes(*node.FirstChild) {
			nodes[k] = v
		}
	}

	if node.SecondChild != nil {
		for k, v := range getVisibleLeafNodes(*node.SecondChild) {
			nodes[k] = v
		}
	}

	return nodes
}
