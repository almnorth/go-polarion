// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"errors"
	"fmt"
	"net/http"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// APIError represents an error response from the Polarion API.
// It contains the HTTP status code, error message, and optional detailed error information.
type APIError = internalhttp.APIError

// ErrorDetail contains detailed error information from the API response.
// This follows the JSON:API error object specification.
type ErrorDetail = internalhttp.ErrorDetail

// ValidationError represents a client-side validation error.
// This is used when input validation fails before making an API request.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// WorkItemError represents a work item specific error.
// It wraps an underlying error and associates it with a specific work item.
type WorkItemError struct {
	WorkItem *WorkItem
	Err      error
}

// Error implements the error interface for WorkItemError.
func (e *WorkItemError) Error() string {
	id := "unknown"
	title := "unknown"
	if e.WorkItem != nil {
		if e.WorkItem.ID != "" {
			id = e.WorkItem.ID
		}
		if e.WorkItem.Attributes != nil && e.WorkItem.Attributes.Title != "" {
			title = e.WorkItem.Attributes.Title
		}
	}
	return fmt.Sprintf("work item error (ID: %s, Title: %s): %v",
		id, title, e.Err)
}

// Unwrap returns the underlying error, allowing errors.Is and errors.As to work.
func (e *WorkItemError) Unwrap() error {
	return e.Err
}

// IsNotFound checks if an error is a 404 Not Found error.
// This is a convenience function for checking API errors.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

// IsValidationError checks if an error is a validation error.
// This is a convenience function for checking validation errors.
func IsValidationError(err error) bool {
	var valErr *ValidationError
	return errors.As(err, &valErr)
}

// IsRetryable checks if an error should trigger a retry.
// Returns true for server errors (5xx) and rate limit errors (429),
// false for client errors (4xx except 429) and other errors.
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Don't retry client errors (4xx) except 429 (rate limit)
		if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
			return apiErr.StatusCode == 429
		}
		// Retry server errors (5xx)
		return apiErr.StatusCode >= 500
	}
	// Retry network errors
	return true
}

// AsAPIError is a helper function that checks if an error is an APIError
// and assigns it to the target if it is. Returns true if the error is an APIError.
func AsAPIError(err error, target **APIError) bool {
	return errors.As(err, target)
}

// AsValidationError is a helper function that checks if an error is a ValidationError
// and assigns it to the target if it is. Returns true if the error is a ValidationError.
func AsValidationError(err error, target **ValidationError) bool {
	return errors.As(err, target)
}

// AsWorkItemError is a helper function that checks if an error is a WorkItemError
// and assigns it to the target if it is. Returns true if the error is a WorkItemError.
func AsWorkItemError(err error, target **WorkItemError) bool {
	return errors.As(err, target)
}

// NewAPIError creates a new APIError from an HTTP response.
func NewAPIError(statusCode int, message string, response *http.Response) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Response:   response,
	}
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewWorkItemError creates a new WorkItemError.
func NewWorkItemError(wi *WorkItem, err error) *WorkItemError {
	return &WorkItemError{
		WorkItem: wi,
		Err:      err,
	}
}

// GetDetailedAPIError returns a detailed error message for an APIError.
// If the error is not an APIError, it returns the standard error message.
// This is useful for debugging when you need more information about what went wrong.
//
// Example usage:
//
//	if err != nil {
//	    log.Printf("Error: %s", polarion.GetDetailedAPIError(err))
//	}
func GetDetailedAPIError(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.GetDetailedError()
	}
	return err.Error()
}

// GetAPIErrorDetails extracts the error details from an APIError.
// Returns nil if the error is not an APIError or has no details.
// This allows you to programmatically inspect specific error details.
//
// Example usage:
//
//	if details := polarion.GetAPIErrorDetails(err); details != nil {
//	    for _, detail := range details {
//	        log.Printf("Field: %s, Error: %s", detail.Pointer, detail.Detail)
//	    }
//	}
func GetAPIErrorDetails(err error) []ErrorDetail {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Details
	}
	return nil
}
