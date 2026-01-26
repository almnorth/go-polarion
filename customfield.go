// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Victorien Elvinger
// Copyright (c) 2025 Siemens AG

package polarion

import "fmt"

// CustomFieldsConfig represents the custom fields configuration for a resource type.
// This is the top-level structure returned by the Custom Fields API.
//
// Requires: Polarion >= 2512
type CustomFieldsConfig struct {
	Type       string                       `json:"type"`
	ID         string                       `json:"id"`
	Attributes CustomFieldsConfigAttributes `json:"attributes"`
	Links      *CustomFieldsConfigLinks     `json:"links,omitempty"`
}

// CustomFieldsConfigAttributes contains the custom fields configuration.
type CustomFieldsConfigAttributes struct {
	// Fields is the list of custom field definitions
	Fields []CustomFieldDefinition `json:"fields"`

	// ResourceType is the resource type (e.g., "workitems", "documents")
	ResourceType string `json:"resourceType,omitempty"`

	// TargetType is the specific type within the resource (e.g., "requirement", "task")
	TargetType string `json:"targetType,omitempty"`
}

// CustomFieldsConfigLinks contains links related to the custom fields configuration.
type CustomFieldsConfigLinks struct {
	// Self is the URL to this resource
	Self string `json:"self,omitempty"`
}

// CustomFieldDefinition represents a single custom field definition.
type CustomFieldDefinition struct {
	// ID is the field identifier
	ID string `json:"id"`

	// Name is the human-readable field name
	Name string `json:"name,omitempty"`

	// Description is the field description
	Description string `json:"description,omitempty"`

	// Required indicates if the field is required
	Required bool `json:"required,omitempty"`

	// DefaultValue is the default value for the field
	DefaultValue string `json:"defaultValue,omitempty"`

	// DependsOn specifies field dependencies
	DependsOn string `json:"dependsOn,omitempty"`

	// Type contains the field type information
	Type CustomFieldType `json:"type"`

	// Parameters contains additional field parameters
	Parameters []CustomFieldParameter `json:"parameters,omitempty"`
}

// CustomFieldType represents the type information for a custom field.
// The kind field indicates the data type (e.g., "string", "integer", "enumeration", "relationship").
type CustomFieldType struct {
	// Kind is the field type kind
	// Possible values: "string", "text", "text/html", "integer", "float", "time", "date", "date-time",
	// "duration", "boolean", "enumeration", "relationship"
	Kind string `json:"kind"`

	// EnumName is the enumeration name for enumeration fields
	EnumName string `json:"enumName,omitempty"`

	// EnumContext is the enumeration context for enumeration fields
	EnumContext string `json:"enumContext,omitempty"`

	// TargetResourceTypes specifies allowed resource types for relationship fields
	TargetResourceTypes []string `json:"targetResourceTypes,omitempty"`

	// Role is the relationship role for relationship fields
	Role string `json:"role,omitempty"`
}

// CustomFieldParameter represents a parameter for a custom field.
type CustomFieldParameter struct {
	// Key is the parameter key
	Key string `json:"key"`

	// Name is the parameter name
	Name string `json:"name,omitempty"`

	// Title is the parameter title
	Title string `json:"title,omitempty"`
}

// CustomFieldID is a helper for constructing custom field IDs.
// Custom field IDs follow the format: {resourceType}/{targetType}
type CustomFieldID struct {
	// ResourceType is the resource type (e.g., "workitems", "documents", "testruns", "plans")
	ResourceType string

	// TargetType is the specific type within the resource (e.g., "requirement", "task")
	// Use "~" to represent no target type
	TargetType string
}

// String returns the custom field ID in the format: {resourceType}/{targetType}
func (id CustomFieldID) String() string {
	return fmt.Sprintf("%s/%s", id.ResourceType, id.TargetType)
}

// NewCustomFieldsConfig creates a new custom fields configuration.
func NewCustomFieldsConfig(resourceType, targetType string) *CustomFieldsConfig {
	id := CustomFieldID{
		ResourceType: resourceType,
		TargetType:   targetType,
	}
	return &CustomFieldsConfig{
		Type: "customfields",
		ID:   id.String(),
		Attributes: CustomFieldsConfigAttributes{
			Fields:       []CustomFieldDefinition{},
			ResourceType: resourceType,
			TargetType:   targetType,
		},
	}
}
