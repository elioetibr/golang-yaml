package node

import (
	"fmt"
	"testing"
)

// TestDemoSectionBasedYAML demonstrates the new section-based YAML structure
func TestDemoSectionBasedYAML(t *testing.T) {
	// Create a comment processor for intelligent comment handling
	processor := NewCommentProcessor()

	// Create a builder for constructing nodes
	builder := NewBuilder()

	// Create Header Section
	headerSection := &Section{
		ID:          "header",
		Type:        SectionTypeHeader,
		Title:       "Application Configuration",
		Description: "Main configuration file for the application",
		Nodes: []Node{
			builder.BuildMapping([]*MappingPair{
				{
					Key:   builder.BuildScalar("name", StylePlain),
					Value: builder.BuildScalar("my-application", StylePlain),
				},
				{
					Key:   builder.BuildScalar("version", StylePlain),
					Value: builder.BuildScalar("1.0.0", StylePlain),
				},
			}, StyleBlock),
		},
	}

	// Process header comments
	processor.AssociateCommentToNode(
		headerSection.Nodes[0],
		"# Application metadata",
		CommentPositionAbove,
		1,
	)

	// Create Configuration Section
	configSection := &Section{
		ID:    "configuration",
		Type:  SectionTypeConfiguration,
		Title: "Server Configuration",
		Nodes: []Node{
			builder.BuildMapping([]*MappingPair{
				{
					Key: builder.BuildScalar("server", StylePlain),
					Value: builder.BuildMapping([]*MappingPair{
						{
							Key:   builder.BuildScalar("host", StylePlain),
							Value: builder.BuildScalar("localhost", StylePlain),
						},
						{
							Key:   builder.BuildScalar("port", StylePlain),
							Value: builder.BuildScalar("8080", StylePlain),
						},
					}, StyleBlock),
				},
			}, StyleBlock),
		},
	}

	// Create Data Section
	dataSection := &Section{
		ID:    "data",
		Type:  SectionTypeData,
		Title: "Application Data",
		Nodes: []Node{
			builder.BuildMapping([]*MappingPair{
				{
					Key: builder.BuildScalar("features", StylePlain),
					Value: builder.BuildSequence([]Node{
						builder.BuildScalar("auth", StylePlain),
						builder.BuildScalar("api", StylePlain),
						builder.BuildScalar("metrics", StylePlain),
					}, StyleBlock),
				},
			}, StyleBlock),
		},
	}

	// Create a document with sections
	doc := builder.BuildDocument(
		[]*Section{headerSection, configSection, dataSection},
		[]Node{}, // No direct nodes outside sections
	)

	// Validate the document structure
	t.Run("Document Structure", func(t *testing.T) {
		if doc.Type() != NodeTypeDocument {
			t.Errorf("Document type = %v, want %v", doc.Type(), NodeTypeDocument)
		}

		if len(doc.Sections) != 3 {
			t.Errorf("Document has %d sections, want 3", len(doc.Sections))
		}

		// Validate section types
		expectedTypes := []SectionType{
			SectionTypeHeader,
			SectionTypeConfiguration,
			SectionTypeData,
		}

		for i, section := range doc.Sections {
			if section.Type != expectedTypes[i] {
				t.Errorf("Section[%d] type = %v, want %v", i, section.Type, expectedTypes[i])
			}
		}
	})

	// Test section detection from comments
	t.Run("Section Detection", func(t *testing.T) {
		comments := []string{
			"# Header: Application Info",
			"# Configuration: Database Settings",
			"# Data: User Preferences",
		}

		for _, comment := range comments {
			sectionType := processor.DetectSectionType(comment)
			t.Logf("Comment: %s -> Section Type: %v", comment, sectionType)
		}
	})

	// Test comment formatting with section context
	t.Run("Comment Formatting", func(t *testing.T) {
		// Different sections should have different comment rules
		headerComment := processor.ProcessComment(
			"Application name and version",
			CommentPositionAbove,
			SectionTypeHeader,
		)

		dataComment := processor.ProcessComment(
			"List of enabled features",
			CommentPositionAbove,
			SectionTypeData,
		)

		// Check that comments are formatted
		if headerComment == "" {
			t.Error("Header comment should not be empty")
		}

		if dataComment == "" {
			t.Error("Data comment should not be empty")
		}

		t.Logf("Header comment: %s", headerComment)
		t.Logf("Data comment: %s", dataComment)
	})

	// Demonstrate the builder pattern
	t.Run("Builder Pattern", func(t *testing.T) {
		// Create a complex nested structure
		complexNode := builder.BuildMapping([]*MappingPair{
			{
				Key: builder.BuildScalar("database", StylePlain),
				Value: builder.BuildMapping([]*MappingPair{
					{
						Key:   builder.BuildScalar("type", StylePlain),
						Value: builder.BuildScalar("postgres", StylePlain),
					},
					{
						Key: builder.BuildScalar("connections", StylePlain),
						Value: builder.BuildMapping([]*MappingPair{
							{
								Key:   builder.BuildScalar("max", StylePlain),
								Value: builder.BuildScalar("100", StylePlain),
							},
							{
								Key:   builder.BuildScalar("min", StylePlain),
								Value: builder.BuildScalar("10", StylePlain),
							},
						}, StyleBlock),
					},
				}, StyleBlock),
			},
		}, StyleBlock)

		if complexNode.Type() != NodeTypeMapping {
			t.Errorf("Complex node type = %v, want %v", complexNode.Type(), NodeTypeMapping)
		}

		// Add tags and anchors
		taggedNode := builder.WithTag(complexNode, "!!database")
		anchoredNode := builder.WithAnchor(taggedNode, "db-config")

		if complexNode.Tag() != "!!database" {
			t.Errorf("Node tag = %v, want '!!database'", complexNode.Tag())
		}

		if anchoredNode.Anchor() != "db-config" {
			t.Errorf("Node anchor = %v, want 'db-config'", anchoredNode.Anchor())
		}
	})

	fmt.Println("✅ All section-based YAML tests passed!")
}

// TestCommentRules demonstrates the comment rule system
func TestCommentRules(t *testing.T) {
	processor := NewCommentProcessor()

	// Test that different section types can have different rules
	testCases := []struct {
		comment     string
		position    CommentPosition
		sectionType SectionType
		description string
	}{
		{
			comment:     "Main application settings",
			position:    CommentPositionAbove,
			sectionType: SectionTypeHeader,
			description: "Header comment",
		},
		{
			comment:     "Database connection pool",
			position:    CommentPositionAbove,
			sectionType: SectionTypeConfiguration,
			description: "Configuration comment",
		},
		{
			comment:     "User data",
			position:    CommentPositionInline,
			sectionType: SectionTypeData,
			description: "Inline data comment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			processed := processor.ProcessComment(tc.comment, tc.position, tc.sectionType)
			if processed == "" {
				t.Errorf("%s: ProcessComment returned empty", tc.description)
			}
			t.Logf("%s: %s", tc.description, processed)
		})
	}
}
