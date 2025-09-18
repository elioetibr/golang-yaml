package serializer

import (
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/node"
)

func TestSerializeScalar(t *testing.T) {
	tests := []struct {
		name     string
		node     *node.ScalarNode
		expected string
		options  *Options
	}{
		{
			name: "plain_scalar",
			node: &node.ScalarNode{
				Value: "hello world",
				Style: node.StylePlain,
			},
			expected: "hello world",
		},
		{
			name: "single_quoted_scalar",
			node: &node.ScalarNode{
				Value: "hello world",
				Style: node.StyleSingleQuoted,
			},
			expected: "'hello world'",
		},
		{
			name: "double_quoted_scalar",
			node: &node.ScalarNode{
				Value: "hello\nworld",
				Style: node.StyleDoubleQuoted,
			},
			expected: `"hello\nworld"`,
		},
		{
			name: "literal_scalar",
			node: &node.ScalarNode{
				Value: "hello\nworld",
				Style: node.StyleLiteral,
			},
			expected: "|\n  hello\n  world\n",
		},
		{
			name: "folded_scalar",
			node: &node.ScalarNode{
				Value: "hello\nworld",
				Style: node.StyleFolded,
			},
			expected: ">\n  hello\n  world\n",
		},
		{
			name: "plain_boolean_literal",
			node: &node.ScalarNode{
				Value: "true",
				Style: node.StylePlain,
			},
			expected: `true`, // Boolean literals are not quoted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.options
			if opts == nil {
				opts = DefaultOptions()
			}

			result, err := SerializeToString(tt.node, opts)
			if err != nil {
				t.Fatalf("Serialize error: %v", err)
			}

			// Trim trailing newline for comparison
			result = strings.TrimSuffix(result, "\n")

			if result != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, result)
			}
		})
	}
}

func TestSerializeSequence(t *testing.T) {
	builder := &node.DefaultBuilder{}

	tests := []struct {
		name     string
		node     *node.SequenceNode
		expected string
		options  *Options
	}{
		{
			name: "block_sequence",
			node: builder.BuildSequence([]node.Node{
				builder.BuildScalar("item1", node.StylePlain),
				builder.BuildScalar("item2", node.StylePlain),
				builder.BuildScalar("item3", node.StylePlain),
			}, node.StyleBlock),
			expected: "- item1\n- item2\n- item3",
		},
		{
			name: "flow_sequence",
			node: builder.BuildSequence([]node.Node{
				builder.BuildScalar("a", node.StylePlain),
				builder.BuildScalar("b", node.StylePlain),
				builder.BuildScalar("c", node.StylePlain),
			}, node.StyleFlow),
			expected: "[a, b, c]",
		},
		{
			name:     "empty_sequence",
			node:     builder.BuildSequence([]node.Node{}, node.StyleFlow),
			expected: "[]",
		},
		{
			name: "nested_sequence",
			node: builder.BuildSequence([]node.Node{
				builder.BuildScalar("item1", node.StylePlain),
				builder.BuildSequence([]node.Node{
					builder.BuildScalar("nested1", node.StylePlain),
					builder.BuildScalar("nested2", node.StylePlain),
				}, node.StyleBlock),
			}, node.StyleBlock),
			expected: "- item1\n-\n  - nested1\n  - nested2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.options
			if opts == nil {
				opts = DefaultOptions()
			}

			result, err := SerializeToString(tt.node, opts)
			if err != nil {
				t.Fatalf("Serialize error: %v", err)
			}

			// Trim trailing newline for comparison
			result = strings.TrimSuffix(result, "\n")

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestSerializeMapping(t *testing.T) {
	builder := &node.DefaultBuilder{}

	tests := []struct {
		name     string
		node     *node.MappingNode
		expected string
		options  *Options
	}{
		{
			name: "block_mapping",
			node: builder.BuildMapping([]*node.MappingPair{
				{
					Key:   builder.BuildScalar("key1", node.StylePlain),
					Value: builder.BuildScalar("value1", node.StylePlain),
				},
				{
					Key:   builder.BuildScalar("key2", node.StylePlain),
					Value: builder.BuildScalar("value2", node.StylePlain),
				},
				{
					Key:   builder.BuildScalar("key3", node.StylePlain),
					Value: builder.BuildScalar("value3", node.StylePlain),
				},
			}, node.StyleBlock),
			expected: "key1: value1\nkey2: value2\nkey3: value3",
		},
		{
			name: "flow_mapping",
			node: builder.BuildMapping([]*node.MappingPair{
				{
					Key:   builder.BuildScalar("a", node.StylePlain),
					Value: builder.BuildScalar("1", node.StylePlain),
				},
				{
					Key:   builder.BuildScalar("b", node.StylePlain),
					Value: builder.BuildScalar("2", node.StylePlain),
				},
			}, node.StyleFlow),
			expected: "{a: 1, b: 2}",
		},
		{
			name:     "empty_mapping",
			node:     builder.BuildMapping([]*node.MappingPair{}, node.StyleFlow),
			expected: "{}",
		},
		{
			name: "nested_mapping",
			node: builder.BuildMapping([]*node.MappingPair{
				{
					Key: builder.BuildScalar("outer", node.StylePlain),
					Value: builder.BuildMapping([]*node.MappingPair{
						{
							Key:   builder.BuildScalar("inner1", node.StylePlain),
							Value: builder.BuildScalar("value1", node.StylePlain),
						},
						{
							Key:   builder.BuildScalar("inner2", node.StylePlain),
							Value: builder.BuildScalar("value2", node.StylePlain),
						},
					}, node.StyleBlock),
				},
			}, node.StyleBlock),
			expected: "outer:\n  inner1: value1\n  inner2: value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.options
			if opts == nil {
				opts = DefaultOptions()
			}

			result, err := SerializeToString(tt.node, opts)
			if err != nil {
				t.Fatalf("Serialize error: %v", err)
			}

			// Trim trailing newline for comparison
			result = strings.TrimSuffix(result, "\n")

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestSerializeMixed(t *testing.T) {
	builder := &node.DefaultBuilder{}

	tests := []struct {
		name     string
		node     node.Node
		expected string
		options  *Options
	}{
		{
			name: "mapping_with_sequence_value",
			node: builder.BuildMapping([]*node.MappingPair{
				{
					Key: builder.BuildScalar("fruits", node.StylePlain),
					Value: builder.BuildSequence([]node.Node{
						builder.BuildScalar("apple", node.StylePlain),
						builder.BuildScalar("banana", node.StylePlain),
					}, node.StyleBlock),
				},
				{
					Key: builder.BuildScalar("vegetables", node.StylePlain),
					Value: builder.BuildSequence([]node.Node{
						builder.BuildScalar("carrot", node.StylePlain),
						builder.BuildScalar("potato", node.StylePlain),
					}, node.StyleBlock),
				},
			}, node.StyleBlock),
			expected: "fruits:\n  - apple\n  - banana\nvegetables:\n  - carrot\n  - potato",
		},
		{
			name: "sequence_with_mapping_items",
			node: builder.BuildSequence([]node.Node{
				builder.BuildMapping([]*node.MappingPair{
					{
						Key:   builder.BuildScalar("name", node.StylePlain),
						Value: builder.BuildScalar("Alice", node.StylePlain),
					},
					{
						Key:   builder.BuildScalar("age", node.StylePlain),
						Value: builder.BuildScalar("30", node.StylePlain),
					},
				}, node.StyleBlock),
				builder.BuildMapping([]*node.MappingPair{
					{
						Key:   builder.BuildScalar("name", node.StylePlain),
						Value: builder.BuildScalar("Bob", node.StylePlain),
					},
					{
						Key:   builder.BuildScalar("age", node.StylePlain),
						Value: builder.BuildScalar("25", node.StylePlain),
					},
				}, node.StyleBlock),
			}, node.StyleBlock),
			expected: "-\n  name: Alice\n  age: 30\n-\n  name: Bob\n  age: 25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.options
			if opts == nil {
				opts = DefaultOptions()
			}

			result, err := SerializeToString(tt.node, opts)
			if err != nil {
				t.Fatalf("Serialize error: %v", err)
			}

			// Trim trailing newline for comparison
			result = strings.TrimSuffix(result, "\n")

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestSerializeWithOptions(t *testing.T) {
	builder := &node.DefaultBuilder{}

	tests := []struct {
		name     string
		node     node.Node
		options  *Options
		expected string
	}{
		{
			name: "custom_indent",
			node: builder.BuildMapping([]*node.MappingPair{
				{
					Key: builder.BuildScalar("key", node.StylePlain),
					Value: builder.BuildSequence([]node.Node{
						builder.BuildScalar("item", node.StylePlain),
					}, node.StyleBlock),
				},
			}, node.StyleBlock),
			options: &Options{
				Indent:           4,
				PreserveComments: true,
			},
			expected: "key:\n    - item",
		},
		{
			name: "explicit_document_markers",
			node: builder.BuildScalar("content", node.StylePlain),
			options: &Options{
				Indent:                2,
				ExplicitDocumentStart: true,
				ExplicitDocumentEnd:   true,
			},
			expected: "---\ncontent\n...\n",
		},
		{
			name: "prefer_flow_style",
			node: builder.BuildSequence([]node.Node{
				builder.BuildScalar("a", node.StylePlain),
				builder.BuildScalar("b", node.StylePlain),
			}, node.StyleBlock),
			options: &Options{
				Indent:          2,
				PreferFlowStyle: true,
			},
			expected: "[a, b]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SerializeToString(tt.node, tt.options)
			if err != nil {
				t.Fatalf("Serialize error: %v", err)
			}

			// Trim trailing newline for comparison (except for document end marker)
			if !tt.options.ExplicitDocumentEnd {
				result = strings.TrimSuffix(result, "\n")
			}

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestSerializeWithComments(t *testing.T) {
	builder := &node.DefaultBuilder{}

	// Create a node with comments
	scalar := builder.BuildScalar("value", node.StylePlain)
	node.AssociateComment(scalar, "# This is a comment", node.CommentPositionAbove, 1)

	mapping := builder.BuildMapping([]*node.MappingPair{
		{
			Key:   builder.BuildScalar("key", node.StylePlain),
			Value: scalar,
		},
	}, node.StyleBlock)

	opts := &Options{
		Indent:                  2,
		PreserveComments:        true,
		BlankLinesBeforeComment: 1,
	}

	result, err := SerializeToString(mapping, opts)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	expected := "key:\n\n  # This is a comment\n  value"
	result = strings.TrimSuffix(result, "\n")

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can parse and serialize back to the same structure
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "simple_mapping",
			yaml: "key1: value1\nkey2: value2",
		},
		{
			name: "simple_sequence",
			yaml: "- item1\n- item2\n- item3",
		},
		{
			name: "nested_structure",
			yaml: "parent:\n  child1: value1\n  child2:\n    - item1\n    - item2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would require parser integration
			// For now, we'll skip the actual round-trip test
			t.Skip("Round-trip test requires parser integration")
		})
	}
}
