// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	"net/url"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// GlobalEnumerationService handles global enumeration operations.
// Global enumerations are not scoped to a specific project and are available
// across the entire Polarion instance.
type GlobalEnumerationService struct {
	client *Client
}

// newGlobalEnumerationService creates a new global enumeration service.
func newGlobalEnumerationService(client *Client) *GlobalEnumerationService {
	return &GlobalEnumerationService{
		client: client,
	}
}

// Get retrieves a specific global enumeration by its ID components.
// The enumeration ID consists of context, name, and target type.
//
// Example:
//
//	enum, err := client.GlobalEnumerations.Get(ctx, "workitem", "status", "requirement")
func (s *GlobalEnumerationService) Get(ctx context.Context, enumContext, enumName, targetType string, opts ...GetOption) (*Enumeration, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	enumPath := fmt.Sprintf("%s/%s/%s", url.PathEscape(enumContext), url.PathEscape(enumName), url.PathEscape(targetType))
	urlStr := fmt.Sprintf("%s/enumerations/%s",
		s.client.baseURL,
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
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &enum)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get global enumeration %s/%s/%s: %w", enumContext, enumName, targetType, err)
	}

	return &enum, nil
}

// GetByID retrieves a specific global enumeration using an EnumerationID.
//
// Example:
//
//	enumID := polarion.NewEnumerationID("workitem", "status", "requirement")
//	enum, err := client.GlobalEnumerations.GetByID(ctx, enumID)
func (s *GlobalEnumerationService) GetByID(ctx context.Context, enumID *EnumerationID, opts ...GetOption) (*Enumeration, error) {
	return s.Get(ctx, enumID.Context, enumID.Name, enumID.TargetType, opts...)
}

// List retrieves all global enumerations.
// Note: This may return a large number of enumerations depending on the Polarion configuration.
//
// Example:
//
//	enums, err := client.GlobalEnumerations.List(ctx)
func (s *GlobalEnumerationService) List(ctx context.Context, opts ...QueryOption) ([]Enumeration, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/enumerations", s.client.baseURL)

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

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list global enumerations: %w", err)
	}

	return response.Data, nil
}

// Create creates a new global enumeration.
// The enumeration must have valid context, name, and target type in its ID.
//
// Example:
//
//	enum := &polarion.Enumeration{
//	    Type: "enumerations",
//	    ID:   "enum/workitem/customStatus/requirement",
//	    Attributes: &polarion.EnumerationAttributes{
//	        Options: []polarion.EnumerationOption{
//	            {ID: "open", Name: "Open", Default: true},
//	            {ID: "closed", Name: "Closed"},
//	        },
//	    },
//	}
//	err := client.GlobalEnumerations.Create(ctx, enum)
func (s *GlobalEnumerationService) Create(ctx context.Context, enum *Enumeration) error {
	if err := s.validateEnumeration(enum); err != nil {
		return err
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/enumerations", s.client.baseURL)

	// Prepare request body
	body := map[string]interface{}{
		"data": enum,
	}

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		// Update the enum with the response
		return internalhttp.DecodeDataResponse(resp, enum)
	})

	if err != nil {
		return fmt.Errorf("failed to create global enumeration: %w", err)
	}

	return nil
}

// Update updates an existing global enumeration.
// The enumeration must have a valid ID set.
//
// Example:
//
//	enum.Attributes.Options = append(enum.Attributes.Options,
//	    polarion.EnumerationOption{ID: "inprogress", Name: "In Progress"})
//	err := client.GlobalEnumerations.Update(ctx, enum)
func (s *GlobalEnumerationService) Update(ctx context.Context, enum *Enumeration) error {
	if enum.ID == "" {
		return NewValidationError("ID", "enumeration ID is required for update")
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/%s", s.client.baseURL, enum.ID)

	// Prepare request body
	body := map[string]interface{}{
		"data": enum,
	}

	// Make request with retry
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		// Update the enum with the response
		return internalhttp.DecodeDataResponse(resp, enum)
	})

	if err != nil {
		return fmt.Errorf("failed to update global enumeration %s: %w", enum.ID, err)
	}

	return nil
}

// Delete deletes a global enumeration by its ID components.
//
// Example:
//
//	err := client.GlobalEnumerations.Delete(ctx, "workitem", "customStatus", "requirement")
func (s *GlobalEnumerationService) Delete(ctx context.Context, enumContext, enumName, targetType string) error {
	// Build URL
	enumPath := fmt.Sprintf("%s/%s/%s", url.PathEscape(enumContext), url.PathEscape(enumName), url.PathEscape(targetType))
	urlStr := fmt.Sprintf("%s/enumerations/%s",
		s.client.baseURL,
		enumPath)

	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "DELETE", urlStr, nil)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete global enumeration %s/%s/%s: %w", enumContext, enumName, targetType, err)
	}

	return nil
}

// DeleteByID deletes a global enumeration using an EnumerationID.
//
// Example:
//
//	enumID := polarion.NewEnumerationID("workitem", "customStatus", "requirement")
//	err := client.GlobalEnumerations.DeleteByID(ctx, enumID)
func (s *GlobalEnumerationService) DeleteByID(ctx context.Context, enumID *EnumerationID) error {
	return s.Delete(ctx, enumID.Context, enumID.Name, enumID.TargetType)
}

// validateEnumeration validates an enumeration before creation or update.
func (s *GlobalEnumerationService) validateEnumeration(enum *Enumeration) error {
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
