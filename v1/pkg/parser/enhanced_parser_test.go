package parser_test

import (
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
	"github.com/elioetibr/golang-yaml/v1/pkg/parser"
)

func TestEnhancedParserCommentAssociation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, root node.Node)
	}{
		{
			name: "Document level comments",
			input: `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

key: value`,
			validate: func(t *testing.T, root node.Node) {
				if root == nil {
					t.Fatal("Expected non-nil root node")
				}

				// Check document has header comments
				if doc, ok := root.(*node.DocumentNode); ok {
					if doc.HeadComment == nil {
						t.Error("Expected document header comments")
					} else if len(doc.HeadComment.Comments) != 5 {
						t.Errorf("Expected 5 header comments, got %d", len(doc.HeadComment.Comments))
					}
				}
			},
		},
		{
			name: "Hierarchical comment association",
			input: `# @schema
# additionalProperties: true
# @schema
# -- Pod Disruption Budget configuration
podDisruptionBudget:
  # -- Enable Pod Disruption Budget
  enabled: false
  # -- Maximum pods unavailable
  maxUnavailable: 1
  # Alternative: minAvailable: 1`,
			validate: func(t *testing.T, root node.Node) {
				// Find the podDisruptionBudget mapping
				mapping := findMapping(root, "podDisruptionBudget")
				if mapping == nil {
					t.Fatal("Expected podDisruptionBudget mapping")
				}

				// Check it has header comments
				if mapping.HeadComment == nil {
					t.Error("Expected header comments on podDisruptionBudget")
				}

				// Check child nodes have proper comments
				for _, pair := range mapping.Pairs {
					if scalar, ok := pair.Key.(*node.ScalarNode); ok {
						switch scalar.Value {
						case "enabled":
							if pair.Key.(*node.ScalarNode).HeadComment == nil {
								t.Error("Expected header comment on 'enabled' key")
							}
						case "maxUnavailable":
							if pair.Key.(*node.ScalarNode).HeadComment == nil {
								t.Error("Expected header comment on 'maxUnavailable' key")
							}
							// Should also have footer comment
							if pair.Value != nil {
								valueNode := pair.Value
								if valueNode.(*node.ScalarNode).FootComment == nil {
									t.Error("Expected footer comment after 'maxUnavailable' value")
								}
							}
						}
					}
				}
			},
		},
		{
			name: "Section detection from empty lines",
			input: `# First section
key1: value1


# Second section
key2: value2

# Third section
key3: value3`,
			validate: func(t *testing.T, root node.Node) {
				// Should detect sections based on multiple empty lines
				if doc, ok := root.(*node.DocumentNode); ok {
					if len(doc.Sections) < 2 {
						t.Errorf("Expected at least 2 sections, got %d", len(doc.Sections))
					}
				}
			},
		},
		{
			name: "Inline comments",
			input: `employees:
  alice@example.com: # Alice's email
    name: Alice
    department: Engineering # Tech department
  bob@example.com: # Bob's email
    name: Bob`,
			validate: func(t *testing.T, root node.Node) {
				mapping := findMapping(root, "employees")
				if mapping == nil {
					t.Fatal("Expected employees mapping")
				}

				// Check inline comments on email keys
				for _, pair := range mapping.Pairs {
					if key, ok := pair.Key.(*node.ScalarNode); ok {
						if strings.Contains(key.Value, "@example.com") {
							if key.LineComment == nil {
								t.Errorf("Expected inline comment on email key: %s", key.Value)
							}
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(strings.NewReader(tt.input))
			p := parser.NewEnhancedParser(l, parser.DefaultParserOptions())

			root, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			tt.validate(t, root)
		})
	}
}

func TestEnhancedParserEmptyLinePreservation(t *testing.T) {
	input := `# Header comment

key1: value1

# Section comment

key2: value2


# Another section
key3: value3`

	l := lexer.NewLexer(strings.NewReader(input))
	opts := &parser.ParserOptions{
		PreserveComments:            true,
		PreserveEmptyLines:          true,
		KeepSectionBoundaries:       true,
		DefaultLinesBetweenSections: 1,
		AutoDetectSections:          true,
	}

	p := parser.NewEnhancedParser(l, opts)
	root, err := p.Parse()

	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Verify sections were detected
	if doc, ok := root.(*node.DocumentNode); ok {
		if len(doc.Sections) == 0 {
			t.Error("Expected sections to be detected")
		}

		// Check that sections have proper formatting
		for _, section := range doc.Sections {
			if section.Formatting == nil {
				t.Error("Expected section formatting")
			}
		}
	}
}

func TestEnhancedParserMergeStrategy(t *testing.T) {
	base := `# Base configuration
name: MyApp
version: 1.0.0

# Settings
settings:
  debug: false
  port: 8080`

	overlay := `# Updated configuration
version: 2.0.0

settings:
  debug: true
  # New setting
  timeout: 30`

	// Parse base
	l1 := lexer.NewLexer(strings.NewReader(base))
	p1 := parser.NewEnhancedParser(l1, parser.DefaultParserOptions())
	baseRoot, err := p1.Parse()
	if err != nil {
		t.Fatalf("Failed to parse base: %v", err)
	}

	// Parse overlay
	l2 := lexer.NewLexer(strings.NewReader(overlay))
	p2 := parser.NewEnhancedParser(l2, parser.DefaultParserOptions())
	overlayRoot, err := p2.Parse()
	if err != nil {
		t.Fatalf("Failed to parse overlay: %v", err)
	}

	// Verify both preserve comments
	if baseRoot.(*node.DocumentNode).HeadComment == nil {
		t.Error("Base should have header comments")
	}

	if overlayRoot.(*node.DocumentNode).HeadComment == nil {
		t.Error("Overlay should have header comments")
	}
}

// Helper functions

func findMapping(n node.Node, key string) *node.MappingNode {
	var result *node.MappingNode

	visitor := &mappingFinder{
		targetKey: key,
		result:    &result,
	}

	n.Accept(visitor)
	return result
}

type mappingFinder struct {
	targetKey string
	result    **node.MappingNode
}

func (f *mappingFinder) VisitScalar(n *node.ScalarNode) error {
	return nil
}

func (f *mappingFinder) VisitSequence(n *node.SequenceNode) error {
	for _, item := range n.Items {
		item.Accept(f)
	}
	return nil
}

func (f *mappingFinder) VisitMapping(n *node.MappingNode) error {
	for _, pair := range n.Pairs {
		if key, ok := pair.Key.(*node.ScalarNode); ok && key.Value == f.targetKey {
			if value, ok := pair.Value.(*node.MappingNode); ok {
				*f.result = value
				return nil
			}
		}
		// Continue searching in nested nodes
		if pair.Value != nil {
			pair.Value.Accept(f)
		}
	}
	return nil
}

func (f *mappingFinder) VisitSection(n *node.SectionNode) error {
	return nil
}

func (f *mappingFinder) VisitDocument(n *node.DocumentNode) error {
	if n.Content != nil {
		n.Content.Accept(f)
	}
	for _, node := range n.Nodes {
		node.Accept(f)
	}
	return nil
}