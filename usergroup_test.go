// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion_test

import (
	"context"
	"os"
	"testing"

	polarion "github.com/almnorth/go-polarion"
)

// TestUserGroupGet tests getting a specific user group
// Note: The Polarion API only provides GET /usergroups/{groupId}, not a LIST endpoint
func TestUserGroupGet(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	groupID := os.Getenv("POLARION_USERGROUP")
	if groupID == "" {
		t.Skip("POLARION_USERGROUP not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("GetExistingUserGroup", func(t *testing.T) {
		group, err := client.UserGroups.Get(ctx, groupID)
		if err != nil {
			t.Fatalf("Failed to get user group: %v", err)
		}

		if group.ID != groupID {
			t.Errorf("Expected group ID %s, got %s", groupID, group.ID)
		}

		if group.Attributes == nil {
			t.Fatal("User group attributes are nil")
		}

		t.Logf("User Group ID: %s", group.ID)
		t.Logf("User Group Name: %s", group.Attributes.Name)
		if group.Attributes.Description != nil {
			t.Logf("User Group Description: %s (type: %s)",
				group.Attributes.Description.Value, group.Attributes.Description.Type)
		} else {
			t.Logf("User Group Description: <nil>")
		}

		// Log relationships if present
		if group.Relationships != nil {
			if group.Relationships.Users != nil {
				t.Logf("Has Users relationship")
			}
			if group.Relationships.GlobalRoles != nil {
				t.Logf("Has GlobalRoles relationship")
			}
			if group.Relationships.ProjectRoles != nil {
				t.Logf("Has ProjectRoles relationship")
			}
		}
	})

	t.Run("GetUserGroupWithRevision", func(t *testing.T) {
		// First get the group to get its revision
		group, err := client.UserGroups.Get(ctx, groupID)
		if err != nil {
			t.Fatalf("Failed to get user group: %v", err)
		}

		if group.Revision == "" {
			t.Skip("User group has no revision, skipping revision test")
		}

		// Get group with specific revision
		groupWithRev, err := client.UserGroups.Get(ctx, groupID, polarion.WithGetRevision(group.Revision))
		if err != nil {
			t.Logf("Warning: Failed to get user group with revision (may not be supported): %v", err)
		} else {
			t.Logf("User group with revision: %s (Revision: %s)",
				groupWithRev.ID, groupWithRev.Revision)
		}
	})

	t.Run("GetNonExistentUserGroup", func(t *testing.T) {
		_, err := client.UserGroups.Get(ctx, "nonexistent_group_12345")
		if err == nil {
			t.Error("Expected error when getting non-existent user group, got nil")
		} else {
			t.Logf("Got expected error for non-existent user group: %v", err)
		}
	})
}

// TestUserGroupValidation tests validation errors
func TestUserGroupValidation(t *testing.T) {
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

	t.Run("GetUserGroupWithEmptyID", func(t *testing.T) {
		_, err := client.UserGroups.Get(ctx, "")
		if err == nil {
			t.Error("Expected error when getting user group with empty ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})
}

// TestUserGroupAttributes tests user group attribute handling
func TestUserGroupAttributes(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	groupID := os.Getenv("POLARION_USERGROUP")
	if groupID == "" {
		t.Skip("POLARION_USERGROUP not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("VerifyGroupStructure", func(t *testing.T) {
		group, err := client.UserGroups.Get(ctx, groupID)
		if err != nil {
			t.Fatalf("Failed to get user group: %v", err)
		}

		// Verify basic structure
		if group.Type != "usergroups" {
			t.Errorf("Expected type 'usergroups', got '%s'", group.Type)
		}

		if group.ID == "" {
			t.Error("Expected non-empty ID")
		}

		if group.Attributes == nil {
			t.Fatal("Expected non-nil attributes")
		}

		// Name should always be present
		if group.Attributes.Name == "" {
			t.Error("Expected non-empty name")
		}

		// Description may or may not be present
		if group.Attributes.Description != nil {
			if group.Attributes.Description.Type == "" {
				t.Error("Expected description type to be set")
			}
		}

		// Links should be present
		if group.Links != nil && group.Links.Self != "" {
			t.Logf("Self link: %s", group.Links.Self)
		}

		t.Logf("Group structure verified successfully")
	})
}

// TestUserGroupRelationships tests user group relationship handling
func TestUserGroupRelationships(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	groupID := os.Getenv("POLARION_USERGROUP")
	if groupID == "" {
		t.Skip("POLARION_USERGROUP not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("CheckRelationships", func(t *testing.T) {
		group, err := client.UserGroups.Get(ctx, groupID)
		if err != nil {
			t.Fatalf("Failed to get user group: %v", err)
		}

		if group.Relationships == nil {
			t.Log("No relationships present (this is acceptable)")
			return
		}

		// Check each relationship type
		if group.Relationships.Users != nil {
			t.Logf("Users relationship present")
			if group.Relationships.Users.Links != nil {
				t.Logf("  Related link: %s", group.Relationships.Users.Links.Related)
			}
		}

		if group.Relationships.GlobalRoles != nil {
			t.Logf("GlobalRoles relationship present")
			if group.Relationships.GlobalRoles.Links != nil {
				t.Logf("  Related link: %s", group.Relationships.GlobalRoles.Links.Related)
			}
		}

		if group.Relationships.ProjectRoles != nil {
			t.Logf("ProjectRoles relationship present")
			if group.Relationships.ProjectRoles.Links != nil {
				t.Logf("  Related link: %s", group.Relationships.ProjectRoles.Links.Related)
			}
		}
	})
}
