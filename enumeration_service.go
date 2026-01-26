// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	"net/url"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// EnumerationService provides operations for enumerations.
type EnumerationService struct {
	project *ProjectClient
}

// newEnumerationService creates a new enumeration service.
func newEnumerationService(project *ProjectClient) *EnumerationService {
	return &EnumerationService{
		project: project,
	}
}

// Get retrieves a specific enumeration by its ID components.
// The enumeration ID consists of context, name, and target type.
//
// Example:
//
//	enum, err := project.Enumerations.Get(ctx, "workitem", "status", "requirement")
func (s *EnumerationService) Get(ctx context.Context, context, name, targetType string, opts ...GetOption) (*Enumeration, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	enumPath := fmt.Sprintf("%s/%s/%s", url.PathEscape(context), url.PathEscape(name), url.PathEscape(targetType))
	urlStr := fmt.Sprintf("%s/projects/%s/enumerations/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		enumPath)

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var enum Enumeration
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &enum)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get enumeration %s/%s/%s: %w", context, name, targetType, err)
	}

	return &enum, nil
}

// GetByID retrieves a specific enumeration using an EnumerationID.
//
// Example:
//
//	enumID := polarion.NewEnumerationID("workitem", "status", "requirement")
//	enum, err := project.Enumerations.GetByID(ctx, enumID)
func (s *EnumerationService) GetByID(ctx context.Context, enumID *EnumerationID, opts ...GetOption) (*Enumeration, error) {
	return s.Get(ctx, enumID.Context, enumID.Name, enumID.TargetType, opts...)
}

// List retrieves all enumerations for the project.
// Note: This may return a large number of enumerations depending on the project configuration.
//
// Example:
//
//	enums, err := project.Enumerations.List(ctx)
func (s *EnumerationService) List(ctx context.Context, opts ...QueryOption) ([]Enumeration, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/enumerations",
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
		Data []Enumeration `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list enumerations: %w", err)
	}

	return response.Data, nil
}

// Create creates a new enumeration.
// The enumeration must have valid context, name, and target type in its ID.
//
// Example:
//
//	enum := &polarion.Enumeration{
//	    Type: "enumerations",
//	    ID:   "project/myproject/enum/workitem/customStatus/requirement",
//	    Attributes: &polarion.EnumerationAttributes{
//	        Options: []polarion.EnumerationOption{
//	            {ID: "open", Name: "Open", Default: true},
//	            {ID: "closed", Name: "Closed"},
//	        },
//	    },
//	}
//	err := project.Enumerations.Create(ctx, enum)
func (s *EnumerationService) Create(ctx context.Context, enum *Enumeration) error {
	if err := s.validateEnumeration(enum); err != nil {
		return err
	}

	// Build URL - use the enumeration ID path
	urlStr := fmt.Sprintf("%s/projects/%s/enumerations",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID))

	// Prepare request body
	body := map[string]interface{}{
		"data": enum,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		// Update the enum with the response
		return internalhttp.DecodeDataResponse(resp, enum)
	})

	if err != nil {
		return fmt.Errorf("failed to create enumeration: %w", err)
	}

	return nil
}

// Update updates an existing enumeration.
// The enumeration must have a valid ID set.
//
// Example:
//
//	enum.Attributes.Options = append(enum.Attributes.Options,
//	    polarion.EnumerationOption{ID: "inprogress", Name: "In Progress"})
//	err := project.Enumerations.Update(ctx, enum)
func (s *EnumerationService) Update(ctx context.Context, enum *Enumeration) error {
	if enum.ID == "" {
		return NewValidationError("ID", "enumeration ID is required for update")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/%s", s.project.client.baseURL, enum.ID)

	// Prepare request body
	body := map[string]interface{}{
		"data": enum,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		// Update the enum with the response
		return internalhttp.DecodeDataResponse(resp, enum)
	})

	if err != nil {
		return fmt.Errorf("failed to update enumeration %s: %w", enum.ID, err)
	}

	return nil
}

// Delete deletes an enumeration by its ID components.
//
// Example:
//
//	err := project.Enumerations.Delete(ctx, "workitem", "customStatus", "requirement")
func (s *EnumerationService) Delete(ctx context.Context, context, name, targetType string) error {
	// Build URL
	enumPath := fmt.Sprintf("%s/%s/%s", url.PathEscape(context), url.PathEscape(name), url.PathEscape(targetType))
	urlStr := fmt.Sprintf("%s/projects/%s/enumerations/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		enumPath)

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, nil)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete enumeration %s/%s/%s: %w", context, name, targetType, err)
	}

	return nil
}

// DeleteByID deletes an enumeration using an EnumerationID.
//
// Example:
//
//	enumID := polarion.NewEnumerationID("workitem", "customStatus", "requirement")
//	err := project.Enumerations.DeleteByID(ctx, enumID)
func (s *EnumerationService) DeleteByID(ctx context.Context, enumID *EnumerationID) error {
	return s.Delete(ctx, enumID.Context, enumID.Name, enumID.TargetType)
}

// validateEnumeration validates an enumeration before creation or update.
func (s *EnumerationService) validateEnumeration(enum *Enumeration) error {
	if enum == nil {
		return NewValidationError("enumeration", "enumeration cannot be nil")
	}

	if enum.Attributes == nil {
		return NewValidationError("attributes", "enumeration attributes cannot be nil")
	}

	if len(enum.Attributes.Options) == 0 {
		return NewValidationError("options", "enumeration must have at least one option")
	}

	// Set type if not set
	if enum.Type == "" {
		enum.Type = "enumerations"
	}

	return nil
}
