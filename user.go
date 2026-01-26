// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// User represents a Polarion user following the JSON:API format.
type User struct {
	// Type is always "users" for users
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the user
	ID string `json:"id,omitempty"`

	// Revision is the user revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all user attributes
	Attributes *UserAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *UserRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *UserLinks `json:"links,omitempty"`

	// Meta contains metadata about the user
	Meta *UserMeta `json:"meta,omitempty"`
}

// UserAttributes contains all user attributes.
type UserAttributes struct {
	// Name is the user's display name
	Name string `json:"name,omitempty"`

	// Email is the user's email address
	Email string `json:"email,omitempty"`

	// Description is the user's description
	Description *TextContent `json:"description,omitempty"`

	// Disabled indicates if the user account is disabled
	Disabled bool `json:"disabled,omitempty"`

	// DisabledForUI indicates if the user is disabled for UI access
	DisabledForUI bool `json:"disabledForUi,omitempty"`

	// VaultUser indicates if this is a vault user
	VaultUser bool `json:"vaultUser,omitempty"`
}

// UserRelationships contains relationships to other resources.
type UserRelationships struct {
	// Avatar is the relationship to the user's avatar
	Avatar *Relationship `json:"avatar,omitempty"`

	// UserGroups is the relationship to the user's groups
	UserGroups *Relationship `json:"userGroups,omitempty"`

	// GlobalRoles is the relationship to the user's global roles
	GlobalRoles *Relationship `json:"globalRoles,omitempty"`

	// ProjectRoles is the relationship to the user's project roles
	ProjectRoles *Relationship `json:"projectRoles,omitempty"`
}

// UserLinks contains hypermedia links for the user.
type UserLinks struct {
	Self string `json:"self,omitempty"`
}

// UserMeta contains metadata about the user.
type UserMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// License represents a Polarion license.
type License struct {
	// Type is always "licenses" for licenses
	Type string `json:"type,omitempty"`

	// ID is the license identifier
	ID string `json:"id,omitempty"`

	// Attributes contains license attributes
	Attributes *LicenseAttributes `json:"attributes,omitempty"`
}

// LicenseAttributes contains license attributes.
type LicenseAttributes struct {
	// Name is the license name
	Name string `json:"name,omitempty"`

	// Description is the license description
	Description string `json:"description,omitempty"`
}

// UserAvatar represents a user's avatar image.
type UserAvatar struct {
	// Data contains the avatar image data
	Data []byte

	// ContentType is the MIME type of the avatar image
	ContentType string
}
