// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

/*
Package polarion provides a Go client for the Polarion REST API.

The client supports comprehensive work item management, project operations,
user and group management, enumerations, and more. It features automatic
retry logic with exponential backoff, sparse field selection, pagination,
and batch operations.

# Installation

	go get github.com/almnorth/go-polarion

# Quick Start

Create a client and start working with Polarion:

	package main

	import (
		"context"
		"fmt"
		"log"

		"github.com/almnorth/go-polarion"
	)

	func main() {
		// Create client
		client, err := polarion.New(
			"https://polarion.example.com/rest/v1",
			"your-bearer-token",
			polarion.WithPageSize(100),
			polarion.WithBatchSize(50),
		)
		if err != nil {
			log.Fatal(err)
		}

		// Get a project-scoped client
		project := client.Project("my-project")

		// Get a work item
		ctx := context.Background()
		wi, err := project.WorkItems.Get(ctx, "WI-123")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Work Item: %s - %s\n", wi.ID, wi.Attributes.Title)
	}

# Features

The client provides the following features:

  - Work Item Operations: Create, read, update, delete, and query work items
  - Project Management: List and manage projects
  - User Management: Manage users and user groups
  - Enumerations: Manage custom enumerations
  - Work Item Types: Query available work item types
  - Work Item Links: Manage relationships between work items
  - Comments: Add and manage work item comments
  - Retry Logic: Automatic retry with exponential backoff
  - Pagination: Automatic handling of paginated responses
  - Sparse Fields: Select only the fields you need
  - Batch Operations: Create multiple work items efficiently

# Work Items

Work with work items using the WorkItemService:

	// Get a work item
	wi, err := project.WorkItems.Get(ctx, "WI-123")

	// Query work items
	query := polarion.NewQuery().
		Where("type", "requirement").
		Where("status", "open").
		OrderBy("created", false)

	items, err := project.WorkItems.Query(ctx, query)

	// Create a work item
	newWI := &polarion.WorkItem{
		Type: "workitems",
		Attributes: &polarion.WorkItemAttributes{
			Type:        "requirement",
			Title:       "New Requirement",
			Description: "Description here",
		},
	}
	err = project.WorkItems.Create(ctx, newWI)

	// Update a work item
	wi.Attributes.Status = "in-progress"
	err = project.WorkItems.Update(ctx, wi)

	// Delete a work item
	err = project.WorkItems.Delete(ctx, "WI-123")

# Sparse Fields

Select only the fields you need to reduce response size:

	fields := polarion.NewFields().
		Add("title", "status", "type").
		AddRelated("assignee", "id", "name")

	wi, err := project.WorkItems.Get(ctx, "WI-123",
		polarion.WithFields(fields))

# Pagination

Query results are automatically paginated:

	query := polarion.NewQuery().Where("type", "requirement")

	// Get all items (handles pagination automatically)
	items, err := project.WorkItems.Query(ctx, query)

	// Or use custom page size
	items, err := project.WorkItems.Query(ctx, query,
		polarion.WithPageSize(50))

# Retry Configuration

Configure automatic retry behavior:

	client, err := polarion.New(
		baseURL,
		token,
		polarion.WithRetryConfig(polarion.RetryConfig{
			MaxRetries: 3,
			MinWait:    time.Second,
			MaxWait:    30 * time.Second,
			RetryIf: func(err error) bool {
				// Custom retry logic
				return polarion.IsRetryable(err)
			},
		}),
	)

# Error Handling

The client provides typed errors for better error handling:

	wi, err := project.WorkItems.Get(ctx, "WI-123")
	if err != nil {
		if polarion.IsNotFound(err) {
			// Handle not found
		} else if polarion.IsValidationError(err) {
			// Handle validation error
		} else {
			// Handle other errors
		}
	}

# Examples

For more examples, see the examples directory:
  - examples/basic/main.go - Complete working example

# API Documentation

For detailed API documentation, visit https://pkg.go.dev/github.com/almnorth/go-polarion

# License

This project is licensed under the Apache License 2.0.
See the LICENSE file for details.
*/
package polarion
