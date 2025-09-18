package test

import (
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
	"github.com/elioetibr/golang-yaml/pkg/transform"
)

// TestCommentSorting tests that comments are preserved correctly during sorting
func TestCommentSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "comments_with_sorting",
			input: `# Application Version
version: 1.0.0
name: TestApp # This must be in PascalCase
active: true # Enable set to true, disable set to false`,
			expected: `active: true  # Enable set to true, disable set to false
name: TestApp  # This must be in PascalCase
# Application Version
version: 1.0.0`,
		},
		{
			name: "multiple_head_comments",
			input: `# First comment
# Second comment
zebra: animal
apple: fruit`,
			expected: `apple: fruit
# First comment
# Second comment
zebra: animal`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sortConfig := &transform.SortConfig{
				Mode:   transform.SortModeAscending,
				SortBy: transform.SortByKey,
			}
			sorter := transform.NewSorter(sortConfig)

			// Parse
			root, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Sort
			sorted := sorter.Sort(root)

			// Serialize back
			result, err := serializer.SerializeToString(sorted, nil)
			if err != nil {
				t.Fatalf("Serialize error: %v", err)
			}

			// Trim trailing newline for comparison
			if len(result) > 0 && result[len(result)-1] == '\n' {
				result = result[:len(result)-1]
			}

			if result != tc.expected {
				t.Errorf("Mismatch:\nExpected:\n%s\n\nGot:\n%s", tc.expected, result)
			}
		})
	}
}
