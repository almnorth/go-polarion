// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	internalhttp "github.com/almnorth/go-polarion/internal/http"
	"net/url"
)

// WorkItemTypeService provides operations for work item type definitions.
type WorkItemTypeService struct {
	project *ProjectClient
}

// newWorkItemTypeService creates a new work item type service.
func newWorkItemTypeService(project *ProjectClient) *WorkItemTypeService {
	return &WorkItemTypeService{
		project: project,
	}
}

// Get retrieves a specific work item type definition by its ID.
//
// Example:
//
//	wiType, err := project.WorkItemTypes.Get(ctx, "requirement")
func (s *WorkItemTypeService) Get(ctx context.Context, typeID string, opts ...GetOption) (*WorkItemType, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/types/workitems/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(typeID))

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var wiType WorkItemType
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &wiType)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get work item type %s: %w", typeID, err)
	}

	return &wiType, nil
}

// List retrieves all work item type definitions for the project.
//
// Example:
//
//	types, err := project.WorkItemTypes.List(ctx)
func (s *WorkItemTypeService) List(ctx context.Context, opts ...QueryOption) ([]WorkItemType, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/types/workitems",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID))

	// Build query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var response struct {
		Data []WorkItemType `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list work item types: %w", err)
	}

	return response.Data, nil
}

// GetFields retrieves the field definitions for a specific work item type.
// This is a convenience method that retrieves the type and returns its fields.
//
// Example:
//
//	fields, err := project.WorkItemTypes.GetFields(ctx, "requirement")
func (s *WorkItemTypeService) GetFields(ctx context.Context, typeID string) ([]FieldDefinition, error) {
	wiType, err := s.Get(ctx, typeID)
	if err != nil {
		return nil, err
	}

	if wiType.Attributes == nil {
		return []FieldDefinition{}, nil
	}

	return wiType.Attributes.Fields, nil
}

// GetFieldByID retrieves a specific field definition from a work item type.
//
// Example:
//
//	field, err := project.WorkItemTypes.GetFieldByID(ctx, "requirement", "status")
func (s *WorkItemTypeService) GetFieldByID(ctx context.Context, typeID, fieldID string) (*FieldDefinition, error) {
	fields, err := s.GetFields(ctx, typeID)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		if field.ID == fieldID {
			return &field, nil
		}
	}

	return nil, fmt.Errorf("field %s not found in work item type %s", fieldID, typeID)
}

// ListFieldsByType returns a map of work item type IDs to their field definitions.
// This is useful for getting an overview of all fields across all types.
//
// Example:
//
//	fieldsByType, err := project.WorkItemTypes.ListFieldsByType(ctx)
func (s *WorkItemTypeService) ListFieldsByType(ctx context.Context) (map[string][]FieldDefinition, error) {
	types, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]FieldDefinition)
	for _, wiType := range types {
		if wiType.Attributes != nil {
			result[wiType.ID] = wiType.Attributes.Fields
		}
	}

	return result, nil
}

// ValidateFieldValue validates a field value against its definition.
// This is a helper method for client-side validation before sending data to the server.
func (s *WorkItemTypeService) ValidateFieldValue(field *FieldDefinition, value interface{}) error {
	if field.Required && value == nil {
		return NewValidationError(field.ID, fmt.Sprintf("field %s is required", field.ID))
	}

	if field.ReadOnly && value != nil {
		return NewValidationError(field.ID, fmt.Sprintf("field %s is read-only", field.ID))
	}

	// Additional type-specific validation could be added here
	// For now, we just check required and read-only

	return nil
}
