package transform

import (
	"strings"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// PathFilter handles path-based exclusions for sorting
type PathFilter struct {
	exclusions []string
}

// NewPathFilter creates a new path filter
func NewPathFilter(exclusions []string) *PathFilter {
	return &PathFilter{
		exclusions: exclusions,
	}
}

// ShouldExclude checks if a path should be excluded from sorting
func (pf *PathFilter) ShouldExclude(path string) bool {
	for _, pattern := range pf.exclusions {
		if pf.matchesPattern(path, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a path matches an exclusion pattern
func (pf *PathFilter) matchesPattern(path, pattern string) bool {
	// Support wildcards
	if strings.Contains(pattern, "*") {
		return pf.matchWildcard(path, pattern)
	}

	// Support path prefixes
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern)
	}

	// Exact match
	return path == pattern
}

// matchWildcard matches a path against a wildcard pattern
func (pf *PathFilter) matchWildcard(path, pattern string) bool {
	// Simple wildcard matching
	// * matches any sequence of characters except /
	// ** matches any sequence of characters including /

	if pattern == "**" {
		return true
	}

	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix)
		}
	}

	// Single wildcard matching
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		prefix := parts[0]
		suffix := parts[1]

		if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
			return false
		}

		// Check that the middle part doesn't contain /
		middle := path[len(prefix) : len(path)-len(suffix)]
		return !strings.Contains(middle, "/")
	}

	return false
}

// PathAwareSorter sorts with path-based exclusions
type PathAwareSorter struct {
	config      *SortConfig
	filter      *PathFilter
	baseSorter  *Sorter
	currentPath []string
}

// NewPathAwareSorter creates a new path-aware sorter
func NewPathAwareSorter(config *SortConfig) *PathAwareSorter {
	return &PathAwareSorter{
		config:      config,
		filter:      NewPathFilter(config.ExcludePatterns),
		baseSorter:  NewSorter(config),
		currentPath: []string{},
	}
}

// Sort sorts a node tree with path-based exclusions
func (pas *PathAwareSorter) Sort(root node.Node) node.Node {
	return pas.sortWithPath(root, []string{})
}

// sortWithPath sorts a node at the given path
func (pas *PathAwareSorter) sortWithPath(n node.Node, path []string) node.Node {
	currentPath := strings.Join(path, "/")

	// Check if this path should be excluded
	if pas.filter.ShouldExclude(currentPath) {
		// Return the node as-is without sorting
		return n
	}

	// Not excluded, proceed with normal sorting
	switch node := n.(type) {
	case *node.MappingNode:
		return pas.sortMappingWithPath(node, path)
	case *node.SequenceNode:
		return pas.sortSequenceWithPath(node, path)
	default:
		return n
	}
}

// sortMappingWithPath sorts a mapping node at the given path
func (pas *PathAwareSorter) sortMappingWithPath(mapping *node.MappingNode, path []string) *node.MappingNode {
	// First, recursively process children to handle their exclusions
	processedPairs := make([]*node.MappingPair, len(mapping.Pairs))
	for i, pair := range mapping.Pairs {
		// Get key for path
		key := pas.getKeyString(pair.Key)
		// Create a new slice to avoid modifying the original path
		childPath := make([]string, len(path)+1)
		copy(childPath, path)
		childPath[len(path)] = key

		// Recursively sort value with path
		processedPairs[i] = &node.MappingPair{
			Key:   pair.Key,
			Value: pas.sortWithPath(pair.Value, childPath),
		}
	}

	// Now create a mapping with processed pairs
	processedMapping := &node.MappingNode{
		BaseNode: mapping.BaseNode,
		Pairs:    processedPairs,
		Style:    mapping.Style,
	}

	// Then sort this level
	return pas.baseSorter.sortMapping(processedMapping)
}

// sortSequenceWithPath sorts a sequence node at the given path
func (pas *PathAwareSorter) sortSequenceWithPath(seq *node.SequenceNode, path []string) *node.SequenceNode {
	// First, sort this level if not excluded
	sorted := pas.baseSorter.sortSequence(seq)

	// Then recursively process children
	for i, item := range sorted.Items {
		// Use index for path in sequences
		childPath := make([]string, len(path)+1)
		copy(childPath, path)
		childPath[len(path)] = string(rune(i))
		sorted.Items[i] = pas.sortWithPath(item, childPath)
	}

	return sorted
}

// processExcludedMapping processes a mapping that's excluded from sorting
func (pas *PathAwareSorter) processExcludedMapping(mapping *node.MappingNode, path []string) *node.MappingNode {
	// Don't sort this level, but process children
	result := &node.MappingNode{
		BaseNode: mapping.BaseNode,
		Pairs:    make([]*node.MappingPair, len(mapping.Pairs)),
		Style:    mapping.Style,
	}

	for i, pair := range mapping.Pairs {
		key := pas.getKeyString(pair.Key)
		childPath := make([]string, len(path)+1)
		copy(childPath, path)
		childPath[len(path)] = key

		result.Pairs[i] = &node.MappingPair{
			Key:   pair.Key,
			Value: pas.sortWithPath(pair.Value, childPath),
		}
	}

	return result
}

// processExcludedSequence processes a sequence that's excluded from sorting
func (pas *PathAwareSorter) processExcludedSequence(seq *node.SequenceNode, path []string) *node.SequenceNode {
	// Don't sort this level, but process children
	result := &node.SequenceNode{
		BaseNode: seq.BaseNode,
		Items:    make([]node.Node, len(seq.Items)),
		Style:    seq.Style,
	}

	for i, item := range seq.Items {
		childPath := make([]string, len(path)+1)
		copy(childPath, path)
		childPath[len(path)] = string(rune(i))
		result.Items[i] = pas.sortWithPath(item, childPath)
	}

	return result
}

// getKeyString extracts string value from a key node
func (pas *PathAwareSorter) getKeyString(n node.Node) string {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return scalar.Value
	}
	return ""
}

// SortWithExclusions is a convenience function that sorts with path exclusions
func SortWithExclusions(root node.Node, config *SortConfig) node.Node {
	if len(config.ExcludePatterns) == 0 {
		// No exclusions, use regular sorter
		return NewSorter(config).Sort(root)
	}

	// Use path-aware sorter
	return NewPathAwareSorter(config).Sort(root)
}
