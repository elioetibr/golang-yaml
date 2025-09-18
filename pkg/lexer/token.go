package lexer

import "fmt"

// TokenType represents the type of a YAML token
type TokenType int

const (
	// Special tokens
	TokenEOF TokenType = iota
	TokenError

	// Structure tokens
	TokenDocumentStart     // ---
	TokenDocumentEnd       // ...
	TokenSequenceEntry     // -
	TokenMappingKey        // ? or implicit
	TokenMappingValue      // :
	TokenFlowSequenceStart // [
	TokenFlowSequenceEnd   // ]
	TokenFlowMappingStart  // {
	TokenFlowMappingEnd    // }
	TokenFlowEntry         // ,

	// Scalar tokens
	TokenScalar
	TokenPlainScalar
	TokenSingleQuotedScalar
	TokenDoubleQuotedScalar
	TokenLiteralScalar // |
	TokenFoldedScalar  // >

	// Special features
	TokenAnchor    // &
	TokenAlias     // *
	TokenTag       // !
	TokenDirective // %
	TokenComment   // #

	// Whitespace
	TokenNewLine
	TokenIndent
)

// Token represents a lexical token in YAML
type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
	Offset  int
	EndLine int
	EndCol  int
	Style   ScalarStyle
	Indent  int
	IsKey   bool
	IsValue bool

	// Comment tracking
	IsInline         bool // True if comment is on same line as content
	BlankLinesBefore int  // Number of blank lines before this token
	BlankLinesAfter  int  // Number of blank lines after this token
}

// ScalarStyle represents the style of a scalar
type ScalarStyle int

const (
	ScalarStylePlain ScalarStyle = iota
	ScalarStyleSingleQuoted
	ScalarStyleDoubleQuoted
	ScalarStyleLiteral
	ScalarStyleFolded
)

func (t TokenType) String() string {
	names := map[TokenType]string{
		TokenEOF:                "EOF",
		TokenError:              "ERROR",
		TokenDocumentStart:      "DOCUMENT_START",
		TokenDocumentEnd:        "DOCUMENT_END",
		TokenSequenceEntry:      "SEQUENCE_ENTRY",
		TokenMappingKey:         "MAPPING_KEY",
		TokenMappingValue:       "MAPPING_VALUE",
		TokenFlowSequenceStart:  "FLOW_SEQ_START",
		TokenFlowSequenceEnd:    "FLOW_SEQ_END",
		TokenFlowMappingStart:   "FLOW_MAP_START",
		TokenFlowMappingEnd:     "FLOW_MAP_END",
		TokenFlowEntry:          "FLOW_ENTRY",
		TokenScalar:             "SCALAR",
		TokenPlainScalar:        "PLAIN_SCALAR",
		TokenSingleQuotedScalar: "SINGLE_QUOTED",
		TokenDoubleQuotedScalar: "DOUBLE_QUOTED",
		TokenLiteralScalar:      "LITERAL",
		TokenFoldedScalar:       "FOLDED",
		TokenAnchor:             "ANCHOR",
		TokenAlias:              "ALIAS",
		TokenTag:                "TAG",
		TokenDirective:          "DIRECTIVE",
		TokenComment:            "COMMENT",
		TokenNewLine:            "NEWLINE",
		TokenIndent:             "INDENT",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}

func (t *Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Value: %q, Pos: %d:%d}",
		t.Type, t.Value, t.Line, t.Column)
}
