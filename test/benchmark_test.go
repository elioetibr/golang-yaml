package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/decoder"
	"github.com/elioetibr/golang-yaml/pkg/encoder"
	"github.com/elioetibr/golang-yaml/pkg/lexer"
	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
)

// Sample YAML documents for benchmarking
var (
	smallYAML = `
name: test
value: 123
enabled: true`

	mediumYAML = `
application:
  name: MyApp
  version: 1.0.0
  description: A test application
server:
  host: localhost
  port: 8080
  ssl: true
  timeout: 30
database:
  driver: postgres
  host: localhost
  port: 5432
  name: mydb
features:
  - logging
  - monitoring
  - caching
  - security`

	largeYAML = generateLargeYAML()
)

// generateLargeYAML creates a large YAML document for benchmarking
func generateLargeYAML() string {
	var sb strings.Builder
	sb.WriteString("root:\n")

	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("  service%d:\n", i))
		sb.WriteString(fmt.Sprintf("    name: Service-%d\n", i))
		sb.WriteString(fmt.Sprintf("    port: %d\n", 8000+i))
		sb.WriteString("    enabled: true\n")
		sb.WriteString("    config:\n")
		sb.WriteString("      timeout: 30\n")
		sb.WriteString("      retries: 3\n")
		sb.WriteString("      cache: true\n")
		sb.WriteString("    endpoints:\n")
		for j := 0; j < 10; j++ {
			sb.WriteString(fmt.Sprintf("      - /api/v1/endpoint%d\n", j))
		}
	}

	return sb.String()
}

// Lexer benchmarks

func BenchmarkLexerSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := lexer.NewLexerFromString(smallYAML)
		l.Initialize()

		for {
			token, err := l.NextToken()
			if err != nil || token.Type == lexer.TokenEOF {
				break
			}
		}
	}
}

func BenchmarkLexerMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := lexer.NewLexerFromString(mediumYAML)
		l.Initialize()

		for {
			token, err := l.NextToken()
			if err != nil || token.Type == lexer.TokenEOF {
				break
			}
		}
	}
}

func BenchmarkLexerLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := lexer.NewLexerFromString(largeYAML)
		l.Initialize()

		for {
			token, err := l.NextToken()
			if err != nil || token.Type == lexer.TokenEOF {
				break
			}
		}
	}
}

// Parser benchmarks

func BenchmarkParserSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(smallYAML)
	}
}

func BenchmarkParserMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(mediumYAML)
	}
}

func BenchmarkParserLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(largeYAML)
	}
}

// Serializer benchmarks

func BenchmarkSerializerSmall(b *testing.B) {
	// Parse once
	root, _ := parser.ParseString(smallYAML)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = serializer.SerializeToString(root, nil)
	}
}

func BenchmarkSerializerMedium(b *testing.B) {
	root, _ := parser.ParseString(mediumYAML)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = serializer.SerializeToString(root, nil)
	}
}

func BenchmarkSerializerLarge(b *testing.B) {
	root, _ := parser.ParseString(largeYAML)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = serializer.SerializeToString(root, nil)
	}
}

// Marshal/Unmarshal benchmarks

type BenchmarkStruct struct {
	Name       string            `yaml:"name"`
	Value      int               `yaml:"value"`
	Enabled    bool              `yaml:"enabled"`
	Tags       []string          `yaml:"tags"`
	Properties map[string]string `yaml:"properties"`
}

func BenchmarkMarshal(b *testing.B) {
	data := BenchmarkStruct{
		Name:    "test",
		Value:   123,
		Enabled: true,
		Tags:    []string{"tag1", "tag2", "tag3"},
		Properties: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.Marshal(data)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	yamlData := []byte(`
name: test
value: 123
enabled: true
tags:
  - tag1
  - tag2
  - tag3
properties:
  key1: value1
  key2: value2
  key3: value3`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data BenchmarkStruct
		_ = decoder.Unmarshal(yamlData, &data)
	}
}

// Round-trip benchmarks

func BenchmarkRoundTripSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root, _ := parser.ParseString(smallYAML)
		output, _ := serializer.SerializeToString(root, nil)
		_, _ = parser.ParseString(output)
	}
}

func BenchmarkRoundTripMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root, _ := parser.ParseString(mediumYAML)
		output, _ := serializer.SerializeToString(root, nil)
		_, _ = parser.ParseString(output)
	}
}

// Parallel benchmarks

func BenchmarkParserParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = parser.ParseString(mediumYAML)
		}
	})
}

func BenchmarkMarshalParallel(b *testing.B) {
	data := BenchmarkStruct{
		Name:    "test",
		Value:   123,
		Enabled: true,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = encoder.Marshal(data)
		}
	})
}

// Memory allocation benchmarks

func BenchmarkParserMemory(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(mediumYAML)
	}
}

func BenchmarkSerializerMemory(b *testing.B) {
	root, _ := parser.ParseString(mediumYAML)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = serializer.SerializeToString(root, nil)
	}
}

// Comparative benchmarks for different YAML styles

var (
	blockYAML = `
mapping:
  key1: value1
  key2: value2
  key3: value3
sequence:
  - item1
  - item2
  - item3`

	flowYAML = `
mapping: {key1: value1, key2: value2, key3: value3}
sequence: [item1, item2, item3]`
)

func BenchmarkParseBlockStyle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(blockYAML)
	}
}

func BenchmarkParseFlowStyle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(flowYAML)
	}
}

// Anchor/Alias benchmarks

var anchorYAML = `
defaults: &defaults
  timeout: 30
  retries: 3

service1:
  <<: *defaults
  name: Service1

service2:
  <<: *defaults
  name: Service2

service3:
  <<: *defaults
  name: Service3`

func BenchmarkParseWithAnchors(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(anchorYAML)
	}
}

// Comment handling benchmarks

var commentYAML = `
# Main configuration file
# Version: 1.0

server:  # Server configuration
  host: localhost  # Server host
  port: 8080       # Server port

# Database settings
database:
  # Connection parameters
  host: localhost
  port: 5432`

func BenchmarkParseWithComments(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(commentYAML)
	}
}
