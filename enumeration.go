// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// Enumeration represents a Polarion enumeration following the JSON:API format.
// Enumerations define the allowed values for enumerated fields in work items.
type Enumeration struct {
	// Type is always "enumerations" for enumerations
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the enumeration
	// Format: "project/{projectId}/enum/{context}/{name}/{targetType}"
	ID string `json:"id,omitempty"`

	// Attributes contains the enumeration attributes
	Attributes *EnumerationAttributes `json:"attributes,omitempty"`

	// Links contains hypermedia links
	Links *EnumerationLinks `json:"links,omitempty"`

	// Meta contains metadata about the enumeration
	Meta *EnumerationMeta `json:"meta,omitempty"`
}

// EnumerationAttributes contains the attributes of an enumeration.
type EnumerationAttributes struct {
	// Options is the list of enumeration values
	Options []EnumerationOption `json:"options,omitempty"`
}

// EnumerationOption represents a single value in an enumeration.
type EnumerationOption struct {
	// ID is the unique identifier of the option
	ID string `json:"id"`

	// Name is the display name of the option
	Name string `json:"name,omitempty"`

	// Description provides additional information about the option
	Description string `json:"description,omitempty"`

	// Color is the color associated with the option (hex format)
	Color string `json:"color,omitempty"`

	// Default indicates if this is the default value
	Default bool `json:"default,omitempty"`

	// Hidden indicates if this option should be hidden
	Hidden bool `json:"hidden,omitempty"`

	// Sequence defines the display order
	Sequence int `json:"sequence,omitempty"`
}

// EnumerationLinks contains hypermedia links for the enumeration.
type EnumerationLinks struct {
	Self string `json:"self,omitempty"`
}

// EnumerationMeta contains metadata about the enumeration.
type EnumerationMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// EnumerationID represents the components of an enumeration ID.
type EnumerationID struct {
	// Context is the enumeration context (e.g., "workitem", "document")
	Context string

	// Name is the enumeration name (e.g., "status", "priority")
	Name string

	// TargetType is the target type (e.g., "requirement", "task")
	TargetType string
}

// String returns the full enumeration ID path component.
func (e *EnumerationID) String() string {
	return e.Context + "/" + e.Name + "/" + e.TargetType
}

// NewEnumerationID creates a new EnumerationID from components.
func NewEnumerationID(context, name, targetType string) *EnumerationID {
	return &EnumerationID{
		Context:    context,
		Name:       name,
		TargetType: targetType,
	}
}
