package transform

import (
	"sort"
	"strconv"
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/node"
)

// SortMode defines how elements should be sorted
type SortMode int

const (
	// Default: Keep original order - no modifications
	SortModeKeepOriginal SortMode = iota
	SortModeAscending             // Sort in ascending order (A-Z, 0-9)
	SortModeDescending            // Sort in descending order (Z-A, 9-0)
)

// SortBy defines what to sort by
type SortBy int

const (
	SortByKey   SortBy = iota // Sort mappings by their keys
	SortByValue               // Sort by values-with-comments (for sequences or mappings)
)

// SortScope defines the scope of sorting
type SortScope int

const (
	SortScopeDocument SortScope = iota // Sort entire document
	SortScopeSection                   // Sort specific sections
	SortScopeNested                    // Sort recursively including nested structures
)

// SortFunction defines a custom comparison function
type SortFunction func(a, b string) bool

// SortConfig configures sorting behavior
type SortConfig struct {
	Mode     SortMode
	SortBy   SortBy
	Scope    SortScope
	Function SortFunction

	// Advanced options
	CaseSensitive   bool       // Case sensitive comparison
	NumericSort     bool       // Treat strings as numbers if possible
	Priority        []string   // Priority order for keys (first in list = highest priority)
	Groups          [][]string // Groups of keys to keep together
	ExcludePatterns []string   // Paths/keys to exclude from sorting
	SectionMarkers  []string   // Comment patterns that mark sections
	StableSort      bool       // Preserve relative order of equal elements

	// Formatting options
	DefaultBlankLines  int  // Default blank lines before comments (default: 1)
	ForceBlankLines    bool // Force blank line count even if original had different
	PreserveBlankLines bool // Preserve original blank lines when false, apply defaults when true

	// Note: Comments ALWAYS move with their associated nodes
	// This is not configurable - it's fundamental behavior
}

// DefaultSortConfig returns a default configuration that preserves original order
func DefaultSortConfig() *SortConfig {
	return &SortConfig{
		Mode:               SortModeKeepOriginal, // DEFAULT: Don't modify order
		SortBy:             SortByKey,
		Scope:              SortScopeDocument,
		CaseSensitive:      true,
		StableSort:         true,
		DefaultBlankLines:  1,     // Default: 1 blank line before comments
		ForceBlankLines:    false, // Don't force by default
		PreserveBlankLines: true,  // Preserve original formatting by default
	}
}

// NewAscendingSortConfig creates config for ascending sort
func NewAscendingSortConfig() *SortConfig {
	config := DefaultSortConfig()
	config.Mode = SortModeAscending
	return config
}

// NewDescendingSortConfig creates config for descending sort
func NewDescendingSortConfig() *SortConfig {
	config := DefaultSortConfig()
	config.Mode = SortModeDescending
	return config
}

// NewSortByKeyConfig creates config for sorting by keys (mappings only)
func NewSortByKeyConfig(mode SortMode) *SortConfig {
	config := DefaultSortConfig()
	config.Mode = mode
	config.SortBy = SortByKey
	return config
}

// NewSortByValueConfig creates config for sorting by values-with-comments (sequences and mappings)
func NewSortByValueConfig(mode SortMode) *SortConfig {
	config := DefaultSortConfig()
	config.Mode = mode
	config.SortBy = SortByValue
	return config
}

// Sorter handles sorting of YAML nodes
type Sorter struct {
	config *SortConfig
}

// NewSorter creates a new sorter with given configuration
func NewSorter(config *SortConfig) *Sorter {
	if config == nil {
		config = DefaultSortConfig()
	}
	return &Sorter{config: config}
}

// Sort sorts a node according to configuration
func (s *Sorter) Sort(n node.Node) node.Node {
	if s.config.Mode == SortModeKeepOriginal {
		return n // DEFAULT: Preserve original order - no modifications
	}

	switch typed := n.(type) {
	case *node.MappingNode:
		return s.sortMapping(typed)
	case *node.SequenceNode:
		return s.sortSequence(typed)
	default:
		return n // Scalars don't need sorting
	}
}

// sortMapping sorts a mapping node by keys or values-with-comments
func (s *Sorter) sortMapping(m *node.MappingNode) *node.MappingNode {
	// Check if this mapping should be excluded
	for _, pattern := range s.config.ExcludePatterns {
		if s.matchesPattern(m, pattern) {
			return m
		}
	}

	// First, transfer inter-field comments to fix comment association
	// Comments between key-value pairs are often attached to the previous value
	// instead of the next key where they logically belong
	pairs := make([]*node.MappingPair, len(m.Pairs))
	copy(pairs, m.Pairs)
	s.transferInterFieldComments(pairs)

	// Create sortable pairs with their comments
	sortablePairs := make([]*sortablePair, len(pairs))
	for i, pair := range pairs {
		var sortValue string
		if s.config.SortBy == SortByKey {
			sortValue = s.getNodeString(pair.Key)
		} else { // SortByValue
			sortValue = s.getNodeString(pair.Value)
		}

		// Create sortable pair - comments are part of the pair
		// They will automatically move with the pair
		sortablePairs[i] = &sortablePair{
			sortValue: sortValue,
			pair:      pair, // This includes KeyComment and ValueComment
			original:  i,
		}
	}

	// Sort based on configuration
	if s.config.StableSort {
		sort.SliceStable(sortablePairs, func(i, j int) bool {
			return s.compare(sortablePairs[i].sortValue, sortablePairs[j].sortValue)
		})
	} else {
		sort.Slice(sortablePairs, func(i, j int) bool {
			return s.compare(sortablePairs[i].sortValue, sortablePairs[j].sortValue)
		})
	}

	// Rebuild pairs in sorted order
	// IMPORTANT: The pairs contain all comment associations (KeyComment, ValueComment)
	// These automatically move with the pair during sorting
	newPairs := make([]*node.MappingPair, len(sortablePairs))
	for i, sp := range sortablePairs {
		newPairs[i] = sp.pair // Comments stay with their pair

		// Recursively sort nested structures if needed
		if s.config.Scope == SortScopeNested {
			newPairs[i].Value = s.Sort(newPairs[i].Value)
		}
	}

	// Create new mapping with sorted pairs
	// The BaseNode contains HeadComment, LineComment, FootComment for the mapping itself
	result := &node.MappingNode{
		BaseNode: m.BaseNode, // Preserves mapping-level comments
		Pairs:    newPairs,
		Style:    m.Style,
	}

	return result
}

// sortSequence sorts a sequence node's values-with-comments
func (s *Sorter) sortSequence(seq *node.SequenceNode) *node.SequenceNode {
	// Sequences only have values-with-comments (no keys), so only sort if SortBy == SortByValue
	if s.config.SortBy != SortByValue {
		// Don't sort the sequence itself, but may need to sort nested structures
		if s.config.Scope == SortScopeNested {
			for i, item := range seq.Items {
				seq.Items[i] = s.Sort(item)
			}
		}
		return seq
	}

	// Create sortable items
	sortableItems := make([]*sortableItem, len(seq.Items))
	for i, item := range seq.Items {
		value := s.getNodeString(item)
		sortableItems[i] = &sortableItem{
			value:    value,
			node:     item,
			original: i,
		}
	}

	// Sort based on configuration
	if s.config.StableSort {
		sort.SliceStable(sortableItems, func(i, j int) bool {
			return s.compare(sortableItems[i].value, sortableItems[j].value)
		})
	} else {
		sort.Slice(sortableItems, func(i, j int) bool {
			return s.compare(sortableItems[i].value, sortableItems[j].value)
		})
	}

	// Rebuild items in sorted order
	newItems := make([]node.Node, len(sortableItems))
	for i, si := range sortableItems {
		newItems[i] = si.node

		// Recursively sort nested structures if needed
		if s.config.Scope == SortScopeNested {
			newItems[i] = s.Sort(newItems[i])
		}
	}

	// Create new sequence with sorted items
	result := &node.SequenceNode{
		BaseNode: seq.BaseNode,
		Items:    newItems,
		Style:    seq.Style,
	}

	return result
}

// compare compares two strings based on configuration
func (s *Sorter) compare(a, b string) bool {
	// Use custom function if provided
	if s.config.Function != nil {
		result := s.config.Function(a, b)
		if s.config.Mode == SortModeDescending {
			return !result
		}
		return result
	}

	// Handle case sensitivity
	if !s.config.CaseSensitive {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}

	// Handle numeric sorting
	if s.config.NumericSort {
		aNum, aErr := strconv.ParseFloat(a, 64)
		bNum, bErr := strconv.ParseFloat(b, 64)
		if aErr == nil && bErr == nil {
			if s.config.Mode == SortModeDescending {
				return aNum > bNum
			}
			return aNum < bNum
		}
	}

	// Default string comparison
	if s.config.Mode == SortModeDescending {
		return a > b
	}
	return a < b
}

// transferInterFieldComments moves comments from value nodes to the next key node
// This fixes the parser's comment association where comments between pairs
// get attached to the previous value instead of the next key
func (s *Sorter) transferInterFieldComments(pairs []*node.MappingPair) {
	for i := 0; i < len(pairs)-1; i++ {
		// Check if the value has a head comment that might belong to the next key
		if scalar, ok := pairs[i].Value.(*node.ScalarNode); ok {
			if scalar.HeadComment != nil && len(scalar.HeadComment.Comments) > 0 {
				// Transfer the comment to the next key if it doesn't have one
				if nextKey, ok := pairs[i+1].Key.(*node.ScalarNode); ok {
					if nextKey.HeadComment == nil {
						nextKey.HeadComment = scalar.HeadComment
						scalar.HeadComment = nil
					}
				}
			}
		}
	}
}

// Helper functions

func (s *Sorter) getKeyString(n node.Node) string {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return scalar.Value
	}
	return ""
}

func (s *Sorter) getNodeString(n node.Node) string {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return scalar.Value
	}
	return ""
}

func (s *Sorter) matchesPattern(m *node.MappingNode, pattern string) bool {
	// Simple implementation - could be enhanced
	return false
}

// sortablePair holds a mapping pair with its sort value
type sortablePair struct {
	sortValue string
	pair      *node.MappingPair
	original  int
}

// sortableItem holds a sequence item with its sort value
type sortableItem struct {
	value    string
	node     node.Node
	original int
}

// Custom sort functions

// AlphabeticalSort sorts alphabetically
func AlphabeticalSort(a, b string) bool {
	return a < b
}

// NumericSort sorts numerically if possible
func NumericSort(a, b string) bool {
	aNum, aErr := strconv.ParseFloat(a, 64)
	bNum, bErr := strconv.ParseFloat(b, 64)
	if aErr == nil && bErr == nil {
		return aNum < bNum
	}
	return a < b
}

// SemanticVersionSort sorts semantic versions (e.g., 1.2.3)
func SemanticVersionSort(a, b string) bool {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aNum, _ := strconv.Atoi(aParts[i])
		bNum, _ := strconv.Atoi(bParts[i])
		if aNum != bNum {
			return aNum < bNum
		}
	}

	return len(aParts) < len(bParts)
}
