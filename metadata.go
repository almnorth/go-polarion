// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Victorien Elvinger
// Copyright (c) 2025 Siemens AG

package polarion

// Metadata represents Polarion instance metadata.
// This provides information about the Polarion server version, configuration,
// and REST API properties.
//
// Requires: Polarion >= 2512
type Metadata struct {
	Type       string             `json:"type"`
	ID         string             `json:"id"`
	Attributes MetadataAttributes `json:"attributes"`
}

// MetadataAttributes contains Polarion instance information including
// version, build details, and REST API configuration.
type MetadataAttributes struct {
	// Version is the Polarion version (e.g., "3.25.12")
	Version string `json:"version"`

	// Build is the build number (e.g., "20250613-1404-master-e594c717")
	Build string `json:"build"`

	// Node is the node identifier
	Node string `json:"node"`

	// Cluster is the cluster identifier
	Cluster string `json:"cluster,omitempty"`

	// Timezone is the server timezone (e.g., "+05:30")
	Timezone string `json:"timezone"`

	// LogoURL is the URL to the Polarion logo
	LogoURL string `json:"logoUrl"`

	// APIProperties contains REST API configuration limits
	APIProperties *APIProperties `json:"apiProperties,omitempty"`
}

// APIProperties contains REST API configuration limits and constraints.
type APIProperties struct {
	// DefaultPageSize is the default number of items per page
	DefaultPageSize int `json:"defaultPageSize"`

	// MaxPageSize is the maximum number of items per page
	MaxPageSize int `json:"maxPageSize"`

	// MaxRelationshipSize is the maximum number of relationships
	MaxRelationshipSize int `json:"maxRelationshipSize"`

	// BodySizeLimit is the maximum request body size in bytes
	BodySizeLimit int `json:"bodySizeLimit"`

	// MaxIncludedSize is the maximum number of includes
	MaxIncludedSize int `json:"maxIncludedSize"`
}

// VersionInfo provides parsed version information for easier version comparison.
type VersionInfo struct {
	// Major version number
	Major int

	// Minor version number
	Minor int

	// Patch version number
	Patch int
}

// FieldsMetadata represents metadata for fields of a resource type.
// This provides information about all available fields including custom fields
// for a specific resource and target type combination.
//
// The response structure contains both attributes (primitive fields) and
// relationships (relationship fields) with their type information.
//
// Requires: Polarion >= 2512
type FieldsMetadata struct {
	// Data contains the fields metadata
	Data FieldsMetadataData `json:"data"`

	// Links contains the self link
	Links *FieldsMetadataLinks `json:"links,omitempty"`
}

// FieldsMetadataData contains the actual field definitions.
type FieldsMetadataData struct {
	// Attributes contains primitive field definitions (string, integer, etc.)
	// The key is the field ID, and the value contains the field metadata
	Attributes map[string]FieldMetadata `json:"attributes,omitempty"`

	// Relationships contains relationship field definitions
	// The key is the field ID, and the value contains the field metadata
	Relationships map[string]FieldMetadata `json:"relationships,omitempty"`
}

// FieldsMetadataLinks contains links related to the fields metadata.
type FieldsMetadataLinks struct {
	// Self is the URL to this resource
	Self string `json:"self,omitempty"`
}

// FieldMetadata represents metadata for a single field including its type,
// label, and constraints.
type FieldMetadata struct {
	// Type contains the field type information
	Type CustomFieldType `json:"type"`

	// Label is the human-readable field name
	Label string `json:"label,omitempty"`
}
