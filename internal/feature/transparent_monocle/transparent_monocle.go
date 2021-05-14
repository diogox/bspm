//go:generate mockgen -package transparentmonocle -destination ./bspc_mock.go github.com/diogox/bspc-go Client
//go:generate mockgen -package transparentmonocle -destination ./transparent_monocle_mock.go -self_package github.com/diogox/bspm/internal/feature/transparent_monocle github.com/diogox/bspm/internal/feature/transparent_monocle Feature

package transparentmonocle

import (
	"errors"
	"fmt"

	"github.com/diogox/bspc-go"
	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/feature/transparent_monocle/topic"
	"github.com/diogox/bspm/internal/subscription"

	"github.com/diogox/bspm/internal/bspwm"
	"github.com/diogox/bspm/internal/bspwm/filter"

	"github.com/diogox/bspm/internal/feature/transparent_monocle/state"
	"github.com/diogox/bspm/internal/log"
)

type (
	Feature interface {
		ToggleCurrentDesktop() error
		FocusPreviousHiddenNode() error
		FocusNextHiddenNode() error
		SubscribeNodeCount() chan int
	}

	transparentMonocle struct {
		logger        *log.Logger
		service       bspwm.Service
		desktops      state.Manager
		subscriptions subscription.Manager
	}
)

var ErrFeatureNotEnabled = errors.New("feature not enabled in current desktop")

func Start(
	logger *log.Logger,
	desktops state.Manager,
	service bspwm.Service,
	subscriptions subscription.Manager,
) (Feature, func(), error) {
	service.Events().On(bspc.EventTypeNodeAdd, func(eventPayload interface{}) error {
		payload, ok := eventPayload.(bspc.EventNodeAdd)
		if !ok {
			return errors.New("invalid event payload")
		}

		if err := handleNodeAdded(logger, service, desktops, payload.DesktopID, payload.NodeID); err != nil {
			logger.Error("failed to handle added node",
				zap.Uint("desktop_id", uint(payload.DesktopID)),
				zap.Error(err),
			)

			return err
		}

		return nil
	})

	service.Events().On(bspc.EventTypeNodeRemove, func(eventPayload interface{}) error {
		payload, ok := eventPayload.(bspc.EventNodeRemove)
		if !ok {
			return errors.New("invalid event payload")
		}

		err := handleNodeRemoved(logger, service, desktops, payload.DesktopID, payload.NodeID)
		if err != nil {
			logger.Error("failed to handle removed node",
				zap.Uint("desktop_id", uint(payload.DesktopID)),
				zap.Error(err),
			)

			return err
		}

		return nil
	})

	service.Events().On(bspc.EventTypeNodeTransfer, func(eventPayload interface{}) error {
		// TODO: Add unit tests for this event
		payload, ok := eventPayload.(bspc.EventNodeTransfer)
		if !ok {
			return errors.New("invalid event payload")
		}

		// The source node id is the id of the node being transferred.
		// It's unclear what the destination node id is.
		// I think it's the id of the node whose position we're going to replace with this one.
		// TODO: Add a godoc to bspc-go to make this clear both for NodeTransfer and NodeSwap events.
		//  I'm assuming it's the same for the latter event type.
		transferredNodeID := payload.SourceNodeID

		if err := handleNodeRemoved(logger, service, desktops, payload.SourceDesktopID, transferredNodeID); err != nil {
			logger.Error("failed to handle node transfers at source",
				zap.Uint("desktop_id", uint(payload.SourceDesktopID)),
				zap.Error(err),
			)

			return err
		}

		if err := handleNodeAdded(logger, service, desktops, payload.DestinationDesktopID, transferredNodeID); err != nil {
			logger.Error("failed to handle node transfer at destination",
				zap.Uint("desktop_id", uint(payload.SourceDesktopID)),
				zap.Error(err),
			)

			return err
		}

		return nil
	})

	service.Events().On(bspc.EventTypeNodeSwap, func(eventPayload interface{}) error {
		// TODO: this one is HUGE. Need to move this to a private method, or a function.
		// TODO: Add unit tests for this event

		payload, ok := eventPayload.(bspc.EventNodeSwap)
		if !ok {
			return errors.New("invalid event payload")
		}

		if payload.SourceDesktopID == payload.DestinationDesktopID {
			// TODO: Is this even possible?
			// It's not going to affect this mode. Move on.
			return nil
		}

		// TODO: Use this strategy in the other log instances.
		loggerOpts := []zap.Field{
			zap.Uint("source_desktop_id", uint(payload.SourceDesktopID)),
			zap.Uint("destination_desktop_id", uint(payload.DestinationDesktopID)),
			zap.Uint("source_node_id", uint(payload.SourceNodeID)),
			zap.Uint("destination_node_id", uint(payload.DestinationNodeID)),
		}

		var isSourceDesktopMonocled bool
		if _, ok := desktops.Get(payload.SourceDesktopID); ok {
			isSourceDesktopMonocled = true
		}

		var isDestinationDesktopMonocled bool
		if _, ok := desktops.Get(payload.DestinationDesktopID); ok {
			isDestinationDesktopMonocled = true
		}

		if !isSourceDesktopMonocled && !isDestinationDesktopMonocled {
			// None of them are in monocle mode. Ignore.
			return nil
		}

		// TODO: This gets called in handleNodeAdded. Is there a way I can reuse this there?
		sourceNode, err := service.Nodes().Get(filter.NodeID(payload.SourceNodeID))
		if err != nil {
			logger.Error("failed to get source node info", append(loggerOpts, zap.Error(err))...)
			return err
		}

		destinationNode, err := service.Nodes().Get(filter.NodeID(payload.DestinationNodeID))
		if err != nil {
			logger.Error("failed to get destination node info", append(loggerOpts, zap.Error(err))...)
			return err
		}

		var (
			sourceNodes      = sourceNode.LeafNodes()
			destinationNodes = destinationNode.LeafNodes()
		)

		for _, n := range sourceNodes {
			// We can't add hidden nodes to a desktop
			if n.Hidden {
				if err := service.Nodes().SetVisibility(n.ID, true); err != nil {
					logger.Error("failed to show hidden node being swapped",
						append(loggerOpts, zap.Error(err))...,
					)
					// TODO: At this point, the mode might be crashed. How to handle this gracefully?
					//  Same for other errors below. Saga pattern won't help here, I think.
					return err
				}
			}
		}

		for _, n := range destinationNodes {
			// We can't add hidden nodes to a desktop
			if n.Hidden {
				if err := service.Nodes().SetVisibility(n.ID, true); err != nil {
					logger.Error("failed to show hidden node being swapped",
						append(loggerOpts, zap.Error(err))...,
					)

					return err
				}
			}
		}

		st, err := service.State()
		if err != nil {
			logger.Error("failed to retrieve bspwm's current state",
				append(loggerOpts, zap.Error(err))...,
			)
		}

		for _, n := range sourceNodes {
			if err := handleNodeRemoved(logger, service, desktops, payload.SourceDesktopID, n.ID); err != nil {
				logger.Error("failed to handle node swap (across desktops) source node removal at source desktop",
					append(loggerOpts, zap.Error(err))...,
				)

				return err
			}
		}
		for _, n := range destinationNodes {
			if err := handleNodeRemoved(logger, service, desktops, payload.DestinationDesktopID, n.ID); err != nil {
				logger.Error("failed to handle node swap (across desktops) destination node added at destination desktop",
					append(loggerOpts, zap.Error(err))...,
				)

				return err
			}
		}

		if len(sourceNodes) == 1 {
			focusedIndex, ok := findMostRecentlyFocusedNode(st.OrderedFocusHistory(), payload.SourceDesktopID, sourceNodes)
			if ok {
				focusedNode := sourceNodes[focusedIndex]

				// Move node to focus to the end of the slice so it gets called last. (giving it focus)
				sourceNodes = append(sourceNodes[:focusedIndex], sourceNodes[focusedIndex+1:]...)
				sourceNodes = append(sourceNodes, focusedNode)
			}
		}

		if len(destinationNodes) == 1 {
			focusedIndex, ok := findMostRecentlyFocusedNode(st.OrderedFocusHistory(), payload.DestinationDesktopID, destinationNodes)
			if ok {
				focusedNode := destinationNodes[focusedIndex]

				// Move node to focus to the end of the slice so it gets called last. (giving it focus)
				destinationNodes = append(destinationNodes[:focusedIndex], destinationNodes[focusedIndex+1:]...)
				destinationNodes = append(destinationNodes, focusedNode)
			}
		}

		for _, n := range sourceNodes {
			if err := handleNodeAdded(logger, service, desktops, payload.DestinationDesktopID, n.ID); err != nil {
				logger.Error("failed to handle node swap (across desktops) source node added at destination desktop",
					append(loggerOpts, zap.Error(err))...,
				)

				return err
			}
		}
		for _, n := range destinationNodes {
			if err := handleNodeAdded(logger, service, desktops, payload.SourceDesktopID, n.ID); err != nil {
				logger.Error("failed to handle node swap (across desktops) destination node added at source desktop",
					append(loggerOpts, zap.Error(err))...,
				)

				return err
			}
		}

		return nil
	})

	// TODO: I should extract these callback definitions to where they make the most sense.
	// Needed to trigger subscriptions when changing monocle mode instances (between desktops).
	service.Events().On(bspc.EventTypeDesktopFocus, func(eventPayload interface{}) error {
		subscriptions.Publish(topic.MonocleDesktopFocusChanged, nil)
		return nil
	})

	service.Events().On(bspc.EventTypeNodeState, func(eventPayload interface{}) error {
		payload, ok := eventPayload.(bspc.EventNodeState)
		if !ok {
			return errors.New("invalid event payload")
		}

		if payload.State != bspc.StateTypeFloating {
			// Ignore state change
			return nil
		}

		if _, ok := desktops.Get(payload.DesktopID); !ok {
			return nil
		}

		switch payload.WasEnabled {
		case true:
			err := handleNodeRemoved(logger, service, desktops, payload.DesktopID, payload.NodeID)
			if err != nil {
				logger.Error("failed to handle removing floating node",
					zap.Uint("desktop_id", uint(payload.DesktopID)),
					zap.Error(err),
				)

				return err
			}

		case false:
			err := handleNodeAdded(logger, service, desktops, payload.DesktopID, payload.NodeID)
			if err != nil {
				logger.Error("failed to handle adding un-floated node",
					zap.Uint("desktop_id", uint(payload.DesktopID)),
					zap.Error(err),
				)

				return err
			}
		}

		return nil
	})

	cancelFunc, err := service.Events().Start()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start event manager")
	}

	return &transparentMonocle{
		logger:        logger,
		service:       service,
		desktops:      desktops,
		subscriptions: subscriptions,
	}, cancelFunc, nil
}

func handleNodeRemoved(
	logger *log.Logger,
	service bspwm.Service,
	desktops state.Manager,
	desktopID bspc.ID,
	nodeID bspc.ID,
) error {
	st, ok := desktops.Get(desktopID)
	if !ok {
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
			desktops.Set(desktopID, state.State{
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

		if err := service.Nodes().SetVisibility(*newSelectedNodeID, true); err != nil {
			return fmt.Errorf("failed to show newly focused node: %w", err)
		}

		newHiddenNodeIDs = removeFromSlice(st.HiddenNodeIDs, *newSelectedNodeID)
	}

	desktops.Set(desktopID, state.State{
		SelectedNodeID: newSelectedNodeID,
		HiddenNodeIDs:  newHiddenNodeIDs,
	})

	return nil
}

func handleNodeAdded(
	logger *log.Logger,
	service bspwm.Service,
	desktops state.Manager,
	desktopID bspc.ID,
	nodeID bspc.ID,
) error {
	addedNode, err := service.Nodes().Get(filter.NodeID(nodeID))
	if err != nil {
		return fmt.Errorf("failed to get added node: %w", err)
	}

	if addedNode.Client.State == bspc.StateTypeFloating {
		logger.Info("Ignoring non-selected node removal",
			zap.Uint("desktop_id", uint(desktopID)),
			zap.Uint("node_id", uint(nodeID)),
		)

		// It's a floating window. Ignore it
		return nil
	}

	st, ok := desktops.Get(desktopID)
	if !ok {
		return nil
	}

	newHiddenNodeIDs := st.HiddenNodeIDs
	if st.SelectedNodeID != nil {
		if err := service.Nodes().SetVisibility(*st.SelectedNodeID, false); err != nil {
			return fmt.Errorf("failed to hide previously focused node: %w", err)
		}

		newHiddenNodeIDs = append(newHiddenNodeIDs, *st.SelectedNodeID)
	}

	desktops.Set(desktopID, state.State{
		SelectedNodeID: &nodeID,
		HiddenNodeIDs:  newHiddenNodeIDs,
	})

	return nil
}

func (tm transparentMonocle) ToggleCurrentDesktop() error {
	desktop, err := tm.service.Desktops().Get(filter.DesktopFocused)
	if err != nil {
		return fmt.Errorf("failed to get current desktop: %w", err)
	}

	st, ok := tm.desktops.Get(desktop.ID)
	if !ok {
		return tm.enableMode(desktop)
	}

	tm.desktops.Delete(desktop.ID)
	return tm.disableMode(st)
}

func (tm transparentMonocle) enableMode(desktop bspc.Desktop) error {
	if err := tm.service.Desktops().SetLayout(filter.DesktopFocused, bspc.LayoutTypeMonocle); err != nil {
		return fmt.Errorf("failed to set current desktop monocle layout: %v", err)
	}

	var (
		selectedNodeID *bspc.ID
		hiddenNodeIDs  []bspc.ID
	)

	if focused := desktop.FocusedNodeID; focused != bspc.NilID {
		leafNodes := make(map[bspc.ID]bspc.Node)
		for _, n := range desktop.Root.LeafNodes() {
			leafNodes[n.ID] = n
		}

		selectedNodeID = &focused
		if n := leafNodes[focused]; n.Client.State == bspc.StateTypeFloating {
			// If the focused node when monocle mode is activated is a floating node,
			// we'll just use the biggest node as the main one.
			biggestNode, err := tm.service.Nodes().Get(filter.NodeLocalBiggest)
			if err != nil {
				return fmt.Errorf("failed to query biggest node in current desktop: %v", err)
			}

			selectedNodeID = &biggestNode.ID
		}

		for id, n := range leafNodes {
			if id == *selectedNodeID {
				continue
			}

			if n.Client.State == bspc.StateTypeFloating {
				continue
			}

			if err := tm.service.Nodes().SetVisibility(id, false); err != nil {
				return fmt.Errorf("failed to hide node: %w", err)
			}

			hiddenNodeIDs = append(hiddenNodeIDs, id)
		}
	}

	tm.desktops.Set(desktop.ID, state.State{
		SelectedNodeID: selectedNodeID,
		HiddenNodeIDs:  hiddenNodeIDs,
	})

	return nil
}

func (tm transparentMonocle) disableMode(st state.State) error {
	for _, n := range st.HiddenNodeIDs {
		if err := tm.service.Nodes().SetVisibility(n, true); err != nil {
			return fmt.Errorf("failed to show node: %w", err)
		}
	}

	if err := tm.service.Desktops().SetLayout(filter.DesktopFocused, bspc.LayoutTypeTiled); err != nil {
		return fmt.Errorf("failed to set current desktop tiled layout: %w", err)
	}

	return nil
}

func (tm transparentMonocle) FocusPreviousHiddenNode() error {
	desktop, err := tm.service.Desktops().Get(filter.DesktopFocused)
	if err != nil {
		return fmt.Errorf("failed to get current desktop state: %v", err)
	}

	st, ok := tm.desktops.Get(desktop.ID)
	if !ok {
		return ErrFeatureNotEnabled
	}

	if st.SelectedNodeID == nil || len(st.HiddenNodeIDs) == 0 {
		// There are no nodes in the current desktop
		return nil
	}

	nextNodeID := st.HiddenNodeIDs[len(st.HiddenNodeIDs)-1]
	if err := tm.service.Nodes().SetVisibility(nextNodeID, true); err != nil {
		return fmt.Errorf("failed to un-hide %d node: %v", nextNodeID, err)
	}

	if err := tm.service.Nodes().SetVisibility(*st.SelectedNodeID, false); err != nil {
		return fmt.Errorf("failed to hide %d node: %v", st.SelectedNodeID, err)
	}

	tm.desktops.Set(desktop.ID, state.State{
		SelectedNodeID: &nextNodeID,
		HiddenNodeIDs:  append([]bspc.ID{*st.SelectedNodeID}, removeFromSlice(st.HiddenNodeIDs, nextNodeID)...),
	})

	return nil
}

func (tm transparentMonocle) FocusNextHiddenNode() error {
	desktop, err := tm.service.Desktops().Get(filter.DesktopFocused)
	if err != nil {
		return fmt.Errorf("failed to get current desktop state: %v", err)
	}

	st, ok := tm.desktops.Get(desktop.ID)
	if !ok {
		return ErrFeatureNotEnabled
	}

	if st.SelectedNodeID == nil || len(st.HiddenNodeIDs) == 0 {
		// There are no nodes in the current desktop
		return nil
	}

	nextNodeID := st.HiddenNodeIDs[0]
	if err := tm.service.Nodes().SetVisibility(nextNodeID, true); err != nil {
		return fmt.Errorf("failed to show %d node: %v", nextNodeID, err)
	}

	if err := tm.service.Nodes().SetVisibility(*st.SelectedNodeID, false); err != nil {
		return fmt.Errorf("failed to hide %d node: %v", st.SelectedNodeID, err)
	}

	tm.desktops.Set(desktop.ID, state.State{
		SelectedNodeID: &nextNodeID,
		HiddenNodeIDs:  append(removeFromSlice(st.HiddenNodeIDs, nextNodeID), *st.SelectedNodeID),
	})

	return nil
}

func (tm transparentMonocle) SubscribeNodeCount() chan int {
	var (
		stateCh        = tm.subscriptions.Subscribe(topic.MonocleStateChanged)
		enabledCh      = tm.subscriptions.Subscribe(topic.MonocleEnabled)
		disabledCh     = tm.subscriptions.Subscribe(topic.MonocleDisabled)
		desktopFocusCh = tm.subscriptions.Subscribe(topic.MonocleDesktopFocusChanged)
	)

	var (
		countCh               = make(chan int, 1)
		publishCountFromState = func(st state.State) {
			count := len(st.HiddenNodeIDs)
			if st.SelectedNodeID != nil {
				count++
			}

			countCh <- count
		}
		getAndPublishCount = func() {
			focusedDesktop, err := tm.service.Desktops().Get(filter.DesktopFocused)
			if err == nil { // TODO: Log if there's an error?
				currentState, ok := tm.desktops.Get(focusedDesktop.ID)
				switch ok {
				case true:
					publishCountFromState(currentState)
				case false:
					// Mode is disabled
					countCh <- -1
				}
			}
		}
	)

	// Publish current number of nodes
	getAndPublishCount()

	go func() {
		for {
			select {
			case payload := <-stateCh:
				updatedState := payload.(state.State)
				publishCountFromState(updatedState)

			case payload := <-enabledCh:
				updatedState := payload.(state.State)
				publishCountFromState(updatedState)

			case <-desktopFocusCh:
				getAndPublishCount()

			case <-disabledCh:
				countCh <- -1
			}
		}
	}()

	return countCh
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

// findMostRecentlyFocusedNode returns the node from the provided slice that shows up first in the focused node history.
func findMostRecentlyFocusedNode(focusHistory []bspc.StateFocusHistoryEntry, relevantDesktopID bspc.ID, nodes []bspc.Node) (int, bool) {
	for _, prevFocusedNode := range focusHistory {
		if prevFocusedNode.DesktopID != relevantDesktopID {
			continue
		}

		for index, n := range nodes {
			if prevFocusedNode.NodeID == n.ID {
				return index, true
			}
		}
	}

	return -1, false
}
