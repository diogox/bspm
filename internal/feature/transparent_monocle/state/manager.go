//go:generate mockgen -package state -destination ./manager_mock.go -self_package github.com/diogox/bspm/internal/feature/transparent_monocle/state github.com/diogox/bspm/internal/feature/transparent_monocle/state Manager

package state

import (
	"sync"

	"github.com/diogox/bspc-go"

	"github.com/diogox/bspm/internal/feature/transparent_monocle/topic"
	"github.com/diogox/bspm/internal/subscription"
)

type (
	Manager interface {
		Get(desktopID bspc.ID) (State, bool)
		Set(desktopID bspc.ID, st State)
		Delete(desktopID bspc.ID)
	}

	State struct {
		SelectedNodeID *bspc.ID
		HiddenNodeIDs  []bspc.ID
	}

	manager struct {
		rwMutex       *sync.RWMutex
		subscriptions subscription.Manager
		desktops      map[bspc.ID]State
	}
)

func NewTransparentMonocle(subscriptions subscription.Manager) manager {
	return manager{
		rwMutex:       &sync.RWMutex{},
		subscriptions: subscriptions,
		desktops:      make(map[bspc.ID]State),
	}
}

func (m manager) Get(desktopID bspc.ID) (State, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()

	st, ok := m.desktops[desktopID]
	return st, ok
}

func (m manager) Set(desktopID bspc.ID, st State) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	if _, ok := m.desktops[desktopID]; !ok {
		m.desktops[desktopID] = st
		m.subscriptions.Publish(topic.MonocleEnabled, st)
		return
	}

	m.desktops[desktopID] = st
	m.subscriptions.Publish(topic.MonocleStateChanged, st)
}

func (m manager) Delete(desktopID bspc.ID) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	prevState := m.desktops[desktopID]

	delete(m.desktops, desktopID)
	m.subscriptions.Publish(topic.MonocleDisabled, prevState)
}
