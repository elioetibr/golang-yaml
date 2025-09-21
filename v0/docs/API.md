# API Documentation

## Package Overview

The YAML library consists of several packages, each with specific responsibilities:

| Package | Purpose |
|---------|---------|
| `pkg/lexer` | Tokenization of YAML input |
| `pkg/parser` | AST construction from tokens |
| `pkg/node` | Node interfaces and builders |
| `pkg/serializer` | YAML generation from AST |
| `pkg/encoder` | High-level Marshal API |
| `pkg/decoder` | High-level Unmarshal API with panic resistance |
| `pkg/transform` | Sorting and formatting |
| `pkg/merge` | YAML merging with configurable strategies |
| `pkg/errors` | Error handling utilities |

## Core Interfaces

### Node Interface

```go
type Node interface {
    Type() NodeType
    Value() interface{}
    SetValue(interface{})

    // Comment management
    HeadComment() string
    SetHeadComment(string)
    LineComment() string
    SetLineComment(string)
    FootComment() string
    SetFootComment(string)
}
```

### Builder Interface

```go
type Builder interface {
    BuildScalar(value string, style ScalarStyle) Node
    BuildSequence(items []Node, style Style) Node
    BuildMapping(pairs []*MappingPair, style Style) Node
    WithAnchor(node Node, anchor string) Node
    WithTag(node Node, tag string) Node
}
```

## Parser Package

### Functions

#### ParseString
```go
func ParseString(input string) (node.Node, error)
```
Parses a YAML string and returns the root node.

#### ParseStream
```go
func ParseStream(input string) (*DocumentStream, error)
```
Parses a multi-document YAML stream.

### Types

#### Parser
```go
type Parser struct {
    // Options
    StrictMode bool
    AllowTabs  bool
}

func NewParser() *Parser
func (p *Parser) Parse() (node.Node, error)
```

#### AnchorRegistry
```go
type AnchorRegistry struct {
    // Manages anchor definitions and references
}

func (r *AnchorRegistry) Define(name string, node node.Node)
func (r *AnchorRegistry) Resolve(name string) (node.Node, bool)
```

## Serializer Package

### Functions

#### SerializeToString
```go
func SerializeToString(node node.Node, opts *Options) (string, error)
```
Serializes a node tree to YAML string.

### Types

#### Options
```go
type Options struct {
    Indent           int
    PreserveComments bool
    SortKeys         bool
    LineWidth        int
    CompactSequence  bool
}
```

## Encoder Package

### Functions

#### Marshal
```go
func Marshal(v interface{}) ([]byte, error)
```
Encodes a Go value to YAML bytes.

### Types

#### Encoder
```go
type Encoder struct {
    writer io.Writer
}

func NewEncoder(w io.Writer) *Encoder
func (e *Encoder) Encode(v interface{}) error
```

## Decoder Package

The decoder package provides robust YAML to Go value conversion with comprehensive validation and error handling.

### Functions

#### Unmarshal
```go
func Unmarshal(data []byte, v interface{}) error
```
Decodes YAML bytes into a Go value. Features:
- Panic-resistant operation with comprehensive validation
- Supports all Go basic types, structs, maps, slices, and arrays
- Handles nil documents gracefully
- Case-insensitive field matching for structs
- YAML struct tags support (`yaml:"name,omitempty"` etc.)

#### UnmarshalStrict
```go
func UnmarshalStrict(data []byte, v interface{}) error
```
Like Unmarshal but returns an error when the destination has unknown fields.

### Types

#### Decoder
```go
type Decoder struct {
    reader io.Reader
    buffer []byte
}

func NewDecoder(r io.Reader) *Decoder
func (d *Decoder) Decode(v interface{}) error
```

### Error Handling

The decoder provides detailed error messages for:
- Invalid reflect.Value operations
- Type conversion failures
- Non-settable destination values
- Unsupported type combinations

### Validation Features

- **Nil Safety**: Handles nil nodes and empty documents gracefully
- **Type Validation**: Validates destination types before setting values
- **Reflection Safety**: Checks if values can be set before attempting operations
- **Graceful Degradation**: Skips invalid fields rather than failing entirely

## Merge Package

The merge package provides comprehensive YAML merging capabilities with multiple strategies and configurable options.

### Core Functions

#### Merge
```go
func Merge(base, override node.Node) (node.Node, error)
```
Combines two YAML nodes using the default deep merge strategy.

#### MergeWithOptions
```go
func MergeWithOptions(base, override node.Node, opts *Options) (node.Node, error)
```
Combines two YAML nodes with specified options for full control over merge behavior.

#### MergeStrings
```go
func MergeStrings(baseYAML, overrideYAML string) (string, error)
```
Merges two YAML strings and returns the result as a string.

#### MergeStringsWithOptions
```go
func MergeStringsWithOptions(baseYAML, overrideYAML string, opts *Options) (string, error)
```
Merges two YAML strings with specified options.

#### MergeFiles
```go
func MergeFiles(basePath, overridePath string) (string, error)
func MergeFilesWithOptions(basePath, overridePath string, opts *Options) (string, error)
```
Merges two YAML files and returns the result as a string.

#### MergeFilesToFile
```go
func MergeFilesToFile(basePath, overridePath, outputPath string) error
func MergeFilesToFileWithOptions(basePath, overridePath, outputPath string, opts *Options) error
```
Merges two YAML files and writes the result to a file.

#### MergeMultiple
```go
func MergeMultiple(nodes []node.Node) (node.Node, error)
func MergeMultipleWithOptions(nodes []node.Node, opts *Options) (node.Node, error)
```
Merges multiple YAML nodes in sequence.

### Merge Strategies

#### Strategy Types
```go
type Strategy int

const (
    StrategyDeep     Strategy = iota  // Deep recursive merging
    StrategyShallow                   // Only merge top-level keys
    StrategyOverride                  // Replace base with override
    StrategyAppend                    // Append arrays instead of replacing
)
```

#### Array Merge Strategies
```go
type ArrayMergeStrategy int

const (
    ArrayReplace      ArrayMergeStrategy = iota  // Replace entire array
    ArrayAppend                                   // Append override to base
    ArrayMergeByIndex                            // Merge element by element
    ArrayMergeByKey                              // Merge array of maps by key
)
```

#### Key Priority
```go
type KeyPriority int

const (
    KeyPriorityBase         KeyPriority = iota  // Preserve base ordering
    KeyPriorityOverride                         // Use override ordering
    KeyPriorityAlphabetical                     // Sort keys alphabetically
)
```

### Options Configuration

#### Options
```go
type Options struct {
    Strategy           Strategy
    PreserveComments   bool
    PreserveBlankLines bool
    ArrayMergeStrategy ArrayMergeStrategy
    OverrideEmpty      bool
    MergeAnchors       bool
    CustomMergeFunc    func(key string, base, override node.Node) (node.Node, bool)
    KeyPriority        KeyPriority
}

func DefaultOptions() *Options
```

#### Option Builders
```go
func (o *Options) WithStrategy(s Strategy) *Options
func (o *Options) WithArrayStrategy(s ArrayMergeStrategy) *Options
func (o *Options) WithKeyPriority(p KeyPriority) *Options
func (o *Options) WithOverrideEmpty(override bool) *Options
```

### Merge Types

#### Merger
```go
type Merger struct {
    // Internal fields
}

func NewMerger(opts *Options) *Merger
func (m *Merger) Merge(base, override node.Node) (node.Node, error)
```

#### Context
```go
type Context struct {
    Options *Options
    Depth   int
    Path    []string
}

func (c *Context) WithPath(segment string) *Context
```

#### MergeStrategy Interface
```go
type MergeStrategy interface {
    Name() string
    Merge(base, override node.Node, ctx *Context) (node.Node, error)
}
```

## Transform Package

### Sorting

#### SortConfig
```go
type SortConfig struct {
    Mode           SortMode
    SortBy         SortBy
    Recursive      bool
    CaseSensitive  bool
    Priority       []string
    Groups         [][]string
    ExcludePaths   []string
}
```

#### SortMode
```go
const (
    SortModeKeepOriginal SortMode = iota
    SortModeAscending
    SortModeDescending
)
```

#### Sorter
```go
type Sorter struct {
    config *SortConfig
}

func NewSorter(config *SortConfig) *Sorter
func (s *Sorter) Sort(node node.Node) node.Node
```

### Formatting

#### FormatConfig
```go
type FormatConfig struct {
    BlankLinesBeforeComment int
    ForceBlankLines        bool
    PreserveOriginal       bool
    SectionMarkers         []string
}
```

#### Formatter
```go
type Formatter struct {
    config *FormatConfig
}

func NewFormatter(config *FormatConfig) *Formatter
func (f *Formatter) Format(node node.Node) node.Node
```

## Error Handling

### YAMLError
```go
type YAMLError struct {
    Message  string
    Position Position
    Type     ErrorType
}

type Position struct {
    Line   int
    Column int
    Offset int
}

type ErrorType int

const (
    ErrorTypeLexer ErrorType = iota
    ErrorTypeParser
    ErrorTypeSerializer
    ErrorTypeEncoder
    ErrorTypeDecoder
)
```

## Struct Tags

The library supports the following struct tags:

```go
type Example struct {
    Field1 string `yaml:"field1"`           // Custom name
    Field2 string `yaml:"field2,omitempty"` // Omit if empty
    Field3 string `yaml:"-"`                // Skip field
    Field4 string `yaml:",inline"`          // Inline map
}
```

## Thread Safety

### Package-Level Thread Safety

| Package | Thread Safety | Notes |
|---------|---------------|-------|
| `pkg/parser` | **Not thread-safe** | Create new instance per goroutine |
| `pkg/serializer` | **Thread-safe for read operations** | Concurrent serialization supported |
| `pkg/node` | **Thread-safe** | Immutable after creation |
| `pkg/encoder` | **Not thread-safe** | Use sync.Pool for reuse across goroutines |
| `pkg/decoder` | **Not thread-safe** | Use sync.Pool for reuse across goroutines |
| `pkg/merge` | **Not thread-safe** | Create new merger instance per operation |
| `pkg/transform` | **Thread-safe** | Stateless operations |

### Concurrent Usage Patterns

#### Safe Concurrent Processing
```go
// Worker pool pattern for concurrent YAML processing
func processYAMLFiles(files []string, workers int) error {
    filesChan := make(chan string, len(files))
    errorsChan := make(chan error, workers)

    // Start workers
    for i := 0; i < workers; i++ {
        go func() {
            // Each worker gets its own decoder
            for file := range filesChan {
                data, err := os.ReadFile(file)
                if err != nil {
                    errorsChan <- err
                    continue
                }

                var config Config
                err = decoder.Unmarshal(data, &config)
                errorsChan <- err
            }
        }()
    }

    // Send work
    for _, file := range files {
        filesChan <- file
    }
    close(filesChan)

    // Collect errors
    for i := 0; i < len(files); i++ {
        if err := <-errorsChan; err != nil {
            return err
        }
    }

    return nil
}
```

#### Using sync.Pool for Decoder Reuse
```go
var decoderPool = sync.Pool{
    New: func() interface{} {
        return &decoder.Decoder{}
    },
}

func processWithPooledDecoder(data []byte, v interface{}) error {
    dec := decoderPool.Get().(*decoder.Decoder)
    defer decoderPool.Put(dec)

    // Reset decoder state
    dec.Reset(bytes.NewReader(data))
    return dec.Decode(v)
}
```

## Performance Considerations

### Memory Usage

| Operation | Memory Impact | Optimization |
|-----------|---------------|--------------|
| **Large Document Parsing** | High | Use streaming parser for >10MB files |
| **Comment Preservation** | +20-30% | Disable if not needed |
| **Deep Merging** | Moderate | Consider shallow merge for simple cases |
| **Node Tree Creation** | High | Reuse builders and avoid deep nesting |

### Performance Optimizations

#### 1. Streaming for Large Documents
```go
// For large files, use streaming to reduce memory usage
func processLargeYAML(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    decoder := decoder.NewDecoder(file)

    for {
        var doc interface{}
        err := decoder.Decode(&doc)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        // Process document in chunks
        if err := processDocument(doc); err != nil {
            return err
        }
    }

    return nil
}
```

#### 2. Parser Instance Reuse
```go
// Reuse parser instances for better performance
type YAMLProcessor struct {
    parser    *parser.Parser
    serializer *serializer.Serializer
}

func NewYAMLProcessor() *YAMLProcessor {
    return &YAMLProcessor{
        parser:     parser.NewParser(),
        serializer: serializer.NewSerializer(),
    }
}

func (p *YAMLProcessor) Process(input string) (string, error) {
    node, err := p.parser.ParseString(input)
    if err != nil {
        return "", err
    }

    return p.serializer.SerializeToString(node, nil)
}
```

#### 3. Merge Performance
```go
// For repeated merges, create merger once
merger := merge.NewMerger(&merge.Options{
    Strategy:           merge.StrategyDeep,
    PreserveComments:   false, // Disable for better performance
    PreserveBlankLines: false,
})

// Reuse for multiple operations
for _, override := range overrides {
    result, err := merger.Merge(base, override)
    if err != nil {
        return err
    }
    // Process result
}
```

### Benchmarking Guidelines

#### CPU Performance
- **Parsing**: ~100-500 MB/s for typical YAML documents
- **Serialization**: ~200-800 MB/s depending on options
- **Merging**: ~50-200 MB/s depending on strategy and document complexity

#### Memory Usage
- **Basic parsing**: ~2-3x input size in memory
- **With comments**: ~3-4x input size in memory
- **Deep merge**: ~4-5x combined input size during operation

## Best Practices

### Error Handling
1. **Always check errors** from Parse/Serialize operations
2. **Use typed errors** for better error handling
3. **Provide context** in error messages
4. **Handle partial failures** gracefully in batch operations

### Memory Management
1. **Use streaming API** for large documents (>10MB)
2. **Disable comment preservation** if not needed
3. **Reuse instances** with sync.Pool for high-throughput applications
4. **Profile memory usage** in production environments

### Performance
1. **Reuse Parser/Serializer instances** when possible
2. **Use appropriate merge strategies** for your use case
3. **Consider shallow merge** for simple overrides
4. **Benchmark with realistic data** for your specific use case

### Validation
1. **Enable StrictMode** for spec compliance when needed
2. **Validate input types** before unmarshaling
3. **Use UnmarshalStrict** when you need to catch unknown fields
4. **Implement custom validation** for business logic

### Comments and Formatting
1. **Use PreserveComments** for round-trip parsing
2. **Use PreserveBlankLines** for maintaining document structure
3. **Consider formatting impact** on performance
4. **Test comment preservation** with your specific YAML patterns

## Examples

### Merge Examples

#### Basic Deep Merge
```go
baseYAML := `
name: myapp
version: 1.0.0
server:
  host: localhost
  port: 8080
database:
  host: localhost
  port: 5432`

overrideYAML := `
server:
  host: api.example.com
  port: 443
database:
  host: db.example.com`

result, err := merge.MergeStrings(baseYAML, overrideYAML)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)
// Output includes merged values with preserved comments
```

#### Array Merge Strategies
```go
baseYAML := `
features:
  - auth
  - api
  - logging`

overrideYAML := `
features:
  - monitoring
  - metrics`

// Replace arrays (default)
replaceResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
    merge.DefaultOptions().WithArrayStrategy(merge.ArrayReplace))

// Append arrays
appendResult, _ := merge.MergeStringsWithOptions(baseYAML, overrideYAML,
    merge.DefaultOptions().WithArrayStrategy(merge.ArrayAppend))
```

#### Custom Merge with Options
```go
opts := &merge.Options{
    Strategy:           merge.StrategyDeep,
    PreserveComments:   true,
    PreserveBlankLines: true,
    ArrayMergeStrategy: merge.ArrayAppend,
    OverrideEmpty:      false,
    KeyPriority:        merge.KeyPriorityBase,
}

result, err := merge.MergeStringsWithOptions(baseYAML, overrideYAML, opts)
```

#### File Operations
```go
// Merge files and get result as string
result, err := merge.MergeFiles("base.yaml", "override.yaml")

// Merge files and write to output file
err := merge.MergeFilesToFile("base.yaml", "override.yaml", "result.yaml")

// Multiple file merge
nodes := []string{yaml1, yaml2, yaml3}
result := yaml1
for i := 1; i < len(nodes); i++ {
    merged, err := merge.MergeStrings(result, nodes[i])
    if err != nil {
        return err
    }
    result = merged
}
```

### Decoder Examples

#### Basic Unmarshaling
```go
type Config struct {
    Name     string            `yaml:"name"`
    Version  string            `yaml:"version,omitempty"`
    Features []string          `yaml:"features"`
    Settings map[string]string `yaml:"settings"`
}

yamlData := []byte(`
name: myapp
version: 1.0.0
features:
  - auth
  - api
settings:
  timeout: "30s"
  retries: "3"`)

var config Config
err := decoder.Unmarshal(yamlData, &config)
if err != nil {
    log.Fatal(err)
}
```

#### Error Handling Patterns
```go
// Robust unmarshaling with error handling
func safeUnmarshal(data []byte, v interface{}) error {
    err := decoder.Unmarshal(data, v)
    if err != nil {
        // Handle specific error types
        if strings.Contains(err.Error(), "cannot unmarshal") {
            return fmt.Errorf("type mismatch: %w", err)
        }
        if strings.Contains(err.Error(), "cannot set value") {
            return fmt.Errorf("destination not settable: %w", err)
        }
        return fmt.Errorf("unmarshal failed: %w", err)
    }
    return nil
}
```

#### Stream Decoding
```go
reader := strings.NewReader(yamlContent)
decoder := decoder.NewDecoder(reader)

for {
    var doc interface{}
    err := decoder.Decode(&doc)
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Printf("Decode error: %v", err)
        continue
    }
    // Process document
    processDocument(doc)
}
```

### Custom Node Building

```go
builder := &node.DefaultBuilder{}

// Create a mapping
mapping := builder.BuildMapping([]*node.MappingPair{
    {
        Key:   builder.BuildScalar("name", node.StylePlain),
        Value: builder.BuildScalar("example", node.StylePlain),
    },
    {
        Key: builder.BuildScalar("items", node.StylePlain),
        Value: builder.BuildSequence([]node.Node{
            builder.BuildScalar("item1", node.StylePlain),
            builder.BuildScalar("item2", node.StylePlain),
        }, node.StyleBlock),
    },
}, node.StyleBlock)

// Add anchor
mapping = builder.WithAnchor(mapping, "example")
```

### Stream Processing

```go
stream, err := parser.ParseStream(multiDocYAML)
if err != nil {
    return err
}

for _, doc := range stream.Documents {
    // Process each document
    output, err := serializer.SerializeToString(doc.Root, nil)
    if err != nil {
        return err
    }
    fmt.Println(output)
}
```

### Custom Sorting

```go
config := &transform.SortConfig{
    Mode:     transform.SortModeAscending,
    SortBy:   transform.SortByKey,
    Priority: []string{"apiVersion", "kind", "metadata"},
    Groups: [][]string{
        {"spec", "status"},
        {"data", "stringData"},
    },
}

sorter := transform.NewSorter(config)
sorted := sorter.Sort(root)
```

### Error Handling Best Practices

#### Comprehensive Error Handling
```go
func processYAMLFile(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }

    // Parse with error context
    node, err := parser.ParseString(string(data))
    if err != nil {
        if yamlErr, ok := err.(*errors.YAMLError); ok {
            return fmt.Errorf("parse error at line %d, column %d: %s",
                yamlErr.Position.Line, yamlErr.Position.Column, yamlErr.Message)
        }
        return fmt.Errorf("parse error: %w", err)
    }

    // Validate structure
    if node == nil {
        return fmt.Errorf("empty document")
    }

    return nil
}
```

#### Recovery from Invalid Operations
```go
func safeNodeAccess(n node.Node, key string) (node.Node, error) {
    mapping, ok := n.(*node.MappingNode)
    if !ok {
        return nil, fmt.Errorf("node is not a mapping")
    }

    for _, pair := range mapping.Pairs {
        if scalar, ok := pair.Key.(*node.ScalarNode); ok && scalar.Value == key {
            return pair.Value, nil
        }
    }

    return nil, fmt.Errorf("key %q not found", key)
}
```