package encoder

import (
	"bytes"
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v0/pkg/decoder"
)

func TestMarshalScalar(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
		{"nil", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			result := strings.TrimSpace(string(data))
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMarshalSlice(t *testing.T) {
	input := []string{"a", "b", "c"}
	expected := "- a\n- b\n- c"

	data, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	result := strings.TrimSpace(string(data))
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestMarshalMap(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	data, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Maps are unordered, so just check that all key-value pairs are present
	result := string(data)
	if !strings.Contains(result, "a: 1") {
		t.Error("Missing 'a: 1' in output")
	}
	if !strings.Contains(result, "b: 2") {
		t.Error("Missing 'b: 2' in output")
	}
	if !strings.Contains(result, "c: 3") {
		t.Error("Missing 'c: 3' in output")
	}
}

func TestMarshalStruct(t *testing.T) {
	type Person struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
		City string `yaml:"city,omitempty"`
	}

	tests := []struct {
		name     string
		input    Person
		contains []string
		excludes []string
	}{
		{
			name:     "all_fields",
			input:    Person{Name: "Alice", Age: 30, City: "NYC"},
			contains: []string{"name: Alice", "age: 30", "city: NYC"},
			excludes: []string{},
		},
		{
			name:     "omitempty",
			input:    Person{Name: "Bob", Age: 25},
			contains: []string{"name: Bob", "age: 25"},
			excludes: []string{"city:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			result := string(data)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected to contain %q, but didn't. Got:\n%s", expected, result)
				}
			}

			for _, excluded := range tt.excludes {
				if strings.Contains(result, excluded) {
					t.Errorf("Expected not to contain %q, but did. Got:\n%s", excluded, result)
				}
			}
		})
	}
}

func TestMarshalNested(t *testing.T) {
	type Address struct {
		Street string `yaml:"street"`
		City   string `yaml:"city"`
	}

	type Person struct {
		Name    string  `yaml:"name"`
		Address Address `yaml:"address"`
	}

	input := Person{
		Name: "Alice",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	data, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	result := string(data)
	if !strings.Contains(result, "name: Alice") {
		t.Error("Missing name field")
	}
	if !strings.Contains(result, "address:") {
		t.Error("Missing address field")
	}
	if !strings.Contains(result, "street: 123 Main St") {
		t.Error("Missing street field")
	}
	if !strings.Contains(result, "city: NYC") {
		t.Error("Missing city field")
	}
}

func TestEncoder(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "key1: value1") {
		t.Error("Missing key1 in encoded output")
	}
	if !strings.Contains(result, "key2: value2") {
		t.Error("Missing key2 in encoded output")
	}
}

func TestRoundTrip(t *testing.T) {
	type TestStruct struct {
		Name   string   `yaml:"name"`
		Age    int      `yaml:"age"`
		Tags   []string `yaml:"tags"`
		Active bool     `yaml:"active"`
	}

	original := TestStruct{
		Name:   "Test",
		Age:    25,
		Tags:   []string{"go", "yaml"},
		Active: true,
	}

	// Marshal to YAML
	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal back
	var decoded TestStruct
	err = decoder.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Compare
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch: got %d, want %d", decoded.Age, original.Age)
	}
	if len(decoded.Tags) != len(original.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(decoded.Tags), len(original.Tags))
	}
	if decoded.Active != original.Active {
		t.Errorf("Active mismatch: got %v, want %v", decoded.Active, original.Active)
	}
}
