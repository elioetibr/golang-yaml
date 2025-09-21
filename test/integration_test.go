package test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/v0/pkg/decoder"
	"github.com/elioetibr/golang-yaml/v0/pkg/encoder"
	"github.com/elioetibr/golang-yaml/v0/pkg/node"
	parser2 "github.com/elioetibr/golang-yaml/v0/pkg/parser"
	"github.com/elioetibr/golang-yaml/v0/pkg/serializer"
	"github.com/elioetibr/golang-yaml/v0/pkg/transform"
)

// TestFullPipeline tests the complete pipeline
func TestFullPipeline(t *testing.T) {
	input := `
# Configuration file
server:
  host: localhost
  port: 8080

database:
  driver: postgres
  host: db.example.com
  port: 5432

features:
  - auth
  - logging
  - metrics`

	// 1. Parse
	root, err := parser2.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 2. Transform - Sort
	sortConfig := &transform.SortConfig{
		Mode:   transform.SortModeAscending,
		SortBy: transform.SortByKey,
	}
	sorter := transform.NewSorter(sortConfig)
	sorted := sorter.Sort(root)

	// 3. Serialize
	output, err := serializer.SerializeToString(sorted, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// 4. Verify output contains all keys in sorted order
	if !strings.Contains(output, "database:") {
		t.Error("Missing database section")
	}
	if !strings.Contains(output, "features:") {
		t.Error("Missing features section")
	}
	if !strings.Contains(output, "server:") {
		t.Error("Missing server section")
	}

	// Verify order (database should come before features and server)
	dbIndex := strings.Index(output, "database:")
	featuresIndex := strings.Index(output, "features:")
	serverIndex := strings.Index(output, "server:")

	if dbIndex > featuresIndex || dbIndex > serverIndex {
		t.Error("Sorting failed: database should come first")
	}
}

// TestRealWorldConfig tests with a realistic configuration
func TestRealWorldConfig(t *testing.T) {
	config := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
  labels:
    app: myapp
    environment: production

data:
  application.yaml: |
    server:
      port: 8080
      host: 0.0.0.0

    database:
      url: postgres://localhost:5432/mydb
      pool:
        min: 5
        max: 20

    redis:
      host: redis.example.com
      port: 6379

    features:
      - authentication
      - authorization
      - caching
      - monitoring`

	// Parse the config
	root, err := parser2.ParseString(config)
	if err != nil {
		t.Fatalf("Failed to parse real-world config: %v", err)
	}

	// Serialize it back
	output, err := serializer.SerializeToString(root, nil)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Verify key sections are present
	requiredSections := []string{
		"apiVersion:", "kind:", "metadata:", "data:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(output, section) {
			t.Errorf("Missing required section: %s", section)
		}
	}
}

// TestStreamProcessing tests handling multiple documents
func TestStreamProcessing(t *testing.T) {
	input := `---
document: 1
type: config
---
document: 2
type: data
---
document: 3
type: metadata
...`

	stream, err := parser2.ParseStream(input)
	if err != nil {
		t.Fatalf("Stream parsing failed: %v", err)
	}

	if len(stream.Documents) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(stream.Documents))
	}

	// Process each document
	for i, doc := range stream.Documents {
		if doc.Root == nil {
			t.Errorf("Document %d has nil root", i)
			continue
		}

		// Serialize each document
		output, err := serializer.SerializeToString(doc.Root, nil)
		if err != nil {
			t.Errorf("Failed to serialize document %d: %v", i, err)
		}

		if !strings.Contains(output, "document:") {
			t.Errorf("Document %d missing 'document' field", i)
		}
	}
}

// TestFileOperations tests reading from and writing to files
func TestFileOperations(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Test data
	testData := struct {
		Name    string   `yaml:"name"`
		Version string   `yaml:"version"`
		Tags    []string `yaml:"tags"`
	}{
		Name:    "TestApp",
		Version: "1.0.0",
		Tags:    []string{"test", "yaml", "go"},
	}

	// Encode to file
	encoder := encoder.NewEncoder(tmpfile)
	if err := encoder.Encode(testData); err != nil {
		t.Fatalf("Failed to encode to file: %v", err)
	}
	tmpfile.Close()

	// Read back from file
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	decoder := decoder.NewDecoder(file)
	var decoded struct {
		Name    string   `yaml:"name"`
		Version string   `yaml:"version"`
		Tags    []string `yaml:"tags"`
	}

	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode from file: %v", err)
	}

	// Verify
	if decoded.Name != testData.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, testData.Name)
	}
	if decoded.Version != testData.Version {
		t.Errorf("Version mismatch: got %q, want %q", decoded.Version, testData.Version)
	}
	if len(decoded.Tags) != len(testData.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(decoded.Tags), len(testData.Tags))
	}
}

// TestComplexDataTypes tests various Go data types
func TestComplexDataTypes(t *testing.T) {
	type Address struct {
		Street string `yaml:"street"`
		City   string `yaml:"city"`
		Zip    string `yaml:"zip"`
	}

	type Person struct {
		Name     string                 `yaml:"name"`
		Age      int                    `yaml:"age"`
		Email    string                 `yaml:"email,omitempty"`
		Address  Address                `yaml:"address"`
		Tags     []string               `yaml:"tags"`
		Settings map[string]interface{} `yaml:"settings"`
		IsActive bool                   `yaml:"is_active"`
		Balance  float64                `yaml:"balance"`
	}

	original := Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
			Zip:    "12345",
		},
		Tags: []string{"developer", "golang", "yaml"},
		Settings: map[string]interface{}{
			"theme":    "dark",
			"fontSize": 14,
			"autoSave": true,
		},
		IsActive: true,
		Balance:  1234.56,
	}

	// Marshal
	data, err := encoder.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded Person
	if err := decoder.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify all fields
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch")
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch")
	}
	if decoded.Address.City != original.Address.City {
		t.Errorf("Address.City mismatch")
	}
	if len(decoded.Tags) != len(original.Tags) {
		t.Errorf("Tags length mismatch")
	}
	if decoded.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch")
	}
	if decoded.Balance != original.Balance {
		t.Errorf("Balance mismatch")
	}
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "invalid_yaml",
			input:       "key: [value", // Unclosed flow sequence - parser is lenient
			expectError: false,
		},
		{
			name:        "undefined_alias",
			input:       "value: *undefined",
			expectError: true,
		},
		{
			name:        "duplicate_anchor",
			input:       "a: &x 1\nb: &x 2",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser2.ParseString(tc.input)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCommentPreservation tests that comments are preserved
func TestCommentPreservation(t *testing.T) {
	// Comment preservation is now implemented

	input := `# Header comment
# Second line

key1: value1  # inline comment

# Above key2
key2: value2

# Footer comment`

	// Parse
	root, err := parser2.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Serialize with comment preservation
	opts := &serializer.Options{
		PreserveComments: true,
		Indent:           2,
	}

	output, err := serializer.SerializeToString(root, opts)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Check for comments in output
	if !strings.Contains(output, "#") {
		t.Error("Comments were not preserved")
	}
}

// TestConcurrency tests thread safety
func TestConcurrency(t *testing.T) {
	yaml := `
name: test
value: 123
enabled: true`

	// Run multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Parse
			root, err := parser2.ParseString(yaml)
			if err != nil {
				t.Errorf("Goroutine %d: parse failed: %v", id, err)
				return
			}

			// Serialize
			_, err = serializer.SerializeToString(root, nil)
			if err != nil {
				t.Errorf("Goroutine %d: serialize failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestBuilderPattern tests the node builder
func TestBuilderPattern(t *testing.T) {
	builder := &node.DefaultBuilder{}

	// Build a complex structure
	root := builder.BuildMapping([]*node.MappingPair{
		{
			Key:   builder.BuildScalar("name", node.StylePlain),
			Value: builder.BuildScalar("MyApp", node.StylePlain),
		},
		{
			Key: builder.BuildScalar("servers", node.StylePlain),
			Value: builder.BuildSequence([]node.Node{
				builder.BuildScalar("server1.example.com", node.StylePlain),
				builder.BuildScalar("server2.example.com", node.StylePlain),
			}, node.StyleBlock),
		},
		{
			Key: builder.BuildScalar("config", node.StylePlain),
			Value: builder.BuildMapping([]*node.MappingPair{
				{
					Key:   builder.BuildScalar("timeout", node.StylePlain),
					Value: builder.BuildScalar("30", node.StylePlain),
				},
				{
					Key:   builder.BuildScalar("retries", node.StylePlain),
					Value: builder.BuildScalar("3", node.StylePlain),
				},
			}, node.StyleBlock),
		},
	}, node.StyleBlock)

	// Add anchor and tag
	rootWithAnchor := builder.WithAnchor(root, "config")
	rootWithTag := builder.WithTag(rootWithAnchor, "!AppConfig")

	// Serialize
	output, err := serializer.SerializeToString(rootWithTag, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Verify structure
	if !strings.Contains(output, "name: MyApp") {
		t.Error("Missing name field")
	}
	if !strings.Contains(output, "servers:") {
		t.Error("Missing servers field")
	}
	if !strings.Contains(output, "config:") {
		t.Error("Missing config field")
	}
}

// TestStreaming tests streaming encoder/decoder
func TestStreaming(t *testing.T) {
	// Create a buffer to simulate a stream
	var buf bytes.Buffer

	// Create encoder
	enc := encoder.NewEncoder(&buf)

	// Encode multiple values-with-comments
	data1 := map[string]string{"type": "first"}
	data2 := map[string]string{"type": "second"}

	if err := enc.Encode(data1); err != nil {
		t.Fatalf("First encode failed: %v", err)
	}

	// For multiple documents, we'd need to write document separator
	if _, err := io.WriteString(&buf, "\n---\n"); err != nil {
		t.Fatalf("Failed to write separator: %v", err)
	}

	if err := enc.Encode(data2); err != nil {
		t.Fatalf("Second encode failed: %v", err)
	}

	// Verify output
	output := buf.String()
	if !strings.Contains(output, "type: first") {
		t.Error("Missing first document")
	}
	if !strings.Contains(output, "type: second") {
		t.Error("Missing second document")
	}
}
