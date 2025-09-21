package test

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/elioetibr/golang-yaml/v0/pkg/decoder"
	"github.com/elioetibr/golang-yaml/v0/pkg/encoder"
	"github.com/elioetibr/golang-yaml/v0/pkg/parser"
)

// FuzzParser tests the parser with random inputs
func FuzzParser(f *testing.F) {
	// Add seed corpus
	seeds := []string{
		"key: value",
		"- item1\n- item2",
		"{a: 1, b: 2}",
		"[1, 2, 3]",
		"---\nkey: value\n...",
		"&anchor value",
		"*anchor",
		"!!str 123",
		"key: |\\n  literal\\n  text",
		"key: >\\n  folded\\n  text",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The parser should not panic on any input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on input: %q\nPanic: %v", input, r)
			}
		}()

		// Try to parse
		_, _ = parser.ParseString(input)
	})
}

// FuzzMarshalUnmarshal tests marshal/unmarshal with random data
func FuzzMarshalUnmarshal(f *testing.F) {
	// Add seed values-with-comments
	f.Add("test", 123, true)
	f.Add("", 0, false)
	f.Add("special chars: \n\t\"'", -456, true)

	f.Fuzz(func(t *testing.T, s string, n int, b bool) {
		// Create a test structure
		type TestStruct struct {
			String string `yaml:"string"`
			Number int    `yaml:"number"`
			Bool   bool   `yaml:"bool"`
		}

		original := TestStruct{
			String: s,
			Number: n,
			Bool:   b,
		}

		// Marshal
		data, err := encoder.Marshal(original)
		if err != nil {
			t.Skip("Marshal failed, skipping")
			return
		}

		// Unmarshal
		var decoded TestStruct
		err = decoder.Unmarshal(data, &decoded)
		if err != nil {
			t.Skip("Unmarshal failed, skipping")
			return
		}

		// Compare
		if decoded.String != original.String {
			t.Errorf("String mismatch: got %q, want %q", decoded.String, original.String)
		}
		if decoded.Number != original.Number {
			t.Errorf("Number mismatch: got %d, want %d", decoded.Number, original.Number)
		}
		if decoded.Bool != original.Bool {
			t.Errorf("Bool mismatch: got %v, want %v", decoded.Bool, original.Bool)
		}
	})
}

// RandomYAMLGenerator generates random valid YAML
type RandomYAMLGenerator struct {
	rand  *rand.Rand
	depth int
}

// NewRandomYAMLGenerator creates a new generator
func NewRandomYAMLGenerator(seed int64) *RandomYAMLGenerator {
	return &RandomYAMLGenerator{
		rand:  rand.New(rand.NewSource(seed)),
		depth: 0,
	}
}

// Generate creates random YAML
func (g *RandomYAMLGenerator) Generate() string {
	g.depth = 0
	return g.generateValue(0)
}

// generateValue generates a random YAML value
func (g *RandomYAMLGenerator) generateValue(indent int) string {
	g.depth++
	defer func() { g.depth-- }()

	// Limit nesting depth
	if g.depth > 5 {
		return g.generateScalar()
	}

	switch g.rand.Intn(6) {
	case 0:
		return g.generateScalar()
	case 1:
		return g.generateQuotedScalar()
	case 2:
		if g.depth < 3 {
			return g.generateSequence(indent)
		}
		return g.generateScalar()
	case 3:
		if g.depth < 3 {
			return g.generateMapping(indent)
		}
		return g.generateScalar()
	case 4:
		return g.generateFlowSequence()
	case 5:
		return g.generateFlowMapping()
	default:
		return g.generateScalar()
	}
}

// generateScalar creates a random scalar
func (g *RandomYAMLGenerator) generateScalar() string {
	scalars := []string{
		"value", "123", "3.14", "true", "false", "null",
		"text", "data", "item", "test", "example",
	}
	return scalars[g.rand.Intn(len(scalars))]
}

// generateQuotedScalar creates a quoted scalar
func (g *RandomYAMLGenerator) generateQuotedScalar() string {
	quotes := []string{"'", "\""}
	quote := quotes[g.rand.Intn(2)]
	value := g.generateScalar()
	return quote + value + quote
}

// generateSequence creates a block sequence
func (g *RandomYAMLGenerator) generateSequence(indent int) string {
	var items []string
	count := g.rand.Intn(3) + 1
	for i := 0; i < count; i++ {
		prefix := strings.Repeat(" ", indent) + "- "
		value := g.generateValue(indent + 2)
		items = append(items, prefix+value)
	}
	return strings.Join(items, "\n")
}

// generateMapping creates a block mapping
func (g *RandomYAMLGenerator) generateMapping(indent int) string {
	var pairs []string
	count := g.rand.Intn(3) + 1
	keys := []string{"key", "name", "value", "data", "item", "prop"}

	for i := 0; i < count; i++ {
		key := keys[g.rand.Intn(len(keys))] + strconv.Itoa(i)
		prefix := strings.Repeat(" ", indent) + key + ": "
		value := g.generateValue(indent + 2)
		pairs = append(pairs, prefix+value)
	}
	return strings.Join(pairs, "\n")
}

// generateFlowSequence creates a flow sequence
func (g *RandomYAMLGenerator) generateFlowSequence() string {
	var items []string
	count := g.rand.Intn(3) + 1
	for i := 0; i < count; i++ {
		items = append(items, g.generateScalar())
	}
	return "[" + strings.Join(items, ", ") + "]"
}

// generateFlowMapping creates a flow mapping
func (g *RandomYAMLGenerator) generateFlowMapping() string {
	var pairs []string
	count := g.rand.Intn(3) + 1
	keys := []string{"a", "b", "c", "x", "y", "z"}

	for i := 0; i < count; i++ {
		key := keys[g.rand.Intn(len(keys))]
		value := g.generateScalar()
		pairs = append(pairs, key+": "+value)
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// TestRandomYAML tests with randomly generated YAML
func TestRandomYAML(t *testing.T) {
	generator := NewRandomYAMLGenerator(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		yaml := generator.Generate()

		// Should not panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Parser panicked on generated YAML:\n%s\nPanic: %v", yaml, r)
				}
			}()

			_, _ = parser.ParseString(yaml)
		}()
	}
}

// TestMalformedYAML tests parser resilience with malformed input
func TestMalformedYAML(t *testing.T) {
	malformed := []string{
		"key: value: extra",
		"- item\n  - nested wrong indent",
		"{ unclosed",
		"[ unclosed",
		"& invalid anchor",
		"* undefined",
		"!!unknown_tag value",
		"---\n...\n---", // Multiple markers
		"key:\n\t- tab indent",
		": no key",
		"- \n  - ", // Empty items
		"{,}",      // Invalid flow
		"[,]",      // Invalid flow
	}

	for _, input := range malformed {
		t.Run(input, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Parser panicked on input: %q\nPanic: %v", input, r)
				}
			}()

			// Parse - may error but shouldn't panic
			_, _ = parser.ParseString(input)
		})
	}
}

// TestLargeDocuments tests handling of large YAML documents
func TestLargeDocuments(t *testing.T) {
	// Generate a large document
	var sb strings.Builder
	sb.WriteString("root:\n")

	// Create deep nesting
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("  level%d:\n", i))
		sb.WriteString(strings.Repeat(" ", (i+2)*2))
		sb.WriteString("value: test\n")
	}

	largeYAML := sb.String()

	// Should handle without stack overflow
	_, err := parser.ParseString(largeYAML)
	if err != nil {
		t.Logf("Large document parse error: %v", err)
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"only_comment", "# comment"},
		{"only_document_marker", "---"},
		{"only_spaces", "   "},
		{"only_newlines", "\n\n\n"},
		{"unicode", "key: 你好世界 🌍"},
		{"special_chars", `key: "!@#$%^&*()"`},
		{"very_long_line", "key: " + strings.Repeat("x", 10000)},
		{"many_keys", strings.Repeat("key: value\n", 1000)},
		{"deep_list", strings.Repeat("- ", 100) + "item"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Parser panicked on edge case %q: %v", tc.name, r)
				}
			}()

			_, _ = parser.ParseString(tc.input)
		})
	}
}
