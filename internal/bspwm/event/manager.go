//go:generate mockgen -package bspwmevent -destination ./manager_mock.go -self_package github.com/diogox/bspm/internal/bspwm/event github.com/diogox/bspm/internal/bspwm/event Manager

package bspwmevent

import (
	"errors"
	"fmt"

	"github.com/diogox/bspc-go"
	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/log"
)

type (
	callbackFunc func(eventPayload interface{}) error
	cancelFunc   func()
)

type (
	Manager interface {
		On(eventType bspc.EventType, callback callbackFunc)
		Start() (cancelFunc, error)
	}
	manager struct {
		logger    *log.Logger
		client    bspc.Client
		callbacks map[bspc.EventType][]callbackFunc
	}
)

func NewManager(logger *log.Logger, client bspc.Client) Manager {
	return manager{
		logger:    logger,
		client:    client,
		callbacks: make(map[bspc.EventType][]callbackFunc),
	}
}

// On takes in an event type and the callback that should be called when a corresponding
// event is triggered. Callbacks are called in order for each event.
func (m manager) On(eventType bspc.EventType, callback callbackFunc) {
	cc, ok := m.callbacks[eventType]
	if !ok {
		m.callbacks[eventType] = []callbackFunc{callback}
		return
	}

	m.callbacks[eventType] = append(cc, callback)
}

// Start subscribes to all the necessary events and calls the callbacks when they are triggered.
// It should be called after all the necessary event callbacks are added.
func (m manager) Start() (cancelFunc, error) {
	var evTypes []bspc.EventType
	for t := range m.callbacks {
		evTypes = append(evTypes, t)
	}

	if len(evTypes) == 0 {
		return nil, errors.New("no events subscribed")
	}

	evCh, errCh, err := m.client.SubscribeEvents(evTypes[0], evTypes[1:]...)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}

	cancelCh := make(chan struct{})
	go m.handleEvents(evCh, errCh, cancelCh)

	cancel := func() { cancelCh <- struct{}{} }
	return cancel, nil
}

func (m manager) handleEvents(evCh chan bspc.Event, errCh chan error, cancelCh chan struct{}) {
	for {
		select {
		case <-cancelCh:
			m.logger.Info("closing transparent monocle events subscription")
			return

		case err, ok := <-errCh:
			if !ok {
				m.logger.Error("error channel closed unexpectedly", zap.Error(err))
				return
			}

			m.logger.Error("error received subscribing to events", zap.Error(err))

		case ev, ok := <-evCh:
			if !ok {
				m.logger.Error("event channel closed unexpectedly")
				return
			}

			for _, callback := range m.callbacks[ev.Type] {
				if err := callback(ev.Payload); err != nil {
					m.logger.Error("error running event callback",
						zap.String("event_type", string(ev.Type)),
						zap.Any("event_payload", ev.Payload),
						zap.Error(err),
					)
				}
			}
		}
	}
}
