// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

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
