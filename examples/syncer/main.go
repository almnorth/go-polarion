// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package main demonstrates syncing external data into Polarion with the
// built-in sync engine.
//
// This example shows:
//   - Defining a mapping struct whose JSON tags are Polarion custom field IDs
//   - Populating a work item from an external record
//   - Staging rows and flushing them as batched creates and updates
//   - Enriching the same work items from a second data source
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// ExternalRecord represents data from an external system (database, API, ...).
type ExternalRecord struct {
	ID          string
	Title       string
	Description string
	DueDate     *time.Time
	Priority    string
	IsCompleted bool
}

// Task maps external data onto Polarion custom fields. The JSON tags are the
// custom field IDs. Fields the struct does not declare are left untouched, so
// two syncs can own different fields of the same work item.
type Task struct {
	ExternalID *string            `json:"externalId,omitempty"`
	DueDate    *polarion.DateOnly `json:"dueDate,omitempty"`
	Priority   *string            `json:"priority,omitempty"`
	Completed  *bool              `json:"completed,omitempty"`
}

// taskKey returns the sync key of an external record. Returning false skips it.
func taskKey(r *ExternalRecord) (string, bool) {
	return r.ID, r.ID != ""
}

// populateTask maps one external record onto a work item. It is the single
// source of truth for the mapping.
func populateTask(wi *polarion.WorkItem, r *ExternalRecord) error {
	// A new work item arrives with an empty ID; set the fields Polarion needs
	// on creation. The status must match the initial status ID in the project's
	// workflow, otherwise Polarion rejects the item.
	if wi.ID == "" {
		wi.Attributes.Type = "task"
		wi.Attributes.Status = "open"
	}

	wi.Attributes.Title = r.Title
	if r.Description != "" {
		wi.Attributes.Description = polarion.NewHTMLContent(r.Description)
	}

	t := &Task{}
	if r.ID != "" {
		t.ExternalID = &r.ID
	}
	if r.DueDate != nil {
		d := polarion.NewDateOnly(*r.DueDate)
		t.DueDate = &d
	}
	if r.Priority != "" {
		t.Priority = &r.Priority
	}
	if r.IsCompleted {
		t.Completed = &r.IsCompleted
	}

	return polarion.SaveCustomFields(wi, t)
}

func main() {
	client, err := polarion.New(
		"https://polarion.example.com/rest/v1",
		"your-bearer-token",
	)
	if err != nil {
		log.Fatal(err)
	}

	project := client.Project("myproject")
	ctx := context.Background()

	// Step 1: open a sync session. Polarion is queried once and the result is
	// indexed by the externalId custom field.
	fmt.Println("=== Opening sync session ===")
	sync, err := polarion.NewSync(ctx, project.WorkItems,
		"type:task AND HAS_VALUE:externalId",
		polarion.KeyByFields("externalId"))
	if err != nil {
		log.Printf("Note: query failed (expected in this example): %v", err)
		return
	}

	// Optional: report progress, and preview a run without writing anything.
	// *zap.SugaredLogger satisfies polarion.SyncLogger directly.
	// sync.Logger = logger
	// sync.DryRun = true

	// Step 2: stage the external rows. Rows with no matching work item are
	// staged for creation because Create is true.
	dueDate := time.Now().AddDate(0, 0, 7)
	records := []ExternalRecord{
		{ID: "EXT-001", Title: "Task 1", Description: "First task", DueDate: &dueDate, Priority: "high"},
		{ID: "EXT-002", Title: "Task 2", Priority: "medium", IsCompleted: true},
		{ID: "EXT-003", Title: "Task 3", Description: "Third task", Priority: "low"},
	}

	fmt.Println("=== Staging records ===")
	polarion.Stage(sync, records, polarion.Pass[ExternalRecord]{
		Key:      taskKey,
		Populate: populateTask,
		Create:   true,
	})

	// Step 2b (optional): a second source can enrich the same work items. With
	// Create left false the pass never adds work items of its own, but it does
	// see the ones staged above, so a single item can be built from both sources
	// and still be written once.
	//
	//	polarion.Stage(sync, orders, polarion.Pass[Order]{
	//	    Key:      orderKey,
	//	    Populate: populateOrder,
	//	})

	// Step 3: flush. Everything staged is diffed against what Polarion served,
	// then written as one batched create and one batched update.
	fmt.Println("=== Flushing ===")
	result, err := sync.Flush(ctx)
	if err != nil {
		log.Printf("Sync finished with errors: %v", err)
	}

	fmt.Println("\n=== Sync Summary ===")
	fmt.Printf("Created: %d, Updated: %d, Unchanged: %d, Skipped: %d, Errors: %d\n",
		result.Created, result.Updated, result.Unchanged, result.Skipped, result.Errors)

	// Created items carry the IDs Polarion assigned, which is what you need to
	// link work items after the sync.
	for _, wi := range result.CreatedItems {
		fmt.Printf("Created: %s (%s)\n", wi.ID, wi.Attributes.Title)
	}

	fmt.Println("\n=== Pattern Benefits ===")
	fmt.Println("  ✓ One query, one batched create, one batched update")
	fmt.Println("  ✓ Change detection built in — unchanged items cost nothing")
	fmt.Println("  ✓ Custom fields not declared by the mapping struct are preserved")
	fmt.Println("  ✓ Several sources can build one work item before it is written")
	fmt.Println("  ✓ DryRun previews a run without touching Polarion")
}
