// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion_test

import (
	"context"
	"os"
	"testing"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// TestWorkItemLifecycle tests the complete lifecycle of a work item:
// create, get, parse, update, and delete
func TestWorkItemLifecycle(t *testing.T) {
	// Skip if no token is provided
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	projectID := os.Getenv("POLARION_PROJECT")
	if projectID == "" {
		t.Skip("POLARION_PROJECT not set, skipping integration test")
	}

	// Create client
	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	project := client.Project(projectID)
	ctx := context.Background()

	// Test 1: Create a work item
	t.Run("CreateWorkItem", func(t *testing.T) {
		wi := &polarion.WorkItem{
			Type: "workitems",
			Attributes: &polarion.WorkItemAttributes{
				Type:         "feature",
				Title:        "Test Feature from Go Test",
				Description:  polarion.NewPlainTextContent("This is a test feature created by automated tests"),
				Status:       "active",
				Priority:     "50.0",
				Severity:     "should_have",
				CustomFields: make(map[string]interface{}),
			},
		}

		// Add some custom fields
		wi.Attributes.CustomFields["someStringField"] = "Test string value"
		wi.Attributes.CustomFields["someNumber"] = 42
		wi.Attributes.CustomFields["booleanField"] = true

		err := project.WorkItems.Create(ctx, wi)
		if err != nil {
			t.Fatalf("Failed to create work item: %v", err)
		}

		if wi.ID == "" {
			t.Fatal("Work item ID is empty after creation")
		}

		t.Logf("Created work item: %s", wi.ID)

		// Store ID for subsequent tests
		workItemID := wi.ID

		// Test 2: Get the work item
		t.Run("GetWorkItem", func(t *testing.T) {
			fetched, err := project.WorkItems.Get(ctx, workItemID)
			if err != nil {
				t.Fatalf("Failed to get work item: %v", err)
			}

			if fetched.ID != workItemID {
				t.Errorf("Expected ID %s, got %s", workItemID, fetched.ID)
			}

			if fetched.Attributes.Title != "Test Feature from Go Test" {
				t.Errorf("Expected title 'Test Feature from Go Test', got '%s'", fetched.Attributes.Title)
			}

			if fetched.Attributes.Type != "feature" {
				t.Errorf("Expected type 'feature', got '%s'", fetched.Attributes.Type)
			}

			t.Logf("Successfully fetched work item: %s", fetched.ID)
		})

		// Test 3: Parse custom fields
		t.Run("ParseCustomFields", func(t *testing.T) {
			fetched, err := project.WorkItems.Get(ctx, workItemID)
			if err != nil {
				t.Fatalf("Failed to get work item: %v", err)
			}

			cf := polarion.CustomFields(fetched.Attributes.CustomFields)

			// Test string field
			if val, ok := cf.GetString("someStringField"); ok {
				if val != "Test string value" {
					t.Errorf("Expected 'Test string value', got '%s'", val)
				}
				t.Logf("String field: %s", val)
			} else {
				t.Error("Failed to get string field")
			}

			// Test integer field
			if val, ok := cf.GetInt("someNumber"); ok {
				if val != 42 {
					t.Errorf("Expected 42, got %d", val)
				}
				t.Logf("Integer field: %d", val)
			} else {
				t.Error("Failed to get integer field")
			}

			// Test boolean field
			if val, ok := cf.GetBool("booleanField"); ok {
				if !val {
					t.Error("Expected true, got false")
				}
				t.Logf("Boolean field: %t", val)
			} else {
				t.Error("Failed to get boolean field")
			}
		})

		// Test 4: Update the work item
		t.Run("UpdateWorkItem", func(t *testing.T) {
			fetched, err := project.WorkItems.Get(ctx, workItemID)
			if err != nil {
				t.Fatalf("Failed to get work item: %v", err)
			}

			// Update standard fields
			fetched.Attributes.Title = "Updated Test Feature"
			fetched.Attributes.Status = "inactive"

			// Update custom fields
			cf := polarion.CustomFields(fetched.Attributes.CustomFields)
			cf.Set("someStringField", "Updated string value")
			cf.Set("someNumber", 100)
			cf.Set("oneFloatField", 3.14)

			err = project.WorkItems.Update(ctx, fetched)
			if err != nil {
				t.Fatalf("Failed to update work item: %v", err)
			}

			t.Logf("Successfully updated work item: %s", fetched.ID)

			// Verify the update
			updated, err := project.WorkItems.Get(ctx, workItemID)
			if err != nil {
				t.Fatalf("Failed to get updated work item: %v", err)
			}

			if updated.Attributes.Title != "Updated Test Feature" {
				t.Errorf("Expected title 'Updated Test Feature', got '%s'", updated.Attributes.Title)
			}

			if updated.Attributes.Status != "inactive" {
				t.Errorf("Expected status 'inactive', got '%s'", updated.Attributes.Status)
			}

			cfUpdated := polarion.CustomFields(updated.Attributes.CustomFields)
			if val, ok := cfUpdated.GetString("someStringField"); ok {
				if val != "Updated string value" {
					t.Errorf("Expected 'Updated string value', got '%s'", val)
				}
			} else {
				t.Error("Failed to get updated string field")
			}

			if val, ok := cfUpdated.GetInt("someNumber"); ok {
				if val != 100 {
					t.Errorf("Expected 100, got %d", val)
				}
			} else {
				t.Error("Failed to get updated integer field")
			}
		})

		// Test 5: Change work item status and custom field
		t.Run("ChangeWorkItemStatusAndCustomField", func(t *testing.T) {
			fetched, err := project.WorkItems.Get(ctx, workItemID)
			if err != nil {
				t.Fatalf("Failed to get work item: %v", err)
			}

			// Change status from inactive to active
			originalStatus := fetched.Attributes.Status
			t.Logf("Original status: %s", originalStatus)

			newStatus := "active"
			if originalStatus == "active" {
				newStatus = "inactive"
			}

			// Also update a custom field
			cf := polarion.CustomFields(fetched.Attributes.CustomFields)
			originalCustomValue, _ := cf.GetString("someStringField")
			t.Logf("Original custom field value: %s", originalCustomValue)

			newCustomValue := "Status changed to " + newStatus
			cf.Set("someStringField", newCustomValue)

			fetched.Attributes.Status = newStatus
			err = project.WorkItems.Update(ctx, fetched)
			if err != nil {
				t.Fatalf("Failed to update work item: %v", err)
			}

			t.Logf("Successfully changed status to: %s and custom field to: %s", newStatus, newCustomValue)

			// Verify both the status change and custom field update
			updated, err := project.WorkItems.Get(ctx, workItemID)
			if err != nil {
				t.Fatalf("Failed to get updated work item: %v", err)
			}

			if updated.Attributes.Status != newStatus {
				t.Errorf("Expected status '%s', got '%s'", newStatus, updated.Attributes.Status)
			}

			cfUpdated := polarion.CustomFields(updated.Attributes.CustomFields)
			if updatedValue, ok := cfUpdated.GetString("someStringField"); ok {
				if updatedValue != newCustomValue {
					t.Errorf("Expected custom field value '%s', got '%s'", newCustomValue, updatedValue)
				}
				t.Logf("Custom field update verified: %s -> %s", originalCustomValue, updatedValue)
			} else {
				t.Error("Failed to get updated custom field value")
			}

			t.Logf("Status change verified: %s -> %s", originalStatus, updated.Attributes.Status)
		})

		// Test 6: Delete the work item (cleanup)
		// Note: DELETE may not be supported by all Polarion instances
		t.Run("DeleteWorkItem", func(t *testing.T) {
			err := project.WorkItems.Delete(ctx, workItemID)
			if err != nil {
				// Log warning but don't fail the test if delete is not supported
				t.Logf("Warning: Failed to delete work item (may not be supported): %v", err)
				t.Skip("Skipping delete verification as delete operation failed")
				return
			}

			t.Logf("Successfully deleted work item: %s", workItemID)

			// Verify deletion
			_, err = project.WorkItems.Get(ctx, workItemID)
			if err == nil {
				t.Error("Expected error when getting deleted work item, got nil")
			}
		})
	})
}

// TestWorkItemWithComplexFields tests work items with complex field types
func TestWorkItemWithComplexFields(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	projectID := os.Getenv("POLARION_PROJECT")
	if projectID == "" {
		t.Skip("POLARION_PROJECT not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	project := client.Project(projectID)
	ctx := context.Background()

	// Create work item with complex fields
	wi := &polarion.WorkItem{
		Type: "workitems",
		Attributes: &polarion.WorkItemAttributes{
			Type:         "feature",
			Title:        "Test Complex Fields",
			Description:  polarion.NewHTMLContent("<p>Test with <strong>HTML</strong> content</p>"),
			Status:       "active",
			CustomFields: make(map[string]interface{}),
		},
	}

	// Add date/time fields
	now := time.Now()
	dateOnly := polarion.DateOnly{Time: now}
	timeOnly := polarion.TimeOnly{Hour: 14, Minute: 30, Second: 0}
	dateTime := polarion.DateTime{Time: now}
	duration := polarion.Duration{Duration: 2 * time.Hour}

	wi.Attributes.CustomFields["dateOnlyField"] = dateOnly.String()
	wi.Attributes.CustomFields["timeonly"] = timeOnly.String()
	wi.Attributes.CustomFields["aDateTimeField"] = dateTime.String()
	wi.Attributes.CustomFields["oneDurationField"] = duration.String()

	// Add text content
	wi.Attributes.CustomFields["somtMultiLineString"] = map[string]interface{}{
		"type":  "text/plain",
		"value": "Multi-line\ntext\ncontent",
	}

	wi.Attributes.CustomFields["htmlField"] = map[string]interface{}{
		"type":  "text/html",
		"value": "<p>HTML <em>formatted</em> text</p>",
	}

	// Add table field
	tableField := map[string]interface{}{
		"keys": []string{"column1", "column2"},
		"rows": []map[string]interface{}{
			{
				"values": []map[string]interface{}{
					{"type": "text/html", "value": "Row 1, Col 1"},
					{"type": "text/html", "value": "Row 1, Col 2"},
				},
			},
			{
				"values": []map[string]interface{}{
					{"type": "text/html", "value": "Row 2, Col 1"},
					{"type": "text/html", "value": "Row 2, Col 2"},
				},
			},
		},
	}
	wi.Attributes.CustomFields["tableField"] = tableField

	err = project.WorkItems.Create(ctx, wi)
	if err != nil {
		t.Fatalf("Failed to create work item: %v", err)
	}

	workItemID := wi.ID
	t.Logf("Created work item with complex fields: %s", workItemID)

	// Fetch and verify
	fetched, err := project.WorkItems.Get(ctx, workItemID)
	if err != nil {
		t.Fatalf("Failed to get work item: %v", err)
	}

	cf := polarion.CustomFields(fetched.Attributes.CustomFields)

	// Test date field
	if val, ok := cf.GetDateOnly("dateOnlyField"); ok {
		t.Logf("Date field: %s", val.String())
	}

	// Test time field
	if val, ok := cf.GetTimeOnly("timeonly"); ok {
		t.Logf("Time field: %s", val.String())
	}

	// Test datetime field
	if val, ok := cf.GetDateTime("aDateTimeField"); ok {
		t.Logf("DateTime field: %s", val.String())
	}

	// Test duration field
	if val, ok := cf.GetDuration("oneDurationField"); ok {
		t.Logf("Duration field: %s", val.String())
	}

	// Test text content
	if val, ok := cf.GetText("somtMultiLineString"); ok {
		t.Logf("Text content: %s", val.Value)
	}

	// Test HTML content
	if val, ok := cf.GetText("htmlField"); ok {
		t.Logf("HTML content: %s", val.Value)
	}

	// Test table field
	if val, ok := cf.GetTable("tableField"); ok {
		t.Logf("Table field: %d rows, %d columns", len(val.Rows), len(val.Keys))
		for i, row := range val.Rows {
			t.Logf("  Row %d: %d values", i+1, len(row.Values))
		}
	}

	// Cleanup
	err = project.WorkItems.Delete(ctx, workItemID)
	if err != nil {
		t.Logf("Warning: Failed to delete work item (may not be supported): %v", err)
	} else {
		t.Logf("Successfully deleted work item: %s", workItemID)
	}
}

// TestWorkItemQuery tests querying work items
func TestWorkItemQuery(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	projectID := os.Getenv("POLARION_PROJECT")
	if projectID == "" {
		t.Skip("POLARION_PROJECT not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	project := client.Project(projectID)
	ctx := context.Background()

	// Query for features
	t.Run("QueryFeatures", func(t *testing.T) {
		result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
			Query:      "type:feature",
			PageSize:   10,
			PageNumber: 1,
		})
		if err != nil {
			t.Fatalf("Failed to query work items: %v", err)
		}

		t.Logf("Found %d features (total: %d)", len(result.Items), result.TotalCount)

		for i, item := range result.Items {
			if i >= 3 {
				break // Only log first 3
			}
			// Add nil check for debugging
			if item.Attributes == nil {
				t.Logf("  [%d] %s: <nil attributes>", i+1, item.ID)
				t.Errorf("Item %d (%s) has nil Attributes", i+1, item.ID)
				continue
			}
			t.Logf("  [%d] %s: %s", i+1, item.ID, item.Attributes.Title)
		}
	})

	// Query with custom field filter
	t.Run("QueryWithCustomField", func(t *testing.T) {
		result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
			Query:      "type:feature AND HAS_VALUE:someStringField",
			PageSize:   5,
			PageNumber: 1,
		})
		if err != nil {
			t.Fatalf("Failed to query work items: %v", err)
		}

		t.Logf("Found %d features with someStringField", len(result.Items))

		for _, item := range result.Items {
			// Add nil check for debugging
			if item.Attributes == nil {
				t.Logf("  %s: <nil attributes>", item.ID)
				t.Errorf("Item %s has nil Attributes", item.ID)
				continue
			}
			cf := polarion.CustomFields(item.Attributes.CustomFields)
			if val, ok := cf.GetString("someStringField"); ok {
				t.Logf("  %s: someStringField = %s", item.ID, val)
			}
		}
	})

	// Query all with pagination
	t.Run("QueryAllWithPagination", func(t *testing.T) {
		items, err := project.WorkItems.QueryAll(ctx, "type:feature", polarion.WithQueryPageSize(5))
		if err != nil {
			t.Fatalf("Failed to query all work items: %v", err)
		}

		t.Logf("Retrieved %d total features using pagination", len(items))
	})
}
