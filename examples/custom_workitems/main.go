// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package main demonstrates how to define and use custom work item types
// with type-safe custom fields in the go-polarion library.
//
// This example shows:
//   - Defining a custom Requirement type with type-safe custom fields
//   - Loading custom fields from a WorkItem
//   - Saving custom fields to a WorkItem
//   - Creating, reading, and updating work items with custom fields
//   - Querying work items by custom field values
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// Requirement represents a custom requirement work item with type-safe custom fields.
// This demonstrates how to wrap a WorkItem with strongly-typed custom field accessors.
type Requirement struct {
	base *polarion.WorkItem

	// Custom fields with proper types (these are example custom field names)
	BusinessValue    *string            // Optional enum field (e.g., "low", "medium", "high", "critical")
	TargetRelease    *polarion.DateOnly // Optional date field
	ComplexityPoints *float64           // Optional float field
	SecurityReviewed *bool              // Optional boolean field
}

// NewRequirement creates a new Requirement with initialized base WorkItem.
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

// GetBusinessValue returns the business value or empty string if not set.
func (r *Requirement) GetBusinessValue() string {
	if r.BusinessValue != nil {
		return *r.BusinessValue
	}
	return ""
}

// SetBusinessValue sets the business value field and syncs it to the base WorkItem.
func (r *Requirement) SetBusinessValue(value string) {
	r.BusinessValue = &value
	if r.base != nil && r.base.Attributes != nil {
		if r.base.Attributes.CustomFields == nil {
			r.base.Attributes.CustomFields = make(map[string]interface{})
		}
		r.base.Attributes.CustomFields["businessValue"] = value
	}
}

// GetTargetRelease returns the target release date or nil if not set.
func (r *Requirement) GetTargetRelease() *polarion.DateOnly {
	return r.TargetRelease
}

// SetTargetRelease sets the target release date field and syncs it to the base WorkItem.
func (r *Requirement) SetTargetRelease(d polarion.DateOnly) {
	r.TargetRelease = &d
	if r.base != nil && r.base.Attributes != nil {
		if r.base.Attributes.CustomFields == nil {
			r.base.Attributes.CustomFields = make(map[string]interface{})
		}
		r.base.Attributes.CustomFields["targetRelease"] = d.String()
	}
}

// GetComplexityPoints returns the complexity points or 0 if not set.
func (r *Requirement) GetComplexityPoints() float64 {
	if r.ComplexityPoints != nil {
		return *r.ComplexityPoints
	}
	return 0.0
}

// SetComplexityPoints sets the complexity points field and syncs it to the base WorkItem.
func (r *Requirement) SetComplexityPoints(points float64) {
	r.ComplexityPoints = &points
	if r.base != nil && r.base.Attributes != nil {
		if r.base.Attributes.CustomFields == nil {
			r.base.Attributes.CustomFields = make(map[string]interface{})
		}
		r.base.Attributes.CustomFields["complexityPoints"] = points
	}
}

// GetSecurityReviewed returns the security reviewed status or false if not set.
func (r *Requirement) GetSecurityReviewed() bool {
	if r.SecurityReviewed != nil {
		return *r.SecurityReviewed
	}
	return false
}

// SetSecurityReviewed sets the security reviewed field and syncs it to the base WorkItem.
func (r *Requirement) SetSecurityReviewed(reviewed bool) {
	r.SecurityReviewed = &reviewed
	if r.base != nil && r.base.Attributes != nil {
		if r.base.Attributes.CustomFields == nil {
			r.base.Attributes.CustomFields = make(map[string]interface{})
		}
		r.base.Attributes.CustomFields["securityReviewed"] = reviewed
	}
}

// LoadFromWorkItem populates the custom type from a WorkItem.
// This method extracts custom fields from the WorkItem and populates
// the type-safe fields in the Requirement struct.
func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
	r.base = wi

	if wi.Attributes == nil {
		return fmt.Errorf("work item attributes are nil")
	}

	cf := polarion.CustomFields(wi.Attributes.CustomFields)

	// Load each custom field using type-safe accessors
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

// SaveToWorkItem updates the base WorkItem with custom field values.
// This method syncs all custom fields from the Requirement struct
// back to the WorkItem's CustomFields map.
func (r *Requirement) SaveToWorkItem() {
	if r.base == nil || r.base.Attributes == nil {
		return
	}

	if r.base.Attributes.CustomFields == nil {
		r.base.Attributes.CustomFields = make(map[string]interface{})
	}

	cf := polarion.CustomFields(r.base.Attributes.CustomFields)

	// Save each custom field
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

// GetBase returns the underlying WorkItem for API operations.
func (r *Requirement) GetBase() *polarion.WorkItem {
	return r.base
}

// GetID returns the work item ID.
func (r *Requirement) GetID() string {
	if r.base != nil {
		return r.base.ID
	}
	return ""
}

// GetTitle returns the work item title.
func (r *Requirement) GetTitle() string {
	if r.base != nil && r.base.Attributes != nil {
		return r.base.Attributes.Title
	}
	return ""
}

func main() {
	// Initialize client
	// Replace with your actual Polarion URL and authentication token
	client, err := polarion.New(
		"https://polarion.example.com/rest/v1",
		"your-bearer-token",
	)
	if err != nil {
		log.Fatal(err)
	}

	project := client.Project("myproject")
	ctx := context.Background()

	// Example 1: Create a new requirement
	fmt.Println("=== Creating New Requirement ===")
	req := NewRequirement("New Requirement with Custom Fields")

	// Set custom fields using type-safe methods
	req.SetBusinessValue("high")
	targetDate := polarion.NewDateOnly(time.Now().AddDate(0, 3, 0)) // 3 months from now
	req.SetTargetRelease(targetDate)
	req.SetComplexityPoints(13.0)
	req.SetSecurityReviewed(false)

	// Save to sync custom fields
	req.SaveToWorkItem()

	// Create in Polarion
	err = project.WorkItems.Create(ctx, req.GetBase())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created requirement: %s\n", req.GetID())
	fmt.Printf("  Title: %s\n", req.GetTitle())
	fmt.Printf("  Business Value: %s\n", req.GetBusinessValue())
	fmt.Printf("  Target Release: %s\n", req.GetTargetRelease().String())
	fmt.Printf("  Complexity Points: %.1f\n", req.GetComplexityPoints())
	fmt.Printf("  Security Reviewed: %t\n", req.GetSecurityReviewed())

	// Example 2: Load existing requirement
	fmt.Println("\n=== Loading Existing Requirement ===")
	wi, err := project.WorkItems.Get(ctx, "REQ-123")
	if err != nil {
		log.Printf("Note: Could not load REQ-123 (this is expected in the example): %v\n", err)
		// Continue with other examples
	} else {
		loadedReq := &Requirement{}
		if err := loadedReq.LoadFromWorkItem(wi); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Loaded requirement: %s\n", loadedReq.GetID())
		fmt.Printf("  Title: %s\n", loadedReq.GetTitle())
		fmt.Printf("  Business Value: %s\n", loadedReq.GetBusinessValue())
		if loadedReq.GetTargetRelease() != nil {
			fmt.Printf("  Target Release: %s\n", loadedReq.GetTargetRelease().String())
		}
		fmt.Printf("  Complexity Points: %.1f\n", loadedReq.GetComplexityPoints())
		fmt.Printf("  Security Reviewed: %t\n", loadedReq.GetSecurityReviewed())

		// Example 3: Update custom fields
		fmt.Println("\n=== Updating Custom Fields ===")
		loadedReq.SetBusinessValue("critical")
		newDate := polarion.NewDateOnly(time.Now().AddDate(0, 1, 0)) // 1 month from now
		loadedReq.SetTargetRelease(newDate)
		loadedReq.SetSecurityReviewed(true)

		// Save changes
		loadedReq.SaveToWorkItem()
		err = project.WorkItems.Update(ctx, loadedReq.GetBase())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Updated requirement: %s\n", loadedReq.GetID())
		fmt.Printf("  New Business Value: %s\n", loadedReq.GetBusinessValue())
		fmt.Printf("  New Target Release: %s\n", loadedReq.GetTargetRelease().String())
		fmt.Printf("  Security Reviewed: %t\n", loadedReq.GetSecurityReviewed())
	}

	// Example 4: Query by type and custom fields
	fmt.Println("\n=== Querying Requirements ===")
	items, err := project.WorkItems.QueryAll(ctx, "type:requirement AND businessValue:high")
	if err != nil {
		log.Printf("Note: Query failed (this is expected in the example): %v\n", err)
	} else {
		fmt.Printf("Found %d high business value requirements:\n", len(items))
		for _, item := range items {
			req := &Requirement{}
			if err := req.LoadFromWorkItem(&item); err != nil {
				continue
			}
			fmt.Printf("  - %s: %s (Business Value: %s, Complexity: %.1f)\n",
				req.GetID(), req.GetTitle(), req.GetBusinessValue(), req.GetComplexityPoints())
		}
	}

	// Example 5: Demonstrate validation
	fmt.Println("\n=== Validation Example ===")
	validationReq := NewRequirement("Requirement with Validation")

	// Validate business value
	validBusinessValues := []string{"low", "medium", "high", "critical"}
	businessValue := "high"
	isValid := false
	for _, vbv := range validBusinessValues {
		if businessValue == vbv {
			isValid = true
			break
		}
	}
	if isValid {
		validationReq.SetBusinessValue(businessValue)
		fmt.Printf("Set valid business value: %s\n", businessValue)
	} else {
		fmt.Printf("Invalid business value: %s\n", businessValue)
	}

	// Validate target release date is in the future
	futureDate := polarion.NewDateOnly(time.Now().AddDate(0, 2, 0))
	if futureDate.After(time.Now()) {
		validationReq.SetTargetRelease(futureDate)
		fmt.Printf("Set future target release: %s\n", futureDate.String())
	}

	// Validate complexity points is positive
	points := 8.0
	if points > 0 {
		validationReq.SetComplexityPoints(points)
		fmt.Printf("Set valid complexity points: %.1f\n", points)
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("  ✓ Defining custom work item types")
	fmt.Println("  ✓ Type-safe custom field access")
	fmt.Println("  ✓ Loading from and saving to WorkItem")
	fmt.Println("  ✓ Creating, reading, and updating work items")
	fmt.Println("  ✓ Querying by custom fields")
	fmt.Println("  ✓ Field validation")
}
