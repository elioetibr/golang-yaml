package merge

import (
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Strategy != StrategyDeep {
		t.Errorf("DefaultOptions().Strategy = %v, want %v", opts.Strategy, StrategyDeep)
	}

	if !opts.PreserveComments {
		t.Error("DefaultOptions().PreserveComments should be true")
	}

	if !opts.PreserveBlankLines {
		t.Error("DefaultOptions().PreserveBlankLines should be true")
	}

	if opts.ArrayMergeStrategy != ArrayReplace {
		t.Errorf("DefaultOptions().ArrayMergeStrategy = %v, want %v", opts.ArrayMergeStrategy, ArrayReplace)
	}

	if opts.OverrideEmpty {
		t.Error("DefaultOptions().OverrideEmpty should be false")
	}

	if !opts.MergeAnchors {
		t.Error("DefaultOptions().MergeAnchors should be true")
	}

	if opts.KeyPriority != KeyPriorityBase {
		t.Errorf("DefaultOptions().KeyPriority = %v, want %v", opts.KeyPriority, KeyPriorityBase)
	}

	if opts.CustomMergeFunc != nil {
		t.Error("DefaultOptions().CustomMergeFunc should be nil")
	}
}

func TestOptionsWithStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
	}{
		{"StrategyDeep", StrategyDeep},
		{"StrategyShallow", StrategyShallow},
		{"StrategyOverride", StrategyOverride},
		{"StrategyAppend", StrategyAppend},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions().WithStrategy(tt.strategy)
			if opts.Strategy != tt.strategy {
				t.Errorf("WithStrategy(%v) = %v, want %v", tt.strategy, opts.Strategy, tt.strategy)
			}
		})
	}
}

func TestOptionsWithArrayStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy ArrayMergeStrategy
	}{
		{"ArrayReplace", ArrayReplace},
		{"ArrayAppend", ArrayAppend},
		{"ArrayMergeByIndex", ArrayMergeByIndex},
		{"ArrayMergeByKey", ArrayMergeByKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions().WithArrayStrategy(tt.strategy)
			if opts.ArrayMergeStrategy != tt.strategy {
				t.Errorf("WithArrayStrategy(%v) = %v, want %v", tt.strategy, opts.ArrayMergeStrategy, tt.strategy)
			}
		})
	}
}

func TestOptionsWithKeyPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority KeyPriority
	}{
		{"KeyPriorityBase", KeyPriorityBase},
		{"KeyPriorityOverride", KeyPriorityOverride},
		{"KeyPriorityAlphabetical", KeyPriorityAlphabetical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions().WithKeyPriority(tt.priority)
			if opts.KeyPriority != tt.priority {
				t.Errorf("WithKeyPriority(%v) = %v, want %v", tt.priority, opts.KeyPriority, tt.priority)
			}
		})
	}
}

func TestOptionsWithOverrideEmpty(t *testing.T) {
	tests := []struct {
		name     string
		override bool
	}{
		{"OverrideEmpty_true", true},
		{"OverrideEmpty_false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions().WithOverrideEmpty(tt.override)
			if opts.OverrideEmpty != tt.override {
				t.Errorf("WithOverrideEmpty(%v) = %v, want %v", tt.override, opts.OverrideEmpty, tt.override)
			}
		})
	}
}

func TestOptionsChaining(t *testing.T) {
	// Test that options can be chained
	opts := DefaultOptions().
		WithStrategy(StrategyShallow).
		WithArrayStrategy(ArrayAppend).
		WithKeyPriority(KeyPriorityOverride).
		WithOverrideEmpty(true)

	if opts.Strategy != StrategyShallow {
		t.Errorf("Chained Strategy = %v, want %v", opts.Strategy, StrategyShallow)
	}

	if opts.ArrayMergeStrategy != ArrayAppend {
		t.Errorf("Chained ArrayMergeStrategy = %v, want %v", opts.ArrayMergeStrategy, ArrayAppend)
	}

	if opts.KeyPriority != KeyPriorityOverride {
		t.Errorf("Chained KeyPriority = %v, want %v", opts.KeyPriority, KeyPriorityOverride)
	}

	if !opts.OverrideEmpty {
		t.Error("Chained OverrideEmpty should be true")
	}
}

func TestOptionsImmutability(t *testing.T) {
	// Test that the original options are not modified when chaining
	original := DefaultOptions()
	modified := original.WithStrategy(StrategyShallow)

	// Both should point to the same object since we're modifying in place
	if original != modified {
		t.Error("WithStrategy should return the same object reference")
	}

	// Verify the strategy was actually changed
	if original.Strategy != StrategyShallow {
		t.Error("Original options should have been modified")
	}
}

func TestCustomMergeFunc(t *testing.T) {
	opts := DefaultOptions()

	// Test setting a custom merge function
	customFunc := func(key string, base, override node.Node) (node.Node, bool) {
		if key == "special" {
			return override, true
		}
		return nil, false
	}

	opts.CustomMergeFunc = customFunc

	if opts.CustomMergeFunc == nil {
		t.Error("CustomMergeFunc should not be nil after setting")
	}

	// Test the custom function
	builder := node.NewBuilder()
	baseNode := builder.BuildScalar("base", node.StylePlain)
	overrideNode := builder.BuildScalar("override", node.StylePlain)

	result, handled := opts.CustomMergeFunc("special", baseNode, overrideNode)
	if !handled {
		t.Error("CustomMergeFunc should handle 'special' key")
	}
	if result != overrideNode {
		t.Error("CustomMergeFunc should return override node for 'special' key")
	}

	result, handled = opts.CustomMergeFunc("normal", baseNode, overrideNode)
	if handled {
		t.Error("CustomMergeFunc should not handle 'normal' key")
	}
	if result != nil {
		t.Error("CustomMergeFunc should return nil for unhandled keys")
	}
}

func TestStrategyConstants(t *testing.T) {
	// Test that strategy constants have expected values
	if StrategyDeep != 0 {
		t.Errorf("StrategyDeep = %d, want 0", StrategyDeep)
	}
	if StrategyShallow != 1 {
		t.Errorf("StrategyShallow = %d, want 1", StrategyShallow)
	}
	if StrategyOverride != 2 {
		t.Errorf("StrategyOverride = %d, want 2", StrategyOverride)
	}
	if StrategyAppend != 3 {
		t.Errorf("StrategyAppend = %d, want 3", StrategyAppend)
	}
}

func TestArrayMergeStrategyConstants(t *testing.T) {
	// Test that array merge strategy constants have expected values
	if ArrayReplace != 0 {
		t.Errorf("ArrayReplace = %d, want 0", ArrayReplace)
	}
	if ArrayAppend != 1 {
		t.Errorf("ArrayAppend = %d, want 1", ArrayAppend)
	}
	if ArrayMergeByIndex != 2 {
		t.Errorf("ArrayMergeByIndex = %d, want 2", ArrayMergeByIndex)
	}
	if ArrayMergeByKey != 3 {
		t.Errorf("ArrayMergeByKey = %d, want 3", ArrayMergeByKey)
	}
}

func TestKeyPriorityConstants(t *testing.T) {
	// Test that key priority constants have expected values
	if KeyPriorityBase != 0 {
		t.Errorf("KeyPriorityBase = %d, want 0", KeyPriorityBase)
	}
	if KeyPriorityOverride != 1 {
		t.Errorf("KeyPriorityOverride = %d, want 1", KeyPriorityOverride)
	}
	if KeyPriorityAlphabetical != 2 {
		t.Errorf("KeyPriorityAlphabetical = %d, want 2", KeyPriorityAlphabetical)
	}
}