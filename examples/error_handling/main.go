// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/almnorth/go-polarion"
)

func main() {
	// This example demonstrates how to handle errors from the Polarion API
	// and extract detailed error information.

	// Create a client (replace with your actual URL and token)
	client, err := polarion.New(
		"https://polarion.example.com/rest/v1",
		"your-bearer-token",
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Basic error handling
	fmt.Println("=== Example 1: Basic Error Handling ===")
	basicErrorHandling(ctx, client)

	// Example 2: Detailed error information
	fmt.Println("\n=== Example 2: Detailed Error Information ===")
	detailedErrorHandling(ctx, client)

	// Example 3: Programmatic error inspection
	fmt.Println("\n=== Example 3: Programmatic Error Inspection ===")
	programmaticErrorHandling(ctx, client)

	// Example 4: Handling specific error types
	fmt.Println("\n=== Example 4: Handling Specific Error Types ===")
	specificErrorHandling(ctx, client)
}

// basicErrorHandling demonstrates simple error checking
func basicErrorHandling(ctx context.Context, client *polarion.Client) {
	project := client.Project("PROJECT")
	_, err := project.WorkItems.Get(ctx, "NONEXISTENT-123")
	if err != nil {
		// Simple error check
		if polarion.IsNotFound(err) {
			fmt.Println("Work item not found (this is expected)")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// detailedErrorHandling demonstrates getting detailed error information
func detailedErrorHandling(ctx context.Context, client *polarion.Client) {
	// Try to create a work item with invalid data
	workItem := &polarion.WorkItem{
		Type: "task",
		Attributes: &polarion.WorkItemAttributes{
			Title: "Test Item",
		},
	}

	project := client.Project("PROJECT")
	err := project.WorkItems.Create(ctx, workItem)
	if err != nil {
		// Get detailed error message including raw response
		detailedMsg := polarion.GetDetailedAPIError(err)
		fmt.Printf("Detailed error:\n%s\n", detailedMsg)
	}
}

// programmaticErrorHandling demonstrates inspecting error details programmatically
func programmaticErrorHandling(ctx context.Context, client *polarion.Client) {
	// Try to create a work item with a type mismatch in custom fields
	workItem := &polarion.WorkItem{
		Type: "task",
		Attributes: &polarion.WorkItemAttributes{
			Title: "Test Item",
		},
	}

	project := client.Project("PROJECT")
	err := project.WorkItems.Create(ctx, workItem)
	if err != nil {
		// Get error details
		details := polarion.GetAPIErrorDetails(err)
		if details != nil {
			fmt.Println("Error details:")
			for i, detail := range details {
				fmt.Printf("  %d. Field: %s\n", i+1, detail.Pointer)
				fmt.Printf("     Error: %s\n", detail.Detail)

				// Handle specific field errors
				if strings.Contains(detail.Pointer, "customFields") {
					fmt.Println("     -> This is a custom field error")
				}

				if strings.Contains(detail.Detail, "STRING expected") {
					fmt.Println("     -> Type mismatch: expected string")
				}
			}
		}
	}
}

// specificErrorHandling demonstrates handling specific error types
func specificErrorHandling(ctx context.Context, client *polarion.Client) {
	workItem := &polarion.WorkItem{
		Type: "task",
		Attributes: &polarion.WorkItemAttributes{
			Title: "Test Item",
		},
	}

	project := client.Project("PROJECT")
	err := project.WorkItems.Create(ctx, workItem)
	if err != nil {
		// Check for validation errors
		if polarion.IsValidationError(err) {
			fmt.Println("Validation error detected")
			var valErr *polarion.ValidationError
			if errors.As(err, &valErr) {
				fmt.Printf("  Field: %s\n", valErr.Field)
				fmt.Printf("  Message: %s\n", valErr.Message)
			}
			return
		}

		// Check for API errors
		var apiErr *polarion.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("API Error (status %d)\n", apiErr.StatusCode)

			// Handle different status codes
			switch apiErr.StatusCode {
			case 400:
				fmt.Println("  -> Bad Request: Check your input data")
				// Show which fields are problematic
				for _, detail := range apiErr.Details {
					if detail.Pointer != "" {
						fmt.Printf("     Problem with field: %s\n", detail.Pointer)
					}
				}
			case 401:
				fmt.Println("  -> Unauthorized: Check your authentication token")
			case 403:
				fmt.Println("  -> Forbidden: You don't have permission for this operation")
			case 404:
				fmt.Println("  -> Not Found: The resource doesn't exist")
			case 429:
				fmt.Println("  -> Rate Limited: Too many requests, retry later")
			case 500:
				fmt.Println("  -> Server Error: Problem on Polarion's side")
			default:
				fmt.Printf("  -> HTTP %d: %s\n", apiErr.StatusCode, apiErr.Message)
			}

			// Check if error is retryable
			if polarion.IsRetryable(err) {
				fmt.Println("  -> This error is retryable")
			}
		}
	}
}

// Example of a robust error handling function
func createWorkItemWithErrorHandling(ctx context.Context, client *polarion.Client, projectID string, wi *polarion.WorkItem) error {
	project := client.Project(projectID)
	err := project.WorkItems.Create(ctx, wi)
	if err != nil {
		// Check for validation errors
		if polarion.IsValidationError(err) {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Get detailed API error information
		var apiErr *polarion.APIError
		if errors.As(err, &apiErr) {
			// Log detailed error for debugging
			log.Printf("API Error (status %d): %s", apiErr.StatusCode, apiErr.Message)

			// Handle specific field errors
			for _, detail := range apiErr.Details {
				log.Printf("  Field: %s", detail.Pointer)
				log.Printf("  Error: %s", detail.Detail)

				// Example: Handle custom field type mismatches
				if strings.Contains(detail.Detail, "STRING expected") &&
					strings.Contains(detail.Pointer, "customFields") {
					// Extract field name from pointer
					parts := strings.Split(detail.Pointer, "/")
					fieldName := parts[len(parts)-1]
					return fmt.Errorf("custom field '%s' expects a string value, but got a different type", fieldName)
				}

				// Example: Handle missing required fields
				if strings.Contains(detail.Detail, "required") {
					parts := strings.Split(detail.Pointer, "/")
					fieldName := parts[len(parts)-1]
					return fmt.Errorf("required field '%s' is missing", fieldName)
				}
			}

			// Return user-friendly error based on status code
			switch apiErr.StatusCode {
			case 400:
				return fmt.Errorf("invalid work item data: %s", apiErr.Message)
			case 401:
				return fmt.Errorf("authentication failed: check your token")
			case 403:
				return fmt.Errorf("permission denied: you don't have access to create work items in project %s", projectID)
			case 404:
				return fmt.Errorf("project %s not found", projectID)
			default:
				return fmt.Errorf("failed to create work item: %s", apiErr.Message)
			}
		}

		return fmt.Errorf("unexpected error: %w", err)
	}

	return nil
}
