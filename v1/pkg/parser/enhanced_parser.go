package parser

import (
	"github.com/elioetibr/golang-yaml/v1/pkg/errors"
	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// EnhancedParser implements an improved YAML parser with better comment handling
type EnhancedParser struct {
	lexer             *lexer.Lexer
	current           *lexer.Token
	peek              *lexer.Token
	errors            []*errors.YAMLError
	indentStack       []int
	nodeBuilder       node.Builder
	commentClassifier *CommentClassifier
	anchorRegistry    *AnchorRegistry
	tagResolver       *TagResolver
	lookahead         *ParserLookahead
	options           *ParserOptions

	// Tracking state
	lineTracker       *LineTracker
	sectionDetector   *SectionDetector
	documentComments  []*CommentInfo
}

// ParserOptions configures parser behavior
type ParserOptions struct {
	PreserveComments          bool
	PreserveEmptyLines        bool
	KeepSectionBoundaries     bool
	DefaultLinesBetweenSections int
	AutoDetectSections        bool
	MergeAdjacentComments     bool
}

// DefaultParserOptions returns default parser options
func DefaultParserOptions() *ParserOptions {
	return &ParserOptions{
		PreserveComments:          true,
		PreserveEmptyLines:        true,
		KeepSectionBoundaries:     true,
		DefaultLinesBetweenSections: 1,
		AutoDetectSections:        true,
		MergeAdjacentComments:     true,
	}
}

// NewEnhancedParser creates a new enhanced parser
func NewEnhancedParser(l *lexer.Lexer, opts *ParserOptions) *EnhancedParser {
	if opts == nil {
		opts = DefaultParserOptions()
	}

	lookahead := NewParserLookahead(l)

	return &EnhancedParser{
		lexer:             l,
		indentStack:       []int{0},
		nodeBuilder:       &node.DefaultBuilder{},
		errors:            make([]*errors.YAMLError, 0),
		anchorRegistry:    NewAnchorRegistry(),
		tagResolver:       NewTagResolver(),
		lookahead:         lookahead,
		commentClassifier: NewCommentClassifier(lookahead),
		options:           opts,
		lineTracker:       NewLineTracker(),
		sectionDetector:   NewSectionDetector(),
		documentComments:  make([]*CommentInfo, 0),
	}
}

// Parse parses the YAML document with enhanced comment handling
func (p *EnhancedParser) Parse() (node.Node, error) {
	// Initialize tokens
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Parse document
	root := p.parseDocumentWithComments()

	// Check for errors
	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	// Post-process for sections and formatting
	if p.options.AutoDetectSections && root != nil {
		p.detectAndMarkSections(root)
	}

	return root, nil
}

// advance moves to the next token while tracking comments and empty lines
func (p *EnhancedParser) advance() error {
	p.current = p.peek

	for {
		token, err := p.lexer.NextToken()
		if err != nil {
			return err
		}

		if token == nil {
			p.peek = nil
			break
		}

		// Track line information
		p.lineTracker.TrackToken(token)

		// Handle comments
		if token.Type == lexer.TokenComment {
			p.processComment(token)
			continue
		}

		// Track empty lines
		if token.Type == lexer.TokenNewLine {
			p.lineTracker.TrackEmptyLine(token.Line)
			if p.options.PreserveEmptyLines {
				// May need to track for section boundaries
				p.checkSectionBoundary(token.Line)
			}
			continue
		}

		p.peek = token
		break
	}

	return nil
}

// processComment handles comment tokens
func (p *EnhancedParser) processComment(token *lexer.Token) {
	info := &CommentInfo{
		Token:            token,
		Text:             token.Value,
		Line:             token.Line,
		Column:           token.Column,
		IndentLevel:      token.Column,
		BlankLinesBefore: p.lineTracker.EmptyLinesBefore(token.Line),
		BlankLinesAfter:  0, // Will be updated when we see what follows
	}

	// Check if this is a document-level comment
	if p.commentClassifier.GetCurrentNode() == nil {
		p.documentComments = append(p.documentComments, info)
	} else {
		// Add to pending comments for classification
		p.commentClassifier.AddPendingComment(info)
	}
}

// parseDocumentWithComments parses document with proper comment association
func (p *EnhancedParser) parseDocumentWithComments() node.Node {
	// Create document node
	doc := p.nodeBuilder.CreateDocument()

	// Handle document start marker
	if p.current != nil && p.current.Type == lexer.TokenDocumentStart {
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}
	}

	// Associate document-level comments
	if len(p.documentComments) > 0 {
		for _, comment := range p.documentComments {
			p.associateDocumentComment(doc, comment)
		}
	}

	// Parse the content
	content := p.parseNodeWithComments(0)
	if content != nil {
		// Wrap in document node if needed
		if docNode, ok := content.(*node.DocumentNode); ok {
			return docNode
		}
		// Set content as document child
		if docNode, ok := doc.(*node.DocumentNode); ok {
			docNode.Content = content
		}
	}

	// Handle document end marker
	if p.current != nil && p.current.Type == lexer.TokenDocumentEnd {
		if err := p.advance(); err != nil {
			p.addError(err.Error())
		}
	}

	// Flush any remaining comments
	p.commentClassifier.FlushPendingComments(doc)

	return doc
}

// parseNodeWithComments parses a node with proper comment handling
func (p *EnhancedParser) parseNodeWithComments(indent int) node.Node {
	if p.current == nil || p.current.Type == lexer.TokenEOF {
		return nil
	}

	// Track node in classifier
	defer func() {
		if p.commentClassifier.GetCurrentNode() != nil {
			p.commentClassifier.PopNode()
		}
	}()

	// Handle anchor
	var anchor string
	if p.current.Type == lexer.TokenAnchor {
		anchor = p.current.Value
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}
	}

	// Handle tag
	var tag string
	if p.current != nil && p.current.Type == lexer.TokenTag {
		tag = p.current.Value
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}
	}

	// Parse based on token type
	var result node.Node
	switch {
	case p.isMapping():
		result = p.parseMappingWithComments(indent)
	case p.isSequence():
		result = p.parseSequenceWithComments(indent)
	default:
		result = p.parseScalarWithComments()
	}

	// Set anchor and tag
	if result != nil {
		if anchor != "" {
			result.SetAnchor(anchor)
			p.anchorRegistry.RegisterAnchor(anchor, result)
		}
		if tag != "" {
			result.SetTag(tag)
		}

		// Push to classifier stack
		p.commentClassifier.PushNode(result, indent)

		// Flush pending comments
		p.commentClassifier.FlushPendingComments(result)
	}

	return result
}

// parseMappingWithComments parses a mapping node with comment association
func (p *EnhancedParser) parseMappingWithComments(indent int) node.Node {
	mapping := p.nodeBuilder.CreateMapping()
	mappingNode := mapping.(*node.MappingNode)

	for p.current != nil && p.current.Type != lexer.TokenEOF {
		// Check indentation
		if p.current.Column < indent {
			break
		}

		// Skip if not a key
		if p.current.Type != lexer.TokenKey {
			break
		}

		// Parse key
		keyToken := p.current
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			break
		}

		key := p.parseNodeWithComments(indent + 2)
		if key == nil {
			p.addError("Expected key after ':'")
			break
		}

		// Expect colon
		if p.current == nil || p.current.Type != lexer.TokenColon {
			p.addError("Expected ':' after key")
			break
		}
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			break
		}

		// Parse value
		value := p.parseNodeWithComments(indent + 2)

		// Create pair
		pair := &node.MappingPair{
			Key:   key,
			Value: value,
		}

		// Check for inline comments on same line as value
		p.checkInlineComment(pair)

		mappingNode.Pairs = append(mappingNode.Pairs, pair)
	}

	return mapping
}

// parseSequenceWithComments parses a sequence node with comment association
func (p *EnhancedParser) parseSequenceWithComments(indent int) node.Node {
	sequence := p.nodeBuilder.CreateSequence()
	seqNode := sequence.(*node.SequenceNode)

	for p.current != nil && p.current.Type == lexer.TokenDash {
		// Check indentation
		if p.current.Column < indent {
			break
		}

		if err := p.advance(); err != nil {
			p.addError(err.Error())
			break
		}

		// Parse item
		item := p.parseNodeWithComments(indent + 2)
		if item != nil {
			seqNode.Items = append(seqNode.Items, item)
		}
	}

	return sequence
}

// parseScalarWithComments parses a scalar node with comment association
func (p *EnhancedParser) parseScalarWithComments() node.Node {
	if p.current == nil {
		return nil
	}

	var value string
	var style node.Style

	switch p.current.Type {
	case lexer.TokenScalar:
		value = p.current.Value
		style = node.StylePlain
	case lexer.TokenQuotedScalar:
		value = p.current.Value
		if p.current.Value[0] == '\'' {
			style = node.StyleSingleQuoted
		} else {
			style = node.StyleDoubleQuoted
		}
	case lexer.TokenLiteral:
		value = p.current.Value
		style = node.StyleLiteral
	case lexer.TokenFolded:
		value = p.current.Value
		style = node.StyleFolded
	default:
		return nil
	}

	scalar := p.nodeBuilder.CreateScalar(value)
	if scalarNode, ok := scalar.(*node.ScalarNode); ok {
		scalarNode.Style = style
	}

	if err := p.advance(); err != nil {
		p.addError(err.Error())
	}

	return scalar
}

// Helper methods

func (p *EnhancedParser) isMapping() bool {
	if p.current == nil {
		return false
	}
	// Look for key pattern
	return p.current.Type == lexer.TokenKey ||
		(p.peek != nil && p.peek.Type == lexer.TokenColon)
}

func (p *EnhancedParser) isSequence() bool {
	return p.current != nil && p.current.Type == lexer.TokenDash
}

func (p *EnhancedParser) checkSectionBoundary(line int) {
	emptyLines := p.lineTracker.ConsecutiveEmptyLines(line)
	if emptyLines >= 2 {
		// Mark as section boundary
		p.sectionDetector.MarkSectionBoundary(line)
	}
}

func (p *EnhancedParser) checkInlineComment(pair *node.MappingPair) {
	// Implementation for checking inline comments
	// This would look at the current line for trailing comments
}

func (p *EnhancedParser) associateDocumentComment(doc node.Node, comment *CommentInfo) {
	if docNode, ok := doc.(*node.DocumentNode); ok {
		// Create comment group for document
		if docNode.HeadComment == nil {
			docNode.HeadComment = &node.CommentGroup{
				Comments:         []string{comment.Text},
				BlankLinesBefore: comment.BlankLinesBefore,
				Format:           node.CommentFormat{PreserveSpacing: true},
			}
		} else {
			docNode.HeadComment.Comments = append(docNode.HeadComment.Comments, comment.Text)
		}
	}
}

func (p *EnhancedParser) detectAndMarkSections(root node.Node) {
	// Walk the tree and detect sections based on comment patterns and empty lines
	visitor := &SectionMarkingVisitor{
		detector: p.sectionDetector,
		options:  p.options,
	}
	root.Accept(visitor)
}

func (p *EnhancedParser) addError(msg string) {
	line := 0
	column := 0
	if p.current != nil {
		line = p.current.Line
		column = p.current.Column
	}

	p.errors = append(p.errors, &errors.YAMLError{
		Message: msg,
		Line:    line,
		Column:  column,
	})
}