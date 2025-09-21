package decoder_test

import (
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/decoder"
)

func TestDecoderNoPanic(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		dest interface{}
	}{
		{
			name: "simple map with interface values",
			yaml: `
key1: value1
key2: 123
key3: true
nested:
  inner: value
  number: 456
`,
			dest: &map[string]interface{}{},
		},
		{
			name: "complex nested structure",
			yaml: `
parent:
  child1:
    grandchild1: value1
    grandchild2: 789
  child2:
    - item1
    - item2
    - item3
  child3: simple
array:
  - name: first
    value: 1
  - name: second
    value: 2
`,
			dest: &map[string]interface{}{},
		},
		{
			name: "map with nil values",
			yaml: `
key1:
key2: null
key3: value3
`,
			dest: &map[string]interface{}{},
		},
		{
			name: "deeply nested interfaces",
			yaml: `
level1:
  level2:
    level3:
      level4:
        level5:
          key: deepvalue
          number: 12345
          bool: false
`,
			dest: &map[string]interface{}{},
		},
		{
			name: "mixed types in map",
			yaml: `
string: hello
int: 42
float: 3.14
bool: true
null: null
empty:
array: [1, 2, 3]
object:
  nested: value
`,
			dest: &map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decoder.Unmarshal([]byte(tt.yaml), tt.dest)
			if err != nil {
				t.Errorf("Unmarshal failed: %v", err)
			}
		})
	}
}

// TestEdgeCases tests various edge cases that might cause panics
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		dest    interface{}
		wantErr bool
	}{
		{
			name: "empty yaml",
			yaml: "",
			dest: &map[string]interface{}{},
		},
		{
			name: "only comments",
			yaml: `# just a comment`,
			dest: &map[string]interface{}{},
		},
		{
			name: "invalid destination",
			yaml: `key: value`,
			dest: nil,
			wantErr: true,
		},
		{
			name: "array to map",
			yaml: `- item1
- item2`,
			dest: &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decoder.Unmarshal([]byte(tt.yaml), tt.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}