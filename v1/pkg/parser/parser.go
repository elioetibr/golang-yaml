package parser

import (
	"strconv"
	"strings"

	"github.com/elioetibr/golang-yaml/v1/pkg/errors"
	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
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
	inMappingValue bool // Track if we're currently parsing a mapping value
	emptyLineQueue []*lexer.Token // Track explicit empty line tokens
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

		// Queue comments and empty lines for later association
		if token.Type == lexer.TokenComment {
			p.commentQueue = append(p.commentQueue, token)
			continue // Skip comments for now
		}

		// Queue empty lines for precise blank line tracking
		if token.Type == lexer.TokenEmptyLine {
			p.emptyLineQueue = append(p.emptyLineQueue, token)
			continue // Skip empty lines for now
		}

		p.peek = token
		break
	}

	return nil
}

// findNextKeyColumn looks ahead in the token stream to find the column of the next key at the same or shallower level
func (p *Parser) findNextKeyColumn(startPos int, currentIndent int) int {
	// Look ahead in the input to find the next key at the same or shallower level
	lexerInput := p.lexer.GetInput()
	if lexerInput == "" || startPos >= len(lexerInput) {
		return -1
	}

	// Create a temporary lexer to scan ahead
	tempLexer := lexer.NewLexerFromString(lexerInput[startPos:])
	if err := tempLexer.Initialize(); err != nil {
		return -1
	}

	for {
		token, err := tempLexer.NextToken()
		if err != nil || token == nil || token.Type == lexer.TokenEOF {
			break
		}

		// Look for scalar keys that could be mapping keys
		if token.Type == lexer.TokenPlainScalar || token.Type == lexer.TokenSingleQuotedScalar || token.Type == lexer.TokenDoubleQuotedScalar {
			// Check if this is followed by a colon (making it a key)
			nextToken, _ := tempLexer.NextToken()
			if nextToken != nil && nextToken.Type == lexer.TokenMappingValue {
				// This is a key, check its indentation level
				keyColumn := token.Column + startPos
				if keyColumn <= currentIndent {
					// Found a key at same or shallower level
					return keyColumn
				}
			}
		}
	}

	return -1 // No next key found
}

// parseDocumentNode parses a YAML document node
func (p *Parser) parseDocumentNode() node.Node {
	// Collect document head comments only if they're truly document-level
	// (i.e., followed by document markers or separated by blank lines from content)
	var documentHeadComments []*lexer.Token

	// Check if we have comments that are truly document-level
	// We'll only treat them as document comments if:
	// 1. They're followed by a document marker (---), or
	// 2. They're generic file headers (like yaml-language-server directives)
	if len(p.commentQueue) > 0 {
		// Check if first comment looks like a document header
		isDocumentHeader := false
		if len(p.commentQueue) > 0 {
			firstComment := p.commentQueue[0]
			// Check for common document-level comment patterns
			if strings.Contains(firstComment.Value, "yaml-language-server") ||
				strings.Contains(firstComment.Value, "Default values") ||
				strings.Contains(firstComment.Value, "This is a YAML") {
				isDocumentHeader = true
			}
		}

		// If it looks like a document header, capture only the header comments
		if isDocumentHeader {
			captureIndex := -1
			for i, comment := range p.commentQueue {
				// Stop capturing once we see comments that describe specific fields (starting with --)
				if !strings.Contains(comment.Value, "yaml-language-server") &&
					!strings.Contains(comment.Value, "Default values") &&
					!strings.Contains(comment.Value, "This is a YAML") &&
					!strings.Contains(comment.Value, "Declare variables") &&
					!strings.Contains(comment.Value, "@schema") &&
					!strings.Contains(comment.Value, "enum:") &&
					!strings.Contains(comment.Value, "required:") &&
					!strings.Contains(comment.Value, "additionalProperties:") &&
					!strings.HasPrefix(comment.Value, "# vim:") &&
					!strings.HasPrefix(comment.Value, "# -*- ") &&
					!strings.Contains(comment.Value, " -- ") {
					// This is likely a field comment, stop here
					captureIndex = i
					break
				}
			}

			if captureIndex > 0 {
				documentHeadComments = p.commentQueue[:captureIndex]
				p.commentQueue = p.commentQueue[captureIndex:]
			} else if captureIndex == 0 {
				// No document comments, all are field comments
				// Keep them in queue
			} else {
				// All comments look like document comments
				documentHeadComments = p.commentQueue
				p.commentQueue = nil
			}
		}
	}

	// Handle document markers
	if p.current != nil && p.current.Type == lexer.TokenDocumentStart {
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}
	}

	// Parse the document content
	root := p.parseNode(0)

	// If we have document head comments and a mapping node, attach them as metadata
	if len(documentHeadComments) > 0 && root != nil {
		if mappingNode, ok := root.(*node.MappingNode); ok {
			// Convert comment tokens to CommentGroup
			var comments []string
			var blankLinesWithin []int
			blankLinesBefore := 0

			for i, token := range documentHeadComments {
				comments = append(comments, token.Value)
				// Track blank lines before each comment
				if i == 0 {
					blankLinesBefore = token.BlankLinesBefore
					blankLinesWithin = append(blankLinesWithin, 0) // First comment has no blank lines within group
				} else {
					blankLinesWithin = append(blankLinesWithin, token.BlankLinesBefore)
				}
			}

			if len(comments) > 0 {
				// Store document head comments in the mapping node
				mappingNode.HeadComment = &node.CommentGroup{
					Comments:         comments,
					BlankLinesBefore: blankLinesBefore,
					BlankLinesWithin: blankLinesWithin,
				}
				// Mark that this mapping has document-level comments
				mappingNode.HasDocumentHeadComments = true
			}
		} else if root != nil {
			// For non-mapping root nodes, still try to associate comments
			if len(documentHeadComments) > 0 {
				// Put the comments back in the queue for normal association
				p.commentQueue = append(documentHeadComments, p.commentQueue...)
			}
		}
	}

	// Associate any remaining comments with the root node
	if root != nil && len(p.commentQueue) > 0 {
		p.associateComments(root)
	}

	// Handle document end marker
	if p.current != nil && p.current.Type == lexer.TokenDocumentEnd {
		if err := p.advance(); err != nil {
			p.addError(err.Error())
		}
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
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}
	}

	// Check for tag
	var tag string
	if p.current != nil && p.current.Type == lexer.TokenTag {
		tag = p.current.Value
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}
	}

	// After processing anchor/tag, if we're on a new line, use the current indentation
	if p.current != nil && anchorLine > 0 && p.current.Line > anchorLine {
		indent = p.current.Column
	}

	// Check for alias reference
	if p.current != nil && p.current.Type == lexer.TokenAlias {
		alias := p.current.Value
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			return nil
		}

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
		// Check if this scalar is a key in a mapping (colon on same line)
		if p.peek != nil && p.peek.Type == lexer.TokenMappingValue && p.current.Line == p.peek.Line {
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

	if err := p.advance(); err != nil {
		p.addError(err.Error())
	}
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
		if err := p.advance(); err != nil {
			p.addError(err.Error())
			break
		}

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
	exitedDueToIndent := false

	for p.current != nil {
		if p.current.Type == lexer.TokenEOF {
			break
		}

		// Check if we're still in the mapping BEFORE capturing comments
		if p.current.Column < indent && indent > 0 {
			// Leave comments in queue for parent level to handle
			exitedDueToIndent = true
			break
		}

		// Capture comments that belong to the NEXT key following hierarchical token path logic
		var keyComments []*lexer.Token
		var emptyLines []*lexer.Token

		// Only capture comments that are at the correct hierarchical level
		if len(p.commentQueue) > 0 {
			// Look ahead to determine if there are more keys at the current level
			nextKeyColumn := p.findNextKeyColumn(p.lexer.GetPos() + 1, indent)

			for i, comment := range p.commentQueue {
				// Comments should belong to this key if:
				// 1. They are at or before current key's indentation, AND
				// 2. They are not intended for a deeper nested key
				shouldCapture := false

				if comment.Column <= p.current.Column {
					// Check if this comment is for a child key by looking at the next key
					if nextKeyColumn == -1 || comment.Column <= nextKeyColumn {
						// This comment belongs to the current key
						shouldCapture = true
					}
				}

				if shouldCapture {
					keyComments = append(keyComments, comment)
				} else {
					// Keep remaining comments for deeper levels or next keys
					p.commentQueue = p.commentQueue[i:]
					break
				}
			}
			// If we captured all comments, clear the queue
			if len(keyComments) == len(p.commentQueue) {
				p.commentQueue = nil
			}
		}

		if len(p.emptyLineQueue) > 0 {
			emptyLines = p.emptyLineQueue
			p.emptyLineQueue = nil
		}

		// Parse explicit key
		if p.current.Type == lexer.TokenMappingKey {
			if err := p.advance(); err != nil {
				p.addError(err.Error())
				break
			}
			var key node.Node
			if p.current != nil {
				// Parse the key directly as a scalar to avoid recursive mapping detection
				key = p.parseScalar()
			} else {
				key = p.nodeBuilder.BuildScalar("", node.StylePlain)
			}

			// Associate captured comments and empty lines with the key
			if (len(keyComments) > 0 || len(emptyLines) > 0) && key != nil {
				p.associateCommentsAndEmptyLines(key, keyComments, emptyLines)
			}

			// Expect ':' for value
			if p.current != nil && p.current.Type == lexer.TokenMappingValue {
				if err := p.advance(); err != nil {
					p.addError(err.Error())
					continue
				}

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
					p.inMappingValue = true // Set flag before parsing value
					value = p.parseNode(p.current.Column)
					p.inMappingValue = false // Reset flag after parsing value
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

			// Associate captured comments and empty lines with the key
			if (len(keyComments) > 0 || len(emptyLines) > 0) && key != nil {
				p.associateCommentsAndEmptyLines(key, keyComments, emptyLines)
			}

			if p.current != nil && p.current.Type == lexer.TokenMappingValue {
				if err := p.advance(); err != nil {
					p.addError(err.Error())
					continue
				}

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
					p.inMappingValue = true // Set flag before parsing value
					value = p.parseNode(p.current.Column)
					p.inMappingValue = false // Reset flag after parsing value
				} else {
					value = p.nodeBuilder.BuildScalar("", node.StylePlain)
				}

				// If we have an inline comment, associate it with the value
				if inlineComment != nil {
					if mapping, ok := value.(*node.MappingNode); ok {
						// Store the inline comment in the mapping's LineComment
						mapping.LineComment = &node.CommentGroup{
							Comments:         []string{inlineComment.Value},
							BlankLinesWithin: []int{0},
						}
					} else if seq, ok := value.(*node.SequenceNode); ok {
						// Store the inline comment in the sequence's LineComment
						seq.LineComment = &node.CommentGroup{
							Comments:         []string{inlineComment.Value},
							BlankLinesWithin: []int{0},
						}
					} else if scalar, ok := value.(*node.ScalarNode); ok {
						// Store the inline comment in the scalar's LineComment
						scalar.LineComment = &node.CommentGroup{
							Comments:         []string{inlineComment.Value},
							BlankLinesWithin: []int{0},
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
	// Only associate comments if we didn't exit due to indent mismatch
	// If we exited due to indent, leave comments for parent level
	if !exitedDueToIndent {
		p.associateComments(mapping)
	}
	return mapping
}

// parseFlowSequence parses a flow-style sequence [a, b, c]
func (p *Parser) parseFlowSequence() node.Node {
	if err := p.advance(); err != nil {
		p.addError(err.Error())
		return nil
	}
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
			if err := p.advance(); err != nil {
				p.addError(err.Error())
				break
			}
		}

		// Check for EOF after advance to prevent infinite loop
		if p.current != nil && p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow sequences
			break
		}
	}

	if p.current != nil && p.current.Type == lexer.TokenFlowSequenceEnd {
		if err := p.advance(); err != nil {
			p.addError(err.Error())
		}
	}

	p.inFlow--
	seq := p.nodeBuilder.BuildSequence(items, node.StyleFlow)
	p.associateComments(seq)
	return seq
}

// parseFlowMapping parses a flow-style mapping {a: 1, b: 2}
func (p *Parser) parseFlowMapping() node.Node {
	if err := p.advance(); err != nil {
		p.addError(err.Error())
		return nil
	}
	p.inFlow++

	pairs := make([]*node.MappingPair, 0)

	for p.current != nil && p.current.Type != lexer.TokenFlowMappingEnd {
		if p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow mappings
			break
		}

		// Skip whitespace/empty
		if p.current.Type == lexer.TokenFlowEntry {
			if err := p.advance(); err != nil {
				p.addError(err.Error())
				break
			}
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
			if err := p.advance(); err != nil {
				p.addError(err.Error())
				break
			}
			if p.current != nil && p.current.Type == lexer.TokenEOF {
				// Be lenient with unclosed flow mappings
				break
			}
			continue
		}

		// Expect ':'
		if p.current != nil && p.current.Type == lexer.TokenMappingValue {
			if err := p.advance(); err != nil {
				p.addError(err.Error())
				continue
			}

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
			if err := p.advance(); err != nil {
				p.addError(err.Error())
				break
			}
		}

		// Check for EOF after advance to prevent infinite loop
		if p.current != nil && p.current.Type == lexer.TokenEOF {
			// Be lenient with unclosed flow mappings
			break
		}
	}

	if p.current != nil && p.current.Type == lexer.TokenFlowMappingEnd {
		if err := p.advance(); err != nil {
			p.addError(err.Error())
		}
	}

	p.inFlow--
	mapping := p.nodeBuilder.BuildMapping(pairs, node.StyleFlow)
	p.associateComments(mapping)
	return mapping
}

// Helper methods

func (p *Parser) isBlockMappingStart() bool {
	// Check if current token is a scalar followed by ':' on the same line
	if p.current == nil || p.peek == nil {
		return false
	}

	isScalar := p.current.Type == lexer.TokenPlainScalar ||
		p.current.Type == lexer.TokenSingleQuotedScalar ||
		p.current.Type == lexer.TokenDoubleQuotedScalar

	// Only consider it a mapping start if the colon is on the same line
	return isScalar && p.peek.Type == lexer.TokenMappingValue && p.current.Line == p.peek.Line
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

// associateCommentsAndEmptyLines associates both comments and empty lines with a node
func (p *Parser) associateCommentsAndEmptyLines(n node.Node, comments []*lexer.Token, emptyLines []*lexer.Token) {
	if n == nil {
		return
	}

	baseNode, ok := n.(interface{ GetBase() *node.BaseNode })
	if !ok {
		return
	}
	base := baseNode.GetBase()

	// Process comments
	if len(comments) > 0 {
		commentStrings := make([]string, len(comments))
		blankLinesWithin := make([]int, len(comments))
		emptyLineMarkers := make([]int, len(comments))

		for i, comment := range comments {
			commentStrings[i] = comment.Value
			blankLinesWithin[i] = comment.BlankLinesBefore
		}

		// Process empty lines and convert them to empty line markers
		for i, _ := range emptyLines {
			if i < len(emptyLineMarkers) {
				emptyLineMarkers[i] = 1 // One empty line marker per ##EMPTY_LINE## token
			}
		}

		base.HeadComment = &node.CommentGroup{
			Comments:         commentStrings,
			BlankLinesWithin: blankLinesWithin,
			EmptyLineMarkers: emptyLineMarkers,
			Format: node.CommentFormat{
				PreserveSpacing: true,
				GroupRelated:    true,
			},
		}
	}
}

func (p *Parser) associateComments(n node.Node) {
	// Associate pending comments with the node
	if len(p.commentQueue) > 0 && n != nil {
		var commentsToAssociate []*lexer.Token
		var commentsToKeep []*lexer.Token

		// Filter comments based on their position relative to the current node
		for _, comment := range p.commentQueue {
			if comment.IsInline {
				// Inline comments on the same line as the current token
				if p.current != nil && comment.Line == p.current.Line {
					commentsToAssociate = append(commentsToAssociate, comment)
				} else {
					commentsToKeep = append(commentsToKeep, comment)
				}
			} else {
				// For block comments, be more selective about association
				shouldAssociate := false

				if p.current != nil {
					// Only associate if comment is directly before this node (within 2 lines)
					lineDiff := p.current.Line - comment.Line
					if lineDiff > 0 && lineDiff <= 2 {
						// Don't associate @schema comments unless they're immediately before
						if strings.Contains(comment.Value, "@schema") {
							if lineDiff == 1 {
								shouldAssociate = true
							}
						} else {
							shouldAssociate = true
						}
					}
				}

				if shouldAssociate {
					commentsToAssociate = append(commentsToAssociate, comment)
				} else {
					commentsToKeep = append(commentsToKeep, comment)
				}
			}
		}

		// Associate the filtered comments - but limit to avoid duplicates
		associatedCount := 0
		schemaCommentCount := 0
		for _, comment := range commentsToAssociate {
			// Allow @schema comments but limit excessive duplication (more than 3 in a row)
			if strings.Contains(comment.Value, "@schema") {
				schemaCommentCount++
				if schemaCommentCount > 10 { // Allow up to 10 @schema comments per node (generous limit)
					continue
				}
			}

			// Create a comment processor to associate comments
			cp := node.NewCommentProcessor()
			if comment.IsInline {
				cp.AssociateCommentToNode(n, comment.Value, node.CommentPositionInline, 0)
			} else {
				cp.AssociateCommentToNode(n, comment.Value, node.CommentPositionAbove, comment.BlankLinesBefore)
			}
			associatedCount++
		}

		// Keep the remaining comments for the next node
		p.commentQueue = commentsToKeep
	}
}

// processFootComments processes comments that appear after a node (foot comments)
func (p *Parser) processFootComments(n node.Node) {
	if len(p.commentQueue) == 0 || n == nil {
		return
	}

	baseNode, ok := n.(interface{ GetBase() *node.BaseNode })
	if !ok {
		return
	}
	base := baseNode.GetBase()

	// Look for comments that should be foot comments (after the current node)
	var footComments []string
	var remainingComments []*lexer.Token

	for _, comment := range p.commentQueue {
		// Comments that are not inline and appear after current parsing position
		// are likely foot comments
		if !comment.IsInline && p.current != nil && comment.Line < p.current.Line {
			footComments = append(footComments, comment.Value)
		} else {
			remainingComments = append(remainingComments, comment)
		}
	}

	// Associate foot comments if we found any
	if len(footComments) > 0 {
		if base.FootComment == nil {
			base.FootComment = &node.CommentGroup{
				Comments: footComments,
				Format: node.CommentFormat{
					PreserveSpacing: true,
					GroupRelated:    true,
				},
			}
		} else {
			// Append to existing foot comments
			base.FootComment.Comments = append(base.FootComment.Comments, footComments...)
		}
	}

	// Update the comment queue
	p.commentQueue = remainingComments
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
