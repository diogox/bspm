package state

import (
	"github.com/diogox/bspc-go"
)

func (m manager) WithState(newState map[bspc.ID]State) manager {
	m.desktops = newState
	return m
}
