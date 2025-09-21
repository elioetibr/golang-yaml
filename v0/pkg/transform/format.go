package transform

import (
	node2 "github.com/elioetibr/golang-yaml/v0/pkg/node"
)

// Formatter handles formatting of YAML nodes
type Formatter struct {
	config *FormatConfig
}

// FormatConfig configures formatting behavior
type FormatConfig struct {
	// Blank line configuration
	DefaultBlankLinesBeforeComment int  // Default blank lines before comments
	ForceBlankLines                bool // Force blank line count
	PreserveOriginal               bool // Preserve original formatting

	// Special rules for different comment positions
	BlankLinesBeforeHeadComment   int
	BlankLinesBeforeKeyComment    int
	BlankLinesBeforeValueComment  int
	BlankLinesBeforeInlineComment int // Usually 0 for inline

	// Section markers get extra spacing
	SectionMarkerExtraLines int
	SectionMarkers          []string // Patterns that mark sections
}

// DefaultFormatConfig returns default formatting configuration
func DefaultFormatConfig() *FormatConfig {
	return &FormatConfig{
		DefaultBlankLinesBeforeComment: 1,
		ForceBlankLines:                false,
		PreserveOriginal:               true,
		BlankLinesBeforeHeadComment:    1,
		BlankLinesBeforeKeyComment:     1,
		BlankLinesBeforeValueComment:   0,
		BlankLinesBeforeInlineComment:  0,
		SectionMarkerExtraLines:        1,
		SectionMarkers: []string{
			"^#\\s*---",      // Section dividers
			"^#\\s*Section:", // Section headers
			"^#\\s*SECTION:", // Uppercase sections
		},
	}
}

// NewFormatter creates a new formatter with given configuration
func NewFormatter(config *FormatConfig) *Formatter {
	if config == nil {
		config = DefaultFormatConfig()
	}
	return &Formatter{config: config}
}

// Format applies formatting to a node
func (f *Formatter) Format(n node2.Node) node2.Node {
	switch typed := n.(type) {
	case *node2.MappingNode:
		return f.formatMapping(typed)
	case *node2.SequenceNode:
		return f.formatSequence(typed)
	case *node2.ScalarNode:
		return f.formatScalar(typed)
	default:
		return n
	}
}

// formatMapping formats a mapping node
func (f *Formatter) formatMapping(m *node2.MappingNode) *node2.MappingNode {
	// Format the mapping's own comments
	m.BaseNode = f.formatBaseNode(m.BaseNode, false)

	// Track if the previous item ended with blank lines to avoid duplication
	prevHadBlankLines := false

	// Format each pair
	for i, pair := range m.Pairs {
		// Check if we already have blank lines from the previous pair
		currentHasBlankLines := pair.BlankLinesBefore > 0

		// Apply blank lines before key comments
		if pair.KeyComment != nil && !f.config.PreserveOriginal {
			desiredBlankLines := f.getBlankLinesForComment(
				pair.KeyComment,
				node2.CommentPositionKey,
			)

			// Only set if we need to change it
			if f.config.ForceBlankLines {
				// If previous had blank lines and we're about to add more, adjust
				if prevHadBlankLines && !currentHasBlankLines {
					desiredBlankLines = max(0, desiredBlankLines-1)
				}
				pair.KeyComment.BlankLinesBefore = desiredBlankLines
			} else if pair.KeyComment.BlankLinesBefore == 0 {
				pair.KeyComment.BlankLinesBefore = desiredBlankLines
			}
		}

		// Apply blank lines before value comments (usually inline, so typically 0)
		if pair.ValueComment != nil && !f.config.PreserveOriginal {
			if f.config.ForceBlankLines || pair.ValueComment.BlankLinesBefore == 0 {
				pair.ValueComment.BlankLinesBefore = f.getBlankLinesForComment(
					pair.ValueComment,
					node2.CommentPositionValue,
				)
			}
		}

		// Format nested nodes
		pair.Key = f.Format(pair.Key)
		pair.Value = f.Format(pair.Value)

		// Track if this pair has blank lines after it
		prevHadBlankLines = pair.BlankLinesAfter > 0 ||
			(pair.KeyComment != nil && pair.KeyComment.BlankLinesBefore > 0)

		// Add blank lines between pairs if configured
		if i > 0 && !f.config.PreserveOriginal {
			if f.shouldAddBlankLineBetweenPairs(m.Pairs[i-1], pair) && !currentHasBlankLines {
				pair.BlankLinesBefore = 1
			}
		}
	}

	return m
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// formatSequence formats a sequence node
func (f *Formatter) formatSequence(s *node2.SequenceNode) *node2.SequenceNode {
	// Format the sequence's own comments
	s.BaseNode = f.formatBaseNode(s.BaseNode, false)

	// Format each item
	for i, item := range s.Items {
		s.Items[i] = f.Format(item)
	}

	return s
}

// formatScalar formats a scalar node
func (f *Formatter) formatScalar(s *node2.ScalarNode) *node2.ScalarNode {
	s.BaseNode = f.formatBaseNode(s.BaseNode, true)
	return s
}

// formatBaseNode formats comments in a base node
func (f *Formatter) formatBaseNode(base node2.BaseNode, isScalar bool) node2.BaseNode {
	if f.config.PreserveOriginal {
		return base
	}

	// Format head comment
	if base.HeadComment != nil {
		if f.config.ForceBlankLines || base.HeadComment.BlankLinesBefore == 0 {
			base.HeadComment.BlankLinesBefore = f.getBlankLinesForComment(
				base.HeadComment,
				node2.CommentPositionAbove,
			)
		}
	}

	// Format line comment (usually inline, so no blank lines)
	if base.LineComment != nil {
		if f.config.ForceBlankLines {
			base.LineComment.BlankLinesBefore = f.config.BlankLinesBeforeInlineComment
		}
	}

	// Format foot comment
	if base.FootComment != nil {
		if f.config.ForceBlankLines || base.FootComment.BlankLinesBefore == 0 {
			base.FootComment.BlankLinesBefore = f.getBlankLinesForComment(
				base.FootComment,
				node2.CommentPositionBelow,
			)
		}
	}

	// Apply default blank lines before the node itself
	if !isScalar && f.config.ForceBlankLines {
		base.BlankLinesBefore = f.config.DefaultBlankLinesBeforeComment
	}

	return base
}

// getBlankLinesForComment determines blank lines for a comment
func (f *Formatter) getBlankLinesForComment(cg *node2.CommentGroup, pos node2.CommentPosition) int {
	if cg == nil || len(cg.Comments) == 0 {
		return 0
	}

	// Check if it's a section marker
	for _, marker := range f.config.SectionMarkers {
		for _, comment := range cg.Comments {
			if matchesPattern(comment, marker) {
				return f.config.DefaultBlankLinesBeforeComment + f.config.SectionMarkerExtraLines
			}
		}
	}

	// Use position-specific configuration
	switch pos {
	case node2.CommentPositionAbove:
		return f.config.BlankLinesBeforeHeadComment
	case node2.CommentPositionKey:
		return f.config.BlankLinesBeforeKeyComment
	case node2.CommentPositionValue:
		return f.config.BlankLinesBeforeValueComment
	case node2.CommentPositionInline:
		return f.config.BlankLinesBeforeInlineComment
	default:
		return f.config.DefaultBlankLinesBeforeComment
	}
}

// shouldAddBlankLineBetweenPairs checks if blank line should be added between pairs
func (f *Formatter) shouldAddBlankLineBetweenPairs(prev, curr *node2.MappingPair) bool {
	// Add blank line if current pair has a key comment
	if curr.KeyComment != nil && len(curr.KeyComment.Comments) > 0 {
		return true
	}

	// Add blank line if previous pair has a foot comment
	if prevValue, ok := prev.Value.(*node2.ScalarNode); ok {
		if prevValue.FootComment != nil && len(prevValue.FootComment.Comments) > 0 {
			return true
		}
	}

	return false
}

// matchesPattern is a simple pattern matcher (could use regex)
func matchesPattern(text, pattern string) bool {
	// Simple implementation - could be enhanced with regex
	return false
}

// FormatWithSorting formats and sorts a node
func FormatWithSorting(n node2.Node, sortConfig *SortConfig, formatConfig *FormatConfig) node2.Node {
	// First format
	formatter := NewFormatter(formatConfig)
	formatted := formatter.Format(n)

	// Then sort (sorting preserves the formatting)
	sorter := NewSorter(sortConfig)
	sorted := sorter.Sort(formatted)

	// Optionally format again to ensure consistency
	return formatter.Format(sorted)
}
