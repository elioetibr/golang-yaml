package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/elioetibr/golang-yaml/pkg/merge"
)

func main() {
	fmt.Println("=== YAML Merge Feature Demo ===")

	// Get directory of main.go
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Demo 1: Simple merge with defaults
	fmt.Println("1. Simple Deep Merge (Default):")
	simpleMerge()

	// Demo 2: Different merge strategies
	fmt.Println("\n2. Merge Strategies Comparison:")
	strategyComparison()

	// Demo 3: Array merge strategies
	fmt.Println("\n3. Array Merge Strategies:")
	arrayMergeDemo()

	// Demo 4: File merging (using example files)
	fmt.Println("\n4. File Merging:")
	fileMergeDemo(dir)

	// Demo 5: Multiple document merging
	fmt.Println("\n5. Multiple Document Merging:")
	multiMergeDemo()

	// Demo 6: Comment and blank line preservation
	fmt.Println("\n6. Format Preservation:")
	formatPreservationDemo()

	fmt.Println("\n=== Demo Complete ===")
}

func simpleMerge() {
	baseYAML := `# Application config
name: myapp
version: 1.0.0
server:
  host: localhost
  port: 8080
database:
  host: localhost
  port: 5432`

	overrideYAML := `# Production overrides
server:
  host: api.example.com
  port: 443
database:
  host: db.example.com`

	result, err := merge.MergeStrings(baseYAML, overrideYAML)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Result:")
	fmt.Println(result)
}

func strategyComparison() {
	baseYAML := `
config:
  nested:
    value: base
    other: keep
  simple: base
list: [1, 2, 3]`

	overrideYAML := `
config:
  nested:
    value: override
  simple: override
  new: added
list: [4, 5]`

	// Deep merge
	deepResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
		merge.DefaultOptions().WithStrategy(merge.StrategyDeep))
	fmt.Println("Deep Merge:")
	fmt.Println(deepResult)

	// Shallow merge
	shallowResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
		merge.DefaultOptions().WithStrategy(merge.StrategyShallow))
	fmt.Println("\nShallow Merge:")
	fmt.Println(shallowResult)

	// Override strategy
	overrideResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
		merge.DefaultOptions().WithStrategy(merge.StrategyOverride))
	fmt.Println("\nOverride Strategy:")
	fmt.Println(overrideResult)
}

func arrayMergeDemo() {
	baseYAML := `
features:
  - auth
  - api
  - logging
settings:
  - name: timeout
    value: 30
  - name: retries
    value: 3`

	overrideYAML := `
features:
  - monitoring
  - metrics
settings:
  - name: timeout
    value: 60
  - name: cache
    value: true`

	// Replace arrays (default)
	replaceResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
		merge.DefaultOptions().WithArrayStrategy(merge.ArrayReplace))
	fmt.Println("Array Replace:")
	fmt.Println(replaceResult)

	// Append arrays
	appendResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
		merge.DefaultOptions().WithArrayStrategy(merge.ArrayAppend))
	fmt.Println("\nArray Append:")
	fmt.Println(appendResult)

	// Merge by index
	indexResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
		merge.DefaultOptions().WithArrayStrategy(merge.ArrayMergeByIndex))
	fmt.Println("\nArray Merge by Index:")
	fmt.Println(indexResult)
}

func fileMergeDemo(dir string) {
	// For demo, we'll use the values example if it exists
	valuesDir := filepath.Join(filepath.Dir(dir), "values")
	valuesFile := filepath.Join(valuesDir, "values.yaml")
	valuesMergeFile := filepath.Join(valuesDir, "values-merge.yaml")

	// Try to use values example files if they exist
	result, err := merge.MergeFiles(valuesFile, valuesMergeFile)
	if err != nil {
		// Fallback to simple demo
		fmt.Println("Using demo values (values example not found)")
		baseYAML := `name: app
version: 1.0.0`
		overrideYAML := `version: 2.0.0
env: production`
		result, _ = merge.MergeStrings(baseYAML, overrideYAML)
	} else {
		fmt.Println("Merged values.yaml with values-merge.yaml:")
	}

	// Show first few lines of result
	lines := splitLines(result)
	for i, line := range lines {
		if i >= 20 {
			fmt.Println("... (truncated)")
			break
		}
		fmt.Println(line)
	}
}

func multiMergeDemo() {
	yaml1 := `base: value1
common: base`

	yaml2 := `common: override1
middle: value2`

	yaml3 := `common: override2
final: value3`

	// Parse to nodes first
	nodes := make([]string, 3)
	nodes[0] = yaml1
	nodes[1] = yaml2
	nodes[2] = yaml3

	// Merge multiple documents in sequence
	result := yaml1
	for i := 1; i < len(nodes); i++ {
		merged, err := merge.MergeStrings(result, nodes[i])
		if err != nil {
			log.Fatal(err)
		}
		result = merged
	}

	fmt.Println("Merged 3 documents in sequence:")
	fmt.Println(result)
}

func formatPreservationDemo() {
	baseYAML := `# Main configuration file
# Version: 1.0

# Server configuration
server:
  # Host to bind to
  host: localhost  # Can be changed for production

  # Port to listen on
  port: 8080

# Database settings
database:
  driver: postgres
  host: localhost
  port: 5432
`

	overrideYAML := `# Production overrides

server:
  host: api.example.com  # Production host
  port: 443

database:
  host: db.prod.example.com
  ssl: true  # Enable SSL in production
`

	// Merge with comment preservation (default)
	result, err := merge.MergeStrings(baseYAML, overrideYAML)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Merged with preserved comments and formatting:")
	fmt.Println(result)

	// Merge without preserving comments
	opts := &merge.Options{
		Strategy:           merge.StrategyDeep,
		PreserveComments:   false,
		PreserveBlankLines: false,
		ArrayMergeStrategy: merge.ArrayReplace,
	}

	resultNoComments, err := merge.MergeStringsWithOptions(baseYAML, overrideYAML, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nMerged WITHOUT preserved comments:")
	fmt.Println(resultNoComments)
}

// Helper function to split string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
