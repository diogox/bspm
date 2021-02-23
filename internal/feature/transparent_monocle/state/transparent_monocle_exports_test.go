package state

import "sync"

func (tm manager) WithMap(m *sync.Map) manager {
	tm.desktops = m
	return tm
}
