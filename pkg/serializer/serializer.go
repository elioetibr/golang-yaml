package serializer

import (
	"fmt"
	"io"
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/node"
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

	// Handle inline comments
	if s.options.PreserveComments && emitComments {
		s.emitComments(n, node.CommentPositionInline, indent)
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
		s.write("[]")
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
		s.write("{}")
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

		// Emit key comments if needed (before writing key)
		if s.options.PreserveComments && pair.Key != nil {
			// fmt.Printf("[DEBUG] Emitting comments for key at indent %d, column %d\n", indent, s.column)
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
		if s.isComplexNode(pair.Value) {
			s.write(":")
			s.writeLine("")
			// Value comments are handled by serializeNode
			err = s.serializeNodeWithComments(pair.Value, indent+s.options.Indent, true)
		} else {
			s.write(": ")
			// Value comments are handled by serializeNode
			err = s.serializeNodeWithComments(pair.Value, indent, true)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// serializeFlowMapping serializes a flow-style mapping
func (s *Serializer) serializeFlowMapping(m *node.MappingNode, indent int) error {
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
	// When the encoder creates these from actual bool/nil values,
	// they come as plain scalars and should not be quoted
	lowerValue := strings.ToLower(value)
	if lowerValue == "true" || lowerValue == "false" || lowerValue == "null" {
		return false
	}

	// Quote other YAML special values that might be ambiguous
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

	// Get base node to access comments
	baseNode, ok := n.(interface{ GetBase() *node.BaseNode })
	if !ok {
		return
	}
	base := baseNode.GetBase()

	var commentGroup *node.CommentGroup
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

	for _, comment := range commentGroup.Comments {
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
			if position == node.CommentPositionAbove && s.options.BlankLinesBeforeComment > 0 {
				for i := 0; i < s.options.BlankLinesBeforeComment; i++ {
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
