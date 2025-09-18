package lexer

import (
	"strings"
	"testing"
)

// TestBasicTokens tests basic YAML tokens
func TestBasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "document markers",
			input: "---\n...",
			expected: []TokenType{
				TokenDocumentStart,
				TokenDocumentEnd,
				TokenEOF,
			},
		},
		{
			name:  "simple key-value",
			input: "key: value",
			expected: []TokenType{
				TokenPlainScalar,
				TokenMappingValue,
				TokenPlainScalar,
				TokenEOF,
			},
		},
		{
			name:  "sequence entry",
			input: "- item",
			expected: []TokenType{
				TokenSequenceEntry,
				TokenPlainScalar,
				TokenEOF,
			},
		},
		{
			name:  "flow sequence",
			input: "[a, b, c]",
			expected: []TokenType{
				TokenFlowSequenceStart,
				TokenPlainScalar,
				TokenFlowEntry,
				TokenPlainScalar,
				TokenFlowEntry,
				TokenPlainScalar,
				TokenFlowSequenceEnd,
				TokenEOF,
			},
		},
		{
			name:  "flow mapping",
			input: "{key: value}",
			expected: []TokenType{
				TokenFlowMappingStart,
				TokenPlainScalar,
				TokenMappingValue,
				TokenPlainScalar,
				TokenFlowMappingEnd,
				TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexerFromString(tt.input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			var tokens []TokenType
			for {
				token, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				tokens = append(tokens, token.Type)
				if token.Type == TokenEOF {
					break
				}
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", tokens)
				return
			}

			for i, expectedType := range tt.expected {
				if tokens[i] != expectedType {
					t.Errorf("Token %d: expected %v, got %v", i, expectedType, tokens[i])
				}
			}
		})
	}
}

// TestScalarStyles tests different scalar styles
func TestScalarStyles(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  TokenType
		expectedValue string
		expectedStyle ScalarStyle
	}{
		{
			name:          "plain scalar",
			input:         "plain text",
			expectedType:  TokenPlainScalar,
			expectedValue: "plain text",
			expectedStyle: ScalarStylePlain,
		},
		{
			name:          "single quoted",
			input:         "'single quoted'",
			expectedType:  TokenSingleQuotedScalar,
			expectedValue: "single quoted",
			expectedStyle: ScalarStyleSingleQuoted,
		},
		{
			name:          "double quoted",
			input:         `"double quoted"`,
			expectedType:  TokenDoubleQuotedScalar,
			expectedValue: "double quoted",
			expectedStyle: ScalarStyleDoubleQuoted,
		},
		{
			name:          "double quoted with escape",
			input:         `"line1\nline2"`,
			expectedType:  TokenDoubleQuotedScalar,
			expectedValue: "line1\nline2",
			expectedStyle: ScalarStyleDoubleQuoted,
		},
		{
			name:          "single quoted with double quote",
			input:         `'it''s'`,
			expectedType:  TokenSingleQuotedScalar,
			expectedValue: "it's",
			expectedStyle: ScalarStyleSingleQuoted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexerFromString(tt.input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if token.Type != tt.expectedType {
				t.Errorf("Expected token type %v, got %v", tt.expectedType, token.Type)
			}

			if token.Value != tt.expectedValue {
				t.Errorf("Expected value %q, got %q", tt.expectedValue, token.Value)
			}

			if token.Style != tt.expectedStyle {
				t.Errorf("Expected style %v, got %v", tt.expectedStyle, token.Style)
			}
		})
	}
}

// TestComments tests comment handling
func TestComments(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedComments   []string
		expectedBlankLines []int
		expectedInline     []bool
	}{
		{
			name: "standalone comment",
			input: `# This is a comment
key: value`,
			expectedComments:   []string{"# This is a comment"},
			expectedBlankLines: []int{0},
			expectedInline:     []bool{false},
		},
		{
			name:               "inline comment",
			input:              `key: value  # inline comment`,
			expectedComments:   []string{"# inline comment"},
			expectedBlankLines: []int{0},
			expectedInline:     []bool{true},
		},
		{
			name: "comment with blank lines",
			input: `

# Comment after blank lines
key: value`,
			expectedComments:   []string{"# Comment after blank lines"},
			expectedBlankLines: []int{2},
			expectedInline:     []bool{false},
		},
		{
			name: "multiple comments",
			input: `# First comment
# Second comment

# Third comment after blank
key: value  # inline`,
			expectedComments: []string{
				"# First comment",
				"# Second comment",
				"# Third comment after blank",
				"# inline",
			},
			expectedBlankLines: []int{0, 0, 1, 0},
			expectedInline:     []bool{false, false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexerFromString(tt.input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			var comments []string
			var blankLines []int
			var inline []bool

			for {
				token, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if token.Type == TokenComment {
					comments = append(comments, token.Value)
					blankLines = append(blankLines, token.BlankLinesBefore)
					inline = append(inline, token.IsInline)
				}

				if token.Type == TokenEOF {
					break
				}
			}

			if len(comments) != len(tt.expectedComments) {
				t.Errorf("Expected %d comments, got %d", len(tt.expectedComments), len(comments))
				return
			}

			for i := range comments {
				if comments[i] != tt.expectedComments[i] {
					t.Errorf("Comment %d: expected %q, got %q", i, tt.expectedComments[i], comments[i])
				}
				if blankLines[i] != tt.expectedBlankLines[i] {
					t.Errorf("Comment %d blank lines: expected %d, got %d", i, tt.expectedBlankLines[i], blankLines[i])
				}
				if inline[i] != tt.expectedInline[i] {
					t.Errorf("Comment %d inline: expected %v, got %v", i, tt.expectedInline[i], inline[i])
				}
			}
		})
	}
}

// TestSpecialTokens tests anchors, aliases, tags, and directives
func TestSpecialTokens(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  TokenType
		expectedValue string
	}{
		{
			name:          "anchor",
			input:         "&anchor",
			expectedType:  TokenAnchor,
			expectedValue: "anchor",
		},
		{
			name:          "alias",
			input:         "*anchor",
			expectedType:  TokenAlias,
			expectedValue: "anchor",
		},
		{
			name:          "tag",
			input:         "!tag",
			expectedType:  TokenTag,
			expectedValue: "!tag",
		},
		{
			name:          "directive",
			input:         "%YAML 1.2",
			expectedType:  TokenDirective,
			expectedValue: "%YAML 1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexerFromString(tt.input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if token.Type != tt.expectedType {
				t.Errorf("Expected token type %v, got %v", tt.expectedType, token.Type)
			}

			if token.Value != tt.expectedValue {
				t.Errorf("Expected value %q, got %q", tt.expectedValue, token.Value)
			}
		})
	}
}

// TestBlockScalars tests literal and folded scalars
func TestBlockScalars(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  TokenType
		expectedValue string
		expectedStyle ScalarStyle
	}{
		{
			name: "literal scalar",
			input: `|
  line1
  line2`,
			expectedType:  TokenLiteralScalar,
			expectedValue: "  line1\n  line2",
			expectedStyle: ScalarStyleLiteral,
		},
		{
			name: "folded scalar",
			input: `>
  line1
  line2`,
			expectedType:  TokenFoldedScalar,
			expectedValue: "  line1\n  line2",
			expectedStyle: ScalarStyleFolded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexerFromString(tt.input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if token.Type != tt.expectedType {
				t.Errorf("Expected token type %v, got %v", tt.expectedType, token.Type)
			}

			if token.Value != tt.expectedValue {
				t.Errorf("Expected value %q, got %q", tt.expectedValue, token.Value)
			}

			if token.Style != tt.expectedStyle {
				t.Errorf("Expected style %v, got %v", tt.expectedStyle, token.Style)
			}
		})
	}
}

// TestComplexDocument tests a complete YAML document
func TestComplexDocument(t *testing.T) {
	input := `---
# Configuration file
name: test-app
version: 1.0.0

# Server settings
server:
  host: localhost
  port: 8080

# Features list
features:
  - authentication  # Core feature
  - logging
  - caching

# Debug options
debug: true
...`

	expectedTokenTypes := []TokenType{
		TokenDocumentStart, // ---
		TokenComment,       // # Configuration file
		TokenPlainScalar,   // name
		TokenMappingValue,  // :
		TokenPlainScalar,   // test-app
		TokenPlainScalar,   // version
		TokenMappingValue,  // :
		TokenPlainScalar,   // 1.0.0
		TokenComment,       // # Server settings
		TokenPlainScalar,   // server
		TokenMappingValue,  // :
		TokenPlainScalar,   // host
		TokenMappingValue,  // :
		TokenPlainScalar,   // localhost
		TokenPlainScalar,   // port
		TokenMappingValue,  // :
		TokenPlainScalar,   // 8080
		TokenComment,       // # Features list
		TokenPlainScalar,   // features
		TokenMappingValue,  // :
		TokenSequenceEntry, // -
		TokenPlainScalar,   // authentication
		TokenComment,       // # Core feature (inline comment)
		TokenSequenceEntry, // -
		TokenPlainScalar,   // logging
		TokenSequenceEntry, // -
		TokenPlainScalar,   // caching
		TokenComment,       // # Debug options
		TokenPlainScalar,   // debug
		TokenMappingValue,  // :
		TokenPlainScalar,   // true
		TokenDocumentEnd,   // ...
		TokenEOF,
	}

	lexer := NewLexerFromString(input)
	err := lexer.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize lexer: %v", err)
	}

	var tokens []*Token
	for {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		tokens = append(tokens, token)
		if token.Type == TokenEOF {
			break
		}
	}

	if len(tokens) != len(expectedTokenTypes) {
		t.Errorf("Expected %d tokens, got %d", len(expectedTokenTypes), len(tokens))
		t.Log("Expected tokens:")
		for i, tt := range expectedTokenTypes {
			t.Logf("  %d: %v", i, tt)
		}
		t.Log("Got tokens:")
		for i, token := range tokens {
			t.Logf("  %d: %v (%q)", i, token.Type, token.Value)
		}
		return
	}

	for i, expectedType := range expectedTokenTypes {
		if tokens[i].Type != expectedType {
			t.Errorf("Token %d: expected %v, got %v (%q)",
				i, expectedType, tokens[i].Type, tokens[i].Value)
		}
	}
}

// TestPositionTracking tests line and column tracking
func TestPositionTracking(t *testing.T) {
	input := `key: value
second: line
  nested: item`

	expectedPositions := []struct {
		tokenType TokenType
		line      int
		column    int
		value     string
	}{
		{TokenPlainScalar, 1, 1, "key"},
		{TokenMappingValue, 1, 4, ":"},
		{TokenPlainScalar, 1, 6, "value"},
		{TokenPlainScalar, 2, 1, "second"},
		{TokenMappingValue, 2, 7, ":"},
		{TokenPlainScalar, 2, 9, "line"},
		{TokenPlainScalar, 3, 3, "nested"},
		{TokenMappingValue, 3, 9, ":"},
		{TokenPlainScalar, 3, 11, "item"},
		{TokenEOF, 4, 1, ""},
	}

	lexer := NewLexerFromString(input)
	err := lexer.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize lexer: %v", err)
	}

	for i, expected := range expectedPositions {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("Unexpected error at token %d: %v", i, err)
		}

		if token.Type != expected.tokenType {
			t.Errorf("Token %d type: expected %v, got %v", i, expected.tokenType, token.Type)
		}

		if token.Line != expected.line {
			t.Errorf("Token %d line: expected %d, got %d (value: %q)",
				i, expected.line, token.Line, token.Value)
		}

		if token.Column != expected.column {
			t.Errorf("Token %d column: expected %d, got %d (value: %q)",
				i, expected.column, token.Column, token.Value)
		}

		if token.Value != expected.value {
			t.Errorf("Token %d value: expected %q, got %q", i, expected.value, token.Value)
		}
	}
}

// TestErrorCases tests error handling
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "empty input",
			input:       "",
			shouldError: false, // Empty input should just return EOF
		},
		{
			name:        "only whitespace",
			input:       "   \n\t\n   ",
			shouldError: false, // Should return EOF
		},
		{
			name:        "only comments",
			input:       "# comment 1\n# comment 2",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexerFromString(tt.input)
			err := lexer.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize lexer: %v", err)
			}

			var hasError bool
			for {
				token, err := lexer.NextToken()
				if err != nil {
					hasError = true
					if !tt.shouldError {
						t.Errorf("Unexpected error: %v", err)
					} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
					}
					break
				}
				if token.Type == TokenEOF {
					break
				}
			}

			if tt.shouldError && !hasError {
				t.Error("Expected an error but got none")
			}
		})
	}
}

// BenchmarkLexer benchmarks lexer performance
func BenchmarkLexer(b *testing.B) {
	input := `---
name: benchmark-test
version: 1.0.0
features:
  - feature1
  - feature2
  - feature3
config:
  debug: true
  timeout: 30
  retries: 3
...`

	for i := 0; i < b.N; i++ {
		lexer := NewLexerFromString(input)
		_ = lexer.Initialize()

		for {
			token, _ := lexer.NextToken()
			if token.Type == TokenEOF {
				break
			}
		}
	}
}

// TestIndentationTracking tests indentation level tracking
func TestIndentationTracking(t *testing.T) {
	input := `root:
  level1:
    level2: value
  another1:
    another2:
      level3: deep`

	lexer := NewLexerFromString(input)
	err := lexer.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize lexer: %v", err)
	}

	// Track indentation changes
	var lastIndent int
	for {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// For scalar tokens, check if indentation makes sense
		if token.Type == TokenPlainScalar && token.Column != lastIndent {
			// Indentation should only increase by consistent amounts
			if token.Column > lastIndent && (token.Column-lastIndent)%2 != 0 {
				t.Logf("Warning: Inconsistent indentation at line %d, column %d",
					token.Line, token.Column)
			}
			lastIndent = token.Column
		}

		if token.Type == TokenEOF {
			break
		}
	}
}
