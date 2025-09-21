package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/merge"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
	"github.com/elioetibr/golang-yaml/v1/pkg/parser"
	"github.com/elioetibr/golang-yaml/v1/pkg/serializer"
)

func main() {
	fmt.Println("=== Advanced YAML Processing with v1 ===")
	fmt.Println()

	// Example YAML demonstrating sections with different spacing
	// The parser currently processes mappings correctly
	baseYAML := `app:
  name: my-app
  version: 1.0.0

database:
  host: localhost
  port: 5432


features:
  debug: false
  metrics: true`

	overrideYAML := `app:
  name: my-app-prod
  version: 2.0.0
database:
  port: 3306
features:
  debug: true
  experimental: true`

	// Parse base YAML
	fmt.Println("📝 Parsing YAML documents...")
	baseLexer := lexer.NewLexerFromString(baseYAML)
	baseParser := parser.NewParser(baseLexer)
	baseNode, err := baseParser.Parse()
	if err != nil {
		log.Fatalf("Failed to parse base YAML: %v", err)
	}

	// Parse override YAML
	overrideLexer := lexer.NewLexerFromString(overrideYAML)
	overrideParser := parser.NewParser(overrideLexer)
	overrideNode, err := overrideParser.Parse()
	if err != nil {
		log.Fatalf("Failed to parse override YAML: %v", err)
	}

	// Example 1: Preserve everything (default)
	fmt.Println("\n✨ Example 1: Full Preservation")
	fmt.Println("   Keeping all blank lines and original section spacing")
	fmt.Println("   Notice: 'database' has 2 blank lines before it in base")
	fmt.Println()

	opts1 := merge.DefaultOptions()
	merged1, err := merge.WithOptions(baseNode, overrideNode, opts1)
	if err != nil {
		log.Fatalf("Merge failed: %v", err)
	}

	result1 := serializeNode(merged1)
	fmt.Println("First few lines of result:")
	printFirstLines(result1, 10)

	// Example 2: Normalize sections to 1 blank line
	fmt.Println("\n✨ Example 2: Normalized Sections")
	fmt.Println("   Normalizing all section separators to 1 blank line")
	fmt.Println()

	opts2 := merge.DefaultOptions().WithNormalizedSections(1)
	merged2, err := merge.WithOptions(baseNode, overrideNode, opts2)
	if err != nil {
		log.Fatalf("Merge failed: %v", err)
	}

	// Check if the merged node has document head comments awareness
	if mapping, ok := merged2.(*node.MappingNode); ok && mapping.HasDocumentHeadComments {
		fmt.Println("   ℹ️  Document-level comments preserved!")
	}

	result2 := serializeNode(merged2)
	fmt.Println("First few lines of result:")
	printFirstLines(result2, 10)

	// Example 3: Compact format (no blank lines)
	fmt.Println("\n✨ Example 3: Compact Format")
	fmt.Println("   Removing all blank lines for a compact output")
	fmt.Println()

	opts3 := &merge.Options{
		Strategy:           merge.StrategyDeep,
		PreserveComments:   true,
		PreserveBlankLines: false, // No blank lines
	}

	merged3, err := merge.WithOptions(baseNode, overrideNode, opts3)
	if err != nil {
		log.Fatalf("Merge failed: %v", err)
	}

	result3 := serializeNode(merged3)
	fmt.Println("First few lines of result:")
	printFirstLines(result3, 10)

	// Demonstrate section detection
	fmt.Println("\n🔍 Section Detection Example")
	processor := merge.NewNodeProcessor()

	// Create a test node with blank lines
	builder := node.NewBuilder()
	testNode := builder.BuildScalar("test", node.StylePlain)
	testNode.BlankLinesBefore = 3

	if processor.IsSectionBoundary(testNode, 2) {
		fmt.Println("   ✓ Node with 3 blank lines detected as section boundary")
	}

	testNode.BlankLinesBefore = 1
	if !processor.IsSectionBoundary(testNode, 2) {
		fmt.Println("   ✓ Node with 1 blank line NOT detected as section boundary")
	}

	fmt.Println()
	fmt.Println("✅ Advanced features demonstration complete!")
}

// serializeNode converts a node to YAML string
func serializeNode(n node.Node) string {
	var buf bytes.Buffer
	s := serializer.NewSerializer(&buf, nil)
	err := s.Serialize(n)
	if err != nil {
		log.Printf("Serialization error: %v", err)
		return ""
	}
	return buf.String()
}

// printFirstLines prints the first n lines of text
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
