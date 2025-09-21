package node

import (
	"regexp"
	"strings"
)

// CommentPosition indicates where a comment appears relative to a node
type CommentPosition int

const (
	CommentPositionAbove CommentPosition = iota
	CommentPositionInline
	CommentPositionBelow
	CommentPositionKey
	CommentPositionValue
	CommentPositionSection  // Comments at section level
	CommentPositionAny     // Matches any position (for rules)
)

// CommentStyle represents different comment formatting styles
type CommentStyle int

const (
	CommentStyleHash CommentStyle = iota // # comment
	CommentStyleDoubleHash                // ## comment
	CommentStylePlain                     // plain comment
)

// CommentRule defines rules for comment formatting and blank line insertion
type CommentRule struct {
	Pattern        *regexp.Regexp
	BlankLines     int
	Position       CommentPosition
	Description    string
	SectionType    SectionType // Apply rule only to specific section types
	ApplyToAll     bool        // Apply to all sections regardless of type
}

// CommentProcessor manages comment formatting, section detection, and blank line rules
type CommentProcessor struct {
	rules              []CommentRule
	preserveBlankLines bool
	maxBlankLines      int
	autoDetectSections bool
	sectionDetectors   []SectionDetector
}

// SectionDetector helps identify sections based on comment patterns
type SectionDetector struct {
	Pattern     *regexp.Regexp
	SectionType SectionType
	Description string
	Priority    int // Higher priority detectors are checked first
}

// NewCommentProcessor creates a new comment processor with default settings
func NewCommentProcessor() *CommentProcessor {
	cm := &CommentProcessor{
		rules:              make([]CommentRule, 0),
		preserveBlankLines: true,
		maxBlankLines:      2,
		autoDetectSections: true,
		sectionDetectors:   make([]SectionDetector, 0),
	}

	// Add default section detectors
	cm.addDefaultSectionDetectors()
	cm.addDefaultCommentRules()

	return cm
}

// addDefaultSectionDetectors adds common patterns for detecting YAML sections
func (cp *CommentProcessor) addDefaultSectionDetectors() {
	// Header section: comments that look like titles
	cp.AddSectionDetector(`^#\s*[A-Z][^-\n]*$`, SectionTypeHeader, "Header section", 10)

	// Configuration section: comments mentioning config/settings
	cp.AddSectionDetector(`(?i)#.*\b(config|configuration|settings|options)\b`, SectionTypeConfiguration, "Configuration section", 8)

	// Data section: comments mentioning data/content
	cp.AddSectionDetector(`(?i)#.*\b(data|content|values|items)\b`, SectionTypeData, "Data section", 6)

	// Footer section: comments at the end or mentioning footer
	cp.AddSectionDetector(`(?i)#.*\b(footer|end|final)\b`, SectionTypeFooter, "Footer section", 4)
}

// addDefaultCommentRules adds common formatting rules
func (cp *CommentProcessor) addDefaultCommentRules() {
	// Header sections get extra spacing
	cp.AddCommentRule(`^#\s*[A-Z][^-\n]*$`, 1, CommentPositionSection, "Header comment spacing", SectionTypeHeader, false)

	// Documentation-style comments (starting with # --)
	cp.AddCommentRule(`^#\s*--`, 0, CommentPositionAbove, "Documentation comment", SectionTypeGeneric, true)

	// Inline comments get no extra spacing
	cp.AddCommentRule(`.*`, 0, CommentPositionInline, "Inline comment", SectionTypeGeneric, true)
}

// AddCommentRule adds a comment formatting rule
func (cp *CommentProcessor) AddCommentRule(pattern string, blankLines int, pos CommentPosition, desc string, sectionType SectionType, applyToAll bool) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	cp.rules = append(cp.rules, CommentRule{
		Pattern:     re,
		BlankLines:  blankLines,
		Position:    pos,
		Description: desc,
		SectionType: sectionType,
		ApplyToAll:  applyToAll,
	})
	return nil
}

// AddSectionDetector adds a section detection pattern
func (cp *CommentProcessor) AddSectionDetector(pattern string, sectionType SectionType, desc string, priority int) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	detector := SectionDetector{
		Pattern:     re,
		SectionType: sectionType,
		Description: desc,
		Priority:    priority,
	}

	// Insert in priority order (highest first)
	inserted := false
	for i, existing := range cp.sectionDetectors {
		if priority > existing.Priority {
			cp.sectionDetectors = append(cp.sectionDetectors[:i], append([]SectionDetector{detector}, cp.sectionDetectors[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		cp.sectionDetectors = append(cp.sectionDetectors, detector)
	}

	return nil
}

// DetectSectionType analyzes a comment to determine what type of section it might introduce
func (cp *CommentProcessor) DetectSectionType(comment string) SectionType {
	if !cp.autoDetectSections {
		return SectionTypeGeneric
	}

	comment = strings.TrimSpace(comment)
	for _, detector := range cp.sectionDetectors {
		if detector.Pattern.MatchString(comment) {
			return detector.SectionType
		}
	}

	return SectionTypeGeneric
}

// GetBlankLinesForComment determines how many blank lines should precede a comment
func (cp *CommentProcessor) GetBlankLinesForComment(comment string, pos CommentPosition, sectionType SectionType) int {
	comment = strings.TrimSpace(comment)

	for _, rule := range cp.rules {
		// Check if rule applies to this position
		if rule.Position != pos && rule.Position != CommentPositionSection {
			continue
		}

		// Check if rule applies to this section type
		if !rule.ApplyToAll && rule.SectionType != sectionType {
			continue
		}

		if rule.Pattern.MatchString(comment) {
			return rule.BlankLines
		}
	}

	// Default: no extra blank lines
	return 0
}

// FormatCommentGroup formats a comment group with appropriate spacing
func (cp *CommentProcessor) FormatCommentGroup(cg *CommentGroup, pos CommentPosition, sectionType SectionType) []string {
	if cg == nil || len(cg.Comments) == 0 {
		return nil
	}

	var result []string

	// Add blank lines before if specified in the group
	for i := 0; i < cg.BlankLinesBefore; i++ {
		result = append(result, "")
	}

	// Process each comment
	for i, comment := range cg.Comments {
		// Check if rule applies for additional spacing
		if i == 0 && pos != CommentPositionInline {
			extraLines := cp.GetBlankLinesForComment(comment, pos, sectionType)
			for j := 0; j < extraLines; j++ {
				result = append(result, "")
			}
		}

		// Apply comment formatting - use hash style by default
		if cg.Format.PreserveSpacing {
			result = append(result, comment)
		} else {
			result = append(result, cp.formatSingleComment(comment, CommentStyleHash))
		}
	}

	// Add blank lines after if specified
	for i := 0; i < cg.BlankLinesAfter; i++ {
		result = append(result, "")
	}

	return result
}

// formatSingleComment applies formatting to a single comment
func (cp *CommentProcessor) formatSingleComment(comment string, style CommentStyle) string {
	// Trim the comment
	trimmed := strings.TrimSpace(comment)

	// Apply the style
	switch style {
	case CommentStyleHash:
		if !strings.HasPrefix(trimmed, "#") {
			trimmed = "# " + trimmed
		}
	case CommentStyleDoubleHash:
		if !strings.HasPrefix(trimmed, "##") {
			trimmed = "## " + trimmed
		}
	case CommentStylePlain:
		// Remove any hash prefixes
		trimmed = strings.TrimPrefix(trimmed, "##")
		trimmed = strings.TrimPrefix(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)
	}

	return trimmed
}

// AssociateCommentToNode associates a comment with a node based on position
func (cp *CommentProcessor) AssociateCommentToNode(node Node, comment string, pos CommentPosition, blankLinesBefore int) {
	// Get base node to set comment
	var baseNode *BaseNode
	switch n := node.(type) {
	case *ScalarNode:
		baseNode = &n.BaseNode
	case *SequenceNode:
		baseNode = &n.BaseNode
	case *MappingNode:
		baseNode = &n.BaseNode
	case *SectionNode:
		baseNode = &n.BaseNode
	case *DocumentNode:
		baseNode = &n.BaseNode
	default:
		return
	}

	// Get or create the appropriate comment group
	var cg *CommentGroup
	switch pos {
	case CommentPositionAbove, CommentPositionSection:
		cg = baseNode.HeadComment
	case CommentPositionInline:
		cg = baseNode.LineComment
	case CommentPositionBelow:
		cg = baseNode.FootComment
	}

	// Create default format
	defaultFormat := CommentFormat{
		IndentLevel:     0,
		PreserveSpacing: cp.preserveBlankLines,
		GroupRelated:    true,
	}

	// If comment group exists, append to it; otherwise create new
	if cg != nil {
		cg.Comments = append(cg.Comments, comment)
		// Update blank lines if this comment has more
		if blankLinesBefore > cg.BlankLinesBefore {
			cg.BlankLinesBefore = blankLinesBefore
		}
	} else {
		cg = &CommentGroup{
			Comments:         []string{comment},
			BlankLinesBefore: blankLinesBefore,
			BlankLinesAfter:  0,
			Format:           defaultFormat,
		}
		// Set the comment group
		switch pos {
		case CommentPositionAbove, CommentPositionSection:
			baseNode.HeadComment = cg
		case CommentPositionInline:
			baseNode.LineComment = cg
		case CommentPositionBelow:
			baseNode.FootComment = cg
		}
	}
}

// AssociateCommentToSection associates a comment directly with a section
func (cp *CommentProcessor) AssociateCommentToSection(section *Section, comment string, blankLinesBefore int) {
	if section == nil {
		return
	}

	defaultFormat := CommentFormat{
		IndentLevel:     0,
		PreserveSpacing: cp.preserveBlankLines,
		GroupRelated:    true,
	}

	if section.Comments != nil {
		section.Comments.Comments = append(section.Comments.Comments, comment)
		if blankLinesBefore > section.Comments.BlankLinesBefore {
			section.Comments.BlankLinesBefore = blankLinesBefore
		}
	} else {
		section.Comments = &CommentGroup{
			Comments:         []string{comment},
			BlankLinesBefore: blankLinesBefore,
			BlankLinesAfter:  0,
			Format:           defaultFormat,
		}
	}
}

// AssociateCommentToPair associates a comment with a mapping pair (key or value)
func (cp *CommentProcessor) AssociateCommentToPair(pair *MappingPair, comment string, pos CommentPosition, blankLinesBefore int) {
	if pair == nil {
		return
	}

	defaultFormat := CommentFormat{
		IndentLevel:     0,
		PreserveSpacing: cp.preserveBlankLines,
		GroupRelated:    true,
	}

	var cg *CommentGroup
	switch pos {
	case CommentPositionKey:
		cg = pair.KeyComment
	case CommentPositionValue:
		cg = pair.ValueComment
	default:
		return
	}

	if cg != nil {
		cg.Comments = append(cg.Comments, comment)
		if blankLinesBefore > cg.BlankLinesBefore {
			cg.BlankLinesBefore = blankLinesBefore
		}
	} else {
		cg = &CommentGroup{
			Comments:         []string{comment},
			BlankLinesBefore: blankLinesBefore,
			BlankLinesAfter:  0,
			Format:           defaultFormat,
		}

		switch pos {
		case CommentPositionKey:
			pair.KeyComment = cg
		case CommentPositionValue:
			pair.ValueComment = cg
		}
	}
}

// MergeCommentGroups merges multiple comment groups intelligently
func (cp *CommentProcessor) MergeCommentGroups(groups ...*CommentGroup) *CommentGroup {
	var allComments []string
	maxBlankLinesBefore := 0
	maxBlankLinesAfter := 0
	var mergedFormat CommentFormat

	for i, g := range groups {
		if g != nil {
			allComments = append(allComments, g.Comments...)
			if g.BlankLinesBefore > maxBlankLinesBefore {
				maxBlankLinesBefore = g.BlankLinesBefore
			}
			if g.BlankLinesAfter > maxBlankLinesAfter {
				maxBlankLinesAfter = g.BlankLinesAfter
			}

			// Use format from first non-empty group
			if i == 0 {
				mergedFormat = g.Format
			}
		}
	}

	if len(allComments) == 0 {
		return nil
	}

	return &CommentGroup{
		Comments:         allComments,
		BlankLinesBefore: maxBlankLinesBefore,
		BlankLinesAfter:  maxBlankLinesAfter,
		Format:           mergedFormat,
	}
}

// CreateSectionFromComments analyzes comments to automatically create sections
func (cp *CommentProcessor) CreateSectionFromComments(comments []string, sectionBuilder SectionBuilder) *Section {
	if len(comments) == 0 {
		return nil
	}

	// Detect section type from first comment
	sectionType := cp.DetectSectionType(comments[0])

	// Generate section ID from first comment
	sectionID := cp.generateSectionID(comments[0])

	// Extract title from first comment if it looks like a header
	title := cp.extractTitle(comments[0])

	// Create section
	section := sectionBuilder.CreateSection(sectionID, sectionType)
	if title != "" {
		section = sectionBuilder.WithTitle(section, title)
	}

	// Associate comments with section
	for _, comment := range comments {
		cp.AssociateCommentToSection(section, comment, 0)
	}

	return section
}

// generateSectionID creates a unique ID from a comment
func (cp *CommentProcessor) generateSectionID(comment string) string {
	// Remove # and extra spaces, convert to lowercase
	id := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comment), "#"))
	id = strings.ToLower(id)

	// Replace spaces and special characters with underscores
	re := regexp.MustCompile(`[^a-z0-9]+`)
	id = re.ReplaceAllString(id, "_")

	// Remove leading/trailing underscores
	id = strings.Trim(id, "_")

	if id == "" {
		id = "section"
	}

	return id
}

// extractTitle extracts a human-readable title from a comment
func (cp *CommentProcessor) extractTitle(comment string) string {
	// Remove # and -- prefixes, trim spaces
	title := strings.TrimSpace(comment)
	title = strings.TrimPrefix(title, "#")
	title = strings.TrimSpace(title)
	title = strings.TrimPrefix(title, "--")
	title = strings.TrimSpace(title)

	return title
}

// SetPreserveBlankLines configures whether to preserve original blank line formatting
func (cp *CommentProcessor) SetPreserveBlankLines(preserve bool) {
	cp.preserveBlankLines = preserve
}

// SetMaxBlankLines sets the maximum number of consecutive blank lines allowed
func (cp *CommentProcessor) SetMaxBlankLines(max int) {
	cp.maxBlankLines = max
}

// SetAutoDetectSections configures whether to automatically detect sections from comments
func (cp *CommentProcessor) SetAutoDetectSections(autoDetect bool) {
	cp.autoDetectSections = autoDetect
}

// ProcessComment formats a comment based on position and section type
func (cp *CommentProcessor) ProcessComment(text string, pos CommentPosition, sectionType SectionType) string {
	// Format the comment with appropriate markers
	formatted := cp.formatSingleComment(text, CommentStyleHash)

	// Return the formatted comment
	return formatted
}