package merge

import (
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestNewOverrideStrategy(t *testing.T) {
	strategy := NewOverrideStrategy()

	if strategy == nil {
		t.Fatal("NewOverrideStrategy should not return nil")
	}

	if strategy.Name() != "override" {
		t.Errorf("Strategy name should be 'override', got '%s'", strategy.Name())
	}
}

func TestOverrideStrategy_Merge(t *testing.T) {
	strategy := NewOverrideStrategy()
	builder := node.NewBuilder()
	ctx := &Context{Options: DefaultOptions()}

	t.Run("nil override", func(t *testing.T) {
		base := builder.BuildScalar("base", node.StylePlain)

		result, err := strategy.Merge(base, nil, ctx)
		if err != nil {
			t.Errorf("Merge with nil override should not error: %v", err)
		}
		if result != base {
			t.Error("Should return base when override is nil")
		}
	})

	t.Run("nil base", func(t *testing.T) {
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(nil, override, ctx)
		if err != nil {
			t.Errorf("Merge with nil base should not error: %v", err)
		}
		if result != override {
			t.Error("Should return override when base is nil")
		}
	})

	t.Run("scalar override", func(t *testing.T) {
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Scalar merge should not error: %v", err)
		}

		if result != override {
			t.Error("Should return override node exactly")
		}

		// Verify it's actually the override node, not a copy
		resultScalar, ok := result.(*node.ScalarNode)
		if !ok {
			t.Fatal("Result should be a scalar node")
		}

		if resultScalar.Value != "override" {
			t.Errorf("Result value should be 'override', got '%s'", resultScalar.Value)
		}
	})

	t.Run("mapping override", func(t *testing.T) {
		// Create base mapping
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		baseKey := builder.BuildScalar("baseKey", node.StylePlain)
		baseValue := builder.BuildScalar("baseValue", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: baseKey, Value: baseValue}}

		// Create override mapping
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("overrideKey", node.StylePlain)
		overrideValue := builder.BuildScalar("overrideValue", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideKey, Value: overrideValue}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Mapping merge should not error: %v", err)
		}

		if result != overrideMapping {
			t.Error("Should return override mapping exactly")
		}

		resultMapping, ok := result.(*node.MappingNode)
		if !ok {
			t.Fatal("Result should be a mapping node")
		}

		// Should only contain override content, not base
		if len(resultMapping.Pairs) != 1 {
			t.Errorf("Result should have 1 pair (from override only), got %d", len(resultMapping.Pairs))
		}

		pair := resultMapping.Pairs[0]
		if keyScalar, ok := pair.Key.(*node.ScalarNode); ok {
			if keyScalar.Value != "overrideKey" {
				t.Errorf("Key should be 'overrideKey', got '%s'", keyScalar.Value)
			}
		} else {
			t.Error("Key should be a scalar")
		}

		if valueScalar, ok := pair.Value.(*node.ScalarNode); ok {
			if valueScalar.Value != "overrideValue" {
				t.Errorf("Value should be 'overrideValue', got '%s'", valueScalar.Value)
			}
		} else {
			t.Error("Value should be a scalar")
		}
	})

	t.Run("sequence override", func(t *testing.T) {
		// Create base sequence
		baseItem1 := builder.BuildScalar("base1", node.StylePlain)
		baseItem2 := builder.BuildScalar("base2", node.StylePlain)
		baseSeq := builder.BuildSequence([]node.Node{baseItem1, baseItem2}, node.StyleBlock)

		// Create override sequence
		overrideItem := builder.BuildScalar("override", node.StylePlain)
		overrideSeq := builder.BuildSequence([]node.Node{overrideItem}, node.StyleBlock)

		result, err := strategy.Merge(baseSeq, overrideSeq, ctx)
		if err != nil {
			t.Errorf("Sequence merge should not error: %v", err)
		}

		if result != overrideSeq {
			t.Error("Should return override sequence exactly")
		}

		resultSeq, ok := result.(*node.SequenceNode)
		if !ok {
			t.Fatal("Result should be a sequence node")
		}

		// Should only contain override content
		if len(resultSeq.Items) != 1 {
			t.Errorf("Result should have 1 item (from override only), got %d", len(resultSeq.Items))
		}

		if itemScalar, ok := resultSeq.Items[0].(*node.ScalarNode); ok {
			if itemScalar.Value != "override" {
				t.Errorf("Item should be 'override', got '%s'", itemScalar.Value)
			}
		} else {
			t.Error("Item should be a scalar")
		}
	})

	t.Run("type mismatch override", func(t *testing.T) {
		// Base is mapping, override is scalar
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		baseKey := builder.BuildScalar("key", node.StylePlain)
		baseValue := builder.BuildScalar("value", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: baseKey, Value: baseValue}}

		overrideScalar := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(baseMapping, overrideScalar, ctx)
		if err != nil {
			t.Errorf("Type mismatch should not error: %v", err)
		}

		if result != overrideScalar {
			t.Error("Should return override regardless of type mismatch")
		}

		// Verify the base mapping is completely ignored
		if resultScalar, ok := result.(*node.ScalarNode); ok {
			if resultScalar.Value != "override" {
				t.Errorf("Result should be override scalar, got '%s'", resultScalar.Value)
			}
		} else {
			t.Error("Result should be the override scalar")
		}
	})

	t.Run("complex nested structure override", func(t *testing.T) {
		// Create complex base structure
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)

		// Add nested mapping
		nestedKey := builder.BuildScalar("nested", node.StylePlain)
		nestedMapping := builder.BuildMapping(nil, node.StyleBlock)
		nestedSubKey := builder.BuildScalar("subkey", node.StylePlain)
		nestedSubValue := builder.BuildScalar("subvalue", node.StylePlain)
		nestedMapping.Pairs = []*node.MappingPair{{Key: nestedSubKey, Value: nestedSubValue}}

		// Add sequence
		seqKey := builder.BuildScalar("sequence", node.StylePlain)
		item1 := builder.BuildScalar("item1", node.StylePlain)
		item2 := builder.BuildScalar("item2", node.StylePlain)
		sequence := builder.BuildSequence([]node.Node{item1, item2}, node.StyleBlock)

		baseMapping.Pairs = []*node.MappingPair{
			{Key: nestedKey, Value: nestedMapping},
			{Key: seqKey, Value: sequence},
		}

		// Simple override
		override := builder.BuildScalar("simple", node.StylePlain)

		result, err := strategy.Merge(baseMapping, override, ctx)
		if err != nil {
			t.Errorf("Complex structure override should not error: %v", err)
		}

		if result != override {
			t.Error("Should return override exactly, ignoring complex base")
		}

		// Verify the entire complex base structure is ignored
		if resultScalar, ok := result.(*node.ScalarNode); ok {
			if resultScalar.Value != "simple" {
				t.Errorf("Result should be 'simple', got '%s'", resultScalar.Value)
			}
		} else {
			t.Error("Result should be the override scalar")
		}
	})

	t.Run("comments and metadata ignored", func(t *testing.T) {
		// Create base with extensive metadata
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.HeadComment = &node.CommentGroup{Comments: []string{"base head comment"}}
		baseScalar.LineComment = &node.CommentGroup{Comments: []string{"base line comment"}}
		baseScalar.FootComment = &node.CommentGroup{Comments: []string{"base foot comment"}}
		baseScalar.BlankLinesBefore = 2
		baseScalar.BlankLinesAfter = 1

		// Create override with different metadata
		overrideScalar := builder.BuildScalar("override", node.StylePlain)
		overrideScalar.LineComment = &node.CommentGroup{Comments: []string{"override line comment"}}

		result, err := strategy.Merge(baseScalar, overrideScalar, ctx)
		if err != nil {
			t.Errorf("Metadata override should not error: %v", err)
		}

		if result != overrideScalar {
			t.Error("Should return override exactly with its metadata")
		}

		resultScalar := result.(*node.ScalarNode)

		// Should have override's metadata, not base's
		if resultScalar.HeadComment != nil {
			t.Error("Result should not have base's head comment")
		}

		if resultScalar.LineComment == nil {
			t.Error("Result should have override's line comment")
		} else if len(resultScalar.LineComment.Comments) == 0 || resultScalar.LineComment.Comments[0] != "override line comment" {
			t.Error("Result should have override's line comment content")
		}

		if resultScalar.FootComment != nil {
			t.Error("Result should not have base's foot comment")
		}

		if resultScalar.BlankLinesBefore != 0 {
			t.Error("Result should not have base's blank lines before")
		}

		if resultScalar.BlankLinesAfter != 0 {
			t.Error("Result should not have base's blank lines after")
		}
	})

	t.Run("context and options ignored", func(t *testing.T) {
		// Create context with various options
		opts := &Options{
			Strategy:           StrategyDeep, // This should be ignored
			PreserveComments:   true,        // This should be ignored
			PreserveBlankLines: true,        // This should be ignored
			ArrayMergeStrategy: ArrayAppend, // This should be ignored
			OverrideEmpty:      false,       // This should be ignored
		}
		ctx := &Context{Options: opts, Depth: 5, Path: []string{"deep", "path"}}

		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Context options should not affect override strategy: %v", err)
		}

		if result != override {
			t.Error("Should return override exactly regardless of context")
		}

		// The override strategy completely ignores all options and context
		resultScalar := result.(*node.ScalarNode)
		if resultScalar.Value != "override" {
			t.Errorf("Result should be 'override', got '%s'", resultScalar.Value)
		}
	})

	t.Run("identity behavior", func(t *testing.T) {
		// Override strategy should return the exact same object reference
		override := builder.BuildScalar("test", node.StylePlain)
		base := builder.BuildScalar("base", node.StylePlain)

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Identity test should not error: %v", err)
		}

		// Should be the exact same object, not a copy
		if result != override {
			t.Error("Override strategy should return the exact same override object")
		}

		// Verify it's not a deep copy by checking object identity
		overridePtr := &override
		resultPtr := &result
		if *overridePtr != *resultPtr {
			t.Error("Should be the exact same object reference")
		}
	})
}