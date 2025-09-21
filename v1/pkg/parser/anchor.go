package parser

import (
	"fmt"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// AnchorRegistry tracks anchor definitions and resolves aliases
type AnchorRegistry struct {
	anchors map[string]node.Node
}

// NewAnchorRegistry creates a new anchor registry
func NewAnchorRegistry() *AnchorRegistry {
	return &AnchorRegistry{
		anchors: make(map[string]node.Node),
	}
}

// RegisterAnchor registers a node with an anchor name
func (r *AnchorRegistry) RegisterAnchor(name string, n node.Node) error {
	if name == "" {
		return fmt.Errorf("anchor name cannot be empty")
	}

	if _, exists := r.anchors[name]; exists {
		return fmt.Errorf("anchor %q already defined", name)
	}

	r.anchors[name] = n
	return nil
}

// ResolveAlias resolves an alias to its anchored node
func (r *AnchorRegistry) ResolveAlias(name string) (node.Node, error) {
	if name == "" {
		return nil, fmt.Errorf("alias name cannot be empty")
	}

	n, exists := r.anchors[name]
	if !exists {
		return nil, fmt.Errorf("undefined alias %q", name)
	}

	// Clone the node to avoid shared references issues
	return cloneNode(n), nil
}

// HasAnchor checks if an anchor is defined
func (r *AnchorRegistry) HasAnchor(name string) bool {
	_, exists := r.anchors[name]
	return exists
}

// Clear removes all anchor definitions
func (r *AnchorRegistry) Clear() {
	r.anchors = make(map[string]node.Node)
}

// cloneNode creates a deep copy of a node
func cloneNode(n node.Node) node.Node {
	if n == nil {
		return nil
	}

	builder := &node.DefaultBuilder{}

	switch v := n.(type) {
	case *node.ScalarNode:
		clone := builder.BuildScalar(v.Value, v.Style)
		// Copy anchor if present
		if anchor := getNodeAnchor(v); anchor != "" {
			setNodeAnchor(clone, anchor)
		}
		return clone

	case *node.SequenceNode:
		items := make([]node.Node, len(v.Items))
		for i, item := range v.Items {
			items[i] = cloneNode(item)
		}
		clone := builder.BuildSequence(items, v.Style)
		if anchor := getNodeAnchor(v); anchor != "" {
			setNodeAnchor(clone, anchor)
		}
		return clone

	case *node.MappingNode:
		pairs := make([]*node.MappingPair, len(v.Pairs))
		for i, pair := range v.Pairs {
			pairs[i] = &node.MappingPair{
				Key:   cloneNode(pair.Key),
				Value: cloneNode(pair.Value),
			}
		}
		clone := builder.BuildMapping(pairs, v.Style)
		if anchor := getNodeAnchor(v); anchor != "" {
			setNodeAnchor(clone, anchor)
		}
		return clone

	default:
		return n // Fallback for unknown types
	}
}

// Helper functions to get/set anchors on nodes
// These would need to be implemented based on how anchors are stored in BaseNode

func getNodeAnchor(n node.Node) string {
	// Get anchor from node using the interface method
	return n.Anchor()
}

func setNodeAnchor(n node.Node, anchor string) {
	// Set anchor on node using the interface method
	n.SetAnchor(anchor)
}

// MergeKey represents the YAML merge key (<<)
const MergeKey = "<<"

// ResolveMergeKeys processes merge keys in a mapping node
func ResolveMergeKeys(n node.Node, registry *AnchorRegistry) error {
	mapping, ok := n.(*node.MappingNode)
	if !ok {
		return nil // Not a mapping, nothing to merge
	}

	var mergedPairs []*node.MappingPair
	var mergeSources []node.Node

	// First pass: collect merge sources and regular pairs
	for _, pair := range mapping.Pairs {
		if scalar, ok := pair.Key.(*node.ScalarNode); ok && scalar.Value == MergeKey {
			// This is a merge key
			switch v := pair.Value.(type) {
			case *node.ScalarNode:
				// Alias reference
				if v.Alias != "" {
					source, err := registry.ResolveAlias(v.Alias)
					if err != nil {
						return fmt.Errorf("merge key alias resolution: %w", err)
					}
					mergeSources = append(mergeSources, source)
				}
			case *node.SequenceNode:
				// Multiple merge sources
				for _, item := range v.Items {
					if scalar, ok := item.(*node.ScalarNode); ok && scalar.Alias != "" {
						source, err := registry.ResolveAlias(scalar.Alias)
						if err != nil {
							return fmt.Errorf("merge key alias resolution: %w", err)
						}
						mergeSources = append(mergeSources, source)
					}
				}
			}
		} else {
			// Regular key-value pair
			mergedPairs = append(mergedPairs, pair)
		}
	}

	// Second pass: merge sources (in order, earlier sources have lower precedence)
	mergedKeys := make(map[string]bool)
	for _, pair := range mergedPairs {
		if scalar, ok := pair.Key.(*node.ScalarNode); ok {
			mergedKeys[scalar.Value] = true
		}
	}

	// Add pairs from merge sources if keys don't exist
	for _, source := range mergeSources {
		if sourceMapping, ok := source.(*node.MappingNode); ok {
			for _, sourcePair := range sourceMapping.Pairs {
				if scalar, ok := sourcePair.Key.(*node.ScalarNode); ok {
					if !mergedKeys[scalar.Value] {
						mergedPairs = append(mergedPairs, sourcePair)
						mergedKeys[scalar.Value] = true
					}
				}
			}
		}
	}

	// Update the mapping with merged pairs
	mapping.Pairs = mergedPairs

	// Recursively resolve merge keys in nested mappings
	for _, pair := range mapping.Pairs {
		if err := ResolveMergeKeys(pair.Value, registry); err != nil {
			return err
		}
	}

	return nil
}
