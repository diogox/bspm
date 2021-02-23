//go:generate mockgen -package state -destination ./state_mock.go -self_package github.com/diogox/bspm/internal/feature/transparent_monocle/state github.com/diogox/bspm/internal/feature/transparent_monocle/state Manager

package state

import (
	"errors"
	"sync"

	"github.com/diogox/bspc-go"
)

type (
	Manager interface {
		Get(desktopID bspc.ID) (State, error)
		Set(desktopID bspc.ID, st State)
		Delete(desktopID bspc.ID) (State, error)
	}

	State struct {
		SelectedNodeID *bspc.ID
		HiddenNodeIDs  []bspc.ID
	}

	manager struct {
		desktops *sync.Map
	}
)

func NewTransparentMonocle() manager {
	return manager{
		desktops: &sync.Map{},
	}
}

func (tm manager) Get(desktopID bspc.ID) (State, error) {
	stObj, ok := tm.desktops.Load(desktopID)
	if !ok {
		return State{}, ErrNotFound
	}

	st, ok := stObj.(State)
	if !ok {
		return State{}, errors.New("invalid state type")
	}

	return st, nil
}

func (tm manager) Set(desktopID bspc.ID, st State) {
	tm.desktops.Store(desktopID, st)
}

func (tm manager) Delete(desktopID bspc.ID) (State, error) {
	stObj, ok := tm.desktops.LoadAndDelete(desktopID)
	if !ok {
		return State{}, ErrNotFound
	}

	st, ok := stObj.(State)
	if !ok {
		return State{}, errors.New("invalid state type")
	}

	return st, nil
}
