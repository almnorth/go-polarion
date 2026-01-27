// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"encoding/json"
	"time"
)

// WorkItem represents a Polarion work item following the JSON:API format.
// The structure matches the Polarion REST API response format where custom fields
// are flat in the attributes object, not nested.
type WorkItem struct {
	// Type is always "workitems" for work items
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the work item (e.g., "MyProject/WI-123")
	ID string `json:"id,omitempty"`

	// Revision is the work item revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all work item attributes (both standard and custom)
	Attributes *WorkItemAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *WorkItemRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *WorkItemLinks `json:"links,omitempty"`

	// Meta contains metadata about the work item
	Meta *WorkItemMeta `json:"meta,omitempty"`
}

// WorkItemAttributes contains all work item attributes.
// This includes both standard Polarion fields and custom fields.
// Custom fields can be added by embedding this struct or using the CustomFields map.
type WorkItemAttributes struct {
	// Standard Polarion fields
	Type              string       `json:"type,omitempty"`
	Created           *time.Time   `json:"created,omitempty"`
	Updated           *time.Time   `json:"updated,omitempty"`
	Title             string       `json:"title,omitempty"`
	Description       *TextContent `json:"description,omitempty"`
	Status            string       `json:"status,omitempty"`
	Resolution        string       `json:"resolution,omitempty"`
	Priority          string       `json:"priority,omitempty"`
	Severity          string       `json:"severity,omitempty"`
	DueDate           string       `json:"dueDate,omitempty"`
	PlannedStart      *time.Time   `json:"plannedStart,omitempty"`
	PlannedEnd        *time.Time   `json:"plannedEnd,omitempty"`
	InitialEstimate   string       `json:"initialEstimate,omitempty"`
	RemainingEstimate string       `json:"remainingEstimate,omitempty"`
	TimeSpent         string       `json:"timeSpent,omitempty"`
	OutlineNumber     string       `json:"outlineNumber,omitempty"`
	ResolvedOn        *time.Time   `json:"resolvedOn,omitempty"`

	// Hyperlinks
	Hyperlinks []Hyperlink `json:"hyperlinks,omitempty"`

	// CustomFields is a map for any additional custom fields
	// This allows for flexible handling of project-specific fields
	CustomFields map[string]interface{} `json:"-"`
}

// TextContent represents rich text content with a specific content type.
type TextContent struct {
	Type  string `json:"type"` // e.g., "text/html", "text/plain"
	Value string `json:"value"`
}

// Hyperlink represents a hyperlink in a work item.
type Hyperlink struct {
	URI  string `json:"uri,omitempty"`
	Role string `json:"role,omitempty"`
}

// WorkItemRelationships contains relationships to other resources.
type WorkItemRelationships struct {
	Assignee         *Relationship `json:"assignee,omitempty"`
	Author           *Relationship `json:"author,omitempty"`
	Categories       *Relationship `json:"categories,omitempty"`
	LinkedWorkItems  *Relationship `json:"linkedWorkItems,omitempty"`
	Attachments      *Relationship `json:"attachments,omitempty"`
	Comments         *Relationship `json:"comments,omitempty"`
	ExternallyLinked *Relationship `json:"externallyLinkedWorkItems,omitempty"`
	LinkedOslc       *Relationship `json:"linkedOslcResources,omitempty"`
	Module           *Relationship `json:"module,omitempty"`
	ModuleFolder     *Relationship `json:"moduleFolder,omitempty"`
	Plan             *Relationship `json:"plan,omitempty"`
	Project          *Relationship `json:"project,omitempty"`
	Votes            *Relationship `json:"votes,omitempty"`
	Watches          *Relationship `json:"watches,omitempty"`
	WorkRecords      *Relationship `json:"workRecords,omitempty"`
	ApprovalRecords  *Relationship `json:"approvals,omitempty"`
}

// Relationship represents a JSON:API relationship.
type Relationship struct {
	Data  interface{}       `json:"data,omitempty"`
	Links *RelationshipLink `json:"links,omitempty"`
	Meta  *RelationshipMeta `json:"meta,omitempty"`
}

// RelationshipLink contains links for a relationship.
type RelationshipLink struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
}

// RelationshipMeta contains metadata for a relationship.
type RelationshipMeta struct {
	TotalCount int `json:"totalCount,omitempty"`
}

// WorkItemLinks contains hypermedia links for the work item.
type WorkItemLinks struct {
	Self string `json:"self,omitempty"`
}

// WorkItemMeta contains metadata about the work item.
type WorkItemMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// WorkItemLink represents a link between two work items.
type WorkItemLink struct {
	Type          string                       `json:"type,omitempty"`
	ID            string                       `json:"id,omitempty"`
	Data          *WorkItemLinkAttributes      `json:"attributes,omitempty"`
	Relationships *LinkedWorkItemRelationships `json:"relationships,omitempty"`
	Links         *WorkItemLinks               `json:"links,omitempty"`
}

// WorkItemLinkAttributes contains attributes of a work item link.
type WorkItemLinkAttributes struct {
	Role     string `json:"role,omitempty"`
	Suspect  bool   `json:"suspect,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// NewTextContent creates a new TextContent with the specified type and value.
func NewTextContent(contentType, value string) *TextContent {
	return &TextContent{
		Type:  contentType,
		Value: value,
	}
}

// NewHTMLContent creates a new TextContent with HTML content.
func NewHTMLContent(html string) *TextContent {
	return &TextContent{
		Type:  "text/html",
		Value: html,
	}
}

// NewPlainTextContent creates a new TextContent with plain text content.
func NewPlainTextContent(text string) *TextContent {
	return &TextContent{
		Type:  "text/plain",
		Value: text,
	}
}

// GetCustomField retrieves a custom field value by name from CustomFields map.
func (a *WorkItemAttributes) GetCustomField(name string) interface{} {
	if a.CustomFields == nil {
		return nil
	}
	return a.CustomFields[name]
}

// SetCustomField sets a custom field value in the CustomFields map.
func (a *WorkItemAttributes) SetCustomField(name string, value interface{}) {
	if a.CustomFields == nil {
		a.CustomFields = make(map[string]interface{})
	}
	a.CustomFields[name] = value
}

// HasCustomField checks if a custom field exists in the CustomFields map.
func (a *WorkItemAttributes) HasCustomField(name string) bool {
	if a.CustomFields == nil {
		return false
	}
	_, exists := a.CustomFields[name]
	return exists
}

// UnmarshalJSON implements custom JSON unmarshaling for WorkItemAttributes.
// It unmarshals known standard fields and captures any remaining fields as custom fields.
func (a *WorkItemAttributes) UnmarshalJSON(data []byte) error {
	// Define a type alias to avoid infinite recursion
	type Alias WorkItemAttributes

	// First, unmarshal into a map to capture all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Unmarshal into the alias to populate standard fields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Define the set of known standard fields
	// These are the fields explicitly defined in WorkItemAttributes struct
	knownFields := map[string]bool{
		"id":                true, // ID field from work item level
		"type":              true,
		"created":           true,
		"updated":           true,
		"title":             true,
		"description":       true,
		"status":            true,
		"resolution":        true,
		"priority":          true,
		"severity":          true,
		"dueDate":           true,
		"plannedStart":      true,
		"plannedEnd":        true,
		"initialEstimate":   true,
		"remainingEstimate": true,
		"timeSpent":         true,
		"outlineNumber":     true,
		"resolvedOn":        true,
		"hyperlinks":        true,
	}

	// Initialize CustomFields map if needed
	if a.CustomFields == nil {
		a.CustomFields = make(map[string]interface{})
	}

	// Populate CustomFields with any fields not in the known set
	for key, value := range raw {
		if !knownFields[key] {
			var v interface{}
			if err := json.Unmarshal(value, &v); err != nil {
				return err
			}
			a.CustomFields[key] = v
		}
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for WorkItemAttributes.
// It marshals standard fields and merges in custom fields at the same level.
func (a *WorkItemAttributes) MarshalJSON() ([]byte, error) {
	// Define a type alias to avoid infinite recursion
	type Alias WorkItemAttributes

	// Marshal the standard fields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// If there are no custom fields, return the standard fields
	if len(a.CustomFields) == 0 {
		return data, nil
	}

	// Unmarshal the standard fields into a map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Merge custom fields into the map
	for key, value := range a.CustomFields {
		result[key] = value
	}

	// Marshal the combined map
	return json.Marshal(result)
}

// Clone creates a deep copy of a WorkItem.
// This is useful when you need to modify a work item without affecting the original.
func (w *WorkItem) Clone() *WorkItem {
	if w == nil {
		return nil
	}

	clone := &WorkItem{
		Type:     w.Type,
		ID:       w.ID,
		Revision: w.Revision,
	}

	// Clone attributes
	if w.Attributes != nil {
		clone.Attributes = &WorkItemAttributes{
			Type:              w.Attributes.Type,
			Created:           w.Attributes.Created,
			Updated:           w.Attributes.Updated,
			Title:             w.Attributes.Title,
			Status:            w.Attributes.Status,
			Resolution:        w.Attributes.Resolution,
			Priority:          w.Attributes.Priority,
			Severity:          w.Attributes.Severity,
			DueDate:           w.Attributes.DueDate,
			PlannedStart:      w.Attributes.PlannedStart,
			PlannedEnd:        w.Attributes.PlannedEnd,
			InitialEstimate:   w.Attributes.InitialEstimate,
			RemainingEstimate: w.Attributes.RemainingEstimate,
			TimeSpent:         w.Attributes.TimeSpent,
			OutlineNumber:     w.Attributes.OutlineNumber,
			ResolvedOn:        w.Attributes.ResolvedOn,
			CustomFields:      make(map[string]interface{}),
		}

		// Clone description
		if w.Attributes.Description != nil {
			clone.Attributes.Description = &TextContent{
				Type:  w.Attributes.Description.Type,
				Value: w.Attributes.Description.Value,
			}
		}

		// Clone hyperlinks
		if len(w.Attributes.Hyperlinks) > 0 {
			clone.Attributes.Hyperlinks = make([]Hyperlink, len(w.Attributes.Hyperlinks))
			copy(clone.Attributes.Hyperlinks, w.Attributes.Hyperlinks)
		}

		// Deep copy custom fields
		for k, v := range w.Attributes.CustomFields {
			clone.Attributes.CustomFields[k] = v
		}
	}

	return clone
}
