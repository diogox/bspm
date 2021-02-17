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
	evCh, errCh, err := client.SubscribeEvents(
		bspc.EventTypeNodeAdd,
		bspc.EventTypeNodeRemove,
		bspc.EventTypeNodeTransfer,
		bspc.EventTypeNodeSwap,
	)
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

					if err := handleNodeAdded(logger, client, desktops, payload.DesktopID, payload.NodeID); err != nil {
						logger.Error("failed to handle added node",
							zap.Uint("desktop_id", uint(payload.DesktopID)),
							zap.Error(err),
						)
						continue
					}

				case bspc.EventTypeNodeRemove:
					payload, ok := ev.Payload.(bspc.EventNodeRemove)
					if !ok {
						logger.Error("failed to type cast event into specified event type",
							zap.String("event_type", string(bspc.EventTypeNodeRemove)),
							zap.Any("event_payload", ev.Payload),
						)
					}

					err := handleNodeRemoved(logger, client, desktops, payload.DesktopID, payload.NodeID)
					if err != nil {
						logger.Error("failed to handle removed node",
							zap.Uint("desktop_id", uint(payload.DesktopID)),
							zap.Error(err),
						)
					}

				case bspc.EventTypeNodeTransfer:
					// TODO: Add unit tests for this event
					payload, ok := ev.Payload.(bspc.EventNodeTransfer)
					if !ok {
						logger.Error("failed to type cast event into specified event type",
							zap.String("event_type", string(bspc.EventTypeNodeTransfer)),
							zap.Any("event_payload", ev.Payload),
						)
						continue
					}

					// TODO: I probably need to check for non-leaf nodes here. Like I did for Swaps below. Test it.

					// The source node id is the id of the node being transferred.
					// It's unclear what the destination node id is.
					// I think it's the id of the node whose position we're going to replace with this one.
					// TODO: Add a godoc to bspc-go to make this clear both for NodeTransfer and NodeSwap events.
					//  I'm assuming it's the same for the latter event type.
					transferredNodeID := payload.SourceNodeID

					if err := handleNodeRemoved(logger, client, desktops, payload.SourceDesktopID, transferredNodeID); err != nil {
						logger.Error("failed to handle node transfers at source",
							zap.Uint("desktop_id", uint(payload.SourceDesktopID)),
							zap.Error(err),
						)
					}

					if err := handleNodeAdded(logger, client, desktops, payload.DestinationDesktopID, transferredNodeID); err != nil {
						logger.Error("failed to handle node transfer at destination",
							zap.Uint("desktop_id", uint(payload.SourceDesktopID)),
							zap.Error(err),
						)
					}

				case bspc.EventTypeNodeSwap:
					// TODO: Add unit tests for this event
					payload, ok := ev.Payload.(bspc.EventNodeSwap)
					if !ok {
						logger.Error("failed to type cast event into specified event type",
							zap.String("event_type", string(bspc.EventTypeNodeTransfer)),
							zap.Any("event_payload", ev.Payload),
						)
						continue
					}

					if payload.SourceDesktopID == payload.DestinationDesktopID {
						// TODO: Is this even possible?
						// It's not going to affect this mode. Move on.
						continue
					}

					// TODO: Use this strategy in the other log instances.
					loggerOpts := []zap.Field{
						zap.Uint("source_desktop_id", uint(payload.SourceDesktopID)),
						zap.Uint("destination_desktop_id", uint(payload.DestinationDesktopID)),
						zap.Uint("source_node_id", uint(payload.SourceNodeID)),
						zap.Uint("destination_node_id", uint(payload.DestinationNodeID)),
					}

					isSourceDesktopMonocled := true
					if _, err := desktops.Get(payload.SourceDesktopID); err != nil {
						if !errors.Is(err, state.ErrNotFound) {
							logger.Error("failed to get source desktop state", append(loggerOpts, zap.Error(err))...)
							continue
						}

						isSourceDesktopMonocled = false
					}

					isDestinationDesktopMonocled := true
					if _, err := desktops.Get(payload.DestinationDesktopID); err != nil {
						if !errors.Is(err, state.ErrNotFound) {
							logger.Error("failed to get destination desktop state", append(loggerOpts, zap.Error(err))...)
							continue
						}

						isDestinationDesktopMonocled = false
					}

					if !isSourceDesktopMonocled && !isDestinationDesktopMonocled {
						// None of them are in monocle mode. Ignore.
						continue
					}

					// TODO: Move these bspc calls to a service layer so they can be reused more idiomatically.
					// TODO: This gets called in handleNodeAdded. Is there a way I can reuse this there?
					var sourceNode bspc.Node
					if err := client.Query(fmt.Sprintf("query -n %d -T", payload.SourceNodeID), bspc.ToStruct(&sourceNode)); err != nil {
						logger.Error("failed to query source node info", append(loggerOpts, zap.Error(err))...)
						continue
					}

					var destinationNode bspc.Node
					if err := client.Query(fmt.Sprintf("query -n %d -T", payload.DestinationNodeID), bspc.ToStruct(&destinationNode)); err != nil {
						logger.Error("failed to query destination node info", append(loggerOpts, zap.Error(err))...)
						continue
					}

					// TODO: This is probably not the best type. I'd need to change the `getVisibleLeafNodes` function to return a slice instead.
					sourceNodes := map[bspc.ID]bspc.Node{sourceNode.ID: sourceNode}
					if !isLeafNode(sourceNode) {
						sourceNodes = getVisibleLeafNodes(sourceNode)
					}

					destinationNodes := map[bspc.ID]bspc.Node{destinationNode.ID: destinationNode}
					if !isLeafNode(destinationNode) {
						destinationNodes = getVisibleLeafNodes(destinationNode)
					}

					for _, n := range sourceNodes {
						// We can't add hidden nodes to a desktop
						if n.Hidden {
							if err := client.Query(fmt.Sprintf("node %d --flag hidden=off", n.ID), nil); err != nil {
								logger.Error("failed to show hidden node being swapped",
									append(loggerOpts, zap.Error(err))...,
								)
								continue
							}
						}
					}

					for _, n := range destinationNodes {
						// We can't add hidden nodes to a desktop
						if n.Hidden {
							if err := client.Query(fmt.Sprintf("node %d --flag hidden=off", n.ID), nil); err != nil {
								logger.Error("failed to show hidden node being swapped",
									append(loggerOpts, zap.Error(err))...,
								)
								continue
							}
						}
					}

					// TODO: we can't know what node was focused atm. So the newly focused node (the one called last below) is
					//  going to be random for now. Refactor this when I add an internal representation of bspwm's state.

					for id := range sourceNodes {
						if err := handleNodeRemoved(logger, client, desktops, payload.SourceDesktopID, id); err != nil {
							logger.Error("failed to handle node swap (across desktops) source node removal at source desktop",
								append(loggerOpts, zap.Error(err))...,
							)
						}
					}
					for id := range destinationNodes {
						if err := handleNodeRemoved(logger, client, desktops, payload.DestinationDesktopID, id); err != nil {
							logger.Error("failed to handle node swap (across desktops) destination node added at destination desktop",
								append(loggerOpts, zap.Error(err))...,
							)
						}
					}

					for id := range sourceNodes {
						if err := handleNodeAdded(logger, client, desktops, payload.DestinationDesktopID, id); err != nil {
							logger.Error("failed to handle node swap (across desktops) source node added at destination desktop",
								append(loggerOpts, zap.Error(err))...,
							)
						}
					}
					for id := range destinationNodes {
						if err := handleNodeAdded(logger, client, desktops, payload.SourceDesktopID, id); err != nil {
							logger.Error("failed to handle node swap (across desktops) destination node added at source desktop",
								append(loggerOpts, zap.Error(err))...,
							)
						}
					}
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

func handleNodeRemoved(
	logger *log.Logger,
	client bspc.Client,
	desktops state.TransparentMonocle,
	desktopID bspc.ID,
	nodeID bspc.ID,
) error {
	st, err := desktops.Get(desktopID)
	if err != nil {
		if !errors.Is(err, state.ErrNotFound) {
			return fmt.Errorf("failed to get desktop state: %w", err)
		}

		return nil
	}

	if st.SelectedNodeID == nil || *st.SelectedNodeID != nodeID {
		var isHiddenNode bool
		for _, hiddenID := range st.HiddenNodeIDs {
			if nodeID == hiddenID {
				isHiddenNode = true
				break
			}
		}

		if isHiddenNode {
			desktops.Set(desktopID, state.TransparentMonocleState{
				SelectedNodeID: st.SelectedNodeID,
				HiddenNodeIDs:  removeFromSlice(st.HiddenNodeIDs, nodeID),
			})

			logger.Info("Ignoring hidden node removal",
				zap.Uint("desktop_id", uint(desktopID)),
				zap.Uint("node_id", uint(nodeID)),
			)
			return nil
		}

		logger.Info("Ignoring floating node removal",
			zap.Uint("desktop_id", uint(desktopID)),
			zap.Uint("node_id", uint(nodeID)),
		)

		// It was most likely a floating window. Ignore it.
		return nil
	}

	var (
		newSelectedNodeID *bspc.ID
		newHiddenNodeIDs  []bspc.ID
	)

	if len(st.HiddenNodeIDs) != 0 {
		newSelectedNodeID = &st.HiddenNodeIDs[len(st.HiddenNodeIDs)-1]

		if err := client.Query(fmt.Sprintf("node %d --flag hidden=off", *newSelectedNodeID), nil); err != nil {
			return fmt.Errorf("failed to show newly focused node: %w", err)
		}

		newHiddenNodeIDs = removeFromSlice(st.HiddenNodeIDs, *newSelectedNodeID)
	}

	desktops.Set(desktopID, state.TransparentMonocleState{
		SelectedNodeID: newSelectedNodeID,
		HiddenNodeIDs:  newHiddenNodeIDs,
	})

	return nil
}

func handleNodeAdded(
	logger *log.Logger,
	client bspc.Client,
	desktops state.TransparentMonocle,
	desktopID bspc.ID,
	nodeID bspc.ID,
) error {
	var addedNode bspc.Node
	if err := client.Query(fmt.Sprintf("query -n %d -T", nodeID), bspc.ToStruct(&addedNode)); err != nil {
		return fmt.Errorf("failed to retrieve info on added node: %w", err)
	}

	if addedNode.Client.State == bspc.StateTypeFloating {
		logger.Info("Ignoring non-selected node removal",
			zap.Uint("desktop_id", uint(desktopID)),
			zap.Uint("node_id", uint(nodeID)),
		)

		// It's a floating window. Ignore it
		return nil
	}

	st, err := desktops.Get(desktopID)
	if err != nil {
		if !errors.Is(err, state.ErrNotFound) {
			return fmt.Errorf("failed to get desktop state: %w", err)
		}

		return nil
	}

	newHiddenNodeIDs := st.HiddenNodeIDs
	if st.SelectedNodeID != nil {
		if err := client.Query(fmt.Sprintf("node %d --flag hidden=on", *st.SelectedNodeID), nil); err != nil {
			return fmt.Errorf("failed to hide previously focused node: %w", err)
		}

		newHiddenNodeIDs = append(newHiddenNodeIDs, *st.SelectedNodeID)
	}

	desktops.Set(desktopID, state.TransparentMonocleState{
		SelectedNodeID: &nodeID,
		HiddenNodeIDs:  newHiddenNodeIDs,
	})

	return nil
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

	if isLeafNode(node) {
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

// TODO: Add this as a method to nodes in bspc-go?
func isLeafNode(node bspc.Node) bool {
	return node.FirstChild == nil && node.SecondChild == nil
}
