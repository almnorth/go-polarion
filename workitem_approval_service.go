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

// WorkItemApprovalService provides operations for work item approvals.
type WorkItemApprovalService struct {
	project *ProjectClient
}

// newWorkItemApprovalService creates a new work item approval service.
func newWorkItemApprovalService(project *ProjectClient) *WorkItemApprovalService {
	return &WorkItemApprovalService{
		project: project,
	}
}

// Get retrieves a specific approval by user ID.
//
// Example:
//
//	approval, err := project.WorkItemApprovals.Get(ctx, "WI-123", "user-id")
func (s *WorkItemApprovalService) Get(ctx context.Context, workItemID, userID string, opts ...GetOption) (*WorkItemApproval, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/approvals/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(userID))

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if options.revision != "" {
		params.Set("revision", options.revision)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var approval WorkItemApproval
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &approval)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get approval for user %s on work item %s: %w", userID, workItemID, err)
	}

	return &approval, nil
}

// List retrieves all approvals for a work item with pagination.
//
// Example:
//
//	approvals, hasNext, err := project.WorkItemApprovals.List(ctx, "WI-123",
//	    polarion.WithQueryPageSize(50), polarion.WithPageNumber(1))
func (s *WorkItemApprovalService) List(ctx context.Context, workItemID string, opts ...QueryOption) ([]WorkItemApproval, bool, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/approvals",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Build query parameters
	params := url.Values{}

	// Set page size
	pageSize := options.pageSize
	if pageSize <= 0 {
		pageSize = s.project.client.config.pageSize
	}
	params.Set("page[size]", strconv.Itoa(pageSize))

	// Set page number
	pageNumber := options.pageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	params.Set("page[number]", strconv.Itoa(pageNumber))

	// Add field selection
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}

	// Add revision if specified
	if options.revision != "" {
		params.Set("revision", options.revision)
	}

	urlStr += "?" + params.Encode()

	// Make request with retry
	var response struct {
		Data  []WorkItemApproval `json:"data"`
		Links struct {
			Next string `json:"next,omitempty"`
		} `json:"links"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, false, fmt.Errorf("failed to list approvals for work item %s: %w", workItemID, err)
	}

	return response.Data, response.Links.Next != "", nil
}

// Create requests approvals from one or more users.
//
// Example:
//
//	req := polarion.NewApprovalRequest("user-id").WithComment("Please review")
//	err := project.WorkItemApprovals.Create(ctx, "WI-123", req)
func (s *WorkItemApprovalService) Create(ctx context.Context, workItemID string, requests ...*ApprovalCreateRequest) error {
	if len(requests) == 0 {
		return nil
	}

	// Validate requests
	for i, req := range requests {
		if err := s.validateCreateRequest(req); err != nil {
			return fmt.Errorf("validation failed for request %d: %w", i, err)
		}
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/approvals",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Prepare request body
	data := make([]map[string]interface{}, len(requests))
	for i, req := range requests {
		attributes := map[string]interface{}{
			"status": req.Status,
		}
		if req.Comment != "" {
			attributes["comment"] = req.Comment
		}

		data[i] = map[string]interface{}{
			"type":       "workitem_approvals",
			"attributes": attributes,
			"relationships": map[string]interface{}{
				"user": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "users",
						"id":   req.UserID,
					},
				},
			},
		}
	}

	body := map[string]interface{}{
		"data": data,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create approvals for work item %s: %w", workItemID, err)
	}

	return nil
}

// Update updates a single approval status.
//
// Example:
//
//	update := polarion.NewApprovalUpdate("user-id", polarion.ApprovalStatusApproved).
//	    WithUpdateComment("Looks good")
//	err := project.WorkItemApprovals.Update(ctx, "WI-123", update)
func (s *WorkItemApprovalService) Update(ctx context.Context, workItemID string, request *ApprovalUpdateRequest) error {
	if request == nil {
		return NewValidationError("request", "update request cannot be nil")
	}

	if request.UserID == "" {
		return NewValidationError("userID", "user ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/approvals/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(request.UserID))

	// Prepare request body
	attributes := map[string]interface{}{
		"status": request.Status,
	}
	if request.Comment != "" {
		attributes["comment"] = request.Comment
	}

	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "workitem_approvals",
			"id":         fmt.Sprintf("%s/%s/%s", s.project.projectID, workItemID, request.UserID),
			"attributes": attributes,
		},
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update approval for user %s: %w", request.UserID, err)
	}

	return nil
}

// UpdateBatch updates multiple approvals at once.
//
// Example:
//
//	updates := []*polarion.ApprovalUpdateRequest{
//	    polarion.NewApprovalUpdate("user1", polarion.ApprovalStatusApproved),
//	    polarion.NewApprovalUpdate("user2", polarion.ApprovalStatusApproved),
//	}
//	err := project.WorkItemApprovals.UpdateBatch(ctx, "WI-123", updates...)
func (s *WorkItemApprovalService) UpdateBatch(ctx context.Context, workItemID string, requests ...*ApprovalUpdateRequest) error {
	if len(requests) == 0 {
		return nil
	}

	// Validate requests
	for i, req := range requests {
		if req == nil {
			return fmt.Errorf("request %d is nil", i)
		}
		if req.UserID == "" {
			return fmt.Errorf("request %d: user ID is required", i)
		}
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/approvals",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Prepare request body
	data := make([]map[string]interface{}, len(requests))
	for i, req := range requests {
		attributes := map[string]interface{}{
			"status": req.Status,
		}
		if req.Comment != "" {
			attributes["comment"] = req.Comment
		}

		data[i] = map[string]interface{}{
			"type":       "workitem_approvals",
			"id":         fmt.Sprintf("%s/%s/%s", s.project.projectID, workItemID, req.UserID),
			"attributes": attributes,
		}
	}

	body := map[string]interface{}{
		"data": data,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update approvals for work item %s: %w", workItemID, err)
	}

	return nil
}

// Delete removes approvals from one or more users.
//
// Example:
//
//	err := project.WorkItemApprovals.Delete(ctx, "WI-123", "user-id-1", "user-id-2")
func (s *WorkItemApprovalService) Delete(ctx context.Context, workItemID string, userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Delete each approval
	for _, userID := range userIDs {
		urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/approvals/%s",
			s.project.client.baseURL,
			url.PathEscape(s.project.projectID),
			url.PathEscape(workItemID),
			url.PathEscape(userID))

		err := s.project.client.retrier.Do(ctx, func() error {
			resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, nil)
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to delete approval for user %s: %w", userID, err)
		}
	}

	return nil
}

// validateCreateRequest validates an approval creation request.
func (s *WorkItemApprovalService) validateCreateRequest(req *ApprovalCreateRequest) error {
	if req == nil {
		return NewValidationError("request", "create request cannot be nil")
	}

	if req.UserID == "" {
		return NewValidationError("userID", "user ID is required")
	}

	if req.Status == "" {
		req.Status = ApprovalStatusWaiting
	}

	return nil
}
