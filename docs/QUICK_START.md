# Quick Start Guide

## Installation

```bash
go get github.com/elioetibr/golang-yaml
```

## Basic Operations

### 1. Parse YAML String

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/parser"
)

func main() {
    yamlStr := `
name: MyApp
version: 1.0.0
features:
  - logging
  - metrics
  - caching
`

    root, err := parser.ParseString(yamlStr)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Parsed successfully!")
}
```

### 2. Marshal Go Struct to YAML

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/encoder"
)

type Config struct {
    Name     string   `yaml:"name"`
    Version  string   `yaml:"version"`
    Features []string `yaml:"features"`
}

func main() {
    config := Config{
        Name:     "MyApp",
        Version:  "1.0.0",
        Features: []string{"logging", "metrics", "caching"},
    }

    data, err := encoder.Marshal(config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(data))
}
```

### 3. Unmarshal YAML to Go Struct

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/decoder"
)

type Config struct {
    Name     string   `yaml:"name"`
    Version  string   `yaml:"version"`
    Features []string `yaml:"features"`
}

func main() {
    yamlData := []byte(`
name: MyApp
version: 1.0.0
features:
  - logging
  - metrics
  - caching
`)

    var config Config
    err := decoder.Unmarshal(yamlData, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Name: %s\n", config.Name)
    fmt.Printf("Version: %s\n", config.Version)
    fmt.Printf("Features: %v\n", config.Features)
}
```

### 4. Sort YAML Content

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/parser"
    "yaml/pkg/serializer"
    "yaml/pkg/transform"
)

func main() {
    yamlStr := `
zoo: animals
bar: drinks
foo: food
apple: fruit
`

    // Parse
    root, err := parser.ParseString(yamlStr)
    if err != nil {
        log.Fatal(err)
    }

    // Sort
    config := &transform.SortConfig{
        Mode:   transform.SortModeAscending,
        SortBy: transform.SortByKey,
    }
    sorter := transform.NewSorter(config)
    sorted := sorter.Sort(root)

    // Serialize back
    output, err := serializer.SerializeToString(sorted, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output)
    // Output:
    // apple: fruit
    // bar: drinks
    // foo: food
    // zoo: animals
}
```

### 5. Preserve Comments

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/parser"
    "yaml/pkg/serializer"
)

func main() {
    yamlStr := `
# Application configuration
name: MyApp  # The application name

# Version information
version: 1.0.0
`

    // Parse
    root, err := parser.ParseString(yamlStr)
    if err != nil {
        log.Fatal(err)
    }

    // Serialize with comments preserved
    opts := &serializer.Options{
        PreserveComments: true,
        Indent:          2,
    }

    output, err := serializer.SerializeToString(root, opts)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output)
    // Comments are preserved!
}
```

### 6. Handle Multiple Documents

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/parser"
)

func main() {
    multiDoc := `
---
document: 1
type: config
---
document: 2
type: data
---
document: 3
type: metadata
...
`

    stream, err := parser.ParseStream(multiDoc)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d documents\n", len(stream.Documents))

    for i, doc := range stream.Documents {
        fmt.Printf("Document %d: %v\n", i+1, doc.Root)
    }
}
```

### 7. Work with Anchors and Aliases

```go
package main

import (
    "fmt"
    "log"
    "yaml/pkg/parser"
    "yaml/pkg/serializer"
)

func main() {
    yamlStr := `
defaults: &defaults
  timeout: 30
  retries: 3
  backoff: exponential

service1:
  <<: *defaults
  endpoint: /api/v1/users

service2:
  <<: *defaults
  endpoint: /api/v1/posts
  timeout: 60  # Override default
`

    root, err := parser.ParseString(yamlStr)
    if err != nil {
        log.Fatal(err)
    }

    // Anchors and aliases are automatically resolved
    output, err := serializer.SerializeToString(root, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output)
}
```

## Common Patterns

### Configuration File Loading

```go
func LoadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := decoder.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

### Configuration File Saving

```go
func SaveConfig(filename string, config *Config) error {
    data, err := encoder.Marshal(config)
    if err != nil {
        return err
    }

    return os.WriteFile(filename, data, 0644)
}
```

### Validate YAML Structure

```go
func ValidateYAML(yamlStr string) error {
    _, err := parser.ParseString(yamlStr)
    return err
}
```

### Merge Multiple YAML Files

```go
func MergeYAMLFiles(files ...string) (node.Node, error) {
    var nodes []node.Node

    for _, file := range files {
        data, err := os.ReadFile(file)
        if err != nil {
            return nil, err
        }

        node, err := parser.ParseString(string(data))
        if err != nil {
            return nil, err
        }

        nodes = append(nodes, node)
    }

    // Merge logic here...
    return nodes[0], nil // Simplified
}
```

## Tips and Best Practices

1. **Always handle errors** - YAML parsing can fail on malformed input
2. **Use struct tags** - Control field names and behavior with `yaml:` tags
3. **Preserve formatting** - Use `PreserveComments: true` for config files
4. **Sort consistently** - Use sorting for reproducible output
5. **Validate early** - Parse and validate YAML at startup
6. **Use appropriate types** - YAML supports various scalar types
7. **Handle special values** - Be aware of null, true, false handling

## Next Steps

- Read the [API Documentation](../v1/docs/API.md) for detailed reference
- Check [examples](../examples/) for more complex use cases
- See [YAML 1.2.2 Compliance](YAML_1.2.2_COMPLIANCE.md) for spec details
- Review the [ROADMAP](ROADMAP.md) for upcoming features