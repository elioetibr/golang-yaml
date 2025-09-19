package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/elioetibr/golang-yaml/pkg/node"
	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
)

func main() {
	// Get directory of main.go
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Now you can access files relative to main.go
	valuesFile := filepath.Join(dir, "values.yaml")
	valuesWithMergeContentFile := filepath.Join(dir, "values-merge.yaml")
	valuesOverriddenFile := filepath.Join(dir, "values-overridden.yaml")

	// Load the base values.yaml
	baseData, err := os.ReadFile(valuesFile)
	if err != nil {
		log.Fatalf("Error reading values.yaml: %v", err)
	}

	// Load the values-merge.yaml
	mergeData, err := os.ReadFile(valuesWithMergeContentFile)
	if err != nil {
		log.Fatalf("Error reading values-merge.yaml: %v", err)
	}

	// Parse base YAML
	baseRoot, err := parser.ParseString(string(baseData))
	if err != nil {
		log.Fatalf("Error parsing values.yaml: %v", err)
	}

	// Parse merge YAML
	mergeRoot, err := parser.ParseString(string(mergeData))
	if err != nil {
		log.Fatalf("Error parsing values-merge.yaml: %v", err)
	}

	// Merge the two YAML documents
	mergedRoot := mergeNodes(baseRoot, mergeRoot)

	// Configure serialization options to preserve formatting
	serializerOpts := &serializer.Options{
		Indent:             2,
		PreserveComments:   true,
		PreserveBlankLines: true,
		LineWidth:          80,
	}

	// Serialize the merged result
	result, err := serializer.SerializeToString(mergedRoot, serializerOpts)
	if err != nil {
		log.Fatalf("Error serializing merged YAML: %v", err)
	}

	// Write to output file
	err = os.WriteFile(valuesOverriddenFile, []byte(result), 0644)
	if err != nil {
		log.Fatalf("Error writing values-overridden.yaml: %v", err)
	}

	fmt.Println("Successfully merged YAML files and wrote to values-overridden.yaml")
}

// mergeNodes recursively merges two YAML nodes, with the second node taking precedence
func mergeNodes(base, override node.Node) node.Node {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}

	// Handle different node types
	switch baseNode := base.(type) {
	case *node.MappingNode:
		overrideMapping, ok := override.(*node.MappingNode)
		if !ok {
			// If override is not a mapping, replace entirely
			return override
		}
		return mergeMappings(baseNode, overrideMapping)

	case *node.SequenceNode:
		// For sequences, typically replace entirely
		return override

	case *node.ScalarNode:
		// For scalars, create a new node with the override value
		overrideScalar, ok := override.(*node.ScalarNode)
		if !ok {
			return override
		}
		// Return the override scalar, preserving any line comments
		return overrideScalar

	default:
		return override
	}
}

// cleanScalarHeadComment removes head comments from scalar nodes to prevent formatting issues
func cleanScalarHeadComment(n node.Node) node.Node {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return &node.ScalarNode{
			BaseNode: node.BaseNode{
				TagValue:    scalar.TagValue,
				AnchorValue: scalar.AnchorValue,
				LineComment: scalar.LineComment,
				FootComment: scalar.FootComment,
				// No HeadComment to keep value on same line as key
			},
			Value: scalar.Value,
			Style: scalar.Style,
			Alias: scalar.Alias,
		}
	}
	return n
}

// mergeMappings merges two mapping nodes
func mergeMappings(base, override *node.MappingNode) *node.MappingNode {
	// Create a new mapping node, preserving base node's metadata
	result := &node.MappingNode{
		BaseNode: base.BaseNode,
		Pairs:    make([]*node.MappingPair, 0),
		Style:    base.Style,
	}

	// Create a map to track which keys have been overridden
	overrideKeys := make(map[string]*node.MappingPair)
	for _, pair := range override.Pairs {
		if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
			overrideKeys[keyScalar.Value] = pair
		}
	}

	// Process base pairs
	for _, basePair := range base.Pairs {
		if keyScalar, ok := basePair.Key.(*node.ScalarNode); ok {
			if overridePair, exists := overrideKeys[keyScalar.Value]; exists {
				// Key exists in override, merge the values
				mergedValue := mergeNodes(basePair.Value, overridePair.Value)

				// Clean head comments from scalar values to prevent formatting issues
				mergedValue = cleanScalarHeadComment(mergedValue)

				newPair := &node.MappingPair{
					Key:              basePair.Key, // Keep base key to preserve its comments
					Value:            mergedValue,
					BlankLinesBefore: basePair.BlankLinesBefore,
					BlankLinesAfter:  basePair.BlankLinesAfter,
				}

				// Preserve comments - prefer base key comments, override value comments
				newPair.KeyComment = basePair.KeyComment
				if overridePair.ValueComment != nil {
					newPair.ValueComment = overridePair.ValueComment
				} else if basePair.ValueComment != nil {
					newPair.ValueComment = basePair.ValueComment
				}

				result.Pairs = append(result.Pairs, newPair)
				// Mark as processed
				delete(overrideKeys, keyScalar.Value)
			} else {
				// Key only exists in base, keep it as is
				result.Pairs = append(result.Pairs, basePair)
			}
		}
	}

	// Add remaining override keys that weren't in base
	for _, pair := range override.Pairs {
		if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
			if _, stillExists := overrideKeys[keyScalar.Value]; stillExists {
				// Clean head comments from scalar values
				cleanedPair := &node.MappingPair{
					Key:              pair.Key,
					Value:            cleanScalarHeadComment(pair.Value),
					KeyComment:       pair.KeyComment,
					ValueComment:     pair.ValueComment,
					BlankLinesBefore: pair.BlankLinesBefore,
					BlankLinesAfter:  pair.BlankLinesAfter,
				}
				result.Pairs = append(result.Pairs, cleanedPair)
			}
		}
	}

	// Preserve document-level comments from base (or override if base doesn't have them)
	if base.HeadComment != nil {
		result.HeadComment = base.HeadComment
	} else if override.HeadComment != nil {
		result.HeadComment = override.HeadComment
	}

	if base.LineComment != nil {
		result.LineComment = base.LineComment
	} else if override.LineComment != nil {
		result.LineComment = override.LineComment
	}

	if base.FootComment != nil {
		result.FootComment = base.FootComment
	} else if override.FootComment != nil {
		result.FootComment = override.FootComment
	}

	return result
}