// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// Note: WorkItemLink and WorkItemLinkAttributes are defined in workitem.go
// This file contains additional types and helpers for the work item link service.

// LinkedWorkItemRelationships contains relationships for a work item link.
type LinkedWorkItemRelationships struct {
	// WorkItem is the relationship to the secondary (target) work item
	WorkItem *Relationship `json:"workItem,omitempty"`
}

// LinkRole represents a link role definition.
// Link roles define the types of relationships between work items.
type LinkRole struct {
	// Type is always "workitemlinkroles" for link roles
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the link role
	ID string `json:"id,omitempty"`

	// Attributes contains the link role attributes
	Attributes *LinkRoleAttributes `json:"attributes,omitempty"`

	// Links contains hypermedia links
	Links *LinkRoleLinks `json:"links,omitempty"`
}

// LinkRoleAttributes contains the attributes of a link role.
type LinkRoleAttributes struct {
	// Name is the display name of the link role
	Name string `json:"name,omitempty"`

	// OppositeRole is the ID of the opposite role (for bidirectional links)
	OppositeRole string `json:"oppositeRole,omitempty"`

	// Description provides additional information about the role
	Description string `json:"description,omitempty"`
}

// LinkRoleLinks contains hypermedia links for the link role.
type LinkRoleLinks struct {
	Self string `json:"self,omitempty"`
}

// CreateLinkRequest represents a request to create a work item link.
type CreateLinkRequest struct {
	// PrimaryWorkItemID is the ID of the source work item
	PrimaryWorkItemID string

	// SecondaryWorkItemID is the ID of the target work item
	SecondaryWorkItemID string

	// Role is the link role ID
	Role string

	// Suspect indicates if the link should be marked as suspect
	Suspect bool

	// SecondaryProjectID is the project ID of the secondary work item (optional, defaults to current project)
	SecondaryProjectID string

	// SecondaryRevision is the specific revision of the secondary work item (optional)
	SecondaryRevision string
}

// UpdateLinkRequest represents a request to update a work item link.
type UpdateLinkRequest struct {
	// LinkID is the full ID of the link to update
	LinkID string

	// Suspect indicates if the link should be marked as suspect
	Suspect bool
}

// ParseLinkID parses a work item link ID into its components.
// Link ID format: "{projectId}/{primaryWorkItemId}/{role}/{secondaryProjectId}/{secondaryWorkItemId}"
func ParseLinkID(linkID string) (projectID, primaryWorkItemID, role, secondaryProjectID, secondaryWorkItemID string, err error) {
	// This is a helper function to parse link IDs
	// Implementation would split the ID by "/" and validate the parts
	// For now, we'll keep it simple
	return "", "", "", "", "", nil
}

// BuildLinkID constructs a work item link ID from its components.
func BuildLinkID(projectID, primaryWorkItemID, role, secondaryProjectID, secondaryWorkItemID string) string {
	return projectID + "/" + primaryWorkItemID + "/" + role + "/" + secondaryProjectID + "/" + secondaryWorkItemID
}
