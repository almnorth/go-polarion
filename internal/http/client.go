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
// It adds the Bearer token and sets appropriate headers for JSON.
func (c *client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original
	req = req.Clone(ctx)

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	// Set JSON headers if not already set
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

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
// The Pointer field typically contains a JSON pointer to the field that caused the error,
// e.g., "/data/0/attributes/customFields/myField" indicates an issue with the custom field "myField".
type ErrorDetail struct {
	Status  string `json:"status"`
	Title   string `json:"title,omitempty"`
	Detail  string `json:"detail"`
	Pointer string `json:"pointer,omitempty"` // JSON pointer to the problematic field
}

// String returns a string representation of the error detail.
func (e ErrorDetail) String() string {
	if e.Pointer != "" {
		return fmt.Sprintf("[%s] %s (at %s)", e.Status, e.Detail, e.Pointer)
	}
	if e.Title != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Status, e.Title, e.Detail)
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

	// Store raw body for debugging
	rawBody := string(body)

	// Try to parse JSON:API error format
	var errorResponse struct {
		Errors []ErrorDetail `json:"errors"`
	}

	if err := json.Unmarshal(body, &errorResponse); err != nil {
		// If parsing fails, return the raw body as the message
		// This helps when the API returns non-JSON:API formatted errors
		return newAPIError(resp.StatusCode, rawBody, resp)
	}

	// If we got errors, use them
	if len(errorResponse.Errors) > 0 {
		apiErr := newAPIError(resp.StatusCode, resp.Status, resp)
		apiErr.Details = errorResponse.Errors
		apiErr.RawBody = rawBody
		return apiErr
	}

	// If no errors in the response, return the raw body
	return newAPIError(resp.StatusCode, rawBody, resp)
}

// APIError represents an error from the Polarion API.
type APIError struct {
	StatusCode int
	Message    string
	Response   *http.Response
	Details    []ErrorDetail
	RawBody    string // Raw response body for debugging
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
	method := ""
	url := ""
	if e.Response != nil && e.Response.Request != nil {
		method = e.Response.Request.Method
		url = e.Response.Request.URL.String()
	}

	if len(e.Details) > 0 {
		// Format error details in a more readable way
		detailsStr := ""
		for i, detail := range e.Details {
			if i > 0 {
				detailsStr += "; "
			}
			if detail.Pointer != "" {
				// Extract field name from JSON pointer (e.g., "/data/0/attributes/fieldName" -> "fieldName")
				detailsStr += fmt.Sprintf("field '%s': %s", detail.Pointer, detail.Detail)
			} else if detail.Title != "" {
				detailsStr += fmt.Sprintf("%s: %s", detail.Title, detail.Detail)
			} else {
				detailsStr += detail.Detail
			}
		}
		return fmt.Sprintf("polarion api error (status %d) for %s %s: %s - %s",
			e.StatusCode, method, url, e.Message, detailsStr)
	}
	return fmt.Sprintf("polarion api error (status %d) for %s %s: %s",
		e.StatusCode, method, url, e.Message)
}

// GetDetailedError returns a detailed error message including the raw response body.
// This is useful for debugging when the standard error message is not sufficient.
func (e *APIError) GetDetailedError() string {
	baseError := e.Error()
	if e.RawBody != "" && len(e.RawBody) < 1000 {
		// Only include raw body if it's not too large
		return fmt.Sprintf("%s\nRaw response: %s", baseError, e.RawBody)
	}
	return baseError
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

// DoRequestWithAccept is a helper function to make HTTP requests with a custom Accept header.
// This is useful for endpoints that don't support JSON:API format (e.g., enumerations).
func DoRequestWithAccept(ctx context.Context, client Client, method, url, acceptHeader string, body interface{}) (*http.Response, error) {
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

	// Set custom Accept header
	req.Header.Set("Accept", acceptHeader)

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
