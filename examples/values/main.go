package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/elioetibr/golang-yaml/pkg/merge"
)

func main() {
	// Get directory of main.go
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// File paths
	valuesFile := filepath.Join(dir, "base-helm-chart-values.yaml")
	valuesWithMergeContentFile := filepath.Join(dir, "legacy-helm-chart-values.yaml")
	valuesOverriddenFile := filepath.Join(dir, "values.yaml")

	// Use the merge package with default options
	// This preserves comments and blank lines by default
	err := merge.MergeFilesToFile(valuesFile, valuesWithMergeContentFile, valuesOverriddenFile)
	if err != nil {
		log.Fatalf("Error merging files: %v", err)
	}

	fmt.Println("Successfully merged YAML files and wrote to values.yaml")
	fmt.Println("✅ Comments and formatting are preserved by default!")

	// Demonstrate different merge strategies
	fmt.Println("\n--- Demonstrating Different Strategies ---")

	// Deep merge (default)
	fmt.Println("\n1. Deep Merge (default):")
	deepResult, _ := merge.MergeFilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge.DefaultOptions().WithStrategy(merge.StrategyDeep))
	printFirstLines(deepResult, 5)

	// Shallow merge
	fmt.Println("\n2. Shallow Merge:")
	shallowResult, _ := merge.MergeFilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge.DefaultOptions().WithStrategy(merge.StrategyShallow))
	printFirstLines(shallowResult, 5)

	// Override strategy
	fmt.Println("\n3. Override Strategy:")
	overrideResult, _ := merge.MergeFilesWithOptions(valuesFile, valuesWithMergeContentFile,
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
