// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Victorien Elvinger
// Copyright (c) 2025 Siemens AG

package polarion

import (
	"context"
	"fmt"
	"net/url"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// CustomFieldService handles project-scoped custom field operations.
// This service provides access to custom field configurations in the project context.
//
// Requires: Polarion >= 2512
type CustomFieldService struct {
	client    *Client
	projectID string
}

// Get retrieves custom fields configuration for a resource type and target type in the project context.
//
// Endpoint: GET /projects/{projectId}/customfields/{resourceType}/{targetType}
// Requires: Polarion >= 2512
//
// Parameters:
//   - resourceType: The resource type (e.g., "workitems", "documents", "testruns", "plans")
//   - targetType: The type of the object. Use "~" to represent no target type.
//
// Example:
//
//	project := client.Project("myproject")
//	config, err := project.CustomFields.Get(ctx, "workitems", "requirement")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, field := range config.Attributes.Fields {
//	    fmt.Printf("Field: %s (%s)\n", field.Name, field.Type.Kind)
//	}
func (s *CustomFieldService) Get(ctx context.Context, resourceType, targetType string, opts ...GetOption) (*CustomFieldsConfig, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resourceType cannot be empty")
	}
	if targetType == "" {
		return nil, fmt.Errorf("targetType cannot be empty")
	}

	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/customfields/%s/%s",
		s.client.baseURL, url.PathEscape(s.projectID),
		url.PathEscape(resourceType), url.PathEscape(targetType))

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var config CustomFieldsConfig
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &config)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get custom fields for %s/%s in project %s: %w",
			resourceType, targetType, s.projectID, err)
	}

	return &config, nil
}

// GetByID retrieves custom fields configuration using a CustomFieldID helper.
//
// Example:
//
//	project := client.Project("myproject")
//	id := polarion.CustomFieldID{
//	    ResourceType: "workitems",
//	    TargetType:   "requirement",
//	}
//	config, err := project.CustomFields.GetByID(ctx, id)
func (s *CustomFieldService) GetByID(ctx context.Context, id CustomFieldID, opts ...GetOption) (*CustomFieldsConfig, error) {
	return s.Get(ctx, id.ResourceType, id.TargetType, opts...)
}

// Create creates custom fields configurations in the project context.
//
// Endpoint: POST /projects/{projectId}/customfields
// Requires: Polarion >= 2512
//
// Example:
//
//	project := client.Project("myproject")
//	config := polarion.NewCustomFieldsConfig("workitems", "requirement")
//	config.Attributes.Fields = []polarion.CustomFieldDefinition{
//	    {
//	        ID:   "customField1",
//	        Name: "Custom Field 1",
//	        Type: polarion.CustomFieldType{Kind: "string"},
//	    },
//	}
//	created, err := project.CustomFields.Create(ctx, config)
func (s *CustomFieldService) Create(ctx context.Context, configs ...*CustomFieldsConfig) ([]*CustomFieldsConfig, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("at least one custom fields configuration must be provided")
	}

	// Prepare request body
	body := map[string]interface{}{
		"data": configs,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/customfields",
		s.client.baseURL, url.PathEscape(s.projectID))

	// Make request with retry
	var response struct {
		Data []CustomFieldsConfig `json:"data"`
	}

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create custom fields in project %s: %w", s.projectID, err)
	}

	// Convert to pointers
	var createdConfigs []*CustomFieldsConfig
	for i := range response.Data {
		createdConfigs = append(createdConfigs, &response.Data[i])
	}

	return createdConfigs, nil
}

// Update updates custom fields configuration for a resource type and target type in the project context.
//
// Endpoint: PATCH /projects/{projectId}/customfields/{resourceType}/{targetType}
// Requires: Polarion >= 2512
//
// Example:
//
//	project := client.Project("myproject")
//	config.Attributes.Fields = append(config.Attributes.Fields, polarion.CustomFieldDefinition{
//	    ID:   "newField",
//	    Name: "New Field",
//	    Type: polarion.CustomFieldType{Kind: "integer"},
//	})
//	err := project.CustomFields.Update(ctx, "workitems", "requirement", config)
func (s *CustomFieldService) Update(ctx context.Context, resourceType, targetType string, config *CustomFieldsConfig) error {
	if resourceType == "" {
		return fmt.Errorf("resourceType cannot be empty")
	}
	if targetType == "" {
		return fmt.Errorf("targetType cannot be empty")
	}
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Prepare request body
	body := map[string]interface{}{
		"data": config,
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/customfields/%s/%s",
		s.client.baseURL, url.PathEscape(s.projectID),
		url.PathEscape(resourceType), url.PathEscape(targetType))

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
		return fmt.Errorf("failed to update custom fields for %s/%s in project %s: %w",
			resourceType, targetType, s.projectID, err)
	}

	return nil
}

// UpdateByID updates custom fields configuration using a CustomFieldID helper.
//
// Example:
//
//	project := client.Project("myproject")
//	id := polarion.CustomFieldID{
//	    ResourceType: "workitems",
//	    TargetType:   "requirement",
//	}
//	err := project.CustomFields.UpdateByID(ctx, id, config)
func (s *CustomFieldService) UpdateByID(ctx context.Context, id CustomFieldID, config *CustomFieldsConfig) error {
	return s.Update(ctx, id.ResourceType, id.TargetType, config)
}

// Delete deletes custom fields configuration for a resource type and target type in the project context.
//
// Endpoint: DELETE /projects/{projectId}/customfields/{resourceType}/{targetType}
// Requires: Polarion >= 2512
//
// Example:
//
//	project := client.Project("myproject")
//	err := project.CustomFields.Delete(ctx, "workitems", "requirement")
func (s *CustomFieldService) Delete(ctx context.Context, resourceType, targetType string) error {
	if resourceType == "" {
		return fmt.Errorf("resourceType cannot be empty")
	}
	if targetType == "" {
		return fmt.Errorf("targetType cannot be empty")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/customfields/%s/%s",
		s.client.baseURL, url.PathEscape(s.projectID),
		url.PathEscape(resourceType), url.PathEscape(targetType))

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
		return fmt.Errorf("failed to delete custom fields for %s/%s in project %s: %w",
			resourceType, targetType, s.projectID, err)
	}

	return nil
}

// DeleteByID deletes custom fields configuration using a CustomFieldID helper.
//
// Example:
//
//	project := client.Project("myproject")
//	id := polarion.CustomFieldID{
//	    ResourceType: "workitems",
//	    TargetType:   "requirement",
//	}
//	err := project.CustomFields.DeleteByID(ctx, id)
func (s *CustomFieldService) DeleteByID(ctx context.Context, id CustomFieldID) error {
	return s.Delete(ctx, id.ResourceType, id.TargetType)
}
