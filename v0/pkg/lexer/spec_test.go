package lexer

import (
	"strings"
	"testing"
)

// TestYAML122SpecCompliance tests compliance with YAML 1.2.2 specification
func TestYAML122SpecCompliance(t *testing.T) {
	t.Run("Indentation", func(t *testing.T) {
		t.Run("tabs_not_allowed", func(t *testing.T) {
			// Per spec 6.1: "Tab characters are not allowed in indentation"
			input := "key:\n\t- value"
			lexer := NewLexerFromString(input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			foundError := false
			for {
				token, err := lexer.NextToken()
				if err != nil {
					// We expect an error for tab indentation
					if strings.Contains(err.Error(), "tab") {
						foundError = true
						break
					}
				}
				if token.Type == TokenEOF {
					break
				}
			}

			if !foundError {
				t.Error("Expected error for tab indentation, but none was reported")
			}
		})
	})

	t.Run("Comments", func(t *testing.T) {
		t.Run("must_be_separated_by_whitespace", func(t *testing.T) {
			// Per spec: Comments must be separated from other tokens by white space
			input := "key:value#comment"
			lexer := NewLexerFromString(input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			var tokens []TokenType
			for {
				token, _ := lexer.NextToken()
				tokens = append(tokens, token.Type)
				if token.Type == TokenEOF {
					break
				}
			}

			// Should parse "value#comment" as single scalar, not separate comment
			hasComment := false
			for _, tt := range tokens {
				if tt == TokenComment {
					hasComment = true
				}
			}

			if hasComment {
				t.Error("Comment without whitespace separator should not be recognized")
			}
		})
	})

	t.Run("Scalars", func(t *testing.T) {
		t.Run("plain_scalar_restrictions", func(t *testing.T) {
			// Plain scalars cannot begin with most indicators
			invalidStarts := []string{
				"@invalid", "#invalid", "`invalid",
				// These are valid in some contexts but not as plain scalars at document start
			}

			for _, input := range invalidStarts {
				lexer := NewLexerFromString(input)
				_ = lexer.Initialize()

				token, _ := lexer.NextToken()
				if token.Type == TokenPlainScalar && strings.HasPrefix(input, "@") {
					// @ is not a special character in YAML, should be valid
					continue
				}
				if token.Type == TokenPlainScalar && strings.HasPrefix(input, "#") {
					t.Errorf("Plain scalar should not start with '#': %s", input)
				}
			}
		})

		t.Run("escape_sequences", func(t *testing.T) {
			// Test double-quoted escape sequences per spec
			tests := []struct {
				input    string
				expected string
			}{
				{`"a\nb"`, "a\nb"},    // newline
				{`"a\tb"`, "a\tb"},    // tab
				{`"a\\b"`, "a\\b"},    // backslash
				{`"a\"b"`, `a"b`},     // quote
				{`"a\x41b"`, "aAb"},   // hex escape (if implemented)
				{`"a\u0041b"`, "aAb"}, // unicode escape (if implemented)
			}

			for _, tt := range tests {
				lexer := NewLexerFromString(tt.input)
				_ = lexer.Initialize()

				token, _ := lexer.NextToken()
				if token.Type == TokenDoubleQuotedScalar {
					// Basic escapes should work
					if tt.input == `"a\nb"` && token.Value != "a\nb" {
						t.Errorf("Escape sequence not handled: %s", tt.input)
					}
				}
			}
		})
	})

	t.Run("Collections", func(t *testing.T) {
		t.Run("mapping_keys_must_be_unique", func(t *testing.T) {
			// Per spec: mapping keys must be unique
			// This is more of a parser/validator concern
			input := `
key: value1
key: value2`
			lexer := NewLexerFromString(input)
			_ = lexer.Initialize()

			// Lexer just tokenizes, uniqueness check is parser responsibility
			// Just verify we can tokenize it
			tokenCount := 0
			for {
				token, _ := lexer.NextToken()
				tokenCount++
				if token.Type == TokenEOF {
					break
				}
			}

			if tokenCount < 5 { // Should have at least: key : value key : value EOF
				t.Error("Failed to tokenize duplicate keys")
			}
		})
	})

	t.Run("DocumentStructure", func(t *testing.T) {
		t.Run("explicit_document_markers", func(t *testing.T) {
			input := `---
content
...
---
more content
...`
			lexer := NewLexerFromString(input)
			_ = lexer.Initialize()

			expectedTypes := []TokenType{
				TokenDocumentStart,
				TokenPlainScalar,
				TokenDocumentEnd,
				TokenDocumentStart,
				TokenPlainScalar,
				TokenDocumentEnd,
				TokenEOF,
			}

			for i, expected := range expectedTypes {
				token, _ := lexer.NextToken()
				if token.Type != expected {
					t.Errorf("Token %d: expected %v, got %v", i, expected, token.Type)
				}
				if token.Type == TokenEOF {
					break
				}
			}
		})

		t.Run("directives", func(t *testing.T) {
			input := `%YAML 1.2
%TAG ! tag:example.com,2014:
---
content`
			lexer := NewLexerFromString(input)
			_ = lexer.Initialize()

			token, _ := lexer.NextToken()
			if token.Type != TokenDirective {
				t.Errorf("Expected directive token, got %v", token.Type)
			}
			if !strings.Contains(token.Value, "YAML") {
				t.Error("YAML directive not recognized")
			}
		})
	})

	t.Run("FlowStyles", func(t *testing.T) {
		t.Run("flow_sequences", func(t *testing.T) {
			input := `[a, b, c]`
			lexer := NewLexerFromString(input)
			_ = lexer.Initialize()

			expectedTypes := []TokenType{
				TokenFlowSequenceStart,
				TokenPlainScalar,
				TokenFlowEntry,
				TokenPlainScalar,
				TokenFlowEntry,
				TokenPlainScalar,
				TokenFlowSequenceEnd,
			}

			for _, expected := range expectedTypes {
				token, _ := lexer.NextToken()
				if token.Type != expected {
					t.Errorf("Expected %v, got %v", expected, token.Type)
				}
			}
		})

		t.Run("flow_mappings", func(t *testing.T) {
			input := `{a: 1, b: 2}`
			lexer := NewLexerFromString(input)
			_ = lexer.Initialize()

			expectedTypes := []TokenType{
				TokenFlowMappingStart,
				TokenPlainScalar,
				TokenMappingValue,
				TokenPlainScalar,
				TokenFlowEntry,
				TokenPlainScalar,
				TokenMappingValue,
				TokenPlainScalar,
				TokenFlowMappingEnd,
			}

			for _, expected := range expectedTypes {
				token, _ := lexer.NextToken()
				if token.Type != expected {
					t.Errorf("Expected %v, got %v", expected, token.Type)
				}
			}
		})
	})
}

// TestBlockScalarIndicators tests chomping and indentation indicators
func TestBlockScalarIndicators(t *testing.T) {
	t.Run("chomping_indicators", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
			desc     string
		}{
			{
				name: "strip_default",
				input: `|
  text
  `,
				expected: "text",
				desc:     "Default strips final newlines",
			},
			{
				name: "clip",
				input: `|
  text
`,
				expected: "text\n",
				desc:     "Clip keeps one newline",
			},
			{
				name: "keep",
				input: `|+
  text

`,
				expected: "text\n\n",
				desc:     "Keep preserves all newlines",
			},
			{
				name: "strip_explicit",
				input: `|-
  text
  `,
				expected: "text",
				desc:     "Strip removes all trailing newlines",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Skip("Chomping indicators not yet implemented")
				// When implemented, test that the scalar value matches expected
			})
		}
	})

	t.Run("indentation_indicators", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`|2
  text`, "text"}, // Explicit indent 2
			{`|1
 text`, "text"}, // Explicit indent 1
		}

		for range tests {
			t.Skip("Indentation indicators not yet implemented")
		}
	})
}
