package merge

import (
	"github.com/elioetibr/golang-yaml/v0/pkg/node"
)

// ShallowMergeStrategy only merges top-level keys
type ShallowMergeStrategy struct {
	processor *NodeProcessor
}

// NewShallowMergeStrategy creates a new shallow merge strategy
func NewShallowMergeStrategy(processor *NodeProcessor) *ShallowMergeStrategy {
	return &ShallowMergeStrategy{
		processor: processor,
	}
}

// Name returns the strategy name
func (s *ShallowMergeStrategy) Name() string {
	return "shallow"
}

// Merge performs shallow merging (top-level only)
func (s *ShallowMergeStrategy) Merge(base, override node.Node, ctx *Context) (node.Node, error) {
	if override == nil {
		return base, nil
	}
	if base == nil {
		return override, nil
	}

	// Only merge if both are mappings
	baseMapping, baseOk := base.(*node.MappingNode)
	overrideMapping, overrideOk := override.(*node.MappingNode)

	if !baseOk || !overrideOk {
		// Not both mappings, return override
		return override, nil
	}

	// Create result mapping
	result := &node.MappingNode{
		BaseNode: baseMapping.BaseNode,
		Pairs:    make([]*node.MappingPair, 0),
		Style:    baseMapping.Style,
	}

	// Build override map
	overrideMap := make(map[string]*node.MappingPair)
	for _, pair := range overrideMapping.Pairs {
		if key, ok := s.processor.GetScalarValue(pair.Key); ok {
			overrideMap[key] = pair
		}
	}

	// Process base pairs
	processedKeys := make(map[string]bool)
	for _, basePair := range baseMapping.Pairs {
		key, ok := s.processor.GetScalarValue(basePair.Key)
		if !ok {
			result.Pairs = append(result.Pairs, basePair)
			continue
		}

		processedKeys[key] = true

		if overridePair, exists := overrideMap[key]; exists {
			// Replace with override (no deep merge)
			cleanedValue := s.processor.CleanScalarHeadComment(overridePair.Value)

			// Preserve the key node with all its comments
			mergedKey := s.processor.PreserveKeyNode(basePair.Key, overridePair.Key, ctx.Options)

			newPair := &node.MappingPair{
				Key:              mergedKey,
				Value:            cleanedValue,
				KeyComment:       basePair.KeyComment,
				ValueComment:     basePair.ValueComment,
				BlankLinesBefore: basePair.BlankLinesBefore,
				BlankLinesAfter:  basePair.BlankLinesAfter,
			}

			// Preserve inline comments from override
			if overridePair.ValueComment != nil {
				newPair.ValueComment = overridePair.ValueComment
			}

			result.Pairs = append(result.Pairs, newPair)
		} else {
			result.Pairs = append(result.Pairs, basePair)
		}
	}

	// Add new keys from override
	for _, overridePair := range overrideMapping.Pairs {
		key, ok := s.processor.GetScalarValue(overridePair.Key)
		if !ok || processedKeys[key] {
			continue
		}

		cleanedValue := s.processor.CleanScalarHeadComment(overridePair.Value)

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
	s.processor.PreserveMetadata(result, baseMapping, ctx.Options)

	return result, nil
}
