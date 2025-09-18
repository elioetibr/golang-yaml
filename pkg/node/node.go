package node

// NodeType represents the type of YAML node
type NodeType int

const (
	NodeTypeScalar NodeType = iota
	NodeTypeSequence
	NodeTypeMapping
)

// Style represents the style of a node (block vs flow)
type Style int

const (
	StyleAny Style = iota
	StyleBlock
	StyleFlow
	StylePlain
	StyleSingleQuoted
	StyleDoubleQuoted
	StyleLiteral
	StyleFolded
)

// Node represents a node in the YAML AST
type Node interface {
	Type() NodeType
	Tag() string
	SetTag(tag string)
	Anchor() string
	SetAnchor(anchor string)
	Line() int
	Column() int
	Accept(visitor Visitor) error
}

// CommentGroup represents a collection of comments
type CommentGroup struct {
	Comments         []string
	BlankLinesBefore int // Number of blank lines before this comment group
}

// BaseNode contains common fields for all node types
type BaseNode struct {
	TagValue     string
	AnchorValue  string
	LineNumber   int
	ColumnNumber int

	// Comment associations
	HeadComment *CommentGroup // Comments before the node
	LineComment *CommentGroup // Inline comment on same line
	FootComment *CommentGroup // Comments after the node

	// Blank line tracking
	BlankLinesBefore int // Number of blank lines before this node
	BlankLinesAfter  int // Number of blank lines after this node

	StyleHint Style
}

func (n *BaseNode) Tag() string        { return n.TagValue }
func (n *BaseNode) SetTag(tag string)  { n.TagValue = tag }
func (n *BaseNode) Anchor() string     { return n.AnchorValue }
func (n *BaseNode) SetAnchor(a string) { n.AnchorValue = a }
func (n *BaseNode) Line() int          { return n.LineNumber }
func (n *BaseNode) Column() int        { return n.ColumnNumber }

// ScalarNode represents a scalar value
type ScalarNode struct {
	BaseNode
	Value string
	Style Style
	Alias string // For alias references (e.g., "*anchor")
}

func (n *ScalarNode) Type() NodeType         { return NodeTypeScalar }
func (n *ScalarNode) Accept(v Visitor) error { return v.VisitScalar(n) }
func (n *ScalarNode) GetBase() *BaseNode     { return &n.BaseNode }

// SequenceNode represents a sequence (list/array)
type SequenceNode struct {
	BaseNode
	Items []Node
	Style Style
}

func (n *SequenceNode) Type() NodeType         { return NodeTypeSequence }
func (n *SequenceNode) Accept(v Visitor) error { return v.VisitSequence(n) }
func (n *SequenceNode) GetBase() *BaseNode     { return &n.BaseNode }

// MappingNode represents a mapping (dict/map)
type MappingNode struct {
	BaseNode
	Pairs []*MappingPair
	Style Style
}

func (n *MappingNode) Type() NodeType         { return NodeTypeMapping }
func (n *MappingNode) Accept(v Visitor) error { return v.VisitMapping(n) }
func (n *MappingNode) GetBase() *BaseNode     { return &n.BaseNode }

// MappingPair represents a key-value pair in a mapping
type MappingPair struct {
	Key   Node
	Value Node

	// Separate comment tracking for key and value
	KeyComment   *CommentGroup // Comments specifically for the key
	ValueComment *CommentGroup // Comments specifically for the value

	// Blank lines between key-value pairs
	BlankLinesBefore int
	BlankLinesAfter  int
}

// Visitor interface for visiting nodes (Visitor pattern)
type Visitor interface {
	VisitScalar(*ScalarNode) error
	VisitSequence(*SequenceNode) error
	VisitMapping(*MappingNode) error
}

// Builder interface for constructing nodes (Builder pattern)
type Builder interface {
	BuildScalar(value string, style Style) *ScalarNode
	BuildSequence(items []Node, style Style) *SequenceNode
	BuildMapping(pairs []*MappingPair, style Style) *MappingNode
	WithTag(node Node, tag string) Node
	WithAnchor(node Node, anchor string) Node
}

// DefaultBuilder implements the Builder interface
type DefaultBuilder struct{}

func (b *DefaultBuilder) BuildScalar(value string, style Style) *ScalarNode {
	return &ScalarNode{
		Value: value,
		Style: style,
	}
}

func (b *DefaultBuilder) BuildSequence(items []Node, style Style) *SequenceNode {
	return &SequenceNode{
		Items: items,
		Style: style,
	}
}

func (b *DefaultBuilder) BuildMapping(pairs []*MappingPair, style Style) *MappingNode {
	return &MappingNode{
		Pairs: pairs,
		Style: style,
	}
}

func (b *DefaultBuilder) WithTag(node Node, tag string) Node {
	node.SetTag(tag)
	return node
}

func (b *DefaultBuilder) WithAnchor(node Node, anchor string) Node {
	node.SetAnchor(anchor)
	return node
}
