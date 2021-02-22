//go:generate mockgen -package bspwm -destination ./service_mock.go -self_package github.com/diogox/bspm/internal/bspwm github.com/diogox/bspm/internal/bspwm Service

package bspwm

import (
	"fmt"

	"github.com/diogox/bspc-go"

	bspwmdesktop "github.com/diogox/bspm/internal/bspwm/desktop"
	bspwmevent "github.com/diogox/bspm/internal/bspwm/event"
	bspwmnode "github.com/diogox/bspm/internal/bspwm/node"
)

type (
	Service interface {
		State() (bspc.State, error)
		Desktops() bspwmdesktop.Service
		Nodes() bspwmnode.Service
		Events() bspwmevent.Manager
	}
)

type service struct {
	client         bspc.Client
	desktopService bspwmdesktop.Service
	nodeService    bspwmnode.Service
	eventManager   bspwmevent.Manager
}

func NewService(
	client bspc.Client,
	desktopService bspwmdesktop.Service,
	nodeService bspwmnode.Service,
	eventManager bspwmevent.Manager,
) Service {
	return service{
		client:         client,
		desktopService: desktopService,
		nodeService:    nodeService,
		eventManager:   eventManager,
	}
}

// State returns bspwm's current state.
func (s service) State() (bspc.State, error) {
	const cmd = "wm --dump-state"

	var state bspc.State
	if err := s.client.Query(cmd, bspc.ToStruct(&state)); err != nil {
		return bspc.State{}, fmt.Errorf("failed to retrieve bspwm state: %w", err)
	}

	return state, nil
}

// Desktops returns the service for dealing with bspwm desktops.
func (s service) Desktops() bspwmdesktop.Service {
	return s.desktopService
}

// Nodes returns the service for dealing with bspwm nodes.
func (s service) Nodes() bspwmnode.Service {
	return s.nodeService
}

// Events returns the service for dealing with bspwm events.
func (s service) Events() bspwmevent.Manager {
	return s.eventManager
}
