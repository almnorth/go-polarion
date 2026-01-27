# Error Handling Guide

This guide explains how to handle errors from the Polarion API client and extract detailed error information.

## Overview

The go-polarion client provides rich error information to help you diagnose and fix issues when working with the Polarion API. When an API request fails, the client returns an `APIError` that contains:

- HTTP status code
- Error message
- Detailed error information (when available)
- Raw response body (for debugging)

## Error Types

### APIError

The most common error type returned by the client. It represents an HTTP error response from the Polarion API.

```go
type APIError struct {
    StatusCode int           // HTTP status code (e.g., 400, 404, 500)
    Message    string        // Error message
    Response   *http.Response // Original HTTP response
    Details    []ErrorDetail // Detailed error information
    RawBody    string        // Raw response body for debugging
}
```

### ErrorDetail

Contains specific information about what went wrong, following the JSON:API error specification.

```go
type ErrorDetail struct {
    Status  string // HTTP status code as string
    Title   string // Short error title (optional)
    Detail  string // Detailed error message
    Pointer string // JSON pointer to the problematic field (e.g., "/data/0/attributes/customFields/myField")
}
```

The `Pointer` field is particularly useful as it tells you exactly which field caused the error.

### ValidationError

Represents client-side validation errors that occur before making an API request.

```go
type ValidationError struct {
    Field   string // Field name that failed validation
    Message string // Validation error message
}
```

### WorkItemError

Wraps an error and associates it with a specific work item.

```go
type WorkItemError struct {
    WorkItem *WorkItem // The work item that caused the error
    Err      error     // The underlying error
}
```

## Basic Error Handling

### Simple Error Check

```go
workItem, err := client.WorkItems.Get(ctx, "PROJECT", "ITEM-123")
if err != nil {
    log.Printf("Error: %v", err)
    return err
}
```

### Check for Specific Error Types

```go
workItem, err := client.WorkItems.Get(ctx, "PROJECT", "ITEM-123")
if err != nil {
    if polarion.IsNotFound(err) {
        log.Printf("Work item not found")
        return nil // Handle gracefully
    }
    return err
}
```

## Advanced Error Handling

### Extract Detailed Error Information

When you need more information about what went wrong:

```go
err := client.WorkItems.Create(ctx, "PROJECT", workItem)
if err != nil {
    // Get detailed error message including raw response
    detailedMsg := polarion.GetDetailedAPIError(err)
    log.Printf("Detailed error: %s", detailedMsg)
}
```

### Inspect Error Details Programmatically

To handle specific field errors:

```go
err := client.WorkItems.Create(ctx, "PROJECT", workItem)
if err != nil {
    details := polarion.GetAPIErrorDetails(err)
    if details != nil {
        for _, detail := range details {
            log.Printf("Field: %s", detail.Pointer)
            log.Printf("Error: %s", detail.Detail)
            
            // Handle specific field errors
            if strings.Contains(detail.Pointer, "customFields") {
                log.Printf("Custom field error detected")
            }
        }
    }
}
```

### Type Assertion for Full Control

For complete access to error information:

```go
err := client.WorkItems.Create(ctx, "PROJECT", workItem)
if err != nil {
    var apiErr *polarion.APIError
    if errors.As(err, &apiErr) {
        log.Printf("Status Code: %d", apiErr.StatusCode)
        log.Printf("Message: %s", apiErr.Message)
        
        // Access detailed errors
        for _, detail := range apiErr.Details {
            log.Printf("  - %s: %s", detail.Pointer, detail.Detail)
        }
        
        // Access raw response for debugging
        if apiErr.RawBody != "" {
            log.Printf("Raw response: %s", apiErr.RawBody)
        }
    }
}
```

## Common Error Scenarios

### Field Type Mismatch

When you send a boolean value for a field that expects a string:

```
Error: polarion api error (status 400) for POST https://polarion.example.com/rest/v1/projects/PROJECT/workitems: 400 Bad Request - field '/data/0/attributes/customFields/myField': Unexpected token, STRING expected, but was : BOOLEAN
```

The error clearly indicates:
- The field path: `/data/0/attributes/customFields/myField`
- The problem: Expected STRING but got BOOLEAN

### Missing Required Field

```
Error: polarion api error (status 400) for POST https://polarion.example.com/rest/v1/projects/PROJECT/workitems: 400 Bad Request - field '/data/0/attributes/title': Field is required
```

### Invalid Field Value

```
Error: polarion api error (status 400) for POST https://polarion.example.com/rest/v1/projects/PROJECT/workitems: 400 Bad Request - field '/data/0/attributes/status': Invalid status value 'invalid-status'
```

## Best Practices

1. **Always check for errors**: Never ignore error returns from API calls.

2. **Use helper functions**: Use `IsNotFound()`, `IsValidationError()`, etc. for common checks.

3. **Log detailed errors in development**: Use `GetDetailedAPIError()` during development to see full error information.

4. **Handle specific errors**: Use `GetAPIErrorDetails()` to handle specific field errors programmatically.

5. **Provide user-friendly messages**: Convert technical errors into user-friendly messages in production.

6. **Retry on retryable errors**: Use `IsRetryable()` to determine if an error should trigger a retry.

## Example: Robust Error Handling

```go
func createWorkItem(ctx context.Context, client *polarion.Client, projectID string, wi *polarion.WorkItem) error {
    err := client.WorkItems.Create(ctx, projectID, wi)
    if err != nil {
        // Check for specific error types
        if polarion.IsValidationError(err) {
            return fmt.Errorf("validation failed: %w", err)
        }
        
        // Get detailed API error information
        var apiErr *polarion.APIError
        if errors.As(err, &apiErr) {
            // Log detailed error for debugging
            log.Printf("API Error (status %d): %s", apiErr.StatusCode, apiErr.Message)
            
            // Handle specific field errors
            for _, detail := range apiErr.Details {
                log.Printf("  Field: %s", detail.Pointer)
                log.Printf("  Error: %s", detail.Detail)
                
                // Example: Handle custom field type mismatches
                if strings.Contains(detail.Detail, "STRING expected") && 
                   strings.Contains(detail.Pointer, "customFields") {
                    return fmt.Errorf("custom field type mismatch: %s", detail.Detail)
                }
            }
            
            // Return user-friendly error
            return fmt.Errorf("failed to create work item: %s", apiErr.Message)
        }
        
        return fmt.Errorf("unexpected error: %w", err)
    }
    
    return nil
}
```

## Debugging Tips

1. **Enable detailed logging**: Use `GetDetailedAPIError()` to see the full error including raw response.

2. **Check field pointers**: The `Pointer` field in `ErrorDetail` tells you exactly which field caused the error.

3. **Verify field types**: Ensure your custom field types match what Polarion expects (string, boolean, integer, etc.).

4. **Check field IDs**: Make sure custom field IDs exist in your Polarion project configuration.

5. **Review API documentation**: Consult the Polarion REST API documentation for field requirements and constraints.
