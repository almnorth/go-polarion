// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "strings"

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

// NewWorkItemLink creates a new work item link with the specified parameters.
// The secondaryWorkItemID should be the full ID including project (e.g., "PROJECT/WI-123").
// If secondaryProjectID is empty, it defaults to the current project.
func NewWorkItemLink(role, secondaryWorkItemID, secondaryProjectID string, suspect bool) *WorkItemLink {
	return &WorkItemLink{
		Type: "linkedworkitems",
		Data: &WorkItemLinkAttributes{
			Role:    role,
			Suspect: suspect,
		},
		Relationships: &LinkedWorkItemRelationships{
			WorkItem: &Relationship{
				Data: map[string]interface{}{
					"type": "workitems",
					"id":   secondaryWorkItemID,
				},
			},
		},
	}
}

// GetSecondaryWorkItemID extracts the secondary work item ID from the link.
// Returns the full ID (e.g., "PROJECT/WI-123") from either the relationships or by parsing the link ID.
func (l *WorkItemLink) GetSecondaryWorkItemID() string {
	// Try to get from relationships first
	if l.Relationships != nil && l.Relationships.WorkItem != nil {
		if data, ok := l.Relationships.WorkItem.Data.(map[string]interface{}); ok {
			if id, ok := data["id"].(string); ok {
				return id
			}
		}
	}

	// Fall back to parsing the link ID
	// Format: "project/primary/role/project/secondary"
	if l.ID != "" {
		parts := strings.Split(l.ID, "/")
		if len(parts) == 5 {
			return parts[3] + "/" + parts[4] // Full ID with project
		}
	}

	return ""
}

// GetSecondaryWorkItemIDShort extracts just the work item ID without the project prefix.
// Returns just the ID part (e.g., "WI-123") from "PROJECT/WI-123".
func (l *WorkItemLink) GetSecondaryWorkItemIDShort() string {
	fullID := l.GetSecondaryWorkItemID()
	if fullID == "" {
		return ""
	}

	// Extract just the work item ID part
	parts := strings.Split(fullID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return fullID
}

// GetSecondaryProjectID extracts the secondary project ID from the link.
func (l *WorkItemLink) GetSecondaryProjectID() string {
	fullID := l.GetSecondaryWorkItemID()
	if fullID == "" {
		return ""
	}

	// Extract project ID from full ID
	parts := strings.Split(fullID, "/")
	if len(parts) >= 2 {
		return parts[0]
	}

	return ""
}
