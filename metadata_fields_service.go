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

// FieldsMetadataService handles fields metadata operations in the global context.
// This service provides access to field definitions for resource types including
// both standard and custom fields.
//
// Metadata support is limited to the following resource types:
// - workitems
// - documents
// - testruns
// - plans
//
// Requires: Polarion >= 2512
type FieldsMetadataService struct {
	client *Client
}

// Get retrieves fields metadata for a resource type and its target type in the global context.
// Returns information about all available fields including their types, labels, and constraints.
//
// Endpoint: GET /actions/getFieldsMetadata
// Requires: Polarion >= 2512
//
// Parameters:
//   - resourceType: The resource type (accepted values: "workitems", "documents", "testruns", "plans")
//   - targetType: The type of the object. Use "~" to represent no target type.
//
// Example:
//
//	// Get fields for work items of type "requirement"
//	metadata, err := client.FieldsMetadata.Get(ctx, "workitems", "requirement")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get fields for all work items (no specific type)
//	metadata, err := client.FieldsMetadata.Get(ctx, "workitems", "~")
func (s *FieldsMetadataService) Get(ctx context.Context, resourceType, targetType string) (*FieldsMetadata, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resourceType cannot be empty")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/actions/getFieldsMetadata", s.client.baseURL)

	// Add query parameters
	params := url.Values{}
	params.Set("resourceType", resourceType)
	if targetType != "" {
		params.Set("targetType", targetType)
	}
	urlStr += "?" + params.Encode()

	// Make request with retry
	var metadata FieldsMetadata
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &metadata)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get fields metadata for %s/%s: %w", resourceType, targetType, err)
	}

	return &metadata, nil
}

// ProjectFieldsMetadataService handles fields metadata operations in the project context.
// This service provides access to project-specific field definitions including
// custom fields configured at the project level.
//
// Requires: Polarion >= 2512
type ProjectFieldsMetadataService struct {
	client    *Client
	projectID string
}

// Get retrieves fields metadata for a resource type and its target type in the project context.
// Returns information about all available fields including project-specific custom fields.
//
// Endpoint: GET /projects/{projectId}/actions/getFieldsMetadata
// Requires: Polarion >= 2512
//
// Parameters:
//   - resourceType: The resource type (accepted values: "workitems", "documents", "testruns", "plans")
//   - targetType: The type of the object. Use "~" to represent no target type.
//
// Example:
//
//	// Get fields for work items of type "requirement" in a project
//	project := client.Project("myproject")
//	metadata, err := project.FieldsMetadata.Get(ctx, "workitems", "requirement")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *ProjectFieldsMetadataService) Get(ctx context.Context, resourceType, targetType string) (*FieldsMetadata, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resourceType cannot be empty")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/actions/getFieldsMetadata",
		s.client.baseURL, url.PathEscape(s.projectID))

	// Add query parameters
	params := url.Values{}
	params.Set("resourceType", resourceType)
	if targetType != "" {
		params.Set("targetType", targetType)
	}
	urlStr += "?" + params.Encode()

	// Make request with retry
	var metadata FieldsMetadata
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &metadata)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get fields metadata for %s/%s in project %s: %w",
			resourceType, targetType, s.projectID, err)
	}

	return &metadata, nil
}
