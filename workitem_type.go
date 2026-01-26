// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// WorkItemType represents a work item type definition following the JSON:API format.
// Work item types define the structure and fields available for different kinds of work items.
type WorkItemType struct {
	// Type is always "workitemtypes" for work item types
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the work item type (e.g., "requirement", "task", "defect")
	ID string `json:"id,omitempty"`

	// Attributes contains the work item type attributes
	Attributes *WorkItemTypeAttributes `json:"attributes,omitempty"`

	// Links contains hypermedia links
	Links *WorkItemTypeLinks `json:"links,omitempty"`

	// Meta contains metadata about the work item type
	Meta *WorkItemTypeMeta `json:"meta,omitempty"`
}

// WorkItemTypeAttributes contains the attributes of a work item type.
type WorkItemTypeAttributes struct {
	// Name is the display name of the work item type
	Name string `json:"name,omitempty"`

	// Icon is the icon identifier for the work item type
	Icon string `json:"icon,omitempty"`

	// Description provides additional information about the type
	Description string `json:"description,omitempty"`

	// Fields contains the field definitions for this work item type
	Fields []FieldDefinition `json:"fields,omitempty"`
}

// FieldDefinition represents a field definition for a work item type.
type FieldDefinition struct {
	// ID is the unique identifier of the field
	ID string `json:"id"`

	// Name is the display name of the field
	Name string `json:"name,omitempty"`

	// Type is the field type (e.g., "string", "text", "enum", "date", "boolean")
	Type string `json:"type,omitempty"`

	// Required indicates if the field is required
	Required bool `json:"required,omitempty"`

	// ReadOnly indicates if the field is read-only
	ReadOnly bool `json:"readOnly,omitempty"`

	// DefaultValue is the default value for the field
	DefaultValue interface{} `json:"defaultValue,omitempty"`

	// EnumerationID references an enumeration for enum-type fields
	EnumerationID string `json:"enumerationId,omitempty"`

	// Description provides additional information about the field
	Description string `json:"description,omitempty"`

	// MultiValue indicates if the field can hold multiple values
	MultiValue bool `json:"multiValue,omitempty"`

	// Computed indicates if the field is computed/calculated
	Computed bool `json:"computed,omitempty"`
}

// WorkItemTypeLinks contains hypermedia links for the work item type.
type WorkItemTypeLinks struct {
	Self string `json:"self,omitempty"`
}

// WorkItemTypeMeta contains metadata about the work item type.
type WorkItemTypeMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// FieldType represents the type of a work item field.
type FieldType string

// Common field types
const (
	FieldTypeString   FieldType = "string"
	FieldTypeText     FieldType = "text"
	FieldTypeEnum     FieldType = "enum"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeInteger  FieldType = "integer"
	FieldTypeFloat    FieldType = "float"
	FieldTypeUser     FieldType = "user"
	FieldTypeUsers    FieldType = "users"
)

// NewFieldDefinition creates a new field definition with the specified ID and type.
func NewFieldDefinition(id string, fieldType FieldType) *FieldDefinition {
	return &FieldDefinition{
		ID:   id,
		Type: string(fieldType),
	}
}

// WithName sets the name of the field definition.
func (f *FieldDefinition) WithName(name string) *FieldDefinition {
	f.Name = name
	return f
}

// WithRequired sets whether the field is required.
func (f *FieldDefinition) WithRequired(required bool) *FieldDefinition {
	f.Required = required
	return f
}

// WithDefaultValue sets the default value of the field.
func (f *FieldDefinition) WithDefaultValue(value interface{}) *FieldDefinition {
	f.DefaultValue = value
	return f
}

// WithEnumeration sets the enumeration ID for enum-type fields.
func (f *FieldDefinition) WithEnumeration(enumID string) *FieldDefinition {
	f.EnumerationID = enumID
	return f
}
