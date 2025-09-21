package serializer

import (
	"strings"
	"testing"

	node2 "github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestSerializeScalar(t *testing.T) {
	tests := []struct {
		name     string
		node     *node2.ScalarNode
		expected string
		options  *Options
	}{
		{
			name: "plain_scalar",
			node: &node2.ScalarNode{
				Value: "hello world",
				Style: node2.StylePlain,
			},
			expected: "hello world",
		},
		{
			name: "single_quoted_scalar",
			node: &node2.ScalarNode{
				Value: "hello world",
				Style: node2.StyleSingleQuoted,
			},
			expected: "'hello world'",
		},
		{
			name: "double_quoted_scalar",
			node: &node2.ScalarNode{
				Value: "hello\nworld",
				Style: node2.StyleDoubleQuoted,
			},
			expected: `"hello\nworld"`,
		},
		{
			name: "literal_scalar",
			node: &node2.ScalarNode{
				Value: "hello\nworld",
				Style: node2.StyleLiteral,
			},
			expected: "|\n  hello\n  world\n",
		},
		{
			name: "folded_scalar",
			node: &node2.ScalarNode{
				Value: "hello\nworld",
				Style: node2.StyleFolded,
			},
			expected: ">\n  hello\n  world\n",
		},
		{
			name: "plain_boolean_literal",
			node: &node2.ScalarNode{
				Value: "true",
				Style: node2.StylePlain,
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
	builder := &node2.DefaultBuilder{}

	tests := []struct {
		name     string
		node     *node2.SequenceNode
		expected string
		options  *Options
	}{
		{
			name: "block_sequence",
			node: builder.BuildSequence([]node2.Node{
				builder.BuildScalar("item1", node2.StylePlain),
				builder.BuildScalar("item2", node2.StylePlain),
				builder.BuildScalar("item3", node2.StylePlain),
			}, node2.StyleBlock),
			expected: "- item1\n- item2\n- item3",
		},
		{
			name: "flow_sequence",
			node: builder.BuildSequence([]node2.Node{
				builder.BuildScalar("a", node2.StylePlain),
				builder.BuildScalar("b", node2.StylePlain),
				builder.BuildScalar("c", node2.StylePlain),
			}, node2.StyleFlow),
			expected: "[a, b, c]",
		},
		{
			name:     "empty_sequence",
			node:     builder.BuildSequence([]node2.Node{}, node2.StyleFlow),
			expected: "[]",
		},
		{
			name: "nested_sequence",
			node: builder.BuildSequence([]node2.Node{
				builder.BuildScalar("item1", node2.StylePlain),
				builder.BuildSequence([]node2.Node{
					builder.BuildScalar("nested1", node2.StylePlain),
					builder.BuildScalar("nested2", node2.StylePlain),
				}, node2.StyleBlock),
			}, node2.StyleBlock),
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
	builder := &node2.DefaultBuilder{}

	tests := []struct {
		name     string
		node     *node2.MappingNode
		expected string
		options  *Options
	}{
		{
			name: "block_mapping",
			node: builder.BuildMapping([]*node2.MappingPair{
				{
					Key:   builder.BuildScalar("key1", node2.StylePlain),
					Value: builder.BuildScalar("value1", node2.StylePlain),
				},
				{
					Key:   builder.BuildScalar("key2", node2.StylePlain),
					Value: builder.BuildScalar("value2", node2.StylePlain),
				},
				{
					Key:   builder.BuildScalar("key3", node2.StylePlain),
					Value: builder.BuildScalar("value3", node2.StylePlain),
				},
			}, node2.StyleBlock),
			expected: "key1: value1\nkey2: value2\nkey3: value3",
		},
		{
			name: "flow_mapping",
			node: builder.BuildMapping([]*node2.MappingPair{
				{
					Key:   builder.BuildScalar("a", node2.StylePlain),
					Value: builder.BuildScalar("1", node2.StylePlain),
				},
				{
					Key:   builder.BuildScalar("b", node2.StylePlain),
					Value: builder.BuildScalar("2", node2.StylePlain),
				},
			}, node2.StyleFlow),
			expected: "{a: 1, b: 2}",
		},
		{
			name:     "empty_mapping",
			node:     builder.BuildMapping([]*node2.MappingPair{}, node2.StyleFlow),
			expected: "{}",
		},
		{
			name: "nested_mapping",
			node: builder.BuildMapping([]*node2.MappingPair{
				{
					Key: builder.BuildScalar("outer", node2.StylePlain),
					Value: builder.BuildMapping([]*node2.MappingPair{
						{
							Key:   builder.BuildScalar("inner1", node2.StylePlain),
							Value: builder.BuildScalar("value1", node2.StylePlain),
						},
						{
							Key:   builder.BuildScalar("inner2", node2.StylePlain),
							Value: builder.BuildScalar("value2", node2.StylePlain),
						},
					}, node2.StyleBlock),
				},
			}, node2.StyleBlock),
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
	builder := &node2.DefaultBuilder{}

	tests := []struct {
		name     string
		node     node2.Node
		expected string
		options  *Options
	}{
		{
			name: "mapping_with_sequence_value",
			node: builder.BuildMapping([]*node2.MappingPair{
				{
					Key: builder.BuildScalar("fruits", node2.StylePlain),
					Value: builder.BuildSequence([]node2.Node{
						builder.BuildScalar("apple", node2.StylePlain),
						builder.BuildScalar("banana", node2.StylePlain),
					}, node2.StyleBlock),
				},
				{
					Key: builder.BuildScalar("vegetables", node2.StylePlain),
					Value: builder.BuildSequence([]node2.Node{
						builder.BuildScalar("carrot", node2.StylePlain),
						builder.BuildScalar("potato", node2.StylePlain),
					}, node2.StyleBlock),
				},
			}, node2.StyleBlock),
			expected: "fruits:\n  - apple\n  - banana\nvegetables:\n  - carrot\n  - potato",
		},
		{
			name: "sequence_with_mapping_items",
			node: builder.BuildSequence([]node2.Node{
				builder.BuildMapping([]*node2.MappingPair{
					{
						Key:   builder.BuildScalar("name", node2.StylePlain),
						Value: builder.BuildScalar("Alice", node2.StylePlain),
					},
					{
						Key:   builder.BuildScalar("age", node2.StylePlain),
						Value: builder.BuildScalar("30", node2.StylePlain),
					},
				}, node2.StyleBlock),
				builder.BuildMapping([]*node2.MappingPair{
					{
						Key:   builder.BuildScalar("name", node2.StylePlain),
						Value: builder.BuildScalar("Bob", node2.StylePlain),
					},
					{
						Key:   builder.BuildScalar("age", node2.StylePlain),
						Value: builder.BuildScalar("25", node2.StylePlain),
					},
				}, node2.StyleBlock),
			}, node2.StyleBlock),
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
	builder := &node2.DefaultBuilder{}

	tests := []struct {
		name     string
		node     node2.Node
		options  *Options
		expected string
	}{
		{
			name: "custom_indent",
			node: builder.BuildMapping([]*node2.MappingPair{
				{
					Key: builder.BuildScalar("key", node2.StylePlain),
					Value: builder.BuildSequence([]node2.Node{
						builder.BuildScalar("item", node2.StylePlain),
					}, node2.StyleBlock),
				},
			}, node2.StyleBlock),
			options: &Options{
				Indent:           4,
				PreserveComments: true,
			},
			expected: "key:\n    - item",
		},
		{
			name: "explicit_document_markers",
			node: builder.BuildScalar("content", node2.StylePlain),
			options: &Options{
				Indent:                2,
				ExplicitDocumentStart: true,
				ExplicitDocumentEnd:   true,
			},
			expected: "---\ncontent\n...\n",
		},
		{
			name: "prefer_flow_style",
			node: builder.BuildSequence([]node2.Node{
				builder.BuildScalar("a", node2.StylePlain),
				builder.BuildScalar("b", node2.StylePlain),
			}, node2.StyleBlock),
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
	builder := &node2.DefaultBuilder{}

	// Create a node with comments
	scalar := builder.BuildScalar("value", node2.StylePlain)
	node2.AssociateComment(scalar, "# This is a comment", node2.CommentPositionAbove, 1)

	mapping := builder.BuildMapping([]*node2.MappingPair{
		{
			Key:   builder.BuildScalar("key", node2.StylePlain),
			Value: scalar,
		},
	}, node2.StyleBlock)

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
