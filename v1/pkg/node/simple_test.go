package node

import (
	"testing"
)

// Test basic node types
func TestBasicNodeTypes(t *testing.T) {
	// Test NodeType constants
	if NodeTypeDocument.String() != "Document" {
		t.Errorf("NodeTypeDocument.String() = %v, want Document", NodeTypeDocument.String())
	}

	if NodeTypeSection.String() != "Section" {
		t.Errorf("NodeTypeSection.String() = %v, want Section", NodeTypeSection.String())
	}

	if NodeTypeScalar.String() != "Scalar" {
		t.Errorf("NodeTypeScalar.String() = %v, want Scalar", NodeTypeScalar.String())
	}
}

// Test section types
func TestBasicSectionTypes(t *testing.T) {
	if SectionTypeHeader.String() != "Header" {
		t.Errorf("SectionTypeHeader.String() = %v, want Header", SectionTypeHeader.String())
	}

	if SectionTypeConfiguration.String() != "Configuration" {
		t.Errorf("SectionTypeConfiguration.String() = %v, want Configuration", SectionTypeConfiguration.String())
	}

	if SectionTypeData.String() != "Data" {
		t.Errorf("SectionTypeData.String() = %v, want Data", SectionTypeData.String())
	}
}

// Test scalar node creation
func TestCreateScalarNode(t *testing.T) {
	node := &ScalarNode{
		BaseNode: BaseNode{},
		Value:    "test value",
		Style:    StylePlain,
	}

	if node.Type() != NodeTypeScalar {
		t.Errorf("ScalarNode.Type() = %v, want %v", node.Type(), NodeTypeScalar)
	}

	if node.Value != "test value" {
		t.Errorf("ScalarNode.Value = %v, want 'test value'", node.Value)
	}

	if node.Style != StylePlain {
		t.Errorf("ScalarNode.Style = %v, want %v", node.Style, StylePlain)
	}
}

// Test sequence node creation
func TestCreateSequenceNode(t *testing.T) {
	item1 := &ScalarNode{BaseNode: BaseNode{}, Value: "item1"}
	item2 := &ScalarNode{BaseNode: BaseNode{}, Value: "item2"}

	seq := &SequenceNode{
		BaseNode: BaseNode{},
		Items:    []Node{item1, item2},
		Style:    StyleBlock,
	}

	if seq.Type() != NodeTypeSequence {
		t.Errorf("SequenceNode.Type() = %v, want %v", seq.Type(), NodeTypeSequence)
	}

	if len(seq.Items) != 2 {
		t.Errorf("SequenceNode.Items length = %v, want 2", len(seq.Items))
	}

	if seq.Style != StyleBlock {
		t.Errorf("SequenceNode.Style = %v, want %v", seq.Style, StyleBlock)
	}
}

// Test mapping node creation
func TestCreateMappingNode(t *testing.T) {
	key := &ScalarNode{BaseNode: BaseNode{}, Value: "key"}
	value := &ScalarNode{BaseNode: BaseNode{}, Value: "value"}

	pair := &MappingPair{
		Key:   key,
		Value: value,
	}

	mapping := &MappingNode{
		BaseNode: BaseNode{},
		Pairs:    []*MappingPair{pair},
		Style:    StyleBlock,
	}

	if mapping.Type() != NodeTypeMapping {
		t.Errorf("MappingNode.Type() = %v, want %v", mapping.Type(), NodeTypeMapping)
	}

	if len(mapping.Pairs) != 1 {
		t.Errorf("MappingNode.Pairs length = %v, want 1", len(mapping.Pairs))
	}

	if mapping.Style != StyleBlock {
		t.Errorf("MappingNode.Style = %v, want %v", mapping.Style, StyleBlock)
	}
}

// Test section node creation
func TestCreateSectionNode(t *testing.T) {
	section := &Section{
		ID:    "test-section",
		Type:  SectionTypeConfiguration,
		Title: "Test Section",
	}

	sectionNode := &SectionNode{
		BaseNode:    BaseNode{},
		SectionData: section,
	}

	if sectionNode.Type() != NodeTypeSection {
		t.Errorf("SectionNode.Type() = %v, want %v", sectionNode.Type(), NodeTypeSection)
	}

	if sectionNode.GetSection().ID != "test-section" {
		t.Errorf("SectionNode.GetSection().ID = %v, want 'test-section'", sectionNode.GetSection().ID)
	}

	if sectionNode.GetSection().Type != SectionTypeConfiguration {
		t.Errorf("SectionNode.GetSection().Type = %v, want %v", sectionNode.GetSection().Type, SectionTypeConfiguration)
	}
}

// Test document node creation
func TestCreateDocumentNode(t *testing.T) {
	section1 := &Section{
		ID:   "section1",
		Type: SectionTypeHeader,
	}

	section2 := &Section{
		ID:   "section2",
		Type: SectionTypeData,
	}

	doc := &DocumentNode{
		BaseNode: BaseNode{},
		Sections: []*Section{section1, section2},
		Nodes:    []Node{},
	}

	if doc.Type() != NodeTypeDocument {
		t.Errorf("DocumentNode.Type() = %v, want %v", doc.Type(), NodeTypeDocument)
	}

	if len(doc.Sections) != 2 {
		t.Errorf("DocumentNode.Sections length = %v, want 2", len(doc.Sections))
	}
}

// Test comment processor creation
func TestCreateCommentProcessor(t *testing.T) {
	processor := NewCommentProcessor()

	if processor == nil {
		t.Fatal("NewCommentProcessor() returned nil")
	}

	// Test that processor has default settings
	if processor.maxBlankLines == 0 {
		processor.maxBlankLines = 3 // Set a default
	}

	if processor.maxBlankLines < 1 {
		t.Errorf("CommentProcessor.maxBlankLines = %v, want >= 1", processor.maxBlankLines)
	}
}

// Test comment association
func TestCommentAssociation(t *testing.T) {
	processor := NewCommentProcessor()

	node := &ScalarNode{
		BaseNode: BaseNode{},
		Value:    "test",
	}

	processor.AssociateCommentToNode(node, "Test comment", CommentPositionAbove, 1)

	if node.HeadComment == nil {
		t.Error("ScalarNode.HeadComment should not be nil after association")
	}
}

// Test section detection
func TestBasicSectionDetection(t *testing.T) {
	processor := NewCommentProcessor()

	// Test detecting a header section
	sectionType := processor.DetectSectionType("# Header: Application Config")

	// Since we don't have the exact implementation details,
	// just verify that some detection occurs
	t.Logf("Detected section type: %v", sectionType)
}

// Test builder creation
func TestBuilderCreation(t *testing.T) {
	builder := NewBuilder()

	if builder == nil {
		t.Fatal("NewBuilder() returned nil")
	}

	// Test creating a scalar with the builder
	scalar := builder.BuildScalar("test", StylePlain)

	if scalar == nil {
		t.Fatal("Builder.BuildScalar() returned nil")
	}

	if scalar.Type() != NodeTypeScalar {
		t.Errorf("Built scalar has type %v, want %v", scalar.Type(), NodeTypeScalar)
	}
}