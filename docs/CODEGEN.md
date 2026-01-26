# Code Generation Tool

The `polarion-codegen` CLI tool automatically generates type-safe Go structs from your Polarion configuration, eliminating the need to manually define custom work item types.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [CLI Usage](#cli-usage)
  - [Programmatic Usage](#programmatic-usage)
- [Generated Code](#generated-code)
- [Refresh Mode](#refresh-mode)
- [Examples](#examples)

## Installation

```bash
go install github.com/almnorth/go-polarion/cmd/polarion-codegen@latest
```

## Quick Start

Generate type-safe structs for a specific work item type:

```bash
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject \
  --type requirement
```

Generate for all work item types in the project:

```bash
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject
```

## Usage

### CLI Usage

The tool supports the following command-line options:

```bash
polarion-codegen [options]

Options:
  --url string       Polarion REST API base URL (required)
  --token string     Bearer token for authentication (required)
  --project string   Project ID (required)
  --type string      Work item type ID (optional, generates all types if not specified)
  --output string    Output directory (default: "./generated")
  --package string   Package name for generated code (default: "generated")
  --refresh          Refresh existing generated files (preserves custom code)
```

**Examples:**

```bash
# Generate a specific type
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject \
  --type requirement \
  --output ./models \
  --package models

# Generate all types
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject

# Refresh existing generated files
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject \
  --refresh
```

### Programmatic Usage

You can also use the codegen package directly in your Go code:

```go
package main

import (
    "context"
    "log"
    
    polarion "github.com/almnorth/go-polarion"
    "github.com/almnorth/go-polarion/codegen"
)

func main() {
    // Create Polarion client
    client, err := polarion.New(
        "https://polarion.example.com/rest/v1",
        "YOUR_TOKEN",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Configure code generation
    config := &codegen.Config{
        OutputDir: "./generated",
        Package:   "generated",
        TypeID:    "", // Empty for all types, or specify like "requirement"
        Refresh:   false,
    }
    
    // Create generator
    gen := codegen.NewGenerator(client, "myproject", config)
    
    // Generate code
    ctx := context.Background()
    if err := gen.Generate(ctx); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Code generation completed successfully")
}
```

## Generated Code

The tool generates the following for each work item type:

### 1. Type-Safe Struct

```go
// Requirement represents a requirement work item with custom fields
type Requirement struct {
    base             *polarion.WorkItem
    BusinessValue    *string            // Custom enum field
    TargetRelease    *polarion.DateOnly // Custom date field
    ComplexityPoints *float64           // Custom float field
    SecurityReviewed *bool              // Custom boolean field
}
```

### 2. Constructor

```go
// NewRequirement creates a new Requirement instance
func NewRequirement() *Requirement {
    return &Requirement{
        base: &polarion.WorkItem{
            Type: "workitems",
            Attributes: &polarion.WorkItemAttributes{
                Type: "requirement",
            },
        },
    }
}
```

### 3. Load/Save Methods

```go
// LoadFromWorkItem loads data from a WorkItem
func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
    r.base = wi
    cf := polarion.CustomFields(wi.Attributes.CustomFields)
    
    if val, ok := cf.GetString("businessValue"); ok {
        r.BusinessValue = &val
    }
    if val, ok := cf.GetDateOnly("targetRelease"); ok {
        r.TargetRelease = &val
    }
    // ... more fields
    
    return nil
}

// SaveToWorkItem saves data to the underlying WorkItem
func (r *Requirement) SaveToWorkItem() {
    cf := polarion.CustomFields(r.base.Attributes.CustomFields)
    
    if r.BusinessValue != nil {
        cf.Set("businessValue", *r.BusinessValue)
    }
    if r.TargetRelease != nil {
        cf.Set("targetRelease", r.TargetRelease.String())
    }
    // ... more fields
}
```

### 4. Getter/Setter Methods

```go
// GetBusinessValue returns the business value
func (r *Requirement) GetBusinessValue() string {
    if r.BusinessValue != nil {
        return *r.BusinessValue
    }
    return ""
}

// SetBusinessValue sets the business value
func (r *Requirement) SetBusinessValue(value string) {
    r.BusinessValue = &value
}

// WorkItem returns the underlying WorkItem
func (r *Requirement) WorkItem() *polarion.WorkItem {
    return r.base
}
```

### 5. Generation Markers

The tool adds special markers to preserve custom code during refresh:

```go
// Code generated by polarion-codegen. DO NOT EDIT.
// To add custom methods, use the CUSTOM CODE section below.

// ... generated code ...

// BEGIN CUSTOM CODE
// Add your custom methods here. This section will be preserved during refresh.

// END CUSTOM CODE
```

## Refresh Mode

When using `--refresh`, the tool:

1. ✅ Updates generated structs with new/changed fields
2. ✅ Preserves code between `BEGIN CUSTOM CODE` and `END CUSTOM CODE` markers
3. ✅ Updates Load/Save methods
4. ✅ Adds new getter/setter methods
5. ⚠️ Removes getter/setters for deleted fields

**Example workflow:**

```bash
# Initial generation
polarion-codegen --url ... --token ... --project myproject

# Add custom validation method
# Edit generated/requirement.go and add between markers:
# BEGIN CUSTOM CODE
func (r *Requirement) Validate() error {
    if r.BusinessValue == nil {
        return errors.New("business value is required")
    }
    return nil
}
# END CUSTOM CODE

# Refresh after Polarion config changes
polarion-codegen --url ... --token ... --project myproject --refresh
# Your Validate() method is preserved!
```

## Examples

### Complete Example

See [`examples/codegen/main.go`](../examples/codegen/main.go) for a complete, runnable example.

### Using Generated Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    polarion "github.com/almnorth/go-polarion"
    "yourproject/generated"
)

func main() {
    client, _ := polarion.New("https://polarion.example.com/rest/v1", "TOKEN")
    project := client.Project("myproject")
    ctx := context.Background()
    
    // Create a new requirement
    req := generated.NewRequirement()
    req.SetBusinessValue("high")
    req.SetComplexityPoints(8.5)
    req.WorkItem().Attributes.Title = "New Security Feature"
    req.WorkItem().Attributes.Status = "draft"
    
    // Save to Polarion
    req.SaveToWorkItem()
    if err := project.WorkItems.Create(ctx, req.WorkItem()); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created requirement: %s\n", req.WorkItem().ID)
    
    // Load existing requirement
    wi, _ := project.WorkItems.Get(ctx, "REQ-123")
    req2 := generated.NewRequirement()
    req2.LoadFromWorkItem(wi)
    
    // Type-safe access
    fmt.Printf("Business Value: %s\n", req2.GetBusinessValue())
    if req2.TargetRelease != nil {
        fmt.Printf("Target Release: %s\n", req2.TargetRelease.String())
    }
}
```

## Field Type Mapping

The tool automatically maps Polarion field types to Go types:

| Polarion Kind | Go Type | Notes |
|---------------|---------|-------|
| `string` | `*string` | Text fields |
| `enumeration` | `*string` | Enum values |
| `integer` | `*int` | Integer numbers |
| `float` | `*float64` | Floating-point numbers |
| `boolean` | `*bool` | Boolean values |
| `time` | `*polarion.TimeOnly` | Time without date |
| `date` | `*polarion.DateOnly` | Date without time |
| `date-time` | `*polarion.DateTime` | Date and time |
| `duration` | `*polarion.Duration` | Time duration |
| `text` | `*polarion.TextContent` | Rich text content |
| `text/html` | `*polarion.TextContent` | HTML content |

All fields are pointers to support nil values (missing/unset fields).

## Best Practices

1. **Version Control**: Commit generated files to track changes over time
2. **Custom Code**: Use the CUSTOM CODE section for validation, business logic, etc.
3. **Refresh Regularly**: Run with `--refresh` when Polarion configuration changes
4. **Separate Package**: Generate into a dedicated package (e.g., `models` or `generated`)
5. **Documentation**: Add comments in the CUSTOM CODE section to document your additions

## Troubleshooting

### Authentication Errors

Ensure your bearer token has sufficient permissions to access:
- Work item type definitions
- Custom field configurations
- Field metadata

### Missing Fields

If fields are missing from generated code:
1. Verify the field exists in Polarion for that work item type
2. Check that the field is properly configured in custom fields
3. Ensure your Polarion version supports the Fields Metadata API (>= 2512)

### Refresh Not Preserving Code

Ensure your custom code is between the markers:
```go
// BEGIN CUSTOM CODE
// Your code here
// END CUSTOM CODE
```

Code outside these markers will be overwritten during refresh.

## See Also

- [Custom Work Items Documentation](CUSTOM-WORKITEMS.md) - Manual approach to custom work items
- [API Reference](API-REFERENCE.md) - Complete API documentation
- [Examples](../examples/) - Working code examples
