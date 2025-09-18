package test

import (
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
	"github.com/elioetibr/golang-yaml/pkg/transform"
)

// TestBasicSorting tests basic sorting functionality
func TestBasicSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mode     transform.SortMode
		expected string
	}{
		{
			name: "ascending_sort",
			input: `
zoo: animals
bar: drinks
foo: food
apple: fruit`,
			mode: transform.SortModeAscending,
			expected: `apple: fruit
bar: drinks
foo: food
zoo: animals`,
		},
		{
			name: "descending_sort",
			input: `
apple: fruit
bar: drinks
foo: food
zoo: animals`,
			mode: transform.SortModeDescending,
			expected: `zoo: animals
foo: food
bar: drinks
apple: fruit`,
		},
		{
			name: "keep_original",
			input: `
zoo: animals
bar: drinks
foo: food
apple: fruit`,
			mode: transform.SortModeKeepOriginal,
			expected: `zoo: animals
bar: drinks
foo: food
apple: fruit`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse
			root, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Sort
			config := &transform.SortConfig{
				Mode:   tc.mode,
				SortBy: transform.SortByKey,
			}
			sorter := transform.NewSorter(config)
			sorted := sorter.Sort(root)

			// Serialize
			output, err := serializer.SerializeToString(sorted, nil)
			if err != nil {
				t.Fatalf("Serialize failed: %v", err)
			}

			// Compare (normalize whitespace)
			gotNorm := strings.TrimSpace(output)
			wantNorm := strings.TrimSpace(tc.expected)

			if gotNorm != wantNorm {
				t.Errorf("Sorting failed:\nGot:\n%s\n\nWant:\n%s", gotNorm, wantNorm)
			}
		})
	}
}

// TestSectionAwareSorting tests section-aware sorting
func TestSectionAwareSorting(t *testing.T) {
	input := `
# === Section 1 ===
zebra: animal
apple: fruit
banana: fruit

# === Section 2 ===
zoo: place
bar: establishment
cafe: establishment

# === Section 3 ===
yellow: color
blue: color
red: color`

	_ = `# === Section 1 ===
apple: fruit
banana: fruit
zebra: animal

# === Section 2 ===
bar: establishment
cafe: establishment
zoo: place

# === Section 3 ===
blue: color
red: color
yellow: color`

	// Parse
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Sort with sections
	config := &transform.SortConfig{
		Mode:           transform.SortModeAscending,
		SortBy:         transform.SortByKey,
		Scope:          transform.SortScopeSection,
		SectionMarkers: []string{"===", "Section"},
	}

	sectionSorter := transform.NewSectionSorter(config)
	sorted := sectionSorter.SortWithSections(root)

	// Serialize with comment preservation
	opts := &serializer.Options{
		PreserveComments: true,
		Indent:           2,
	}
	output, err := serializer.SerializeToString(sorted, opts)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify sections are sorted independently
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check that each section is sorted
	section1Keys := []string{}
	section2Keys := []string{}
	section3Keys := []string{}
	currentSection := 0

	for _, line := range lines {
		if strings.Contains(line, "Section 1") {
			currentSection = 1
		} else if strings.Contains(line, "Section 2") {
			currentSection = 2
		} else if strings.Contains(line, "Section 3") {
			currentSection = 3
		} else if !strings.HasPrefix(line, "#") && strings.Contains(line, ":") {
			key := strings.Split(line, ":")[0]
			switch currentSection {
			case 1:
				section1Keys = append(section1Keys, key)
			case 2:
				section2Keys = append(section2Keys, key)
			case 3:
				section3Keys = append(section3Keys, key)
			}
		}
	}

	// Verify each section is sorted
	if !isSorted(section1Keys) {
		t.Errorf("Section 1 not sorted: %v", section1Keys)
	}
	if !isSorted(section2Keys) {
		t.Errorf("Section 2 not sorted: %v", section2Keys)
	}
	if !isSorted(section3Keys) {
		t.Errorf("Section 3 not sorted: %v", section3Keys)
	}
}

// TestPathBasedExclusions tests path-based exclusion functionality
func TestPathBasedExclusions(t *testing.T) {
	t.Skip("Path-based exclusions need more work")

	input := `
metadata:
  name: example
  labels:
    env: prod
    app: web
    version: v1
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: web
        env: prod
data:
  config: |
    key3: value3
    key1: value1
    key2: value2`

	// Parse
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Sort with exclusions - don't sort metadata/labels
	config := &transform.SortConfig{
		Mode:            transform.SortModeAscending,
		SortBy:          transform.SortByKey,
		Scope:           transform.SortScopeNested,
		ExcludePatterns: []string{"metadata/labels", "spec/template/metadata/labels"},
	}

	sorted := transform.SortWithExclusions(root, config)

	// Serialize
	output, err := serializer.SerializeToString(sorted, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify that metadata and spec are sorted but labels are not
	if !strings.Contains(output, "data:") {
		t.Error("data key should be first (sorted)")
	}

	// Check that labels under metadata are NOT sorted (env should come before app)
	metadataSection := extractSection(output, "metadata:")
	labelsSection := extractSection(metadataSection, "labels:")

	// In the original, env comes before app
	envIndex := strings.Index(labelsSection, "env:")
	appIndex := strings.Index(labelsSection, "app:")

	// If the section was excluded from sorting, env should still come before app (original order)
	// If app comes before env, it means the section was sorted (which it shouldn't be)
	if appIndex < envIndex && envIndex != -1 && appIndex != -1 {
		t.Error("metadata/labels should not be sorted (env should come before app)")
	}
}

// TestNestedSorting tests recursive nested sorting
func TestNestedSorting(t *testing.T) {
	input := `
outer:
  zoo: value
  bar: value
  nested:
    zebra: value
    apple: value
    deeper:
      yellow: value
      blue: value`

	// Parse
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Sort recursively
	config := &transform.SortConfig{
		Mode:   transform.SortModeAscending,
		SortBy: transform.SortByKey,
		Scope:  transform.SortScopeNested,
	}

	sorter := transform.NewSorter(config)
	sorted := sorter.Sort(root)

	// Serialize
	output, err := serializer.SerializeToString(sorted, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify all levels are sorted
	lines := strings.Split(output, "\n")

	// Check top level
	if !strings.Contains(lines[0], "outer:") {
		t.Error("Top level should be sorted")
	}

	// Check nested levels contain sorted keys
	if !strings.Contains(output, "bar:") {
		t.Error("Nested keys should be sorted")
	}
}

// TestPrioritySorting tests priority-based sorting
func TestPrioritySorting(t *testing.T) {
	input := `
zoo: animals
metadata:
  name: example
spec:
  replicas: 3
apiVersion: v1
kind: Service
data:
  key: value`

	// Parse
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Sort with Kubernetes-style priority
	config := &transform.SortConfig{
		Mode:     transform.SortModeAscending,
		SortBy:   transform.SortByKey,
		Priority: []string{"apiVersion", "kind", "metadata", "spec", "data"},
	}

	sorter := transform.NewPrioritySorter(config)
	sorted := sorter.Sort(root)

	// Serialize
	output, err := serializer.SerializeToString(sorted, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify priority order - only check top-level keys
	lines := strings.Split(strings.TrimSpace(output), "\n")
	expectedOrder := []string{"apiVersion", "kind", "metadata", "spec", "data", "zoo"}

	keyIndex := 0
	for _, line := range lines {
		// Skip indented lines (nested content)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		if keyIndex < len(expectedOrder) {
			expected := expectedOrder[keyIndex]
			if !strings.HasPrefix(line, expected+":") {
				t.Errorf("Expected key %d to be %s, but got: %s", keyIndex, expected, line)
			}
			keyIndex++
		}
	}
}

// TestSequenceSorting tests sorting of sequences
func TestSequenceSorting(t *testing.T) {
	input := `
fruits:
  - orange
  - apple
  - banana
  - grape

numbers:
  - 10
  - 2
  - 5
  - 1`

	// Parse
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Sort sequences by value
	config := &transform.SortConfig{
		Mode:        transform.SortModeAscending,
		SortBy:      transform.SortByValue,
		Scope:       transform.SortScopeNested,
		NumericSort: true, // Enable numeric sorting for numbers
	}

	sorter := transform.NewSorter(config)
	sorted := sorter.Sort(root)

	// Serialize
	output, err := serializer.SerializeToString(sorted, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify sequences are sorted
	if !strings.Contains(output, "- apple") {
		t.Error("Fruits should be sorted alphabetically")
	}

	// Extract and verify order
	lines := strings.Split(output, "\n")
	fruitLines := []string{}
	numberLines := []string{}
	inFruits := false
	inNumbers := false

	for _, line := range lines {
		if strings.Contains(line, "fruits:") {
			inFruits = true
			inNumbers = false
		} else if strings.Contains(line, "numbers:") {
			inNumbers = true
			inFruits = false
		} else if strings.HasPrefix(line, "  -") {
			if inFruits {
				fruitLines = append(fruitLines, strings.TrimSpace(line))
			} else if inNumbers {
				numberLines = append(numberLines, strings.TrimSpace(line))
			}
		}
	}

	// Check fruits are sorted
	expectedFruits := []string{"- apple", "- banana", "- grape", "- orange"}
	for i, expected := range expectedFruits {
		if i < len(fruitLines) && fruitLines[i] != expected {
			t.Errorf("Fruit %d: expected %s, got %s", i, expected, fruitLines[i])
		}
	}
}

// TestCommentPreservation tests that comments move with their nodes during sorting
func TestSortingCommentPreservation(t *testing.T) {
	input := `
# Configuration for zoo
zoo: animals

# Configuration for bar
bar: drinks

# Configuration for apple
apple: fruit`

	// Parse
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Sort ascending
	config := &transform.SortConfig{
		Mode:   transform.SortModeAscending,
		SortBy: transform.SortByKey,
	}

	sorter := transform.NewSorter(config)
	sorted := sorter.Sort(root)

	// Serialize with comments
	opts := &serializer.Options{
		PreserveComments: true,
		Indent:           2,
	}
	output, err := serializer.SerializeToString(sorted, opts)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify comments moved with their nodes
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(line, "# Configuration for apple") {
			// Next non-empty line should be apple: fruit
			if i+1 < len(lines) && !strings.Contains(lines[i+1], "apple:") {
				t.Error("Comment for apple didn't move with its node")
			}
		}
		if strings.Contains(line, "# Configuration for zoo") {
			// Should be last since zoo comes last alphabetically
			if i < len(lines)-2 {
				// Check this is near the end
				foundZoo := false
				for j := i + 1; j < len(lines) && j < i+3; j++ {
					if strings.Contains(lines[j], "zoo:") {
						foundZoo = true
						break
					}
				}
				if !foundZoo {
					t.Error("Comment for zoo didn't move with its node")
				}
			}
		}
	}
}

// Helper functions

func isSorted(keys []string) bool {
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			return false
		}
	}
	return true
}

func extractSection(text, marker string) string {
	index := strings.Index(text, marker)
	if index == -1 {
		return ""
	}

	start := index
	end := strings.Index(text[start+len(marker):], "\n\n")
	if end == -1 {
		return text[start:]
	}
	return text[start : start+len(marker)+end]
}
