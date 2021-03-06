package topic

import "github.com/diogox/bspm/internal/subscription"

const (
	MonocleEnabled             subscription.Topic = "monocle_enabled"
	MonocleDisabled            subscription.Topic = "monocle_disabled"
	MonocleStateChanged        subscription.Topic = "monocle_state_changed"
	MonocleDesktopFocusChanged subscription.Topic = "monocle_focused_desktop_changed"
)
