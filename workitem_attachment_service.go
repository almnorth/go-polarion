// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// WorkItemAttachmentService provides operations for work item attachments.
type WorkItemAttachmentService struct {
	project *ProjectClient
}

// newWorkItemAttachmentService creates a new work item attachment service.
func newWorkItemAttachmentService(project *ProjectClient) *WorkItemAttachmentService {
	return &WorkItemAttachmentService{
		project: project,
	}
}

// Get retrieves a specific attachment by ID.
//
// Example:
//
//	attachment, err := project.WorkItemAttachments.Get(ctx, "WI-123", "attachment-id")
func (s *WorkItemAttachmentService) Get(ctx context.Context, workItemID, attachmentID string, opts ...GetOption) (*WorkItemAttachment, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/attachments/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(attachmentID))

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
	var attachment WorkItemAttachment
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &attachment)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get attachment %s for work item %s: %w", attachmentID, workItemID, err)
	}

	return &attachment, nil
}

// List retrieves all attachments for a work item with pagination.
//
// Example:
//
//	attachments, hasNext, err := project.WorkItemAttachments.List(ctx, "WI-123",
//	    polarion.WithPageSize(50), polarion.WithPageNumber(1))
func (s *WorkItemAttachmentService) List(ctx context.Context, workItemID string, opts ...QueryOption) ([]WorkItemAttachment, bool, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/attachments",
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
		Data  []WorkItemAttachment `json:"data"`
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
		return nil, false, fmt.Errorf("failed to list attachments for work item %s: %w", workItemID, err)
	}

	return response.Data, response.Links.Next != "", nil
}

// GetContent downloads the content of an attachment.
// Returns an io.ReadCloser that must be closed by the caller.
//
// Example:
//
//	content, err := project.WorkItemAttachments.GetContent(ctx, "WI-123", "attachment-id")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer content.Close()
//	data, err := io.ReadAll(content)
func (s *WorkItemAttachmentService) GetContent(ctx context.Context, workItemID, attachmentID string) (io.ReadCloser, error) {
	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/attachments/%s/content",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(attachmentID))

	// Make request with retry
	var content io.ReadCloser
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		content = resp.Body
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get content for attachment %s: %w", attachmentID, err)
	}

	return content, nil
}

// Create uploads one or more attachments to a work item.
//
// Example:
//
//	req := polarion.NewAttachmentCreateRequest("document.pdf", fileBytes, "application/pdf").
//	    WithTitle("Requirements Document")
//	err := project.WorkItemAttachments.Create(ctx, "WI-123", req)
func (s *WorkItemAttachmentService) Create(ctx context.Context, workItemID string, requests ...*AttachmentCreateRequest) error {
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
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/attachments",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Create multipart request
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoMultipartRequest(ctx, s.project.client.httpClient, "POST", urlStr, requests)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create attachments for work item %s: %w", workItemID, err)
	}

	return nil
}

// Update updates an attachment's metadata and optionally its content.
//
// Example:
//
//	req := &polarion.AttachmentUpdateRequest{
//	    AttachmentID: "attachment-id",
//	    Title:        "Updated Title",
//	}
//	err := project.WorkItemAttachments.Update(ctx, "WI-123", req)
func (s *WorkItemAttachmentService) Update(ctx context.Context, workItemID string, request *AttachmentUpdateRequest) error {
	if request == nil {
		return NewValidationError("request", "update request cannot be nil")
	}

	if request.AttachmentID == "" {
		return NewValidationError("attachmentID", "attachment ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/attachments/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(request.AttachmentID))

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoMultipartUpdateRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, request)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update attachment %s: %w", request.AttachmentID, err)
	}

	return nil
}

// Delete deletes one or more attachments from a work item.
//
// Example:
//
//	err := project.WorkItemAttachments.Delete(ctx, "WI-123", "attachment-id-1", "attachment-id-2")
func (s *WorkItemAttachmentService) Delete(ctx context.Context, workItemID string, attachmentIDs ...string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	// Delete each attachment
	for _, attachmentID := range attachmentIDs {
		urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/attachments/%s",
			s.project.client.baseURL,
			url.PathEscape(s.project.projectID),
			url.PathEscape(workItemID),
			url.PathEscape(attachmentID))

		err := s.project.client.retrier.Do(ctx, func() error {
			resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, nil)
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to delete attachment %s: %w", attachmentID, err)
		}
	}

	return nil
}

// validateCreateRequest validates an attachment creation request.
func (s *WorkItemAttachmentService) validateCreateRequest(req *AttachmentCreateRequest) error {
	if req == nil {
		return NewValidationError("request", "create request cannot be nil")
	}

	if req.FileName == "" {
		return NewValidationError("fileName", "file name is required")
	}

	if len(req.Content) == 0 {
		return NewValidationError("content", "file content is required")
	}

	if req.ContentType == "" {
		return NewValidationError("contentType", "content type is required")
	}

	return nil
}

// buildWorkItemID builds a full work item ID with project prefix if needed.
func (s *WorkItemAttachmentService) buildWorkItemID(id string) string {
	// If ID already contains project prefix, return as-is
	if len(id) > 0 && id[0] != '/' {
		return fmt.Sprintf("%s/%s", s.project.projectID, id)
	}
	return id
}
