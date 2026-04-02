// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"fmt"
	"strings"
)

// WorkItemExternalLink represents an externally linked work item from a
// work item hosted in another Polarion repository or server.
type WorkItemExternalLink struct {
	// Type is always "externallylinkedworkitems" for external work item links.
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the link.
	// Format: "{projectId}/{workItemId}/{roleId}/{hostname}/{targetProjectId}/{linkedWorkItemId}".
	ID string `json:"id,omitempty"`

	// Revision is the link revision.
	Revision string `json:"revision,omitempty"`

	// Attributes contains the external link attributes.
	Attributes *WorkItemExternalLinkAttributes `json:"attributes,omitempty"`

	// Links contains hypermedia links.
	Links *WorkItemExternalLinkLinks `json:"links,omitempty"`

	// Meta contains metadata about the external link.
	Meta *WorkItemExternalLinkMeta `json:"meta,omitempty"`
}

// WorkItemExternalLinkAttributes contains the external link attributes.
type WorkItemExternalLinkAttributes struct {
	// Role is the external link role ID.
	Role string `json:"role,omitempty"`

	// WorkItemURI is the URI of the externally linked work item.
	WorkItemURI string `json:"workItemURI,omitempty"`
}

// WorkItemExternalLinkLinks contains hypermedia links for an externally linked work item.
type WorkItemExternalLinkLinks struct {
	Self string `json:"self,omitempty"`
}

// WorkItemExternalLinkMeta contains metadata about the external link.
type WorkItemExternalLinkMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// ParseExternalLinkID parses an external link ID into its components.
// Expected format:
// "{projectId}/{workItemId}/{roleId}/{hostname}/{targetProjectId}/{linkedWorkItemId}".
func ParseExternalLinkID(linkID string) (projectID, workItemID, roleID, hostname, targetProjectID, linkedWorkItemID string, err error) {
	parts := strings.Split(linkID, "/")
	if len(parts) != 6 {
		return "", "", "", "", "", "", fmt.Errorf("invalid external link ID %q", linkID)
	}

	return parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], nil
}

// BuildExternalLinkID constructs an external link ID from its components.
func BuildExternalLinkID(projectID, workItemID, roleID, hostname, targetProjectID, linkedWorkItemID string) string {
	return strings.Join([]string{
		projectID,
		workItemID,
		roleID,
		hostname,
		targetProjectID,
		linkedWorkItemID,
	}, "/")
}

// NewWorkItemExternalLink creates a new externally linked work item.
func NewWorkItemExternalLink(role, workItemURI string) *WorkItemExternalLink {
	return &WorkItemExternalLink{
		Type: "externallylinkedworkitems",
		Attributes: &WorkItemExternalLinkAttributes{
			Role:        role,
			WorkItemURI: workItemURI,
		},
	}
}
