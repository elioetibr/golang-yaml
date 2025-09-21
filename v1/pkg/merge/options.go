package merge

import (
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// Strategy defines how values-with-comments should be merged
type Strategy int

const (
	// StrategyDeep performs deep recursive merging of nested structures
	StrategyDeep Strategy = iota
	// StrategyShallow only merges top-level keys
	StrategyShallow
	// StrategyOverride completely replaces base with override
	StrategyOverride
	// StrategyAppend appends arrays instead of replacing them
	StrategyAppend
)

// Options configures the merge behavior
type Options struct {
	// Strategy defines the merge strategy to use
	Strategy Strategy

	// PreserveComments controls whether comments from base are preserved
	PreserveComments bool

	// PreserveBlankLines controls whether blank lines are preserved
	PreserveBlankLines bool

	// ArrayMergeStrategy defines how arrays should be merged
	ArrayMergeStrategy ArrayMergeStrategy

	// OverrideEmpty allows empty values-with-comments to override non-empty ones
	OverrideEmpty bool

	// MergeAnchors controls whether anchor/alias references should be resolved
	MergeAnchors bool

	// CustomMergeFunc allows custom merge logic for specific keys
	CustomMergeFunc func(key string, base, override node.Node) (node.Node, bool)

	// KeyPriority defines which document's keys take priority for ordering
	KeyPriority KeyPriority

	// KeepDefaultLineBetweenSections preserves original blank line count between sections
	// When true, the original number of blank lines is preserved
	// When false, uses DefaultLineBetweenSections value
	KeepDefaultLineBetweenSections bool

	// DefaultLineBetweenSections specifies number of blank lines between sections
	// Only used when KeepDefaultLineBetweenSections is false
	// Default value is 1
	DefaultLineBetweenSections int
}

// ArrayMergeStrategy defines how arrays should be merged
type ArrayMergeStrategy int

const (
	// ArrayReplace replaces the entire array (default)
	ArrayReplace ArrayMergeStrategy = iota
	// ArrayAppend appends override array to base array
	ArrayAppend
	// ArrayMergeByIndex merges arrays element by element
	ArrayMergeByIndex
	// ArrayMergeByKey merges array of maps by a key field
	ArrayMergeByKey
)

// KeyPriority defines which document's key ordering takes priority
type KeyPriority int

const (
	// KeyPriorityBase preserves base document's key ordering
	KeyPriorityBase KeyPriority = iota
	// KeyPriorityOverride uses override document's key ordering
	KeyPriorityOverride
	// KeyPriorityAlphabetical sorts keys alphabetically
	KeyPriorityAlphabetical
)

// DefaultOptions returns the default merge options
// By default, we preserve comments and blank lines to maintain
// the original document structure and documentation
func DefaultOptions() *Options {
	return &Options{
		Strategy:                       StrategyDeep,
		PreserveComments:               true, // Always preserve comments by default
		PreserveBlankLines:             true, // Always preserve blank lines by default
		ArrayMergeStrategy:             ArrayReplace,
		OverrideEmpty:                  false,
		MergeAnchors:                   true,
		KeyPriority:                    KeyPriorityBase, // Maintain base document's structure
		KeepDefaultLineBetweenSections: true,            // Keep original blank line count by default
		DefaultLineBetweenSections:     1,               // 1 blank line between sections when normalized
	}
}

// WithStrategy returns options with the specified strategy
func (o *Options) WithStrategy(s Strategy) *Options {
	o.Strategy = s
	return o
}

// WithArrayStrategy returns options with the specified array merge strategy
func (o *Options) WithArrayStrategy(s ArrayMergeStrategy) *Options {
	o.ArrayMergeStrategy = s
	return o
}

// WithKeyPriority returns options with the specified key priority
func (o *Options) WithKeyPriority(p KeyPriority) *Options {
	o.KeyPriority = p
	return o
}

// WithOverrideEmpty returns options with the specified override empty behavior
func (o *Options) WithOverrideEmpty(override bool) *Options {
	o.OverrideEmpty = override
	return o
}

// WithSectionHandling returns options with the specified section handling configuration
func (o *Options) WithSectionHandling(keepDefault bool, defaultLines int) *Options {
	o.KeepDefaultLineBetweenSections = keepDefault
	o.DefaultLineBetweenSections = defaultLines
	return o
}

// WithNormalizedSections returns options that normalize blank lines between sections
func (o *Options) WithNormalizedSections(lines int) *Options {
	o.KeepDefaultLineBetweenSections = false
	o.DefaultLineBetweenSections = lines
	return o
}
