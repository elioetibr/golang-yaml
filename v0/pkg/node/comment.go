package node

import (
	"regexp"
	"strings"
)

// CommentPosition indicates where a comment appears relative to a node
type CommentPosition int

const (
	CommentPositionAbove CommentPosition = iota
	CommentPositionInline
	CommentPositionBelow
	CommentPositionKey
	CommentPositionValue
)

// BlankLineMatcher defines rules for automatic blank line insertion
type BlankLineMatcher struct {
	Pattern     *regexp.Regexp
	BlankLines  int
	Position    CommentPosition
	Description string
}

// CommentManager manages comment formatting and blank line rules
type CommentManager struct {
	matchers           []BlankLineMatcher
	preserveBlankLines bool
	maxBlankLines      int
}

// NewCommentManager creates a new comment manager with default settings
func NewCommentManager() *CommentManager {
	return &CommentManager{
		matchers:           make([]BlankLineMatcher, 0),
		preserveBlankLines: true,
		maxBlankLines:      2,
	}
}

// AddMatcher adds a blank line matcher rule
func (cm *CommentManager) AddMatcher(pattern string, blankLines int, pos CommentPosition, desc string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	cm.matchers = append(cm.matchers, BlankLineMatcher{
		Pattern:     re,
		BlankLines:  blankLines,
		Position:    pos,
		Description: desc,
	})
	return nil
}

// GetBlankLinesForComment determines how many blank lines should precede a comment
func (cm *CommentManager) GetBlankLinesForComment(comment string, pos CommentPosition) int {
	comment = strings.TrimSpace(comment)

	for _, matcher := range cm.matchers {
		if matcher.Position == pos && matcher.Pattern.MatchString(comment) {
			return matcher.BlankLines
		}
	}

	// Default: no extra blank lines
	return 0
}

// FormatCommentGroup formats a comment group with appropriate spacing
func (cm *CommentManager) FormatCommentGroup(cg *CommentGroup, pos CommentPosition) []string {
	if cg == nil || len(cg.Comments) == 0 {
		return nil
	}

	var result []string

	// Add blank lines before if needed
	for i := 0; i < cg.BlankLinesBefore; i++ {
		result = append(result, "")
	}

	// Add comments
	for i, comment := range cg.Comments {
		// Check if matcher applies for additional spacing
		if i == 0 && pos != CommentPositionInline {
			extraLines := cm.GetBlankLinesForComment(comment, pos)
			for j := 0; j < extraLines; j++ {
				result = append(result, "")
			}
		}
		result = append(result, comment)
	}

	return result
}

// AssociateComment associates a comment with a node based on position
func AssociateComment(node Node, comment string, pos CommentPosition, blankLinesBefore int) {
	// Get base node to set comment
	var baseNode *BaseNode
	switch n := node.(type) {
	case *ScalarNode:
		baseNode = &n.BaseNode
	case *SequenceNode:
		baseNode = &n.BaseNode
	case *MappingNode:
		baseNode = &n.BaseNode
	default:
		return
	}

	// Get or create the appropriate comment group
	var cg *CommentGroup
	switch pos {
	case CommentPositionAbove:
		cg = baseNode.HeadComment
	case CommentPositionInline:
		cg = baseNode.LineComment
	case CommentPositionBelow:
		cg = baseNode.FootComment
	}

	// If comment group exists, append to it; otherwise create new
	if cg != nil {
		cg.Comments = append(cg.Comments, comment)
		// Update blank lines if this comment has more
		if blankLinesBefore > cg.BlankLinesBefore {
			cg.BlankLinesBefore = blankLinesBefore
		}
	} else {
		cg = &CommentGroup{
			Comments:         []string{comment},
			BlankLinesBefore: blankLinesBefore,
		}
		// Set the comment group
		switch pos {
		case CommentPositionAbove:
			baseNode.HeadComment = cg
		case CommentPositionInline:
			baseNode.LineComment = cg
		case CommentPositionBelow:
			baseNode.FootComment = cg
		}
	}
}

// MergeCommentGroups merges multiple comment groups
func MergeCommentGroups(groups ...*CommentGroup) *CommentGroup {
	var allComments []string
	maxBlankLines := 0

	for _, g := range groups {
		if g != nil {
			allComments = append(allComments, g.Comments...)
			if g.BlankLinesBefore > maxBlankLines {
				maxBlankLines = g.BlankLinesBefore
			}
		}
	}

	if len(allComments) == 0 {
		return nil
	}

	return &CommentGroup{
		Comments:         allComments,
		BlankLinesBefore: maxBlankLines,
	}
}
