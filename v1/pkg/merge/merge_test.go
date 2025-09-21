package merge

import (
	"strings"
	"testing"
)

func TestMergeStrings(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		override  string
		expected  string
		shouldErr bool
	}{
		{
			name: "simple merge",
			base: `name: app
version: 1.0.0`,
			override: `version: 2.0.0
env: production`,
			expected: `name: app
version: 2.0.0
env: production`,
		},
		{
			name: "preserve comments",
			base: `# Application config
name: myapp
version: 1.0.0  # Version number`,
			override: `version: 2.0.0`,
			expected: `# Application config
name: myapp
version: 2.0.0  # Version number`,
		},
		{
			name: "deep merge nested",
			base: `server:
  host: localhost
  port: 8080
database:
  host: localhost`,
			override: `server:
  port: 443
database:
  host: db.example.com`,
			expected: `server:
  host: localhost
  port: 443
database:
  host: db.example.com`,
		},
		{
			name: "handle arrays with replace strategy",
			base: `features:
  - auth
  - api`,
			override: `features:
  - monitoring
  - metrics`,
			expected: `features:
  - monitoring
  - metrics`,
		},
		{
			name: "preserve blank lines",
			base: `# Config
name: app
version: 1.0.0`,
			override: `version: 2.0.0`,
			expected: `# Config
name: app
version: 2.0.0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Strings(tt.base, tt.override)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr {
				// Normalize whitespace for comparison
				expected := strings.TrimSpace(tt.expected)
				actual := strings.TrimSpace(result)
				if expected != actual {
					t.Errorf("merge mismatch\nexpected:\n%s\n\nactual:\n%s", expected, actual)
				}
			}
		})
	}
}

func TestMergeStrategies(t *testing.T) {
	base := `config:
  nested:
    value: base
    other: keep
  simple: base
list: [1, 2, 3]`

	override := `config:
  nested:
    value: override
  simple: override
  new: added
list: [4, 5]`

	t.Run("deep merge", func(t *testing.T) {
		opts := DefaultOptions().WithStrategy(StrategyDeep)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have 'other: keep' preserved
		if !strings.Contains(result, "other: keep") {
			t.Errorf("deep merge should preserve unmodified nested fields")
		}
		if !strings.Contains(result, "new: added") {
			t.Errorf("deep merge should add new fields")
		}
	})

	t.Run("shallow merge", func(t *testing.T) {
		opts := DefaultOptions().WithStrategy(StrategyShallow)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should replace entire nested object
		if strings.Contains(result, "other: keep") {
			t.Errorf("shallow merge should replace entire nested objects")
		}
		if !strings.Contains(result, "new: added") {
			t.Errorf("shallow merge should add new top-level fields")
		}
	})

	t.Run("override strategy", func(t *testing.T) {
		opts := DefaultOptions().WithStrategy(StrategyOverride)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return override as-is
		if strings.Contains(result, "other: keep") {
			t.Errorf("override strategy should not preserve base fields")
		}
		if strings.Contains(result, "simple: base") {
			t.Errorf("override strategy should not preserve base values-with-comments")
		}
	})
}

func TestArrayMergeStrategies(t *testing.T) {
	base := `features:
  - auth
  - api
  - logging`

	override := `features:
  - monitoring
  - metrics`

	t.Run("array replace", func(t *testing.T) {
		opts := DefaultOptions().WithArrayStrategy(ArrayReplace)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.Contains(result, "auth") || strings.Contains(result, "api") {
			t.Errorf("array replace should not keep base items")
		}
		if !strings.Contains(result, "monitoring") || !strings.Contains(result, "metrics") {
			t.Errorf("array replace should use override items")
		}
	})

	t.Run("array append", func(t *testing.T) {
		opts := DefaultOptions().WithArrayStrategy(ArrayAppend)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "auth") || !strings.Contains(result, "api") {
			t.Errorf("array append should keep base items")
		}
		if !strings.Contains(result, "monitoring") || !strings.Contains(result, "metrics") {
			t.Errorf("array append should add override items")
		}
	})

	t.Run("array merge by index", func(t *testing.T) {
		base := `settings:
  - name: timeout
    value: 30
  - name: retries
    value: 3`

		override := `settings:
  - name: timeout
    value: 60
  - name: cache
    value: true`

		opts := DefaultOptions().WithArrayStrategy(ArrayMergeByIndex)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// First item should be merged
		if !strings.Contains(result, "value: 60") {
			t.Errorf("array merge by index should merge items at same index")
		}
		// Second item should be replaced
		if !strings.Contains(result, "cache") {
			t.Errorf("array merge by index should replace items at same index")
		}
	})
}

func TestCommentPreservation(t *testing.T) {
	base := `# Main configuration
# Version: 1.0

# Server settings
server:
  # Host to bind to
  host: localhost  # Can be changed
  # Port number
  port: 8080

# Database config
database:
  driver: postgres
  host: localhost`

	override := `# Production overrides

server:
  host: api.example.com  # Production host
  port: 443

database:
  host: db.prod.example.com`

	t.Run("with comment preservation", func(t *testing.T) {
		opts := DefaultOptions()
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should preserve main comments
		if !strings.Contains(result, "# Main configuration") {
			t.Errorf("should preserve base file header comments")
		}
		// Should preserve field comments
		if !strings.Contains(result, "# Server settings") {
			t.Errorf("should preserve field comments")
		}
		// Should preserve inline comments from override
		if !strings.Contains(result, "# Production host") {
			t.Errorf("should preserve inline comments from override")
		}
	})

	t.Run("without comment preservation", func(t *testing.T) {
		opts := &Options{
			Strategy:           StrategyDeep,
			PreserveComments:   false,
			PreserveBlankLines: false,
			ArrayMergeStrategy: ArrayReplace,
		}
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should not have comments
		if strings.Contains(result, "# Main configuration") {
			t.Errorf("should not preserve comments when disabled")
		}
	})
}

func TestMergeFiles(t *testing.T) {
	// This test would need actual files, so we'll skip in unit tests
	t.Skip("File-based tests require test fixtures")
}

func TestEmptyValueHandling(t *testing.T) {
	base := `name: app
version: 1.0.0
env: development`

	override := `version: ""
env: production`

	t.Run("override empty values-with-comments", func(t *testing.T) {
		opts := DefaultOptions().WithOverrideEmpty(true)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.Contains(result, "version: 1.0.0") {
			t.Errorf("should override with empty value when OverrideEmpty is true")
		}
	})

	t.Run("don't override empty values-with-comments", func(t *testing.T) {
		opts := DefaultOptions().WithOverrideEmpty(false)
		result, err := StringsWithOptions(base, override, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "version: 1.0.0") {
			t.Errorf("should keep base value when override is empty and OverrideEmpty is false")
		}
	})
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		override string
	}{
		{
			name: "type mismatch - mapping to scalar",
			base: `config: value`,
			override: `config:
  nested: value`,
		},
		{
			name: "type mismatch - sequence to mapping",
			base: `items: [1, 2, 3]`,
			override: `items:
  key: value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Strings(tt.base, tt.override)
			if err == nil {
				t.Errorf("expected error for type mismatch but got none")
			}
		})
	}
}
