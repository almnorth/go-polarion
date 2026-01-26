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

// ProjectTemplateService handles project template operations.
// Project templates are used as blueprints when creating new projects.
type ProjectTemplateService struct {
	client *Client
}

// newProjectTemplateService creates a new project template service.
func newProjectTemplateService(client *Client) *ProjectTemplateService {
	return &ProjectTemplateService{
		client: client,
	}
}

// List returns all available project templates.
// Project templates can be used when creating new projects to apply
// predefined configurations and structures.
//
// Endpoint: GET /projecttemplates
//
// Example:
//
//	templates, err := client.ProjectTemplates.List(ctx)
//	for _, template := range templates {
//	    fmt.Printf("Template: %s (%s)\n", template.Attributes.Name, template.ID)
//	}
func (s *ProjectTemplateService) List(ctx context.Context, opts ...QueryOption) ([]*ProjectTemplate, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projecttemplates", s.client.baseURL)

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
		Data []*ProjectTemplate `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list project templates: %w", err)
	}

	return response.Data, nil
}
