//go:generate mockgen -package bspwm -destination ./service_mock.go -self_package github.com/diogox/bspm/internal/bspwm github.com/diogox/bspm/internal/bspwm Service

package bspwm

import (
	bspwmdesktop "github.com/diogox/bspm/internal/bspwm/desktop"
	bspwmevent "github.com/diogox/bspm/internal/bspwm/event"
	bspwmnode "github.com/diogox/bspm/internal/bspwm/node"
)

type (
	Service interface {
		Desktops() bspwmdesktop.Service
		Nodes() bspwmnode.Service
		Events() bspwmevent.Manager
	}
)

type service struct {
	desktopService bspwmdesktop.Service
	nodeService    bspwmnode.Service
	eventManager   bspwmevent.Manager
}

func NewService(
	desktopService bspwmdesktop.Service,
	nodeService bspwmnode.Service,
	eventManager bspwmevent.Manager,
) Service {
	return service{
		desktopService: desktopService,
		nodeService:    nodeService,
		eventManager:   eventManager,
	}
}

func (s service) Desktops() bspwmdesktop.Service {
	return s.desktopService
}

func (s service) Nodes() bspwmnode.Service {
	return s.nodeService
}

func (s service) Events() bspwmevent.Manager {
	return s.eventManager
}
