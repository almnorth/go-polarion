// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"bytes"
	"context"
	"fmt"
	internalhttp "github.com/almnorth/go-polarion/internal/http"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// UserService provides operations for managing Polarion users.
// Users are global resources, not project-scoped.
type UserService struct {
	client *Client
}

// newUserService creates a new user service.
func newUserService(client *Client) *UserService {
	return &UserService{
		client: client,
	}
}

// Get retrieves a specific user by ID.
//
// Example:
//
//	user, err := client.Users.Get(ctx, "user123")
func (s *UserService) Get(ctx context.Context, userID string, opts ...GetOption) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/users/%s", s.client.baseURL, url.PathEscape(userID))

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
	var user User
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &user)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
	}

	return &user, nil
}

// List retrieves a list of users with optional query parameters.
//
// Example:
//
//	users, err := client.Users.List(ctx, polarion.WithQuery("disabled:false"))
func (s *UserService) List(ctx context.Context, opts ...QueryOption) ([]*User, error) {
	// Apply options
	options := defaultQueryOptions()
	options.pageSize = s.client.config.pageSize
	for _, opt := range opts {
		opt(&options)
	}

	var allUsers []*User
	pageNum := 1

	for {
		// Build URL
		urlStr := fmt.Sprintf("%s/users", s.client.baseURL)

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
			Data  []User `json:"data"`
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
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		// Append users from this page
		for i := range response.Data {
			allUsers = append(allUsers, &response.Data[i])
		}

		// Check if there are more pages
		if response.Links.Next == "" {
			break
		}

		pageNum++
	}

	return allUsers, nil
}

// Create creates one or more users.
//
// Example:
//
//	user := &polarion.User{
//	    Type: "users",
//	    ID: "newuser",
//	    Attributes: &polarion.UserAttributes{
//	        Name: "New User",
//	        Email: "newuser@example.com",
//	    },
//	}
//	created, err := client.Users.Create(ctx, user)
func (s *UserService) Create(ctx context.Context, users ...*User) ([]*User, error) {
	if len(users) == 0 {
		return nil, fmt.Errorf("at least one user must be provided")
	}

	// Prepare request body
	body := map[string]interface{}{
		"data": users,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/users", s.client.baseURL)

	// Make request with retry
	var response struct {
		Data []User `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create users: %w", err)
	}

	// Convert to pointers
	var createdUsers []*User
	for i := range response.Data {
		createdUsers = append(createdUsers, &response.Data[i])
	}

	return createdUsers, nil
}

// Update updates an existing user.
//
// Example:
//
//	user.Attributes.Name = "Updated Name"
//	err := client.Users.Update(ctx, user)
func (s *UserService) Update(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	if user.ID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Prepare request body
	body := map[string]interface{}{
		"data": user,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/users/%s", s.client.baseURL, url.PathEscape(user.ID))

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update user %s: %w", user.ID, err)
	}

	return nil
}

// GetAvatar retrieves a user's avatar image.
//
// Example:
//
//	avatar, err := client.Users.GetAvatar(ctx, "user123")
func (s *UserService) GetAvatar(ctx context.Context, userID string) (*UserAvatar, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/users/%s/avatar", s.client.baseURL, url.PathEscape(userID))

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make request with retry
	var avatar *UserAvatar
	err = s.client.retrier.Do(ctx, func() error {
		resp, err := s.client.httpClient.Do(ctx, req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		avatar = &UserAvatar{
			Data:        data,
			ContentType: resp.Header.Get("Content-Type"),
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get avatar for user %s: %w", userID, err)
	}

	return avatar, nil
}

// UpdateAvatar updates a user's avatar image.
//
// Example:
//
//	avatarData, _ := os.ReadFile("avatar.png")
//	err := client.Users.UpdateAvatar(ctx, "user123", avatarData, "image/png")
func (s *UserService) UpdateAvatar(ctx context.Context, userID string, avatarData []byte, contentType string) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}
	if len(avatarData) == 0 {
		return fmt.Errorf("avatarData cannot be empty")
	}
	if contentType == "" {
		contentType = "image/png" // Default to PNG
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/users/%s/avatar", s.client.baseURL, url.PathEscape(userID))

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, bytes.NewReader(avatarData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", contentType)

		resp, err := s.client.httpClient.Do(ctx, req)
		if err != nil {
			return err
		}
		resp.Body.Close()

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update avatar for user %s: %w", userID, err)
	}

	return nil
}

// SetLicense sets a license for a user.
//
// Example:
//
//	license := &polarion.License{
//	    Type: "licenses",
//	    ID: "developer",
//	}
//	err := client.Users.SetLicense(ctx, "user123", license)
func (s *UserService) SetLicense(ctx context.Context, userID string, license *License) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}
	if license == nil {
		return fmt.Errorf("license cannot be nil")
	}

	// Prepare request body
	body := map[string]interface{}{
		"data": license,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/users/%s/relationships/license", s.client.baseURL, url.PathEscape(userID))

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to set license for user %s: %w", userID, err)
	}

	return nil
}
