// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package main demonstrates a clean pattern for syncing external data to Polarion.
//
// This example shows:
//   - Defining a typed work item wrapper with JSON tags for custom fields
//   - Populating work items from external data sources
//   - Efficient sync with change detection using Clone() and Equals()
//   - Creating new items and updating only changed items
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// ExternalRecord represents data from an external system (e.g., database, API, etc.)
type ExternalRecord struct {
	ID          string
	Title       string
	Description string
	DueDate     *time.Time
	Priority    string
	IsCompleted bool
}

// Task represents a Polarion work item with type-safe custom fields.
// JSON tags map directly to Polarion custom field IDs.
type Task struct {
	base       *polarion.WorkItem
	ExternalID *string            `json:"externalId,omitempty"` // Links to external system
	DueDate    *polarion.DateOnly `json:"dueDate,omitempty"`
	Priority   *string            `json:"priority,omitempty"`
	Completed  *bool              `json:"completed,omitempty"`
}

// PopulateFromExternal populates the Task from an external record.
// This is the single source of truth for mapping external data to Polarion.
func (t *Task) PopulateFromExternal(record *ExternalRecord) error {
	// Ensure base work item exists
	if t.base == nil {
		t.base = &polarion.WorkItem{
			Type: "workitems",
			Attributes: &polarion.WorkItemAttributes{
				Type:   "task",
				Status: "open", // This must match the initial status ID in Polarion, otherwise you will get an invalid status
			},
		}
	}
	if t.base.Attributes == nil {
		t.base.Attributes = &polarion.WorkItemAttributes{
			Type:   "task",
			Status: "open", // This must match the initial status ID in Polarion, otherwise you will get an invalid status
		}
	}
	// Ensure CustomFields map exists (preserve fields we don't manage)
	if t.base.Attributes.CustomFields == nil {
		t.base.Attributes.CustomFields = make(map[string]interface{})
	}

	// Map standard attributes
	t.base.Attributes.Title = record.Title
	if record.Description != "" {
		t.base.Attributes.Description = polarion.NewHTMLContent(record.Description)
	} else {
		t.base.Attributes.Description = nil
	}

	// Reset typed fields (nil fields will be deleted from CustomFields)
	t.ExternalID = nil
	t.DueDate = nil
	t.Priority = nil
	t.Completed = nil

	// Map external fields to typed fields
	if record.ID != "" {
		t.ExternalID = &record.ID
	}
	if record.DueDate != nil {
		date := polarion.NewDateOnly(*record.DueDate)
		t.DueDate = &date
	}
	if record.Priority != "" {
		t.Priority = &record.Priority
	}
	// Only set completed if true (avoids false positives in change detection)
	if record.IsCompleted {
		t.Completed = &record.IsCompleted
	}

	// Save typed fields to work item's CustomFields map
	return polarion.SaveCustomFields(t.base, t)
}

// SyncResult tracks synchronization statistics
type SyncResult struct {
	Created int
	Updated int
	Skipped int
	Errors  int
}

// BuildWorkItemMap creates a map of work items indexed by external ID
func BuildWorkItemMap(items []polarion.WorkItem) map[string]*polarion.WorkItem {
	result := make(map[string]*polarion.WorkItem)
	for i := range items {
		item := &items[i]
		if item.Attributes != nil && item.Attributes.CustomFields != nil {
			if extID, ok := item.Attributes.CustomFields["externalId"].(string); ok && extID != "" {
				result[extID] = item
			}
		}
	}
	return result
}

// Sync synchronizes a single external record to Polarion
func Sync(
	ctx context.Context,
	project *polarion.ProjectClient,
	record *ExternalRecord,
	workItemMap map[string]*polarion.WorkItem,
	result *SyncResult,
) error {
	if record.ID == "" {
		return fmt.Errorf("record has no ID")
	}

	existingWorkItem, exists := workItemMap[record.ID]

	if exists {
		// Clone the work item and apply updates
		updatedWorkItem := existingWorkItem.Clone()
		task := &Task{base: updatedWorkItem}
		if err := task.PopulateFromExternal(record); err != nil {
			return fmt.Errorf("failed to apply updates: %w", err)
		}

		// Check if there are actual changes
		if !existingWorkItem.Equals(updatedWorkItem, project.WorkItems) {
			if err := project.WorkItems.UpdateWithOldValue(ctx, existingWorkItem, updatedWorkItem); err != nil {
				return fmt.Errorf("failed to update: %w", err)
			}
			result.Updated++
			fmt.Printf("Updated: %s (External ID: %s)\n", existingWorkItem.ID, record.ID)
		} else {
			result.Skipped++
			fmt.Printf("Skipped: %s (no changes)\n", existingWorkItem.ID)
		}
	} else {
		// Create new work item
		task := &Task{}
		if err := task.PopulateFromExternal(record); err != nil {
			return fmt.Errorf("failed to convert: %w", err)
		}
		if err := project.WorkItems.Create(ctx, task.base); err != nil {
			return fmt.Errorf("failed to create: %w", err)
		}
		result.Created++
		fmt.Printf("Created: %s (External ID: %s)\n", task.base.ID, record.ID)

		// Add to map for future lookups
		workItemMap[record.ID] = task.base
	}

	return nil
}

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

	// Step 1: Fetch existing work items from Polarion
	fmt.Println("=== Fetching existing work items ===")
	existingItems, err := project.WorkItems.QueryAll(ctx, "type:task AND externalId:*")
	if err != nil {
		log.Printf("Note: Query failed (expected in example): %v\n", err)
		existingItems = []polarion.WorkItem{} // Continue with empty list
	}
	fmt.Printf("Found %d existing items\n", len(existingItems))

	// Step 2: Build lookup map by external ID
	workItemMap := BuildWorkItemMap(existingItems)

	// Step 3: Simulate external data (replace with your data source)
	dueDate := time.Now().AddDate(0, 0, 7) // 1 week from now
	externalRecords := []*ExternalRecord{
		{ID: "EXT-001", Title: "Task 1", Description: "First task", DueDate: &dueDate, Priority: "high"},
		{ID: "EXT-002", Title: "Task 2", Priority: "medium", IsCompleted: true},
		{ID: "EXT-003", Title: "Task 3", Description: "Third task", Priority: "low"},
	}

	// Step 4: Sync each record
	fmt.Println("\n=== Syncing records ===")
	result := &SyncResult{}
	for _, record := range externalRecords {
		if err := Sync(ctx, project, record, workItemMap, result); err != nil {
			fmt.Printf("Error syncing %s: %v\n", record.ID, err)
			result.Errors++
		}
	}

	// Step 5: Print summary
	fmt.Println("\n=== Sync Summary ===")
	fmt.Printf("Created: %d, Updated: %d, Skipped: %d, Errors: %d\n",
		result.Created, result.Updated, result.Skipped, result.Errors)

	fmt.Println("\n=== Pattern Benefits ===")
	fmt.Println("  ✓ Single PopulateFromExternal method for all mapping logic")
	fmt.Println("  ✓ Clone() + Equals() for efficient change detection")
	fmt.Println("  ✓ UpdateWithOldValue() sends only changed fields")
	fmt.Println("  ✓ Preserves custom fields not managed by the sync")
	fmt.Println("  ✓ Type-safe custom fields with JSON tags")
	fmt.Println("  ✓ Minimal boilerplate code")
}
