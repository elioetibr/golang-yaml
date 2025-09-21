package parser

import (
	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
)

// LineTracker tracks line information and empty lines in the document
type LineTracker struct {
	lines           map[int]*LineInfo
	emptyLines      []int
	lastLine        int
	maxLine         int
	tokensByLine    map[int][]*lexer.Token
}

// LineInfo stores information about a specific line
type LineInfo struct {
	Number        int
	IsEmpty       bool
	HasComment    bool
	HasContent    bool
	IndentLevel   int
	Tokens        []*lexer.Token
}

// NewLineTracker creates a new line tracker
func NewLineTracker() *LineTracker {
	return &LineTracker{
		lines:        make(map[int]*LineInfo),
		emptyLines:   make([]int, 0),
		tokensByLine: make(map[int][]*lexer.Token),
	}
}

// TrackToken records a token's line information
func (lt *LineTracker) TrackToken(token *lexer.Token) {
	if token == nil {
		return
	}

	line := token.Line
	if line > lt.maxLine {
		lt.maxLine = line
	}

	// Get or create line info
	info := lt.getOrCreateLine(line)

	// Update line info based on token type
	switch token.Type {
	case lexer.TokenComment:
		info.HasComment = true
	case lexer.TokenNewLine:
		if !info.HasContent && !info.HasComment {
			info.IsEmpty = true
			lt.emptyLines = append(lt.emptyLines, line)
		}
	default:
		info.HasContent = true
		if info.IndentLevel == 0 && token.Column > 0 {
			info.IndentLevel = token.Column
		}
	}

	// Track token
	info.Tokens = append(info.Tokens, token)
	lt.tokensByLine[line] = append(lt.tokensByLine[line], token)
	lt.lastLine = line
}

// TrackEmptyLine marks a line as empty
func (lt *LineTracker) TrackEmptyLine(line int) {
	info := lt.getOrCreateLine(line)
	info.IsEmpty = true
	if !contains(lt.emptyLines, line) {
		lt.emptyLines = append(lt.emptyLines, line)
	}
}

// EmptyLinesBefore counts empty lines before a given line
func (lt *LineTracker) EmptyLinesBefore(line int) int {
	count := 0
	for i := line - 1; i > 0; i-- {
		if info, exists := lt.lines[i]; exists && info.IsEmpty {
			count++
		} else if info != nil && (info.HasContent || info.HasComment) {
			break
		}
	}
	return count
}

// EmptyLinesAfter counts empty lines after a given line
func (lt *LineTracker) EmptyLinesAfter(line int) int {
	count := 0
	for i := line + 1; i <= lt.maxLine; i++ {
		if info, exists := lt.lines[i]; exists && info.IsEmpty {
			count++
		} else if info != nil && (info.HasContent || info.HasComment) {
			break
		}
	}
	return count
}

// ConsecutiveEmptyLines counts consecutive empty lines at a position
func (lt *LineTracker) ConsecutiveEmptyLines(line int) int {
	if info, exists := lt.lines[line]; !exists || !info.IsEmpty {
		return 0
	}

	count := 1

	// Count before
	for i := line - 1; i > 0; i-- {
		if info, exists := lt.lines[i]; exists && info.IsEmpty {
			count++
		} else {
			break
		}
	}

	// Count after
	for i := line + 1; i <= lt.maxLine; i++ {
		if info, exists := lt.lines[i]; exists && info.IsEmpty {
			count++
		} else {
			break
		}
	}

	return count
}

// IsEmptyLine checks if a line is empty
func (lt *LineTracker) IsEmptyLine(line int) bool {
	if info, exists := lt.lines[line]; exists {
		return info.IsEmpty
	}
	return false
}

// HasCommentOnLine checks if a line has a comment
func (lt *LineTracker) HasCommentOnLine(line int) bool {
	if info, exists := lt.lines[line]; exists {
		return info.HasComment
	}
	return false
}

// GetLineIndent returns the indentation level of a line
func (lt *LineTracker) GetLineIndent(line int) int {
	if info, exists := lt.lines[line]; exists {
		return info.IndentLevel
	}
	return 0
}

// GetTokensOnLine returns all tokens on a specific line
func (lt *LineTracker) GetTokensOnLine(line int) []*lexer.Token {
	return lt.tokensByLine[line]
}

// getOrCreateLine gets or creates line info
func (lt *LineTracker) getOrCreateLine(line int) *LineInfo {
	if info, exists := lt.lines[line]; exists {
		return info
	}

	info := &LineInfo{
		Number: line,
		Tokens: make([]*lexer.Token, 0),
	}
	lt.lines[line] = info
	return info
}

// Helper function
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// ParserLookahead provides lookahead capability for the parser
type ParserLookahead struct {
	lexer      *lexer.Lexer
	buffer     []*lexer.Token
	bufferSize int
	lineTracker *LineTracker
}

// NewParserLookahead creates a new parser lookahead
func NewParserLookahead(l *lexer.Lexer) *ParserLookahead {
	return &ParserLookahead{
		lexer:       l,
		buffer:      make([]*lexer.Token, 0),
		bufferSize:  5, // Look ahead up to 5 tokens
		lineTracker: NewLineTracker(),
	}
}

// PeekNext returns the next token without consuming it
func (pl *ParserLookahead) PeekNext() *lexer.Token {
	if len(pl.buffer) == 0 {
		pl.fillBuffer(1)
	}
	if len(pl.buffer) > 0 {
		return pl.buffer[0]
	}
	return nil
}

// PeekNextNonEmpty returns the next non-empty token
func (pl *ParserLookahead) PeekNextNonEmpty() *lexer.Token {
	pl.fillBuffer(pl.bufferSize)

	for _, token := range pl.buffer {
		if token != nil && token.Type != lexer.TokenNewLine {
			return token
		}
	}
	return nil
}

// CountEmptyLinesBefore counts empty lines before a line
func (pl *ParserLookahead) CountEmptyLinesBefore(line int) int {
	return pl.lineTracker.EmptyLinesBefore(line)
}

// CountEmptyLinesAfter counts empty lines after a line
func (pl *ParserLookahead) CountEmptyLinesAfter(line int) int {
	return pl.lineTracker.EmptyLinesAfter(line)
}

// fillBuffer fills the lookahead buffer
func (pl *ParserLookahead) fillBuffer(n int) {
	for len(pl.buffer) < n {
		token, err := pl.lexer.NextToken()
		if err != nil || token == nil {
			break
		}
		pl.buffer = append(pl.buffer, token)
		pl.lineTracker.TrackToken(token)
	}
}

// Consume removes a token from the buffer
func (pl *ParserLookahead) Consume() *lexer.Token {
	if len(pl.buffer) == 0 {
		pl.fillBuffer(1)
	}
	if len(pl.buffer) > 0 {
		token := pl.buffer[0]
		pl.buffer = pl.buffer[1:]
		return token
	}
	return nil
}