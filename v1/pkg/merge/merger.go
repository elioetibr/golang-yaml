package merge

import (
	"fmt"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// Merger orchestrates the merge process
type Merger struct {
	options   *Options
	strategy  MergeStrategy
	processor *NodeProcessor
}

// NewMerger creates a new merger with the given options
func NewMerger(opts *Options) *Merger {
	if opts == nil {
		opts = DefaultOptions()
	}

	m := &Merger{
		options:   opts,
		processor: NewNodeProcessor(),
	}

	// Set strategy based on options
	switch opts.Strategy {
	case StrategyDeep:
		m.strategy = NewDeepMergeStrategy(m.processor)
	case StrategyShallow:
		m.strategy = NewShallowMergeStrategy(m.processor)
	case StrategyOverride:
		m.strategy = NewOverrideStrategy()
	default:
		m.strategy = NewDeepMergeStrategy(m.processor)
	}

	return m
}

// Merge combines two nodes according to the configured strategy
func (m *Merger) Merge(base, override node.Node) (node.Node, error) {
	// Create context
	ctx := &Context{
		Options: m.options,
		Depth:   0,
		Path:    []string{},
	}

	// Perform merge
	result, err := m.strategy.Merge(base, override, ctx)
	if err != nil {
		return nil, fmt.Errorf("merge failed: %w", err)
	}

	return result, nil
}

// Context carries merge operation context
type Context struct {
	Options *Options
	Depth   int
	Path    []string
}

// WithPath returns a new context with the path appended
func (c *Context) WithPath(segment string) *Context {
	newPath := make([]string, len(c.Path)+1)
	copy(newPath, c.Path)
	newPath[len(c.Path)] = segment

	return &Context{
		Options: c.Options,
		Depth:   c.Depth + 1,
		Path:    newPath,
	}
}
