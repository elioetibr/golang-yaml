package merge

import (
	"errors"
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestNewMerger(t *testing.T) {
	t.Run("with default options", func(t *testing.T) {
		merger := NewMerger(nil)

		if merger == nil {
			t.Fatal("NewMerger(nil) should not return nil")
		}

		if merger.options == nil {
			t.Error("Merger should have options set")
		}

		if merger.strategy == nil {
			t.Error("Merger should have strategy set")
		}

		if merger.processor == nil {
			t.Error("Merger should have processor set")
		}

		// Should default to deep merge strategy
		if merger.options.Strategy != StrategyDeep {
			t.Errorf("Default strategy should be StrategyDeep, got %v", merger.options.Strategy)
		}
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := &Options{
			Strategy:           StrategyShallow,
			PreserveComments:   false,
			PreserveBlankLines: false,
		}

		merger := NewMerger(opts)

		if merger.options != opts {
			t.Error("Merger should use provided options")
		}

		if merger.options.Strategy != StrategyShallow {
			t.Errorf("Strategy should be StrategyShallow, got %v", merger.options.Strategy)
		}
	})

	t.Run("strategy selection", func(t *testing.T) {
		tests := []struct {
			name     string
			strategy Strategy
		}{
			{"StrategyDeep", StrategyDeep},
			{"StrategyShallow", StrategyShallow},
			{"StrategyOverride", StrategyOverride},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				opts := &Options{Strategy: tt.strategy}
				merger := NewMerger(opts)

				if merger.strategy == nil {
					t.Error("Strategy should be set")
				}

				// We can't directly test the strategy type without exposing it,
				// but we can verify it was set and is not nil
			})
		}
	})

	t.Run("unknown strategy defaults to deep", func(t *testing.T) {
		opts := &Options{Strategy: Strategy(999)} // Invalid strategy
		merger := NewMerger(opts)

		if merger.strategy == nil {
			t.Error("Strategy should be set even for unknown values")
		}
	})
}

func TestMergerMerge(t *testing.T) {
	t.Run("simple scalar merge", func(t *testing.T) {
		merger := NewMerger(DefaultOptions())
		builder := node.NewBuilder()

		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := merger.Merge(base, override)
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		if scalar, ok := result.(*node.ScalarNode); ok {
			if scalar.Value != "override" {
				t.Errorf("Expected 'override', got '%s'", scalar.Value)
			}
		} else {
			t.Error("Result should be a scalar node")
		}
	})

	t.Run("mapping merge", func(t *testing.T) {
		merger := NewMerger(DefaultOptions())
		builder := node.NewBuilder()

		// Create base mapping
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		baseKey := builder.BuildScalar("key1", node.StylePlain)
		baseValue := builder.BuildScalar("value1", node.StylePlain)
		baseMapping.Pairs = []*node.MappingPair{{Key: baseKey, Value: baseValue}}

		// Create override mapping
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideKey := builder.BuildScalar("key2", node.StylePlain)
		overrideValue := builder.BuildScalar("value2", node.StylePlain)
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideKey, Value: overrideValue}}

		result, err := merger.Merge(baseMapping, overrideMapping)
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		if mapping, ok := result.(*node.MappingNode); ok {
			if len(mapping.Pairs) != 2 {
				t.Errorf("Expected 2 pairs, got %d", len(mapping.Pairs))
			}
		} else {
			t.Error("Result should be a mapping node")
		}
	})

	t.Run("error handling", func(t *testing.T) {
		// Create a merger with a mock strategy that returns an error
		merger := &Merger{
			options:   DefaultOptions(),
			processor: NewNodeProcessor(),
			strategy:  &mockStrategy{shouldError: true},
		}

		builder := node.NewBuilder()
		base := builder.BuildScalar("base", node.StylePlain)
		override := builder.BuildScalar("override", node.StylePlain)

		result, err := merger.Merge(base, override)
		if err == nil {
			t.Error("Expected error from merge")
		}

		if result != nil {
			t.Error("Result should be nil when error occurs")
		}

		if !strings.Contains(err.Error(), "merge failed") {
			t.Errorf("Error should contain 'merge failed', got: %v", err)
		}
	})
}

func TestContext(t *testing.T) {
	t.Run("context creation", func(t *testing.T) {
		opts := DefaultOptions()
		ctx := &Context{
			Options: opts,
			Depth:   0,
			Path:    []string{},
		}

		if ctx.Options != opts {
			t.Error("Context should have correct options")
		}

		if ctx.Depth != 0 {
			t.Error("Context should have depth 0")
		}

		if len(ctx.Path) != 0 {
			t.Error("Context should have empty path")
		}
	})

	t.Run("context with path", func(t *testing.T) {
		opts := DefaultOptions()
		originalCtx := &Context{
			Options: opts,
			Depth:   1,
			Path:    []string{"root"},
		}

		newCtx := originalCtx.WithPath("child")

		// Verify original context is unchanged
		if len(originalCtx.Path) != 1 {
			t.Error("Original context should be unchanged")
		}

		if originalCtx.Depth != 1 {
			t.Error("Original context depth should be unchanged")
		}

		// Verify new context
		if newCtx.Depth != 2 {
			t.Errorf("New context depth should be 2, got %d", newCtx.Depth)
		}

		if len(newCtx.Path) != 2 {
			t.Errorf("New context path should have 2 elements, got %d", len(newCtx.Path))
		}

		if newCtx.Path[0] != "root" {
			t.Errorf("New context path[0] should be 'root', got '%s'", newCtx.Path[0])
		}

		if newCtx.Path[1] != "child" {
			t.Errorf("New context path[1] should be 'child', got '%s'", newCtx.Path[1])
		}

		if newCtx.Options != opts {
			t.Error("New context should share options")
		}
	})

	t.Run("multiple path segments", func(t *testing.T) {
		opts := DefaultOptions()
		ctx := &Context{
			Options: opts,
			Depth:   0,
			Path:    []string{},
		}

		ctx1 := ctx.WithPath("level1")
		ctx2 := ctx1.WithPath("level2")
		ctx3 := ctx2.WithPath("level3")

		if ctx3.Depth != 3 {
			t.Errorf("Context depth should be 3, got %d", ctx3.Depth)
		}

		expectedPath := []string{"level1", "level2", "level3"}
		if len(ctx3.Path) != len(expectedPath) {
			t.Errorf("Path length should be %d, got %d", len(expectedPath), len(ctx3.Path))
		}

		for i, expected := range expectedPath {
			if ctx3.Path[i] != expected {
				t.Errorf("Path[%d] should be '%s', got '%s'", i, expected, ctx3.Path[i])
			}
		}
	})

	t.Run("empty path segment", func(t *testing.T) {
		opts := DefaultOptions()
		ctx := &Context{
			Options: opts,
			Depth:   0,
			Path:    []string{"existing"},
		}

		newCtx := ctx.WithPath("")

		if len(newCtx.Path) != 2 {
			t.Errorf("Path should have 2 elements, got %d", len(newCtx.Path))
		}

		if newCtx.Path[1] != "" {
			t.Errorf("Second path element should be empty string, got '%s'", newCtx.Path[1])
		}
	})
}

// mockStrategy is a test helper that implements MergeStrategy
type mockStrategy struct {
	shouldError bool
	result      node.Node
}

func (m *mockStrategy) Name() string {
	return "mock"
}

func (m *mockStrategy) Merge(base, override node.Node, ctx *Context) (node.Node, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	if m.result != nil {
		return m.result, nil
	}
	return override, nil
}

func TestMergerIntegration(t *testing.T) {
	t.Run("deep merge integration", func(t *testing.T) {
		opts := DefaultOptions().WithStrategy(StrategyDeep)
		merger := NewMerger(opts)
		builder := node.NewBuilder()

		// Create nested structure
		baseMapping := builder.BuildMapping(nil, node.StyleBlock)
		configKey := builder.BuildScalar("config", node.StylePlain)
		configMapping := builder.BuildMapping(nil, node.StyleBlock)
		dbKey := builder.BuildScalar("database", node.StylePlain)
		dbValue := builder.BuildScalar("mysql", node.StylePlain)
		configMapping.Pairs = []*node.MappingPair{{Key: dbKey, Value: dbValue}}
		baseMapping.Pairs = []*node.MappingPair{{Key: configKey, Value: configMapping}}

		// Create override structure
		overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
		overrideConfigKey := builder.BuildScalar("config", node.StylePlain)
		overrideConfigMapping := builder.BuildMapping(nil, node.StyleBlock)
		portKey := builder.BuildScalar("port", node.StylePlain)
		portValue := builder.BuildScalar("3306", node.StylePlain)
		overrideConfigMapping.Pairs = []*node.MappingPair{{Key: portKey, Value: portValue}}
		overrideMapping.Pairs = []*node.MappingPair{{Key: overrideConfigKey, Value: overrideConfigMapping}}

		result, err := merger.Merge(baseMapping, overrideMapping)
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		// Verify deep merge preserved both database and port
		if mapping, ok := result.(*node.MappingNode); ok {
			if len(mapping.Pairs) != 1 {
				t.Fatalf("Expected 1 top-level pair, got %d", len(mapping.Pairs))
			}

			if configPair := mapping.Pairs[0]; configPair != nil {
				if configNode, ok := configPair.Value.(*node.MappingNode); ok {
					if len(configNode.Pairs) != 2 {
						t.Errorf("Expected 2 config pairs, got %d", len(configNode.Pairs))
					}
				} else {
					t.Error("Config value should be a mapping node")
				}
			}
		} else {
			t.Error("Result should be a mapping node")
		}
	})
}
