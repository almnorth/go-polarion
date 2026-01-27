// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package main demonstrates the simplified approach to custom work items
// using automatic field mapping with JSON tags and reflection.
//
// This example shows how to use LoadCustomFields and SaveCustomFields
// to eliminate boilerplate code for custom field handling.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// Requirement represents a custom requirement work item with type-safe custom fields.
// Using JSON tags, the fields are automatically mapped to/from Polarion custom fields.
type Requirement struct {
	base *polarion.WorkItem

	// Custom fields - just add json tags matching the Polarion field IDs
	BusinessValue    *string            `json:"businessValue"`
	TargetRelease    *polarion.DateOnly `json:"targetRelease"`
	ComplexityPoints *float64           `json:"complexityPoints"`
	SecurityReviewed *bool              `json:"securityReviewed"`
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

// LoadFromWorkItem populates the custom type from a WorkItem using automatic field mapping.
func (r *Requirement) LoadFromWorkItem(wi *polarion.WorkItem) error {
	r.base = wi
	return polarion.LoadCustomFields(wi, r)
}

// SaveToWorkItem updates the base WorkItem with custom field values using automatic field mapping.
func (r *Requirement) SaveToWorkItem() error {
	return polarion.SaveCustomFields(r.base, r)
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

// Helper functions for creating pointers
func stringPtr(s string) *string                     { return &s }
func float64Ptr(f float64) *float64                  { return &f }
func boolPtr(b bool) *bool                           { return &b }
func datePtr(d polarion.DateOnly) *polarion.DateOnly { return &d }

func main() {
	// Initialize client
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
	req := NewRequirement("New Requirement with Simplified Custom Fields")

	// Set custom fields directly - no setter methods needed!
	req.BusinessValue = stringPtr("high")
	req.TargetRelease = datePtr(polarion.NewDateOnly(time.Now().AddDate(0, 3, 0)))
	req.ComplexityPoints = float64Ptr(13.0)
	req.SecurityReviewed = boolPtr(false)

	// Save to sync custom fields - automatic mapping!
	if err := req.SaveToWorkItem(); err != nil {
		log.Fatal(err)
	}

	// Create in Polarion
	err = project.WorkItems.Create(ctx, req.GetBase())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created requirement: %s\n", req.GetID())
	fmt.Printf("  Title: %s\n", req.GetTitle())
	if req.BusinessValue != nil {
		fmt.Printf("  Business Value: %s\n", *req.BusinessValue)
	}
	if req.TargetRelease != nil {
		fmt.Printf("  Target Release: %s\n", req.TargetRelease.String())
	}
	if req.ComplexityPoints != nil {
		fmt.Printf("  Complexity Points: %.1f\n", *req.ComplexityPoints)
	}
	if req.SecurityReviewed != nil {
		fmt.Printf("  Security Reviewed: %t\n", *req.SecurityReviewed)
	}

	// Example 2: Load existing requirement
	fmt.Println("\n=== Loading Existing Requirement ===")
	wi, err := project.WorkItems.Get(ctx, "REQ-123")
	if err != nil {
		log.Printf("Note: Could not load REQ-123 (this is expected in the example): %v\n", err)
	} else {
		loadedReq := &Requirement{}
		// Automatic loading - no manual field extraction!
		if err := loadedReq.LoadFromWorkItem(wi); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Loaded requirement: %s\n", loadedReq.GetID())
		fmt.Printf("  Title: %s\n", loadedReq.GetTitle())
		if loadedReq.BusinessValue != nil {
			fmt.Printf("  Business Value: %s\n", *loadedReq.BusinessValue)
		}
		if loadedReq.TargetRelease != nil {
			fmt.Printf("  Target Release: %s\n", loadedReq.TargetRelease.String())
		}
		if loadedReq.ComplexityPoints != nil {
			fmt.Printf("  Complexity Points: %.1f\n", *loadedReq.ComplexityPoints)
		}
		if loadedReq.SecurityReviewed != nil {
			fmt.Printf("  Security Reviewed: %t\n", *loadedReq.SecurityReviewed)
		}

		// Example 3: Update custom fields
		fmt.Println("\n=== Updating Custom Fields ===")
		loadedReq.BusinessValue = stringPtr("critical")
		loadedReq.TargetRelease = datePtr(polarion.NewDateOnly(time.Now().AddDate(0, 1, 0)))
		loadedReq.SecurityReviewed = boolPtr(true)

		// Save changes - automatic mapping!
		if err := loadedReq.SaveToWorkItem(); err != nil {
			log.Fatal(err)
		}
		err = project.WorkItems.Update(ctx, loadedReq.GetBase())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Updated requirement: %s\n", loadedReq.GetID())
	}

	// Example 4: Query and batch load
	fmt.Println("\n=== Querying Requirements ===")
	items, err := project.WorkItems.QueryAll(ctx, "type:requirement AND businessValue:high")
	if err != nil {
		log.Printf("Note: Query failed (this is expected in the example): %v\n", err)
	} else {
		fmt.Printf("Found %d high business value requirements:\n", len(items))
		for _, item := range items {
			req := &Requirement{}
			// Automatic loading for each item
			if err := req.LoadFromWorkItem(&item); err != nil {
				continue
			}
			businessValue := ""
			if req.BusinessValue != nil {
				businessValue = *req.BusinessValue
			}
			complexity := 0.0
			if req.ComplexityPoints != nil {
				complexity = *req.ComplexityPoints
			}
			fmt.Printf("  - %s: %s (Business Value: %s, Complexity: %.1f)\n",
				req.GetID(), req.GetTitle(), businessValue, complexity)
		}
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Benefits of the simplified approach:")
	fmt.Println("  ✓ No manual LoadFromWorkItem implementation needed")
	fmt.Println("  ✓ No manual SaveToWorkItem implementation needed")
	fmt.Println("  ✓ No getter/setter methods required")
	fmt.Println("  ✓ Just use JSON tags matching Polarion field IDs")
	fmt.Println("  ✓ Automatic type conversion and validation")
	fmt.Println("  ✓ Less boilerplate, fewer bugs")
}
