package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elioetibr/golang-yaml/v1/pkg/errors"
)

// Lexer tokenizes YAML input
type Lexer struct {
	reader      *bufio.Reader
	input       string
	pos         int
	line        int
	column      int
	indentStack []int
	tokens      []Token
	current     rune
	next        rune
	inFlow      int
	errors      []*errors.YAMLError

	// Blank line and comment tracking
	blankLineCount   int
	lastNonBlankLine int
	pendingComments  []Token
	lastTokenLine    int
}

// NewLexer creates a new lexer from a reader
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader:      bufio.NewReader(r),
		line:        1,
		column:      1,
		indentStack: []int{0},
		tokens:      make([]Token, 0),
	}
}

// NewLexerFromString creates a new lexer from a string
func NewLexerFromString(input string) *Lexer {
	l := &Lexer{
		input:       input,
		line:        1,
		column:      1,
		indentStack: []int{0},
		tokens:      make([]Token, 0),
	}
	// Initialize current to the first character
	if len(input) > 0 {
		l.current = rune(input[0])
	}
	return l
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() (*Token, error) {
	l.skipWhitespace()

	// Check if we encountered any errors during whitespace processing
	if len(l.errors) > 0 {
		err := l.errors[0]
		l.errors = l.errors[1:] // Remove the error we're returning
		return nil, err
	}

	if l.isEOF() {
		// For EOF, position it at the start of the next line if we're not already at column 1
		token := &Token{
			Type:   TokenEOF,
			Value:  "",
			Line:   l.line,
			Column: l.column,
			Offset: l.pos,
		}
		// If we're not at the start of a line, position EOF at the next line
		if l.column > 1 {
			token.Line = l.line + 1
			token.Column = 1
		}
		return token, nil
	}

	var token *Token
	var err error

	// Check for document markers
	if l.checkString("---") {
		l.advance(3)
		token = l.createToken(TokenDocumentStart, "---")
		l.lastTokenLine = token.Line
		return token, nil
	}
	if l.checkString("...") {
		l.advance(3)
		token = l.createToken(TokenDocumentEnd, "...")
		l.lastTokenLine = token.Line
		return token, nil
	}

	// Check for comments
	if l.current == '#' {
		return l.scanComment()
	}

	// Check for flow indicators
	if l.inFlow > 0 {
		switch l.current {
		case ',':
			l.advance(1)
			token = l.createToken(TokenFlowEntry, ",")
			l.lastTokenLine = token.Line
			return token, nil
		case ']':
			l.advance(1)
			l.inFlow--
			token = l.createToken(TokenFlowSequenceEnd, "]")
			l.lastTokenLine = token.Line
			return token, nil
		case '}':
			l.advance(1)
			l.inFlow--
			token = l.createToken(TokenFlowMappingEnd, "}")
			l.lastTokenLine = token.Line
			return token, nil
		}
	}

	// Check for structure indicators
	switch l.current {
	case '-':
		if l.isWhitespace(l.peek()) {
			l.advance(1)
			token = l.createToken(TokenSequenceEntry, "-")
			l.lastTokenLine = token.Line
			return token, nil
		}
	case ':':
		if l.isWhitespace(l.peek()) || l.isEOF() {
			l.advance(1)
			token = l.createToken(TokenMappingValue, ":")
			l.lastTokenLine = token.Line
			return token, nil
		}
	case '?':
		if l.isWhitespace(l.peek()) {
			l.advance(1)
			token = l.createToken(TokenMappingKey, "?")
			l.lastTokenLine = token.Line
			return token, nil
		}
	case '[':
		l.advance(1)
		l.inFlow++
		token = l.createToken(TokenFlowSequenceStart, "[")
		l.lastTokenLine = token.Line
		return token, nil
	case '{':
		l.advance(1)
		l.inFlow++
		token = l.createToken(TokenFlowMappingStart, "{")
		l.lastTokenLine = token.Line
		return token, nil
	case '&':
		token, err = l.scanAnchor()
	case '*':
		token, err = l.scanAlias()
	case '!':
		token, err = l.scanTag()
	case '%':
		token, err = l.scanDirective()
	case '|':
		token, err = l.scanLiteralScalar()
	case '>':
		token, err = l.scanFoldedScalar()
	case '"':
		token, err = l.scanDoubleQuotedScalar()
	case '\'':
		token, err = l.scanSingleQuotedScalar()
	default:
		// Default to plain scalar
		token, err = l.scanPlainScalar()
	}

	if err != nil {
		return nil, err
	}
	if token != nil && token.Type != TokenComment {
		l.lastTokenLine = token.Line
	}
	return token, nil
}

// Helper methods

func (l *Lexer) advance(n int) {
	for i := 0; i < n; i++ {
		if l.current == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.pos++
		if l.pos < len(l.input) {
			l.current = rune(l.input[l.pos])
		} else {
			l.current = 0
		}
	}
}

func (l *Lexer) peek() rune {
	if l.pos+1 < len(l.input) {
		return rune(l.input[l.pos+1])
	}
	return 0
}

func (l *Lexer) isEOF() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func (l *Lexer) skipWhitespace() {
	consecutiveNewlines := 0
	atLineStart := l.column == 1
	startedAtBeginning := l.pos == 0
	hasTabAtLineStart := false
	tabLine := 0
	tabColumn := 0
	tabOffset := 0

	for l.isWhitespace(l.current) && !l.isEOF() {
		// Check for tabs at line start (YAML 1.2.2 spec forbids tabs in indentation)
		if atLineStart && l.current == '\t' {
			// Remember we saw a tab at line start, but don't error yet
			// We'll only error if there's actual content after the whitespace
			if !hasTabAtLineStart {
				hasTabAtLineStart = true
				tabLine = l.line
				tabColumn = l.column
				tabOffset = l.pos
			}
		}

		if l.current == '\n' {
			consecutiveNewlines++
			atLineStart = true
		} else if l.current == ' ' {
			// Spaces are fine, keep line start status unchanged
		} else if l.current == '\t' {
			// Tab after content on same line is OK, just not for indentation
			atLineStart = false
		} else if l.current != '\r' {
			// Reset on non-whitespace
			if startedAtBeginning && consecutiveNewlines > 0 {
				// At start of input, each newline creates a blank line
				l.blankLineCount = consecutiveNewlines
			} else if consecutiveNewlines > 1 {
				// Otherwise, N newlines = N-1 blank lines
				l.blankLineCount = consecutiveNewlines - 1
			}
			consecutiveNewlines = 0
			atLineStart = false
		}
		l.advance(1)
	}

	// Final check for blank lines at end of whitespace
	if startedAtBeginning && consecutiveNewlines > 0 {
		l.blankLineCount = consecutiveNewlines
	} else if consecutiveNewlines > 1 {
		l.blankLineCount = consecutiveNewlines - 1
	}

	// Only error on tabs if we're about to parse actual content
	if hasTabAtLineStart && !l.isEOF() {
		// We saw a tab at line start and there's content after the whitespace
		l.errors = append(l.errors, errors.New(
			"tabs are not allowed for indentation (YAML 1.2.2 spec section 6.1)",
			errors.Position{Line: tabLine, Column: tabColumn, Offset: tabOffset},
			errors.ErrorTypeLexer,
		))
	}
}

// scanEmptyLine creates explicit empty line tokens following user's ##EMPTY_LINE## strategy
func (l *Lexer) scanEmptyLine() *Token {
	startLine := l.line

	// Skip the newline character
	l.advance(1)

	return &Token{
		Type:   TokenEmptyLine,
		Value:  "##EMPTY_LINE##",
		Line:   startLine,
		Column: 1,
		Offset: l.pos - 1,
	}
}

func (l *Lexer) checkString(s string) bool {
	if l.pos+len(s) > len(l.input) {
		return false
	}
	return l.input[l.pos:l.pos+len(s)] == s
}

func (l *Lexer) createToken(typ TokenType, value string) *Token {
	// Store the column at the start of the token
	col := l.column - len(value)
	if col < 1 {
		col = l.column
	}
	return &Token{
		Type:   typ,
		Value:  value,
		Line:   l.line,
		Column: col,
		Offset: l.pos - len(value),
	}
}

func (l *Lexer) scanComment() (*Token, error) {
	start := l.pos
	startLine := l.line
	startCol := l.column

	l.advance(1) // skip #

	for l.current != '\n' && !l.isEOF() {
		l.advance(1)
	}

	value := l.input[start:l.pos]
	token := &Token{
		Type:             TokenComment,
		Value:            value,
		Line:             startLine,
		Column:           startCol,
		Offset:           start,
		BlankLinesBefore: l.blankLineCount,
	}

	// Check if this is an inline comment
	// It's inline if we've already returned a non-comment token on this line
	token.IsInline = (startLine == l.lastTokenLine && l.lastTokenLine > 0)

	// Reset blank line counter after comment
	l.blankLineCount = 0

	// Only update lastTokenLine if this is NOT an inline comment
	if !token.IsInline {
		l.lastTokenLine = startLine
	}

	return token, nil
}

func (l *Lexer) scanAnchor() (*Token, error) {
	start := l.pos
	l.advance(1) // skip &

	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' || l.current == '-' {
		l.advance(1)
	}

	value := l.input[start+1 : l.pos]
	return l.createToken(TokenAnchor, value), nil
}

func (l *Lexer) scanAlias() (*Token, error) {
	start := l.pos
	l.advance(1) // skip *

	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' || l.current == '-' {
		l.advance(1)
	}

	value := l.input[start+1 : l.pos]
	return l.createToken(TokenAlias, value), nil
}

func (l *Lexer) scanTag() (*Token, error) {
	start := l.pos
	l.advance(1) // skip !

	for !l.isWhitespace(l.current) && !l.isEOF() {
		l.advance(1)
	}

	value := l.input[start:l.pos]
	return l.createToken(TokenTag, value), nil
}

func (l *Lexer) scanDirective() (*Token, error) {
	start := l.pos
	l.advance(1) // skip %

	for l.current != '\n' && !l.isEOF() {
		l.advance(1)
	}

	value := l.input[start:l.pos]
	return l.createToken(TokenDirective, value), nil
}

func (l *Lexer) scanPlainScalar() (*Token, error) {
	var sb strings.Builder
	startCol := l.column

	for !l.isEOF() {
		// Check for comment start
		if l.current == '#' && (sb.Len() == 0 || l.isWhitespace(rune(sb.String()[sb.Len()-1]))) {
			// We hit a comment, trim trailing whitespace and return scalar
			value := strings.TrimSpace(sb.String())
			if value != "" {
				token := l.createToken(TokenPlainScalar, value)
				token.Style = ScalarStylePlain
				token.Column = startCol
				return token, nil
			}
			// If empty, this is actually a comment
			return l.scanComment()
		}

		if l.current == ':' && l.isWhitespace(l.peek()) {
			break
		}
		if l.current == '\n' {
			break
		}
		if l.inFlow > 0 && (l.current == ',' || l.current == ']' || l.current == '}') {
			break
		}

		sb.WriteRune(l.current)
		l.advance(1)
	}

	value := strings.TrimSpace(sb.String())
	token := l.createToken(TokenPlainScalar, value)
	token.Style = ScalarStylePlain
	token.Column = startCol
	return token, nil
}

func (l *Lexer) scanSingleQuotedScalar() (*Token, error) {
	var sb strings.Builder
	l.advance(1) // skip opening '

	for !l.isEOF() {
		if l.current == '\'' {
			if l.peek() == '\'' {
				sb.WriteRune('\'')
				l.advance(2)
			} else {
				l.advance(1) // skip closing '
				break
			}
		} else {
			sb.WriteRune(l.current)
			l.advance(1)
		}
	}

	token := l.createToken(TokenSingleQuotedScalar, sb.String())
	token.Style = ScalarStyleSingleQuoted
	return token, nil
}

func (l *Lexer) scanDoubleQuotedScalar() (*Token, error) {
	var sb strings.Builder
	l.advance(1) // skip opening "

	for !l.isEOF() {
		if l.current == '"' {
			l.advance(1) // skip closing "
			break
		} else if l.current == '\\' {
			l.advance(1)
			switch l.current {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune(l.current)
			}
			l.advance(1)
		} else {
			sb.WriteRune(l.current)
			l.advance(1)
		}
	}

	token := l.createToken(TokenDoubleQuotedScalar, sb.String())
	token.Style = ScalarStyleDoubleQuoted
	return token, nil
}

func (l *Lexer) scanLiteralScalar() (*Token, error) {
	var sb strings.Builder
	l.advance(1) // skip |

	// Skip to next line
	for l.current != '\n' && !l.isEOF() {
		l.advance(1)
	}
	l.advance(1)

	// Read literal block
	baseIndent := l.column
	for !l.isEOF() {
		if l.column < baseIndent && !l.isWhitespace(l.current) {
			break
		}
		sb.WriteRune(l.current)
		l.advance(1)
	}

	token := l.createToken(TokenLiteralScalar, strings.TrimRight(sb.String(), "\n"))
	token.Style = ScalarStyleLiteral
	return token, nil
}

func (l *Lexer) scanFoldedScalar() (*Token, error) {
	var sb strings.Builder
	l.advance(1) // skip >

	// Skip to next line
	for l.current != '\n' && !l.isEOF() {
		l.advance(1)
	}
	l.advance(1)

	// Read folded block
	baseIndent := l.column
	for !l.isEOF() {
		if l.column < baseIndent && !l.isWhitespace(l.current) {
			break
		}
		sb.WriteRune(l.current)
		l.advance(1)
	}

	token := l.createToken(TokenFoldedScalar, strings.TrimRight(sb.String(), "\n"))
	token.Style = ScalarStyleFolded
	return token, nil
}

// Initialize reads the first character
func (l *Lexer) Initialize() error {
	if l.input != "" {
		if len(l.input) > 0 {
			r, _ := utf8.DecodeRuneInString(l.input)
			l.current = r
		}
	} else if l.reader != nil {
		// Read from reader implementation
		// This would be implemented based on the reader
	}
	return nil
}

// GetInput returns the input string
func (l *Lexer) GetInput() string {
	return l.input
}

// GetPos returns the current position
func (l *Lexer) GetPos() int {
	return l.pos
}
