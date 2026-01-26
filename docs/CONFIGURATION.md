# Configuration

Complete guide to configuring the go-polarion client.

## Table of Contents

- [Client Creation](#client-creation)
- [Configuration Options](#configuration-options)
- [Batch Operations](#batch-operations)
- [Pagination](#pagination)
- [Retry Logic](#retry-logic)
- [Timeouts](#timeouts)
- [Custom HTTP Client](#custom-http-client)
- [Field Selection](#field-selection)
- [Best Practices](#best-practices)

## Client Creation

### Basic Client

```go
import (
    "github.com/almnorth/go-polarion"
)

client, err := polarion.New(
    "https://polarion.example.com/rest/v1",
    "your-bearer-token",
)
if err != nil {
    log.Fatal(err)
}
```

### Client with Options

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithBatchSize(50),
    polarion.WithPageSize(100),
    polarion.WithTimeout(60*time.Second),
)
```

## Configuration Options

### WithBatchSize

Sets the batch size for bulk operations (create, update, delete).

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithBatchSize(50), // Process 50 items per batch
)
```

**Default:** 100

**Use Cases:**
- Reduce batch size if hitting API rate limits
- Increase batch size for better performance with large datasets
- Adjust based on network conditions

**Example:**

```go
// Create 500 work items with batch size of 50
// This will make 10 API calls (500 / 50)
items := make([]*polarion.WorkItem, 500)
for i := range items {
    items[i] = &polarion.WorkItem{
        Type: "workitems",
        Attributes: &polarion.WorkItemAttributes{
            Title: fmt.Sprintf("Item %d", i+1),
        },
    }
}

err := project.WorkItems.Create(ctx, items...)
```

### WithPageSize

Sets the default page size for query operations.

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithPageSize(100), // Fetch 100 items per page
)
```

**Default:** 100

**Use Cases:**
- Reduce page size for faster initial response
- Increase page size to reduce number of API calls
- Balance between memory usage and network overhead

**Example:**

```go
// Query with default page size
items, err := project.WorkItems.QueryAll(ctx, "type:requirement")

// Override page size for specific query
result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
    Query:    "type:requirement",
    PageSize: 50, // Override default
})
```

### WithMaxContentSize

Sets the maximum request body size in bytes.

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithMaxContentSize(2 * 1024 * 1024), // 2MB limit
)
```

**Default:** 10MB (10 * 1024 * 1024)

**Use Cases:**
- Prevent memory issues with large payloads
- Enforce size limits for attachments
- Control resource usage

### WithRetryConfig

Configures retry behavior for failed requests.

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 3,                  // Maximum number of retries
        MinWait:    5 * time.Second,    // Minimum wait between retries
        MaxWait:    15 * time.Second,   // Maximum wait between retries
    }),
)
```

**Defaults:**
- MaxRetries: 3
- MinWait: 1 second
- MaxWait: 30 seconds

**Retry Strategy:**
- Exponential backoff with jitter
- Only retries on transient errors (5xx, network errors)
- Does not retry on client errors (4xx)

**Example:**

```go
// Aggressive retry for unstable networks
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 5,
        MinWait:    2 * time.Second,
        MaxWait:    60 * time.Second,
    }),
)

// Disable retries
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 0,
    }),
)
```

### WithTimeout

Sets the HTTP client timeout for all requests.

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithTimeout(60 * time.Second), // 60 second timeout
)
```

**Default:** 30 seconds

**Use Cases:**
- Increase timeout for slow networks
- Decrease timeout for faster failure detection
- Set based on expected operation duration

**Example:**

```go
// Long timeout for bulk operations
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithTimeout(5 * time.Minute),
)

// Short timeout for quick operations
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithTimeout(10 * time.Second),
)
```

### WithHTTPClient

Uses a custom HTTP client instead of the default.

```go
customClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 60 * time.Second,
}

client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithHTTPClient(customClient),
)
```

**Use Cases:**
- Custom TLS configuration
- Proxy configuration
- Connection pooling tuning
- Custom transport middleware

**Example with Proxy:**

```go
proxyURL, _ := url.Parse("http://proxy.example.com:8080")
customClient := &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}

client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithHTTPClient(customClient),
)
```

**Example with Custom TLS:**

```go
tlsConfig := &tls.Config{
    InsecureSkipVerify: false,
    MinVersion:         tls.VersionTLS12,
}

customClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: tlsConfig,
    },
}

client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithHTTPClient(customClient),
)
```

## Batch Operations

### Automatic Batching

The client automatically batches bulk operations based on the configured batch size.

```go
// Configure batch size
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithBatchSize(50),
)

project := client.Project("myproject")

// Create 200 work items
// Automatically batched into 4 requests (200 / 50)
items := make([]*polarion.WorkItem, 200)
for i := range items {
    items[i] = &polarion.WorkItem{
        Type: "workitems",
        Attributes: &polarion.WorkItemAttributes{
            Title: fmt.Sprintf("Item %d", i+1),
        },
    }
}

err = project.WorkItems.Create(ctx, items...)
```

### Batch Size Considerations

**Small Batch Size (10-25):**
- ✅ Lower memory usage
- ✅ Faster individual request completion
- ✅ Better for rate-limited APIs
- ❌ More API calls
- ❌ Higher total latency

**Medium Batch Size (50-100):**
- ✅ Good balance
- ✅ Reasonable memory usage
- ✅ Acceptable latency
- ✅ Recommended default

**Large Batch Size (200+):**
- ✅ Fewer API calls
- ✅ Lower total latency
- ❌ Higher memory usage
- ❌ Longer individual requests
- ❌ May hit API limits

## Pagination

### Automatic Pagination

Use `QueryAll` for automatic pagination:

```go
// Fetches all results automatically
items, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement",
    polarion.WithFields(polarion.FieldsBasic),
)
```

### Manual Pagination

Use `Query` for manual pagination control:

```go
pageNumber := 1
pageSize := 50

for {
    result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
        Query:      "type:requirement",
        PageSize:   pageSize,
        PageNumber: pageNumber,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Process items
    for _, item := range result.Items {
        fmt.Printf("Item: %s\n", item.ID)
    }
    
    // Check if more pages
    if !result.HasNext {
        break
    }
    
    pageNumber++
}
```

### Pagination Options

```go
// Query with pagination options
result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
    Query:      "type:requirement",
    PageSize:   100,           // Items per page
    PageNumber: 1,             // Page number (1-indexed)
    Fields:     fieldSelector, // Field selection
})
```

## Retry Logic

### Retry Behavior

The client automatically retries failed requests with exponential backoff:

1. **Retryable Errors:**
   - HTTP 5xx errors (server errors)
   - Network timeouts
   - Connection errors
   - DNS resolution failures

2. **Non-Retryable Errors:**
   - HTTP 4xx errors (client errors)
   - Authentication failures
   - Validation errors
   - Context cancellation

### Exponential Backoff

Wait time between retries increases exponentially:

```
Retry 1: MinWait + jitter
Retry 2: MinWait * 2 + jitter
Retry 3: MinWait * 4 + jitter
...
Max: MaxWait
```

**Jitter:** Random value added to prevent thundering herd

### Custom Retry Configuration

```go
// Conservative retry (slow but reliable)
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 5,
        MinWait:    5 * time.Second,
        MaxWait:    120 * time.Second,
    }),
)

// Aggressive retry (fast but may overwhelm server)
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 3,
        MinWait:    500 * time.Millisecond,
        MaxWait:    5 * time.Second,
    }),
)
```

## Timeouts

### Client-Level Timeout

Set default timeout for all requests:

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithTimeout(60 * time.Second),
)
```

### Context-Level Timeout

Override timeout for specific operations:

```go
// Timeout for specific operation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

items, err := project.WorkItems.QueryAll(ctx, "type:requirement")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Query timed out")
    }
}
```

### Timeout Recommendations

| Operation | Recommended Timeout |
|-----------|-------------------|
| Get single item | 10-30 seconds |
| Query (small) | 30-60 seconds |
| Query (large) | 2-5 minutes |
| Create/Update (single) | 10-30 seconds |
| Create/Update (bulk) | 1-5 minutes |
| Delete | 10-30 seconds |
| Upload attachment | 1-5 minutes |

## Custom HTTP Client

### Connection Pooling

```go
customClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,              // Total idle connections
        MaxIdleConnsPerHost: 10,               // Idle connections per host
        MaxConnsPerHost:     50,               // Max connections per host
        IdleConnTimeout:     90 * time.Second, // Idle connection timeout
    },
}

client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithHTTPClient(customClient),
)
```

### TLS Configuration

```go
tlsConfig := &tls.Config{
    MinVersion:         tls.VersionTLS12,
    InsecureSkipVerify: false,
    // Add custom CA certificates
    RootCAs: certPool,
}

customClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: tlsConfig,
    },
}

client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithHTTPClient(customClient),
)
```

### Proxy Configuration

```go
proxyURL, _ := url.Parse("http://proxy.example.com:8080")

customClient := &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}

client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithHTTPClient(customClient),
)
```

## Field Selection

### Predefined Field Sets

```go
// Basic fields only
items, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement",
    polarion.WithFields(polarion.FieldsBasic),
)

// All fields
items, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement",
    polarion.WithFields(polarion.FieldsAll),
)

// No fields (IDs only)
items, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement",
    polarion.WithFields(polarion.FieldsNone),
)
```

### Custom Field Selection

```go
// Select specific fields
fields := polarion.NewFieldSelector().
    WithWorkItemFields("id,title,status,type").
    WithLinkedWorkItemFields("id,role")

result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
    Query:  "status:open",
    Fields: fields,
})
```

### Field Selection Benefits

- ✅ Reduced response size
- ✅ Faster API responses
- ✅ Lower bandwidth usage
- ✅ Improved performance

## Best Practices

### 1. Use Appropriate Batch Sizes

```go
// Good: Reasonable batch size
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithBatchSize(50),
)

// Avoid: Too large batch size
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithBatchSize(1000), // May cause timeouts
)
```

### 2. Use Context for Cancellation

```go
// Good: Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

items, err := project.WorkItems.QueryAll(ctx, "type:requirement")

// Avoid: No timeout
items, err := project.WorkItems.QueryAll(context.Background(), "type:requirement")
```

### 3. Select Only Needed Fields

```go
// Good: Select specific fields
fields := polarion.NewFieldSelector().
    WithWorkItemFields("id,title,status")

// Avoid: Fetch all fields when not needed
fields := polarion.FieldsAll
```

### 4. Configure Retries Appropriately

```go
// Good: Reasonable retry configuration
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 3,
        MinWait:    1 * time.Second,
        MaxWait:    30 * time.Second,
    }),
)

// Avoid: Too aggressive retries
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 10,
        MinWait:    100 * time.Millisecond,
        MaxWait:    1 * time.Second,
    }),
)
```

### 5. Reuse Client Instances

```go
// Good: Create client once, reuse
client, err := polarion.New(baseURL, bearerToken)
project := client.Project("myproject")

for _, id := range workItemIDs {
    wi, err := project.WorkItems.Get(ctx, id)
    // Process work item
}

// Avoid: Creating new client for each request
for _, id := range workItemIDs {
    client, _ := polarion.New(baseURL, bearerToken)
    project := client.Project("myproject")
    wi, _ := project.WorkItems.Get(ctx, id)
}
```

### 6. Handle Errors Properly

```go
// Good: Check for specific error types
wi, err := project.WorkItems.Get(ctx, "WI-123")
if err != nil {
    if polarion.IsNotFound(err) {
        // Handle not found
        return
    }
    
    var apiErr *polarion.APIError
    if polarion.AsAPIError(err, &apiErr) {
        // Handle API error
        log.Printf("API Error: %s", apiErr.Message)
        return
    }
    
    // Handle other errors
    log.Fatal(err)
}

// Avoid: Generic error handling
wi, err := project.WorkItems.Get(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}
```

## Environment Variables

While the client doesn't directly support environment variables, you can use them in your application:

```go
package main

import (
    "os"
    "time"
    
    "github.com/almnorth/go-polarion"
)

func main() {
    baseURL := os.Getenv("POLARION_URL")
    token := os.Getenv("POLARION_TOKEN")
    
    client, err := polarion.New(
        baseURL,
        token,
        polarion.WithTimeout(60*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Use client
}
```

## See Also

- [API Reference](API-REFERENCE.md) - Complete API documentation
- [Architecture](ARCHITECTURE.md) - Design principles and architecture
- [Examples](../examples/) - Working code examples
