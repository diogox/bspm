//go:generate mockgen -package subscription -destination ./manager_mock.go -self_package github.com/diogox/bspm/internal/subscription github.com/diogox/bspm/internal/subscription Manager

package subscription

import "sync"

type Topic string

type (
	Manager interface {
		Publish(topic Topic, payload interface{})
		Subscribe(topic Topic) chan interface{}
	}

	manager struct {
		rwMutex       *sync.RWMutex
		subscriptions map[Topic][]chan interface{}
	}
)

func NewManager() *manager {
	return &manager{
		rwMutex:       &sync.RWMutex{},
		subscriptions: make(map[Topic][]chan interface{}),
	}
}

func (m manager) Publish(topic Topic, payload interface{}) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()

	for _, subs := range m.subscriptions[topic] {
		subs <- payload
	}
}

func (m *manager) Subscribe(topic Topic) chan interface{} {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	sub := make(chan interface{}, 1)
	m.subscriptions[topic] = append(m.subscriptions[topic], sub)

	return sub
}
