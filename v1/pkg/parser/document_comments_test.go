package parser

import (
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestParseDocumentWithHeadComments(t *testing.T) {
	tests := []struct {
		name                  string
		yaml                  string
		expectDocComments     bool
		expectedCommentCount  int
		expectedFirstKey      string
		expectedMappingPairs  int
	}{
		{
			name: "YAML with leading comments",
			yaml: `# This is a comment
# Another comment
name: test
value: 123`,
			expectDocComments:    true,
			expectedCommentCount: 2,
			expectedFirstKey:     "name",
			expectedMappingPairs: 2,
		},
		{
			name: "YAML without leading comments",
			yaml: `name: test
value: 123`,
			expectDocComments:    false,
			expectedCommentCount: 0,
			expectedFirstKey:     "name",
			expectedMappingPairs: 2,
		},
		{
			name: "YAML with comments and blank lines",
			yaml: `# Document header
# Version: 1.0

# Configuration section
config:
  name: test`,
			expectDocComments:    true,
			expectedCommentCount: 2,
			expectedFirstKey:     "config",
			expectedMappingPairs: 1,
		},
		{
			name: "Comments with proper YAML structure",
			yaml: `config:
  # Comment inside structure
  name: test
  value: 123`,
			expectDocComments:    false,
			expectedCommentCount: 0,
			expectedFirstKey:     "config",
			expectedMappingPairs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexerFromString(tt.yaml)
			p := NewParser(l)

			result, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			// Check if we got a MappingNode
			mapping, ok := result.(*node.MappingNode)
			if !ok {
				t.Fatalf("Expected MappingNode, got %T", result)
			}

			// Check document head comments
			if tt.expectDocComments {
				if !mapping.HasDocumentHeadComments {
					t.Error("Expected HasDocumentHeadComments to be true")
				}

				if mapping.HeadComment == nil {
					t.Fatal("Expected HeadComment to be non-nil")
				}

				if len(mapping.HeadComment.Comments) != tt.expectedCommentCount {
					t.Errorf("Expected %d comments, got %d",
						tt.expectedCommentCount, len(mapping.HeadComment.Comments))
				}

				// Verify comments don't include the # prefix
				for _, comment := range mapping.HeadComment.Comments {
					if strings.HasPrefix(comment, "#") {
						t.Errorf("Comment should not include # prefix: %s", comment)
					}
				}
			} else {
				if mapping.HasDocumentHeadComments {
					t.Error("Expected HasDocumentHeadComments to be false")
				}
			}

			// Check mapping pairs
			if len(mapping.Pairs) != tt.expectedMappingPairs {
				t.Errorf("Expected %d mapping pairs, got %d",
					tt.expectedMappingPairs, len(mapping.Pairs))
			}

			// Check first key
			if len(mapping.Pairs) > 0 {
				firstPair := mapping.Pairs[0]
				if keyNode, ok := firstPair.Key.(*node.ScalarNode); ok {
					if keyNode.Value != tt.expectedFirstKey {
						t.Errorf("Expected first key '%s', got '%s'",
							tt.expectedFirstKey, keyNode.Value)
					}
				} else {
					t.Errorf("Expected first key to be ScalarNode, got %T", firstPair.Key)
				}
			}
		})
	}
}

func TestParseComplexDocumentWithComments(t *testing.T) {
	yaml := `# Application Configuration
# Version: 2.0
# Environment: Production

# Database settings
database:
  host: localhost
  port: 5432
  # Connection pool settings
  pool:
    min: 5
    max: 20

# Cache configuration
cache:
  enabled: true
  ttl: 3600`

	l := lexer.NewLexerFromString(yaml)
	p := NewParser(l)

	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	mapping, ok := result.(*node.MappingNode)
	if !ok {
		t.Fatalf("Expected MappingNode, got %T", result)
	}

	// Should have document head comments
	if !mapping.HasDocumentHeadComments {
		t.Error("Expected document to have head comments")
	}

	if mapping.HeadComment == nil {
		t.Fatal("Expected HeadComment to be non-nil")
	}

	// Should have 3 head comments (before any content)
	if len(mapping.HeadComment.Comments) != 3 {
		t.Errorf("Expected 3 head comments, got %d: %v",
			len(mapping.HeadComment.Comments), mapping.HeadComment.Comments)
	}

	// Should have 2 top-level keys (database, cache)
	if len(mapping.Pairs) != 2 {
		t.Errorf("Expected 2 mapping pairs, got %d", len(mapping.Pairs))
	}

	// First key should be "database"
	if len(mapping.Pairs) > 0 {
		firstPair := mapping.Pairs[0]
		if keyNode, ok := firstPair.Key.(*node.ScalarNode); ok {
			if keyNode.Value != "database" {
				t.Errorf("Expected first key 'database', got '%s'", keyNode.Value)
			}
		}
	}
}

func TestParseOnlyComments(t *testing.T) {
	yaml := `# Only comments
# No actual YAML content`

	l := lexer.NewLexerFromString(yaml)
	p := NewParser(l)

	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// When there's only comments and no content, parser returns nil
	// This is expected behavior - comments alone don't constitute a document
	if result != nil {
		t.Errorf("Expected nil for comment-only document, got %T", result)
	}
}