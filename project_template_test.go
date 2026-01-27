// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion_test

import (
	"context"
	"os"
	"testing"

	polarion "github.com/almnorth/go-polarion"
)

// TestProjectTemplateList tests listing project templates
func TestProjectTemplateList(t *testing.T) {
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

	t.Run("ListAllTemplates", func(t *testing.T) {
		templates, err := client.ProjectTemplates.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list project templates: %v", err)
		}

		t.Logf("Found %d project templates", len(templates))

		if len(templates) == 0 {
			t.Log("No project templates found (this may be expected in some Polarion instances)")
			return
		}

		// Log details of first few templates
		for i, template := range templates {
			if i >= 5 {
				break // Only log first 5
			}

			if template.Attributes == nil {
				t.Logf("  [%d] %s: <nil attributes>", i+1, template.ID)
				continue
			}

			t.Logf("  [%d] %s: %s", i+1, template.ID, template.Attributes.Name)
			if template.Attributes.Description != "" {
				t.Logf("      Description: %s", template.Attributes.Description)
			}
			if template.Attributes.TemplateID != "" {
				t.Logf("      Template ID: %s", template.Attributes.TemplateID)
			}
		}

		// Verify structure of first template
		if len(templates) > 0 {
			template := templates[0]

			if template.Type != "projecttemplates" {
				t.Errorf("Expected type 'projecttemplates', got '%s'", template.Type)
			}

			if template.ID == "" {
				t.Error("Template ID is empty")
			}

			// Attributes may be nil for some templates
			if template.Attributes != nil {
				t.Logf("First template has attributes: Name=%s", template.Attributes.Name)
			}
		}
	})

	t.Run("ListTemplatesWithPageSize", func(t *testing.T) {
		templates, err := client.ProjectTemplates.List(ctx, polarion.WithQueryPageSize(2))
		if err != nil {
			t.Fatalf("Failed to list project templates with page size: %v", err)
		}

		t.Logf("Found %d project templates with page size 2", len(templates))

		if len(templates) > 2 {
			t.Logf("Note: Received more than requested page size (pagination may have been applied)")
		}
	})

	t.Run("ListTemplatesWithFields", func(t *testing.T) {
		// Request only specific fields using FieldSelector
		// Note: FieldSelector is primarily for work items, but we can test it here
		fields := polarion.NewFieldSelector()

		templates, err := client.ProjectTemplates.List(ctx, polarion.WithFields(fields))
		if err != nil {
			t.Fatalf("Failed to list project templates with field selection: %v", err)
		}

		t.Logf("Found %d project templates with field selection", len(templates))

		if len(templates) > 0 {
			template := templates[0]
			t.Logf("First template: ID=%s", template.ID)
			if template.Attributes != nil {
				t.Logf("  Name: %s", template.Attributes.Name)
				t.Logf("  Description: %s", template.Attributes.Description)
			}
		}
	})
}

// TestProjectTemplateStructure tests the structure of project templates
func TestProjectTemplateStructure(t *testing.T) {
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

	templates, err := client.ProjectTemplates.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list project templates: %v", err)
	}

	if len(templates) == 0 {
		t.Skip("No project templates available for structure testing")
	}

	t.Run("VerifyTemplateFields", func(t *testing.T) {
		for i, template := range templates {
			// Verify required fields
			if template.Type == "" {
				t.Errorf("Template %d: Type is empty", i)
			}

			if template.ID == "" {
				t.Errorf("Template %d: ID is empty", i)
			}

			// Attributes may be nil, but if present, verify structure
			if template.Attributes != nil {
				// Name is typically present but not strictly required
				if template.Attributes.Name == "" {
					t.Logf("Template %d (%s): Name is empty", i, template.ID)
				}
			}

			// Only check first few templates to avoid excessive output
			if i >= 2 {
				break
			}
		}
	})

	t.Run("VerifyTemplateLinks", func(t *testing.T) {
		for i, template := range templates {
			if template.Links != nil && template.Links.Self != "" {
				t.Logf("Template %d (%s) has self link: %s", i, template.ID, template.Links.Self)
			}

			// Only check first few templates
			if i >= 2 {
				break
			}
		}
	})
}

// TestProjectTemplateUsage tests using templates in project creation
func TestProjectTemplateUsage(t *testing.T) {
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

	// Get available templates
	templates, err := client.ProjectTemplates.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list project templates: %v", err)
	}

	if len(templates) == 0 {
		t.Skip("No project templates available for usage testing")
	}

	t.Run("VerifyTemplateIDsForCreation", func(t *testing.T) {
		// Verify that templates have IDs that can be used for project creation
		for i, template := range templates {
			if template.ID == "" {
				t.Errorf("Template %d has empty ID, cannot be used for project creation", i)
			} else {
				t.Logf("Template %d: ID '%s' can be used for project creation", i, template.ID)
			}

			// Only check first few templates
			if i >= 2 {
				break
			}
		}
	})

	t.Run("DocumentTemplateUsage", func(t *testing.T) {
		// Document how to use templates in project creation
		if len(templates) > 0 {
			template := templates[0]
			t.Logf("Example: To create a project with template '%s':", template.ID)
			t.Logf("  req := &polarion.CreateProjectRequest{")
			t.Logf("      ProjectID:   \"myproject\",")
			t.Logf("      Name:        \"My Project\",")
			t.Logf("      Location:    \"/default\",")
			t.Logf("      TemplateID:  \"%s\",", template.ID)
			t.Logf("  }")
			t.Logf("  project, err := client.Projects.Create(ctx, req)")
		}
	})
}
