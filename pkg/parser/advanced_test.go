package parser

import (
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/node"
)

func TestAnchorAndAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, node.Node)
	}{
		{
			name: "simple_anchor_alias",
			input: `
default: &defaults
  host: localhost
  port: 8080

development:
  <<: *defaults
  debug: true

production:
  <<: *defaults
  host: prod.example.com`,
			validate: func(t *testing.T, root node.Node) {
				mapping, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatal("Expected MappingNode")
				}
				if len(mapping.Pairs) != 3 {
					t.Errorf("Expected 3 pairs, got %d", len(mapping.Pairs))
				}
			},
		},
		{
			name: "sequence_with_anchors",
			input: `
- &item1
  name: Item 1
  value: 100
- &item2
  name: Item 2
  value: 200
- *item1
- *item2`,
			validate: func(t *testing.T, root node.Node) {
				seq, ok := root.(*node.SequenceNode)
				if !ok {
					t.Fatal("Expected SequenceNode")
				}
				if len(seq.Items) != 4 {
					t.Errorf("Expected 4 items, got %d", len(seq.Items))
				}
			},
		},
		{
			name: "anchor_on_scalar",
			input: `
name: &myname "Application"
description: *myname`,
			validate: func(t *testing.T, root node.Node) {
				mapping, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatal("Expected MappingNode")
				}
				if len(mapping.Pairs) != 2 {
					t.Errorf("Expected 2 pairs, got %d", len(mapping.Pairs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			tt.validate(t, root)
		})
	}
}

func TestTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, node.Node)
	}{
		{
			name:  "explicit_string_tag",
			input: `value: !!str 123`,
			validate: func(t *testing.T, root node.Node) {
				mapping, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatal("Expected MappingNode")
				}
				if len(mapping.Pairs) == 0 {
					t.Fatal("No pairs in mapping")
				}
				value := mapping.Pairs[0].Value
				if scalar, ok := value.(*node.ScalarNode); ok {
					if scalar.Value != "123" {
						t.Errorf("Expected value '123', got %q", scalar.Value)
					}
					// Tag should be set
					if scalar.Tag() != "!!str" && scalar.Tag() != "tag:yaml.org,2002:str" {
						t.Errorf("Expected string tag, got %q", scalar.Tag())
					}
				}
			},
		},
		{
			name: "custom_tag",
			input: `person: !Person
  name: Alice
  age: 30`,
			validate: func(t *testing.T, root node.Node) {
				mapping, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatal("Expected MappingNode")
				}
				if len(mapping.Pairs) == 0 {
					t.Fatal("No pairs in mapping")
				}
				value := mapping.Pairs[0].Value
				if value.Tag() != "!Person" {
					t.Errorf("Expected !Person tag, got %q", value.Tag())
				}
			},
		},
		{
			name: "binary_tag",
			input: `image: !!binary |
  R0lGODdhDQAIAIAAAAAAANn
  Z2SwAAAAADQAIAAACF4SDGQ
  ar3xxbJ9p0qa7R0YxwzaFME
  1IAADs=`,
			validate: func(t *testing.T, root node.Node) {
				_, ok := root.(*node.MappingNode)
				if !ok {
					t.Fatal("Expected MappingNode")
				}
				// Just verify it parses
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if root != nil && tt.validate != nil {
				tt.validate(t, root)
			}
		})
	}
}

func TestMultipleDocuments(t *testing.T) {
	input := `---
document: 1
type: first
---
document: 2
type: second
---
document: 3
type: third
...`

	stream, err := ParseStream(input)
	if err != nil {
		t.Fatalf("ParseStream error: %v", err)
	}

	if len(stream.Documents) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(stream.Documents))
	}

	for i, doc := range stream.Documents {
		if doc.Root == nil {
			t.Errorf("Document %d has nil root", i)
			continue
		}

		mapping, ok := doc.Root.(*node.MappingNode)
		if !ok {
			t.Errorf("Document %d: expected MappingNode, got %T", i, doc.Root)
			continue
		}

		// Find the 'document' field
		for _, pair := range mapping.Pairs {
			if key, ok := pair.Key.(*node.ScalarNode); ok && key.Value == "document" {
				if value, ok := pair.Value.(*node.ScalarNode); ok {
					expectedValue := string(rune('1' + i))
					if value.Value != expectedValue {
						t.Errorf("Document %d: expected value %q, got %q", i, expectedValue, value.Value)
					}
				}
			}
		}
	}
}

func TestMergeKey(t *testing.T) {
	input := `
defaults: &defaults
  timeout: 30
  retries: 3
  debug: false

service1:
  <<: *defaults
  name: Service One
  port: 8080

service2:
  <<: *defaults
  name: Service Two
  port: 8081
  debug: true  # Override default`

	root, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	mapping, ok := root.(*node.MappingNode)
	if !ok {
		t.Fatal("Expected MappingNode")
	}

	// Find service1
	var service1 *node.MappingNode
	for _, pair := range mapping.Pairs {
		if key, ok := pair.Key.(*node.ScalarNode); ok && key.Value == "service1" {
			service1, _ = pair.Value.(*node.MappingNode)
			break
		}
	}

	if service1 == nil {
		t.Fatal("service1 not found")
	}

	// Check that merged fields are present
	hasTimeout := false
	hasRetries := false
	hasName := false
	for _, pair := range service1.Pairs {
		if key, ok := pair.Key.(*node.ScalarNode); ok {
			switch key.Value {
			case "timeout":
				hasTimeout = true
			case "retries":
				hasRetries = true
			case "name":
				hasName = true
			}
		}
	}

	if !hasTimeout {
		t.Error("Merged field 'timeout' not found in service1")
	}
	if !hasRetries {
		t.Error("Merged field 'retries' not found in service1")
	}
	if !hasName {
		t.Error("Field 'name' not found in service1")
	}
}

func TestDirectives(t *testing.T) {
	input := `%YAML 1.2
%TAG ! tag:example.com,2014:
---
!article
title: Test Article
author: !person
  name: John Doe
  email: john@example.com`

	stream, err := ParseStream(input)
	if err != nil {
		t.Fatalf("ParseStream error: %v", err)
	}

	if len(stream.Documents) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(stream.Documents))
	}

	doc := stream.Documents[0]

	// Check directives
	if len(doc.Directives) != 2 {
		t.Errorf("Expected 2 directives, got %d", len(doc.Directives))
	}

	if doc.Directives[0].Name != "YAML" {
		t.Errorf("Expected YAML directive, got %s", doc.Directives[0].Name)
	}

	if doc.Directives[1].Name != "TAG" {
		t.Errorf("Expected TAG directive, got %s", doc.Directives[1].Name)
	}

	// Check that root has the custom tag
	if doc.Root != nil && doc.Root.Tag() != "!article" {
		t.Errorf("Expected !article tag on root, got %q", doc.Root.Tag())
	}
}

func TestTagResolution(t *testing.T) {
	resolver := NewTagResolver()

	tests := []struct {
		tag      string
		expected string
	}{
		{"!!str", "tag:yaml.org,2002:str"},
		{"!!int", "tag:yaml.org,2002:int"},
		{"!!float", "tag:yaml.org,2002:float"},
		{"!!bool", "tag:yaml.org,2002:bool"},
		{"!custom", "!custom"},
		{"", ""},
	}

	for _, tt := range tests {
		resolved := resolver.ResolveTag(tt.tag)
		if resolved != tt.expected {
			t.Errorf("ResolveTag(%q): expected %q, got %q", tt.tag, tt.expected, resolved)
		}
	}
}

func TestTagInference(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"hello", CommonTags.Str},
		{"123", CommonTags.Int},
		{"3.14", CommonTags.Float},
		{"true", CommonTags.Bool},
		{"false", CommonTags.Bool},
		{"null", CommonTags.Null},
		{"~", CommonTags.Null},
		{"", CommonTags.Null},
		{"yes", CommonTags.Bool},
		{"no", CommonTags.Bool},
	}

	for _, tt := range tests {
		inferred := InferTag(tt.value)
		if inferred != tt.expected {
			t.Errorf("InferTag(%q): expected %q, got %q", tt.value, tt.expected, inferred)
		}
	}
}

func TestComplexAnchorScenario(t *testing.T) {
	input := `
base: &base
  name: Base Config
  settings: &settings
    timeout: 30
    retries: 3

extended:
  <<: *base
  settings:
    <<: *settings
    timeout: 60  # Override
  extra: value`

	root, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Just verify it parses without error
	if root == nil {
		t.Error("Expected non-nil root")
	}
}

func TestAnchorErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name: "undefined_alias",
			input: `
value: *undefined`,
			shouldError: true,
		},
		{
			name: "duplicate_anchor",
			input: `
first: &anchor value1
second: &anchor value2`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseString(tt.input)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDocumentMarkers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *Stream)
	}{
		{
			name: "explicit_markers",
			input: `---
content: here
...`,
			check: func(t *testing.T, stream *Stream) {
				if len(stream.Documents) != 1 {
					t.Errorf("Expected 1 document, got %d", len(stream.Documents))
					return
				}
				doc := stream.Documents[0]
				if !doc.ExplicitStart {
					t.Error("Expected explicit start marker")
				}
				if !doc.ExplicitEnd {
					t.Error("Expected explicit end marker")
				}
			},
		},
		{
			name:  "implicit_document",
			input: `key: value`,
			check: func(t *testing.T, stream *Stream) {
				if len(stream.Documents) != 1 {
					t.Errorf("Expected 1 document, got %d", len(stream.Documents))
					return
				}
				doc := stream.Documents[0]
				if doc.ExplicitStart {
					t.Error("Expected no explicit start marker")
				}
				if doc.ExplicitEnd {
					t.Error("Expected no explicit end marker")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := ParseStream(tt.input)
			if err != nil {
				t.Fatalf("ParseStream error: %v", err)
			}
			tt.check(t, stream)
		})
	}
}

func TestEmptyDocument(t *testing.T) {
	input := `---
---
content: here
---
...`

	stream, err := ParseStream(input)
	if err != nil {
		t.Fatalf("ParseStream error: %v", err)
	}

	// Should have 3 documents (first is empty)
	if len(stream.Documents) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(stream.Documents))
	}

	// First document should have nil root
	if stream.Documents[0].Root != nil {
		t.Error("Expected first document to be empty")
	}

	// Second document should have content
	if stream.Documents[1].Root == nil {
		t.Error("Expected second document to have content")
	}

	// Third document should be empty
	if stream.Documents[2].Root != nil {
		t.Error("Expected third document to be empty")
	}
}
