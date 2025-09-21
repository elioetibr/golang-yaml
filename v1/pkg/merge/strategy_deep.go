package merge

import (
	"fmt"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// DeepMergeStrategy performs recursive deep merging
type DeepMergeStrategy struct {
	processor *NodeProcessor
}

// NewDeepMergeStrategy creates a new deep merge strategy
func NewDeepMergeStrategy(processor *NodeProcessor) *DeepMergeStrategy {
	return &DeepMergeStrategy{
		processor: processor,
	}
}

// Name returns the strategy name
func (s *DeepMergeStrategy) Name() string {
	return "deep"
}

// Merge performs deep recursive merging
func (s *DeepMergeStrategy) Merge(base, override node.Node, ctx *Context) (node.Node, error) {
	if override == nil {
		return base, nil
	}
	if base == nil {
		return override, nil
	}

	// Type-specific merging
	switch baseNode := base.(type) {
	case *node.MappingNode:
		return s.mergeMappings(baseNode, override, ctx)
	case *node.SequenceNode:
		return s.mergeSequences(baseNode, override, ctx)
	case *node.ScalarNode:
		return s.mergeScalars(baseNode, override, ctx)
	default:
		return override, nil
	}
}

// mergeMappings merges two mapping nodes
func (s *DeepMergeStrategy) mergeMappings(base *node.MappingNode, override node.Node, ctx *Context) (*node.MappingNode, error) {
	overrideMapping, ok := override.(*node.MappingNode)
	if !ok {
		// Type mismatch - return override
		return nil, fmt.Errorf("type mismatch: expected mapping, got %T", override)
	}

	// Transfer inter-field comments from base before processing
	// This moves comments from value HeadComments to next key HeadComments where they belong
	baseCopy := &node.MappingNode{
		BaseNode: base.BaseNode,
		Pairs:    make([]*node.MappingPair, len(base.Pairs)),
		Style:    base.Style,
	}
	copy(baseCopy.Pairs, base.Pairs)
	s.processor.TransferInterFieldComments(baseCopy.Pairs)

	// Create result mapping, preserving base metadata
	result := &node.MappingNode{
		BaseNode: baseCopy.BaseNode,
		Pairs:    make([]*node.MappingPair, 0),
		Style:    baseCopy.Style,
	}

	// If override has a LineComment (inline comment on mapping key), preserve it
	if overrideMapping.LineComment != nil && ctx.Options.PreserveComments {
		result.LineComment = overrideMapping.LineComment
	}

	// Build override map for quick lookup
	overrideMap := make(map[string]*node.MappingPair)
	for _, pair := range overrideMapping.Pairs {
		if key, ok := s.processor.GetScalarValue(pair.Key); ok {
			overrideMap[key] = pair
		}
	}

	// Process base pairs
	processedKeys := make(map[string]bool)
	for _, basePair := range baseCopy.Pairs {
		key, ok := s.processor.GetScalarValue(basePair.Key)
		if !ok {
			// Non-scalar key, keep as is
			result.Pairs = append(result.Pairs, basePair)
			continue
		}

		processedKeys[key] = true

		if overridePair, exists := overrideMap[key]; exists {
			// Key exists in override, merge values-with-comments
			mergedValue, err := s.Merge(basePair.Value, overridePair.Value, ctx.WithPath(key))
			if err != nil {
				return nil, err
			}

			// Clean head comment from scalar values-with-comments to prevent formatting issues
			mergedValue = s.processor.CleanScalarHeadComment(mergedValue)

			// Preserve the key node with all its comments
			mergedKey := s.processor.PreserveKeyNode(basePair.Key, overridePair.Key, ctx.Options)

			// Create merged pair preserving all metadata
			newPair := &node.MappingPair{
				Key:              mergedKey,
				Value:            mergedValue,
				KeyComment:       basePair.KeyComment,
				ValueComment:     basePair.ValueComment,
				BlankLinesBefore: basePair.BlankLinesBefore,
				BlankLinesAfter:  basePair.BlankLinesAfter,
			}

			// Override with override's metadata if present and options allow
			if ctx.Options.PreserveComments {
				// Preserve ValueComment from override if present
				if overridePair.ValueComment != nil {
					newPair.ValueComment = overridePair.ValueComment
				}
				// Preserve KeyComment from base unless override has one
				if overridePair.KeyComment != nil {
					newPair.KeyComment = overridePair.KeyComment
				}
			}

			// Handle line comments for scalar values-with-comments
			if scalar, ok := overridePair.Value.(*node.ScalarNode); ok && scalar.LineComment != nil {
				if mergedScalar, ok := mergedValue.(*node.ScalarNode); ok {
					mergedScalar.LineComment = scalar.LineComment
				}
			}

			result.Pairs = append(result.Pairs, newPair)
		} else {
			// Key only exists in base, keep it completely as is
			result.Pairs = append(result.Pairs, basePair)
		}
	}

	// Add keys from override that weren't in base
	for _, overridePair := range overrideMapping.Pairs {
		key, ok := s.processor.GetScalarValue(overridePair.Key)
		if !ok || processedKeys[key] {
			continue
		}

		// Clean head comment from scalar values-with-comments
		cleanedValue := s.processor.CleanScalarHeadComment(overridePair.Value)

		// Keep the override pair completely with all its comments
		newPair := &node.MappingPair{
			Key:              overridePair.Key,
			Value:            cleanedValue,
			KeyComment:       overridePair.KeyComment,
			ValueComment:     overridePair.ValueComment,
			BlankLinesBefore: overridePair.BlankLinesBefore,
			BlankLinesAfter:  overridePair.BlankLinesAfter,
		}

		result.Pairs = append(result.Pairs, newPair)
	}

	// Preserve metadata
	s.processor.PreserveMetadata(result, baseCopy, ctx.Options)

	return result, nil
}

// mergeSequences merges two sequence nodes
func (s *DeepMergeStrategy) mergeSequences(base *node.SequenceNode, override node.Node, ctx *Context) (*node.SequenceNode, error) {
	overrideSeq, ok := override.(*node.SequenceNode)
	if !ok {
		return nil, fmt.Errorf("type mismatch: expected sequence, got %T", override)
	}

	// Handle array merge strategy
	switch ctx.Options.ArrayMergeStrategy {
	case ArrayAppend:
		// Append arrays
		result := &node.SequenceNode{
			BaseNode: base.BaseNode,
			Items:    make([]node.Node, 0, len(base.Items)+len(overrideSeq.Items)),
			Style:    base.Style,
		}
		result.Items = append(result.Items, base.Items...)
		result.Items = append(result.Items, overrideSeq.Items...)
		s.processor.PreserveMetadata(result, base, ctx.Options)
		return result, nil

	case ArrayMergeByIndex:
		// Merge by index
		maxLen := len(base.Items)
		if len(overrideSeq.Items) > maxLen {
			maxLen = len(overrideSeq.Items)
		}

		result := &node.SequenceNode{
			BaseNode: base.BaseNode,
			Items:    make([]node.Node, 0, maxLen),
			Style:    base.Style,
		}

		for i := 0; i < maxLen; i++ {
			var item node.Node
			if i < len(base.Items) && i < len(overrideSeq.Items) {
				// Both have item at this index, merge them
				merged, err := s.Merge(base.Items[i], overrideSeq.Items[i], ctx)
				if err != nil {
					return nil, err
				}
				item = merged
			} else if i < len(base.Items) {
				// Only base has item
				item = base.Items[i]
			} else {
				// Only override has item
				item = overrideSeq.Items[i]
			}
			result.Items = append(result.Items, item)
		}
		s.processor.PreserveMetadata(result, base, ctx.Options)
		return result, nil

	default: // ArrayReplace
		// Replace entire array
		result := &node.SequenceNode{
			BaseNode: overrideSeq.BaseNode,
			Items:    overrideSeq.Items,
			Style:    overrideSeq.Style,
		}
		// Preserve comments from base if override doesn't have them
		s.processor.PreserveMetadata(result, base, ctx.Options)
		return result, nil
	}
}

// mergeScalars merges two scalar nodes
func (s *DeepMergeStrategy) mergeScalars(base *node.ScalarNode, override node.Node, ctx *Context) (*node.ScalarNode, error) {
	overrideScalar, ok := override.(*node.ScalarNode)
	if !ok {
		return nil, fmt.Errorf("type mismatch: expected scalar, got %T", override)
	}

	// Check if override is empty and OverrideEmpty is false
	if !ctx.Options.OverrideEmpty && overrideScalar.Value == "" && base.Value != "" {
		return base, nil
	}

	// Create result with override value but cleaned head comment
	result := &node.ScalarNode{
		BaseNode: node.BaseNode{
			TagValue:    overrideScalar.TagValue,
			AnchorValue: overrideScalar.AnchorValue,
			// Don't set LineComment here - let it be handled below
			FootComment: overrideScalar.FootComment,
			// No HeadComment to keep value on same line as key
		},
		Value: overrideScalar.Value,
		Style: overrideScalar.Style,
		Alias: overrideScalar.Alias,
	}

	// Preserve LineComment: use override's if it has one, otherwise keep base's
	if overrideScalar.LineComment != nil {
		result.LineComment = overrideScalar.LineComment
	} else if base.LineComment != nil && ctx.Options.PreserveComments {
		result.LineComment = base.LineComment
	}

	// Preserve other metadata from base if not in override
	s.processor.PreserveMetadata(result, base, ctx.Options)

	return result, nil
}
