package main

import (
	"fmt"
	"log"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
	"github.com/elioetibr/golang-yaml/v1/pkg/parser"
)

func main() {
	fmt.Println("=== Parser Behavior Investigation ===\n")

	// Test Case 1: YAML with leading comments
	yaml1 := `# This is a comment
# Another comment
name: test
value: 123`

	fmt.Println("Test 1: YAML with leading comments")
	fmt.Printf("Input:\n%s\n", yaml1)
	testParse(yaml1)

	// Test Case 2: YAML without leading comments
	yaml2 := `name: test
value: 123`

	fmt.Println("\nTest 2: YAML without leading comments")
	fmt.Printf("Input:\n%s\n", yaml2)
	testParse(yaml2)

	// Test Case 3: Just comments
	yaml3 := `# Only comments
# No actual YAML content`

	fmt.Println("\nTest 3: Only comments")
	fmt.Printf("Input:\n%s\n", yaml3)
	testParse(yaml3)

	// Test Case 4: Comments with proper YAML structure
	yaml4 := `config:
  # Comment inside structure
  name: test
  value: 123`

	fmt.Println("\nTest 4: Comments within YAML structure")
	fmt.Printf("Input:\n%s\n", yaml4)
	testParse(yaml4)
}

func testParse(yamlStr string) {
	l := lexer.NewLexerFromString(yamlStr)
	p := parser.NewParser(l)
	n, err := p.Parse()
	if err != nil {
		log.Printf("  ❌ Parse error: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Parsed successfully\n")
	fmt.Printf("  Node type: %T\n", n)

	switch node := n.(type) {
	case *node.ScalarNode:
		fmt.Printf("  ScalarNode value: %q\n", node.Value)
		if node.HeadComment != nil {
			fmt.Printf("  Has head comment: %v\n", node.HeadComment.Comments)
		}
	case *node.MappingNode:
		fmt.Printf("  MappingNode with %d pairs\n", len(node.Pairs))
		if node.HeadComment != nil {
			fmt.Printf("  Has head comment: %v\n", node.HeadComment.Comments)
		}
		for i, pair := range node.Pairs {
			if key, ok := pair.Key.(*node.ScalarNode); ok {
				fmt.Printf("    Pair %d: key=%q\n", i, key.Value)
			}
		}
	case *node.SequenceNode:
		fmt.Printf("  SequenceNode with %d items\n", len(node.Items))
	default:
		fmt.Printf("  Unknown node type\n")
	}
}