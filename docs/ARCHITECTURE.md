# Architecture

Design principles, architecture, and implementation details of go-polarion.

## Table of Contents

- [Design Principles](#design-principles)
- [Architecture Overview](#architecture-overview)
- [Package Structure](#package-structure)
- [Client Layer](#client-layer)
- [Service Layer](#service-layer)
- [Internal Layer](#internal-layer)
- [Data Models](#data-models)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)

## Design Principles

### 1. Simple but Smart

The library provides a clean, intuitive API while handling complexity internally.

**Simple API:**
```go
// Easy to use
wi, err := project.WorkItems.Get(ctx, "WI-123")
```

**Smart Internals:**
- Automatic batching for bulk operations
- Exponential backoff with jitter for retries
- Efficient pagination handling
- Request/response validation

### 2. Idiomatic Go

Follows Go best practices and conventions:

- Clear, descriptive names
- Minimal interfaces
- Composition over inheritance
- Error values, not exceptions
- Context for cancellation

**Example:**
```go
// Idiomatic error handling
wi, err := project.WorkItems.Get(ctx, "WI-123")
if err != nil {
    if polarion.IsNotFound(err) {
        // Handle not found
    }
    return err
}
```

### 3. Context Everywhere

Full support for cancellation and timeouts:

```go
// Context-based cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

items, err := project.WorkItems.QueryAll(ctx, "type:requirement")
```

### 4. Interface-Based

Easy to mock and test:

```go
// Services are interfaces
type WorkItemService interface {
    Get(ctx context.Context, id string) (*WorkItem, error)
    Create(ctx context.Context, items ...*WorkItem) error
    // ...
}

// Easy to mock for testing
type mockWorkItemService struct {
    getFunc func(ctx context.Context, id string) (*WorkItem, error)
}
```

### 5. Zero Dependencies

Uses only Go standard library:

- No external dependencies
- Smaller binary size
- Easier maintenance
- Better security

### 6. Encapsulation

Internal details hidden from public API:

- `internal/` package for implementation details
- Clean separation of concerns
- Stable public API
- Freedom to refactor internals

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Client    │  │   Project    │  │   Config     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       Service Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  WorkItems   │  │    Users     │  │   Projects   │      │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤      │
│  │  Comments    │  │  UserGroups  │  │ Enumerations │      │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤      │
│  │ Attachments  │  │   Metadata   │  │CustomFields  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Internal Layer                          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                   HTTP Client                         │   │
│  │  • Authentication  • Retry Logic  • Rate Limiting    │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │  Polarion API   │
                   └─────────────────┘
```

## Package Structure

```
go-polarion/
├── client.go              # Main client
├── config.go              # Configuration
├── project.go             # Project-scoped client
├── errors.go              # Error types
├── query.go               # Query helpers
│
├── workitem*.go           # Work item models and services
├── user*.go               # User models and services
├── project*.go            # Project models and services
├── enumeration*.go        # Enumeration models and services
├── metadata*.go           # Metadata models and services
├── customfield*.go        # Custom field models and services
│
├── internal/              # Internal implementation
│   └── http/              # HTTP client and retry logic
│       ├── client.go      # HTTP client
│       └── retry.go       # Retry logic
│
├── codegen/               # Code generation
│   ├── generator.go       # Generator
│   ├── parser.go          # Parser
│   └── template.go        # Templates
│
├── cmd/                   # CLI tools
│   └── polarion-codegen/  # Code generation tool
│
├── examples/              # Examples
│   ├── basic/             # Basic usage
│   ├── custom_workitems/  # Custom work items
│   └── codegen/           # Code generation
│
└── docs/                  # Documentation
    ├── API-REFERENCE.md
    ├── CONFIGURATION.md
    ├── ARCHITECTURE.md
    ├── CODEGEN.md
    └── CUSTOM-WORKITEMS.md
```

## Client Layer

### Main Client

The main client provides access to global services and project-scoped operations.

```go
type Client struct {
    config *Config
    http   HTTPClient
    
    // Global services
    Projects          ProjectService
    ProjectTemplates  ProjectTemplateService
    Users             UserService
    UserGroups        UserGroupService
    GlobalEnumerations EnumerationGlobalService
    GlobalCustomFields CustomFieldGlobalService
    Metadata          MetadataService
    FieldsMetadata    MetadataFieldsService
}
```

**Responsibilities:**
- Client initialization
- Configuration management
- Access to global services
- Project client creation

### Project Client

Project-scoped client for project-specific operations.

```go
type Project struct {
    client    *Client
    projectID string
    
    // Project services
    WorkItems         WorkItemService
    WorkItemComments  WorkItemCommentService
    WorkItemAttachments WorkItemAttachmentService
    WorkItemApprovals WorkItemApprovalService
    WorkItemWorkRecords WorkItemWorkRecordService
    WorkItemLinks     WorkItemLinkService
    WorkItemTypes     WorkItemTypeService
    TestParameters    TestParameterService
    Enumerations      EnumerationService
    CustomFields      CustomFieldService
    FieldsMetadata    MetadataFieldsService
}
```

**Responsibilities:**
- Project-scoped operations
- Access to project services
- Project ID management

## Service Layer

Services implement resource-specific operations following a consistent pattern.

### Service Interface Pattern

```go
type WorkItemService interface {
    // CRUD operations
    Get(ctx context.Context, id string, opts ...QueryOption) (*WorkItem, error)
    Create(ctx context.Context, items ...*WorkItem) error
    Update(ctx context.Context, items ...*WorkItem) error
    Delete(ctx context.Context, ids ...string) error
    
    // Query operations
    Query(ctx context.Context, opts QueryOptions) (*QueryResult, error)
    QueryAll(ctx context.Context, query string, opts ...QueryOption) ([]*WorkItem, error)
    
    // Relationship operations
    GetRelationships(ctx context.Context, id, relationship string) (interface{}, error)
    CreateRelationships(ctx context.Context, id, relationship string, data ...interface{}) error
    UpdateRelationships(ctx context.Context, id, relationship string, data ...interface{}) error
    DeleteRelationships(ctx context.Context, id, relationship string) error
}
```

### Service Implementation

```go
type workItemService struct {
    client    *Client
    projectID string
}

func (s *workItemService) Get(ctx context.Context, id string, opts ...QueryOption) (*WorkItem, error) {
    // Build URL
    url := fmt.Sprintf("/projects/%s/workitems/%s", s.projectID, id)
    
    // Apply options
    params := applyQueryOptions(opts...)
    
    // Make request
    var response struct {
        Data *WorkItem `json:"data"`
    }
    
    if err := s.client.http.Get(ctx, url, params, &response); err != nil {
        return nil, err
    }
    
    return response.Data, nil
}
```

### Service Responsibilities

- Request construction
- Parameter validation
- Response parsing
- Error handling
- Batching logic
- Pagination handling

## Internal Layer

### HTTP Client

The internal HTTP client handles low-level HTTP operations.

**Location:** `internal/http/client.go`

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    retry      *RetryConfig
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Add authentication
    req.Header.Set("Authorization", "Bearer "+c.token)
    
    // Add headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    
    // Execute with retry
    return c.doWithRetry(ctx, req)
}
```

**Responsibilities:**
- HTTP request execution
- Authentication
- Header management
- Response handling
- Connection pooling

### Retry Logic

Implements exponential backoff with jitter.

**Location:** `internal/http/retry.go`

```go
type RetryConfig struct {
    MaxRetries int
    MinWait    time.Duration
    MaxWait    time.Duration
}

func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
        // Clone request for retry
        reqClone := req.Clone(ctx)
        
        // Execute request
        resp, err := c.httpClient.Do(reqClone)
        
        // Check if should retry
        if !shouldRetry(resp, err) {
            return resp, err
        }
        
        lastErr = err
        
        // Calculate backoff
        wait := calculateBackoff(attempt, c.retry)
        
        // Wait with context
        select {
        case <-time.After(wait):
            continue
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    return nil, lastErr
}
```

**Features:**
- Exponential backoff
- Jitter to prevent thundering herd
- Context-aware waiting
- Configurable retry limits

## Data Models

### JSON:API Format

All models follow the JSON:API specification:

```go
type WorkItem struct {
    Type          string                 `json:"type"`
    ID            string                 `json:"id,omitempty"`
    Attributes    *WorkItemAttributes    `json:"attributes,omitempty"`
    Relationships *WorkItemRelationships `json:"relationships,omitempty"`
    Links         *WorkItemLinks         `json:"links,omitempty"`
}

type WorkItemAttributes struct {
    Title       string       `json:"title,omitempty"`
    Status      string       `json:"status,omitempty"`
    Description *TextContent `json:"description,omitempty"`
    CustomFields map[string]interface{} `json:"customFields,omitempty"`
    // ...
}
```

### Type Safety

Strong typing for all models:

```go
// Type-safe enums
type ApprovalStatus string

const (
    ApprovalStatusApproved  ApprovalStatus = "approved"
    ApprovalStatusRejected  ApprovalStatus = "rejected"
    ApprovalStatusPending   ApprovalStatus = "pending"
)

// Type-safe field types
type DateOnly struct {
    time.Time
}

type TimeOnly struct {
    Hour   int
    Minute int
    Second int
}
```

## Error Handling

### Error Types

```go
// APIError represents an error from the Polarion API
type APIError struct {
    StatusCode int
    Message    string
    Details    []ErrorDetail
}

// ErrorDetail provides additional error information
type ErrorDetail struct {
    Source string
    Detail string
}
```

### Error Helpers

```go
// IsNotFound checks if error is a 404 Not Found
func IsNotFound(err error) bool {
    var apiErr *APIError
    return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

// AsAPIError extracts APIError from error
func AsAPIError(err error, target **APIError) bool {
    return errors.As(err, target)
}
```

### Error Handling Pattern

```go
wi, err := project.WorkItems.Get(ctx, "WI-123")
if err != nil {
    // Check for specific error types
    if polarion.IsNotFound(err) {
        return nil, fmt.Errorf("work item not found")
    }
    
    // Extract API error details
    var apiErr *polarion.APIError
    if polarion.AsAPIError(err, &apiErr) {
        return nil, fmt.Errorf("API error: %s", apiErr.Message)
    }
    
    // Handle other errors
    return nil, err
}
```

## Testing Strategy

### Unit Tests

Test individual functions and methods:

```go
func TestWorkItemService_Get(t *testing.T) {
    // Create mock HTTP client
    mock := &mockHTTPClient{
        doFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
            // Return mock response
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(`{"data":{"id":"WI-123"}}`)),
            }, nil
        },
    }
    
    // Create service with mock
    service := &workItemService{
        client: &Client{http: mock},
        projectID: "test",
    }
    
    // Test
    wi, err := service.Get(context.Background(), "WI-123")
    assert.NoError(t, err)
    assert.Equal(t, "WI-123", wi.ID)
}
```

### Integration Tests

Test against real Polarion instance:

```go
func TestIntegration_WorkItems(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    client, err := polarion.New(
        os.Getenv("POLARION_URL"),
        os.Getenv("POLARION_TOKEN"),
    )
    require.NoError(t, err)
    
    project := client.Project(os.Getenv("POLARION_PROJECT"))
    
    // Test operations
    wi, err := project.WorkItems.Get(context.Background(), "WI-123")
    require.NoError(t, err)
    assert.NotEmpty(t, wi.ID)
}
```

### Mock Interfaces

Easy to mock for testing:

```go
type mockWorkItemService struct {
    getFunc    func(ctx context.Context, id string) (*WorkItem, error)
    createFunc func(ctx context.Context, items ...*WorkItem) error
}

func (m *mockWorkItemService) Get(ctx context.Context, id string, opts ...QueryOption) (*WorkItem, error) {
    if m.getFunc != nil {
        return m.getFunc(ctx, id)
    }
    return nil, nil
}
```

## Performance Considerations

### 1. Connection Pooling

Reuse HTTP connections:

```go
// Default transport with connection pooling
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
```

### 2. Batching

Automatic batching for bulk operations:

```go
// Automatically batched into multiple requests
items := make([]*WorkItem, 500)
err := project.WorkItems.Create(ctx, items...)
```

### 3. Pagination

Efficient pagination handling:

```go
// Automatic pagination
items, err := project.WorkItems.QueryAll(ctx, "type:requirement")

// Manual pagination for memory control
result, err := project.WorkItems.Query(ctx, QueryOptions{
    PageSize: 50,
})
```

### 4. Field Selection

Request only needed fields:

```go
// Reduce response size
fields := polarion.NewFieldSelector().
    WithWorkItemFields("id,title,status")

items, err := project.WorkItems.QueryAll(ctx, "type:requirement",
    polarion.WithFields(fields))
```

## Security Considerations

### 1. Token Management

- Never log bearer tokens
- Store tokens securely
- Use environment variables

### 2. TLS Configuration

- Enforce TLS 1.2+
- Validate certificates
- Support custom CA certificates

### 3. Input Validation

- Validate all inputs
- Sanitize user data
- Prevent injection attacks

### 4. Error Messages

- Don't expose sensitive data in errors
- Log detailed errors server-side
- Return generic errors to clients

## See Also

- [API Reference](API-REFERENCE.md) - Complete API documentation
- [Configuration](CONFIGURATION.md) - Client configuration options
- [Custom Work Items](CUSTOM-WORKITEMS.md) - Type-safe custom work items
- [Code Generation](CODEGEN.md) - Automatic code generation tool
