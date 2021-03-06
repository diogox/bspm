package subscription_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/diogox/bspm/internal/subscription"
)

const testTopic subscription.Topic = "test"

func TestManager_Publish(t *testing.T) {
	t.Run("should publish to subscribers", func(t *testing.T) {
		var (
			sub  = make(chan interface{}, 1)
			subs = map[subscription.Topic][]chan interface{}{
				testTopic: {sub},
			}
			expected = "expected-string"
		)

		subscription.NewManager().WithSubs(subs).Publish(testTopic, expected)

		assert.Equal(t, expected, <-sub)
	})
}

func TestManager_Subscribe(t *testing.T) {
	t.Run("should subscribe to topic", func(t *testing.T) {
		subs := map[subscription.Topic][]chan interface{}{}
		sub := subscription.NewManager().WithSubs(subs).Subscribe(testTopic)

		require.NotEmpty(t, subs)
		require.NotEmpty(t, subs[testTopic])
		assert.Equal(t, subs[testTopic][0], sub)
	})
}
