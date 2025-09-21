package merge

import (
	"fmt"
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// TestMergeStrategyInterface tests the MergeStrategy interface behavior
func TestMergeStrategyInterface(t *testing.T) {
	t.Run("deep merge strategy implements interface", func(t *testing.T) {
		processor := NewNodeProcessor()
		strategy := NewDeepMergeStrategy(processor)

		var _ MergeStrategy = strategy

		if strategy == nil {
			t.Fatal("NewDeepMergeStrategy should not return nil")
		}

		name := strategy.Name()
		if name == "" {
			t.Error("Strategy should have a non-empty name")
		}

		// Test that Merge method exists and can be called
		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{
			Options: DefaultOptions(),
			Depth:   0,
			Path:    []string{},
		}

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Merge should not error for simple scalars: %v", err)
		}

		if result == nil {
			t.Error("Merge should return a result")
		}
	})

	t.Run("shallow merge strategy implements interface", func(t *testing.T) {
		processor := NewNodeProcessor()
		strategy := NewShallowMergeStrategy(processor)

		var _ MergeStrategy = strategy

		if strategy == nil {
			t.Fatal("NewShallowMergeStrategy should not return nil")
		}

		name := strategy.Name()
		if name == "" {
			t.Error("Strategy should have a non-empty name")
		}

		// Test that Merge method exists and can be called
		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{
			Options: DefaultOptions(),
			Depth:   0,
			Path:    []string{},
		}

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Merge should not error for simple scalars: %v", err)
		}

		if result == nil {
			t.Error("Merge should return a result")
		}
	})

	t.Run("override strategy implements interface", func(t *testing.T) {
		strategy := NewOverrideStrategy()

		var _ MergeStrategy = strategy

		if strategy == nil {
			t.Fatal("NewOverrideStrategy should not return nil")
		}

		name := strategy.Name()
		if name == "" {
			t.Error("Strategy should have a non-empty name")
		}

		// Test that Merge method exists and can be called
		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{
			Options: DefaultOptions(),
			Depth:   0,
			Path:    []string{},
		}

		result, err := strategy.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Merge should not error for simple scalars: %v", err)
		}

		if result == nil {
			t.Error("Merge should return a result")
		}
	})

	t.Run("strategy name uniqueness", func(t *testing.T) {
		processor := NewNodeProcessor()
		deepStrategy := NewDeepMergeStrategy(processor)
		shallowStrategy := NewShallowMergeStrategy(processor)
		overrideStrategy := NewOverrideStrategy()

		deepName := deepStrategy.Name()
		shallowName := shallowStrategy.Name()
		overrideName := overrideStrategy.Name()

		// Names should be different
		if deepName == shallowName {
			t.Error("Deep and shallow strategies should have different names")
		}

		if deepName == overrideName {
			t.Error("Deep and override strategies should have different names")
		}

		if shallowName == overrideName {
			t.Error("Shallow and override strategies should have different names")
		}
	})
}

// mockMergeStrategy is a test implementation of MergeStrategy
type mockMergeStrategy struct {
	name        string
	result      node.Node
	shouldError bool
	errorMsg    string
}

func (m *mockMergeStrategy) Name() string {
	return m.name
}

func (m *mockMergeStrategy) Merge(base, override node.Node, ctx *Context) (node.Node, error) {
	if m.shouldError {
		return nil, fmt.Errorf(m.errorMsg)
	}
	if m.result != nil {
		return m.result, nil
	}
	return override, nil
}

func TestMockMergeStrategy(t *testing.T) {
	t.Run("mock strategy basic functionality", func(t *testing.T) {
		mock := &mockMergeStrategy{
			name: "MockStrategy",
		}

		var _ MergeStrategy = mock

		if mock.Name() != "MockStrategy" {
			t.Errorf("Expected name 'MockStrategy', got '%s'", mock.Name())
		}

		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{Options: DefaultOptions()}

		result, err := mock.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Mock merge should not error: %v", err)
		}

		if result != override {
			t.Error("Mock should return override by default")
		}
	})

	t.Run("mock strategy with custom result", func(t *testing.T) {
		builder := node.NewBuilder()
		customResult := builder.BuildScalar("custom", node.StylePlain)

		mock := &mockMergeStrategy{
			name:   "CustomMockStrategy",
			result: customResult,
		}

		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{Options: DefaultOptions()}

		result, err := mock.Merge(base, override, ctx)
		if err != nil {
			t.Errorf("Mock merge should not error: %v", err)
		}

		if result != customResult {
			t.Error("Mock should return custom result when set")
		}
	})

	t.Run("mock strategy with error", func(t *testing.T) {
		mock := &mockMergeStrategy{
			name:        "ErrorMockStrategy",
			shouldError: true,
			errorMsg:    "test error",
		}

		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{Options: DefaultOptions()}

		result, err := mock.Merge(base, override, ctx)
		if err == nil {
			t.Error("Mock should return error when shouldError is true")
		}

		if result != nil {
			t.Error("Mock should return nil result when error occurs")
		}

		if err.Error() != "test error" {
			t.Errorf("Expected error message 'test error', got '%s'", err.Error())
		}
	})
}

func TestMergeStrategyIntegration(t *testing.T) {
	t.Run("strategies can be used interchangeably", func(t *testing.T) {
		processor := NewNodeProcessor()
		strategies := []MergeStrategy{
			NewDeepMergeStrategy(processor),
			NewShallowMergeStrategy(processor),
			NewOverrideStrategy(),
		}

		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)
		ctx := &Context{Options: DefaultOptions()}

		for _, strategy := range strategies {
			t.Run(strategy.Name(), func(t *testing.T) {
				result, err := strategy.Merge(base, override, ctx)
				if err != nil {
					t.Errorf("Strategy %s should not error: %v", strategy.Name(), err)
				}

				if result == nil {
					t.Errorf("Strategy %s should return a result", strategy.Name())
				}

				// All strategies should return the override for simple scalars
				if resultScalar, ok := result.(*node.ScalarNode); ok {
					if resultScalar.Value != "override" {
						t.Errorf("Strategy %s should return override value for scalars", strategy.Name())
					}
				}
			})
		}
	})
}