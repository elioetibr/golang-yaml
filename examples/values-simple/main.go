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
	valuesFile := filepath.Join(dir, "values-with-comments.yaml")
	valuesWithMergeContentFile := filepath.Join(dir, "values-with-comments-merge.yaml")
	valuesOverriddenFile := filepath.Join(dir, "values-with-comments-overridden.yaml")

	// Use the new merge package - it preserves comments and blank lines by default
	err := merge.MergeFilesToFile(valuesFile, valuesWithMergeContentFile, valuesOverriddenFile)
	if err != nil {
		log.Fatalf("Error merging files: %v", err)
	}

	fmt.Println("✅ Successfully merged YAML files using the new merge package!")
	fmt.Printf("   Base: %s\n", valuesFile)
	fmt.Printf("   Override: %s\n", valuesWithMergeContentFile)
	fmt.Printf("   Output: %s\n", valuesOverriddenFile)
	fmt.Println("\n📝 Comments and formatting are preserved by default!")
}
