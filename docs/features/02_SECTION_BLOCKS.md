# Comments and Empty Lines

I want to develop and improve the comments and empty lines including the better block separations.

## Proposal

Restructure the structs to have a Better Node detection.

```go
// CommentGroup represents a collection of comments
type CommentGroup struct {
	Comments         []string
	BlankLinesBefore int // Number of blank lines before this comment group
}

// BaseNode contains common fields for all node types
type BaseNode struct {
	TagValue     string
	AnchorValue  string
	LineNumber   int
	ColumnNumber int

	// Comment associations
	HeadComment *CommentGroup // Comments before the node
	LineComment *CommentGroup // Inline comment on same line
	FootComment *CommentGroup // Comments after the node

	// Blank line tracking
	BlankLinesBefore int // Number of blank lines before this node
	BlankLinesAfter  int // Number of blank lines after this node

	StyleHint Style
}
```

## Cases

### Simple Case

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

#### To Merge yaml:

```yaml
company: Umbrella Corporation.
city: Raccoon City
employees:
  redqueen@umbreallacorp.co:
    name: Red Queen
    department: Security
```

#### Expected merged yaml result:

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

### Commented Case without breaking lines

The sample below we will enhance the parsing.
We will have comments but now empty lines. So, we need to keep as is, except I set the configuration to enable the empty lines.

#### Base yaml

```yaml
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
```

#### Expected merged yaml result:

```yaml
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

### Commented Case with breaking lines configuration enabled

The sample below we will enhance the parsing.
We will have comments but now empty lines. So, we need to add the blank lines because we have set the configuration to enable the empty lines.
Parser need to identify the comments "# "

#### Base yaml

```yaml
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
```

#### Expected merged yaml result:

```yaml
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

