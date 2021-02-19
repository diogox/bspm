//go:generate mockgen -package bspwmdesktop -destination ./service_mock.go -self_package github.com/diogox/bspm/internal/bspwm/desktop github.com/diogox/bspm/internal/bspwm/desktop Service

package bspwmdesktop

import (
	"fmt"

	"github.com/diogox/bspc-go"

	"github.com/diogox/bspm/internal/bspwm/filter"
)

type (
	Service interface {
		Get(filter filter.DesktopFilter) (bspc.Desktop, error)
		SetLayout(filter filter.DesktopFilter, layout bspc.LayoutType) error
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

func (s service) Get(filter filter.DesktopFilter) (bspc.Desktop, error) {
	const descriptor = "query -d %s -T"

	cmd := fmt.Sprintf(descriptor, filter)

	var desktop bspc.Desktop
	if err := s.client.Query(cmd, bspc.ToStruct(&desktop)); err != nil {
		return bspc.Desktop{}, fmt.Errorf("failed to get desktop: %w", err)
	}

	return desktop, nil
}

func (s service) SetLayout(filter filter.DesktopFilter, layout bspc.LayoutType) error {
	const descriptor = "desktop %s -l %s"

	cmd := fmt.Sprintf(descriptor, filter, layout)

	if err := s.client.Query(cmd, nil); err != nil {
		return fmt.Errorf("failed to set desktop layout: %w", err)
	}

	return nil
}
