package transform

import (
	"github.com/elioetibr/golang-yaml/pkg/node"
	"strings"
)

// SectionSorter handles section-aware sorting
type SectionSorter struct {
	config *SortConfig
	sorter *Sorter
}

// NewSectionSorter creates a new section-aware sorter
func NewSectionSorter(config *SortConfig) *SectionSorter {
	return &SectionSorter{
		config: config,
		sorter: NewSorter(config),
	}
}

// SortWithSections sorts a node tree respecting section boundaries
func (ss *SectionSorter) SortWithSections(root node.Node) node.Node {
	switch n := root.(type) {
	case *node.MappingNode:
		return ss.sortMappingWithSections(n)
	case *node.SequenceNode:
		return ss.sortSequenceWithSections(n)
	default:
		// For scalars and other types, just return as-is
		return root
	}
}

// sortMappingWithSections sorts a mapping node respecting section boundaries
func (ss *SectionSorter) sortMappingWithSections(mapping *node.MappingNode) *node.MappingNode {
	if len(mapping.Pairs) == 0 {
		return mapping
	}

	// Detect sections based on markers
	sections := ss.detectSections(mapping)

	// If no sections found, sort normally
	if len(sections) == 0 {
		return ss.sorter.sortMapping(mapping)
	}

	// Sort each section independently
	sortedPairs := make([]*node.MappingPair, 0, len(mapping.Pairs))

	for _, section := range sections {
		// Extract pairs for this section
		sectionPairs := section.pairs

		// Sort this section's pairs
		if !section.isMarker && len(sectionPairs) > 0 {
			tempMapping := &node.MappingNode{
				BaseNode: mapping.BaseNode,
				Pairs:    sectionPairs,
				Style:    mapping.Style,
			}

			sorted := ss.sorter.sortMapping(tempMapping)
			sortedPairs = append(sortedPairs, sorted.Pairs...)
		} else {
			// Keep marker pairs as-is
			sortedPairs = append(sortedPairs, sectionPairs...)
		}
	}

	// Create new mapping with sorted sections
	return &node.MappingNode{
		BaseNode: mapping.BaseNode,
		Pairs:    sortedPairs,
		Style:    mapping.Style,
	}
}

// sortSequenceWithSections sorts a sequence node respecting section boundaries
func (ss *SectionSorter) sortSequenceWithSections(seq *node.SequenceNode) *node.SequenceNode {
	// For sequences, detect sections based on comments
	sections := ss.detectSequenceSections(seq)

	if len(sections) == 0 {
		return ss.sorter.sortSequence(seq)
	}

	// Sort each section independently
	sortedItems := make([]node.Node, 0, len(seq.Items))

	for _, section := range sections {
		if !section.isMarker && len(section.items) > 0 {
			tempSeq := &node.SequenceNode{
				BaseNode: seq.BaseNode,
				Items:    section.items,
				Style:    seq.Style,
			}

			sorted := ss.sorter.sortSequence(tempSeq)
			sortedItems = append(sortedItems, sorted.Items...)
		} else {
			sortedItems = append(sortedItems, section.items...)
		}
	}

	return &node.SequenceNode{
		BaseNode: seq.BaseNode,
		Items:    sortedItems,
		Style:    seq.Style,
	}
}

// section represents a section in the document
type section struct {
	startIndex int
	endIndex   int
	isMarker   bool
	pairs      []*node.MappingPair
	items      []node.Node
}

// detectSections detects section boundaries in a mapping
func (ss *SectionSorter) detectSections(mapping *node.MappingNode) []section {
	// First, transfer inter-field comments to fix comment association
	// Comments between key-value pairs are often attached to the previous value
	// instead of the next key where they logically belong
	pairs := make([]*node.MappingPair, len(mapping.Pairs))
	copy(pairs, mapping.Pairs)
	ss.transferInterFieldComments(pairs)

	var sections []section
	currentSection := section{
		startIndex: 0,
		pairs:      []*node.MappingPair{},
	}

	for i, pair := range pairs {
		// Check if this pair's comment marks a section
		if ss.isSectionMarker(pair.Key) {
			// Save current section if it has pairs
			if len(currentSection.pairs) > 0 {
				sections = append(sections, currentSection)
			}

			// Don't include the marker pair in any section - just start a new section
			currentSection = section{
				startIndex: i + 1,
				pairs:      []*node.MappingPair{},
			}
		} else {
			currentSection.pairs = append(currentSection.pairs, pair)
			currentSection.endIndex = i
		}
	}

	// Add final section if it has pairs
	if len(currentSection.pairs) > 0 {
		sections = append(sections, currentSection)
	}

	return sections
}

// detectSequenceSections detects section boundaries in a sequence
func (ss *SectionSorter) detectSequenceSections(seq *node.SequenceNode) []section {
	var sections []section
	currentSection := section{
		startIndex: 0,
		items:      []node.Node{},
	}

	for i, item := range seq.Items {
		// Check if this item's comment marks a section
		if ss.isSectionMarker(item) {
			// Save current section if it has items
			if len(currentSection.items) > 0 {
				sections = append(sections, currentSection)
			}

			// Add marker as its own section
			sections = append(sections, section{
				startIndex: i,
				endIndex:   i,
				isMarker:   true,
				items:      []node.Node{item},
			})

			// Start new section
			currentSection = section{
				startIndex: i + 1,
				items:      []node.Node{},
			}
		} else {
			currentSection.items = append(currentSection.items, item)
			currentSection.endIndex = i
		}
	}

	// Add final section if it has items
	if len(currentSection.items) > 0 {
		sections = append(sections, currentSection)
	}

	return sections
}

// transferInterFieldComments moves comments from value nodes to the next key node
// This fixes the parser's comment association where comments between pairs
// get attached to the previous value instead of the next key
func (ss *SectionSorter) transferInterFieldComments(pairs []*node.MappingPair) {
	for i := 0; i < len(pairs)-1; i++ {
		// Check if the value has a head comment that might belong to the next key
		if scalar, ok := pairs[i].Value.(*node.ScalarNode); ok {
			if scalar.HeadComment != nil && len(scalar.HeadComment.Comments) > 0 {
				// Check if any comment contains section markers
				for _, comment := range scalar.HeadComment.Comments {
					for _, marker := range ss.config.SectionMarkers {
						if strings.Contains(comment, marker) {
							// This comment should belong to the next key
							if nextKey, ok := pairs[i+1].Key.(*node.ScalarNode); ok {
								if nextKey.HeadComment == nil {
									nextKey.HeadComment = scalar.HeadComment
									scalar.HeadComment = nil
								}
							}
							break
						}
					}
				}
			}
		}
	}
}

// isSectionMarker checks if a node's comment marks a section boundary
func (ss *SectionSorter) isSectionMarker(n node.Node) bool {
	if n == nil {
		return false
	}

	// Check head comment based on node type
	var headComment string
	switch node := n.(type) {
	case *node.ScalarNode:
		if node.HeadComment != nil {
			for _, comment := range node.HeadComment.Comments {
				headComment += comment + " "
			}
		}
	case *node.MappingNode:
		if node.HeadComment != nil {
			for _, comment := range node.HeadComment.Comments {
				headComment += comment + " "
			}
		}
	case *node.SequenceNode:
		if node.HeadComment != nil {
			for _, comment := range node.HeadComment.Comments {
				headComment += comment + " "
			}
		}
	}

	if headComment != "" {
		for _, marker := range ss.config.SectionMarkers {
			if strings.Contains(headComment, marker) {
				return true
			}
		}
	}

	// Check if the node itself is a comment-like marker
	if scalar, ok := n.(*node.ScalarNode); ok {
		for _, marker := range ss.config.SectionMarkers {
			if strings.HasPrefix(scalar.Value, marker) ||
				strings.Contains(scalar.Value, "---") ||
				strings.Contains(scalar.Value, "===") {
				return true
			}
		}
	}

	return false
}
