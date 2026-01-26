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

// TestParameterService handles test parameter definition operations for a project.
// Test parameters define configurable values that can be used in test cases.
type TestParameterService struct {
	client    *Client
	projectID string
}

// newTestParameterService creates a new test parameter service.
func newTestParameterService(client *Client, projectID string) *TestParameterService {
	return &TestParameterService{
		client:    client,
		projectID: projectID,
	}
}

// Get retrieves a specific test parameter definition.
//
// Endpoint: GET /projects/{projectId}/testparameterdefinitions/{testParamId}
//
// Example:
//
//	param, err := project.TestParameters.Get(ctx, "browser")
func (s *TestParameterService) Get(ctx context.Context, testParamID string) (*TestParameter, error) {
	if testParamID == "" {
		return nil, NewValidationError("testParamID", "test parameter ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/testparameterdefinitions/%s",
		s.client.baseURL,
		url.PathEscape(s.projectID),
		url.PathEscape(testParamID))

	// Make request with retry
	var param TestParameter
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &param)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get test parameter %s: %w", testParamID, err)
	}

	return &param, nil
}

// List returns all test parameter definitions for the project.
//
// Endpoint: GET /projects/{projectId}/testparameterdefinitions
//
// Example:
//
//	params, err := project.TestParameters.List(ctx)
//	for _, param := range params {
//	    fmt.Printf("Parameter: %s (%s)\n", param.Attributes.Name, param.ID)
//	}
func (s *TestParameterService) List(ctx context.Context, opts ...QueryOption) ([]*TestParameter, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/testparameterdefinitions",
		s.client.baseURL,
		url.PathEscape(s.projectID))

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
		Data []*TestParameter `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list test parameters: %w", err)
	}

	return response.Data, nil
}

// Create creates one or more test parameter definitions.
//
// Endpoint: POST /projects/{projectId}/testparameterdefinitions
//
// Example:
//
//	param := &polarion.TestParameter{
//	    Type: "testparameterdefinitions",
//	    Attributes: &polarion.TestParameterAttributes{
//	        Name: "Browser",
//	        Type: "enum",
//	        AllowedValues: []string{"Chrome", "Firefox", "Safari"},
//	    },
//	}
//	err := project.TestParameters.Create(ctx, param)
func (s *TestParameterService) Create(ctx context.Context, params ...*TestParameter) error {
	if len(params) == 0 {
		return nil
	}

	// Validate all parameters first
	for i, param := range params {
		if err := s.validateTestParameter(param); err != nil {
			return fmt.Errorf("validation failed for parameter %d: %w", i, err)
		}
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/testparameterdefinitions",
		s.client.baseURL,
		url.PathEscape(s.projectID))

	// Prepare request body
	body := map[string]interface{}{
		"data": params,
	}

	// Make request with retry
	var response struct {
		Data []*TestParameter `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return fmt.Errorf("failed to create test parameters: %w", err)
	}

	// Update parameters with created IDs
	for i, created := range response.Data {
		if i < len(params) {
			params[i].ID = created.ID
			if created.Links != nil {
				params[i].Links = created.Links
			}
		}
	}

	return nil
}

// Delete deletes a specific test parameter definition.
//
// Endpoint: DELETE /projects/{projectId}/testparameterdefinitions/{testParamId}
//
// Example:
//
//	err := project.TestParameters.Delete(ctx, "browser")
func (s *TestParameterService) Delete(ctx context.Context, testParamID string) error {
	if testParamID == "" {
		return NewValidationError("testParamID", "test parameter ID is required")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/testparameterdefinitions/%s",
		s.client.baseURL,
		url.PathEscape(s.projectID),
		url.PathEscape(testParamID))

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
		return fmt.Errorf("failed to delete test parameter %s: %w", testParamID, err)
	}

	return nil
}

// DeleteBatch deletes multiple test parameter definitions.
//
// Endpoint: DELETE /projects/{projectId}/testparameterdefinitions
//
// Example:
//
//	err := project.TestParameters.DeleteBatch(ctx, "browser", "os", "version")
func (s *TestParameterService) DeleteBatch(ctx context.Context, paramIDs ...string) error {
	if len(paramIDs) == 0 {
		return nil
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/testparameterdefinitions",
		s.client.baseURL,
		url.PathEscape(s.projectID))

	// Build query parameters with IDs
	params := url.Values{}
	for _, id := range paramIDs {
		params.Add("id", id)
	}
	urlStr += "?" + params.Encode()

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
		return fmt.Errorf("failed to delete test parameters: %w", err)
	}

	return nil
}

// validateTestParameter validates a test parameter before creation.
func (s *TestParameterService) validateTestParameter(param *TestParameter) error {
	if param == nil {
		return NewValidationError("param", "test parameter cannot be nil")
	}

	if param.Attributes == nil {
		return NewValidationError("attributes", "test parameter attributes cannot be nil")
	}

	if param.Attributes.Name == "" {
		return NewValidationError("name", "test parameter name is required")
	}

	// Set type if not set
	if param.Type == "" {
		param.Type = "testparameterdefinitions"
	}

	return nil
}
