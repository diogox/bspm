package subscription

func (m manager) WithSubs(newSubs map[Topic][]chan interface{}) *manager {
	m.subscriptions = newSubs
	return &m
}
