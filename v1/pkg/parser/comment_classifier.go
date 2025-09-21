package parser

import (
	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// CommentContext tracks the state for comment classification
type CommentContext struct {
	CurrentIndent    int
	ParentNode      node.Node
	PendingComments []*CommentInfo
	LastTokenType   lexer.TokenType
	NodeStack       []node.Node
	IndentStack     []int
}

// CommentInfo stores comment metadata for classification
type CommentInfo struct {
	Token            *lexer.Token
	Text             string
	Line             int
	Column           int
	BlankLinesBefore int
	BlankLinesAfter  int
	IndentLevel      int
}

// CommentClassifier determines the ownership of comments
type CommentClassifier struct {
	context          *CommentContext
	commentProcessor *node.CommentProcessor
	lookahead        TokenLookahead
}

// TokenLookahead provides lookahead capability for the classifier
type TokenLookahead interface {
	PeekNext() *lexer.Token
	PeekNextNonEmpty() *lexer.Token
	CountEmptyLinesBefore(line int) int
	CountEmptyLinesAfter(line int) int
}

// NewCommentClassifier creates a new comment classifier
func NewCommentClassifier(lookahead TokenLookahead) *CommentClassifier {
	return &CommentClassifier{
		context: &CommentContext{
			PendingComments: make([]*CommentInfo, 0),
			NodeStack:       make([]node.Node, 0),
			IndentStack:     []int{0},
		},
		commentProcessor: node.NewCommentProcessor(),
		lookahead:        lookahead,
	}
}

// ClassifyComment determines the ownership of a comment
func (cc *CommentClassifier) ClassifyComment(comment *lexer.Token) CommentOwnership {
	if comment == nil || comment.Type != lexer.TokenComment {
		return CommentOwnershipOrphan
	}

	// Create comment info
	info := &CommentInfo{
		Token:            comment,
		Text:             comment.Value,
		Line:             comment.Line,
		Column:           comment.Column,
		IndentLevel:      comment.Column,
		BlankLinesBefore: cc.lookahead.CountEmptyLinesBefore(comment.Line),
		BlankLinesAfter:  cc.lookahead.CountEmptyLinesAfter(comment.Line),
	}

	// Check what comes next
	nextToken := cc.lookahead.PeekNextNonEmpty()

	// Document-level comments (before any keys)
	if len(cc.context.NodeStack) == 0 && cc.isDocumentComment(info) {
		return CommentOwnershipDocument
	}

	// Check if this is an inline comment (same line as a key/value)
	if cc.isInlineComment(info, nextToken) {
		return CommentOwnershipInline
	}

	// Check if comment belongs to the next key (header comment)
	if cc.isHeaderComment(info, nextToken) {
		return CommentOwnershipHeader
	}

	// Check if comment belongs to current node (footer comment)
	if cc.isFooterComment(info, nextToken) {
		return CommentOwnershipFooter
	}

	// Section-level comment
	if cc.isSectionComment(info) {
		return CommentOwnershipSection
	}

	return CommentOwnershipOrphan
}

// isDocumentComment checks if comment is at document level
func (cc *CommentClassifier) isDocumentComment(info *CommentInfo) bool {
	// Comments before any structural elements
	return len(cc.context.NodeStack) == 0 && info.Column == 0
}

// isInlineComment checks if comment is on the same line as a node
func (cc *CommentClassifier) isInlineComment(info *CommentInfo, nextToken *lexer.Token) bool {
	if len(cc.context.NodeStack) == 0 {
		return false
	}

	currentNode := cc.context.NodeStack[len(cc.context.NodeStack)-1]

	// Check if comment is on the same line as the node
	switch n := currentNode.(type) {
	case *node.ScalarNode:
		return n.Line() == info.Line
	case *node.MappingNode:
		// Check last pair
		if len(n.Pairs) > 0 {
			lastPair := n.Pairs[len(n.Pairs)-1]
			return lastPair.Value != nil && lastPair.Value.Line() == info.Line
		}
	}

	return false
}

// isHeaderComment checks if comment belongs to the next key
func (cc *CommentClassifier) isHeaderComment(info *CommentInfo, nextToken *lexer.Token) bool {
	if nextToken == nil {
		return false
	}

	// Comment appears before a key at the same or deeper indentation
	if nextToken.Type == lexer.TokenKey || nextToken.Type == lexer.TokenScalar {
		// Check indentation relationship
		if nextToken.Column >= info.IndentLevel {
			// Comment is followed by a key - it's a header comment
			return true
		}
	}

	// Comment followed by empty lines then a key
	if info.BlankLinesAfter > 0 && nextToken.Type == lexer.TokenKey {
		return true
	}

	return false
}

// isFooterComment checks if comment belongs to the current node
func (cc *CommentClassifier) isFooterComment(info *CommentInfo, nextToken *lexer.Token) bool {
	if len(cc.context.NodeStack) == 0 {
		return false
	}

	// Comment after a value, before empty line or next key at same/shallower level
	if nextToken == nil || nextToken.Type == lexer.TokenEOF {
		return true
	}

	// If next token is at shallower indentation, this is a footer
	if nextToken.Column < info.IndentLevel {
		return true
	}

	// Multiple empty lines after comment suggests footer
	if info.BlankLinesAfter >= 2 {
		return true
	}

	return false
}

// isSectionComment checks if comment marks a section boundary
func (cc *CommentClassifier) isSectionComment(info *CommentInfo) bool {
	// Section comments typically have blank lines before them
	// and are at root indentation
	return info.BlankLinesBefore >= 2 && info.IndentLevel == 0
}

// AddPendingComment adds a comment to the pending queue
func (cc *CommentClassifier) AddPendingComment(comment *CommentInfo) {
	cc.context.PendingComments = append(cc.context.PendingComments, comment)
}

// FlushPendingComments assigns all pending comments to the appropriate node
func (cc *CommentClassifier) FlushPendingComments(targetNode node.Node) {
	if targetNode == nil || len(cc.context.PendingComments) == 0 {
		return
	}

	for _, comment := range cc.context.PendingComments {
		ownership := cc.ClassifyComment(comment.Token)
		cc.associateComment(targetNode, comment, ownership)
	}

	// Clear pending comments
	cc.context.PendingComments = nil
}

// associateComment associates a comment with a node based on ownership
func (cc *CommentClassifier) associateComment(targetNode node.Node, comment *CommentInfo, ownership CommentOwnership) {
	position := cc.ownershipToPosition(ownership)

	// Detect section type from comment
	sectionType := cc.commentProcessor.DetectSectionType(comment.Text)

	// Associate the comment
	cc.commentProcessor.AssociateCommentToNode(
		targetNode,
		comment.Text,
		position,
		comment.BlankLinesBefore,
	)

	// Update node's section if this is a section comment
	if ownership == CommentOwnershipSection && targetNode.Section() == nil {
		section := &node.Section{
			ID:          cc.generateSectionID(comment.Text),
			Type:        sectionType,
			Title:       cc.extractTitle(comment.Text),
			StartLine:   comment.Line,
			IndentLevel: comment.IndentLevel,
		}
		targetNode.SetSection(section)
	}
}

// ownershipToPosition converts ownership to comment position
func (cc *CommentClassifier) ownershipToPosition(ownership CommentOwnership) node.CommentPosition {
	switch ownership {
	case CommentOwnershipDocument:
		return node.CommentPositionSection
	case CommentOwnershipHeader:
		return node.CommentPositionAbove
	case CommentOwnershipInline:
		return node.CommentPositionInline
	case CommentOwnershipFooter:
		return node.CommentPositionBelow
	case CommentOwnershipSection:
		return node.CommentPositionSection
	default:
		return node.CommentPositionAbove
	}
}

// generateSectionID creates a section ID from comment text
func (cc *CommentClassifier) generateSectionID(text string) string {
	// Delegate to comment processor
	return cc.commentProcessor.ProcessComment(text, node.CommentPositionSection, node.SectionTypeGeneric)
}

// extractTitle extracts a title from comment text
func (cc *CommentClassifier) extractTitle(text string) string {
	// Simple title extraction - can be enhanced
	if len(text) > 2 && text[0] == '#' {
		return text[1:]
	}
	return text
}

// PushNode adds a node to the context stack
func (cc *CommentClassifier) PushNode(node node.Node, indent int) {
	cc.context.NodeStack = append(cc.context.NodeStack, node)
	cc.context.IndentStack = append(cc.context.IndentStack, indent)
	cc.context.CurrentIndent = indent
}

// PopNode removes the top node from the context stack
func (cc *CommentClassifier) PopNode() node.Node {
	if len(cc.context.NodeStack) == 0 {
		return nil
	}

	n := len(cc.context.NodeStack) - 1
	node := cc.context.NodeStack[n]
	cc.context.NodeStack = cc.context.NodeStack[:n]

	if len(cc.context.IndentStack) > 1 {
		cc.context.IndentStack = cc.context.IndentStack[:len(cc.context.IndentStack)-1]
		cc.context.CurrentIndent = cc.context.IndentStack[len(cc.context.IndentStack)-1]
	}

	return node
}

// GetCurrentNode returns the current node being processed
func (cc *CommentClassifier) GetCurrentNode() node.Node {
	if len(cc.context.NodeStack) == 0 {
		return nil
	}
	return cc.context.NodeStack[len(cc.context.NodeStack)-1]
}

// CommentOwnership represents the ownership of a comment
type CommentOwnership int

const (
	CommentOwnershipOrphan CommentOwnership = iota
	CommentOwnershipDocument
	CommentOwnershipHeader
	CommentOwnershipInline
	CommentOwnershipFooter
	CommentOwnershipSection
)

func (co CommentOwnership) String() string {
	switch co {
	case CommentOwnershipDocument:
		return "Document"
	case CommentOwnershipHeader:
		return "Header"
	case CommentOwnershipInline:
		return "Inline"
	case CommentOwnershipFooter:
		return "Footer"
	case CommentOwnershipSection:
		return "Section"
	default:
		return "Orphan"
	}
}