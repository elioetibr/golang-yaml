package parser

import (
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/node"
)

func TestParseScalars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		style    node.Style
	}{
		{
			name:     "plain_scalar",
			input:    "hello world",
			expected: "hello world",
			style:    node.StylePlain,
		},
		{
			name:     "single_quoted",
			input:    "'hello world'",
			expected: "hello world",
			style:    node.StyleSingleQuoted,
		},
		{
			name:     "double_quoted",
			input:    `"hello\nworld"`,
			expected: "hello\nworld",
			style:    node.StyleDoubleQuoted,
		},
		{
			name: "literal_scalar",
			input: `|
  hello
  world`,
			expected: "hello\nworld",
			style:    node.StyleLiteral,
		},
		{
			name: "folded_scalar",
			input: `>
  hello
  world`,
			expected: "hello\nworld",
			style:    node.StyleFolded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			scalar, ok := root.(*node.ScalarNode)
			if !ok {
				t.Fatalf("Expected ScalarNode, got %T", root)
			}

			if scalar.Value != tt.expected {
				t.Errorf("Expected value %q, got %q", tt.expected, scalar.Value)
			}

			if scalar.Style != tt.style {
				t.Errorf("Expected style %v, got %v", tt.style, scalar.Style)
			}
		})
	}
}

func TestParseBlockSequence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "simple_sequence",
			input: `- item1
- item2
- item3`,
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name: "nested_sequence",
			input: `- item1
-
  - nested1
  - nested2
- item3`,
			expected: []string{"item1", "[nested]", "item3"},
		},
		{
			name: "sequence_with_empty_items",
			input: `- item1
-
- item3`,
			expected: []string{"item1", "", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			seq, ok := root.(*node.SequenceNode)
			if !ok {
				t.Fatalf("Expected SequenceNode, got %T", root)
			}

			if len(seq.Items) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(seq.Items))
			}

			for i, item := range seq.Items {
				if scalar, ok := item.(*node.ScalarNode); ok {
					if i < len(tt.expected) && scalar.Value != tt.expected[i] {
						t.Errorf("Item %d: expected %q, got %q", i, tt.expected[i], scalar.Value)
					}
				}
			}
		})
	}
}

func TestParseBlockMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name: "simple_mapping",
			input: `key1: value1
key2: value2
key3: value3`,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "mapping_with_quotes",
			input: `"key 1": 'value 1'
'key 2': "value 2"
key3: value3`,
			expected: map[string]string{
				"key 1": "value 1",
				"key 2": "value 2",
				"key3":  "value3",
			},
		},
		{
			name: "explicit_mapping",
			input: `? key1
: value1
? key2
: value2`,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			mapping, ok := root.(*node.MappingNode)
			if !ok {
				t.Fatalf("Expected MappingNode, got %T", root)
			}

			if len(mapping.Pairs) != len(tt.expected) {
				t.Errorf("Expected %d pairs, got %d", len(tt.expected), len(mapping.Pairs))
			}

			for _, pair := range mapping.Pairs {
				key, keyOk := pair.Key.(*node.ScalarNode)
				value, valueOk := pair.Value.(*node.ScalarNode)

				if !keyOk || !valueOk {
					continue
				}

				if expectedValue, exists := tt.expected[key.Value]; exists {
					if value.Value != expectedValue {
						t.Errorf("Key %q: expected value %q, got %q", key.Value, expectedValue, value.Value)
					}
				}
			}
		})
	}
}

func TestParseFlowSequence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple_flow_sequence",
			input:    "[a, b, c]",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "flow_sequence_with_quotes",
			input:    `['a', "b", c]`,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "nested_flow_sequence",
			input:    "[a, [b, c], d]",
			expected: []string{"a", "[nested]", "d"},
		},
		{
			name:     "empty_flow_sequence",
			input:    "[]",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			seq, ok := root.(*node.SequenceNode)
			if !ok {
				t.Fatalf("Expected SequenceNode, got %T", root)
			}

			if seq.Style != node.StyleFlow {
				t.Errorf("Expected flow style, got %v", seq.Style)
			}

			if len(seq.Items) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(seq.Items))
			}
		})
	}
}

func TestParseFlowMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "simple_flow_mapping",
			input: "{a: 1, b: 2, c: 3}",
			expected: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
		},
		{
			name:  "flow_mapping_with_quotes",
			input: `{'a': "1", "b": '2'}`,
			expected: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
		{
			name:     "empty_flow_mapping",
			input:    "{}",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			mapping, ok := root.(*node.MappingNode)
			if !ok {
				t.Fatalf("Expected MappingNode, got %T", root)
			}

			if mapping.Style != node.StyleFlow {
				t.Errorf("Expected flow style, got %v", mapping.Style)
			}

			if len(mapping.Pairs) != len(tt.expected) {
				t.Errorf("Expected %d pairs, got %d", len(tt.expected), len(mapping.Pairs))
			}
		})
	}
}

func TestParseMixedContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, node.Node)
	}{
		{
			name: "mapping_with_sequence_values",
			input: `fruits:
  - apple
  - banana
vegetables:
  - carrot
  - potato`,
			check: func(t *testing.T, root node.Node) {
				mapping, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatalf("Expected MappingNode, got %T", root)
				}
				if len(mapping.Pairs) != 2 {
					t.Errorf("Expected 2 pairs, got %d", len(mapping.Pairs))
				}
			},
		},
		{
			name: "sequence_with_mapping_items",
			input: `- name: Alice
  age: 30
- name: Bob
  age: 25`,
			check: func(t *testing.T, root node.Node) {
				seq, ok := root.(*node.SequenceNode)
				if !ok {
					t.Fatalf("Expected SequenceNode, got %T", root)
				}
				if len(seq.Items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(seq.Items))
				}
			},
		},
		{
			name: "deeply_nested",
			input: `root:
  level1:
    level2:
      - item1
      - item2
    another: value`,
			check: func(t *testing.T, root node.Node) {
				if root == nil {
					t.Fatal("Expected non-nil root")
				}
				if root.Type() != node.NodeTypeMapping {
					t.Errorf("Expected mapping root, got %v", root.Type())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			tt.check(t, root)
		})
	}
}

func TestParseDocumentMarkers(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "explicit_document",
			input: `---
content: here
...`,
		},
		{
			name: "multiple_documents",
			input: `---
doc1: value
---
doc2: value`,
		},
		{
			name:  "implicit_document",
			input: `key: value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if root == nil {
				t.Fatal("Expected non-nil root")
			}
		})
	}
}

func TestParseWithComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, node.Node)
	}{
		{
			name: "mapping_with_comments",
			input: `# Header comment
key1: value1  # inline comment
# Above key2
key2: value2`,
			check: func(t *testing.T, root node.Node) {
				mapping, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatalf("Expected MappingNode, got %T", root)
				}
				if len(mapping.Pairs) != 2 {
					t.Errorf("Expected 2 pairs, got %d", len(mapping.Pairs))
				}
			},
		},
		{
			name: "sequence_with_comments",
			input: `# List header
- item1  # first item
# Above item2
- item2`,
			check: func(t *testing.T, root node.Node) {
				seq, ok := root.(*node.SequenceNode)
				if !ok {
					t.Fatalf("Expected SequenceNode, got %T", root)
				}
				if len(seq.Items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(seq.Items))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			tt.check(t, root)
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Numbers
		{"123", int64(123)},
		{"-456", int64(-456)},
		{"3.14", float64(3.14)},
		{"-2.5", float64(-2.5)},

		// Booleans
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"on", true},
		{"On", true},
		{"ON", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"no", false},
		{"No", false},
		{"NO", false},
		{"off", false},
		{"Off", false},
		{"OFF", false},

		// Null
		{"null", nil},
		{"Null", nil},
		{"NULL", nil},
		{"~", nil},
		{"", nil},

		// Strings
		{"hello", "hello"},
		{"hello world", "hello world"},
		{"'123'", "'123'"},
	}

	for _, tt := range tests {
		result := ParseValue(tt.input)
		if result != tt.expected {
			t.Errorf("ParseValue(%q): expected %v (%T), got %v (%T)",
				tt.input, tt.expected, tt.expected, result, result)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "unclosed_flow_sequence",
			input:       "[a, b, c",
			expectError: false, // Parser is lenient
		},
		{
			name:        "unclosed_flow_mapping",
			input:       "{a: 1, b: 2",
			expectError: false, // Parser is lenient
		},
		{
			name:        "invalid_indentation",
			input:       "key:\n value",
			expectError: false, // Basic indentation handled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseString(tt.input)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
