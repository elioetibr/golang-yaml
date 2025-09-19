package main

import (
	"fmt"
	"log"

	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
	"github.com/elioetibr/golang-yaml/pkg/transform"
)

func main() {
	roundTripYAML := `# Application Version
version: 1.0.0
name: TestApp # This must be in PascalCase
active: true # Enable set to true, disable set to false`

	sortConfig := &transform.SortConfig{
		Mode:   transform.SortModeAscending,
		SortBy: transform.SortByKey,
	}
	sorter := transform.NewSorter(sortConfig)

	// Parse
	rtRoot, err := parser.ParseString(roundTripYAML)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Modify (sort)
	rtSorted := sorter.Sort(rtRoot)

	// Serialize back
	rtResult, err := serializer.SerializeToString(rtSorted, nil)
	if err != nil {
		log.Fatalf("Serialize error: %v", err)
	}

	fmt.Println("Original:")
	fmt.Println(roundTripYAML)
	fmt.Println("\nAfter round-trip with sorting:")
	fmt.Println(rtResult)
}
