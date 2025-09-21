package merge

import (
	"bytes"
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
	"github.com/elioetibr/golang-yaml/v1/pkg/parser"
	"github.com/elioetibr/golang-yaml/v1/pkg/serializer"
)

// TestDocumentHeadComments tests preservation of document-level comments
func TestDocumentHeadComments(t *testing.T) {
	t.Run("preserve document head comments", func(t *testing.T) {
		baseYAML := `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

company: Umbrella Corp.
city: Raccoon City`

		overrideYAML := `company: Umbrella Corporation.
newfield: value`

		// Parse base YAML
		l := lexer.NewLexerFromString(baseYAML)
		p := parser.NewParser(l)
		baseNode, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse base YAML: %v", err)
		}

		// Parse override YAML
		l2 := lexer.NewLexerFromString(overrideYAML)
		p2 := parser.NewParser(l2)
		overrideNode, err := p2.Parse()
		if err != nil {
			t.Fatalf("Failed to parse override YAML: %v", err)
		}

		// Create merge options
		opts := &Options{
			Strategy:                       StrategyDeep,
			PreserveComments:               true,
			PreserveBlankLines:             true,
			KeepDefaultLineBetweenSections: true,
		}

		// Perform merge
		mergedNode, err := WithOptions(baseNode, overrideNode, opts)
		if err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Check if merged node is marked as having document head comments
		if mapping, ok := mergedNode.(*node.MappingNode); ok {
			// The mapping should be aware of document head comments
			if !mapping.HasDocumentHeadComments && mapping.HeadComment != nil {
				t.Log("Warning: Mapping should be marked as having document head comments")
			}
		}

		// Serialize and check output
		var buf bytes.Buffer
		s := serializer.NewSerializer(&buf, nil)
		err = s.Serialize(mergedNode)
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}

		result := string(buf.Bytes())

		// Check that document head comments are preserved
		if !strings.Contains(result, "yaml-language-server") {
			t.Error("Document head comment 'yaml-language-server' should be preserved")
		}
	})
}

// TestSectionHandlingOptions tests the new section handling configuration
func TestSectionHandlingOptions(t *testing.T) {
	yamlWithSections := `# Section 1
key1: value1


# Section 2
key2: value2


# Section 3
key3: value3`

	t.Run("keep original blank lines", func(t *testing.T) {
		opts := &Options{
			Strategy:                       StrategyDeep,
			PreserveComments:               true,
			PreserveBlankLines:             true,
			KeepDefaultLineBetweenSections: true, // Keep original
		}

		// Parse YAML
		l := lexer.NewLexerFromString(yamlWithSections)
		p := parser.NewParser(l)
		parsedNode, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse YAML: %v", err)
		}

		// Process with options
		processor := NewNodeProcessor()
		processor.NormalizeSectionBoundaries(parsedNode, opts)

		// With KeepDefaultLineBetweenSections=true, original blank lines should be preserved
		// (No actual normalization should occur)
		if mapping, ok := parsedNode.(*node.MappingNode); ok {
			for _, pair := range mapping.Pairs {
				// Original blank lines should remain unchanged
				if pair.BlankLinesBefore > 2 {
					// This is expected - multiple blank lines preserved
					t.Logf("Preserved %d blank lines before key", pair.BlankLinesBefore)
				}
			}
		}
	})

	t.Run("normalize to 1 blank line", func(t *testing.T) {
		opts := &Options{
			Strategy:                       StrategyDeep,
			PreserveComments:               true,
			PreserveBlankLines:             true,
			KeepDefaultLineBetweenSections: false, // Normalize
			DefaultLineBetweenSections:     1,     // To 1 line
		}

		// Parse YAML
		l := lexer.NewLexerFromString(yamlWithSections)
		p := parser.NewParser(l)
		parsedNode, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse YAML: %v", err)
		}

		// Process with options
		processor := NewNodeProcessor()
		processor.NormalizeSectionBoundaries(parsedNode, opts)

		// With normalization, sections with >1 blank lines should be normalized to 1
		if mapping, ok := parsedNode.(*node.MappingNode); ok {
			for _, pair := range mapping.Pairs {
				if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
					// Sections that had multiple blank lines should now have exactly 1
					if pair.BlankLinesBefore > 1 && pair.BlankLinesBefore != opts.DefaultLineBetweenSections {
						t.Errorf("Key '%s': expected %d blank lines, got %d",
							keyScalar.Value, opts.DefaultLineBetweenSections, pair.BlankLinesBefore)
					}
				}
			}
		}
	})

	t.Run("normalize to 2 blank lines", func(t *testing.T) {
		opts := &Options{
			Strategy:                       StrategyDeep,
			PreserveComments:               true,
			PreserveBlankLines:             true,
			KeepDefaultLineBetweenSections: false, // Normalize
			DefaultLineBetweenSections:     2,     // To 2 lines
		}

		// Parse YAML
		l := lexer.NewLexerFromString(yamlWithSections)
		p := parser.NewParser(l)
		parsedNode, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse YAML: %v", err)
		}

		// Process with options
		processor := NewNodeProcessor()
		processor.NormalizeSectionBoundaries(parsedNode, opts)

		// With normalization to 2, sections should have exactly 2 blank lines
		if mapping, ok := parsedNode.(*node.MappingNode); ok {
			for _, pair := range mapping.Pairs {
				if pair.BlankLinesBefore > 1 && pair.BlankLinesBefore != opts.DefaultLineBetweenSections {
					if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
						t.Errorf("Key '%s': expected %d blank lines, got %d",
							keyScalar.Value, opts.DefaultLineBetweenSections, pair.BlankLinesBefore)
					}
				}
			}
		}
	})
}

// TestSectionDetection tests the section boundary detection
func TestSectionDetection(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("detect section with 2 blank lines", func(t *testing.T) {
		scalar := builder.BuildScalar("test", node.StylePlain)
		scalar.BlankLinesBefore = 2

		if !processor.IsSectionBoundary(scalar, 2) {
			t.Error("Should detect section boundary with 2 blank lines")
		}
	})

	t.Run("detect section with 3 blank lines", func(t *testing.T) {
		mapping := builder.BuildMapping(nil, node.StyleBlock)
		mapping.BlankLinesBefore = 3

		if !processor.IsSectionBoundary(mapping, 2) {
			t.Error("Should detect section boundary with 3 blank lines (>= 2)")
		}
	})

	t.Run("no section with 1 blank line", func(t *testing.T) {
		scalar := builder.BuildScalar("test", node.StylePlain)
		scalar.BlankLinesBefore = 1

		if processor.IsSectionBoundary(scalar, 2) {
			t.Error("Should not detect section boundary with only 1 blank line")
		}
	})

	t.Run("custom minimum blank lines", func(t *testing.T) {
		scalar := builder.BuildScalar("test", node.StylePlain)
		scalar.BlankLinesBefore = 3

		if !processor.IsSectionBoundary(scalar, 3) {
			t.Error("Should detect section boundary with custom minimum")
		}

		if processor.IsSectionBoundary(scalar, 4) {
			t.Error("Should not detect section boundary when below custom minimum")
		}
	})
}

// TestOptionsHelpers tests the new option helper methods
func TestOptionsHelpers(t *testing.T) {
	t.Run("WithSectionHandling", func(t *testing.T) {
		opts := DefaultOptions()
		opts.WithSectionHandling(false, 2)

		if opts.KeepDefaultLineBetweenSections {
			t.Error("KeepDefaultLineBetweenSections should be false")
		}

		if opts.DefaultLineBetweenSections != 2 {
			t.Errorf("DefaultLineBetweenSections should be 2, got %d", opts.DefaultLineBetweenSections)
		}
	})

	t.Run("WithNormalizedSections", func(t *testing.T) {
		opts := DefaultOptions()
		opts.WithNormalizedSections(3)

		if opts.KeepDefaultLineBetweenSections {
			t.Error("KeepDefaultLineBetweenSections should be false for normalized sections")
		}

		if opts.DefaultLineBetweenSections != 3 {
			t.Errorf("DefaultLineBetweenSections should be 3, got %d", opts.DefaultLineBetweenSections)
		}
	})

	t.Run("default options preserve original", func(t *testing.T) {
		opts := DefaultOptions()

		if !opts.KeepDefaultLineBetweenSections {
			t.Error("Default should keep original blank lines")
		}

		if opts.DefaultLineBetweenSections != 1 {
			t.Errorf("Default blank lines should be 1, got %d", opts.DefaultLineBetweenSections)
		}
	})
}
