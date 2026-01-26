# Type-Safe Custom Work Items

This guide explains how to define and use custom work item types with type-safe custom fields in go-polarion.

## Table of Contents

- [Introduction](#introduction)
- [Why Use Type-Safe Custom Fields?](#why-use-type-safe-custom-fields)
- [Quick Start](#quick-start)
- [Defining Custom Types](#defining-custom-types)
- [Working with Custom Fields](#working-with-custom-fields)
- [Field Type Reference](#field-type-reference)
- [Advanced Topics](#advanced-topics)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

Polarion allows you to define custom fields on work items to capture project-specific information. While you can access these fields using the generic `CustomFields` map, defining custom types provides several advantages:

- **Type Safety**: Compile-time checking prevents type errors
- **Better IDE Support**: Autocomplete and inline documentation
- **Validation**: Easy to add custom validation logic
- **Maintainability**: Clear structure and documentation
- **Refactoring**: Easier to update when field names change

## Why Use Type-Safe Custom Fields?

### Without Type Safety

```go
// Generic approach - prone to errors
wi, _ := project.WorkItems.Get(ctx, "REQ-123")
priority := wi.Attributes.CustomFields["priority"] // interface{}
if priority != nil {
    // Need type assertion, might panic
    p := priority.(string)
    fmt.Println(p)
}

// Easy to make typos
value := wi.Attributes.CustomFields["priorty"] // Typo! No compile error
```

### With Type Safety

```go
// Type-safe approach
req := &Requirement{}
wi, _ := project.WorkItems.Get(ctx, "REQ-123")
req.LoadFromWorkItem(wi)

// Compile-time checking, autocomplete works
if req.BusinessValue != nil {
    fmt.Printf("Business Value: %s\n", *req.BusinessValue) // Type-safe!
}
```

## Quick Start

Here's a minimal example to get you started:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/almnorth/go-polarion"
)

// Define your custom type
type Requirement struct {
    base          *polarion.WorkItem
    BusinessValue *string
}

// Load from WorkItem
func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
    r.base = wi
    cf := polarion.CustomFields(wi.Attributes.CustomFields)
    
    if val, ok := cf.GetString("businessValue"); ok {
        r.BusinessValue = &val
    }
    
    return nil
}

// Save to WorkItem
func (r *Requirement) SaveToWorkItem() {
    cf := polarion.CustomFields(r.base.Attributes.CustomFields)
    
    if r.BusinessValue != nil {
        cf.Set("businessValue", *r.BusinessValue)
    } else {
        cf.Delete("businessValue")
    }
}

func main() {
    client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
    project := client.Project("myproject")
    ctx := context.Background()
    
    // Load and use
    req := &Requirement{}
    wi, _ := project.WorkItems.Get(ctx, "REQ-123")
    req.LoadFromWorkItem(wi)
    
    if req.BusinessValue != nil {
        fmt.Printf("Business Value: %s\n", *req.BusinessValue)
    }
}
```

## Defining Custom Types

### Basic Structure

A custom work item type typically has:

1. A reference to the base `WorkItem`
2. Typed fields for custom fields
3. Methods to load from and save to `WorkItem`
4. Helper methods for accessing fields

```go
type Requirement struct {
    // Base WorkItem - required for API operations
    base *polarion.WorkItem
    
    // Custom fields with appropriate types
    BusinessValue    *string            // Optional enum field
    TargetRelease    *polarion.DateOnly // Optional date field
    ComplexityPoints *float64           // Optional float field
    SecurityReviewed *bool              // Optional boolean field
}
```

### Constructor Pattern

Provide a constructor for creating new instances:

```go
func NewRequirement(title string) *Requirement {
    return &Requirement{
        base: &polarion.WorkItem{
            Type: "workitems",
            Attributes: &polarion.WorkItemAttributes{
                Title:        title,
                CustomFields: make(map[string]interface{}),
            },
        },
    }
}
```

### Getter Methods

Provide safe getter methods that handle nil values:

```go
func (r *Requirement) GetBusinessValue() string {
    if r.BusinessValue != nil {
        return *r.BusinessValue
    }
    return "" // or a sensible default
}

func (r *Requirement) GetComplexityPoints() float64 {
    if r.ComplexityPoints != nil {
        return *r.ComplexityPoints
    }
    return 0.0
}
```

### Setter Methods

Provide setter methods that sync with the base WorkItem:

```go
func (r *Requirement) SetBusinessValue(value string) {
    r.BusinessValue = &value
    if r.base != nil && r.base.Attributes != nil {
        if r.base.Attributes.CustomFields == nil {
            r.base.Attributes.CustomFields = make(map[string]interface{})
        }
        r.base.Attributes.CustomFields["businessValue"] = value
    }
}
```

## Working with Custom Fields

### Loading from WorkItem

The `LoadFromWorkItem` method extracts custom fields from a WorkItem:

```go
func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
    r.base = wi
    
    if wi.Attributes == nil {
        return fmt.Errorf("work item attributes are nil")
    }
    
    // Use CustomFields wrapper for type-safe access
    cf := polarion.CustomFields(wi.Attributes.CustomFields)
    
    // Load each field with appropriate type
    if val, ok := cf.GetString("businessValue"); ok {
        r.BusinessValue = &val
    }
    
    if val, ok := cf.GetDateOnly("targetRelease"); ok {
        r.TargetRelease = &val
    }
    
    if val, ok := cf.GetFloat("complexityPoints"); ok {
        r.ComplexityPoints = &val
    }
    
    if val, ok := cf.GetBool("securityReviewed"); ok {
        r.SecurityReviewed = &val
    }
    
    return nil
}
```

### Saving to WorkItem

The `SaveToWorkItem` method syncs custom fields back to the WorkItem:

```go
func (r *Requirement) SaveToWorkItem() {
    if r.base == nil || r.base.Attributes == nil {
        return
    }
    
    if r.base.Attributes.CustomFields == nil {
        r.base.Attributes.CustomFields = make(map[string]interface{})
    }
    
    cf := polarion.CustomFields(r.base.Attributes.CustomFields)
    
    // Save or delete each field
    if r.BusinessValue != nil {
        cf.Set("businessValue", *r.BusinessValue)
    } else {
        cf.Delete("businessValue")
    }
    
    if r.TargetRelease != nil {
        cf.Set("targetRelease", r.TargetRelease.String())
    } else {
        cf.Delete("targetRelease")
    }
    
    if r.ComplexityPoints != nil {
        cf.Set("complexityPoints", *r.ComplexityPoints)
    } else {
        cf.Delete("complexityPoints")
    }
    
    if r.SecurityReviewed != nil {
        cf.Set("securityReviewed", *r.SecurityReviewed)
    } else {
        cf.Delete("securityReviewed")
    }
}
```

### Handling Optional Fields

Use pointers for optional fields to distinguish between "not set" and "zero value":

```go
// With pointer - can distinguish nil from false
SecurityReviewed *bool

// Check if set
if req.SecurityReviewed != nil {
    if *req.SecurityReviewed {
        fmt.Println("Security reviewed and approved")
    } else {
        fmt.Println("Security reviewed but not approved")
    }
} else {
    fmt.Println("Security review not performed")
}
```

### Validation

Add validation methods to ensure data integrity:

```go
func (r *Requirement) Validate() error {
    // Check required fields
    if r.base == nil || r.base.Attributes == nil {
        return fmt.Errorf("base work item not initialized")
    }
    
    if r.base.Attributes.Title == "" {
        return fmt.Errorf("title is required")
    }
    
    // Validate business value
    if r.BusinessValue != nil {
        validValues := []string{"low", "medium", "high", "critical"}
        valid := false
        for _, v := range validValues {
            if *r.BusinessValue == v {
                valid = true
                break
            }
        }
        if !valid {
            return fmt.Errorf("invalid business value: %s", *r.BusinessValue)
        }
    }
    
    // Validate complexity points
    if r.ComplexityPoints != nil && *r.ComplexityPoints < 0 {
        return fmt.Errorf("complexity points cannot be negative")
    }
    
    // Validate target release is in the future
    if r.TargetRelease != nil && r.TargetRelease.Before(time.Now()) {
        return fmt.Errorf("target release must be in the future")
    }
    
    return nil
}
```

## Field Type Reference

### String Fields

Used for text and enumeration fields.

```go
// Definition
BusinessValue *string

// Loading
if val, ok := cf.GetString("businessValue"); ok {
    r.BusinessValue = &val
}

// Saving
if r.BusinessValue != nil {
    cf.Set("businessValue", *r.BusinessValue)
}

// Usage
if req.BusinessValue != nil {
    fmt.Printf("Value: %s\n", *req.BusinessValue)
}
```

### Integer Fields

Used for whole numbers.

```go
// Definition
ItemCount *int

// Loading
if val, ok := cf.GetInt("itemCount"); ok {
    r.ItemCount = &val
}

// Saving
if r.ItemCount != nil {
    cf.Set("itemCount", *r.ItemCount)
}
```

### Float Fields

Used for decimal numbers.

```go
// Definition
ComplexityPoints *float64

// Loading
if val, ok := cf.GetFloat("complexityPoints"); ok {
    r.ComplexityPoints = &val
}

// Saving
if r.ComplexityPoints != nil {
    cf.Set("complexityPoints", *r.ComplexityPoints)
}
```

### Boolean Fields

Used for true/false values.

```go
// Definition
SecurityReviewed *bool

// Loading
if val, ok := cf.GetBool("securityReviewed"); ok {
    r.SecurityReviewed = &val
}

// Saving
if r.SecurityReviewed != nil {
    cf.Set("securityReviewed", *r.SecurityReviewed)
}
```

### Date Fields

Used for dates without time information.

```go
// Definition
TargetRelease *polarion.DateOnly

// Loading
if val, ok := cf.GetDateOnly("targetRelease"); ok {
    r.TargetRelease = &val
}

// Saving
if r.TargetRelease != nil {
    cf.Set("targetRelease", r.TargetRelease.String())
}

// Creating
date := polarion.NewDateOnly(time.Now().AddDate(0, 3, 0))
req.TargetRelease = &date

// Parsing
date, err := polarion.ParseDateOnly("2026-06-15")
if err == nil {
    req.TargetRelease = &date
}
```

### Time Fields

Used for time without date information.

```go
// Definition
StartTime *polarion.TimeOnly

// Loading
if val, ok := cf.GetTimeOnly("startTime"); ok {
    r.StartTime = &val
}

// Saving
if r.StartTime != nil {
    cf.Set("startTime", r.StartTime.String())
}

// Creating
time, _ := polarion.NewTimeOnly(14, 30, 0) // 14:30:00
req.StartTime = &time

// Parsing
time, err := polarion.ParseTimeOnly("14:30:00")
```

### DateTime Fields

Used for date and time information.

```go
// Definition
ReviewedAt *polarion.DateTime

// Loading
if val, ok := cf.GetDateTime("reviewedAt"); ok {
    r.ReviewedAt = &val
}

// Saving
if r.ReviewedAt != nil {
    cf.Set("reviewedAt", r.ReviewedAt.String())
}

// Creating
dt := polarion.NewDateTime(time.Now())
req.ReviewedAt = &dt
```

### Duration Fields

Used for time durations.

```go
// Definition
EstimatedEffort *polarion.Duration

// Loading
if val, ok := cf.GetDuration("estimatedEffort"); ok {
    r.EstimatedEffort = &val
}

// Saving
if r.EstimatedEffort != nil {
    cf.Set("estimatedEffort", r.EstimatedEffort.String())
}

// Creating
duration := polarion.NewDuration(8 * time.Hour) // 8 hours
req.EstimatedEffort = &duration

// Parsing
duration, err := polarion.ParseDuration("2d 3h 30m")
```

### Text Content Fields

Used for rich text (HTML) or plain text.

```go
// Definition
DetailedDescription *polarion.TextContent

// Loading
if val, ok := cf.GetText("detailedDescription"); ok {
    r.DetailedDescription = val
}

// Saving
if r.DetailedDescription != nil {
    cf.Set("detailedDescription", r.DetailedDescription)
}

// Creating HTML content
html := polarion.NewHTMLContent("<p>Detailed description</p>")
req.DetailedDescription = html

// Creating plain text
text := polarion.NewPlainTextContent("Simple description")
req.DetailedDescription = text
```

## Advanced Topics

### Builder Pattern

For complex types, consider using a builder pattern:

```go
type RequirementBuilder struct {
    req *Requirement
}

func NewRequirementBuilder(title string) *RequirementBuilder {
    return &RequirementBuilder{
        req: NewRequirement(title),
    }
}

func (b *RequirementBuilder) WithBusinessValue(value string) *RequirementBuilder {
    b.req.SetBusinessValue(value)
    return b
}

func (b *RequirementBuilder) WithTargetRelease(date polarion.DateOnly) *RequirementBuilder {
    b.req.SetTargetRelease(date)
    return b
}

func (b *RequirementBuilder) WithComplexityPoints(points float64) *RequirementBuilder {
    b.req.SetComplexityPoints(points)
    return b
}

func (b *RequirementBuilder) Build() (*Requirement, error) {
    if err := b.req.Validate(); err != nil {
        return nil, err
    }
    return b.req, nil
}

// Usage
req, err := NewRequirementBuilder("New Feature").
    WithBusinessValue("high").
    WithComplexityPoints(13.0).
    WithTargetRelease(polarion.NewDateOnly(time.Now().AddDate(0, 3, 0))).
    Build()
```

### Inheritance and Composition

For shared fields across multiple types, use composition:

```go
// Common fields
type BaseCustomFields struct {
    BusinessValue    *string
    TargetRelease    *polarion.DateOnly
}

func (b *BaseCustomFields) LoadBase(cf polarion.CustomFields) {
    if val, ok := cf.GetString("businessValue"); ok {
        b.BusinessValue = &val
    }
    if val, ok := cf.GetDateOnly("targetRelease"); ok {
        b.TargetRelease = &val
    }
}

func (b *BaseCustomFields) SaveBase(cf polarion.CustomFields) {
    if b.BusinessValue != nil {
        cf.Set("businessValue", *b.BusinessValue)
    }
    if b.TargetRelease != nil {
        cf.Set("targetRelease", b.TargetRelease.String())
    }
}

// Specific types
type Requirement struct {
    base *polarion.WorkItem
    BaseCustomFields
    ComplexityPoints *float64
}

func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
    r.base = wi
    cf := polarion.CustomFields(wi.Attributes.CustomFields)
    
    r.LoadBase(cf)
    
    if val, ok := cf.GetFloat("complexityPoints"); ok {
        r.ComplexityPoints = &val
    }
    
    return nil
}
```

### Batch Operations

When working with multiple custom work items:

```go
// Load multiple requirements
func LoadRequirements(ctx context.Context, project *polarion.ProjectClient, query string) ([]*Requirement, error) {
    items, err := project.WorkItems.QueryAll(ctx, query)
    if err != nil {
        return nil, err
    }
    
    requirements := make([]*Requirement, 0, len(items))
    for _, item := range items {
        req := &Requirement{}
        if err := req.LoadFromWorkItem(&item); err != nil {
            continue // or return error
        }
        requirements = append(requirements, req)
    }
    
    return requirements, nil
}

// Update multiple requirements
func UpdateRequirements(ctx context.Context, project *polarion.ProjectClient, reqs []*Requirement) error {
    for _, req := range reqs {
        req.SaveToWorkItem()
        if err := project.WorkItems.Update(ctx, req.base); err != nil {
            return fmt.Errorf("failed to update %s: %w", req.GetID(), err)
        }
    }
    return nil
}
```

### Performance Considerations

When loading many work items, use field selection to reduce payload size:

```go
// Only load fields you need
items, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement",
    polarion.WithFields(polarion.NewFieldSelector().
        WithWorkItemFields("id,title,businessValue,targetRelease")),
)
```

## Best Practices

### 1. Use Pointers for Optional Fields

Always use pointers for optional custom fields to distinguish between "not set" and "zero value":

```go
// Good
SecurityReviewed *bool

// Bad - can't tell if false means "not reviewed" or "reviewed and rejected"
SecurityReviewed bool
```

### 2. Validate Before Saving

Always validate data before saving to Polarion:

```go
func (r *Requirement) Save(ctx context.Context, project *polarion.ProjectClient) error {
    if err := r.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    r.SaveToWorkItem()
    return project.WorkItems.Update(ctx, r.base)
}
```

### 3. Handle Nil Safely

Always check for nil before dereferencing pointers:

```go
// Good
if req.BusinessValue != nil {
    fmt.Printf("Value: %s\n", *req.BusinessValue)
}

// Bad - will panic if nil
fmt.Printf("Value: %s\n", *req.BusinessValue)
```

### 4. Provide Helper Methods

Make your types easy to use with helper methods:

```go
func (r *Requirement) GetID() string {
    if r.base != nil {
        return r.base.ID
    }
    return ""
}

func (r *Requirement) GetTitle() string {
    if r.base != nil && r.base.Attributes != nil {
        return r.base.Attributes.Title
    }
    return ""
}

func (r *Requirement) IsHighPriority() bool {
    return r.BusinessValue != nil && 
           (*r.BusinessValue == "high" || *r.BusinessValue == "critical")
}
```

### 5. Document Field Names

Document the Polarion field names in comments:

```go
type Requirement struct {
    base *polarion.WorkItem
    
    // BusinessValue maps to the "businessValue" custom field in Polarion
    // Valid values: "low", "medium", "high", "critical"
    BusinessValue *string
    
    // TargetRelease maps to the "targetRelease" custom field in Polarion
    // Format: YYYY-MM-DD
    TargetRelease *polarion.DateOnly
}
```

### 6. Use Constants for Enum Values

Define constants for enumeration field values:

```go
const (
    BusinessValueLow      = "low"
    BusinessValueMedium   = "medium"
    BusinessValueHigh     = "high"
    BusinessValueCritical = "critical"
)

// Usage
req.SetBusinessValue(BusinessValueHigh)
```

## Troubleshooting

### Missing Keys

**Problem**: Custom field is not loaded even though it exists in Polarion.

**Solution**: Check the field name spelling and ensure the field exists in the CustomFields map:

```go
func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
    r.base = wi
    cf := polarion.CustomFields(wi.Attributes.CustomFields)
    
    // Debug: Print all available keys
    for key := range cf {
        fmt.Printf("Available field: %s\n", key)
    }
    
    if val, ok := cf.GetString("businessValue"); ok {
        r.BusinessValue = &val
    } else {
        fmt.Println("businessValue field not found")
    }
    
    return nil
}
```

### Type Mismatches

**Problem**: Field exists but `Get*` method returns false.

**Solution**: The field type in Polarion might not match your expectation. Check the actual type:

```go
cf := polarion.CustomFields(wi.Attributes.CustomFields)

// Check raw value
if raw, exists := cf["businessValue"]; exists {
    fmt.Printf("Raw value type: %T, value: %v\n", raw, raw)
}

// Try different getters
if val, ok := cf.GetString("businessValue"); ok {
    fmt.Printf("As string: %s\n", val)
}
if val, ok := cf.GetInt("businessValue"); ok {
    fmt.Printf("As int: %d\n", val)
}
```

### Parse Errors

**Problem**: Date, time, or duration parsing fails.

**Solution**: Check the format of the value in Polarion:

```go
// Date fields should be "YYYY-MM-DD"
if val, ok := cf.GetDateOnly("targetRelease"); ok {
    r.TargetRelease = &val
} else {
    // Try getting as string to see the format
    if str, ok := cf.GetString("targetRelease"); ok {
        fmt.Printf("Date string format: %s\n", str)
        // Try parsing manually
        if date, err := polarion.ParseDateOnly(str); err == nil {
            r.TargetRelease = &date
        } else {
            fmt.Printf("Parse error: %v\n", err)
        }
    }
}
```

### Nil Pointer Panics

**Problem**: Application panics with nil pointer dereference.

**Solution**: Always check for nil before dereferencing:

```go
// Bad
fmt.Printf("Value: %s\n", *req.BusinessValue) // Panics if nil

// Good
if req.BusinessValue != nil {
    fmt.Printf("Value: %s\n", *req.BusinessValue)
} else {
    fmt.Println("Value not set")
}

// Or use a getter with default
func (r *Requirement) GetBusinessValue() string {
    if r.BusinessValue != nil {
        return *r.BusinessValue
    }
    return "not-set" // default value
}
```

### Fields Not Saving

**Problem**: Custom fields are not saved to Polarion.

**Solution**: Ensure you call `SaveToWorkItem()` before updating:

```go
// Bad
req.SetBusinessValue("high")
project.WorkItems.Update(ctx, req.base) // Field not synced!

// Good
req.SetBusinessValue("high")
req.SaveToWorkItem() // Sync to base WorkItem
project.WorkItems.Update(ctx, req.base)

// Or use setter that syncs automatically
func (r *Requirement) SetBusinessValue(value string) {
    r.BusinessValue = &value
    // Automatically sync to base
    if r.base != nil && r.base.Attributes != nil {
        if r.base.Attributes.CustomFields == nil {
            r.base.Attributes.CustomFields = make(map[string]interface{})
        }
        r.base.Attributes.CustomFields["businessValue"] = value
    }
}
```

## Complete Example

For a complete, runnable example, see [`examples/custom_workitems/main.go`](../examples/custom_workitems/main.go).

## Further Reading

- [CustomFields API Documentation](../workitem_custom_fields.go)
- [Field Types Documentation](../workitem_field_types.go)
- [WorkItem Documentation](../workitem.go)
- [Main README](../README.md)
