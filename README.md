# go-polarion

A simple yet smart Go client for the Polarion REST API.

[![Go Reference](https://pkg.go.dev/badge/github.com/almnorth/go-polarion.svg)](https://pkg.go.dev/github.com/almnorth/go-polarion)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Features

- ✅ **Simple API** - Clean, idiomatic Go interface
- ✅ **Work Item CRUD** - Full support for create, read, update, delete operations
- ✅ **User Management** - Manage users, avatars, and licenses
- ✅ **User Groups** - Access and manage user groups
- ✅ **Work Item Comments** - Create and manage threaded comments on work items
- ✅ **Enumeration Management** - Manage project enumerations and their values
- ✅ **Work Item Linking** - Create and manage links between work items
- ✅ **Type Definitions** - Access work item type definitions and field metadata
- ✅ **Query & Pagination** - Powerful querying with automatic pagination
- ✅ **Sparse Fields** - Request only the fields you need
- ✅ **Automatic Batching** - Efficient bulk operations
- ✅ **Retry Logic** - Exponential backoff with jitter
- ✅ **Context Support** - Cancellation and timeout support
- ✅ **Type Safe** - Strongly typed models
- ✅ **Zero Dependencies** - Uses only Go standard library

## Installation

```bash
go get github.com/almnorth/go-polarion
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/almnorth/go-polarion"
)

func main() {
    // Create client
    client, err := polarion.New(
        "https://polarion.example.com/rest/v1",
        "your-bearer-token",
        polarion.WithBatchSize(50),
        polarion.WithPageSize(100),
        polarion.WithTimeout(60*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get project client
    project := client.Project("my-project")

    ctx := context.Background()

    // Get a work item
    wi, err := project.WorkItems.Get(ctx, "WI-123")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Work Item: %s - %s\n", wi.ID, wi.Attributes.Title)
}
```

## Usage Examples

### Creating Work Items

```go
// Create a single work item
wi := &polarion.WorkItem{
    Type: "workitems",
    Attributes: &polarion.WorkItemAttributes{
        Title:  "New Security Requirement",
        Status: "draft",
        Description: polarion.NewHTMLContent(
            "<p>All user data must be encrypted at rest</p>",
        ),
    },
}

err := project.WorkItems.Create(ctx, wi)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created work item: %s\n", wi.ID)

// Create multiple work items (automatic batching)
items := make([]*polarion.WorkItem, 150)
for i := range items {
    items[i] = &polarion.WorkItem{
        Type: "workitems",
        Attributes: &polarion.WorkItemAttributes{
            Title:  fmt.Sprintf("Task %d", i+1),
            Status: "open",
        },
    }
}

err = project.WorkItems.Create(ctx, items...)
if err != nil {
    log.Fatal(err)
}
```

### Querying Work Items

```go
// Query with manual pagination
result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
    Query:      "type:requirement AND status:open",
    PageSize:   50,
    PageNumber: 1,
    Fields: &polarion.FieldSelector{
        WorkItems: "@basic",
    },
})
if err != nil {
    log.Fatal(err)
}

for _, wi := range result.Items {
    fmt.Printf("Work Item: %s - %s\n", wi.ID, wi.Attributes.Title)
}

// Query all with automatic pagination
allItems, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement AND status:open",
    polarion.WithFields(polarion.FieldsBasic),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d work items\n", len(allItems))
```

### Updating Work Items

```go
// Get work item
wi, err := project.WorkItems.Get(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

// Modify fields
wi.Attributes.Status = "approved"
wi.Attributes.SetCustomField("priority", "high")

// Update
err = project.WorkItems.Update(ctx, wi)
if err != nil {
    log.Fatal(err)
}
```

### Deleting Work Items

```go
// Delete single work item
err := project.WorkItems.Delete(ctx, "WI-123")

// Delete multiple work items
err = project.WorkItems.Delete(ctx, "WI-123", "WI-124", "WI-125")
```

### Field Selection (Sparse Fields)

```go
// Use predefined field sets
items, err := project.WorkItems.QueryAll(
    ctx,
    "type:requirement",
    polarion.WithFields(polarion.FieldsAll),
)

// Custom field selection
customFields := polarion.NewFieldSelector().
    WithWorkItemFields("id,title,status,type").
    WithLinkedWorkItemFields("id,role")

result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
    Query:  "status:open",
    Fields: customFields,
})
```

### Error Handling

```go
wi, err := project.WorkItems.Get(ctx, "WI-999")
if err != nil {
    // Check for specific error types
    if polarion.IsNotFound(err) {
        fmt.Println("Work item not found")
        return
    }

    var apiErr *polarion.APIError
    if polarion.AsAPIError(err, &apiErr) {
        fmt.Printf("API Error: Status=%d, Message=%s\n",
            apiErr.StatusCode, apiErr.Message)
        for _, detail := range apiErr.Details {
            fmt.Printf("  Detail: %s\n", detail.Detail)
        }
        return
    }

    log.Fatal(err)
}
```

### Context and Cancellation

```go
// Timeout context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

items, err := project.WorkItems.QueryAll(ctx, "type:requirement")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Query timed out")
        return
    }
    log.Fatal(err)
}

// Cancellable operation
ctx, cancel = context.WithCancel(context.Background())

go func() {
    time.Sleep(5 * time.Second)
    cancel()
}()

items, err = project.WorkItems.QueryAll(ctx, "type:requirement")
if err != nil {
    if errors.Is(err, context.Canceled) {
        fmt.Println("Operation cancelled")
        return
    }
    log.Fatal(err)
}
```

### Custom Fields

```go
// Set custom fields
wi.Attributes.SetCustomField("priority", "high")
wi.Attributes.SetCustomField("assignee", "user123")

// Get custom fields
priority := wi.Attributes.GetCustomField("priority")
if priority != nil {
    fmt.Printf("Priority: %v\n", priority)
}

// Check if custom field exists
if wi.Attributes.HasCustomField("assignee") {
    fmt.Println("Assignee is set")
}
```

### Enumerations

```go
// Get a specific enumeration
enum, err := project.Enumerations.Get(ctx, "workitem", "status", "requirement")
if err != nil {
    log.Fatal(err)
}

for _, option := range enum.Attributes.Options {
    fmt.Printf("Option: %s - %s (default: %v)\n",
        option.ID, option.Name, option.Default)
}

// List all enumerations
enums, err := project.Enumerations.List(ctx)
if err != nil {
    log.Fatal(err)
}

// Create a custom enumeration
newEnum := &polarion.Enumeration{
    Type: "enumerations",
    Attributes: &polarion.EnumerationAttributes{
        Options: []polarion.EnumerationOption{
            {ID: "new", Name: "New", Default: true, Color: "#00FF00"},
            {ID: "inprogress", Name: "In Progress", Color: "#FFFF00"},
            {ID: "done", Name: "Done", Color: "#0000FF"},
        },
    },
}
err = project.Enumerations.Create(ctx, newEnum)

// Update enumeration
enum.Attributes.Options = append(enum.Attributes.Options,
    polarion.EnumerationOption{ID: "blocked", Name: "Blocked", Color: "#FF0000"})
err = project.Enumerations.Update(ctx, enum)

// Delete enumeration
err = project.Enumerations.Delete(ctx, "workitem", "customStatus", "requirement")
```

### Work Item Links

```go
// List all links for a work item
links, err := project.WorkItemLinks.List(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

for _, link := range links {
    fmt.Printf("Link: %s (suspect: %v)\n", link.Data.Role, link.Data.Suspect)
}

// Create a link between work items
link := &polarion.WorkItemLink{
    Type: "linkedworkitems",
    Data: &polarion.WorkItemLinkAttributes{
        Role:    "relates_to",
        Suspect: false,
    },
}
err = project.WorkItemLinks.Create(ctx, "WI-123", link)

// Update link (e.g., mark as suspect)
link.Data.Suspect = true
err = project.WorkItemLinks.Update(ctx, link)

// Delete links
err = project.WorkItemLinks.Delete(ctx,
    "myproject/WI-123/relates_to/myproject/WI-456")
```

### Work Item Types

```go
// Get a specific work item type
wiType, err := project.WorkItemTypes.Get(ctx, "requirement")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Type: %s - %s\n", wiType.ID, wiType.Attributes.Name)
fmt.Printf("Icon: %s\n", wiType.Attributes.Icon)

// List all work item types
types, err := project.WorkItemTypes.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, t := range types {
    fmt.Printf("Type: %s - %s\n", t.ID, t.Attributes.Name)
}

// Get field definitions for a type
fields, err := project.WorkItemTypes.GetFields(ctx, "requirement")
if err != nil {
    log.Fatal(err)
}

for _, field := range fields {
    fmt.Printf("Field: %s (%s) - Required: %v\n",
        field.ID, field.Type, field.Required)
}

// Get a specific field definition
field, err := project.WorkItemTypes.GetFieldByID(ctx, "requirement", "status")
if err != nil {
    log.Fatal(err)
}

if field.EnumerationID != "" {
    fmt.Printf("Field uses enumeration: %s\n", field.EnumerationID)
}

// Get all fields by type
fieldsByType, err := project.WorkItemTypes.ListFieldsByType(ctx)
for typeID, fields := range fieldsByType {
    fmt.Printf("Type %s has %d fields\n", typeID, len(fields))
}
```

### Users

```go
// Get a specific user
user, err := client.Users.Get(ctx, "user123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s (%s)\n", user.Attributes.Name, user.Attributes.Email)

// List all users
users, err := client.Users.List(ctx)
if err != nil {
    log.Fatal(err)
}

// List users with query
users, err = client.Users.List(ctx, polarion.WithQuery("disabled:false"))

// Create a new user
newUser := &polarion.User{
    Type: "users",
    ID:   "newuser",
    Attributes: &polarion.UserAttributes{
        Name:  "New User",
        Email: "newuser@example.com",
    },
}
created, err := client.Users.Create(ctx, newUser)

// Update user
user.Attributes.Name = "Updated Name"
err = client.Users.Update(ctx, user)

// Get user avatar
avatar, err := client.Users.GetAvatar(ctx, "user123")
if err != nil {
    log.Fatal(err)
}
// avatar.Data contains the image bytes
// avatar.ContentType contains the MIME type

// Update user avatar
avatarData, _ := os.ReadFile("avatar.png")
err = client.Users.UpdateAvatar(ctx, "user123", avatarData, "image/png")

// Set user license
license := &polarion.License{
    Type: "licenses",
    ID:   "developer",
}
err = client.Users.SetLicense(ctx, "user123", license)
```

### User Groups

```go
// Get a specific user group
group, err := client.UserGroups.Get(ctx, "developers")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Group: %s\n", group.Attributes.Name)

// List all user groups
groups, err := client.UserGroups.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, group := range groups {
    fmt.Printf("Group: %s - %s\n", group.ID, group.Attributes.Name)
}
```

### Work Item Comments

```go
// Get a specific comment
comment, err := project.WorkItemComments.Get(ctx, "WI-123", "comment-456")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Comment by %s: %s\n", comment.Relationships.Author, comment.Attributes.Text.Value)

// List all comments for a work item
comments, err := project.WorkItemComments.List(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

for _, comment := range comments {
    fmt.Printf("Comment: %s\n", comment.Attributes.Text.Value)
}

// Create a new comment
newComment := &polarion.WorkItemComment{
    Type: "workitem_comments",
    Attributes: &polarion.WorkItemCommentAttributes{
        Text: polarion.NewHTMLContent("<p>This is a comment</p>"),
    },
}
created, err := project.WorkItemComments.Create(ctx, "WI-123", newComment)

// Create a threaded comment (reply to another comment)
reply := &polarion.WorkItemComment{
    Type: "workitem_comments",
    Attributes: &polarion.WorkItemCommentAttributes{
        Text: polarion.NewHTMLContent("<p>This is a reply</p>"),
    },
    Relationships: &polarion.WorkItemCommentRelationships{
        ParentComment: &polarion.Relationship{
            Data: map[string]string{
                "type": "workitem_comments",
                "id":   "comment-456",
            },
        },
    },
}
created, err = project.WorkItemComments.Create(ctx, "WI-123", reply)

// Update a comment
comment.Attributes.Text = polarion.NewHTMLContent("<p>Updated comment</p>")
err = project.WorkItemComments.Update(ctx, "WI-123", comment)

// Mark comment as resolved
comment.Attributes.Resolved = true
err = project.WorkItemComments.Update(ctx, "WI-123", comment)
```

## Configuration Options

The client supports various configuration options:

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    // Set batch size for bulk operations
    polarion.WithBatchSize(50),
    
    // Set default page size for queries
    polarion.WithPageSize(100),
    
    // Set maximum request body size
    polarion.WithMaxContentSize(2 * 1024 * 1024), // 2MB
    
    // Configure retry behavior
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 3,
        MinWait:    5 * time.Second,
        MaxWait:    15 * time.Second,
    }),
    
    // Set HTTP client timeout
    polarion.WithTimeout(60 * time.Second),
    
    // Use custom HTTP client
    polarion.WithHTTPClient(customHTTPClient),
)
```

## Examples

For complete working examples, see the [`examples/`](examples/) directory:

- [`examples/basic/main.go`](examples/basic/main.go) - Comprehensive example showing all major features

To run the basic example:

```bash
export POLARION_URL="https://polarion.example.com/rest/v1"
export POLARION_TOKEN="your-bearer-token"
export POLARION_PROJECT="your-project-id"
cd examples/basic
go run main.go
```

## Architecture

The client follows a clean, layered architecture:

- **Client Layer** - Main client and project-scoped operations
- **Service Layer** - Resource-specific operations (WorkItems, Users, etc.)
- **Internal Layer** - HTTP transport, retry logic, and implementation details
  - `internal/http` - HTTP client with authentication and retry logic
- **Models** - Strongly typed data structures

The `internal/` package contains implementation details that are not part of the public API and may change without notice.

## Design Principles

- **Simple but Smart** - Easy to use, but handles complexity internally
- **Idiomatic Go** - Follows Go best practices and conventions
- **Context Everywhere** - Full support for cancellation and timeouts
- **Interface-Based** - Easy to mock and test
- **Zero Dependencies** - Uses only Go standard library
- **Encapsulation** - Internal details hidden from public API

## JSON:API Format

The client follows the Polarion REST API's JSON:API format:

```json
{
  "data": {
    "type": "workitems",
    "id": "myproject/WI-123",
    "attributes": {
      "title": "Example Work Item",
      "status": "open",
      "description": {
        "type": "text/html",
        "value": "<p>Description</p>"
      }
    },
    "relationships": {
      "assignee": {
        "data": {
          "type": "users",
          "id": "user123"
        }
      }
    }
  }
}
```

## Testing

The client is designed to be easily testable through interfaces:

```go
// Mock HTTP client for testing
type mockHTTPClient struct {
    doFunc func(ctx context.Context, req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    return m.doFunc(ctx, req)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [polarion-rest-api-client-python](https://github.com/almnorth/polarion-rest-api-client-python) - Python client for Polarion REST API

## Support

For issues and questions, please use the [GitHub issue tracker](https://github.com/almnorth/go-polarion/issues).
