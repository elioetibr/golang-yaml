package main

import (
	"fmt"

	"github.com/elioetibr/golang-yaml/v0/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v0/pkg/parser"
	"github.com/elioetibr/golang-yaml/v0/pkg/serializer"
)

func main() {
	// Parse simple test case
	input := `strategy:
  # Comment about maxSurge
  maxSurge: 1
  # Comment about maxUnavailable
  maxUnavailable: 0`

	fmt.Println("=== Input ===")
	fmt.Println(input)
	fmt.Println()

	// Create lexer and parser
	l := lexer.NewLexerFromString(input)
	err := l.Initialize()
	if err != nil {
		panic(err)
	}

	p := parser.NewParser(l)
	root, err := p.Parse()
	if err != nil {
		panic(err)
	}

	// Serialize back
	output, err := serializer.SerializeToString(root, serializer.DefaultOptions())
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Output ===")
	fmt.Println(output)
}
