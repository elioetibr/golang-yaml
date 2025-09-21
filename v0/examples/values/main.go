package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	merge2 "github.com/elioetibr/golang-yaml/v0/pkg/merge"
)

func main() {
	// Get directory of main.go
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// File paths
	valuesFile := filepath.Join(dir, "base-helm-chart-values.yaml")
	valuesWithMergeContentFile := filepath.Join(dir, "legacy-helm-chart-values.yaml")
	valuesOverriddenFile := filepath.Join(dir, "values.yaml")

	// Use the merge package with options to NOT preserve blank lines
	opts := merge2.DefaultOptions()
	opts.PreserveBlankLines = false
	err := merge2.MergeFilesToFileWithOptions(valuesFile, valuesWithMergeContentFile, valuesOverriddenFile, opts)
	if err != nil {
		log.Fatalf("Error merging files: %v", err)
	}

	fmt.Println("Successfully merged YAML files and wrote to values.yaml")
	fmt.Println("✅ Comments and formatting are preserved by default!")

	// Demonstrate different merge strategies
	fmt.Println("\n--- Demonstrating Different Strategies ---")

	// Deep merge (default)
	fmt.Println("\n1. Deep Merge (default):")
	deepResult, _ := merge2.MergeFilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge2.DefaultOptions().WithStrategy(merge2.StrategyDeep))
	printFirstLines(deepResult, 5)

	// Shallow merge
	fmt.Println("\n2. Shallow Merge:")
	shallowResult, _ := merge2.MergeFilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge2.DefaultOptions().WithStrategy(merge2.StrategyShallow))
	printFirstLines(shallowResult, 5)

	// Override strategy
	fmt.Println("\n3. Override Strategy:")
	overrideResult, _ := merge2.MergeFilesWithOptions(valuesFile, valuesWithMergeContentFile,
		merge2.DefaultOptions().WithStrategy(merge2.StrategyOverride))
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
