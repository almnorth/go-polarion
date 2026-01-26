// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package main demonstrates basic usage of the Polarion Go client.
//
// This example shows how to:
//   - Create a client
//   - Query work items
//   - Get a specific work item
//   - Create a new work item
//   - Update a work item
//   - Work with users and enumerations
//   - Handle errors
//
// To run this example:
//
//	export POLARION_URL="https://polarion.example.com/rest/v1"
//	export POLARION_TOKEN="your-bearer-token"
//	export POLARION_PROJECT="your-project-id"
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/almnorth/go-polarion"
)

func main() {
	// Get configuration from environment
	baseURL := os.Getenv("POLARION_URL")
	token := os.Getenv("POLARION_TOKEN")
	projectID := os.Getenv("POLARION_PROJECT")

	if baseURL == "" || token == "" || projectID == "" {
		log.Fatal("Please set POLARION_URL, POLARION_TOKEN, and POLARION_PROJECT environment variables")
	}

	// Create client with custom configuration
	client, err := polarion.New(
		baseURL,
		token,
		polarion.WithPageSize(100),
		polarion.WithBatchSize(50),
		polarion.WithTimeout(30*time.Second),
		polarion.WithRetryConfig(polarion.RetryConfig{
			MaxRetries: 3,
			MinWait:    time.Second,
			MaxWait:    30 * time.Second,
			RetryIf:    polarion.IsRetryable,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Connected to Polarion at %s\n", baseURL)

	// Get project-scoped client
	project := client.Project(projectID)

	// Create context
	ctx := context.Background()

	// Example 1: Query work items
	fmt.Println("\n=== Example 1: Query Work Items ===")
	if err := queryWorkItems(ctx, project); err != nil {
		log.Printf("Query error: %v", err)
	}

	// Example 2: Get a specific work item
	fmt.Println("\n=== Example 2: Get Work Item ===")
	if err := getWorkItem(ctx, project); err != nil {
		log.Printf("Get error: %v", err)
	}

	// Example 3: Create a work item
	fmt.Println("\n=== Example 3: Create Work Item ===")
	if err := createWorkItem(ctx, project); err != nil {
		log.Printf("Create error: %v", err)
	}

	// Example 4: Update a work item
	fmt.Println("\n=== Example 4: Update Work Item ===")
	if err := updateWorkItem(ctx, project); err != nil {
		log.Printf("Update error: %v", err)
	}

	// Example 5: List users
	fmt.Println("\n=== Example 5: List Users ===")
	if err := listUsers(ctx, client); err != nil {
		log.Printf("List users error: %v", err)
	}

	// Example 6: List enumerations
	fmt.Println("\n=== Example 6: List Enumerations ===")
	if err := listEnumerations(ctx, project); err != nil {
		log.Printf("List enumerations error: %v", err)
	}

	fmt.Println("\n=== All examples completed ===")
}

// queryWorkItems demonstrates querying work items with filters
func queryWorkItems(ctx context.Context, project *polarion.ProjectClient) error {
	// Define sparse fields to reduce response size
	fields := polarion.NewFieldSelector().
		WithWorkItemFields("title,status,created").
		WithLinkedWorkItemFields("id,role")

	// Execute query with options
	result, err := project.WorkItems.Query(ctx,
		polarion.QueryOptions{
			Query:    "type:requirement",
			PageSize: 10,
			Fields:   fields,
		})
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	fmt.Printf("Found %d work items\n", len(result.Items))
	for i, item := range result.Items {
		if i >= 5 { // Show only first 5
			break
		}
		fmt.Printf("  - %s: %s (Status: %s)\n",
			item.ID,
			item.Attributes.Title,
			item.Attributes.Status)
	}

	return nil
}

// getWorkItem demonstrates getting a specific work item
func getWorkItem(ctx context.Context, project *polarion.ProjectClient) error {
	// Note: Replace with an actual work item ID from your project
	workItemID := "WI-1"

	// Get work item with sparse fields
	fields := polarion.NewFieldSelector().
		WithWorkItemFields("title,description,status,created,updated")

	wi, err := project.WorkItems.Get(ctx, workItemID,
		polarion.WithGetFields(fields))
	if err != nil {
		if polarion.IsNotFound(err) {
			fmt.Printf("Work item %s not found (this is expected if it doesn't exist)\n", workItemID)
			return nil
		}
		return fmt.Errorf("get failed: %w", err)
	}

	fmt.Printf("Work Item: %s\n", wi.ID)
	fmt.Printf("  Title: %s\n", wi.Attributes.Title)
	fmt.Printf("  Status: %s\n", wi.Attributes.Status)
	if wi.Attributes.Description != nil {
		desc := wi.Attributes.Description.Value
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		fmt.Printf("  Description: %s\n", desc)
	}

	return nil
}

// createWorkItem demonstrates creating a new work item
func createWorkItem(ctx context.Context, project *polarion.ProjectClient) error {
	// Create a new work item
	newWI := &polarion.WorkItem{
		Type: "workitems",
		Attributes: &polarion.WorkItemAttributes{
			Title: fmt.Sprintf("Example Task - %s", time.Now().Format("2006-01-02 15:04:05")),
			Description: polarion.NewPlainTextContent(
				"This is an example work item created by the Go client",
			),
			Status: "open",
		},
	}

	// Note: You may need to set the work item type via custom field
	// depending on your Polarion configuration
	newWI.Attributes.SetCustomField("type", "task")

	err := project.WorkItems.Create(ctx, newWI)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}

	fmt.Printf("Created work item: %s\n", newWI.ID)
	fmt.Printf("  Title: %s\n", newWI.Attributes.Title)

	return nil
}

// updateWorkItem demonstrates updating a work item
func updateWorkItem(ctx context.Context, project *polarion.ProjectClient) error {
	// Note: Replace with an actual work item ID from your project
	workItemID := "WI-1"

	// Get the work item first
	wi, err := project.WorkItems.Get(ctx, workItemID)
	if err != nil {
		if polarion.IsNotFound(err) {
			fmt.Printf("Work item %s not found (skipping update example)\n", workItemID)
			return nil
		}
		return fmt.Errorf("get failed: %w", err)
	}

	// Update the work item
	originalTitle := wi.Attributes.Title
	wi.Attributes.Title = fmt.Sprintf("%s (Updated at %s)", originalTitle, time.Now().Format("15:04:05"))

	err = project.WorkItems.Update(ctx, wi)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("Updated work item: %s\n", wi.ID)
	fmt.Printf("  New title: %s\n", wi.Attributes.Title)

	return nil
}

// listUsers demonstrates listing users
func listUsers(ctx context.Context, client *polarion.Client) error {
	// List users with pagination
	users, err := client.Users.List(ctx,
		polarion.WithQueryPageSize(10))
	if err != nil {
		return fmt.Errorf("list users failed: %w", err)
	}

	fmt.Printf("Found %d users\n", len(users))
	for i, user := range users {
		if i >= 5 { // Show only first 5
			break
		}
		fmt.Printf("  - %s: %s (%s)\n",
			user.ID,
			user.Attributes.Name,
			user.Attributes.Email)
	}

	return nil
}

// listEnumerations demonstrates listing enumerations
func listEnumerations(ctx context.Context, project *polarion.ProjectClient) error {
	enums, err := project.Enumerations.List(ctx)
	if err != nil {
		return fmt.Errorf("list enumerations failed: %w", err)
	}

	fmt.Printf("Found %d enumerations\n", len(enums))
	for i, enum := range enums {
		if i >= 5 { // Show only first 5
			break
		}
		optCount := 0
		if enum.Attributes != nil {
			optCount = len(enum.Attributes.Options)
		}
		fmt.Printf("  - %s (%d options)\n", enum.ID, optCount)
	}

	return nil
}
