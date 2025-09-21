package main

import (
	"fmt"

	"github.com/elioetibr/golang-yaml/v0/pkg/merge"
	"github.com/elioetibr/golang-yaml/v0/pkg/parser"
	"github.com/elioetibr/golang-yaml/v0/pkg/serializer"
)

func main() {
	// Test simple merge case
	base := `imagePullSecrets:
  - name: ecr-secret`

	override := `imagePullSecrets: []`

	fmt.Println("=== Base ===")
	fmt.Println(base)
	fmt.Println("\n=== Override ===")
	fmt.Println(override)

	// Parse both
	baseNode, err := parser.ParseString(base)
	if err != nil {
		panic(err)
	}

	overrideNode, err := parser.ParseString(override)
	if err != nil {
		panic(err)
	}

	// Merge
	merged, err := merge.Merge(baseNode, overrideNode)
	if err != nil {
		panic(err)
	}

	// Serialize
	result, err := serializer.SerializeToString(merged, &serializer.Options{
		Indent:             2,
		PreserveComments:   true,
		PreserveBlankLines: false, // Don't preserve blank lines
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("\n=== Merged ===")
	fmt.Println(result)
}
