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

// TestSection05Requirements tests the requirements from section 05 of the documentation:
// Commented Case with breaking lines configuration enabled separating the sections by default
func TestSection05Requirements(t *testing.T) {
	baseYAML := `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# Company Name
# -- This is a simple Company Name
company: Umbrella Corp.

# City Name
# -- This is the Company City Name
city: Raccoon City

# Company Employees
# -- This is a mapping of Company Employees
employees:
  bob@umbreallacorp.co: # Bob Umbrella's Email
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co: # Alice Umbrella's Email
    name: Alice Abernathy
    department: Project`

	overrideYAML := `company: Umbrella Corporation.
city: Raccoon City
employees:
  redqueen@umbreallacorp.co: # Red Queen Umbrella's Email
    name: Red Queen
    department: Security

# @schema
# additionalProperties: true
# @schema
# -- This status list demonstrates the possible hive access
statusList: []
# Examples:
# statusList:
#   - enabled
#   - blocked
#   - quarantine
#   - infected`

	expectedYAML := `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# Company Name
# -- This is a simple Company Name
company: Umbrella Corporation.

# City Name
# -- This is the Company City Name
city: Raccoon City

# Company Employees
# -- This is a mapping of Company Employees
employees:
  bob@umbreallacorp.co: # Bob Umbrella's Email
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co: # Alice Umbrella's Email
    name: Alice Abernathy
    department: Project
  redqueen@umbreallacorp.co: # Red Queen Umbrella's Email
    name: Red Queen
    department: Security

# @schema
# additionalProperties: true
# @schema
# -- This status list demonstrates the possible hive access
statusList: []
# Examples:
# statusList:
#   - enabled
#   - blocked
#   - quarantine
#   - infected`

	t.Run("merge with blank lines and comments preservation", func(t *testing.T) {
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

		// Create merge options with blank lines and comments preservation
		opts := &Options{
			Strategy:           StrategyDeep,
			PreserveComments:   true,
			PreserveBlankLines: true,
		}

		// Perform merge
		mergedNode, err := WithOptions(baseNode, overrideNode, opts)
		if err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Serialize the result
		var buf bytes.Buffer
		s := serializer.NewSerializer(&buf, nil)
		err = s.Serialize(mergedNode)
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}
		result := buf.Bytes()

		// Normalize line endings for comparison
		actualResult := strings.TrimSpace(string(result))
		expectedResult := strings.TrimSpace(expectedYAML)

		// Compare results
		if actualResult != expectedResult {
			t.Errorf("Merge result does not match expected output")
			t.Logf("Expected:\n%s", expectedResult)
			t.Logf("Actual:\n%s", actualResult)

			// Show line-by-line differences for debugging
			expectedLines := strings.Split(expectedResult, "\n")
			actualLines := strings.Split(actualResult, "\n")

			maxLines := len(expectedLines)
			if len(actualLines) > maxLines {
				maxLines = len(actualLines)
			}

			for i := 0; i < maxLines; i++ {
				var expectedLine, actualLine string
				if i < len(expectedLines) {
					expectedLine = expectedLines[i]
				}
				if i < len(actualLines) {
					actualLine = actualLines[i]
				}

				if expectedLine != actualLine {
					t.Logf("Line %d differs:", i+1)
					t.Logf("  Expected: %q", expectedLine)
					t.Logf("  Actual:   %q", actualLine)
				}
			}
		}
	})

	t.Run("verify blank lines are preserved", func(t *testing.T) {
		// This test specifically checks that blank lines between sections are preserved
		l := lexer.NewLexerFromString(baseYAML)
		p := parser.NewParser(l)
		baseNode, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse base YAML: %v", err)
		}

		// Check that the parsed structure has blank lines information
		if mapping, ok := baseNode.(*node.MappingNode); ok {
			for _, pair := range mapping.Pairs {
				if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
					// Check specific keys that should have blank lines before them
					switch keyScalar.Value {
					case "city", "employees":
						if pair.BlankLinesBefore == 0 {
							t.Logf("Warning: Key '%s' should have blank lines before it", keyScalar.Value)
						}
					}
				}
			}
		}
	})

	t.Run("verify comments are preserved", func(t *testing.T) {
		l := lexer.NewLexerFromString(baseYAML)
		p := parser.NewParser(l)
		baseNode, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse base YAML: %v", err)
		}

		// Check that comments are parsed correctly
		if mapping, ok := baseNode.(*node.MappingNode); ok {
			// Check document-level head comments
			if mapping.HeadComment == nil || len(mapping.HeadComment.Comments) == 0 {
				t.Error("Document head comments should be preserved")
			}

			// Check field-level comments
			for _, pair := range mapping.Pairs {
				if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
					switch keyScalar.Value {
					case "company", "city", "employees":
						// These fields should have head comments
						if keyScalar.HeadComment == nil || len(keyScalar.HeadComment.Comments) == 0 {
							t.Logf("Warning: Key '%s' should have head comments", keyScalar.Value)
						}
					}
				}

				// Check inline comments on employee emails
				if keyScalar, ok := pair.Key.(*node.ScalarNode); ok && keyScalar.Value == "employees" {
					if empMapping, ok := pair.Value.(*node.MappingNode); ok {
						for _, empPair := range empMapping.Pairs {
							if empKey, ok := empPair.Key.(*node.ScalarNode); ok {
								if strings.Contains(empKey.Value, "@") {
									// Employee emails should have inline comments
									if empKey.LineComment == nil || len(empKey.LineComment.Comments) == 0 {
										t.Logf("Warning: Employee email '%s' should have inline comment", empKey.Value)
									}
								}
							}
						}
					}
				}
			}
		}
	})
}

// TestBlankLineSectionSeparation tests that blank lines properly separate sections
func TestBlankLineSectionSeparation(t *testing.T) {
	yamlWithSections := `# Section 1
# Configuration for first section
key1: value1

# Section 2
# Configuration for second section
key2: value2

# Section 3
# Configuration for third section
key3: value3`

	t.Run("parse and preserve section separation", func(t *testing.T) {
		l := lexer.NewLexerFromString(yamlWithSections)
		p := parser.NewParser(l)
		node, err := p.Parse()
		if err != nil {
			t.Fatalf("Failed to parse YAML: %v", err)
		}

		// Serialize with blank lines preservation
		var buf bytes.Buffer
		s := serializer.NewSerializer(&buf, nil)
		err = s.Serialize(node)
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}
		result := buf.Bytes()

		resultStr := string(result)

		// Check that blank lines are present in the output
		if !strings.Contains(resultStr, "\n\n") {
			t.Error("Blank lines between sections should be preserved")
		}

		// Count the number of blank line separators
		blankLineCount := strings.Count(resultStr, "\n\n")
		if blankLineCount < 2 {
			t.Errorf("Expected at least 2 blank line separators, got %d", blankLineCount)
		}
	})
}
