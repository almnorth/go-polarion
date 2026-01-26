// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package http provides internal HTTP client functionality for the Polarion API client.
// This package is internal and not part of the public API.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client defines the interface for making HTTP requests.
// This interface allows for easy mocking in tests.
type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// client wraps http.Client with authentication and JSON:API support.
type client struct {
	httpClient  *http.Client
	bearerToken string
}

// NewClient creates a new HTTP client with Bearer token authentication.
func NewClient(httpClient *http.Client, bearerToken string) Client {
	return &client{
		httpClient:  httpClient,
		bearerToken: bearerToken,
	}
}

// Do executes an HTTP request with authentication headers.
// It adds the Bearer token and sets appropriate headers for JSON:API.
func (c *client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original
	req = req.Clone(ctx)

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	// Set JSON:API headers
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// Check for API errors
	if resp.StatusCode >= 400 {
		return resp, c.parseAPIError(resp)
	}

	return resp, nil
}

// ErrorDetail represents a single error detail from the Polarion API.
// This follows the JSON:API error object specification.
type ErrorDetail struct {
	Status  string `json:"status"`
	Title   string `json:"title,omitempty"`
	Detail  string `json:"detail"`
	Pointer string `json:"pointer,omitempty"`
}

// String returns a string representation of the error detail.
func (e ErrorDetail) String() string {
	if e.Pointer != "" {
		return fmt.Sprintf("[%s] %s (at %s)", e.Status, e.Detail, e.Pointer)
	}
	return fmt.Sprintf("[%s] %s", e.Status, e.Detail)
}

// parseAPIError parses an error response from the Polarion API.
// It attempts to extract error details from the JSON:API error format.
func (c *client) parseAPIError(resp *http.Response) error {
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return newAPIError(resp.StatusCode, "failed to read error response", resp)
	}

	// Try to parse JSON:API error format
	var errorResponse struct {
		Errors []ErrorDetail `json:"errors"`
	}

	if err := json.Unmarshal(body, &errorResponse); err != nil {
		// If parsing fails, return the raw body as the message
		return newAPIError(resp.StatusCode, string(body), resp)
	}

	// Create API error with details
	apiErr := newAPIError(resp.StatusCode, resp.Status, resp)
	apiErr.Details = errorResponse.Errors

	return apiErr
}

// APIError represents an error from the Polarion API.
type APIError struct {
	StatusCode int
	Message    string
	Response   *http.Response
	Details    []ErrorDetail
}

// newAPIError creates a new API error.
func newAPIError(statusCode int, message string, resp *http.Response) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Response:   resp,
	}
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("polarion api error (status %d): %s - %v",
			e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("polarion api error (status %d): %s",
		e.StatusCode, e.Message)
}

// DoRequest is a helper function to make HTTP requests with JSON encoding/decoding.
func DoRequest(ctx context.Context, client Client, method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return client.Do(ctx, req)
}

// DecodeResponse decodes a JSON:API response into the target struct.
func DecodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// DecodeDataResponse decodes a JSON:API response with a "data" wrapper.
func DecodeDataResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("failed to decode response wrapper: %w", err)
	}

	if err := json.Unmarshal(wrapper.Data, target); err != nil {
		return fmt.Errorf("failed to decode response data: %w", err)
	}

	return nil
}

// DoMultipartRequest makes a multipart form request for file uploads.
// The requests parameter should contain attachment creation requests.
func DoMultipartRequest(ctx context.Context, client Client, method, url string, requests interface{}) (*http.Response, error) {
	// For now, we'll use a simplified approach with JSON:API format
	// In a full implementation, this would create a proper multipart/form-data request
	// with file content and JSON metadata

	// Build the JSON:API request body
	body := map[string]interface{}{
		"data": requests,
	}

	return DoRequest(ctx, client, method, url, body)
}

// DoMultipartUpdateRequest makes a multipart form request for updating attachments.
func DoMultipartUpdateRequest(ctx context.Context, client Client, method, url string, request interface{}) (*http.Response, error) {
	// Build the JSON:API request body
	body := map[string]interface{}{
		"data": request,
	}

	return DoRequest(ctx, client, method, url, body)
}
