package parser

import (
	"strings"

	lexer2 "github.com/elioetibr/golang-yaml/v0/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v0/pkg/node"
)

// Document represents a YAML document with optional directives
type Document struct {
	// Directives at the beginning of the document
	Directives []Directive

	// The root node of the document
	Root node.Node

	// Whether the document has explicit start/end markers
	ExplicitStart bool
	ExplicitEnd   bool
}

// Directive represents a YAML directive
type Directive struct {
	Name       string
	Parameters []string
}

// Stream represents a stream of YAML documents
type Stream struct {
	Documents []*Document
}

// ParseStream parses a YAML stream containing multiple documents
func ParseStream(input string) (*Stream, error) {
	l := lexer2.NewLexerFromString(input)
	if err := l.Initialize(); err != nil {
		return nil, err
	}

	parser := NewParser(l)
	return parser.ParseStream()
}

// ParseStream parses multiple YAML documents from the input
func (p *Parser) ParseStream() (*Stream, error) {
	stream := &Stream{
		Documents: make([]*Document, 0),
	}

	// Initialize by reading the first two tokens
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	for p.current != nil && p.current.Type != lexer2.TokenEOF {
		doc, err := p.parseDocument()
		if err != nil {
			return nil, err
		}
		if doc != nil {
			stream.Documents = append(stream.Documents, doc)
		}

		// Check for more documents
		if p.current == nil || p.current.Type == lexer2.TokenEOF {
			break
		}
	}

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	return stream, nil
}

// parseDocument parses a single YAML document
func (p *Parser) parseDocument() (*Document, error) {
	doc := &Document{
		Directives: make([]Directive, 0),
	}

	// Clear anchor registry for new document
	p.anchorRegistry.Clear()

	// Parse directives
	for p.current != nil && p.current.Type == lexer2.TokenDirective {
		directive := p.parseDirective()
		if directive != nil {
			doc.Directives = append(doc.Directives, *directive)
		}
	}

	// Check for document start marker
	if p.current != nil && p.current.Type == lexer2.TokenDocumentStart {
		doc.ExplicitStart = true
		p.advance()
	}

	// Parse the document content
	if p.current != nil && p.current.Type != lexer2.TokenDocumentEnd && p.current.Type != lexer2.TokenEOF {
		doc.Root = p.parseNode(0)
	}

	// Check for document end marker
	if p.current != nil && p.current.Type == lexer2.TokenDocumentEnd {
		doc.ExplicitEnd = true
		p.advance()
	}

	// Apply merge keys if present
	if doc.Root != nil {
		if err := ResolveMergeKeys(doc.Root, p.anchorRegistry); err != nil {
			return nil, err
		}
	}

	return doc, nil
}

// parseDirective parses a YAML directive
func (p *Parser) parseDirective() *Directive {
	if p.current == nil || p.current.Type != lexer2.TokenDirective {
		return nil
	}

	// Parse directive value (e.g., "%YAML 1.2" or "%TAG ! tag:example.com,2014:")
	value := p.current.Value
	p.advance()

	// Remove the % prefix
	if strings.HasPrefix(value, "%") {
		value = value[1:]
	}

	// Split into name and parameters
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return nil
	}

	directive := &Directive{
		Name:       parts[0],
		Parameters: parts[1:],
	}

	// Handle specific directives
	switch directive.Name {
	case "YAML":
		// YAML version directive
		if len(directive.Parameters) > 0 {
			// Store version for validation
		}
	case "TAG":
		// TAG directive for custom tags
		if len(directive.Parameters) >= 2 {
			handle := directive.Parameters[0]
			prefix := directive.Parameters[1]
			// Register with tag resolver if available
			if p.tagResolver != nil {
				p.tagResolver.AddTagDirective(handle, prefix)
			}
		}
	}

	return directive
}

// GetDocumentCount returns the number of documents in a YAML string
func GetDocumentCount(input string) (int, error) {
	stream, err := ParseStream(input)
	if err != nil {
		return 0, err
	}
	return len(stream.Documents), nil
}

// ParseFirstDocument parses only the first document in a stream
func ParseFirstDocument(input string) (node.Node, error) {
	stream, err := ParseStream(input)
	if err != nil {
		return nil, err
	}

	if len(stream.Documents) == 0 {
		return nil, nil
	}

	return stream.Documents[0].Root, nil
}

// ParseAllDocuments parses all documents and returns their root nodes
func ParseAllDocuments(input string) ([]node.Node, error) {
	stream, err := ParseStream(input)
	if err != nil {
		return nil, err
	}

	nodes := make([]node.Node, 0, len(stream.Documents))
	for _, doc := range stream.Documents {
		if doc.Root != nil {
			nodes = append(nodes, doc.Root)
		}
	}

	return nodes, nil
}
