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
| `pkg/decoder` | High-level Unmarshal API |
| `pkg/transform` | Sorting and formatting |
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

### Functions

#### Unmarshal
```go
func Unmarshal(data []byte, v interface{}) error
```
Decodes YAML bytes into a Go value.

### Types

#### Decoder
```go
type Decoder struct {
    reader io.Reader
}

func NewDecoder(r io.Reader) *Decoder
func (d *Decoder) Decode(v interface{}) error
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

- Parser: Not thread-safe (create new instance per goroutine)
- Serializer: Thread-safe for read operations
- Nodes: Immutable after creation (thread-safe)
- Encoder/Decoder: Not thread-safe (use sync.Pool for reuse)

## Best Practices

1. **Error Handling**: Always check errors from Parse/Serialize operations
2. **Memory**: For large documents, use streaming API
3. **Performance**: Reuse Parser/Serializer instances when possible
4. **Comments**: Use PreserveComments option for round-trip parsing
5. **Validation**: Enable StrictMode for spec compliance

## Examples

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