//go:generate mockgen -package state -destination ./transparent_monocle_mock.go -self_package github.com/diogox/bspm/internal/feature/state github.com/diogox/bspm/internal/feature/state TransparentMonocle

package state

import (
	"errors"
	"sync"

	"github.com/diogox/bspc-go"
)

type (
	TransparentMonocle interface {
		Get(desktopID bspc.ID) (TransparentMonocleState, error)
		Set(desktopID bspc.ID, st TransparentMonocleState)
		Delete(desktopID bspc.ID) (TransparentMonocleState, error)
	}

	TransparentMonocleState struct {
		SelectedNodeID *bspc.ID
		HiddenNodeIDs  []bspc.ID
	}

	transparentMonocle struct {
		desktops *sync.Map
	}
)

func NewTransparentMonocle() TransparentMonocle {
	return transparentMonocle{
		desktops: &sync.Map{},
	}
}

func (tm transparentMonocle) Get(desktopID bspc.ID) (TransparentMonocleState, error) {
	stObj, ok := tm.desktops.Load(desktopID)
	if !ok {
		return TransparentMonocleState{}, ErrNotFound
	}

	st, ok := stObj.(TransparentMonocleState)
	if !ok {
		return TransparentMonocleState{}, errors.New("invalid state type")
	}

	return st, nil
}

func (tm transparentMonocle) Set(desktopID bspc.ID, st TransparentMonocleState) {
	tm.desktops.Store(desktopID, st)
}

func (tm transparentMonocle) Delete(desktopID bspc.ID) (TransparentMonocleState, error) {
	stObj, ok := tm.desktops.LoadAndDelete(desktopID)
	if !ok {
		return TransparentMonocleState{}, ErrNotFound
	}

	st, ok := stObj.(TransparentMonocleState)
	if !ok {
		return TransparentMonocleState{}, errors.New("invalid state type")
	}

	return st, nil
}
