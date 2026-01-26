// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	internalhttp "github.com/almnorth/go-polarion/internal/http"
	"net/url"
	"strconv"
)

// UserGroupService provides operations for managing Polarion user groups.
// User groups are global resources, not project-scoped.
type UserGroupService struct {
	client *Client
}

// newUserGroupService creates a new user group service.
func newUserGroupService(client *Client) *UserGroupService {
	return &UserGroupService{
		client: client,
	}
}

// Get retrieves a specific user group by ID.
//
// Example:
//
//	group, err := client.UserGroups.Get(ctx, "developers")
func (s *UserGroupService) Get(ctx context.Context, groupID string, opts ...GetOption) (*UserGroup, error) {
	if groupID == "" {
		return nil, fmt.Errorf("groupID cannot be empty")
	}

	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/usergroups/%s", s.client.baseURL, url.PathEscape(groupID))

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
	var group UserGroup
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &group)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get user group %s: %w", groupID, err)
	}

	return &group, nil
}

// List retrieves a list of user groups with optional query parameters.
//
// Example:
//
//	groups, err := client.UserGroups.List(ctx)
func (s *UserGroupService) List(ctx context.Context, opts ...QueryOption) ([]*UserGroup, error) {
	// Apply options
	options := defaultQueryOptions()
	options.pageSize = s.client.config.pageSize
	for _, opt := range opts {
		opt(&options)
	}

	var allGroups []*UserGroup
	pageNum := 1

	for {
		// Build URL
		urlStr := fmt.Sprintf("%s/usergroups", s.client.baseURL)

		// Build query parameters
		params := url.Values{}
		if options.query != "" {
			params.Set("query", options.query)
		}

		// Set page size
		pageSize := options.pageSize
		if pageSize <= 0 {
			pageSize = s.client.config.pageSize
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
			Data  []UserGroup `json:"data"`
			Links struct {
				Next string `json:"next,omitempty"`
			} `json:"links"`
		}

		err := s.client.retrier.Do(ctx, func() error {
			resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
			if err != nil {
				return err
			}
			return internalhttp.DecodeResponse(resp, &response)
		})

		if err != nil {
			return nil, fmt.Errorf("failed to list user groups: %w", err)
		}

		// Append groups from this page
		for i := range response.Data {
			allGroups = append(allGroups, &response.Data[i])
		}

		// Check if there are more pages
		if response.Links.Next == "" {
			break
		}

		pageNum++
	}

	return allGroups, nil
}
