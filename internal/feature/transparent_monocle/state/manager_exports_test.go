package state

import (
	"github.com/diogox/bspc-go"
)

const (
	ExportMonocleEnabled      = topicMonocleEnabled
	ExportMonocleDisabled     = topicMonocleDisabled
	ExportMonocleStateChanged = topicMonocleStateChanged
)

func (m manager) WithState(newState map[bspc.ID]State) manager {
	m.desktops = newState
	return m
}
