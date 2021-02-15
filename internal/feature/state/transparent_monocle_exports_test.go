package state

import "sync"

func (tm transparentMonocle) WithMap(m *sync.Map) transparentMonocle {
	tm.desktops = m
	return tm
}
