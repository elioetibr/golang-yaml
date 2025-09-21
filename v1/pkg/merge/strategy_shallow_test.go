package merge

import (
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestNewShallowMergeStrategy(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewShallowMergeStrategy(processor)

	if strategy == nil {
		t.Fatal("NewShallowMergeStrategy should not return nil")
	}

	if strategy.processor != processor {
		t.Error("Strategy should have the provided processor")
	}

	if strategy.Name() != "shallow" {
		t.Errorf("Strategy name should be 'shallow', got '%s'", strategy.Name())
	}
}

func TestShallowMergeStrategy_Merge(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewShallowMergeStrategy(processor)
	builder := node.NewBuilder()
	ctx := &Context{Options: DefaultOptions()}

	t.Run("nil nodes", func(t *testing.T) {
		base := builder.BuildScalar("base", node.StylePlain)

		// Override is nil
		result, err := strategy.Merge(base, nil, ctx)
		if err != nil {
			t.Errorf("Merge with nil override should not error: %v", err)
		}
		if result != base {
			t.Error("Should return base when override is nil")
		}

		// Base is nil
		override := builder.BuildScalar("override", node.StylePlain)
		result, err = strategy.Merge(nil, override, ctx)
		if err != nil {
			t.Errorf("Merge with nil base should not error: %v", err)
		}
		if result != override {
			t.Error("Should return override when base is nil")
		}
	})

	t.Run("non-mapping nodes", func(t *testing.T) {
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Non-mapping merge should not error: %v", err)
		}

		if result != override {
			t.Error("Should return override for non-mapping nodes")
		}
	})

	t.Run("mapping to scalar", func(t *testing.T) {
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideScalar := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(baseMapping, overrideScalar, ctx)
		if err != nil {
			t.Errorf("Mapping to scalar should not error: %v", err)
		}

		if result != overrideScalar {
			t.Error("Should return override when types don't match")
		}
	})

	t.Run("scalar to mapping", func(t *testing.T) {
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)

		result, err := strategy.Merge(baseScalar, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Scalar to mapping should not error: %v", err)
		}

		if result != overrideMapping {
			t.Error("Should return override when types don't match")
		}
	})

	t.Run("mapping merge - shallow only", func(t *testing.T) {
		// Create base mapping with nested structure
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)

		// Add a nested mapping to base
		nestedKey := builder.BuildScalar("nested", node.StylePlain)
		nestedMapping := builder.BuildMapping(nil, node.StyleBlock)
		nestedBaseKey := builder.BuildScalar("baseKey", node.StylePlain)
		nestedBaseValue := builder.BuildScalar("baseValue", node.StylePlain)
		nestedMapping.Pairs = []*node.MappingPair{{Key: nestedBaseKey, Value: nestedBaseValue}}

		// Add a simple key
		simpleKey := builder.BuildScalar("simple", node.StylePlain)
		simpleValue := builder.BuildScalar("baseSimple", node.StylePlain)

		baseMapping.Pairs = []*node.MappingPair{
			{Key: nestedKey, Value: nestedMapping},
			{Key: simpleKey, Value: simpleValue},
		}

		// Create override mapping
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)

		// Override the nested mapping completely (shallow merge)
		overrideNestedKey := builder.BuildScalar("nested", node.StylePlain)
		overrideNestedMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideNestedKey2 := builder.BuildScalar("overrideKey", node.StylePlain)
		overrideNestedValue := builder.BuildScalar("overrideValue", node.StylePlain)
		overrideNestedMapping.Pairs = []*node.MappingPair{{Key: overrideNestedKey2, Value: overrideNestedValue}}

		// Override the simple key
		overrideSimpleKey := builder.BuildScalar("simple", node.StylePlain)
		overrideSimpleValue := builder.BuildScalar("overrideSimple", node.StylePlain)

		overrideMapping.Pairs = []*node.MappingPair{
			{Key: overrideNestedKey, Value: overrideNestedMapping},
			{Key: overrideSimpleKey, Value: overrideSimpleValue},
		}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Shallow mapping merge should not error: %v", err)
		}

		resultMapping, ok := result.(*node.MappingNode)
		if !ok {
			t.Fatal("Result should be a mapping node")
		}

		if len(resultMapping.Pairs) != 2 {
			t.Errorf("Result should have 2 pairs, got %d", len(resultMapping.Pairs))
		}

		// Check that nested mapping was completely replaced (not merged)
		var nestedResult *node.MappingNode
		for _, pair := range resultMapping.Pairs {
			if keyScalar, ok := pair.Key.(*node.ScalarNode); ok && keyScalar.Value == "nested" {
				if mappingNode, ok := pair.Value.(*node.MappingNode); ok {
					nestedResult = mappingNode
				}
			}
		}

		if nestedResult == nil {
			t.Fatal("Nested mapping should exist in result")
		}

		// Should only have override key (baseKey should not be present - shallow merge)
		if len(nestedResult.Pairs) != 1 {
			t.Errorf("Nested result should have 1 pair (shallow merge), got %d", len(nestedResult.Pairs))
		}

		if nestedPair := nestedResult.Pairs[0]; nestedPair != nil {
			if keyScalar, ok := nestedPair.Key.(*node.ScalarNode); ok {
				if keyScalar.Value != "overrideKey" {
					t.Errorf("Nested key should be 'overrideKey', got '%s'", keyScalar.Value)
				}
			}
		}
	})

	t.Run("new keys from override", func(t *testing.T) {
		// Create base mapping
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		baseKey := builder.BuildScalar("baseKey", node.StylePlain)
		baseValue := builder.BuildScalar("baseValue", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: baseKey, Value: baseValue}}

		// Create override mapping with new key
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("overrideKey", node.StylePlain)
		overrideValue := builder.BuildScalar("overrideValue", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideKey, Value: overrideValue}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("New keys merge should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		if len(resultMapping.Pairs) != 2 {
			t.Errorf("Result should have 2 pairs, got %d", len(resultMapping.Pairs))
		}

		// Verify both keys are present
		foundBase := false
		foundOverride := false
		for _, pair := range resultMapping.Pairs {
			if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
				if keyScalar.Value == "baseKey" {
					foundBase = true
				} else if keyScalar.Value == "overrideKey" {
					foundOverride = true
				}
			}
		}

		if !foundBase {
			t.Error("Base key should be preserved")
		}

		if !foundOverride {
			t.Error("Override key should be added")
		}
	})

	t.Run("comment preservation", func(t *testing.T) {
		opts := &Options{PreserveComments: true}
		ctx := &Context{Options: opts}

		// Create base with comments
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		key := builder.BuildScalar("key", node.StylePlain)
		baseValue := builder.BuildScalar("base", node.StylePlain)
		basePair := &node.MappingPair{
			Key:          key,
			Value:        baseValue,
			KeyComment:   &node.CommentGroup{Comments: []string{"key comment"}},
			ValueComment: &node.CommentGroup{Comments: []string{"base value comment"}},
		}
		baseMapping.Pairs = []*node.MappingPair{basePair}

		// Create override
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("key", node.StylePlain)
		overrideValue := builder.BuildScalar("override", node.StylePlain)
		overridePair := &node.MappingPair{
			Key:          overrideKey,
			Value:        overrideValue,
			ValueComment: &node.CommentGroup{Comments: []string{"override value comment"}},
		}
		overrideMapping.Pairs = []*node.MappingPair{overridePair}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Comment preservation should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		resultPair := resultMapping.Pairs[0]

		// Key comment should be preserved from base
		if resultPair.KeyComment == nil {
			t.Error("Key comment should be preserved from base")
		}

		// Value comment should come from override
		if resultPair.ValueComment == nil {
			t.Error("Value comment should be preserved")
		}

		if len(resultPair.ValueComment.Comments) == 0 || resultPair.ValueComment.Comments[0] != "override value comment" {
			t.Error("Override value comment should take precedence")
		}
	})

	t.Run("non-scalar keys", func(t *testing.T) {
		// Create base mapping with non-scalar key
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		complexKey := builder.BuildMapping(nil, node.StyleFlow) // Non-scalar key
		baseValue := builder.BuildScalar("base", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: complexKey, Value: baseValue}}

		// Create override mapping
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		scalarKey := builder.BuildScalar("scalarKey", node.StylePlain)
		overrideValue := builder.BuildScalar("override", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: scalarKey, Value: overrideValue}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Non-scalar keys should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		if len(resultMapping.Pairs) != 2 {
			t.Errorf("Result should have 2 pairs, got %d", len(resultMapping.Pairs))
		}

		// Both pairs should be preserved
		foundComplex := false
		foundScalar := false
		for _, pair := range resultMapping.Pairs {
			if _, ok := pair.Key.(*node.MappingNode); ok {
				foundComplex = true
			} else if scalar, ok := pair.Key.(*node.ScalarNode); ok && scalar.Value == "scalarKey" {
				foundScalar = true
			}
		}

		if !foundComplex {
			t.Error("Complex key should be preserved")
		}

		if !foundScalar {
			t.Error("Scalar key should be added")
		}
	})

	t.Run("empty mappings", func(t *testing.T) {
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Empty mappings should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		if len(resultMapping.Pairs) != 0 {
			t.Errorf("Result should have 0 pairs, got %d", len(resultMapping.Pairs))
		}
	})

	t.Run("metadata preservation", func(t *testing.T) {
		opts := &Options{PreserveComments: true, PreserveBlankLines: true}
		ctx := &Context{Options: opts}

		// Create base with metadata
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		baseMapping.HeadComment = &node.CommentGroup{Comments: []string{"base comment"}}
		baseMapping.BlankLinesBefore = 2

		// Create override
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Metadata preservation should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)

		// Base metadata should be preserved
		if resultMapping.HeadComment == nil {
			t.Error("Base comment should be preserved")
		}

		if resultMapping.BlankLinesBefore != 2 {
			t.Error("Base blank lines should be preserved")
		}
	})
}
