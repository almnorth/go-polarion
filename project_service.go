// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// ProjectService handles project-level operations.
// Projects are the top-level organizational units in Polarion.
type ProjectService struct {
	client *Client
}

// newProjectService creates a new project service.
func newProjectService(client *Client) *ProjectService {
	return &ProjectService{
		client: client,
	}
}

// Get retrieves a specific project.
//
// Endpoint: GET /projects/{projectId}
//
// Example:
//
//	project, err := client.Projects.Get(ctx, "myproject")
func (s *ProjectService) Get(ctx context.Context, projectID string, opts ...QueryOption) (*Project, error) {
	if projectID == "" {
		return nil, NewValidationError("projectID", "project ID is required")
	}

	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s", s.client.baseURL, url.PathEscape(projectID))

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var project Project
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &project)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectID, err)
	}

	return &project, nil
}

// List returns all projects.
//
// Endpoint: GET /projects
//
// Example:
//
//	projects, err := client.Projects.List(ctx)
//	for _, project := range projects {
//	    fmt.Printf("Project: %s (%s)\n", project.Attributes.Name, project.ID)
//	}
func (s *ProjectService) List(ctx context.Context, opts ...QueryOption) ([]*Project, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects", s.client.baseURL)

	// Build query parameters
	params := url.Values{}

	// Set page size if specified
	if options.pageSize > 0 {
		params.Set("page[size]", strconv.Itoa(options.pageSize))
	}

	// Add field selection
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}

	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var response struct {
		Data []*Project `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return response.Data, nil
}

// Create creates a new project.
// Project creation is an asynchronous operation that returns a job.
//
// Endpoint: POST /projects/actions/createProject
//
// Example:
//
//	req := &polarion.CreateProjectRequest{
//	    ProjectID:   "newproject",
//	    Name:        "New Project",
//	    Description: "Project description",
//	    TemplateID:  "template_id",
//	}
//	project, err := client.Projects.Create(ctx, req)
func (s *ProjectService) Create(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
	if req == nil {
		return nil, NewValidationError("req", "create project request is required")
	}

	if req.ProjectID == "" {
		return nil, NewValidationError("projectID", "project ID is required")
	}

	if req.Name == "" {
		return nil, NewValidationError("name", "project name is required")
	}

	if req.Location == "" {
		return nil, NewValidationError("location", "project location is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/actions/createProject", s.client.baseURL)

	// Prepare request body - note: this endpoint does NOT use JSON:API format
	// It expects a flat structure with projectId, location, trackerPrefix, templateId, and params
	trackerPrefix := req.TrackerPrefix
	if trackerPrefix == "" {
		// Default to project ID if not specified
		trackerPrefix = req.ProjectID
	}

	body := map[string]interface{}{
		"projectId":     req.ProjectID,
		"location":      req.Location,
		"trackerPrefix": trackerPrefix,
		"params": map[string]interface{}{
			"name": req.Name,
		},
	}

	// Add optional fields
	if req.TemplateID != "" {
		body["templateId"] = req.TemplateID
	}
	if req.ParentID != "" {
		body["parentId"] = req.ParentID
	}

	// Note: Description might need to be set after creation via Update
	// as the create endpoint may not support it directly

	// Make request with retry
	// Note: This returns a job, but we'll return the project for simplicity
	var response struct {
		Data struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				Status string `json:"status"`
			} `json:"attributes"`
		} `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Return a basic project structure
	// In a real scenario, you might want to poll the job status
	project := &Project{
		Type: "projects",
		ID:   req.ProjectID,
		Attributes: &ProjectAttributes{
			Name: req.Name,
		},
	}
	if req.Description != "" {
		project.Attributes.Description = NewPlainTextContent(req.Description)
	}
	return project, nil
}

// Update updates a project.
//
// Endpoint: PATCH /projects/{projectId}
//
// Example:
//
//	project.Attributes.Description = "Updated description"
//	updated, err := client.Projects.Update(ctx, project)
func (s *ProjectService) Update(ctx context.Context, project *Project) (*Project, error) {
	if project == nil {
		return nil, NewValidationError("project", "project cannot be nil")
	}

	if project.ID == "" {
		return nil, NewValidationError("ID", "project ID is required for update")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s", s.client.baseURL, url.PathEscape(project.ID))

	// Prepare request body - exclude links as they're read-only
	projectData := map[string]interface{}{
		"type":       project.Type,
		"id":         project.ID,
		"attributes": project.Attributes,
	}

	body := map[string]interface{}{
		"data": projectData,
	}

	// Make request with retry
	var updated Project
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}

		// Check if response has content (some PATCH operations return 204 No Content)
		if resp.StatusCode == 204 || resp.ContentLength == 0 {
			// No content returned, use the input project as the result
			updated = *project
			resp.Body.Close()
			return nil
		}

		return internalhttp.DecodeDataResponse(resp, &updated)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", project.ID, err)
	}

	return &updated, nil
}

// Delete deletes a project.
//
// Endpoint: DELETE /projects/{projectId}
//
// Example:
//
//	err := client.Projects.Delete(ctx, "myproject")
func (s *ProjectService) Delete(ctx context.Context, projectID string) error {
	if projectID == "" {
		return NewValidationError("projectID", "project ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s", s.client.baseURL, url.PathEscape(projectID))

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "DELETE", urlStr, nil)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete project %s: %w", projectID, err)
	}

	return nil
}

// Mark marks a project.
// Marking a project adds it to the user's list of favorite projects.
//
// Endpoint: POST /projects/actions/markProject
//
// Example:
//
//	err := client.Projects.Mark(ctx, "myproject")
func (s *ProjectService) Mark(ctx context.Context, projectID string) error {
	if projectID == "" {
		return NewValidationError("projectID", "project ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/actions/markProject", s.client.baseURL)

	// Prepare request body
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "projects",
			"id":   projectID,
		},
	}

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to mark project %s: %w", projectID, err)
	}

	return nil
}

// Unmark unmarks a project.
// Unmarking a project removes it from the user's list of favorite projects.
//
// Endpoint: POST /projects/{projectId}/actions/unmarkProject
//
// Example:
//
//	err := client.Projects.Unmark(ctx, "myproject")
func (s *ProjectService) Unmark(ctx context.Context, projectID string) error {
	if projectID == "" {
		return NewValidationError("projectID", "project ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/actions/unmarkProject",
		s.client.baseURL,
		url.PathEscape(projectID))

	// Prepare request body
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "projects",
			"id":   projectID,
		},
	}

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to unmark project %s: %w", projectID, err)
	}

	return nil
}

// Move moves a project to a different location.
//
// Endpoint: POST /projects/{projectId}/actions/moveProject
//
// Example:
//
//	err := client.Projects.Move(ctx, "myproject", &polarion.MoveProjectRequest{
//	    NewLocation: "/new/location",
//	})
func (s *ProjectService) Move(ctx context.Context, projectID string, req *MoveProjectRequest) error {
	if projectID == "" {
		return NewValidationError("projectID", "project ID is required")
	}

	if req == nil {
		return NewValidationError("req", "move project request is required")
	}

	if req.NewLocation == "" {
		return NewValidationError("newLocation", "new location is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/actions/moveProject",
		s.client.baseURL,
		url.PathEscape(projectID))

	// Prepare request body
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "projects",
			"id":   projectID,
			"attributes": map[string]interface{}{
				"newLocation": req.NewLocation,
			},
		},
	}

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to move project %s: %w", projectID, err)
	}

	return nil
}
