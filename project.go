// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

// Project represents a Polarion project.
// Projects are the top-level organizational units in Polarion.
type Project struct {
	// Type is the JSON:API resource type (always "projects")
	Type string `json:"type"`

	// ID is the unique identifier for the project
	ID string `json:"id"`

	// Attributes contains the project properties
	Attributes *ProjectAttributes `json:"attributes,omitempty"`

	// Links contains related resource links
	Links *ProjectLinks `json:"links,omitempty"`
}

// ProjectAttributes contains project properties.
type ProjectAttributes struct {
	// Name is the display name of the project
	Name string `json:"name,omitempty"`

	// Description provides details about the project
	Description *TextContent `json:"description,omitempty"`

	// Active indicates if the project is active
	Active bool `json:"active,omitempty"`

	// Location is the project's location in the repository
	Location string `json:"location,omitempty"`

	// Lead is the project lead user ID
	Lead string `json:"lead,omitempty"`

	// StartDate is the project start date
	StartDate string `json:"startDate,omitempty"`

	// FinishDate is the project finish date
	FinishDate string `json:"finishDate,omitempty"`
}

// ProjectLinks contains links to related resources.
type ProjectLinks struct {
	// Self is the link to this resource
	Self string `json:"self,omitempty"`
}

// CreateProjectRequest represents project creation parameters.
type CreateProjectRequest struct {
	// ProjectID is the unique identifier for the new project
	ProjectID string `json:"projectId"`

	// Name is the display name of the project
	Name string `json:"name"`

	// Location is the repository location for the project (e.g., "/default")
	// If not provided, defaults to "/default"
	Location string `json:"location,omitempty"`

	// TrackerPrefix is the prefix for work item IDs in this project (e.g., "TEST" for TEST-123)
	// If not provided, defaults to the project ID
	TrackerPrefix string `json:"trackerPrefix,omitempty"`

	// Description provides details about the project
	Description string `json:"description,omitempty"`

	// TemplateID is the ID of the template to use for project creation
	TemplateID string `json:"templateId,omitempty"`

	// ParentID is the ID of the parent project (for subprojects)
	ParentID string `json:"parentId,omitempty"`
}

// MoveProjectRequest represents project move parameters.
type MoveProjectRequest struct {
	// NewLocation is the new location path for the project
	NewLocation string `json:"newLocation"`
}

// ProjectClient provides project-scoped operations.
// It contains services for different resource types within a project.
type ProjectClient struct {
	projectID string
	client    *Client

	// WorkItems provides access to work item operations
	WorkItems *WorkItemService

	// Enumerations provides access to enumeration operations
	Enumerations *EnumerationService

	// WorkItemLinks provides access to work item link operations
	WorkItemLinks *WorkItemLinkService

	// WorkItemTypes provides access to work item type definition operations
	WorkItemTypes *WorkItemTypeService

	// WorkItemComments provides access to work item comment operations
	WorkItemComments *WorkItemCommentService

	// WorkItemAttachments provides access to work item attachment operations
	WorkItemAttachments *WorkItemAttachmentService

	// WorkItemApprovals provides access to work item approval operations
	WorkItemApprovals *WorkItemApprovalService

	// WorkItemWorkRecords provides access to work item work record (time tracking) operations
	WorkItemWorkRecords *WorkItemWorkRecordService

	// TestParameters provides access to test parameter definition operations
	TestParameters *TestParameterService

	// CustomFields provides access to project custom field operations (Polarion >= 2512)
	CustomFields *CustomFieldService

	// FieldsMetadata provides access to project fields metadata operations (Polarion >= 2512)
	FieldsMetadata *ProjectFieldsMetadataService
}

// newProjectClient creates a new project-scoped client.
func newProjectClient(client *Client, projectID string) *ProjectClient {
	pc := &ProjectClient{
		projectID: projectID,
		client:    client,
	}

	// Initialize services
	pc.WorkItems = newWorkItemService(pc)
	pc.Enumerations = newEnumerationService(pc)
	pc.WorkItemLinks = newWorkItemLinkService(pc)
	pc.WorkItemTypes = newWorkItemTypeService(pc)
	pc.WorkItemComments = newWorkItemCommentService(pc)
	pc.WorkItemAttachments = newWorkItemAttachmentService(pc)
	pc.WorkItemApprovals = newWorkItemApprovalService(pc)
	pc.WorkItemWorkRecords = newWorkItemWorkRecordService(pc)
	pc.TestParameters = newTestParameterService(client, projectID)
	pc.CustomFields = &CustomFieldService{client: client, projectID: projectID}
	pc.FieldsMetadata = &ProjectFieldsMetadataService{client: client, projectID: projectID}

	return pc
}

// ProjectID returns the project ID for this client.
func (pc *ProjectClient) ProjectID() string {
	return pc.projectID
}

// Client returns the underlying Client.
func (pc *ProjectClient) Client() *Client {
	return pc.client
}
