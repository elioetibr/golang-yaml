# Feature #01: YAML Merge Functionality

## Executive Summary

Implement a comprehensive YAML merging capability in the golang-yaml library that allows combining multiple YAML documents with configurable strategies while adhering to SOLID principles and maintaining clean architecture.

## Problem Statement

Currently, users must implement custom merge logic when combining YAML configurations from multiple sources. This is a critical requirement for:
- **Helm Charts**: Merging values.yaml with environment-specific overrides
- **Kubernetes**: Combining base configurations with patches
- **CI/CD**: Layering pipeline configurations
- **Microservices**: Managing multi-environment configurations

### Critical Requirements

1. **Preserve Documentation**: Comments in YAML files often contain critical documentation, schema information, and usage instructions that MUST be preserved
2. **Maintain Readability**: Blank lines improve readability and logical grouping - they should be maintained by default
3. **Respect Structure**: The original document's structure and organization should be preserved unless explicitly changed

## Design Following Core Principles

### 1. SOLID Principles Application

#### Single Responsibility Principle (SRP)
Each component has one clear responsibility:
- `Merger`: Orchestrates the merge process
- `Strategy`: Defines how merging occurs
- `NodeMerger`: Handles node-specific merge logic
- `Validator`: Validates merge operations

#### Open/Closed Principle (OCP)
The design is extensible without modification:
```go
// Strategy interface - open for extension
type MergeStrategy interface {
    Merge(base, override node.Node, ctx *MergeContext) (node.Node, error)
    CanMerge(base, override node.Node) bool
}

// New strategies can be added without modifying existing code
type CustomStrategy struct{}
func (c *CustomStrategy) Merge(...) {...}
```

#### Liskov Substitution Principle (LSP)
All strategies are interchangeable:
```go
var strategy MergeStrategy
strategy = &DeepMergeStrategy{}
strategy = &ShallowMergeStrategy{}
strategy = &OverrideStrategy{}
// All work with the same interface
```

#### Interface Segregation Principle (ISP)
Small, focused interfaces:
```go
// Separate interfaces for different concerns
type Merger interface {
    Merge(base, override node.Node) (node.Node, error)
}

type Validator interface {
    Validate(node.Node) error
}

type Formatter interface {
    PreserveFormatting(node.Node) node.Node
}
```

#### Dependency Inversion Principle (DIP)
Depend on abstractions:
```go
type MergeService struct {
    strategy  MergeStrategy    // Interface, not concrete
    validator Validator        // Interface, not concrete
    formatter Formatter        // Interface, not concrete
}
```

### 2. KISS - Keep It Simple

Simple API for common use cases with sensible defaults:
```go
// Simple function for 80% of use cases
// Defaults: Deep merge, preserve comments & blank lines
result, err := merge.Simple(base, override)

// Advanced API when needed
result, err := merge.WithOptions(base, override, opts)
```

### 3. DRY - Don't Repeat Yourself

Reusable components:
```go
// Shared node processing logic
type NodeProcessor struct {
    cleanScalarHeadComment(node.Node) node.Node
    preserveComments(base, override node.Node) node.Node
    mergeMetadata(base, override *node.BaseNode) *node.BaseNode
}
```

### 4. Dependency Injection

Constructor injection for testability:
```go
func NewMerger(
    strategy MergeStrategy,
    opts ...MergerOption,
) *Merger {
    m := &Merger{
        strategy: strategy,
        // Default dependencies
        validator: DefaultValidator(),
        formatter: DefaultFormatter(),
    }

    // Apply options to inject custom dependencies
    for _, opt := range opts {
        opt(m)
    }
    return m
}

// Options for dependency injection
func WithValidator(v Validator) MergerOption {
    return func(m *Merger) {
        m.validator = v
    }
}
```

### 5. Must Pattern for Error Handling

Chainable operations with deferred error handling:
```go
type MergeBuilder struct {
    base     node.Node
    override node.Node
    err      error
}

func (b *MergeBuilder) Must() node.Node {
    if b.err != nil {
        panic(b.err)
    }
    return b.base
}

// Usage
result := NewMergeBuilder().
    WithBase(baseYAML).
    WithOverride(overrideYAML).
    WithStrategy(DeepMerge).
    Merge().
    Must()
```

### 6. Layered Architecture

Clear separation of concerns:

```
┌─────────────────────────────────────┐
│         API Layer                   │  - Public interfaces
├─────────────────────────────────────┤
│         Service Layer               │  - Business logic
├─────────────────────────────────────┤
│         Strategy Layer              │  - Merge strategies
├─────────────────────────────────────┤
│         Core Layer                  │  - Node operations
└─────────────────────────────────────┘
```

## Proposed Package Structure

```
pkg/merge/
├── merge.go           # Public API (API Layer)
├── service.go         # MergeService (Service Layer)
├── strategies/        # Strategy implementations
│   ├── strategy.go    # Strategy interface
│   ├── deep.go        # Deep merge strategy
│   ├── shallow.go     # Shallow merge strategy
│   ├── override.go    # Override strategy
│   └── custom.go      # Custom strategy support
├── core/              # Core operations
│   ├── processor.go   # Node processing
│   ├── validator.go   # Validation logic
│   └── formatter.go   # Formatting preservation
├── options.go         # Configuration
├── builder.go         # Fluent API builder
├── errors.go          # Custom error types
└── merge_test.go      # Tests
```

## Implementation Architecture

### Core Interfaces

```go
// Package merge provides YAML merging capabilities
package merge

// Strategy defines the merge behavior
type Strategy interface {
    Name() string
    Merge(base, override node.Node, ctx *Context) (node.Node, error)
    CanMerge(base, override node.Node) bool
}

// Context carries merge operation context
type Context struct {
    Options         *Options
    Depth          int
    Path           []string
    CustomHandlers map[string]Handler
}

// Handler for custom merge logic
type Handler func(key string, base, override node.Node) (node.Node, bool)

// Merger orchestrates the merge process
type Merger interface {
    Merge(base, override node.Node) (node.Node, error)
    MergeStrings(base, override string) (string, error)
    MergeFiles(basePath, overridePath string) (string, error)
}
```

### Service Implementation

```go
// MergeService implements Merger with dependency injection
type MergeService struct {
    strategy   Strategy
    validator  Validator
    formatter  Formatter
    parser     Parser
    serializer Serializer
}

// NewMergeService creates a new service with dependencies
func NewMergeService(opts ...ServiceOption) *MergeService {
    svc := &MergeService{
        strategy:   NewDeepMergeStrategy(),
        validator:  NewDefaultValidator(),
        formatter:  NewDefaultFormatter(),
        parser:     parser.NewParser(),
        serializer: serializer.NewSerializer(),
    }

    for _, opt := range opts {
        opt(svc)
    }

    return svc
}

// Merge implements the main merge logic
func (s *MergeService) Merge(base, override node.Node) (node.Node, error) {
    // Validate inputs
    if err := s.validator.ValidateNodes(base, override); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Create context
    ctx := &Context{
        Options: s.options,
        Depth:   0,
        Path:    []string{},
    }

    // Perform merge
    result, err := s.strategy.Merge(base, override, ctx)
    if err != nil {
        return nil, fmt.Errorf("merge failed: %w", err)
    }

    // Preserve formatting
    if s.options.PreserveFormatting {
        result = s.formatter.PreserveFormatting(result)
    }

    return result, nil
}
```

### Strategy Pattern Implementation

```go
// DeepMergeStrategy performs recursive deep merging
type DeepMergeStrategy struct {
    processor *NodeProcessor
}

func (s *DeepMergeStrategy) Name() string {
    return "deep"
}

func (s *DeepMergeStrategy) Merge(base, override node.Node, ctx *Context) (node.Node, error) {
    if override == nil {
        return base, nil
    }
    if base == nil {
        return override, nil
    }

    // Type-specific merging
    switch baseNode := base.(type) {
    case *node.MappingNode:
        return s.mergeMappings(baseNode, override, ctx)
    case *node.SequenceNode:
        return s.mergeSequences(baseNode, override, ctx)
    case *node.ScalarNode:
        return s.mergeScalars(baseNode, override, ctx)
    default:
        return override, nil
    }
}

func (s *DeepMergeStrategy) mergeMappings(
    base *node.MappingNode,
    override node.Node,
    ctx *Context,
) (*node.MappingNode, error) {
    overrideMapping, ok := override.(*node.MappingNode)
    if !ok {
        return nil, fmt.Errorf("type mismatch at %s", strings.Join(ctx.Path, "."))
    }

    result := &node.MappingNode{
        BaseNode: base.BaseNode, // Preserve all base metadata
        Pairs:    make([]*node.MappingPair, 0),
        Style:    base.Style,
    }

    // CRITICAL: Preserve document-level comments and blank lines
    if ctx.Options.PreserveComments {
        result.HeadComment = base.HeadComment
        result.FootComment = base.FootComment
        result.LineComment = base.LineComment
    }

    if ctx.Options.PreserveBlankLines {
        result.BlankLinesBefore = base.BlankLinesBefore
        result.BlankLinesAfter = base.BlankLinesAfter
    }

    // Merge key-value pairs
    for _, basePair := range base.Pairs {
        mergedPair := s.mergePair(basePair, overrideMapping, ctx)

        // CRITICAL: Preserve pair-level formatting
        if ctx.Options.PreserveComments {
            mergedPair.KeyComment = basePair.KeyComment
            mergedPair.ValueComment = basePair.ValueComment
        }

        if ctx.Options.PreserveBlankLines {
            mergedPair.BlankLinesBefore = basePair.BlankLinesBefore
            mergedPair.BlankLinesAfter = basePair.BlankLinesAfter
        }

        result.Pairs = append(result.Pairs, mergedPair)
    }

    return result, nil
}
```

## API Design

### Simple API (80% Use Case)

```go
// Simple merge with defaults:
// - Deep merge strategy
// - Preserves comments and blank lines
// - Maintains base document's key ordering
// - Arrays are replaced (not appended)
result, err := merge.Merge(baseYAML, overrideYAML)

// Merge files with same defaults
result, err := merge.Files("base.yaml", "override.yaml")

// Merge strings with same defaults
result, err := merge.Strings(baseStr, overrideStr)
```

### Advanced API

```go
// Configure merge behavior
merger := merge.NewMerger(
    merge.WithStrategy(merge.DeepStrategy),
    merge.WithArrayStrategy(merge.AppendArrays),
    merge.WithCommentPreservation(true),
    merge.WithValidator(customValidator),
)

result, err := merger.Merge(base, override)
```

### Fluent Builder API

```go
result := merge.NewBuilder().
    Base(baseYAML).
    Override(overrideYAML).
    Strategy(merge.Deep).
    PreserveComments().
    PreserveBlanks().
    OnConflict(merge.UseOverride).
    Build().
    Execute()
```

## Use Cases

### 1. Helm Values Merging
```go
merger := merge.NewMerger(
    merge.WithStrategy(merge.DeepStrategy),
    merge.WithCommentPreservation(true),
)

result, err := merger.MergeFiles(
    "values.yaml",
    "values.production.yaml",
)
```

### 2. Configuration Layering
```go
configs := []string{
    "config/base.yaml",
    "config/env.yaml",
    "config/local.yaml",
}

result := merge.NewLayeredMerger().
    AddLayers(configs...).
    Merge()
```

### 3. Custom Merge Logic
```go
merger := merge.NewMerger(
    merge.WithCustomHandler("version", func(key string, base, override node.Node) (node.Node, bool) {
        // Always keep base version
        return base, true
    }),
)
```

## Testing Strategy

### Unit Tests
- Each strategy independently
- Each core component
- Edge cases and error conditions

### Integration Tests
- Real-world YAML scenarios
- Multi-document streams
- Complex nested structures

### Performance Tests
- Large document merging
- Deep nesting performance
- Memory usage analysis

### Test Organization
```go
// Following table-driven tests
func TestDeepMergeStrategy(t *testing.T) {
    tests := []struct {
        name     string
        base     string
        override string
        expected string
        wantErr  bool
    }{
        // Test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Error Handling

### Custom Error Types
```go
type MergeError struct {
    Path      []string
    BaseType  string
    OverrideType string
    Message   string
}

func (e *MergeError) Error() string {
    return fmt.Sprintf("merge error at %s: %s",
        strings.Join(e.Path, "."), e.Message)
}
```

### Error Wrapping
```go
if err != nil {
    return nil, fmt.Errorf("failed to merge at path %s: %w",
        path, err)
}
```

## Performance Considerations

1. **Lazy Evaluation**: Don't process nodes until necessary
2. **Caching**: Cache processed nodes to avoid recomputation
3. **Memory Pool**: Reuse node allocations
4. **Streaming**: Support for large files without full memory load

## Migration Path

### Phase 1: Core Implementation
- Basic merge functionality
- Deep merge strategy
- Unit tests

### Phase 2: Advanced Features
- Additional strategies
- Custom handlers
- Array merge options

### Phase 3: Integration
- Examples update
- Documentation
- Performance optimization

### Phase 4: Production Ready
- Comprehensive testing
- Benchmarks
- Real-world validation

## Benefits

1. **Clean Architecture**: Follows SOLID principles
2. **Testable**: Dependency injection enables easy testing
3. **Extensible**: New strategies without modification
4. **Performant**: Optimized implementation
5. **User-Friendly**: Simple API for common cases
6. **Flexible**: Advanced API for complex needs
7. **Preserves Documentation**: Comments and formatting are maintained by default
8. **Production Ready**: Handles real-world YAML with all its complexity

### Format Preservation Benefits

By preserving comments and blank lines by default, we ensure:

- **Documentation Integrity**: Schema annotations, usage instructions, and inline documentation remain intact
- **Readability**: Visual structure and logical grouping through blank lines are maintained
- **Traceability**: Comments explaining why certain values exist are preserved
- **Compliance**: Maintains any compliance or audit-related comments
- **Developer Experience**: Engineers can understand the merged configuration as easily as the original

## Conclusion

This design provides a robust, maintainable, and extensible YAML merge feature that adheres to software engineering best practices while solving real-world configuration management challenges.