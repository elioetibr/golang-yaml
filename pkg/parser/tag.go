package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// TagResolver handles YAML tag resolution
type TagResolver struct {
	// Tag shorthand definitions
	tagShorthands map[string]string

	// Custom tag handlers
	customHandlers map[string]TagHandler
}

// TagHandler is a function that processes a tagged value
type TagHandler func(value string) (interface{}, error)

// NewTagResolver creates a new tag resolver
func NewTagResolver() *TagResolver {
	tr := &TagResolver{
		tagShorthands:  make(map[string]string),
		customHandlers: make(map[string]TagHandler),
	}

	// Initialize default tag shorthands
	tr.initializeDefaults()
	return tr
}

// initializeDefaults sets up the default YAML 1.2 tags
func (tr *TagResolver) initializeDefaults() {
	// Core schema tags
	tr.tagShorthands["!!"] = "tag:yaml.org,2002:"

	// Common type tags
	tr.tagShorthands["!!str"] = "tag:yaml.org,2002:str"
	tr.tagShorthands["!!int"] = "tag:yaml.org,2002:int"
	tr.tagShorthands["!!float"] = "tag:yaml.org,2002:float"
	tr.tagShorthands["!!bool"] = "tag:yaml.org,2002:bool"
	tr.tagShorthands["!!null"] = "tag:yaml.org,2002:null"
	tr.tagShorthands["!!binary"] = "tag:yaml.org,2002:binary"
	tr.tagShorthands["!!timestamp"] = "tag:yaml.org,2002:timestamp"
	tr.tagShorthands["!!omap"] = "tag:yaml.org,2002:omap"
	tr.tagShorthands["!!pairs"] = "tag:yaml.org,2002:pairs"
	tr.tagShorthands["!!set"] = "tag:yaml.org,2002:set"
	tr.tagShorthands["!!seq"] = "tag:yaml.org,2002:seq"
	tr.tagShorthands["!!map"] = "tag:yaml.org,2002:map"

	// Register default handlers
	tr.customHandlers["tag:yaml.org,2002:str"] = tr.handleString
	tr.customHandlers["tag:yaml.org,2002:int"] = tr.handleInt
	tr.customHandlers["tag:yaml.org,2002:float"] = tr.handleFloat
	tr.customHandlers["tag:yaml.org,2002:bool"] = tr.handleBool
	tr.customHandlers["tag:yaml.org,2002:null"] = tr.handleNull
	tr.customHandlers["tag:yaml.org,2002:timestamp"] = tr.handleTimestamp
}

// AddTagDirective adds a TAG directive mapping
func (tr *TagResolver) AddTagDirective(handle, prefix string) {
	tr.tagShorthands[handle] = prefix
}

// RegisterCustomHandler registers a custom tag handler
func (tr *TagResolver) RegisterCustomHandler(tag string, handler TagHandler) {
	tr.customHandlers[tag] = handler
}

// ResolveTag resolves a tag to its full form
func (tr *TagResolver) ResolveTag(tag string) string {
	if tag == "" {
		return ""
	}

	// Check if it's a verbatim tag (starts with !)
	if strings.HasPrefix(tag, "!<") && strings.HasSuffix(tag, ">") {
		// Verbatim tag: !<tag:example.com,2014:something>
		return tag[2 : len(tag)-1]
	}

	// Check if it's a local tag (!something)
	if strings.HasPrefix(tag, "!") && !strings.HasPrefix(tag, "!!") {
		// Local tag, return as-is
		return tag
	}

	// Check for shorthand
	if shorthand, exists := tr.tagShorthands[tag]; exists {
		return shorthand
	}

	// Check for partial match (e.g., !!str)
	for short, full := range tr.tagShorthands {
		if tag == short {
			return full
		}
		// Handle secondary tags like !!myapp:custom
		if strings.HasPrefix(tag, short) {
			suffix := tag[len(short):]
			return full + suffix
		}
	}

	return tag
}

// ProcessTaggedValue processes a value according to its tag
func (tr *TagResolver) ProcessTaggedValue(tag, value string) (interface{}, error) {
	resolvedTag := tr.ResolveTag(tag)

	// Check for custom handler
	if handler, exists := tr.customHandlers[resolvedTag]; exists {
		return handler(value)
	}

	// Default: return the string value
	return value, nil
}

// Default tag handlers

func (tr *TagResolver) handleString(value string) (interface{}, error) {
	return value, nil
}

func (tr *TagResolver) handleInt(value string) (interface{}, error) {
	// Try different integer formats
	value = strings.TrimSpace(value)

	// Handle hex (0x...), octal (0o...), binary (0b...)
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		i, err := strconv.ParseInt(value[2:], 16, 64)
		return i, err
	}
	if strings.HasPrefix(value, "0o") || strings.HasPrefix(value, "0O") {
		i, err := strconv.ParseInt(value[2:], 8, 64)
		return i, err
	}
	if strings.HasPrefix(value, "0b") || strings.HasPrefix(value, "0B") {
		i, err := strconv.ParseInt(value[2:], 2, 64)
		return i, err
	}

	// Handle underscores in numbers (e.g., 1_000_000)
	value = strings.ReplaceAll(value, "_", "")

	// Parse as decimal
	return strconv.ParseInt(value, 10, 64)
}

func (tr *TagResolver) handleFloat(value string) (interface{}, error) {
	value = strings.TrimSpace(value)

	// Handle special float values-with-comments
	switch strings.ToLower(value) {
	case ".inf", "+.inf", "inf", "+inf":
		return math.Inf(1), nil // +Inf
	case "-.inf", "-inf":
		return math.Inf(-1), nil // -Inf
	case ".nan", "nan":
		return math.NaN(), nil // NaN
	}

	// Handle underscores in numbers
	value = strings.ReplaceAll(value, "_", "")

	return strconv.ParseFloat(value, 64)
}

func (tr *TagResolver) handleBool(value string) (interface{}, error) {
	switch strings.ToLower(value) {
	case "true", "yes", "on", "y":
		return true, nil
	case "false", "no", "off", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

func (tr *TagResolver) handleNull(value string) (interface{}, error) {
	return nil, nil
}

func (tr *TagResolver) handleTimestamp(value string) (interface{}, error) {
	// Try parsing various timestamp formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00", // RFC3339Nano
		"2006-01-02T15:04:05",                 // Without timezone
		"2006-01-02",                          // Date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return nil, fmt.Errorf("invalid timestamp: %s", value)
}

// CommonTags provides common YAML tags
var CommonTags = struct {
	Str       string
	Int       string
	Float     string
	Bool      string
	Null      string
	Binary    string
	Timestamp string
	Seq       string
	Map       string
}{
	Str:       "!!str",
	Int:       "!!int",
	Float:     "!!float",
	Bool:      "!!bool",
	Null:      "!!null",
	Binary:    "!!binary",
	Timestamp: "!!timestamp",
	Seq:       "!!seq",
	Map:       "!!map",
}

// InferTag infers the appropriate tag for a plain scalar value
func InferTag(value string) string {
	// Check for null
	if value == "" || value == "~" || strings.ToLower(value) == "null" {
		return CommonTags.Null
	}

	// Check for boolean
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" || lower == "yes" || lower == "no" ||
		lower == "on" || lower == "off" {
		return CommonTags.Bool
	}

	// Check for integer
	if _, err := strconv.ParseInt(strings.ReplaceAll(value, "_", ""), 10, 64); err == nil {
		return CommonTags.Int
	}

	// Check for float
	if _, err := strconv.ParseFloat(strings.ReplaceAll(value, "_", ""), 64); err == nil {
		return CommonTags.Float
	}

	// Check for timestamp-like strings
	if strings.Contains(value, "-") && strings.Contains(value, ":") {
		if _, err := time.Parse(time.RFC3339, value); err == nil {
			return CommonTags.Timestamp
		}
	}

	// Default to string
	return CommonTags.Str
}
