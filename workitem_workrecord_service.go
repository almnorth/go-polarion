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

// WorkItemWorkRecordService provides operations for work item work records (time tracking).
type WorkItemWorkRecordService struct {
	project *ProjectClient
}

// newWorkItemWorkRecordService creates a new work item work record service.
func newWorkItemWorkRecordService(project *ProjectClient) *WorkItemWorkRecordService {
	return &WorkItemWorkRecordService{
		project: project,
	}
}

// Get retrieves a specific work record by ID.
//
// Example:
//
//	record, err := project.WorkItemWorkRecords.Get(ctx, "WI-123", "record-id")
func (s *WorkItemWorkRecordService) Get(ctx context.Context, workItemID, recordID string, opts ...GetOption) (*WorkRecord, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/workrecords/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(recordID))

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
	var record WorkRecord
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &record)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get work record %s for work item %s: %w", recordID, workItemID, err)
	}

	return &record, nil
}

// List retrieves all work records for a work item with pagination.
//
// Example:
//
//	records, hasNext, err := project.WorkItemWorkRecords.List(ctx, "WI-123",
//	    polarion.WithQueryPageSize(50), polarion.WithPageNumber(1))
func (s *WorkItemWorkRecordService) List(ctx context.Context, workItemID string, opts ...QueryOption) ([]WorkRecord, bool, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/workrecords",
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
		Data  []WorkRecord `json:"data"`
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
		return nil, false, fmt.Errorf("failed to list work records for work item %s: %w", workItemID, err)
	}

	return response.Data, response.Links.Next != "", nil
}

// Create logs time on a work item by creating one or more work records.
//
// Example:
//
//	req := polarion.NewWorkRecordRequest("user-id", time.Now(), polarion.NewTimeSpent(2, 30)).
//	    WithComment("Implemented feature X")
//	err := project.WorkItemWorkRecords.Create(ctx, "WI-123", req)
func (s *WorkItemWorkRecordService) Create(ctx context.Context, workItemID string, requests ...*WorkRecordCreateRequest) error {
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
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/workrecords",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Prepare request body
	data := make([]map[string]interface{}, len(requests))
	for i, req := range requests {
		attributes := map[string]interface{}{
			"date":      req.Date.Format("2006-01-02"),
			"timeSpent": req.TimeSpent.String(),
		}
		if req.Comment != "" {
			attributes["comment"] = map[string]interface{}{
				"type":  "text/plain",
				"value": req.Comment,
			}
		}

		data[i] = map[string]interface{}{
			"type":       "workrecords",
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
		return fmt.Errorf("failed to create work records for work item %s: %w", workItemID, err)
	}

	return nil
}

// Delete deletes one or more work records from a work item.
//
// Example:
//
//	err := project.WorkItemWorkRecords.Delete(ctx, "WI-123", "record-id-1", "record-id-2")
func (s *WorkItemWorkRecordService) Delete(ctx context.Context, workItemID string, recordIDs ...string) error {
	if len(recordIDs) == 0 {
		return nil
	}

	// Delete each work record
	for _, recordID := range recordIDs {
		urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/workrecords/%s",
			s.project.client.baseURL,
			url.PathEscape(s.project.projectID),
			url.PathEscape(workItemID),
			url.PathEscape(recordID))

		err := s.project.client.retrier.Do(ctx, func() error {
			resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, nil)
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to delete work record %s: %w", recordID, err)
		}
	}

	return nil
}

// validateCreateRequest validates a work record creation request.
func (s *WorkItemWorkRecordService) validateCreateRequest(req *WorkRecordCreateRequest) error {
	if req == nil {
		return NewValidationError("request", "create request cannot be nil")
	}

	if req.UserID == "" {
		return NewValidationError("userID", "user ID is required")
	}

	if req.Date.IsZero() {
		return NewValidationError("date", "date is required")
	}

	if req.TimeSpent.Hours == 0 && req.TimeSpent.Minutes == 0 {
		return NewValidationError("timeSpent", "time spent must be greater than zero")
	}

	return nil
}
