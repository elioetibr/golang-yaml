package errors

import "fmt"

// Position represents a position in the YAML document
type Position struct {
	Line   int
	Column int
	Offset int
}

// YAMLError represents a YAML parsing/serialization error with position information
type YAMLError struct {
	Message  string
	Position Position
	Context  string
	Type     ErrorType
}

// ErrorType represents the type of YAML error
type ErrorType int

const (
	ErrorTypeLexer ErrorType = iota
	ErrorTypeParser
	ErrorTypeSerializer
	ErrorTypeDecoder
	ErrorTypeEncoder
	ErrorTypeValidation
)

func (e *YAMLError) Error() string {
	return fmt.Sprintf("YAML error at line %d, column %d: %s",
		e.Position.Line, e.Position.Column, e.Message)
}

// New creates a new YAML error
func New(msg string, pos Position, errType ErrorType) *YAMLError {
	return &YAMLError{
		Message:  msg,
		Position: pos,
		Type:     errType,
	}
}

// Wrap wraps an existing error with YAML context
func Wrap(err error, pos Position, errType ErrorType) *YAMLError {
	if err == nil {
		return nil
	}
	return &YAMLError{
		Message:  err.Error(),
		Position: pos,
		Type:     errType,
	}
}
