package parser

import (
	"strconv"
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/errors"
	"github.com/elioetibr/golang-yaml/pkg/lexer"
	"github.com/elioetibr/golang-yaml/pkg/node"
)

// Parser implements a recursive descent parser for YAML
type Parser struct {
	lexer          *lexer.Lexer
	current        *lexer.Token
	peek           *lexer.Token
	errors         []*errors.YAMLError
	indentStack    []int
	inFlow         int
	nodeBuilder    node.Builder
	commentQueue   []*lexer.Token
	anchorRegistry *AnchorRegistry
	tagResolver    *TagResolver
	inMergeKey     bool
}

// NewParser creates a new parser instance
func NewParser(l *lexer.Lexer) *Parser {
	return &Parser{
		lexer:          l,
		indentStack:    []int{0},
		nodeBuilder:    &node.DefaultBuilder{},
		errors:         make([]*errors.YAMLError, 0),
		anchorRegistry: NewAnchorRegistry(),
		tagResolver:    NewTagResolver(),
	}
}

// Parse parses the input and returns the root node
func (p *Parser) Parse() (node.Node, error) {
	// Initialize by reading the first two tokens
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	root := p.parseDocumentNode()

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	return root, nil
}

// advance moves to the next token
func (p *Parser) advance() error {
	if p.peek == nil && p.current != nil {
		// Only set current to nil if we're at EOF
		if p.current.Type == lexer.TokenEOF {
			p.current = nil
			return nil
		}
	}
	p.current = p.peek

	for {
		token, err := p.lexer.NextToken()
		if err != nil {
			return err
		}

		// Check for nil token
		if token == nil {
			p.peek = nil
			break
		}

		// Queue comments for later association
		if token.Type == lexer.TokenComment {
			p.commentQueue = append(p.commentQueue, token)
			continue // Skip comments for now
		}

		p.peek = token
		break
	}

	return nil
}

// parseDocumentNode parses a YAML document node
func (p *Parser) parseDocumentNode() node.Node {
	// Handle document markers
	if p.current != nil && p.current.Type == lexer.TokenDocumentStart {
		p.advance()
	}

	// Parse the document content
	root := p.parseNode(0)

	// Associate any leading comments with the root node
	if root != nil && len(p.commentQueue) > 0 {
		p.associateComments(root)
	}

	// Handle document end marker
	if p.current != nil && p.current.Type == lexer.TokenDocumentEnd {
		p.advance()
	}

	// Resolve merge keys after parsing
	if root != nil {
		ResolveMergeKeys(root, p.anchorRegistry)
	}

	return root
}

// parseNode parses any YAML node at the given indentation level
func (p *Parser) parseNode(indent int) node.Node {
	if p.current == nil || p.current.Type == lexer.TokenEOF {
		return nil
	}

	// Don't process pending comments here - they should be handled elsewhere

	// Check for anchor definition
	var anchor string
	anchorLine := 0
	if p.current.Type == lexer.TokenAnchor {
		anchor = p.current.Value
		anchorLine = p.current.Line
		p.advance() // skip anchor token
	}

	// Check for tag
	var tag string
	if p.current.Type == lexer.TokenTag {
		tag = p.current.Value
		p.advance() // skip tag token
	}

	// After processing anchor/tag, if we're on a new line, use the current indentation
	if p.current != nil && anchorLine > 0 && p.current.Line > anchorLine {
		indent = p.current.Column
	}

	// Check for alias reference
	if p.current.Type == lexer.TokenAlias {
		alias := p.current.Value
		p.advance() // skip alias token

		// For merge keys, keep the alias unresolved
		// It will be resolved later by ResolveMergeKeys
		if p.inMergeKey {
			scalarNode := &node.ScalarNode{
				Value: "",
				Style: node.StylePlain,
				Alias: alias,
			}
			return scalarNode
		}

		// Resolve the alias
		aliasNode, err := p.anchorRegistry.ResolveAlias(alias)
		if err != nil {
			p.addError(err.Error())
			return nil
		}
		return aliasNode
	}

	var n node.Node

	switch p.current.Type {
	case lexer.TokenFlowSequenceStart:
		n = p.parseFlowSequence()
	case lexer.TokenFlowMappingStart:
		n = p.parseFlowMapping()
	case lexer.TokenSequenceEntry:
		n = p.parseBlockSequence(indent)
	case lexer.TokenMappingKey:
		n = p.parseBlockMapping(indent)
	case lexer.TokenPlainScalar, lexer.TokenSingleQuotedScalar,
		lexer.TokenDoubleQuotedScalar, lexer.TokenLiteralScalar,
		lexer.TokenFoldedScalar:
		// Check if this scalar is a key in a mapping
		if p.peek != nil && p.peek.Type == lexer.TokenMappingValue {
			n = p.parseBlockMapping(indent)
		} else {
			n = p.parseScalar()
		}
	default:
		// Try to parse as block mapping (key: value pattern)
		if p.isBlockMappingStart() {
			n = p.parseBlockMapping(indent)
		}
	}

	// Set anchor and tag if present
	if n != nil {
		if anchor != "" {
			n.SetAnchor(anchor)
			// Register the anchor
			if err := p.anchorRegistry.RegisterAnchor(anchor, n); err != nil {
				p.addError(err.Error())
			}
		}
		if tag != "" {
			n.SetTag(tag)
		}
	}

	return n
}

// parseScalar parses a scalar value
func (p *Parser) parseScalar() node.Node {
	if p.current == nil {
		return nil
	}

	var style node.Style
	value := p.current.Value

	switch p.current.Style {
	case lexer.ScalarStylePlain:
		style = node.StylePlain
	case lexer.ScalarStyleSingleQuoted:
		style = node.StyleSingleQuoted
	case lexer.ScalarStyleDoubleQuoted:
		style = node.StyleDoubleQuoted
	case lexer.ScalarStyleLiteral:
		style = node.StyleLiteral
		// Strip common indentation from literal scalars
		value = p.stripBlockScalarIndent(value)
	case lexer.ScalarStyleFolded:
		style = node.StyleFolded
		// Strip common indentation from folded scalars
		value = p.stripBlockScalarIndent(value)
	default:
		style = node.StylePlain
	}

	n := p.nodeBuilder.BuildScalar(value, style)

	// Associate comments
	p.associateComments(n)

	p.advance()
	return n
}

// parseBlockSequence parses a block-style sequence
func (p *Parser) parseBlockSequence(indent int) node.Node {
	items := make([]node.Node, 0)

	for p.current != nil && p.current.Type == lexer.TokenSequenceEntry {
		if p.current.Type == lexer.TokenEOF {
			break
		}

		// Check indentation
		if p.current.Column < indent {
			break
		}

		currentIndent := p.current.Column
		p.advance() // skip '-'

		// Parse the sequence item
		var item node.Node
		if p.current != nil && p.current.Type != lexer.TokenEOF {
			// Check if this is an empty item (next token is another sequence entry at same or less indentation)
			if p.current.Type == lexer.TokenSequenceEntry && p.current.Column <= currentIndent {
				// It's the next item at the same or parent level, so this is an empty item
				item = p.nodeBuilder.BuildScalar("", node.StylePlain)
			} else {
				// Parse the actual content (could be nested or scalar)
				item = p.parseNode(p.current.Column)
			}
		} else {
			// Empty item at EOF
			item = p.nodeBuilder.BuildScalar("", node.StylePlain)
		}
		if item != nil {
			items = append(items, item)
		}
	}

	seq := p.nodeBuilder.BuildSequence(items, node.StyleBlock)
	p.associateComments(seq)
	return seq
}

// parseBlockMapping parses a block-style mapping
func (p *Parser) parseBlockMapping(indent int) node.Node {
	pairs := make([]*node.MappingPair, 0)

	for p.current != nil {
		if p.current.Type == lexer.TokenEOF {
			break
		}

		// Check if we're still in the mapping
		if p.current.Column < indent && indent > 0 {
			break
		}

		// Parse explicit key
		if p.current.Type == lexer.TokenMappingKey {
			p.advance() // skip '?'
			var key node.Node
			if p.current != nil {
				// Parse the key directly as a scalar to avoid recursive mapping detection
				key = p.parseScalar()
			} else {
				key = p.nodeBuilder.BuildScalar("", node.StylePlain)
			}

			// Expect ':' for value
			if p.current != nil && p.current.Type == lexer.TokenMappingValue {
				p.advance() // skip ':'

				// Check for inline comment after colon but before value
				var inlineComment *lexer.Token
				if len(p.commentQueue) > 0 {
					for i, comment := range p.commentQueue {
						if comment.IsInline {
							inlineComment = comment
							// Remove from queue
							p.commentQueue = append(p.commentQueue[:i], p.commentQueue[i+1:]...)
							break
						}
					}
				}

				// Check if this is a merge key
				isMergeKey := false
				if scalarKey, ok := key.(*node.ScalarNode); ok && scalarKey.Value == "<<" {
					isMergeKey = true
					p.inMergeKey = true
				}

				var value node.Node
				if p.current != nil {
					value = p.parseNode(p.current.Column)
				} else {
					value = p.nodeBuilder.BuildScalar("", node.StylePlain)
				}

				// If value is a mapping and we have an inline comment, associate it
				if inlineComment != nil {
					if mapping, ok := value.(*node.MappingNode); ok {
						mapping.LineComment = &node.CommentGroup{
							Comments: []string{inlineComment.Value},
						}
					}
				}

				// Reset merge key flag
				if isMergeKey {
					p.inMergeKey = false
				}

				pairs = append(pairs, &node.MappingPair{Key: key, Value: value})
			}
		} else if p.isBlockMappingStart() {
			// Parse implicit key (plain scalar followed by ':')
			key := p.parseScalar()

			if p.current != nil && p.current.Type == lexer.TokenMappingValue {
				p.advance() // skip ':'

				// Check for inline comment after colon but before value
				var inlineComment *lexer.Token
				if len(p.commentQueue) > 0 {
					// Check if there's an inline comment on the same line as the colon
					for i, comment := range p.commentQueue {
						if comment.IsInline {
							inlineComment = comment
							// Remove from queue
							p.commentQueue = append(p.commentQueue[:i], p.commentQueue[i+1:]...)
							break
						}
					}
				}

				// Check if this is a merge key
				isMergeKey := false
				if scalarKey, ok := key.(*node.ScalarNode); ok && scalarKey.Value == "<<" {
					isMergeKey = true
					p.inMergeKey = true
				}

				var value node.Node
				if p.current != nil {
					value = p.parseNode(p.current.Column)
				} else {
					value = p.nodeBuilder.BuildScalar("", node.StylePlain)
				}

				// If we have an inline comment, associate it with the value
				if inlineComment != nil {
					if mapping, ok := value.(*node.MappingNode); ok {
						// Store the inline comment in the mapping's LineComment
						mapping.LineComment = &node.CommentGroup{
							Comments: []string{inlineComment.Value},
						}
					} else if seq, ok := value.(*node.SequenceNode); ok {
						// Store the inline comment in the sequence's LineComment
						seq.LineComment = &node.CommentGroup{
							Comments: []string{inlineComment.Value},
						}
					} else if scalar, ok := value.(*node.ScalarNode); ok {
						// Store the inline comment in the scalar's LineComment
						scalar.LineComment = &node.CommentGroup{
							Comments: []string{inlineComment.Value},
						}
					}
				}

				// Reset merge key flag
				if isMergeKey {
					p.inMergeKey = false
				}

				pairs = append(pairs, &node.MappingPair{Key: key, Value: value})
			}
		} else {
			break
		}
	}

	mapping := p.nodeBuilder.BuildMapping(pairs, node.StyleBlock)
	p.associateComments(mapping)
	return mapping
}

// parseFlowSequence parses a flow-style sequence [a, b, c]
func (p *Parser) parseFlowSequence() node.Node {
	p.advance() // skip '['
	p.inFlow++

	items := make([]node.Node, 0)

	for p.current != nil && p.current.Type != lexer.TokenFlowSequenceEnd {
		if p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow sequences
			break
		}

		// Parse item
		item := p.parseNode(0)
		if item != nil {
			items = append(items, item)
		}

		// Handle comma
		if p.current != nil && p.current.Type == lexer.TokenFlowEntry {
			p.advance() // skip ','
		}

		// Check for EOF after advance to prevent infinite loop
		if p.current != nil && p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow sequences
			break
		}
	}

	if p.current != nil && p.current.Type == lexer.TokenFlowSequenceEnd {
		p.advance() // skip ']'
	}

	p.inFlow--
	seq := p.nodeBuilder.BuildSequence(items, node.StyleFlow)
	p.associateComments(seq)
	return seq
}

// parseFlowMapping parses a flow-style mapping {a: 1, b: 2}
func (p *Parser) parseFlowMapping() node.Node {
	p.advance() // skip '{'
	p.inFlow++

	pairs := make([]*node.MappingPair, 0)

	for p.current != nil && p.current.Type != lexer.TokenFlowMappingEnd {
		if p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow mappings
			break
		}

		// Skip whitespace/empty
		if p.current.Type == lexer.TokenFlowEntry {
			p.advance()
			// Check for EOF after advance
			if p.current != nil && p.current.Type == lexer.TokenEOF {
				// Be lenient with unclosed flow mappings
				break
			}
			continue
		}

		// Parse key - could be any scalar
		var key node.Node
		switch p.current.Type {
		case lexer.TokenPlainScalar, lexer.TokenSingleQuotedScalar, lexer.TokenDoubleQuotedScalar:
			key = p.parseScalar()
		default:
			// Unexpected token - advance and check for EOF
			p.advance()
			if p.current != nil && p.current.Type == lexer.TokenEOF {
				// Be lenient with unclosed flow mappings
				break
			}
			continue
		}

		// Expect ':'
		if p.current != nil && p.current.Type == lexer.TokenMappingValue {
			p.advance() // skip ':'

			// Parse value - could be any scalar
			var value node.Node
			switch p.current.Type {
			case lexer.TokenPlainScalar, lexer.TokenSingleQuotedScalar, lexer.TokenDoubleQuotedScalar:
				value = p.parseScalar()
			default:
				value = p.nodeBuilder.BuildScalar("", node.StylePlain)
			}

			pairs = append(pairs, &node.MappingPair{Key: key, Value: value})
		}

		// Handle comma - already handled in loop condition
		if p.current != nil && p.current.Type == lexer.TokenFlowEntry {
			p.advance() // skip ','
		}

		// Check for EOF after advance to prevent infinite loop
		if p.current != nil && p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow mappings
			break
		}
	}

	if p.current != nil && p.current.Type == lexer.TokenFlowMappingEnd {
		p.advance() // skip '}'
	}

	p.inFlow--
	mapping := p.nodeBuilder.BuildMapping(pairs, node.StyleFlow)
	p.associateComments(mapping)
	return mapping
}

// Helper methods

func (p *Parser) isBlockMappingStart() bool {
	// Check if current token is a scalar followed by ':'
	if p.current == nil || p.peek == nil {
		return false
	}

	isScalar := p.current.Type == lexer.TokenPlainScalar ||
		p.current.Type == lexer.TokenSingleQuotedScalar ||
		p.current.Type == lexer.TokenDoubleQuotedScalar

	return isScalar && p.peek.Type == lexer.TokenMappingValue
}

func (p *Parser) processPendingComments() {
	// Process queued comments and determine their association
	for _, comment := range p.commentQueue {
		// This is simplified - full implementation would determine
		// comment position relative to nodes
		_ = comment
	}
	p.commentQueue = nil
}

func (p *Parser) associateComments(n node.Node) {
	// Associate pending comments with the node
	if len(p.commentQueue) > 0 && n != nil {
		// Associate all non-inline comments as head comments
		for _, comment := range p.commentQueue {
			if !comment.IsInline {
				node.AssociateComment(n, comment.Value, node.CommentPositionAbove, comment.BlankLinesBefore)
			}
		}

		// Look for inline comments on the same line
		if p.current != nil {
			for _, comment := range p.commentQueue {
				if comment.IsInline && comment.Line == p.current.Line {
					node.AssociateComment(n, comment.Value, node.CommentPositionInline, 0)
				}
			}
		}

		p.commentQueue = nil
	}
}

// ParseString is a convenience method to parse a YAML string
func ParseString(input string) (node.Node, error) {
	l := lexer.NewLexerFromString(input)
	if err := l.Initialize(); err != nil {
		return nil, err
	}

	parser := NewParser(l)
	return parser.Parse()
}

// ParseValue parses a string and converts it to appropriate Go type
func ParseValue(s string) interface{} {
	// Try to parse as number
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try to parse as boolean
	switch s {
	case "true", "True", "TRUE", "yes", "Yes", "YES", "on", "On", "ON":
		return true
	case "false", "False", "FALSE", "no", "No", "NO", "off", "Off", "OFF":
		return false
	case "null", "Null", "NULL", "~", "":
		return nil
	}

	// Return as string
	return s
}

// Errors returns any parsing errors encountered
func (p *Parser) Errors() []*errors.YAMLError {
	return p.errors
}

func (p *Parser) addError(msg string) {
	pos := errors.Position{
		Line:   1,
		Column: 1,
		Offset: 0,
	}

	if p.current != nil {
		pos.Line = p.current.Line
		pos.Column = p.current.Column
		pos.Offset = p.current.Offset
	}

	p.errors = append(p.errors, errors.New(msg, pos, errors.ErrorTypeParser))
}

// stripBlockScalarIndent strips the common indentation from block scalar content
func (p *Parser) stripBlockScalarIndent(value string) string {
	if value == "" {
		return value
	}

	lines := strings.Split(value, "\n")
	if len(lines) == 0 {
		return value
	}

	// Find the minimum indentation (excluding empty lines)
	minIndent := -1
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else {
				break
			}
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// If no non-empty lines, return as is
	if minIndent == -1 {
		return value
	}

	// Strip the common indentation from all lines
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
}
