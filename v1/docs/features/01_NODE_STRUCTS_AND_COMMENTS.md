# Comments and Empty Lines

I want to develop and improve the comments and empty lines including the better block separations.

---

## Architecture Proposal

### Design Principles

- **SOLID** - Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **KISS** - Keep It Simple, Stupid
- **DRY** - Don't Repeat Yourself
- **Dependency Injection** - For testability and flexibility
- **Must Pattern** - For better error handling
- **Layered Architecture** - Clear separation of concerns

---

## Proposal

Restructure the structs to have a Better Node detection.

---

## TDD Cases

### 01. Simple Case Empty YAML file or string

Let's create a TDD to loads an empty yaml file | string.

### 02. Only Comments Yaml Case

Let's create a TDD to loads commented lines only.

#### Test

Arrange:

1. The node needs to be enough smart to load the string or document and setup the write structure
2. In this case I just want to store the comments in a commented map group struct since we don't have any key.
3. We need to be able to track in each comment line the { line id, the comment, next line / token is a comment|emptyLine|yaml structure}

Act:

Load the yaml's, without error keeping the structure untouched.

Assert:

We should generate a yaml and have the string below.

Assertion 1:

```yaml
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.
```

Assertion 2:

```yaml
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.
```

### 03. Simple Case

If do we have a yaml with without comments like below, we will treat as common yaml and load nothing to do.

#### Base yaml

```yaml
company: Umbrella Corp.
city: Raccoon City
employees:
  bob@umbreallacorp.co:
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co:
    name: Alice Abernathy
    department: Project
```

#### To Merge yaml

```yaml
company: Umbrella Corporation.
city: Raccoon City
employees:
  redqueen@umbreallacorp.co:
    name: Red Queen
    department: Security
```

#### Expected merged yaml result

```yaml
company: Umbrella Corporation.
city: Raccoon City
employees:
  bob@umbreallacorp.co:
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co:
    name: Alice Abernathy
    department: Project
  redqueen@umbreallacorp.co:
    name: Red Queen
    department: Security
```

### 04. Commented Case without breaking lines

The sample below we will enhance the parsing.
We will have comments but now empty lines. So, we need to keep as is, except I set the configuration to enable the empty lines.

#### Base yaml

```yaml
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# Company Name
# -- This is a simple Company Name
company: Umbrella Corp.
# City Name
# -- This is the Company City Name
city: Raccoon City
# Company Employees
# -- This is a mapping of Company Employees
employees:
  bob@umbreallacorp.co: # Bob Umbrella's Email
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co: # Alice Umbrella's Email
    name: Alice Abernathy
    department: Project
```

#### To Merge yaml

```yaml
company: Umbrella Corporation.
city: Raccoon City
employees:
  redqueen@umbreallacorp.co: # Red Queen Umbrella's Email
    name: Red Queen
    department: Security
```

#### Expected merged yaml result

```yaml
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# Company Name
# -- This is a simple Company Name
company: Umbrella Corporation.
# City Name
# -- This is the Company City Name
city: Raccoon City
# Company Employees
# -- This is a mapping of Company Employees
employees:
  bob@umbreallacorp.co: # Bob Umbrella's Email
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co: # Alice Umbrella's Email
    name: Alice Abernathy
    department: Project
  redqueen@umbreallacorp.co: # Red Queen Umbrella's Email
    name: Red Queen
    department: Security
```

### 05. Commented Case with breaking lines configuration enabled separating the sections by default

The sample below we will enhance the parsing.
We will have comments but now empty lines. So, we need to add the blank lines because we have set the configuration to enable the empty lines.
Parser need to identify the comments "# "
Needs to check if key has HeadComments or InLineComments and FootComments (if next token is comment check the next, till the next token be an empty line or key or scalar definition, treat as comment mapped group with key as the parent token)
If we found empty line as token initialize a next block section

#### Implementation Findings

Based on the current codebase analysis, we have:

**✅ Already Implemented:**
- Node structures support `BlankLinesBefore` and `BlankLinesAfter` fields
- Comment tracking with `HeadComment`, `LineComment`, and `FootComment`
- MappingPairs track blank lines and comments
- Options for `PreserveComments` and `PreserveBlankLines`
- `NodeProcessor.PreserveMetadata()` handles preservation logic

**🔧 Needs Implementation:**
1. **Comment Association Logic**: Comments need to be properly tied to their keys during parsing
2. **Section Detection**: Detect root and child sections based on blank lines
3. **Configurable Blank Lines**: Add configuration options:
   - `KeepDefaultLineBetweenSections` (bool) - preserve original blank line count (default: true)
   - `DefaultLineBetweenSections` (int) - number of blank lines (default: 1)
4. **Smart Section Handling**: If more than 1 blank line exists, keep as-is unless overridden

#### Proposed Configuration Structure

```go
type Options struct {
    // Existing options
    Strategy           Strategy
    PreserveComments   bool
    PreserveBlankLines bool

    // New section handling options
    KeepDefaultLineBetweenSections bool // Keep original blank line count
    DefaultLineBetweenSections     int  // Default blank lines between sections (default: 1)
}
```

#### Comment Association Algorithm

1. **Document Head Comments**: Comments at the beginning of the document before any keys
   - Stored in a special `HeadCommentDocumentSections` field
   - First scalar key must be aware these exist
2. **Pre-key Comments**: Comments before a key become HeadComment
3. **Inline Comments**: Comments on same line as key become LineComment
4. **Post-value Comments**: Comments after value until blank line or next key become FootComment
5. **Section Separation**: Blank lines indicate section boundaries
   - If more than 1 blank line exists, keep as-is (unless overridden)
   - Configuration controls normalization behavior

#### Implementation Status

**✅ Completed:**
- Added `KeepDefaultLineBetweenSections` option (default: true)
- Added `DefaultLineBetweenSections` option (default: 1)
- Enhanced `PreserveMetadata()` to handle section normalization
- Added `IsSectionBoundary()` method to detect sections
- Added `NormalizeSectionBoundaries()` method for consistent formatting

**✅ Now Complete:**
- `HeadCommentDocumentSections` field added to DocumentNode
- `HasDocumentHeadComments` flag added to MappingNode
- `PreserveDocumentHeadComments()` method for handling document-level comments
- Section boundary detection with configurable thresholds
- Comprehensive test coverage for all features

#### Key Features Implemented

1. **Section Handling Configuration:**
   ```go
   opts := &Options{
       KeepDefaultLineBetweenSections: true,  // Keep original blank line count
       DefaultLineBetweenSections:     1,     // Or normalize to N lines
   }
   ```

2. **Document Head Comments:**
   - Comments at document start are tracked separately
   - First key is aware of document-level comments
   - Preserved during merge operations

3. **Intelligent Section Detection:**
   - Detects sections based on blank line count (default: 2+ lines = section)
   - Preserves original formatting or normalizes as configured
   - Handles nested structures recursively

#### Base yaml

```yaml
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# Company Name
# -- This is a simple Company Name
company: Umbrella Corp.

# City Name
# -- This is the Company City Name
city: Raccoon City

# Company Employees
# -- This is a mapping of Company Employees
employees:
  bob@umbreallacorp.co: # Bob Umbrella's Email
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co: # Alice Umbrella's Email
    name: Alice Abernathy
    department: Project
```

#### To Merge yaml:

```yaml
company: Umbrella Corporation.
city: Raccoon City
employees:
  redqueen@umbreallacorp.co: # Red Queen Umbrella's Email
    name: Red Queen
    department: Security

# @schema
# additionalProperties: true
# @schema
# -- This status list demonstrates the possible hive access
statusList: []
# Examples:
# statusList:
#   - enabled
#   - blocked
#   - quarantine
#   - infected
```

#### Expected merged yaml result:

```yaml
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# Company Name
# -- This is a simple Company Name
company: Umbrella Corporation.

# City Name
# -- This is the Company City Name
city: Raccoon City

# Company Employees
# -- This is a mapping of Company Employees
employees:
  bob@umbreallacorp.co: # Bob Umbrella's Email
    name: Bob Sinclair
    department: Cloud Computing
  alice@umbreallacorp.co: # Alice Umbrella's Email
    name: Alice Abernathy
    department: Project
  redqueen@umbreallacorp.co: # Red Queen Umbrella's Email
    name: Red Queen
    department: Security

# @schema
# additionalProperties: true
# @schema
# -- This status list demonstrates the possible hive access
statusList: []
# Examples:
# statusList:
#   - enabled
#   - blocked
#   - quarantine
#   - infected
```
