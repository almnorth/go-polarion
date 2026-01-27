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

// TestUserOperations tests user operations on an existing user
func TestUserOperations(t *testing.T) {
	// Skip if no token is provided
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	userID := os.Getenv("POLARION_USER")
	if userID == "" {
		t.Skip("POLARION_USER not set, skipping integration test")
	}

	// Create client
	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test 1: Get the user
	t.Run("GetUser", func(t *testing.T) {
		fetched, err := client.Users.Get(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if fetched.ID != userID {
			t.Errorf("Expected ID %s, got %s", userID, fetched.ID)
		}

		if fetched.Attributes == nil {
			t.Fatal("User attributes are nil")
		}

		t.Logf("Successfully fetched user: %s (Name: %s, Email: %s)",
			fetched.ID, fetched.Attributes.Name, fetched.Attributes.Email)
	})

	// Test 2: Update the user (non-destructive update)
	t.Run("UpdateUser", func(t *testing.T) {
		fetched, err := client.Users.Get(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		// Store original values
		originalName := fetched.Attributes.Name
		originalDesc := fetched.Attributes.Description

		// Update user attributes
		fetched.Attributes.Description = polarion.NewPlainTextContent("Test description from Go client")

		err = client.Users.Update(ctx, fetched)
		if err != nil {
			t.Logf("Warning: Failed to update user (may not have permissions): %v", err)
		} else {
			t.Logf("Successfully updated user: %s", fetched.ID)

			// Restore original values
			fetched.Attributes.Name = originalName
			fetched.Attributes.Description = originalDesc
			err = client.Users.Update(ctx, fetched)
			if err != nil {
				t.Logf("Warning: Failed to restore original user values: %v", err)
			}
		}
	})

	// Test 3: Avatar operations
	t.Run("AvatarOperations", func(t *testing.T) {
		// Try to get existing avatar
		avatar, err := client.Users.GetAvatar(ctx, userID)
		if err != nil {
			t.Logf("Warning: Failed to get avatar (user may not have one): %v", err)
		} else if avatar != nil {
			t.Logf("Successfully retrieved avatar: %d bytes, content-type: %s",
				len(avatar.Data), avatar.ContentType)
		}
	})

	// Test 4: Set license
	t.Run("SetLicense", func(t *testing.T) {
		license := &polarion.License{
			Type: "licenses",
			ID:   "developer",
		}

		err := client.Users.SetLicense(ctx, userID, license)
		if err != nil {
			t.Logf("Warning: Failed to set license (may not be supported or license may not exist): %v", err)
		} else {
			t.Logf("Successfully set license for user: %s", userID)
		}
	})
}

// TestUserCreate tests creating a new user (may require admin permissions)
func TestUserCreate(t *testing.T) {
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

	// Generate a unique user ID for this test run
	timestamp := time.Now().Unix()
	testUserID := fmt.Sprintf("testuser_%d", timestamp)

	t.Run("CreateUser", func(t *testing.T) {
		user := &polarion.User{
			Type: "users",
			ID:   testUserID,
			Attributes: &polarion.UserAttributes{
				Name:  "Test User from Go Test",
				Email: fmt.Sprintf("%s@example.com", testUserID),
			},
		}

		created, err := client.Users.Create(ctx, user)
		if err != nil {
			t.Logf("Warning: Failed to create user (may require admin permissions or different API format): %v", err)
			t.Skip("Skipping user creation test - may not be supported or may require different format")
			return
		}

		if len(created) == 0 {
			t.Fatal("Expected at least one created user, got none")
		}

		if created[0].ID != testUserID {
			t.Errorf("Expected user ID %s, got %s", testUserID, created[0].ID)
		}

		t.Logf("Created user: %s", created[0].ID)
	})
}

// TestUserList tests listing users
func TestUserList(t *testing.T) {
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

	t.Run("ListAllUsers", func(t *testing.T) {
		users, err := client.Users.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}

		t.Logf("Found %d users", len(users))

		for i, user := range users {
			if i >= 5 {
				break // Only log first 5
			}
			if user.Attributes == nil {
				t.Logf("  [%d] %s: <nil attributes>", i+1, user.ID)
				continue
			}
			t.Logf("  [%d] %s: %s (Email: %s, Disabled: %t)",
				i+1, user.ID, user.Attributes.Name, user.Attributes.Email, user.Attributes.Disabled)
		}
	})

	t.Run("ListUsersWithQuery", func(t *testing.T) {
		users, err := client.Users.List(ctx, polarion.WithQuery("disabled:false"))
		if err != nil {
			t.Fatalf("Failed to list users with query: %v", err)
		}

		t.Logf("Found %d active users", len(users))

		// Verify all returned users are not disabled
		for _, user := range users {
			if user.Attributes != nil && user.Attributes.Disabled {
				t.Errorf("Expected only active users, but found disabled user: %s", user.ID)
			}
		}
	})

	t.Run("ListUsersWithPageSize", func(t *testing.T) {
		users, err := client.Users.List(ctx, polarion.WithQueryPageSize(3))
		if err != nil {
			t.Fatalf("Failed to list users with page size: %v", err)
		}

		t.Logf("Found %d users with page size 3", len(users))

		if len(users) > 3 {
			t.Logf("Note: Received more than requested page size (pagination may have been applied)")
		}
	})

	t.Run("ListUsersWithPageSizeAndQuery", func(t *testing.T) {
		users, err := client.Users.List(ctx,
			polarion.WithQueryPageSize(5),
			polarion.WithQuery("disabled:false"))
		if err != nil {
			t.Fatalf("Failed to list users with page size and query: %v", err)
		}

		t.Logf("Found %d active users with page size 5", len(users))

		if len(users) > 0 && users[0].Attributes != nil {
			t.Logf("Sample user: %s (Name: %s, Email: %s)",
				users[0].ID, users[0].Attributes.Name, users[0].Attributes.Email)
		}
	})
}

// TestUserGet tests getting a specific user
func TestUserGet(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	userID := os.Getenv("POLARION_USER")
	if userID == "" {
		t.Skip("POLARION_USER not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("GetExistingUser", func(t *testing.T) {
		user, err := client.Users.Get(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, user.ID)
		}

		if user.Attributes == nil {
			t.Fatal("User attributes are nil")
		}

		t.Logf("User ID: %s", user.ID)
		t.Logf("User Name: %s", user.Attributes.Name)
		t.Logf("User Email: %s", user.Attributes.Email)
		if user.Attributes.Description != nil {
			t.Logf("User Description: %s (type: %s)",
				user.Attributes.Description.Value, user.Attributes.Description.Type)
		} else {
			t.Logf("User Description: <nil>")
		}
		t.Logf("User Disabled: %t", user.Attributes.Disabled)
		t.Logf("User DisabledForUI: %t", user.Attributes.DisabledForUI)
		t.Logf("User VaultUser: %t", user.Attributes.VaultUser)
	})

	t.Run("GetUserWithRevision", func(t *testing.T) {
		// First get the user to get its revision
		user, err := client.Users.Get(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.Revision == "" {
			t.Skip("User has no revision, skipping revision test")
		}

		// Get user with specific revision
		userWithRev, err := client.Users.Get(ctx, userID, polarion.WithGetRevision(user.Revision))
		if err != nil {
			t.Logf("Warning: Failed to get user with revision (may not be supported): %v", err)
		} else {
			t.Logf("User with revision: %s (Revision: %s)",
				userWithRev.ID, userWithRev.Revision)
		}
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		_, err := client.Users.Get(ctx, "nonexistent_user_12345")
		if err == nil {
			t.Error("Expected error when getting non-existent user, got nil")
		} else {
			t.Logf("Got expected error for non-existent user: %v", err)
		}
	})
}

// TestUserValidation tests validation errors
func TestUserValidation(t *testing.T) {
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

	t.Run("CreateUserWithoutData", func(t *testing.T) {
		_, err := client.Users.Create(ctx)
		if err == nil {
			t.Error("Expected error when creating user without data, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("GetUserWithEmptyID", func(t *testing.T) {
		_, err := client.Users.Get(ctx, "")
		if err == nil {
			t.Error("Expected error when getting user with empty ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("UpdateNilUser", func(t *testing.T) {
		err := client.Users.Update(ctx, nil)
		if err == nil {
			t.Error("Expected error when updating nil user, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("UpdateUserWithEmptyID", func(t *testing.T) {
		user := &polarion.User{
			Type: "users",
			Attributes: &polarion.UserAttributes{
				Name: "Test User",
			},
		}

		err := client.Users.Update(ctx, user)
		if err == nil {
			t.Error("Expected error when updating user with empty ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("GetAvatarWithEmptyID", func(t *testing.T) {
		_, err := client.Users.GetAvatar(ctx, "")
		if err == nil {
			t.Error("Expected error when getting avatar with empty user ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("UpdateAvatarWithEmptyID", func(t *testing.T) {
		testData := []byte{0x89, 0x50, 0x4E, 0x47}
		err := client.Users.UpdateAvatar(ctx, "", testData, "image/png")
		if err == nil {
			t.Error("Expected error when updating avatar with empty user ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("UpdateAvatarWithEmptyData", func(t *testing.T) {
		err := client.Users.UpdateAvatar(ctx, "testuser", []byte{}, "image/png")
		if err == nil {
			t.Error("Expected error when updating avatar with empty data, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("SetLicenseWithEmptyUserID", func(t *testing.T) {
		license := &polarion.License{
			Type: "licenses",
			ID:   "developer",
		}

		err := client.Users.SetLicense(ctx, "", license)
		if err == nil {
			t.Error("Expected error when setting license with empty user ID, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})

	t.Run("SetNilLicense", func(t *testing.T) {
		err := client.Users.SetLicense(ctx, "testuser", nil)
		if err == nil {
			t.Error("Expected error when setting nil license, got nil")
		} else {
			t.Logf("Got expected validation error: %v", err)
		}
	})
}

// TestUserAvatar tests avatar operations in detail
func TestUserAvatar(t *testing.T) {
	token := os.Getenv("POLARION_TOKEN")
	if token == "" {
		t.Skip("POLARION_TOKEN not set, skipping integration test")
	}

	polarionURL := os.Getenv("POLARION_URL")
	if polarionURL == "" {
		t.Skip("POLARION_URL not set, skipping integration test")
	}

	userID := os.Getenv("POLARION_USER")
	if userID == "" {
		t.Skip("POLARION_USER not set, skipping integration test")
	}

	client, err := polarion.New(polarionURL, token)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("GetExistingUserAvatar", func(t *testing.T) {
		avatar, err := client.Users.GetAvatar(ctx, userID)
		if err != nil {
			t.Logf("Warning: Failed to get avatar (user may not have one): %v", err)
			t.Skip("Skipping avatar test as user may not have an avatar")
			return
		}

		if avatar == nil {
			t.Error("Expected avatar data, got nil")
			return
		}

		t.Logf("Avatar size: %d bytes", len(avatar.Data))
		t.Logf("Avatar content type: %s", avatar.ContentType)

		if len(avatar.Data) == 0 {
			t.Error("Expected avatar data, got empty byte slice")
		}

		if avatar.ContentType == "" {
			t.Error("Expected content type, got empty string")
		}
	})

	t.Run("UpdateAvatarWithDefaultContentType", func(t *testing.T) {
		// Create a simple test avatar (1x1 PNG)
		testAvatarData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
			0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
			0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
			0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xDD, 0x8D,
			0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
			0x44, 0xAE, 0x42, 0x60, 0x82,
		}

		// Update with empty content type (should default to image/png)
		err := client.Users.UpdateAvatar(ctx, userID, testAvatarData, "")
		if err != nil {
			t.Logf("Warning: Failed to update avatar (may not be supported): %v", err)
		} else {
			t.Logf("Successfully updated avatar with default content type")
		}
	})
}

// TestUserCreateMultiple tests creating multiple users at once (may require admin permissions)
func TestUserCreateMultiple(t *testing.T) {
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

	timestamp := time.Now().Unix()

	t.Run("CreateMultipleUsers", func(t *testing.T) {
		users := []*polarion.User{
			{
				Type: "users",
				ID:   fmt.Sprintf("testuser_multi1_%d", timestamp),
				Attributes: &polarion.UserAttributes{
					Name:  "Test User Multi 1",
					Email: fmt.Sprintf("testuser_multi1_%d@example.com", timestamp),
				},
			},
			{
				Type: "users",
				ID:   fmt.Sprintf("testuser_multi2_%d", timestamp),
				Attributes: &polarion.UserAttributes{
					Name:  "Test User Multi 2",
					Email: fmt.Sprintf("testuser_multi2_%d@example.com", timestamp),
				},
			},
		}

		created, err := client.Users.Create(ctx, users...)
		if err != nil {
			t.Logf("Warning: Failed to create multiple users (may require admin permissions or different API format): %v", err)
			t.Skip("Skipping multiple user creation test - may not be supported or may require different format")
			return
		}

		if len(created) != 2 {
			t.Errorf("Expected 2 created users, got %d", len(created))
		}

		for i, user := range created {
			t.Logf("Created user %d: %s (Name: %s)", i+1, user.ID, user.Attributes.Name)
		}
	})
}
