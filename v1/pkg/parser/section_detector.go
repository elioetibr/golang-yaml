package parser

import (
	"regexp"
	"strings"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// SectionDetector identifies and manages document sections
type SectionDetector struct {
	boundaries      []int
	sections        []*DetectedSection
	patterns        []SectionPattern
	minEmptyLines   int
}

// DetectedSection represents a detected section in the document
type DetectedSection struct {
	StartLine   int
	EndLine     int
	Type        node.SectionType
	Title       string
	ID          string
	IndentLevel int
	Comments    []string
}

// SectionPattern defines a pattern for detecting sections
type SectionPattern struct {
	Pattern     *regexp.Regexp
	Type        node.SectionType
	Priority    int
	Description string
}

// NewSectionDetector creates a new section detector
func NewSectionDetector() *SectionDetector {
	sd := &SectionDetector{
		boundaries:    make([]int, 0),
		sections:      make([]*DetectedSection, 0),
		patterns:      make([]SectionPattern, 0),
		minEmptyLines: 2,
	}

	// Add default patterns
	sd.addDefaultPatterns()

	return sd
}

// addDefaultPatterns adds common section detection patterns
func (sd *SectionDetector) addDefaultPatterns() {
	// Schema section pattern
	sd.AddPattern(`(?i)@schema`, node.SectionTypeConfiguration, 10, "Schema section")

	// Configuration section patterns
	sd.AddPattern(`(?i)config(uration)?`, node.SectionTypeConfiguration, 8, "Configuration section")
	sd.AddPattern(`(?i)settings?`, node.SectionTypeConfiguration, 8, "Settings section")
	sd.AddPattern(`(?i)options?`, node.SectionTypeConfiguration, 8, "Options section")

	// Header section patterns
	sd.AddPattern(`^#\s*[A-Z][A-Z\s]*$`, node.SectionTypeHeader, 9, "Header section")
	sd.AddPattern(`(?i)^#\s*(introduction|overview|description)`, node.SectionTypeHeader, 9, "Header section")

	// Data section patterns
	sd.AddPattern(`(?i)data`, node.SectionTypeData, 7, "Data section")
	sd.AddPattern(`(?i)values?`, node.SectionTypeData, 7, "Values section")
	sd.AddPattern(`(?i)items?`, node.SectionTypeData, 7, "Items section")

	// Footer section patterns
	sd.AddPattern(`(?i)footer`, node.SectionTypeFooter, 6, "Footer section")
	sd.AddPattern(`(?i)examples?`, node.SectionTypeFooter, 6, "Examples section")
	sd.AddPattern(`(?i)notes?`, node.SectionTypeFooter, 6, "Notes section")
}

// AddPattern adds a section detection pattern
func (sd *SectionDetector) AddPattern(pattern string, sectionType node.SectionType, priority int, description string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	sp := SectionPattern{
		Pattern:     re,
		Type:        sectionType,
		Priority:    priority,
		Description: description,
	}

	// Insert in priority order
	inserted := false
	for i, existing := range sd.patterns {
		if priority > existing.Priority {
			sd.patterns = append(sd.patterns[:i], append([]SectionPattern{sp}, sd.patterns[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		sd.patterns = append(sd.patterns, sp)
	}

	return nil
}

// MarkSectionBoundary marks a line as a section boundary
func (sd *SectionDetector) MarkSectionBoundary(line int) {
	if !sd.isBoundary(line) {
		sd.boundaries = append(sd.boundaries, line)
	}
}

// DetectSectionFromComments analyzes comments to detect a section
func (sd *SectionDetector) DetectSectionFromComments(comments []string, startLine int, indentLevel int) *DetectedSection {
	if len(comments) == 0 {
		return nil
	}

	// Check each comment against patterns
	for _, comment := range comments {
		cleanComment := sd.cleanComment(comment)

		for _, pattern := range sd.patterns {
			if pattern.Pattern.MatchString(cleanComment) {
				return &DetectedSection{
					StartLine:   startLine,
					Type:        pattern.Type,
					Title:       sd.extractTitle(cleanComment),
					ID:          sd.generateID(cleanComment),
					IndentLevel: indentLevel,
					Comments:    comments,
				}
			}
		}
	}

	// Default section if no pattern matches but has section characteristics
	if sd.looksLikeSection(comments, startLine) {
		return &DetectedSection{
			StartLine:   startLine,
			Type:        node.SectionTypeGeneric,
			Title:       sd.extractTitle(comments[0]),
			ID:          sd.generateID(comments[0]),
			IndentLevel: indentLevel,
			Comments:    comments,
		}
	}

	return nil
}

// looksLikeSection checks if comments look like a section header
func (sd *SectionDetector) looksLikeSection(comments []string, line int) bool {
	// Check if this is at a section boundary
	if sd.isBoundary(line) {
		return true
	}

	// Check if first comment looks like a title
	if len(comments) > 0 {
		first := sd.cleanComment(comments[0])
		// Title-like: starts with capital, no trailing punctuation
		if len(first) > 0 {
			firstChar := first[0]
			lastChar := first[len(first)-1]
			if firstChar >= 'A' && firstChar <= 'Z' &&
				lastChar != '.' && lastChar != ',' && lastChar != ';' {
				return true
			}
		}
	}

	return false
}

// FindSectionForLine finds which section a line belongs to
func (sd *SectionDetector) FindSectionForLine(line int) *DetectedSection {
	for _, section := range sd.sections {
		if line >= section.StartLine && (section.EndLine == 0 || line <= section.EndLine) {
			return section
		}
	}
	return nil
}

// AddSection adds a detected section
func (sd *SectionDetector) AddSection(section *DetectedSection) {
	// Update end lines of previous sections
	if len(sd.sections) > 0 {
		lastSection := sd.sections[len(sd.sections)-1]
		if lastSection.EndLine == 0 {
			lastSection.EndLine = section.StartLine - 1
		}
	}

	sd.sections = append(sd.sections, section)
}

// GetSections returns all detected sections
func (sd *SectionDetector) GetSections() []*DetectedSection {
	return sd.sections
}

// GetBoundaries returns all section boundaries
func (sd *SectionDetector) GetBoundaries() []int {
	return sd.boundaries
}

// cleanComment removes comment markers and trims whitespace
func (sd *SectionDetector) cleanComment(comment string) string {
	cleaned := strings.TrimSpace(comment)
	cleaned = strings.TrimPrefix(cleaned, "#")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimPrefix(cleaned, "--")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// extractTitle extracts a title from a comment
func (sd *SectionDetector) extractTitle(comment string) string {
	title := sd.cleanComment(comment)

	// Remove common prefixes
	prefixes := []string{
		"Section:", "Chapter:", "Part:",
		"Config:", "Settings:", "Options:",
		"Data:", "Values:", "Items:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(title, prefix) {
			title = strings.TrimPrefix(title, prefix)
			title = strings.TrimSpace(title)
			break
		}
	}

	return title
}

// generateID generates a section ID from text
func (sd *SectionDetector) generateID(text string) string {
	id := sd.cleanComment(text)
	id = strings.ToLower(id)

	// Replace non-alphanumeric with underscores
	re := regexp.MustCompile(`[^a-z0-9]+`)
	id = re.ReplaceAllString(id, "_")

	// Remove leading/trailing underscores
	id = strings.Trim(id, "_")

	if id == "" {
		id = "section"
	}

	// Ensure uniqueness
	baseID := id
	counter := 1
	for sd.idExists(id) {
		id = baseID + "_" + string(counter)
		counter++
	}

	return id
}

// idExists checks if a section ID already exists
func (sd *SectionDetector) idExists(id string) bool {
	for _, section := range sd.sections {
		if section.ID == id {
			return true
		}
	}
	return false
}

// isBoundary checks if a line is marked as a boundary
func (sd *SectionDetector) isBoundary(line int) bool {
	for _, boundary := range sd.boundaries {
		if boundary == line {
			return true
		}
	}
	return false
}

// SectionMarkingVisitor visits nodes to mark sections
type SectionMarkingVisitor struct {
	detector *SectionDetector
	options  *ParserOptions
	currentSection *DetectedSection
}

// Visit implements the Visitor interface for section marking
func (v *SectionMarkingVisitor) Visit(n node.Node) error {
	// Check if node has comments that indicate a section
	var comments []string
	var startLine int

	// Extract comments from node
	switch node := n.(type) {
	case *node.ScalarNode:
		if node.HeadComment != nil {
			comments = node.HeadComment.Comments
			startLine = node.Line() - len(comments)
		}
	case *node.MappingNode:
		if node.HeadComment != nil {
			comments = node.HeadComment.Comments
			startLine = node.Line() - len(comments)
		}
	case *node.SequenceNode:
		if node.HeadComment != nil {
			comments = node.HeadComment.Comments
			startLine = node.Line() - len(comments)
		}
	}

	// Detect section from comments
	if len(comments) > 0 {
		if section := v.detector.DetectSectionFromComments(comments, startLine, n.Column()); section != nil {
			// Create and set section on node
			nodeSection := &node.Section{
				ID:          section.ID,
				Type:        section.Type,
				Title:       section.Title,
				StartLine:   section.StartLine,
				EndLine:     section.EndLine,
				IndentLevel: section.IndentLevel,
			}
			n.SetSection(nodeSection)

			// Track current section
			v.currentSection = section
			v.detector.AddSection(section)
		}
	}

	// If no section detected but inside a section, inherit it
	if n.Section() == nil && v.currentSection != nil {
		if n.Line() >= v.currentSection.StartLine {
			nodeSection := &node.Section{
				ID:          v.currentSection.ID,
				Type:        v.currentSection.Type,
				Title:       v.currentSection.Title,
				StartLine:   v.currentSection.StartLine,
				EndLine:     v.currentSection.EndLine,
				IndentLevel: v.currentSection.IndentLevel,
			}
			n.SetSection(nodeSection)
		}
	}

	return nil
}

// SectionBuilder helps build section structures
type SectionBuilder interface {
	CreateSection(id string, sectionType node.SectionType) *node.Section
	WithTitle(section *node.Section, title string) *node.Section
}