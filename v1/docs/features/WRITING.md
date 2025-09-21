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

### 05. Commented Case with breaking lines configuration enabled

The sample below we will enhance the parsing.
We will have comments but now empty lines. So, we need to add the blank lines because we have set the configuration to enable the empty lines.
Parser need to identify the comments "# "

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
```

