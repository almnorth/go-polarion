// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/almnorth/go-polarion"
)

// Example demonstrates basic usage of the Polarion client.
func Example() {
	// Create client
	client, err := polarion.New(
		"https://polarion.example.com/rest/v1",
		"your-bearer-token",
		polarion.WithBatchSize(50),
		polarion.WithPageSize(100),
		polarion.WithTimeout(60*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get project client
	project := client.Project("my-project")

	ctx := context.Background()

	// Get a work item
	wi, err := project.WorkItems.Get(ctx, "WI-123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Work Item: %s - %s\n", wi.ID, wi.Attributes.Title)
}

// Example_createWorkItem demonstrates creating a work item.
func Example_createWorkItem() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	wi := &polarion.WorkItem{
		Type: "workitems",
		Attributes: &polarion.WorkItemAttributes{
			Title:  "New Security Requirement",
			Status: "draft",
			Description: polarion.NewHTMLContent(
				"<p>All user data must be encrypted at rest</p>",
			),
		},
	}

	err := project.WorkItems.Create(ctx, wi)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created work item: %s\n", wi.ID)
}

// Example_queryWorkItems demonstrates querying work items.
func Example_queryWorkItems() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	// Query all with automatic pagination
	items, err := project.WorkItems.QueryAll(
		ctx,
		"type:requirement AND status:open",
		polarion.WithFields(polarion.FieldsBasic),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d work items\n", len(items))
}

// Example_updateWorkItem demonstrates updating a work item.
func Example_updateWorkItem() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	// Get work item
	wi, err := project.WorkItems.Get(ctx, "WI-123")
	if err != nil {
		log.Fatal(err)
	}

	// Modify fields
	wi.Attributes.Status = "approved"
	wi.Attributes.SetCustomField("priority", "high")

	// Update
	err = project.WorkItems.Update(ctx, wi)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Work item updated")
}

// Example_errorHandling demonstrates error handling.
func Example_errorHandling() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	wi, err := project.WorkItems.Get(ctx, "WI-999")
	if err != nil {
		// Check for specific error types
		if polarion.IsNotFound(err) {
			fmt.Println("Work item not found")
			return
		}

		var apiErr *polarion.APIError
		if polarion.AsAPIError(err, &apiErr) {
			fmt.Printf("API Error: Status=%d\n", apiErr.StatusCode)
			return
		}

		log.Fatal(err)
	}
	fmt.Printf("Found: %s\n", wi.ID)
}

// Example_enumerations demonstrates working with enumerations.
func Example_enumerations() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	// Get a specific enumeration
	enum, err := project.Enumerations.Get(ctx, "workitem", "status", "requirement")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Enumeration: %s\n", enum.ID)
	for _, option := range enum.Attributes.Options {
		fmt.Printf("  Option: %s - %s\n", option.ID, option.Name)
	}

	// List all enumerations
	enums, err := project.Enumerations.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d enumerations\n", len(enums))
}

// Example_workItemLinks demonstrates working with work item links.
func Example_workItemLinks() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	// List all links for a work item
	links, err := project.WorkItemLinks.List(ctx, "WI-123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d links\n", len(links))
	for _, link := range links {
		fmt.Printf("  Link: %s (suspect: %v)\n", link.Data.Role, link.Data.Suspect)
	}

	// Create a link between work items
	link := &polarion.WorkItemLink{
		Type: "linkedworkitems",
		Data: &polarion.WorkItemLinkAttributes{
			Role:    "relates_to",
			Suspect: false,
		},
	}
	err = project.WorkItemLinks.Create(ctx, "WI-123", link)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Link created")
}

// Example_workItemTypes demonstrates working with work item type definitions.
func Example_workItemTypes() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	// Get a specific work item type
	wiType, err := project.WorkItemTypes.Get(ctx, "requirement")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Type: %s - %s\n", wiType.ID, wiType.Attributes.Name)
	fmt.Printf("Icon: %s\n", wiType.Attributes.Icon)

	// Get field definitions for a type
	fields, err := project.WorkItemTypes.GetFields(ctx, "requirement")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d fields\n", len(fields))
	for _, field := range fields {
		fmt.Printf("  Field: %s (%s) - Required: %v\n",
			field.ID, field.Type, field.Required)
	}

	// Get a specific field definition
	field, err := project.WorkItemTypes.GetFieldByID(ctx, "requirement", "status")
	if err != nil {
		log.Fatal(err)
	}

	if field.EnumerationID != "" {
		fmt.Printf("Status field uses enumeration: %s\n", field.EnumerationID)
	}
}

// Example_users demonstrates working with users.
func Example_users() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	ctx := context.Background()

	// Get a specific user
	user, err := client.Users.Get(ctx, "user123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User: %s (%s)\n", user.Attributes.Name, user.Attributes.Email)

	// List all users
	users, err := client.Users.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d users\n", len(users))

	// List users with query
	activeUsers, err := client.Users.List(ctx, polarion.WithQuery("disabled:false"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d active users\n", len(activeUsers))
}

// Example_userGroups demonstrates working with user groups.
func Example_userGroups() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	ctx := context.Background()

	// Get a specific user group
	group, err := client.UserGroups.Get(ctx, "developers")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Group: %s\n", group.Attributes.Name)

	// List all user groups
	groups, err := client.UserGroups.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d user groups\n", len(groups))
}

// Example_workItemComments demonstrates working with work item comments.
func Example_workItemComments() {
	client, _ := polarion.New("https://polarion.example.com/rest/v1", "token")
	project := client.Project("my-project")
	ctx := context.Background()

	// List all comments for a work item
	comments, err := project.WorkItemComments.List(ctx, "WI-123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d comments\n", len(comments))

	// Create a new comment
	newComment := &polarion.WorkItemComment{
		Type: "workitem_comments",
		Attributes: &polarion.WorkItemCommentAttributes{
			Text: polarion.NewHTMLContent("<p>This is a comment</p>"),
		},
	}
	created, err := project.WorkItemComments.Create(ctx, "WI-123", newComment)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created comment: %s\n", created[0].ID)

	// Get a specific comment
	comment, err := project.WorkItemComments.Get(ctx, "WI-123", created[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Comment text: %s\n", comment.Attributes.Text.Value)

	// Update a comment
	comment.Attributes.Text = polarion.NewHTMLContent("<p>Updated comment</p>")
	err = project.WorkItemComments.Update(ctx, "WI-123", comment)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Comment updated")
}
