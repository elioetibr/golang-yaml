package transform

import "github.com/elioetibr/golang-yaml/pkg/node"

// ConfigBuilder provides a fluent API for building sort and format configurations
type ConfigBuilder struct {
	sortConfig   *SortConfig
	formatConfig *FormatConfig
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		sortConfig:   DefaultSortConfig(),
		formatConfig: DefaultFormatConfig(),
	}
}

// Sorting Configuration Methods

// KeepOriginal sets mode to keep original order (default)
func (b *ConfigBuilder) KeepOriginal() *ConfigBuilder {
	b.sortConfig.Mode = SortModeKeepOriginal
	return b
}

// SortAscending sets ascending sort mode
func (b *ConfigBuilder) SortAscending() *ConfigBuilder {
	b.sortConfig.Mode = SortModeAscending
	return b
}

// SortDescending sets descending sort mode
func (b *ConfigBuilder) SortDescending() *ConfigBuilder {
	b.sortConfig.Mode = SortModeDescending
	return b
}

// ByKey sets sorting by keys (for mappings)
func (b *ConfigBuilder) ByKey() *ConfigBuilder {
	b.sortConfig.SortBy = SortByKey
	return b
}

// ByValue sets sorting by values (for sequences and mappings)
func (b *ConfigBuilder) ByValue() *ConfigBuilder {
	b.sortConfig.SortBy = SortByValue
	return b
}

// WithNumericSort enables numeric-aware sorting
func (b *ConfigBuilder) WithNumericSort() *ConfigBuilder {
	b.sortConfig.NumericSort = true
	return b
}

// CaseInsensitive sets case-insensitive sorting
func (b *ConfigBuilder) CaseInsensitive() *ConfigBuilder {
	b.sortConfig.CaseSensitive = false
	return b
}

// Formatting Configuration Methods

// WithBlankLines sets the default number of blank lines before comments
func (b *ConfigBuilder) WithBlankLines(n int) *ConfigBuilder {
	b.formatConfig.DefaultBlankLinesBeforeComment = n
	b.formatConfig.BlankLinesBeforeHeadComment = n
	b.formatConfig.BlankLinesBeforeKeyComment = n
	return b
}

// WithKeyCommentSpacing sets blank lines before key comments
func (b *ConfigBuilder) WithKeyCommentSpacing(n int) *ConfigBuilder {
	b.formatConfig.BlankLinesBeforeKeyComment = n
	return b
}

// WithValueCommentSpacing sets blank lines before value comments
func (b *ConfigBuilder) WithValueCommentSpacing(n int) *ConfigBuilder {
	b.formatConfig.BlankLinesBeforeValueComment = n
	return b
}

// ForceFormatting forces blank line formatting even if original had different spacing
func (b *ConfigBuilder) ForceFormatting() *ConfigBuilder {
	b.formatConfig.PreserveOriginal = false
	b.formatConfig.ForceBlankLines = true
	return b
}

// PreserveFormatting preserves original formatting
func (b *ConfigBuilder) PreserveFormatting() *ConfigBuilder {
	b.formatConfig.PreserveOriginal = true
	b.formatConfig.ForceBlankLines = false
	return b
}

// WithSectionMarkers adds patterns for section markers that get extra spacing
func (b *ConfigBuilder) WithSectionMarkers(patterns ...string) *ConfigBuilder {
	b.formatConfig.SectionMarkers = append(b.formatConfig.SectionMarkers, patterns...)
	return b
}

// WithSectionExtraLines sets extra blank lines for section markers
func (b *ConfigBuilder) WithSectionExtraLines(n int) *ConfigBuilder {
	b.formatConfig.SectionMarkerExtraLines = n
	return b
}

// Build returns the configured sort and format configurations
func (b *ConfigBuilder) Build() (*SortConfig, *FormatConfig) {
	return b.sortConfig, b.formatConfig
}

// BuildSorter creates a sorter with the current configuration
func (b *ConfigBuilder) BuildSorter() *Sorter {
	return NewSorter(b.sortConfig)
}

// BuildFormatter creates a formatter with the current configuration
func (b *ConfigBuilder) BuildFormatter() *Formatter {
	return NewFormatter(b.formatConfig)
}

// Apply applies both sorting and formatting to a node
func (b *ConfigBuilder) Apply(n node.Node) node.Node {
	return FormatWithSorting(n, b.sortConfig, b.formatConfig)
}

// Common Presets

// StandardConfig returns a standard configuration with sensible defaults
func StandardConfig() *ConfigBuilder {
	return NewConfigBuilder().
		KeepOriginal().
		ByKey().
		WithBlankLines(1).
		PreserveFormatting()
}

// CleanupConfig returns configuration for cleaning up YAML files
func CleanupConfig() *ConfigBuilder {
	return NewConfigBuilder().
		SortAscending().
		ByKey().
		WithBlankLines(1).
		ForceFormatting().
		WithSectionMarkers("^#\\s*---", "^#\\s*Section:").
		WithSectionExtraLines(1)
}

// MinimalConfig returns configuration for minimal formatting
func MinimalConfig() *ConfigBuilder {
	return NewConfigBuilder().
		KeepOriginal().
		WithBlankLines(0).
		ForceFormatting()
}

// ReadableConfig returns configuration optimized for readability
func ReadableConfig() *ConfigBuilder {
	return NewConfigBuilder().
		SortAscending().
		ByKey().
		WithBlankLines(1).
		WithKeyCommentSpacing(1).
		WithValueCommentSpacing(0).
		ForceFormatting().
		WithSectionExtraLines(2)
}
