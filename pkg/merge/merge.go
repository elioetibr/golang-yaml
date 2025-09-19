// Package merge provides YAML merging capabilities with configurable strategies
package merge

import (
	"fmt"
	"os"

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

	return result, nil
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
