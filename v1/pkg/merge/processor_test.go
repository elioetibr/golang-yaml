package merge

import (
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestNewNodeProcessor(t *testing.T) {
	processor := NewNodeProcessor()

	if processor == nil {
		t.Fatal("NewNodeProcessor() should not return nil")
	}
}

func TestNodeProcessor_CleanScalarHeadComment(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("scalar node with comments", func(t *testing.T) {
		scalarNode := builder.BuildScalar("test", node.StylePlain)

		// Add comments
		scalarNode.HeadComment = &node.CommentGroup{Comments: []string{"head comment"}}
		scalarNode.LineComment = &node.CommentGroup{Comments: []string{"line comment"}}
		scalarNode.FootComment = &node.CommentGroup{Comments: []string{"foot comment"}}

		result := processor.CleanScalarHeadComment(scalarNode)

		if result == nil {
			t.Fatal("CleanScalarHeadComment should not return nil")
		}

		resultScalar, ok := result.(*node.ScalarNode)
		if !ok {
			t.Fatal("Result should be a scalar node")
		}

		// Head comment should be removed
		if resultScalar.HeadComment != nil {
			t.Error("Head comment should be removed")
		}

		// Other comments should be preserved
		if resultScalar.LineComment == nil {
			t.Error("Line comment should be preserved")
		}

		if resultScalar.FootComment == nil {
			t.Error("Foot comment should be preserved")
		}

		// Value should be preserved
		if resultScalar.Value != "test" {
			t.Errorf("Value should be preserved, got %s", resultScalar.Value)
		}
	})

	t.Run("non-scalar node", func(t *testing.T) {
		mapping := builder.BuildMapping(nil, node.StyleBlock)

		result := processor.CleanScalarHeadComment(mapping)

		if result != mapping {
			t.Error("Non-scalar nodes should be returned unchanged")
		}
	})

	t.Run("scalar without comments", func(t *testing.T) {
		scalar := builder.BuildScalar("test", node.StylePlain)

		result := processor.CleanScalarHeadComment(scalar)

		resultScalar, ok := result.(*node.ScalarNode)
		if !ok {
			t.Fatal("Result should be a scalar node")
		}

		if resultScalar.Value != "test" {
			t.Errorf("Value should be preserved, got %s", resultScalar.Value)
		}
	})
}

func TestNodeProcessor_PreserveMetadata(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("preserve comments", func(t *testing.T) {
		opts := &Options{PreserveComments: true, PreserveBlankLines: false}

		// Create base node with comments
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.HeadComment = &node.CommentGroup{Comments: []string{"base head"}}
		baseScalar.LineComment = &node.CommentGroup{Comments: []string{"base line"}}
		baseScalar.FootComment = &node.CommentGroup{Comments: []string{"base foot"}}

		// Create result node without comments
		resultScalar := builder.BuildScalar("result", node.StylePlain)

		processor.PreserveMetadata(resultScalar, baseScalar, opts)

		if resultScalar.HeadComment == nil || len(resultScalar.HeadComment.Comments) == 0 {
			t.Error("Head comment should be preserved")
		}

		if resultScalar.LineComment == nil || len(resultScalar.LineComment.Comments) == 0 {
			t.Error("Line comment should be preserved")
		}

		if resultScalar.FootComment == nil || len(resultScalar.FootComment.Comments) == 0 {
			t.Error("Foot comment should be preserved")
		}
	})

	t.Run("preserve blank lines", func(t *testing.T) {
		opts := &Options{PreserveComments: false, PreserveBlankLines: true}

		// Create base node with blank lines
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.BlankLinesBefore = 2
		baseScalar.BlankLinesAfter = 1

		// Create result node
		resultScalar := builder.BuildScalar("result", node.StylePlain)

		processor.PreserveMetadata(resultScalar, baseScalar, opts)

		if resultScalar.BlankLinesBefore != 2 {
			t.Errorf("BlankLinesBefore should be 2, got %d", resultScalar.BlankLinesBefore)
		}

		if resultScalar.BlankLinesAfter != 1 {
			t.Errorf("BlankLinesAfter should be 1, got %d", resultScalar.BlankLinesAfter)
		}
	})

	t.Run("don't override existing comments", func(t *testing.T) {
		opts := &Options{PreserveComments: true}

		// Create base node with comments
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.HeadComment = &node.CommentGroup{Comments: []string{"base head"}}

		// Create result node with existing comments
		resultScalar := builder.BuildScalar("result", node.StylePlain)
		resultScalar.HeadComment = &node.CommentGroup{Comments: []string{"existing head"}}

		processor.PreserveMetadata(resultScalar, baseScalar, opts)

		// Existing comment should be preserved
		if len(resultScalar.HeadComment.Comments) == 0 || resultScalar.HeadComment.Comments[0] != "existing head" {
			t.Error("Existing comments should not be overridden")
		}
	})

	t.Run("mapping nodes", func(t *testing.T) {
		opts := &Options{PreserveComments: true}

		// Create base mapping with comments
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		baseMapping.HeadComment = &node.CommentGroup{Comments: []string{"mapping comment"}}

		// Create result mapping
		resultMapping := builder.BuildMapping(nil, node.StyleBlock)

		processor.PreserveMetadata(resultMapping, baseMapping, opts)

		if resultMapping.HeadComment == nil || len(resultMapping.HeadComment.Comments) == 0 {
			t.Error("Mapping comments should be preserved")
		}
	})

	t.Run("sequence nodes", func(t *testing.T) {
		opts := &Options{PreserveComments: true}

		// Create base sequence with comments
		baseSequence := builder.BuildSequence(nil, node.StyleBlock)
		baseSequence.HeadComment = &node.CommentGroup{Comments: []string{"sequence comment"}}

		// Create result sequence
		resultSequence := builder.BuildSequence(nil, node.StyleBlock)

		processor.PreserveMetadata(resultSequence, baseSequence, opts)

		if resultSequence.HeadComment == nil || len(resultSequence.HeadComment.Comments) == 0 {
			t.Error("Sequence comments should be preserved")
		}
	})

	t.Run("mismatched node types", func(t *testing.T) {
		opts := &Options{PreserveComments: true}

		// Create base scalar
		baseScalar := builder.BuildScalar("base", node.StylePlain)
		baseScalar.HeadComment = &node.CommentGroup{Comments: []string{"base comment"}}

		// Create result mapping (different type)
		resultMapping := builder.BuildMapping(nil, node.StyleBlock)

		// This should not panic
		processor.PreserveMetadata(resultMapping, baseScalar, opts)
		if resultMapping.HeadComment != nil {
			t.Error("Comments should not be transferred between different node types")
		}
	})
}

func TestNodeProcessor_CreateMappingPair(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("with base pair", func(t *testing.T) {
		opts := &Options{PreserveComments: true, PreserveBlankLines: true}

		key := builder.BuildScalar("key", node.StylePlain)
		value := builder.BuildScalar("value", node.StylePlain)

		// Create base pair with metadata
		basePair := &node.MappingPair{
			Key:              key,
			Value:            value,
			KeyComment:       &node.CommentGroup{Comments: []string{"key comment"}},
			ValueComment:     &node.CommentGroup{Comments: []string{"value comment"}},
			BlankLinesBefore: 1,
			BlankLinesAfter:  2,
		}

		newKey := builder.BuildScalar("newkey", node.StylePlain)
		newValue := builder.BuildScalar("newvalue", node.StylePlain)

		result := processor.CreateMappingPair(newKey, newValue, basePair, opts)

		if result.Key != newKey {
			t.Error("Key should be the new key")
		}

		if result.Value != newValue {
			t.Error("Value should be the new value")
		}

		if result.KeyComment == nil || len(result.KeyComment.Comments) == 0 {
			t.Error("Key comment should be preserved")
		}

		if result.ValueComment == nil || len(result.ValueComment.Comments) == 0 {
			t.Error("Value comment should be preserved")
		}

		if result.BlankLinesBefore != 1 {
			t.Errorf("BlankLinesBefore should be 1, got %d", result.BlankLinesBefore)
		}

		if result.BlankLinesAfter != 2 {
			t.Errorf("BlankLinesAfter should be 2, got %d", result.BlankLinesAfter)
		}
	})

	t.Run("without base pair", func(t *testing.T) {
		opts := &Options{PreserveComments: true, PreserveBlankLines: true}

		key := builder.BuildScalar("key", node.StylePlain)
		value := builder.BuildScalar("value", node.StylePlain)

		result := processor.CreateMappingPair(key, value, nil, opts)

		if result.Key != key {
			t.Error("Key should be preserved")
		}

		if result.Value != value {
			t.Error("Value should be preserved")
		}

		if result.KeyComment != nil {
			t.Error("Key comment should be nil when no base pair")
		}

		if result.ValueComment != nil {
			t.Error("Value comment should be nil when no base pair")
		}

		if result.BlankLinesBefore != 0 {
			t.Error("BlankLinesBefore should be 0 when no base pair")
		}

		if result.BlankLinesAfter != 0 {
			t.Error("BlankLinesAfter should be 0 when no base pair")
		}
	})

	t.Run("preserve comments disabled", func(t *testing.T) {
		opts := &Options{PreserveComments: false, PreserveBlankLines: false}

		key := builder.BuildScalar("key", node.StylePlain)
		value := builder.BuildScalar("value", node.StylePlain)

		basePair := &node.MappingPair{
			Key:              key,
			Value:            value,
			KeyComment:       &node.CommentGroup{Comments: []string{"key comment"}},
			ValueComment:     &node.CommentGroup{Comments: []string{"value comment"}},
			BlankLinesBefore: 1,
			BlankLinesAfter:  2,
		}

		newKey := builder.BuildScalar("newkey", node.StylePlain)
		newValue := builder.BuildScalar("newvalue", node.StylePlain)

		result := processor.CreateMappingPair(newKey, newValue, basePair, opts)

		if result.KeyComment != nil {
			t.Error("Key comment should not be preserved when disabled")
		}

		if result.ValueComment != nil {
			t.Error("Value comment should not be preserved when disabled")
		}

		if result.BlankLinesBefore != 0 {
			t.Error("BlankLinesBefore should not be preserved when disabled")
		}

		if result.BlankLinesAfter != 0 {
			t.Error("BlankLinesAfter should not be preserved when disabled")
		}
	})
}

func TestNodeProcessor_GetScalarValue(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("scalar node", func(t *testing.T) {
		scalar := builder.BuildScalar("test value", node.StylePlain)

		value, ok := processor.GetScalarValue(scalar)

		if !ok {
			t.Error("Should return true for scalar node")
		}

		if value != "test value" {
			t.Errorf("Expected 'test value', got '%s'", value)
		}
	})

	t.Run("non-scalar node", func(t *testing.T) {
		mapping := builder.BuildMapping(nil, node.StyleBlock)

		value, ok := processor.GetScalarValue(mapping)

		if ok {
			t.Error("Should return false for non-scalar node")
		}

		if value != "" {
			t.Errorf("Expected empty string, got '%s'", value)
		}
	})

	t.Run("nil node", func(t *testing.T) {
		value, ok := processor.GetScalarValue(nil)

		if ok {
			t.Error("Should return false for nil node")
		}

		if value != "" {
			t.Errorf("Expected empty string, got '%s'", value)
		}
	})
}

func TestNodeProcessor_PreserveKeyNode(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("preserve comments enabled", func(t *testing.T) {
		opts := &Options{PreserveComments: true}

		// Create base key with comments
		baseScalar := builder.BuildScalar("key", node.StylePlain)
		baseScalar.HeadComment = &node.CommentGroup{Comments: []string{"head comment"}}
		baseScalar.LineComment = &node.CommentGroup{Comments: []string{"line comment"}}
		baseScalar.FootComment = &node.CommentGroup{Comments: []string{"foot comment"}}

		// Create override key
		overrideKey := builder.BuildScalar("key", node.StylePlain)

		result := processor.PreserveKeyNode(baseScalar, overrideKey, opts)

		resultScalar, ok := result.(*node.ScalarNode)
		if !ok {
			t.Fatal("Result should be a scalar node")
		}

		if resultScalar.HeadComment == nil || len(resultScalar.HeadComment.Comments) == 0 {
			t.Error("Head comment should be preserved")
		}

		if resultScalar.LineComment == nil || len(resultScalar.LineComment.Comments) == 0 {
			t.Error("Line comment should be preserved")
		}

		if resultScalar.FootComment == nil || len(resultScalar.FootComment.Comments) == 0 {
			t.Error("Foot comment should be preserved")
		}
	})

	t.Run("preserve comments disabled", func(t *testing.T) {
		opts := &Options{PreserveComments: false}

		baseKey := builder.BuildScalar("key", node.StylePlain)
		overrideKey := builder.BuildScalar("key", node.StylePlain)

		result := processor.PreserveKeyNode(baseKey, overrideKey, opts)

		if result != baseKey {
			t.Error("Should return base key when comments are disabled")
		}
	})

	t.Run("non-scalar keys", func(t *testing.T) {
		opts := &Options{PreserveComments: true}

		baseKey := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("key", node.StylePlain)

		result := processor.PreserveKeyNode(baseKey, overrideKey, opts)

		if result != baseKey {
			t.Error("Should return base key for non-scalar keys")
		}
	})
}

func TestNodeProcessor_TransferInterFieldComments(t *testing.T) {
	processor := NewNodeProcessor()
	builder := node.NewBuilder()

	t.Run("transfer comments", func(t *testing.T) {
		// Create pairs with inter-field comments
		key1 := builder.BuildScalar("key1", node.StylePlain)
		value1Scalar := builder.BuildScalar("value1", node.StylePlain)
		value1Scalar.HeadComment = &node.CommentGroup{Comments: []string{"comment for next field"}}

		key2Scalar := builder.BuildScalar("key2", node.StylePlain)
		value2 := builder.BuildScalar("value2", node.StylePlain)

		pairs := []*node.MappingPair{
			{Key: key1, Value: value1Scalar},
			{Key: key2Scalar, Value: value2},
		}

		processor.TransferInterFieldComments(pairs)

		// Comment should be transferred from value1 to key2
		if value1Scalar.HeadComment != nil {
			t.Error("Comment should be removed from source value")
		}
		if key2Scalar.HeadComment == nil || len(key2Scalar.HeadComment.Comments) == 0 {
			t.Error("Comment should be transferred to next key")
		}

		if key2Scalar.HeadComment.Comments[0] != "comment for next field" {
			t.Error("Comment content should match")
		}
	})

	t.Run("don't override existing comments", func(t *testing.T) {
		// Create pairs with existing comment on target key
		key1 := builder.BuildScalar("key1", node.StylePlain)
		value1Scalar := builder.BuildScalar("value1", node.StylePlain)
		value1Scalar.HeadComment = &node.CommentGroup{Comments: []string{"comment to transfer"}}

		key2Scalar := builder.BuildScalar("key2", node.StylePlain)
		key2Scalar.HeadComment = &node.CommentGroup{Comments: []string{"existing comment"}}
		value2 := builder.BuildScalar("value2", node.StylePlain)

		pairs := []*node.MappingPair{
			{Key: key1, Value: value1Scalar},
			{Key: key2Scalar, Value: value2},
		}

		processor.TransferInterFieldComments(pairs)

		// Comment should not be transferred, existing should remain
		if value1Scalar.HeadComment == nil {
			t.Error("Original comment should remain when target has existing comment")
		}

		if key2Scalar.HeadComment.Comments[0] != "existing comment" {
			t.Error("Existing comment should not be overridden")
		}
	})

	t.Run("empty pairs", func(t *testing.T) {
		pairs := []*node.MappingPair{}

		// Should not panic
		processor.TransferInterFieldComments(pairs)
	})

	t.Run("single pair", func(t *testing.T) {
		key1 := builder.BuildScalar("key1", node.StylePlain)
		value1 := builder.BuildScalar("value1", node.StylePlain)

		pairs := []*node.MappingPair{
			{Key: key1, Value: value1},
		}

		// Should not panic
		processor.TransferInterFieldComments(pairs)
	})

	t.Run("non-scalar nodes", func(t *testing.T) {
		key1 := builder.BuildScalar("key1", node.StylePlain)
		value1 := builder.BuildMapping(nil, node.StyleBlock) // Non-scalar value

		key2 := builder.BuildScalar("key2", node.StylePlain)
		value2 := builder.BuildScalar("value2", node.StylePlain)

		pairs := []*node.MappingPair{
			{Key: key1, Value: value1},
			{Key: key2, Value: value2},
		}

		// Should not panic with non-scalar values
		processor.TransferInterFieldComments(pairs)
	})
}
