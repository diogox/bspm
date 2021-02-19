package filter

import (
	"fmt"

	"github.com/diogox/bspc-go"
)

type NodeFilter string

const (
	NodeLocalBiggest NodeFilter = "biggest.local"
	NodeFocused      NodeFilter = "focused"
)

func NodeID(id bspc.ID) NodeFilter {
	return NodeFilter(fmt.Sprintf("%d", id))
}
