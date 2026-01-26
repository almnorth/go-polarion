// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// UserGroup represents a Polarion user group following the JSON:API format.
type UserGroup struct {
	// Type is always "usergroups" for user groups
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the user group
	ID string `json:"id,omitempty"`

	// Revision is the user group revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all user group attributes
	Attributes *UserGroupAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *UserGroupRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *UserGroupLinks `json:"links,omitempty"`

	// Meta contains metadata about the user group
	Meta *UserGroupMeta `json:"meta,omitempty"`
}

// UserGroupAttributes contains all user group attributes.
type UserGroupAttributes struct {
	// Name is the user group's display name
	Name string `json:"name,omitempty"`

	// Description is the user group's description
	Description *TextContent `json:"description,omitempty"`
}

// UserGroupRelationships contains relationships to other resources.
type UserGroupRelationships struct {
	// Users is the relationship to the group's users
	Users *Relationship `json:"users,omitempty"`

	// GlobalRoles is the relationship to the group's global roles
	GlobalRoles *Relationship `json:"globalRoles,omitempty"`

	// ProjectRoles is the relationship to the group's project roles
	ProjectRoles *Relationship `json:"projectRoles,omitempty"`
}

// UserGroupLinks contains hypermedia links for the user group.
type UserGroupLinks struct {
	Self string `json:"self,omitempty"`
}

// UserGroupMeta contains metadata about the user group.
type UserGroupMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}
