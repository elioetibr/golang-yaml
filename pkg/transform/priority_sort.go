package transform

import (
	"github.com/elioetibr/golang-yaml/pkg/node"
	"sort"
)

// PrioritySorter sorts with priority for specific keys
type PrioritySorter struct {
	config      *SortConfig
	priorityMap map[string]int
}

// NewPrioritySorter creates a new priority-based sorter
func NewPrioritySorter(config *SortConfig) *PrioritySorter {
	// Build priority map for quick lookup
	priorityMap := make(map[string]int)
	for i, key := range config.Priority {
		priorityMap[key] = i
	}

	return &PrioritySorter{
		config:      config,
		priorityMap: priorityMap,
	}
}

// Sort sorts a node tree with priority ordering
func (ps *PrioritySorter) Sort(root node.Node) node.Node {
	switch n := root.(type) {
	case *node.MappingNode:
		return ps.sortMapping(n)
	case *node.SequenceNode:
		// For sequences, use regular sorting
		return NewSorter(ps.config).sortSequence(n)
	default:
		return root
	}
}

// sortMapping sorts a mapping node with priority ordering
func (ps *PrioritySorter) sortMapping(mapping *node.MappingNode) *node.MappingNode {
	if len(mapping.Pairs) == 0 {
		return mapping
	}

	// Create sortable pairs with priority
	type priorityPair struct {
		pair     *node.MappingPair
		key      string
		priority int
		original int
	}

	sortablePairs := make([]*priorityPair, len(mapping.Pairs))

	for i, pair := range mapping.Pairs {
		key := ps.getKeyString(pair.Key)
		priority := ps.getPriority(key)

		sortablePairs[i] = &priorityPair{
			pair:     pair,
			key:      key,
			priority: priority,
			original: i,
		}
	}

	// Sort with priority, then by mode
	sort.SliceStable(sortablePairs, func(i, j int) bool {
		// First, sort by priority (lower priority number comes first)
		if sortablePairs[i].priority != sortablePairs[j].priority {
			// -1 means not in priority list, should go to end
			if sortablePairs[i].priority == -1 {
				return false
			}
			if sortablePairs[j].priority == -1 {
				return true
			}
			return sortablePairs[i].priority < sortablePairs[j].priority
		}

		// If same priority (both -1 or same priority level), sort by mode
		if sortablePairs[i].priority == -1 && sortablePairs[j].priority == -1 {
			switch ps.config.Mode {
			case SortModeAscending:
				return sortablePairs[i].key < sortablePairs[j].key
			case SortModeDescending:
				return sortablePairs[i].key > sortablePairs[j].key
			default: // SortModeKeepOriginal
				return sortablePairs[i].original < sortablePairs[j].original
			}
		}

		// Same priority level, maintain original order
		return sortablePairs[i].original < sortablePairs[j].original
	})

	// Rebuild pairs in sorted order
	newPairs := make([]*node.MappingPair, len(sortablePairs))
	for i, sp := range sortablePairs {
		newPairs[i] = sp.pair

		// Recursively sort nested structures if needed
		if ps.config.Scope == SortScopeNested {
			newPairs[i].Value = ps.Sort(newPairs[i].Value)
		}
	}

	// Create new mapping with sorted pairs
	return &node.MappingNode{
		BaseNode: mapping.BaseNode,
		Pairs:    newPairs,
		Style:    mapping.Style,
	}
}

// getPriority returns the priority of a key (-1 if not in priority list)
func (ps *PrioritySorter) getPriority(key string) int {
	if priority, ok := ps.priorityMap[key]; ok {
		return priority
	}
	return -1
}

// getKeyString extracts string value from a key node
func (ps *PrioritySorter) getKeyString(n node.Node) string {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return scalar.Value
	}
	return ""
}
