package serializer

import (
	"fmt"
	"io"
	"strings"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// Options configures the serialization behavior
type Options struct {
	// Indentation settings
	Indent      int  // Number of spaces for indentation (default: 2)
	UseTabsOnly bool // Use tabs instead of spaces

	// Style preferences
	PreferBlockStyle bool // Prefer block style over flow style
	PreferFlowStyle  bool // Prefer flow style over block style
	LineWidth        int  // Maximum line width for flow style (default: 80)

	// Comment handling
	PreserveComments bool // Preserve comments from AST
	CommentColumn    int  // Column to align inline comments (default: 0 = no alignment)

	// Blank lines
	PreserveBlankLines      bool // Preserve blank lines from AST
	BlankLinesBeforeComment int  // Number of blank lines before comments (default: 0)

	// Document markers
	ExplicitDocumentStart bool // Always emit --- at document start
	ExplicitDocumentEnd   bool // Always emit ... at document end

	// Tag handling
	EmitTags bool // Emit tags from nodes
}

// DefaultOptions returns the default serialization options
func DefaultOptions() *Options {
	return &Options{
		Indent:             2,
		PreserveComments:   true,
		PreserveBlankLines: true,
		LineWidth:          80,
	}
}

// Serializer converts YAML AST to text
type Serializer struct {
	writer      io.Writer
	options     *Options
	indentLevel int
	column      int
	line        int
	inFlow      bool
	buffer      strings.Builder
}

// NewSerializer creates a new serializer with the given writer and options
func NewSerializer(w io.Writer, opts *Options) *Serializer {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &Serializer{
		writer:  w,
		options: opts,
		line:    1,
		column:  1,
	}
}

// Serialize serializes a node to the writer
func (s *Serializer) Serialize(n node.Node) error {
	// Don't emit document head comments here anymore - they're handled by first key
	// This prevents duplication

	if s.options.ExplicitDocumentStart {
		s.writeLine("---")
	}

	if err := s.serializeNode(n, 0); err != nil {
		return err
	}

	if s.options.ExplicitDocumentEnd {
		s.writeLine("")
		s.writeLine("...")
	}

	// Flush buffer to writer
	_, err := s.writer.Write([]byte(s.buffer.String()))
	return err
}

// serializeNode serializes any node based on its type
func (s *Serializer) serializeNode(n node.Node, indent int) error {
	return s.serializeNodeWithComments(n, indent, true)
}

// serializeNodeWithComments serializes a node with optional comment emission
func (s *Serializer) serializeNodeWithComments(n node.Node, indent int, emitComments bool) error {
	if n == nil {
		return nil
	}

	// Handle comments before the node (if requested)
	if s.options.PreserveComments && emitComments {
		s.emitComments(n, node.CommentPositionAbove, indent)
	}

	switch v := n.(type) {
	case *node.ScalarNode:
		err := s.serializeScalar(v, indent)
		if err != nil {
			return err
		}

	case *node.SequenceNode:
		err := s.serializeSequence(v, indent)
		if err != nil {
			return err
		}

	case *node.MappingNode:
		err := s.serializeMapping(v, indent)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown node type: %T", n)
	}

	// Handle inline comments (skip for mappings/sequences as they're handled specially)
	if s.options.PreserveComments && emitComments {
		if _, isMapping := n.(*node.MappingNode); !isMapping {
			if _, isSequence := n.(*node.SequenceNode); !isSequence {
				s.emitComments(n, node.CommentPositionInline, indent)
			}
		}
	}

	// Handle comments after the node
	if s.options.PreserveComments && !s.inFlow && emitComments {
		s.emitComments(n, node.CommentPositionBelow, indent)
	}

	return nil
}

// serializeScalar serializes a scalar node
func (s *Serializer) serializeScalar(n *node.ScalarNode, indent int) error {
	value := n.Value

	switch n.Style {
	case node.StyleDoubleQuoted:
		value = s.doubleQuoteScalar(value)
	case node.StyleSingleQuoted:
		value = s.singleQuoteScalar(value)
	case node.StyleLiteral:
		return s.serializeLiteralScalar(value, indent)
	case node.StyleFolded:
		return s.serializeFoldedScalar(value, indent)
	default:
		// Plain scalar - check if quoting needed
		if s.needsQuoting(value) {
			value = s.doubleQuoteScalar(value)
		}
	}

	// If we're at the beginning of a line (after a newline), add indentation
	// Debug: column and indent info
	if s.column == 0 && indent > 0 {
		s.writeIndent(indent)
	}
	s.write(value)
	return nil
}

// serializeSequence serializes a sequence node
func (s *Serializer) serializeSequence(seq *node.SequenceNode, indent int) error {
	if len(seq.Items) == 0 {
		// For empty sequences, check if we're in a mapping value position
		// If we are and on the same line, just write []
		// Otherwise write [] on next line with proper indent
		if s.column > 0 {
			// We're on the same line as a key, write inline
			s.write(" []")
		} else {
			// We're on a new line, write with indent
			s.writeIndent(indent)
			s.write("[]")
		}
		return nil
	}

	// Determine style
	useFlow := seq.Style == node.StyleFlow ||
		(s.options.PreferFlowStyle && !s.options.PreferBlockStyle)

	if useFlow {
		return s.serializeFlowSequence(seq, indent)
	}
	return s.serializeBlockSequence(seq, indent)
}

// serializeBlockSequence serializes a block-style sequence
func (s *Serializer) serializeBlockSequence(seq *node.SequenceNode, indent int) error {
	for i, item := range seq.Items {
		if i > 0 || s.column > 1 {
			s.writeLine("")
		}

		// Write indent and dash
		s.writeIndent(indent)

		// Check if item is complex (needs new line)
		if s.isComplexNode(item) {
			s.write("-")
			s.writeLine("")
			err := s.serializeNode(item, indent+s.options.Indent)
			if err != nil {
				return err
			}
		} else {
			s.write("- ")
			err := s.serializeNode(item, indent+s.options.Indent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// serializeFlowSequence serializes a flow-style sequence
func (s *Serializer) serializeFlowSequence(seq *node.SequenceNode, indent int) error {
	s.write("[")
	s.inFlow = true

	for i, item := range seq.Items {
		if i > 0 {
			s.write(", ")
		}
		err := s.serializeNode(item, indent)
		if err != nil {
			return err
		}
	}

	s.write("]")
	s.inFlow = false
	return nil
}

// serializeMapping serializes a mapping node
func (s *Serializer) serializeMapping(m *node.MappingNode, indent int) error {
	if len(m.Pairs) == 0 {
		// For empty mappings, check if we're in a mapping value position
		// If we are and on the same line, just write {}
		// Otherwise write {} on next line with proper indent
		if s.column > 0 {
			// We're on the same line as a key, write inline
			s.write(" {}")
		} else {
			// We're on a new line, write with indent
			s.writeIndent(indent)
			s.write("{}")
		}
		return nil
	}

	// Determine style
	useFlow := m.Style == node.StyleFlow ||
		(s.options.PreferFlowStyle && !s.options.PreferBlockStyle)

	if useFlow {
		return s.serializeFlowMapping(m, indent)
	}
	return s.serializeBlockMapping(m, indent)
}

// serializeBlockMapping serializes a block-style mapping
func (s *Serializer) serializeBlockMapping(m *node.MappingNode, indent int) error {
	for i, pair := range m.Pairs {
		if i > 0 || s.column > 1 {
			s.writeLine("")
		}

		// Handle blank lines before entry
		if s.options.PreserveBlankLines && pair.Key != nil {
			if baseNode, ok := pair.Key.(interface{ GetBase() *node.BaseNode }); ok {
				for j := 0; j < baseNode.GetBase().BlankLinesBefore; j++ {
					s.writeLine("")
				}
			}
		}

		// Add header comment metadata to the first matching key (as requested)
		// Only do this if document head comments haven't been emitted yet
		if i == 0 && m.HasDocumentHeadComments && m.HeadComment != nil && pair.Key != nil {
			if keyBaseNode, ok := pair.Key.(interface{ GetBase() *node.BaseNode }); ok {
				keyBase := keyBaseNode.GetBase()
				// Transfer document head comments to first key if it doesn't already have head comments
				if keyBase.HeadComment == nil {
					keyBase.HeadComment = m.HeadComment
				} else {
					// Merge the document head comments with existing key comments
					mergedComments := append(m.HeadComment.Comments, keyBase.HeadComment.Comments...)
					keyBase.HeadComment.Comments = mergedComments
				}
				// Clear the document head comments to prevent double emission
				m.HeadComment = nil
				m.HasDocumentHeadComments = false
			}
		}

		// Emit key comments if needed (before writing key)
		if s.options.PreserveComments && pair.Key != nil {
			s.emitComments(pair.Key, node.CommentPositionAbove, indent)
		}

		// Write key (with indent if not already at position)
		if s.column == 0 {
			s.writeIndent(indent)
		}

		// Don't emit comments for the key node itself - we already handled them
		err := s.serializeNodeWithComments(pair.Key, indent, false)
		if err != nil {
			return err
		}

		// Check if value is complex (needs new line)
		// Empty sequences and mappings should be inline
		isComplex := s.isComplexNode(pair.Value)
		isEmpty := false
		if seq, ok := pair.Value.(*node.SequenceNode); ok && len(seq.Items) == 0 {
			isEmpty = true
			isComplex = false
		} else if mapping, ok := pair.Value.(*node.MappingNode); ok && len(mapping.Pairs) == 0 {
			isEmpty = true
			isComplex = false
		}

		if isComplex {
			s.write(":")

			// Check if the value (mapping/sequence) has an inline comment
			hasInlineComment := false
			if s.options.PreserveComments {
				if mapping, ok := pair.Value.(*node.MappingNode); ok && mapping.LineComment != nil {
					// Write inline comment for mapping key
					s.write("  ")
					for _, comment := range mapping.LineComment.Comments {
						s.write(comment)
					}
					hasInlineComment = true
				} else if seq, ok := pair.Value.(*node.SequenceNode); ok && seq.LineComment != nil {
					// Write inline comment for sequence key
					s.write("  ")
					for _, comment := range seq.LineComment.Comments {
						s.write(comment)
					}
					hasInlineComment = true
				}
			}

			// Only add newline if there's no immediate head comment on the value
			// The serializeNodeWithComments will handle proper spacing based on comments
			valueHasHeadComment := false
			if s.options.PreserveComments {
				if baseNode, ok := pair.Value.(interface{ GetBase() *node.BaseNode }); ok {
					base := baseNode.GetBase()
					if base.HeadComment != nil && len(base.HeadComment.Comments) > 0 {
						valueHasHeadComment = true
					}
				}
			}

			if !valueHasHeadComment {
				s.writeLine("")
			}
			// Value comments are handled by serializeNode
			err = s.serializeNodeWithComments(pair.Value, indent+s.options.Indent, !hasInlineComment)
		} else {
			s.write(": ")
			// For empty collections, don't let them add extra spacing
			if isEmpty {
				// Write empty collection inline
				if _, ok := pair.Value.(*node.SequenceNode); ok {
					s.write("[]")
					err = nil
				} else if _, ok := pair.Value.(*node.MappingNode); ok {
					s.write("{}")
					err = nil
				} else {
					err = s.serializeNodeWithComments(pair.Value, indent, true)
				}
			} else {
				// Value comments are handled by serializeNode
				err = s.serializeNodeWithComments(pair.Value, indent, true)
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// serializeFlowMapping serializes a flow-style mapping
func (s *Serializer) serializeFlowMapping(m *node.MappingNode, indent int) error {
	// Check if we're at the start of a line and need indentation
	if s.column == 0 {
		s.writeIndent(indent)
	}

	s.write("{")
	s.inFlow = true

	for i, pair := range m.Pairs {
		if i > 0 {
			s.write(", ")
		}

		// Don't emit comments for the key node itself - we already handled them
		err := s.serializeNodeWithComments(pair.Key, indent, false)
		if err != nil {
			return err
		}

		s.write(": ")

		err = s.serializeNode(pair.Value, indent)
		if err != nil {
			return err
		}
	}

	s.write("}")
	s.inFlow = false
	return nil
}

// serializeLiteralScalar serializes a literal block scalar
func (s *Serializer) serializeLiteralScalar(value string, indent int) error {
	s.write("|")
	s.writeLine("")

	lines := strings.Split(value, "\n")
	for _, line := range lines {
		s.writeIndent(indent + s.options.Indent)
		s.writeLine(line)
	}
	// Add trailing newline for block scalars
	s.writeLine("")

	return nil
}

// serializeFoldedScalar serializes a folded block scalar
func (s *Serializer) serializeFoldedScalar(value string, indent int) error {
	s.write(">")
	s.writeLine("")

	lines := strings.Split(value, "\n")
	for _, line := range lines {
		s.writeIndent(indent + s.options.Indent)
		s.writeLine(line)
	}
	// Add trailing newline for block scalars
	s.writeLine("")

	return nil
}

// Helper methods

func (s *Serializer) write(str string) {
	s.buffer.WriteString(str)
	s.column += len(str)
}

func (s *Serializer) writeLine(str string) {
	if str != "" {
		s.write(str)
	}
	s.buffer.WriteString("\n")
	s.line++
	s.column = 0
}

func (s *Serializer) writeIndent(indent int) {
	if s.options.UseTabsOnly {
		tabs := indent / 8
		for i := 0; i < tabs; i++ {
			s.write("\t")
		}
	} else {
		for i := 0; i < indent; i++ {
			s.write(" ")
		}
	}
}

func (s *Serializer) needsQuoting(value string) bool {
	// Check if value needs quoting based on YAML rules
	if value == "" {
		return true
	}

	// Don't quote boolean and null literals
	// When the encoder creates these from actual bool/nil values-with-comments,
	// they come as plain scalars and should not be quoted
	lowerValue := strings.ToLower(value)
	if lowerValue == "true" || lowerValue == "false" || lowerValue == "null" {
		return false
	}

	// Quote other YAML special values-with-comments that might be ambiguous
	specialValues := []string{"yes", "no", "on", "off"}
	for _, special := range specialValues {
		if lowerValue == special {
			return true
		}
	}

	// Check for characters that require quoting
	if strings.ContainsAny(value, ":#@*&{}[]|>'\"\n") {
		return true
	}

	// Check if it looks like a number
	if _, err := fmt.Sscanf(value, "%f", new(float64)); err == nil {
		return false // Numbers don't need quoting
	}

	return false
}

func (s *Serializer) doubleQuoteScalar(value string) string {
	// Escape special characters
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return fmt.Sprintf("\"%s\"", value)
}

func (s *Serializer) singleQuoteScalar(value string) string {
	// Escape single quotes by doubling them
	value = strings.ReplaceAll(value, "'", "''")
	return fmt.Sprintf("'%s'", value)
}

func (s *Serializer) isComplexNode(n node.Node) bool {
	if n == nil {
		return false
	}

	switch v := n.(type) {
	case *node.SequenceNode, *node.MappingNode:
		return true
	case *node.ScalarNode:
		// Scalars with comments above them need to be on a new line
		if s.options.PreserveComments && v.HeadComment != nil && len(v.HeadComment.Comments) > 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func (s *Serializer) emitComments(n node.Node, position node.CommentPosition, indent int) {
	if n == nil {
		return
	}

	// Skip document head comments for MappingNode - they're handled separately in Serialize()
	if position == node.CommentPositionAbove {
		if mappingNode, ok := n.(*node.MappingNode); ok && mappingNode.HasDocumentHeadComments {
			// Document head comments are emitted at the document level, not here
			return
		}
	}

	// Get base node to access comments
	baseNode, ok := n.(interface{ GetBase() *node.BaseNode })
	if !ok {
		return
	}
	base := baseNode.GetBase()
	if base == nil {
		return
	}

	var commentGroup *node.CommentGroup
	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch position {
	case node.CommentPositionAbove:
		commentGroup = base.HeadComment
	case node.CommentPositionInline:
		commentGroup = base.LineComment
	case node.CommentPositionBelow:
		commentGroup = base.FootComment
	}

	if commentGroup == nil || len(commentGroup.Comments) == 0 {
		return
	}
	for i, comment := range commentGroup.Comments {
		if position == node.CommentPositionInline {
			// Inline comment - add spacing
			if s.options.CommentColumn > 0 && s.column < s.options.CommentColumn {
				for s.column < s.options.CommentColumn {
					s.write(" ")
				}
			} else {
				s.write("  ")
			}
			s.write(comment)
		} else {
			// Head or foot comment - on its own line

			// Handle blank lines within comment group
			if i > 0 && len(commentGroup.BlankLinesWithin) > i {
				blankLines := commentGroup.BlankLinesWithin[i]
				for j := 0; j < blankLines; j++ {
					s.writeLine("")
				}
			}

			// Handle ##EMPTY_LINE## markers for precise blank line preservation
			if len(commentGroup.EmptyLineMarkers) > i && commentGroup.EmptyLineMarkers[i] > 0 {
				for j := 0; j < commentGroup.EmptyLineMarkers[i]; j++ {
					s.writeLine("")
				}
			}

			if position == node.CommentPositionAbove && s.options.BlankLinesBeforeComment > 0 && i == 0 {
				for j := 0; j < s.options.BlankLinesBeforeComment; j++ {
					s.writeLine("")
				}
			}
			// Write comment with indentation
			s.writeIndent(indent)
			s.writeLine(comment)
		}
	}
}

// SerializeToString is a convenience method to serialize to a string
func SerializeToString(n node.Node, opts *Options) (string, error) {
	var buf strings.Builder
	serializer := NewSerializer(&buf, opts)
	err := serializer.Serialize(n)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
