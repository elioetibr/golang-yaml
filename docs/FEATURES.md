# Feature Overview

## Core Features

### 1. YAML 1.2.2 Specification Compliance

- **~90% compliance** with the official YAML 1.2.2 specification
- Proper handling of all scalar styles (plain, quoted, literal, folded)
- Block and flow collection support
- Document markers and directives
- Tab rejection in indentation (spec section 6.1)

### 2. Advanced YAML Features

#### Anchors & Aliases
- Full anchor definition and reference resolution
- Deep cloning of aliased nodes
- Circular reference detection
- Anchor registry management

#### Merge Keys (<<)
- Complete merge key support
- Multiple merge key handling
- Override capability
- Nested merge resolution

#### Tags
- YAML 1.2 default tag resolution
- Custom tag support
- Tag shorthand handling
- Explicit type tags (!!, !)

#### Multiple Documents
- Stream parsing for multi-document files
- Document directives (---, ...)
- Independent document processing
- Stream serialization

### 3. Parsing Capabilities

- **Recursive descent parser** with error recovery
- **Comprehensive error messages** with line/column info
- **Position tracking** for all tokens
- **Streaming support** for large files
- **Panic-resistant** with proper error handling

### 4. Serialization Features

- **Round-trip preservation** of formatting
- **Style preservation** (block vs flow)
- **Configurable indentation**
- **Line width control**
- **Smart quoting** for special values

### 5. Comment System

#### Comment Types
- **Head comments**: Before nodes
- **Line comments**: Inline with nodes
- **Foot comments**: After nodes
- **Key comments**: For mapping keys
- **Value comments**: For mapping values

#### Comment Features
- Full comment preservation
- Comment association with nodes
- Blank line tracking
- Configurable spacing

### 6. Transformation Capabilities

#### Sorting System

**Sort Modes:**
- KeepOriginal (default)
- Ascending
- Descending

**Sort Strategies:**
- By key (for mappings)
- By value (for sequences)
- Priority-based sorting
- Group-based sorting
- Custom comparators

**Sort Features:**
- Recursive sorting
- Case sensitivity control
- Path-based exclusions
- Stable sort algorithm
- Comment preservation during sort

#### Formatting System

**Format Options:**
- Configurable blank lines
- Smart blank line detection
- Section markers
- Comment spacing
- Force formatting

**Preset Configurations:**
- Standard
- Cleanup
- Minimal
- Readable

### 7. High-Level API

#### Marshal/Unmarshal
- Similar to encoding/json
- Struct tag support
- Custom type handling
- Omitempty support
- Inline fields

#### Encoder/Decoder
- Stream-based processing
- io.Reader/Writer support
- Multiple encode/decode calls
- Error accumulation

### 8. Node Building

- Programmatic node construction
- Builder pattern implementation
- Fluent API for complex structures
- Anchor/tag attachment
- Style control

### 9. Performance Features

- **Optimized tokenizer** with minimal allocations
- **Efficient parser** with single-pass design
- **Memory-conscious** serialization
- **Benchmark suite** for regression testing
- **Parallel processing** support

### 10. Safety Features

- **No code execution** (safe by default)
- **Resource limits** (configurable)
- **Input validation**
- **Error recovery**
- **Panic protection** in fuzzing

## Unique Features (Beyond Spec)

### 1. Advanced Sorting
No other YAML library provides such comprehensive sorting capabilities with priority lists, groups, and custom comparators.

### 2. Comment Association
Enhanced comment handling that goes beyond the specification to maintain comments with their associated nodes during transformations.

### 3. Blank Line Management
Intelligent blank line tracking and preservation with configurable rules and matchers.

### 4. Formatting Engine
Complete formatting system with presets and fine-grained control over spacing and layout.

### 5. Transform Pipeline
Ability to chain transformations (sort → format → serialize) in a single operation.

## Use Cases

### Configuration Management
- Application configs
- Kubernetes manifests
- CI/CD pipelines
- Docker Compose files

### Data Processing
- Data transformation
- Config migration
- Schema validation
- Format conversion

### Development Tools
- Code generation
- Config linting
- Format standardization
- Documentation generation

### DevOps
- Helm charts
- Ansible playbooks
- Cloud formations
- Infrastructure as Code

## Comparison with Other Libraries

| Feature | This Library | gopkg.in/yaml.v3 | Other Libraries |
|---------|------------|------------------|-----------------|
| YAML 1.2.2 Compliance | ~90% | ~70% | Varies |
| Anchor/Alias Resolution | ✅ Full | ✅ Full | ⚠️ Partial |
| Merge Keys | ✅ Full | ✅ Full | ❌ Often missing |
| Comment Preservation | ✅ Enhanced | ✅ Basic | ❌ Usually none |
| Sorting System | ✅ Advanced | ❌ None | ❌ None |
| Formatting Engine | ✅ Full | ❌ None | ❌ None |
| Tag Resolution | ✅ Full | ✅ Full | ⚠️ Partial |
| Multiple Documents | ✅ Full | ✅ Full | ⚠️ Partial |
| Performance | ⚡ Fast | ⚡ Fast | Varies |
| Memory Usage | 🔋 Efficient | 🔋 Efficient | Varies |

## Performance Metrics

Based on benchmarks on Intel Xeon W-2150B @ 3.00GHz:

### Parsing Performance
- **Small documents**: 3.9 microseconds
- **Medium documents**: 14.6 microseconds
- **Large documents**: 1.25 milliseconds

### Serialization Performance
- **Small documents**: 5.0 microseconds
- **Medium documents**: 27.1 microseconds
- **Large documents**: 2.18 milliseconds

### Memory Efficiency
- **Small documents**: 17 allocations
- **Medium documents**: 173 allocations
- **Large documents**: 14,128 allocations

## Limitations

### Current Limitations
1. UTF-8 encoding only (no UTF-16/32)
2. No chomping indicators (|--, |+, etc.)
3. Partial indentation validation
4. No implicit document handling
5. No node comparison/equality methods

### Design Decisions
1. Comments always move with nodes (not configurable)
2. Default sort mode is KeepOriginal
3. Tabs rejected in block context
4. Deep cloning for alias resolution

## Future Enhancements

See [ROADMAP.md](ROADMAP.md) for planned features:
- UTF-16 encoding support
- Chomping/indentation indicators
- Full indentation enforcement
- Node comparison methods
- Streaming optimization