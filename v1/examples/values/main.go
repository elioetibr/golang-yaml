package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/elioetibr/golang-yaml/v1/pkg/merge"
)

func main() {
	// Get directory of main.go
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// File paths
	valuesFile := filepath.Join(dir, "base-helm-chart-values.yaml")
	valuesWithMergeContentFile := filepath.Join(dir, "legacy-helm-chart-values.yaml")
	valuesOverriddenFile := filepath.Join(dir, "values.yaml")

	fmt.Println("=== YAML Merge v1 - Advanced Features Demo ===")
	fmt.Println()

	// Example 1: Default merge with all preservation features
	fmt.Println("📋 Example 1: Default Configuration")
	fmt.Println("   • Preserves all comments")
	fmt.Println("   • Preserves blank lines")
	fmt.Println("   • Keeps original section spacing")
	fmt.Println("   • Document head comments preserved")
	fmt.Println()

	opts := merge.DefaultOptions()
	err := merge.FilesToFileWithOptions(valuesFile, valuesWithMergeContentFile, valuesOverriddenFile, opts)
	if err != nil {
		log.Fatalf("Error merging files: %v", err)
	}
	fmt.Println("   ✅ Successfully merged to values.yaml")

	// Example 2: Normalize section spacing
	fmt.Println()
	fmt.Println("📋 Example 2: Normalized Section Spacing")
	fmt.Println("   • All sections normalized to 1 blank line")
	fmt.Println("   • Comments still preserved")
	fmt.Println()

	normalizedOpts := merge.DefaultOptions().WithNormalizedSections(1)
	_, err = merge.FilesWithOptions(valuesFile, valuesWithMergeContentFile, normalizedOpts)
	if err != nil {
		log.Printf("   ❌ Error: %v", err)
	} else {
		fmt.Println("   ✅ Sections normalized to consistent spacing")
	}

	// Example 3: Custom section handling
	fmt.Println()
	fmt.Println("📋 Example 3: Custom Section Configuration")
	fmt.Println("   • Sections with >1 blank line normalized to 2 lines")
	fmt.Println("   • Document structure maintained")
	fmt.Println()

	customOpts := merge.DefaultOptions().WithSectionHandling(false, 2)
	_, err = merge.FilesWithOptions(valuesFile, valuesWithMergeContentFile, customOpts)
	if err != nil {
		log.Printf("   ❌ Error: %v", err)
	} else {
		fmt.Println("   ✅ Custom section spacing applied")
	}

	// Example 4: No blank line preservation
	fmt.Println()
	fmt.Println("📋 Example 4: Compact Format (No Blank Lines)")
	fmt.Println("   • All blank lines removed")
	fmt.Println("   • Comments still preserved")
	fmt.Println()

	compactOpts := merge.DefaultOptions()
	compactOpts.PreserveBlankLines = false
	_, err = merge.FilesWithOptions(valuesFile, valuesWithMergeContentFile, compactOpts)
	if err != nil {
		log.Printf("   ❌ Error: %v", err)
	} else {
		fmt.Println("   ✅ Compact format applied")
	}

	// Demonstrate different merge strategies
	fmt.Println("\n--- Merge Strategy Comparison ---")

	// Deep merge (default)
	fmt.Println("\n1. Deep Merge (default):")
	deepResult, _ := merge.FilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge.DefaultOptions().WithStrategy(merge.StrategyDeep))
	printFirstLines(deepResult, 5)

	// Shallow merge
	fmt.Println("\n2. Shallow Merge:")
	shallowResult, _ := merge.FilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge.DefaultOptions().WithStrategy(merge.StrategyShallow))
	printFirstLines(shallowResult, 5)

	// Override strategy
	fmt.Println("\n3. Override Strategy:")
	overrideResult, _ := merge.FilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge.DefaultOptions().WithStrategy(merge.StrategyOverride))
	printFirstLines(overrideResult, 5)
}

// Helper function to print first N lines
func printFirstLines(text string, n int) {
	lines := 0
	for i, ch := range text {
		if ch == '\n' {
			lines++
			if lines >= n {
				fmt.Println(text[:i])
				fmt.Println("...")
				return
			}
		}
	}
	fmt.Print(text)
}
