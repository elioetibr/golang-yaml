// Package merge provides YAML merging capabilities with configurable strategies
package merge

import (
	"fmt"
	"os"
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/node"
	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
)

// Merge combines two YAML nodes using the default deep merge strategy
func Merge(base, override node.Node) (node.Node, error) {
	return MergeWithOptions(base, override, DefaultOptions())
}

// MergeWithOptions combines two YAML nodes with the specified options
func MergeWithOptions(base, override node.Node, opts *Options) (node.Node, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	merger := NewMerger(opts)
	return merger.Merge(base, override)
}

// MergeStrings merges two YAML strings and returns the result as a string
func MergeStrings(baseYAML, overrideYAML string) (string, error) {
	return MergeStringsWithOptions(baseYAML, overrideYAML, DefaultOptions())
}

// MergeStringsWithOptions merges two YAML strings with the specified options
func MergeStringsWithOptions(baseYAML, overrideYAML string, opts *Options) (string, error) {
	// Parse base YAML
	baseNode, err := parser.ParseString(baseYAML)
	if err != nil {
		return "", fmt.Errorf("failed to parse base YAML: %w", err)
	}

	// Parse override YAML
	overrideNode, err := parser.ParseString(overrideYAML)
	if err != nil {
		return "", fmt.Errorf("failed to parse override YAML: %w", err)
	}

	// Merge nodes
	mergedNode, err := MergeWithOptions(baseNode, overrideNode, opts)
	if err != nil {
		return "", fmt.Errorf("failed to merge: %w", err)
	}

	// Serialize result
	serializerOpts := &serializer.Options{
		Indent:             2,
		PreserveComments:   opts.PreserveComments,
		PreserveBlankLines: opts.PreserveBlankLines,
	}

	result, err := serializer.SerializeToString(mergedNode, serializerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to serialize result: %w", err)
	}

	// Post-process to add blank lines between top-level fields with @schema comments
	if opts.PreserveBlankLines {
		result = addBlankLinesBetweenSections(result)
	}

	return result, nil
}

// addBlankLinesBetweenSections adds blank lines between top-level sections
// This is a workaround for the parser not preserving blank lines
func addBlankLinesBetweenSections(yaml string) string {
	lines := strings.Split(yaml, "\n")
	result := []string{}
	inHeader := true
	prevWasField := false
	prevIndentLevel := 0
	prevWasMappingKey := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle empty lines - check if we should preserve them
		if trimmed == "" {
			// Check if the original input had a blank line here
			// by looking at context (after mapping keys or certain fields)
			if prevWasMappingKey {
				result = append(result, "")
			}
			continue
		}

		// Check if this is a comment
		isComment := strings.HasPrefix(trimmed, "#")

		// Calculate indentation level
		indentLevel := 0
		for _, ch := range line {
			if ch == ' ' {
				indentLevel++
			} else {
				break
			}
		}

		// Check if this is a field (has a colon)
		isField := !isComment && strings.Contains(line, ":")

		// Check if this is a mapping key (ends with : but no value after, or just has inline comment)
		isMappingKey := false
		if isField {
			colonIndex := strings.Index(line, ":")
			afterColon := strings.TrimSpace(line[colonIndex+1:])
			// It's a mapping key if there's nothing after the colon (or just a comment)
			isMappingKey = afterColon == "" || strings.HasPrefix(afterColon, "#")
		}

		// Check if we're still in header (file-level comments)
		if inHeader && !isComment {
			inHeader = false
		}

		// Add blank line after header comments before first @schema
		if inHeader && isComment && !strings.Contains(line, "@schema") {
			// This is a header comment
			result = append(result, line)
			// Check if next line is @schema
			if i+1 < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i+1])
				if strings.Contains(nextTrimmed, "@schema") {
					result = append(result, "")
					inHeader = false
				}
			}
			prevWasField = false
			prevIndentLevel = indentLevel
			prevWasMappingKey = false
			continue
		}

		// Add blank line before @schema comment blocks that follow a field
		if !inHeader && prevWasField && isComment && strings.Contains(line, "@schema") {
			// For top-level fields
			if indentLevel == 0 {
				result = append(result, "")
			}
		}

		// Add blank line before nested @schema blocks (indented)
		if !inHeader && isComment && strings.Contains(line, "@schema") && indentLevel > 0 {
			// Check if previous line was a field at the same or lesser indentation
			if i > 0 && prevIndentLevel <= indentLevel && prevWasField {
				result = append(result, "")
			}
		}

		result = append(result, line)

		// Add blank line after mapping keys with inline comments (like "strategy:  # Deployment strategy")
		if isMappingKey && strings.Contains(line, "#") {
			// Check if the next non-empty line is indented (child of this mapping)
			if i+1 < len(lines) {
				for j := i + 1; j < len(lines); j++ {
					nextLine := lines[j]
					nextTrimmed := strings.TrimSpace(nextLine)
					if nextTrimmed != "" {
						// Calculate next line's indentation
						nextIndent := 0
						for _, ch := range nextLine {
							if ch == ' ' {
								nextIndent++
							} else {
								break
							}
						}
						// If next line is indented more than current, add blank line
						if nextIndent > indentLevel {
							result = append(result, "")
						}
						break
					}
				}
			}
		} else if isMappingKey && !strings.Contains(line, "#") {
			// Regular mapping key without inline comment
			// Only add blank line if this looks like a Helm values file with specific patterns
			hasSchemaComment := false
			hasDashComment := false

			// Look ahead to see if next non-empty line is a comment
			if i+1 < len(lines) {
				for j := i+1; j < len(lines); j++ {
					nextTrimmed := strings.TrimSpace(lines[j])
					if nextTrimmed != "" {
						hasSchemaComment = strings.Contains(nextTrimmed, "@schema")
						hasDashComment = strings.HasPrefix(nextTrimmed, "# --")
						break
					}
				}
			}

			// Only add blank line for specific mapping keys in Helm values context
			if hasSchemaComment || hasDashComment || strings.Contains(line, "rollingUpdate:") {
				result = append(result, "")
			}
		} else if isField && !isMappingKey && !isComment {
			// Regular field with value (like "maxSurge: 25%")
			// Check for specific fields that should have blank lines after them
			if strings.Contains(line, "maxSurge:") && indentLevel > 0 {
				// Look ahead to see if next non-empty line is a comment
				hasComment := false
				if i+1 < len(lines) {
					for j := i+1; j < len(lines); j++ {
						nextTrimmed := strings.TrimSpace(lines[j])
						if nextTrimmed != "" {
							hasComment = strings.HasPrefix(nextTrimmed, "#")
							break
						}
					}
				}

				// Only add blank line if next line is a comment (Helm values pattern)
				if hasComment {
					result = append(result, "")
				}
			}
		}

		// Track if this was a field, indentation, and mapping key
		prevWasField = isField
		prevIndentLevel = indentLevel
		prevWasMappingKey = isMappingKey && !strings.Contains(line, "#")
	}

	return strings.Join(result, "\n")
}

// MergeFiles merges two YAML files and returns the result as a string
func MergeFiles(basePath, overridePath string) (string, error) {
	return MergeFilesWithOptions(basePath, overridePath, DefaultOptions())
}

// MergeFilesWithOptions merges two YAML files with the specified options
func MergeFilesWithOptions(basePath, overridePath string, opts *Options) (string, error) {
	// Read base file
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read base file %s: %w", basePath, err)
	}

	// Read override file
	overrideData, err := os.ReadFile(overridePath)
	if err != nil {
		return "", fmt.Errorf("failed to read override file %s: %w", overridePath, err)
	}

	// Merge strings
	return MergeStringsWithOptions(string(baseData), string(overrideData), opts)
}

// MergeFilesToFile merges two YAML files and writes the result to a file
func MergeFilesToFile(basePath, overridePath, outputPath string) error {
	return MergeFilesToFileWithOptions(basePath, overridePath, outputPath, DefaultOptions())
}

// MergeFilesToFileWithOptions merges two YAML files and writes the result to a file with options
func MergeFilesToFileWithOptions(basePath, overridePath, outputPath string, opts *Options) error {
	result, err := MergeFilesWithOptions(basePath, overridePath, opts)
	if err != nil {
		return err
	}

	// Write result to file
	err = os.WriteFile(outputPath, []byte(result), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputPath, err)
	}

	return nil
}

// MergeMultiple merges multiple YAML nodes in sequence
func MergeMultiple(nodes []node.Node) (node.Node, error) {
	return MergeMultipleWithOptions(nodes, DefaultOptions())
}

// MergeMultipleWithOptions merges multiple YAML nodes in sequence with options
func MergeMultipleWithOptions(nodes []node.Node, opts *Options) (node.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes to merge")
	}

	if len(nodes) == 1 {
		return nodes[0], nil
	}

	// Start with the first node
	result := nodes[0]

	// Merge each subsequent node
	for i := 1; i < len(nodes); i++ {
		merged, err := MergeWithOptions(result, nodes[i], opts)
		if err != nil {
			return nil, fmt.Errorf("failed to merge node %d: %w", i, err)
		}
		result = merged
	}

	return result, nil
}
