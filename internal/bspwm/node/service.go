//go:generate mockgen -package bspwmnode -destination ./service_mock.go -self_package github.com/diogox/bspm/internal/bspwm/node github.com/diogox/bspm/internal/bspwm/node Service

package bspwmnode

import (
	"fmt"

	"github.com/diogox/bspc-go"

	"github.com/diogox/bspm/internal/bspwm/filter"
)

type (
	Service interface {
		Get(filter filter.NodeFilter) (bspc.Node, error)
		SetVisibility(id bspc.ID, isVisible bool) error
	}
	service struct {
		client bspc.Client
	}
)

func NewService(client bspc.Client) Service {
	return service{
		client: client,
	}
}

func (s service) Get(filter filter.NodeFilter) (bspc.Node, error) {
	const descriptor = "query -n %s -T"

	cmd := fmt.Sprintf(descriptor, filter)

	var node bspc.Node
	if err := s.client.Query(cmd, bspc.ToStruct(&node)); err != nil {
		return bspc.Node{}, fmt.Errorf("failed to get node: %w", err)
	}

	return node, nil
}

func (s service) SetVisibility(id bspc.ID, isVisible bool) error {
	const descriptor = "node %d --flag hidden=%s"

	const (
		actionOn  = "on"
		actionOff = "off"
	)

	var action string
	switch isVisible {
	case true:
		action = actionOff
	case false:
		action = actionOn
	}

	cmd := fmt.Sprintf(descriptor, id, action)

	if err := s.client.Query(cmd, nil); err != nil {
		return fmt.Errorf("failed to set visibility: %w", err)
	}

	return nil
}
