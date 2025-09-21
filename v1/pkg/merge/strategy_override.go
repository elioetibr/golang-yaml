package merge

import (
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// OverrideStrategy completely replaces base with override
type OverrideStrategy struct{}

// NewOverrideStrategy creates a new override strategy
func NewOverrideStrategy() *OverrideStrategy {
	return &OverrideStrategy{}
}

// Name returns the strategy name
func (s *OverrideStrategy) Name() string {
	return "override"
}

// Merge returns the override node, completely replacing base
func (s *OverrideStrategy) Merge(base, override node.Node, ctx *Context) (node.Node, error) {
	if override == nil {
		return base, nil
	}
	// Simply return the override, ignoring base
	return override, nil
}
