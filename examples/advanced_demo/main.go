package main

import (
	"fmt"
	"log"

	"github.com/elioetibr/golang-yaml/pkg/decoder"
	"github.com/elioetibr/golang-yaml/pkg/encoder"
	"github.com/elioetibr/golang-yaml/pkg/node"
	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
	"github.com/elioetibr/golang-yaml/pkg/transform"
)

func main() {
	fmt.Println("=== Advanced YAML Library Demo ===")

	// 1. Demonstrate Anchor/Alias Resolution
	fmt.Println("1. Anchor/Alias Resolution:")
	anchorDemo()

	// 2. Demonstrate Merge Keys
	fmt.Println("\n2. Merge Key Support:")
	mergeKeyDemo()

	// 3. Demonstrate Multi-Document Streams
	fmt.Println("\n3. Multi-Document Streams:")
	multiDocDemo()

	// 4. Demonstrate Sorting with Priority
	fmt.Println("\n4. Advanced Sorting:")
	sortingDemo()

	// 5. Demonstrate Comment Preservation
	fmt.Println("\n5. Comment Preservation:")
	commentDemo()

	// 6. Demonstrate Tag Support
	fmt.Println("\n6. Tag Support:")
	tagDemo()
}

func anchorDemo() {
	yaml := `
base_config: &base
  timeout: 30
  retries: 3
  log_level: info

development:
  <<: *base
  host: localhost
  log_level: debug

production:
  <<: *base
  host: api.example.com
  timeout: 60
`

	root, err := parser.ParseString(yaml)
	if err != nil {
		log.Fatal(err)
	}

	output, _ := serializer.SerializeToString(root, nil)
	fmt.Println(output)
}

func mergeKeyDemo() {
	yaml := `
defaults: &defaults
  cpu: 100m
  memory: 128Mi

containers:
  - name: web
    <<: *defaults
    image: nginx:latest
    cpu: 200m  # Override

  - name: api
    <<: *defaults
    image: node:16
    memory: 256Mi  # Override
`

	root, err := parser.ParseString(yaml)
	if err != nil {
		log.Fatal(err)
	}

	// The merge keys are automatically resolved
	output, _ := serializer.SerializeToString(root, nil)
	fmt.Println(output)
}

func multiDocDemo() {
	yaml := `---
# First document
apiVersion: v1
kind: Service
metadata:
  name: web-service
---
# Second document
apiVersion: v1
kind: Deployment
metadata:
  name: web-deployment
---
# Third document
apiVersion: v1
kind: ConfigMap
metadata:
  name: web-config
...`

	stream, err := parser.ParseStream(yaml)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed %d documents:\n", len(stream.Documents))
	for i, doc := range stream.Documents {
		fmt.Printf("  Document %d: ", i+1)
		// Serialize just the first level to show structure
		output, _ := serializer.SerializeToString(doc.Root, &serializer.Options{
			Indent: 2,
		})
		// Print first line only for demo
		if len(output) > 50 {
			fmt.Printf("%s...\n", output[:50])
		} else {
			fmt.Println(output)
		}
	}
}

func sortingDemo() {
	yaml := `
zoo: animals
metadata:
  name: example
  labels:
    env: prod
apiVersion: v1
kind: Service
spec:
  ports:
    - port: 80
data:
  key: value
`

	root, err := parser.ParseString(yaml)
	if err != nil {
		log.Fatal(err)
	}

	// Sort with Kubernetes-style priority
	config := &transform.SortConfig{
		Mode:     transform.SortModeAscending,
		SortBy:   transform.SortByKey,
		Scope:    transform.SortScopeNested,
		Priority: []string{"apiVersion", "kind", "metadata", "spec", "data"},
	}

	sorter := transform.NewSorter(config)
	sorted := sorter.Sort(root)

	output, _ := serializer.SerializeToString(sorted, nil)
	fmt.Println("Sorted with priority:")
	fmt.Println(output)
}

func commentDemo() {
	yaml := `# Application Configuration
# Version: 2.0

server:  # Server settings
  host: localhost  # Bind address
  port: 8080       # Listen port

# Database configuration
database:
  # Connection string
  url: postgres://localhost/mydb
  pool_size: 10  # Connection pool size
`

	root, err := parser.ParseString(yaml)
	if err != nil {
		log.Fatal(err)
	}

	// Serialize with comments preserved
	opts := &serializer.Options{
		PreserveComments: true,
		Indent:           2,
	}

	output, _ := serializer.SerializeToString(root, opts)
	fmt.Println("With preserved comments:")
	fmt.Println(output)
}

func tagDemo() {
	yaml := `
# Custom tags example
binary: !!binary "SGVsbG8gV29ybGQ="
timestamp: !!timestamp 2024-01-15T10:30:00Z
set: !!set
  ? item1
  ? item2
  ? item3
custom: !CustomType
  field: value
`

	root, err := parser.ParseString(yaml)
	if err != nil {
		log.Fatal(err)
	}

	// Tags are parsed and preserved
	output, _ := serializer.SerializeToString(root, nil)
	fmt.Println("Tags preserved:")
	fmt.Println(output)
}

// AppConfig Example: Custom struct marshaling/unmarshaling
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Features []string       `yaml:"features,omitempty"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	URL      string `yaml:"url"`
	PoolSize int    `yaml:"pool_size"`
}

func structDemo() {
	config := AppConfig{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			URL:      "postgres://localhost/mydb",
			PoolSize: 20,
		},
		Features: []string{"auth", "api", "websocket"},
	}

	// Marshal to YAML
	data, err := encoder.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Marshaled struct:")
	fmt.Println(string(data))

	// Unmarshal back
	var loaded AppConfig
	err = decoder.Unmarshal(data, &loaded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded: %+v\n", loaded)
}

// Example: Building nodes programmatically
func builderDemo() {
	builder := &node.DefaultBuilder{}

	// Build a complex structure
	root := builder.BuildMapping([]*node.MappingPair{
		{
			Key:   builder.BuildScalar("apiVersion", node.StylePlain),
			Value: builder.BuildScalar("v1", node.StylePlain),
		},
		{
			Key:   builder.BuildScalar("kind", node.StylePlain),
			Value: builder.BuildScalar("ConfigMap", node.StylePlain),
		},
		{
			Key: builder.BuildScalar("metadata", node.StylePlain),
			Value: builder.BuildMapping([]*node.MappingPair{
				{
					Key:   builder.BuildScalar("name", node.StylePlain),
					Value: builder.BuildScalar("app-config", node.StylePlain),
				},
				{
					Key: builder.BuildScalar("labels", node.StylePlain),
					Value: builder.BuildMapping([]*node.MappingPair{
						{
							Key:   builder.BuildScalar("app", node.StylePlain),
							Value: builder.BuildScalar("myapp", node.StylePlain),
						},
					}, node.StyleBlock),
				},
			}, node.StyleBlock),
		},
		{
			Key: builder.BuildScalar("data", node.StylePlain),
			Value: builder.BuildMapping([]*node.MappingPair{
				{
					Key:   builder.BuildScalar("config.yaml", node.StylePlain),
					Value: builder.BuildScalar("key: value", node.StyleLiteral),
				},
			}, node.StyleBlock),
		},
	}, node.StyleBlock)

	// Add anchor
	rootWithAnchor := builder.WithAnchor(root, "config")

	// Serialize
	output, _ := serializer.SerializeToString(rootWithAnchor, nil)
	fmt.Println("Programmatically built:")
	fmt.Println(output)
}
