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

// WorkItemCommentService provides operations for managing work item comments.
// Comments are project-scoped and belong to specific work items.
type WorkItemCommentService struct {
	project *ProjectClient
}

// newWorkItemCommentService creates a new work item comment service.
func newWorkItemCommentService(project *ProjectClient) *WorkItemCommentService {
	return &WorkItemCommentService{
		project: project,
	}
}

// Get retrieves a specific comment by ID.
//
// Example:
//
//	comment, err := project.WorkItemComments.Get(ctx, "WI-123", "comment-456")
func (s *WorkItemCommentService) Get(ctx context.Context, workItemID, commentID string, opts ...GetOption) (*WorkItemComment, error) {
	if workItemID == "" {
		return nil, fmt.Errorf("workItemID cannot be empty")
	}
	if commentID == "" {
		return nil, fmt.Errorf("commentID cannot be empty")
	}

	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Extract work item ID from full ID if needed
	cleanWorkItemID := extractWorkItemID(workItemID)

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/comments/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID),
		url.PathEscape(commentID))

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
	var comment WorkItemComment
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &comment)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get comment %s for work item %s: %w", commentID, workItemID, err)
	}

	return &comment, nil
}

// List retrieves all comments for a work item.
//
// Example:
//
//	comments, err := project.WorkItemComments.List(ctx, "WI-123")
func (s *WorkItemCommentService) List(ctx context.Context, workItemID string, opts ...QueryOption) ([]*WorkItemComment, error) {
	if workItemID == "" {
		return nil, fmt.Errorf("workItemID cannot be empty")
	}

	// Apply options
	options := defaultQueryOptions()
	options.pageSize = s.project.client.config.pageSize
	for _, opt := range opts {
		opt(&options)
	}

	// Extract work item ID from full ID if needed
	cleanWorkItemID := extractWorkItemID(workItemID)

	var allComments []*WorkItemComment
	pageNum := 1

	for {
		// Build URL
		urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/comments",
			s.project.client.baseURL,
			url.PathEscape(s.project.projectID),
			url.PathEscape(cleanWorkItemID))

		// Build query parameters
		params := url.Values{}

		// Set page size
		pageSize := options.pageSize
		if pageSize <= 0 {
			pageSize = s.project.client.config.pageSize
		}
		params.Set("page[size]", strconv.Itoa(pageSize))
		params.Set("page[number]", strconv.Itoa(pageNum))

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
			Data  []WorkItemComment `json:"data"`
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
			return nil, fmt.Errorf("failed to list comments for work item %s: %w", workItemID, err)
		}

		// Append comments from this page
		for i := range response.Data {
			allComments = append(allComments, &response.Data[i])
		}

		// Check if there are more pages
		if response.Links.Next == "" {
			break
		}

		pageNum++
	}

	return allComments, nil
}

// Create creates one or more comments on a work item.
//
// Example:
//
//	comment := &polarion.WorkItemComment{
//	    Type: "workitem_comments",
//	    Attributes: &polarion.WorkItemCommentAttributes{
//	        Text: polarion.NewHTMLContent("<p>This is a comment</p>"),
//	    },
//	}
//	created, err := project.WorkItemComments.Create(ctx, "WI-123", comment)
func (s *WorkItemCommentService) Create(ctx context.Context, workItemID string, comments ...*WorkItemComment) ([]*WorkItemComment, error) {
	if workItemID == "" {
		return nil, fmt.Errorf("workItemID cannot be empty")
	}
	if len(comments) == 0 {
		return nil, fmt.Errorf("at least one comment must be provided")
	}

	// Extract work item ID from full ID if needed
	cleanWorkItemID := extractWorkItemID(workItemID)

	// Prepare request body
	body := map[string]interface{}{
		"data": comments,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/comments",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID))

	// Make request with retry
	var response struct {
		Data []WorkItemComment `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create comments for work item %s: %w", workItemID, err)
	}

	// Convert to pointers
	var createdComments []*WorkItemComment
	for i := range response.Data {
		createdComments = append(createdComments, &response.Data[i])
	}

	return createdComments, nil
}

// Update updates an existing comment.
//
// Example:
//
//	comment.Attributes.Text = polarion.NewHTMLContent("<p>Updated comment</p>")
//	err := project.WorkItemComments.Update(ctx, "WI-123", comment)
func (s *WorkItemCommentService) Update(ctx context.Context, workItemID string, comment *WorkItemComment) error {
	if workItemID == "" {
		return fmt.Errorf("workItemID cannot be empty")
	}
	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}
	if comment.ID == "" {
		return fmt.Errorf("comment ID cannot be empty")
	}

	// Extract work item ID from full ID if needed
	cleanWorkItemID := extractWorkItemID(workItemID)

	// Prepare request body
	body := map[string]interface{}{
		"data": comment,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/comments/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID),
		url.PathEscape(comment.ID))

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
		return fmt.Errorf("failed to update comment %s for work item %s: %w", comment.ID, workItemID, err)
	}

	return nil
}
