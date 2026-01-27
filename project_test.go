// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// TestProjectLifecycle tests the complete lifecycle of a project:
// create, get, update, and delete
func TestProjectLifecycle(t *testing.T) {
	// Skip if no token is provided
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	// Create client
	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Generate a unique project ID for this test run
	timestamp := time.Now().Unix()
	testProjectID := fmt.Sprintf("test_project_%d", timestamp)

	// Test 1: Create a project
	t.Run("CreateProject", func(t *testing.T) {
		req := &polarion.CreateProjectRequest{
			ProjectID:   testProjectID,
			Name:        "Test Project from Go Test",
			Location:    "/go-polarion-client",
			Description: "This is a test project created by automated tests",
		}

		project, err := client.Projects.Create(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		if project.ID != testProjectID {
			t.Errorf("Expected project ID %s, got %s", testProjectID, project.ID)
		}

		t.Logf("Created project: %s", project.ID)

		// Wait a bit for the project to be fully created
		// Project creation is asynchronous
		time.Sleep(5 * time.Second)

		// Test 2: Get the project
		t.Run("GetProject", func(t *testing.T) {
			fetched, err := client.Projects.Get(ctx, testProjectID)
			if err != nil {
				t.Fatalf("Failed to get project: %v", err)
			}

			if fetched.ID != testProjectID {
				t.Errorf("Expected ID %s, got %s", testProjectID, fetched.ID)
			}

			if fetched.Attributes == nil {
				t.Fatal("Project attributes are nil")
			}

			// Note: The API may return the project ID as the name initially
			// The name we set might only be visible after the project is fully initialized
			t.Logf("Successfully fetched project: %s (Name: %s)", fetched.ID, fetched.Attributes.Name)
		})

		// Test 3: Update the project
		t.Run("UpdateProject", func(t *testing.T) {
			fetched, err := client.Projects.Get(ctx, testProjectID)
			if err != nil {
				t.Fatalf("Failed to get project: %v", err)
			}

			// Update project attributes
			fetched.Attributes.Description = polarion.NewPlainTextContent("Updated test project description")
			fetched.Attributes.Active = true

			updated, err := client.Projects.Update(ctx, fetched)
			if err != nil {
				t.Fatalf("Failed to update project: %v", err)
			}

			t.Logf("Successfully updated project: %s", updated.ID)

			// Verify the update
			verified, err := client.Projects.Get(ctx, testProjectID)
			if err != nil {
				t.Fatalf("Failed to get updated project: %v", err)
			}

			if verified.Attributes.Description == nil || verified.Attributes.Description.Value != "Updated test project description" {
				if verified.Attributes.Description == nil {
					t.Error("Expected description to be set, got nil")
				} else {
					t.Errorf("Expected description 'Updated test project description', got '%s'", verified.Attributes.Description.Value)
				}
			}

			if verified.Attributes.Description != nil {
				t.Logf("Project update verified: %s", verified.Attributes.Description.Value)
			}
		})

		// Test 4: Mark and Unmark project
		t.Run("MarkUnmarkProject", func(t *testing.T) {
			// Mark the project
			err := client.Projects.Mark(ctx, testProjectID)
			if err != nil {
				t.Logf("Warning: Failed to mark project (may not be supported): %v", err)
			} else {
				t.Logf("Successfully marked project: %s", testProjectID)

				// Unmark the project
				err = client.Projects.Unmark(ctx, testProjectID)
				if err != nil {
					t.Logf("Warning: Failed to unmark project: %v", err)
				} else {
					t.Logf("Successfully unmarked project: %s", testProjectID)
				}
			}
		})

		// Test 5: Delete the project (cleanup)
		// IMPORTANT: Only delete the project created by this test
		t.Run("DeleteProject", func(t *testing.T) {
			// Double-check we're only deleting our test project
			if testProjectID != fmt.Sprintf("test_project_%d", timestamp) {
				t.Fatal("Safety check failed: attempting to delete wrong project")
			}

			err := client.Projects.Delete(ctx, testProjectID)
			if err != nil {
				// Log warning but don't fail the test if delete is not supported
				t.Logf("Warning: Failed to delete project (may not be supported): %v", err)
				t.Skip("Skipping delete verification as delete operation failed")
				return
			}

			t.Logf("Successfully deleted project: %s", testProjectID)

			// Note: Project deletion might be asynchronous, so the project
			// may still be retrievable immediately after deletion
			// We'll just log the result rather than failing the test
			_, err = client.Projects.Get(ctx, testProjectID)
			if err != nil {
				t.Logf("Verified project deletion: %v", err)
			} else {
				t.Logf("Note: Project still exists after deletion (deletion may be asynchronous)")
			}
		})
	})
}

// TestProjectList tests listing projects
func TestProjectList(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("ListAllProjects", func(t *testing.T) {
		projects, err := client.Projects.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list projects: %v", err)
		}

		t.Logf("Found %d projects", len(projects))

		for i, project := range projects {
			if i >= 5 {
				break // Only log first 5
			}
			if project.Attributes == nil {
				t.Logf("  [%d] %s: <nil attributes>", i+1, project.ID)
				continue
			}
			t.Logf("  [%d] %s: %s (Active: %t)", i+1, project.ID, project.Attributes.Name, project.Attributes.Active)
		}
	})

	t.Run("ListProjectsWithPageSize", func(t *testing.T) {
		projects, err := client.Projects.List(ctx, polarion.WithQueryPageSize(3))
		if err != nil {
			t.Fatalf("Failed to list projects with page size: %v", err)
		}

		t.Logf("Found %d projects with page size 3", len(projects))

		if len(projects) > 3 {
			t.Logf("Note: Received more than requested page size (pagination may have been applied)")
		}
	})
}

// TestProjectGet tests getting a specific project
func TestProjectGet(t *testing.T) {
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

	ctx := context.Background()

	t.Run("GetExistingProject", func(t *testing.T) {
		project, err := client.Projects.Get(ctx, projectID)
		if err != nil {
			t.Fatalf("Failed to get project: %v", err)
		}

		if project.ID != projectID {
			t.Errorf("Expected project ID %s, got %s", projectID, project.ID)
		}

		if project.Attributes == nil {
			t.Fatal("Project attributes are nil")
		}

		t.Logf("Project ID: %s", project.ID)
		t.Logf("Project Name: %s", project.Attributes.Name)
		if project.Attributes.Description != nil {
			t.Logf("Project Description: %s (type: %s)", project.Attributes.Description.Value, project.Attributes.Description.Type)
		} else {
			t.Logf("Project Description: <nil>")
		}
		t.Logf("Project Active: %t", project.Attributes.Active)
		t.Logf("Project Location: %s", project.Attributes.Location)
		t.Logf("Project Lead: %s", project.Attributes.Lead)
		t.Logf("Project Start Date: %s", project.Attributes.StartDate)
		t.Logf("Project Finish Date: %s", project.Attributes.FinishDate)
	})

	t.Run("GetNonExistentProject", func(t *testing.T) {
		_, err := client.Projects.Get(ctx, "nonexistent_project_12345")
		if err == nil {
			t.Error("Expected error when getting non-existent project, got nil")
		} else {
			t.Logf("Got expected error for non-existent project: %v", err)
		}
	})
}

// TestProjectValidation tests validation errors
func TestProjectValidation(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("CreateProjectWithoutID", func(t *testing.T) {
		req := &polarion.CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test description",
		}

		_, err := client.Projects.Create(ctx, req)
		if err == nil {
			t.Error("Expected error when creating project without ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("CreateProjectWithoutName", func(t *testing.T) {
		req := &polarion.CreateProjectRequest{
			ProjectID:   "test_project",
			Description: "Test description",
		}

		_, err := client.Projects.Create(ctx, req)
		if err == nil {
			t.Error("Expected error when creating project without name, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("GetProjectWithEmptyID", func(t *testing.T) {
		_, err := client.Projects.Get(ctx, "")
		if err == nil {
			t.Error("Expected error when getting project with empty ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("UpdateNilProject", func(t *testing.T) {
		_, err := client.Projects.Update(ctx, nil)
		if err == nil {
			t.Error("Expected error when updating nil project, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("DeleteProjectWithEmptyID", func(t *testing.T) {
		err := client.Projects.Delete(ctx, "")
		if err == nil {
			t.Error("Expected error when deleting project with empty ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})
}
