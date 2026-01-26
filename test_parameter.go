// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// TestParameter represents a test parameter definition in Polarion.
// Test parameters define configurable values that can be used in test cases.
type TestParameter struct {
	// Type is the JSON:API resource type (always "testparameterdefinitions")
	Type string `json:"type"`

	// ID is the unique identifier for the test parameter
	ID string `json:"id"`

	// Attributes contains the parameter properties
	Attributes *TestParameterAttributes `json:"attributes,omitempty"`

	// Links contains related resource links
	Links *TestParameterLinks `json:"links,omitempty"`
}

// TestParameterAttributes contains test parameter properties.
type TestParameterAttributes struct {
	// Name is the display name of the parameter
	Name string `json:"name,omitempty"`

	// Description provides details about the parameter
	Description string `json:"description,omitempty"`

	// Type specifies the parameter type (e.g., "string", "enum", "boolean")
	Type string `json:"type,omitempty"`

	// DefaultValue is the default value for the parameter
	DefaultValue string `json:"defaultValue,omitempty"`

	// AllowedValues contains the list of allowed values for enum types
	AllowedValues []string `json:"allowedValues,omitempty"`

	// Required indicates if the parameter is mandatory
	Required bool `json:"required,omitempty"`
}

// TestParameterLinks contains links to related resources.
type TestParameterLinks struct {
	// Self is the link to this resource
	Self string `json:"self,omitempty"`
}
