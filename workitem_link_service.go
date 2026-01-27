// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// WorkItemLinkService provides operations for work item links.
type WorkItemLinkService struct {
	project *ProjectClient
}

// newWorkItemLinkService creates a new work item link service.
func newWorkItemLinkService(project *ProjectClient) *WorkItemLinkService {
	return &WorkItemLinkService{
		project: project,
	}
}

// Get retrieves a specific work item link by its ID.
// The link ID format is: "{projectId}/{primaryWorkItemId}/{role}/{secondaryProjectId}/{secondaryWorkItemId}"
//
// Example:
//
//	link, err := project.WorkItemLinks.Get(ctx, "myproject/WI-123/relates_to/myproject/WI-456")
func (s *WorkItemLinkService) Get(ctx context.Context, linkID string, opts ...GetOption) (*WorkItemLink, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/linkedworkitems/%s", s.project.client.baseURL, url.PathEscape(linkID))

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var link WorkItemLink
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &link)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get work item link %s: %w", linkID, err)
	}

	return &link, nil
}

// List retrieves all links for a specific work item.
//
// Example:
//
//	links, err := project.WorkItemLinks.List(ctx, "WI-123")
func (s *WorkItemLinkService) List(ctx context.Context, workItemID string, opts ...QueryOption) ([]WorkItemLink, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Extract work item ID from full ID if needed (e.g., "OP869335/OP869335-34496" -> "OP869335-34496")
	cleanWorkItemID := extractWorkItemID(workItemID)

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/linkedworkitems",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID))

	// Build query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var response struct {
		Data []WorkItemLink `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list work item links for %s: %w", workItemID, err)
	}

	return response.Data, nil
}

// Create creates one or more work item links.
// All links must have the same primary work item ID.
//
// Example using the helper function:
//
//	link := polarion.NewWorkItemLink("relates_to", "MyProject/WI-456", "", false)
//	err := project.WorkItemLinks.Create(ctx, "WI-123", link)
//
// Example with manual construction:
//
//	link := &polarion.WorkItemLink{
//	    Type: "linkedworkitems",
//	    Data: &polarion.WorkItemLinkAttributes{
//	        Role:    "relates_to",
//	        Suspect: false,
//	    },
//	    Relationships: &polarion.LinkedWorkItemRelationships{
//	        WorkItem: &polarion.Relationship{
//	            Data: map[string]interface{}{
//	                "type": "workitems",
//	                "id":   "MyProject/WI-456",
//	            },
//	        },
//	    },
//	}
//	err := project.WorkItemLinks.Create(ctx, "WI-123", link)
func (s *WorkItemLinkService) Create(ctx context.Context, primaryWorkItemID string, links ...*WorkItemLink) error {
	if len(links) == 0 {
		return nil
	}

	// Validate all links
	for i, link := range links {
		if err := s.validateLink(link); err != nil {
			return fmt.Errorf("validation failed for link %d: %w", i, err)
		}
	}

	// Extract work item ID from full ID if needed (e.g., "OP869335/OP869335-34496" -> "OP869335-34496")
	cleanWorkItemID := extractWorkItemID(primaryWorkItemID)

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/linkedworkitems",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID))

	// Prepare request body - ensure relationships are properly set
	// If the link doesn't have relationships set, we need to construct them from the old fields
	requestData := make([]interface{}, len(links))
	for i, link := range links {
		linkData := map[string]interface{}{
			"type": link.Type,
			"attributes": map[string]interface{}{
				"role":    link.Data.Role,
				"suspect": link.Data.Suspect,
			},
		}

		// Add revision if present
		if link.Data.Revision != "" {
			linkData["attributes"].(map[string]interface{})["revision"] = link.Data.Revision
		}

		// Add relationships if present
		if link.Relationships != nil && link.Relationships.WorkItem != nil {
			linkData["relationships"] = link.Relationships
		}

		requestData[i] = linkData
	}

	// Prepare request body
	body := map[string]interface{}{
		"data": requestData,
	}

	// Make request with retry
	var response struct {
		Data []WorkItemLink `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return fmt.Errorf("failed to create work item links: %w", err)
	}

	// Update links with created IDs
	for i, created := range response.Data {
		if i < len(links) {
			links[i].ID = created.ID
			if created.Links != nil {
				links[i].Links = created.Links
			}
		}
	}

	return nil
}

// Update updates a work item link (typically to change the suspect flag).
// The link must have an ID set.
//
// Example:
//
//	link.Data.Suspect = true
//	err := project.WorkItemLinks.Update(ctx, link)
func (s *WorkItemLinkService) Update(ctx context.Context, link *WorkItemLink) error {
	if link.ID == "" {
		return NewValidationError("ID", "work item link ID is required for update")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/linkedworkitems/%s", s.project.client.baseURL, url.PathEscape(link.ID))

	// Prepare request body
	body := map[string]interface{}{
		"data": link,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		// Update the link with the response
		return internalhttp.DecodeDataResponse(resp, link)
	})

	if err != nil {
		return fmt.Errorf("failed to update work item link %s: %w", link.ID, err)
	}

	return nil
}

// Delete deletes one or more work item links by their IDs.
//
// Example:
//
//	err := project.WorkItemLinks.Delete(ctx, "myproject/WI-123/relates_to/myproject/WI-456")
func (s *WorkItemLinkService) Delete(ctx context.Context, linkIDs ...string) error {
	if len(linkIDs) == 0 {
		return nil
	}

	// Group links by primary work item for batch deletion
	linksByWorkItem := make(map[string][]string)
	for _, linkID := range linkIDs {
		// Extract primary work item ID from link ID
		parts := strings.Split(linkID, "/")
		if len(parts) >= 2 {
			primaryWorkItemID := parts[0] + "/" + parts[1]
			linksByWorkItem[primaryWorkItemID] = append(linksByWorkItem[primaryWorkItemID], linkID)
		}
	}

	// Delete links for each work item
	for primaryWorkItemID, links := range linksByWorkItem {
		if err := s.deleteBatch(ctx, primaryWorkItemID, links); err != nil {
			return err
		}
	}

	return nil
}

// deleteBatch deletes a batch of links for a specific work item.
func (s *WorkItemLinkService) deleteBatch(ctx context.Context, primaryWorkItemID string, linkIDs []string) error {
	// Extract work item ID from full ID if needed (e.g., "OP869335/OP869335-34496" -> "OP869335-34496")
	cleanWorkItemID := extractWorkItemID(primaryWorkItemID)

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/linkedworkitems",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID))

	// Prepare request body with link IDs
	linkData := make([]map[string]interface{}, len(linkIDs))
	for i, linkID := range linkIDs {
		linkData[i] = map[string]interface{}{
			"type": "linkedworkitems",
			"id":   linkID,
		}
	}

	body := map[string]interface{}{
		"data": linkData,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete work item links: %w", err)
	}

	return nil
}

// validateLink validates a work item link before creation or update.
func (s *WorkItemLinkService) validateLink(link *WorkItemLink) error {
	if link == nil {
		return NewValidationError("link", "work item link cannot be nil")
	}

	if link.Data == nil {
		return NewValidationError("attributes", "work item link attributes cannot be nil")
	}

	if link.Data.Role == "" {
		return NewValidationError("role", "work item link role is required")
	}

	// Set type if not set
	if link.Type == "" {
		link.Type = "linkedworkitems"
	}

	return nil
}

// buildWorkItemID builds a full work item ID with project prefix if needed.
// This is used for building request bodies, not URLs.
func (s *WorkItemLinkService) buildWorkItemID(id string) string {
	// If ID already contains project prefix, return as-is
	if strings.Contains(id, "/") {
		return id
	}
	// Otherwise, prepend project ID
	return fmt.Sprintf("%s/%s", s.project.projectID, id)
}
