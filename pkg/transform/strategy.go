package transform

import (
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/node"
)

// SortStrategy defines the interface for sorting strategies
type SortStrategy interface {
	// Name returns the strategy name
	Name() string

	// Description returns a human-readable description
	Description() string

	// ShouldSort returns true if sorting should be applied
	ShouldSort() bool

	// Compare compares two values and returns true if a should come before b
	Compare(a, b string) bool

	// PreProcess can modify values before comparison (e.g., lowercase)
	PreProcess(value string) string

	// CanSort checks if a specific node can be sorted by this strategy
	CanSort(node node.Node) bool
}

// BaseStrategy provides common functionality for strategies
type BaseStrategy struct {
	name        string
	description string
	shouldSort  bool
}

func (s *BaseStrategy) Name() string               { return s.name }
func (s *BaseStrategy) Description() string        { return s.description }
func (s *BaseStrategy) ShouldSort() bool           { return s.shouldSort }
func (s *BaseStrategy) PreProcess(v string) string { return v }
func (s *BaseStrategy) CanSort(n node.Node) bool   { return true }

// KeepOriginalStrategy preserves the original order (default)
type KeepOriginalStrategy struct {
	BaseStrategy
}

func NewKeepOriginalStrategy() *KeepOriginalStrategy {
	return &KeepOriginalStrategy{
		BaseStrategy{
			name:        "keep-original",
			description: "Preserves the original order without any modifications (DEFAULT)",
			shouldSort:  false,
		},
	}
}

func (s *KeepOriginalStrategy) Compare(a, b string) bool {
	return false // Never change order
}

// AscendingStrategy sorts in ascending order
type AscendingStrategy struct {
	BaseStrategy
	CaseSensitive bool
	NumericAware  bool
}

func NewAscendingStrategy() *AscendingStrategy {
	return &AscendingStrategy{
		BaseStrategy: BaseStrategy{
			name:        "ascending",
			description: "Sorts in ascending order (A-Z, 0-9)",
			shouldSort:  true,
		},
		CaseSensitive: true,
		NumericAware:  false,
	}
}

func (s *AscendingStrategy) Compare(a, b string) bool {
	if !s.CaseSensitive {
		a = s.PreProcess(a)
		b = s.PreProcess(b)
	}
	return a < b
}

func (s *AscendingStrategy) PreProcess(value string) string {
	if !s.CaseSensitive {
		return strings.ToLower(value)
	}
	return value
}

// DescendingStrategy sorts in descending order
type DescendingStrategy struct {
	BaseStrategy
	CaseSensitive bool
	NumericAware  bool
}

func NewDescendingStrategy() *DescendingStrategy {
	return &DescendingStrategy{
		BaseStrategy: BaseStrategy{
			name:        "descending",
			description: "Sorts in descending order (Z-A, 9-0)",
			shouldSort:  true,
		},
		CaseSensitive: true,
		NumericAware:  false,
	}
}

func (s *DescendingStrategy) Compare(a, b string) bool {
	if !s.CaseSensitive {
		a = s.PreProcess(a)
		b = s.PreProcess(b)
	}
	return a > b
}

func (s *DescendingStrategy) PreProcess(value string) string {
	if !s.CaseSensitive {
		return strings.ToLower(value)
	}
	return value
}

// PriorityStrategy sorts based on priority lists
type PriorityStrategy struct {
	BaseStrategy
	PriorityKeys []string     // Keys in priority order
	ThenStrategy SortStrategy // Strategy for non-priority items
}

func NewPriorityStrategy(priorities []string, fallback SortStrategy) *PriorityStrategy {
	if fallback == nil {
		fallback = NewAscendingStrategy()
	}
	return &PriorityStrategy{
		BaseStrategy: BaseStrategy{
			name:        "priority",
			description: "Sorts based on predefined priority order",
			shouldSort:  true,
		},
		PriorityKeys: priorities,
		ThenStrategy: fallback,
	}
}

func (s *PriorityStrategy) Compare(a, b string) bool {
	aIdx, bIdx := -1, -1

	// Find priority indices
	for i, key := range s.PriorityKeys {
		if key == a {
			aIdx = i
		}
		if key == b {
			bIdx = i
		}
	}

	// Both have priority - sort by priority order
	if aIdx >= 0 && bIdx >= 0 {
		return aIdx < bIdx
	}

	// Only a has priority
	if aIdx >= 0 {
		return true
	}

	// Only b has priority
	if bIdx >= 0 {
		return false
	}

	// Neither has priority - use fallback strategy
	return s.ThenStrategy.Compare(a, b)
}

// GroupStrategy groups related items together
type GroupStrategy struct {
	BaseStrategy
	Groups          map[string][]string // Group name -> items in that group
	GroupOrder      []string            // Order of groups
	InGroupStrategy SortStrategy        // How to sort within groups
}

func NewGroupStrategy() *GroupStrategy {
	return &GroupStrategy{
		BaseStrategy: BaseStrategy{
			name:        "group",
			description: "Groups related items together",
			shouldSort:  true,
		},
		Groups:          make(map[string][]string),
		GroupOrder:      []string{},
		InGroupStrategy: NewAscendingStrategy(),
	}
}

func (s *GroupStrategy) AddGroup(name string, items ...string) {
	s.Groups[name] = items
	s.GroupOrder = append(s.GroupOrder, name)
}

func (s *GroupStrategy) Compare(a, b string) bool {
	aGroup, bGroup := s.findGroup(a), s.findGroup(b)

	// Same group or no group - use in-group strategy
	if aGroup == bGroup {
		return s.InGroupStrategy.Compare(a, b)
	}

	// Different groups - sort by group order
	for _, group := range s.GroupOrder {
		if group == aGroup {
			return true
		}
		if group == bGroup {
			return false
		}
	}

	// Fallback to in-group strategy
	return s.InGroupStrategy.Compare(a, b)
}

func (s *GroupStrategy) findGroup(item string) string {
	for group, items := range s.Groups {
		for _, i := range items {
			if i == item {
				return group
			}
		}
	}
	return ""
}

// CustomStrategy allows user-defined sorting logic
type CustomStrategy struct {
	BaseStrategy
	CompareFunc func(a, b string) bool
}

func NewCustomStrategy(name, description string, compareFunc func(a, b string) bool) *CustomStrategy {
	return &CustomStrategy{
		BaseStrategy: BaseStrategy{
			name:        name,
			description: description,
			shouldSort:  true,
		},
		CompareFunc: compareFunc,
	}
}

func (s *CustomStrategy) Compare(a, b string) bool {
	if s.CompareFunc != nil {
		return s.CompareFunc(a, b)
	}
	return false
}

// Common pre-built strategies

// YAMLDocumentStrategy - Common YAML document key ordering
func NewYAMLDocumentStrategy() *PriorityStrategy {
	priorities := []string{
		"apiVersion",
		"kind",
		"metadata",
		"name",
		"namespace",
		"spec",
		"data",
		"status",
	}
	return NewPriorityStrategy(priorities, NewAscendingStrategy())
}

// PackageJSONStrategy - Common package.json key ordering
func NewPackageJSONStrategy() *PriorityStrategy {
	priorities := []string{
		"name",
		"version",
		"description",
		"main",
		"scripts",
		"dependencies",
		"devDependencies",
		"peerDependencies",
	}
	return NewPriorityStrategy(priorities, NewAscendingStrategy())
}
