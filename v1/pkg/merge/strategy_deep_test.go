package merge

import (
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestNewDeepMergeStrategy(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewDeepMergeStrategy(processor)

	if strategy == nil {
		t.Fatal("NewDeepMergeStrategy should not return nil")
	}

	if strategy.processor != processor {
		t.Error("Strategy should have the provided processor")
	}

	if strategy.Name() != "deep" {
		t.Errorf("Strategy name should be 'deep', got '%s'", strategy.Name())
	}
}

func TestDeepMergeStrategy_Merge(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewDeepMergeStrategy(processor)
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

	t.Run("scalar merge", func(t *testing.T) {
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Scalar merge should not error: %v", err)
		}

		resultScalar, ok := result.(*node.ScalarNode)
		if !ok {
			t.Fatal("Result should be a scalar node")
		}

		if resultScalar.Value != "override" {
			t.Errorf("Result value should be 'override', got '%s'", resultScalar.Value)
		}
	})

	t.Run("mapping merge", func(t *testing.T) {
		// Create base mapping
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		key1 := builder.BuildScalar("key1", node.StylePlain)
		value1 := builder.BuildScalar("value1", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: key1, Value: value1}}

		// Create override mapping
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		key2 := builder.BuildScalar("key2", node.StylePlain)
		value2 := builder.BuildScalar("value2", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: key2, Value: value2}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Mapping merge should not error: %v", err)
		}

		resultMapping, ok := result.(*node.MappingNode)
		if !ok {
			t.Fatal("Result should be a mapping node")
		}

		if len(resultMapping.Pairs) != 2 {
			t.Errorf("Result should have 2 pairs, got %d", len(resultMapping.Pairs))
		}
	})

	t.Run("sequence merge", func(t *testing.T) {
		// Create base sequence
		baseItem := builder.BuildScalar("base", node.StylePlain)
		baseSeq := builder.BuildSequence([]node.Node{baseItem}, node.StyleBlock)

		// Create override sequence
		overrideItem := builder.BuildScalar("override", node.StylePlain)
		overrideSeq := builder.BuildSequence([]node.Node{overrideItem}, node.StyleBlock)

		result, err := strategy.Merge(baseSeq, overrideSeq, ctx)
		if err != nil {
			t.Errorf("Sequence merge should not error: %v", err)
		}

		resultSeq, ok := result.(*node.SequenceNode)
		if !ok {
			t.Fatal("Result should be a sequence node")
		}

		// Default behavior should replace the array
		if len(resultSeq.Items) != 1 {
			t.Errorf("Result should have 1 item (replace), got %d", len(resultSeq.Items))
		}
	})

	t.Run("unknown node type", func(t *testing.T) {
		// This test uses a mock unknown node type
		unknownNode := &mockUnknownNode{}
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(unknownNode, override, ctx)
		if err != nil {
			t.Errorf("Unknown node type should not error: %v", err)
		}

		if result != override {
			t.Error("Should return override for unknown node types")
		}
	})
}

func TestDeepMergeStrategy_MergeMappings(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewDeepMergeStrategy(processor)
	builder := node.NewBuilder()
	ctx := &Context{Options: DefaultOptions()}

	t.Run("type mismatch", func(t *testing.T) {
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideScalar := builder.BuildScalar("scalar", node.StylePlain)

		_, err := strategy.Merge(baseMapping, overrideScalar, ctx)
		if err == nil {
			t.Error("Should error on type mismatch")
		}

		if !strings.Contains(err.Error(), "type mismatch") {
			t.Errorf("Error should mention type mismatch, got: %v", err)
		}
	})

	t.Run("key override", func(t *testing.T) {
		// Create base mapping
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		key := builder.BuildScalar("key", node.StylePlain)
		baseValue := builder.BuildScalar("base", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: key, Value: baseValue}}

		// Create override mapping with same key
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("key", node.StylePlain)
		overrideValue := builder.BuildScalar("override", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideKey, Value: overrideValue}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Key override should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		if len(resultMapping.Pairs) != 1 {
			t.Errorf("Result should have 1 pair, got %d", len(resultMapping.Pairs))
		}

		pair := resultMapping.Pairs[0]
		if scalar, ok := pair.Value.(*node.ScalarNode); ok {
			if scalar.Value != "override" {
				t.Errorf("Value should be 'override', got '%s'", scalar.Value)
			}
		} else {
			t.Error("Value should be a scalar")
		}
	})

	t.Run("nested mapping merge", func(t *testing.T) {
		// Create base with nested mapping
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		nestedKey := builder.BuildScalar("nested", node.StylePlain)
		nestedMapping := builder.BuildMapping(nil, node.StyleBlock)
		nestedBaseKey := builder.BuildScalar("baseKey", node.StylePlain)
		nestedBaseValue := builder.BuildScalar("baseValue", node.StylePlain)
		nestedMapping.Pairs = []*node.MappingPair{{Key: nestedBaseKey, Value: nestedBaseValue}}
		baseMapping.Pairs = []*node.MappingPair{{Key: nestedKey, Value: nestedMapping}}

		// Create override with nested mapping
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideNestedKey := builder.BuildScalar("nested", node.StylePlain)
		overrideNestedMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideNestedKey2 := builder.BuildScalar("overrideKey", node.StylePlain)
		overrideNestedValue := builder.BuildScalar("overrideValue", node.StylePlain)
		overrideNestedMapping.Pairs = []*node.MappingPair{{Key: overrideNestedKey2, Value: overrideNestedValue}}
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideNestedKey, Value: overrideNestedMapping}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Nested mapping merge should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		if len(resultMapping.Pairs) != 1 {
			t.Errorf("Result should have 1 top-level pair, got %d", len(resultMapping.Pairs))
		}

		nestedResult := resultMapping.Pairs[0].Value.(*node.MappingNode)
		if len(nestedResult.Pairs) != 2 {
			t.Errorf("Nested result should have 2 pairs, got %d", len(nestedResult.Pairs))
		}
	})

	t.Run("comment preservation", func(t *testing.T) {
		opts := &Options{PreserveComments: true}
		ctx := &Context{Options: opts}

		// Create base with comments
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		key := builder.BuildScalar("key", node.StylePlain)
		baseValue := builder.BuildScalar("base", node.StylePlain)
		baseValue.LineComment = &node.CommentGroup{
			Comments: []string{"base comment"},
		}
		baseMapping.Pairs = []*node.MappingPair{{Key: key, Value: baseValue}}

		// Create override without comments
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("key", node.StylePlain)
		overrideValue := builder.BuildScalar("override", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideKey, Value: overrideValue}}

		result, err := strategy.Merge(baseMapping, overrideMapping, ctx)
		if err != nil {
			t.Errorf("Comment preservation should not error: %v", err)
		}

		resultMapping := result.(*node.MappingNode)
		resultValue := resultMapping.Pairs[0].Value.(*node.ScalarNode)

		// Comment should be preserved from base
		if resultValue.LineComment == nil {
			t.Error("Base comment should be preserved")
		}
	})
}

func TestDeepMergeStrategy_MergeSequences(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewDeepMergeStrategy(processor)
	builder := node.NewBuilder()

	t.Run("type mismatch", func(t *testing.T) {
		ctx := &Context{Options: DefaultOptions()}
		baseSeq := builder.BuildSequence(nil, node.StyleBlock)
		overrideScalar := builder.BuildScalar("scalar", node.StylePlain)

		_, err := strategy.Merge(baseSeq, overrideScalar, ctx)
		if err == nil {
			t.Error("Should error on type mismatch")
		}
	})

	t.Run("array replace strategy", func(t *testing.T) {
		opts := DefaultOptions().WithArrayStrategy(ArrayReplace)
		ctx := &Context{Options: opts}

		baseItem1 := builder.BuildScalar("base1", node.StylePlain)
		baseItem2 := builder.BuildScalar("base2", node.StylePlain)
		baseSeq := builder.BuildSequence([]node.Node{baseItem1, baseItem2}, node.StyleBlock)

		overrideItem := builder.BuildScalar("override", node.StylePlain)
		overrideSeq := builder.BuildSequence([]node.Node{overrideItem}, node.StyleBlock)

		result, err := strategy.Merge(baseSeq, overrideSeq, ctx)
		if err != nil {
			t.Errorf("Array replace should not error: %v", err)
		}

		resultSeq := result.(*node.SequenceNode)
		if len(resultSeq.Items) != 1 {
			t.Errorf("Result should have 1 item, got %d", len(resultSeq.Items))
		}

		if scalar, ok := resultSeq.Items[0].(*node.ScalarNode); ok {
			if scalar.Value != "override" {
				t.Errorf("Item should be 'override', got '%s'", scalar.Value)
			}
		}
	})

	t.Run("array append strategy", func(t *testing.T) {
		opts := DefaultOptions().WithArrayStrategy(ArrayAppend)
		ctx := &Context{Options: opts}

		baseItem := builder.BuildScalar("base", node.StylePlain)
		baseSeq := builder.BuildSequence([]node.Node{baseItem}, node.StyleBlock)

		overrideItem := builder.BuildScalar("override", node.StylePlain)
		overrideSeq := builder.BuildSequence([]node.Node{overrideItem}, node.StyleBlock)

		result, err := strategy.Merge(baseSeq, overrideSeq, ctx)
		if err != nil {
			t.Errorf("Array append should not error: %v", err)
		}

		resultSeq := result.(*node.SequenceNode)
		if len(resultSeq.Items) != 2 {
			t.Errorf("Result should have 2 items, got %d", len(resultSeq.Items))
		}
	})

	t.Run("array merge by index strategy", func(t *testing.T) {
		opts := DefaultOptions().WithArrayStrategy(ArrayMergeByIndex)
		ctx := &Context{Options: opts}

		baseItem1 := builder.BuildScalar("base1", node.StylePlain)
		baseItem2 := builder.BuildScalar("base2", node.StylePlain)
		baseSeq := builder.BuildSequence([]node.Node{baseItem1, baseItem2}, node.StyleBlock)

		overrideItem1 := builder.BuildScalar("override1", node.StylePlain)
		overrideSeq := builder.BuildSequence([]node.Node{overrideItem1}, node.StyleBlock)

		result, err := strategy.Merge(baseSeq, overrideSeq, ctx)
		if err != nil {
			t.Errorf("Array merge by index should not error: %v", err)
		}

		resultSeq := result.(*node.SequenceNode)
		if len(resultSeq.Items) != 2 {
			t.Errorf("Result should have 2 items, got %d", len(resultSeq.Items))
		}

		// First item should be merged (override)
		if scalar, ok := resultSeq.Items[0].(*node.ScalarNode); ok {
			if scalar.Value != "override1" {
				t.Errorf("First item should be 'override1', got '%s'", scalar.Value)
			}
		}

		// Second item should be from base
		if scalar, ok := resultSeq.Items[1].(*node.ScalarNode); ok {
			if scalar.Value != "base2" {
				t.Errorf("Second item should be 'base2', got '%s'", scalar.Value)
			}
		}
	})
}

func TestDeepMergeStrategy_MergeScalars(t *testing.T) {
	processor := NewNodeProcessor()
	strategy := NewDeepMergeStrategy(processor)
	builder := node.NewBuilder()
	ctx := &Context{Options: DefaultOptions()}

	t.Run("type mismatch", func(t *testing.T) {
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)

		_, err := strategy.Merge(baseScalar, overrideMapping, ctx)
		if err == nil {
			t.Error("Should error on type mismatch")
		}
	})

	t.Run("simple scalar merge", func(t *testing.T) {
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Scalar merge should not error: %v", err)
		}

		resultScalar := result.(*node.ScalarNode)
		if resultScalar.Value != "override" {
			t.Errorf("Result should be 'override', got '%s'", resultScalar.Value)
		}
	})

	t.Run("empty override with OverrideEmpty false", func(t *testing.T) {
		opts := DefaultOptions().WithOverrideEmpty(false)
		ctx := &Context{Options: opts}

		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("", node.StylePlain) // Empty override

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Empty override should not error: %v", err)
		}

		resultScalar := result.(*node.ScalarNode)
		if resultScalar.Value != "base" {
			t.Errorf("Should keep base value when override is empty, got '%s'", resultScalar.Value)
		}
	})

	t.Run("empty override with OverrideEmpty true", func(t *testing.T) {
		opts := DefaultOptions().WithOverrideEmpty(true)
		ctx := &Context{Options: opts}

		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("", node.StylePlain) // Empty override

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Empty override should not error: %v", err)
		}

		resultScalar := result.(*node.ScalarNode)
		if resultScalar.Value != "" {
			t.Errorf("Should use empty override when OverrideEmpty is true, got '%s'", resultScalar.Value)
		}
	})

	t.Run("comment preservation", func(t *testing.T) {
		opts := &Options{PreserveComments: true}
		ctx := &Context{Options: opts}

		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.LineComment = &node.CommentGroup{
			Comments: []string{"base comment"},
		}

		override := builder.BuildScalar("override", node.StylePlain)

		result, err := strategy.Merge(baseScalar, override, ctx)
		if err != nil {
			t.Errorf("Comment preservation should not error: %v", err)
		}

		resultScalar := result.(*node.ScalarNode)

		// Base comment should be preserved
		if resultScalar.LineComment == nil {
			t.Error("Base comment should be preserved")
		}

		if len(resultScalar.LineComment.Comments) == 0 || resultScalar.LineComment.Comments[0] != "base comment" {
			t.Error("Base comment content should be preserved")
		}
	})

	t.Run("override comment takes precedence", func(t *testing.T) {
		opts := &Options{PreserveComments: true}
		ctx := &Context{Options: opts}

		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.LineComment = &node.CommentGroup{
			Comments: []string{"base comment"},
		}

		overrideScalar := builder.BuildScalar("override", node.StylePlain)
		overrideScalar.LineComment = &node.CommentGroup{
			Comments: []string{"override comment"},
		}

		result, err := strategy.Merge(baseScalar, overrideScalar, ctx)
		if err != nil {
			t.Errorf("Comment preservation should not error: %v", err)
		}

		resultScalar := result.(*node.ScalarNode)

		// Override comment should take precedence
		if resultScalar.LineComment == nil {
			t.Error("Override comment should be preserved")
		}

		if len(resultScalar.LineComment.Comments) == 0 || resultScalar.LineComment.Comments[0] != "override comment" {
			t.Error("Override comment should take precedence")
		}
	})
}

// mockUnknownNode is a test helper for unknown node types
type mockUnknownNode struct {
	node.BaseNode
}

func (m *mockUnknownNode) Type() node.NodeType { return node.NodeType(999) }
func (m *mockUnknownNode) Accept(v node.Visitor) error { return nil }
func (m *mockUnknownNode) Clone() node.Node { return m }
func (m *mockUnknownNode) String() string { return "mockUnknownNode" }