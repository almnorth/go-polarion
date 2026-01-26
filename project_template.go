// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// ProjectTemplate represents a Polarion project template.
// Project templates are used as blueprints when creating new projects.
type ProjectTemplate struct {
	// Type is the JSON:API resource type (always "projecttemplates")
	Type string `json:"type"`

	// ID is the unique identifier for the project template
	ID string `json:"id"`

	// Attributes contains the template properties
	Attributes *ProjectTemplateAttributes `json:"attributes,omitempty"`

	// Links contains related resource links
	Links *ProjectTemplateLinks `json:"links,omitempty"`
}

// ProjectTemplateAttributes contains project template properties.
type ProjectTemplateAttributes struct {
	// Name is the display name of the template
	Name string `json:"name,omitempty"`

	// Description provides details about the template
	Description string `json:"description,omitempty"`

	// TemplateID is the internal template identifier
	TemplateID string `json:"templateId,omitempty"`
}

// ProjectTemplateLinks contains links to related resources.
type ProjectTemplateLinks struct {
	// Self is the link to this resource
	Self string `json:"self,omitempty"`
}
