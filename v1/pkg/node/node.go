package node

// NodeType represents the type of YAML node
type NodeType int

const (
	NodeTypeScalar NodeType = iota
	NodeTypeSequence
	NodeTypeMapping
	NodeTypeSection
	NodeTypeDocument
)

func (t NodeType) String() string {
	switch t {
	case NodeTypeDocument:
		return "Document"
	case NodeTypeSection:
		return "Section"
	case NodeTypeScalar:
		return "Scalar"
	case NodeTypeSequence:
		return "Sequence"
	case NodeTypeMapping:
		return "Mapping"
	default:
		return "Unknown"
	}
}

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

// SectionType represents different types of sections in YAML
type SectionType int

const (
	SectionTypeGeneric SectionType = iota
	SectionTypeHeader      // Section with header comments (e.g., "# Company Name")
	SectionTypeConfiguration // Configuration blocks
	SectionTypeData         // Data blocks
	SectionTypeFooter       // Footer sections
	SectionTypeAny          // Matches any section type (for rules)
)

func (t SectionType) String() string {
	switch t {
	case SectionTypeGeneric:
		return "Generic"
	case SectionTypeHeader:
		return "Header"
	case SectionTypeConfiguration:
		return "Configuration"
	case SectionTypeData:
		return "Data"
	case SectionTypeFooter:
		return "Footer"
	case SectionTypeAny:
		return "Any"
	default:
		return "Unknown"
	}
}

// Node represents a node in the YAML AST
type Node interface {
	Type() NodeType
	Tag() string
	SetTag(tag string)
	Anchor() string
	SetAnchor(anchor string)
	Line() int
	Column() int
	Section() *Section
	SetSection(section *Section)
	Accept(visitor Visitor) error
}

// CommentGroup represents a collection of comments with formatting rules
type CommentGroup struct {
	Comments         []string
	BlankLinesBefore int    // Number of blank lines before this comment group
	BlankLinesAfter  int    // Number of blank lines after this comment group
	Format           CommentFormat // How to format these comments
}

// CommentFormat defines how comments should be formatted
type CommentFormat struct {
	IndentLevel     int  // Indentation level for comments
	PreserveSpacing bool // Whether to preserve original spacing
	GroupRelated    bool // Whether to group related comments
}

// Section represents a logical section of YAML with its own comment context
type Section struct {
	ID          string      // Unique identifier for the section
	Type        SectionType // Type of section
	Title       string      // Optional title for the section
	Description string      // Optional description
	Nodes       []Node      // Nodes belonging to this section
	Comments    *CommentGroup // Section-level comments

	// Section formatting options
	Formatting *SectionFormat
}

// SectionFormat defines formatting rules for sections
type SectionFormat struct {
	BlankLinesBefore int  // Blank lines before section
	BlankLinesAfter  int  // Blank lines after section
	IndentChildren   bool // Whether to indent child nodes
	PreserveStructure bool // Whether to preserve original structure
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

	// Section association
	ParentSection *Section

	// Formatting hints
	StyleHint        Style
	BlankLinesBefore int // Number of blank lines before this node
	BlankLinesAfter  int // Number of blank lines after this node
}

func (n *BaseNode) Tag() string              { return n.TagValue }
func (n *BaseNode) SetTag(tag string)        { n.TagValue = tag }
func (n *BaseNode) Anchor() string           { return n.AnchorValue }
func (n *BaseNode) SetAnchor(a string)       { n.AnchorValue = a }
func (n *BaseNode) Line() int                { return n.LineNumber }
func (n *BaseNode) Column() int              { return n.ColumnNumber }
func (n *BaseNode) Section() *Section        { return n.ParentSection }
func (n *BaseNode) SetSection(s *Section)    { n.ParentSection = s }

// ScalarNode represents a scalar value
type ScalarNode struct {
	BaseNode
	Value string
	Style Style
	Alias string // For alias references (e.g., "*anchor")
}

func (n *ScalarNode) Type() NodeType         { return NodeTypeScalar }
func (n *ScalarNode) Accept(v Visitor) error { return v.VisitScalar(n) }

// SequenceNode represents a sequence (list/array)
type SequenceNode struct {
	BaseNode
	Items []Node
	Style Style
}

func (n *SequenceNode) Type() NodeType         { return NodeTypeSequence }
func (n *SequenceNode) Accept(v Visitor) error { return v.VisitSequence(n) }

// MappingNode represents a mapping (dict/map)
type MappingNode struct {
	BaseNode
	Pairs []*MappingPair
	Style Style
}

func (n *MappingNode) Type() NodeType         { return NodeTypeMapping }
func (n *MappingNode) Accept(v Visitor) error { return v.VisitMapping(n) }

// SectionNode represents a logical section containing related nodes
type SectionNode struct {
	BaseNode
	SectionData *Section
}

func (n *SectionNode) Type() NodeType         { return NodeTypeSection }
func (n *SectionNode) Accept(v Visitor) error { return v.VisitSection(n) }
func (n *SectionNode) GetSection() *Section   { return n.SectionData }

// DocumentNode represents the root document containing sections
type DocumentNode struct {
	BaseNode
	Sections []*Section
	Nodes    []Node // Direct child nodes not in sections
}

func (n *DocumentNode) Type() NodeType         { return NodeTypeDocument }
func (n *DocumentNode) Accept(v Visitor) error { return v.VisitDocument(n) }

// MappingPair represents a key-value pair in a mapping
type MappingPair struct {
	Key   Node
	Value Node

	// Comment tracking for key and value
	KeyComment   *CommentGroup
	ValueComment *CommentGroup

	// Section association
	ParentSection *Section

	// Formatting
	BlankLinesBefore int
	BlankLinesAfter  int
}

// Visitor interface for visiting nodes (Visitor pattern)
type Visitor interface {
	VisitScalar(*ScalarNode) error
	VisitSequence(*SequenceNode) error
	VisitMapping(*MappingNode) error
	VisitSection(*SectionNode) error
	VisitDocument(*DocumentNode) error
}

// Builder interface for constructing nodes (Builder pattern)
type Builder interface {
	BuildScalar(value string, style Style) *ScalarNode
	BuildSequence(items []Node, style Style) *SequenceNode
	BuildMapping(pairs []*MappingPair, style Style) *MappingNode
	BuildSection(section *Section) *SectionNode
	BuildDocument(sections []*Section, nodes []Node) *DocumentNode
	WithTag(node Node, tag string) Node
	WithAnchor(node Node, anchor string) Node
	WithSection(node Node, section *Section) Node
}

// SectionBuilder creates and manages sections
type SectionBuilder interface {
	CreateSection(id string, sectionType SectionType) *Section
	WithTitle(section *Section, title string) *Section
	WithDescription(section *Section, description string) *Section
	WithFormatting(section *Section, format *SectionFormat) *Section
	AddNode(section *Section, node Node) *Section
}

// DefaultBuilder implements the Builder interface
type DefaultBuilder struct{}

// NewBuilder creates a new DefaultBuilder
func NewBuilder() Builder {
	return &DefaultBuilder{}
}

func (b *DefaultBuilder) Scalar(value string, style Style) Node {
	return &ScalarNode{
		Value: value,
		Style: style,
	}
}

func (b *DefaultBuilder) Sequence(items []Node, style Style) Node {
	return &SequenceNode{
		Items: items,
		Style: style,
	}
}

func (b *DefaultBuilder) Mapping(pairs []*MappingPair, style Style) Node {
	return &MappingNode{
		Pairs: pairs,
		Style: style,
	}
}

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

func (b *DefaultBuilder) BuildSection(section *Section) *SectionNode {
	return &SectionNode{
		SectionData: section,
	}
}

func (b *DefaultBuilder) BuildDocument(sections []*Section, nodes []Node) *DocumentNode {
	return &DocumentNode{
		Sections: sections,
		Nodes:    nodes,
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

func (b *DefaultBuilder) WithSection(node Node, section *Section) Node {
	node.SetSection(section)
	return node
}

// DefaultSectionBuilder implements the SectionBuilder interface
type DefaultSectionBuilder struct{}

func (b *DefaultSectionBuilder) CreateSection(id string, sectionType SectionType) *Section {
	return &Section{
		ID:   id,
		Type: sectionType,
		Formatting: &SectionFormat{
			BlankLinesBefore:  1,
			BlankLinesAfter:   0,
			IndentChildren:    false,
			PreserveStructure: true,
		},
	}
}

func (b *DefaultSectionBuilder) WithTitle(section *Section, title string) *Section {
	if section != nil {
		section.Title = title
	}
	return section
}

func (b *DefaultSectionBuilder) WithDescription(section *Section, description string) *Section {
	if section != nil {
		section.Description = description
	}
	return section
}

func (b *DefaultSectionBuilder) WithFormatting(section *Section, format *SectionFormat) *Section {
	if section != nil {
		section.Formatting = format
	}
	return section
}

func (b *DefaultSectionBuilder) AddNode(section *Section, node Node) *Section {
	if section != nil {
		section.Nodes = append(section.Nodes, node)
		node.SetSection(section)
	}
	return section
}

// Helper functions for working with sections

// FindSectionByID finds a section by its ID in a document
func FindSectionByID(doc *DocumentNode, id string) *Section {
	for _, section := range doc.Sections {
		if section.ID == id {
			return section
		}
	}
	return nil
}

// GetNodeSection returns the section that contains the given node
func GetNodeSection(node Node) *Section {
	return node.Section()
}

// GroupNodesBySection groups nodes by their section
func GroupNodesBySection(nodes []Node) map[*Section][]Node {
	groups := make(map[*Section][]Node)
	for _, node := range nodes {
		section := node.Section()
		groups[section] = append(groups[section], node)
	}
	return groups
}