# Values Example

This example demonstrates YAML merging for configuration files, particularly useful for Helm-style values files.

## Files

- `values.yaml` - Base configuration file
- `values-merge.yaml` - Override values to merge
- `values-overridden.yaml` - Output file after merging
- `main.go` - Original implementation (manual merge logic)
- `main_simple.go.example` - Simplified implementation using the merge package

## Usage

### Using the original implementation:
```bash
go run main.go
```

### Using the merge package (recommended):

```go
// See main.go for the simplified implementation
package main

import (
	"github.com/elioetibr/golang-yaml/v0/pkg/merge"
)

func main() {
	// Simple one-liner merge with all defaults
	err := merge.MergeFilesToFile("values-with-comments.yaml", "values-with-comments-merge.yaml", "values-with-comments-overridden.yaml")

	// The merge package automatically:
	// - Preserves comments and documentation
	// - Maintains blank lines and formatting
	// - Performs deep merging of nested structures
	// - Handles arrays, maps, and scalars correctly
}
```

## Features Demonstrated

1. **Deep Merging**: Nested structures are merged recursively
2. **Comment Preservation**: Schema annotations and documentation are maintained
3. **Format Preservation**: Blank lines and structure are preserved
4. **Value Override**: Override values take precedence over base values

## Output

The merged file (`values-overridden.yaml`) will contain:
- All keys from the base file
- Values overridden from the merge file
- New keys added from the merge file
- All comments and formatting preserved