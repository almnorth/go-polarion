# go-polarion

A simple yet smart Go client for the Polarion REST API.

[![Go Reference](https://pkg.go.dev/badge/github.com/almnorth/go-polarion.svg)](https://pkg.go.dev/github.com/almnorth/go-polarion)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Features

- **Simple API** - Clean, idiomatic Go interface
- **Comprehensive Coverage** - Work items, projects, users, enumerations, and more ([see API coverage](API-COVERAGE.md))
- **Type-Safe Custom Fields** - Strongly typed custom work item types
- **Code Generation** - Automatic generation from your Polarion configuration (requires 2512+)
- **Query & Pagination** - Powerful querying with automatic pagination
- **Automatic Batching** - Efficient bulk operations
- **Retry Logic** - Exponential backoff with jitter
- **Context Support** - Cancellation and timeout support
- **Zero Dependencies** - Uses only Go standard library

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

    // Query work items
    items, err := project.WorkItems.QueryAll(ctx, "type:requirement AND status:open")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d open requirements\n", len(items))

    // Create a work item
    newWI := &polarion.WorkItem{
        Type: "workitems",
        Attributes: &polarion.WorkItemAttributes{
            Title:  "New Feature Request",
            Status: "draft",
            Description: polarion.NewHTMLContent(
                "<p>Implement new authentication system</p>",
            ),
        },
    }
    err = project.WorkItems.Create(ctx, newWI)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created work item: %s\n", newWI.ID)
}
```

## Usage

### Work Items

Full CRUD operations, querying, and relationships:

```go
// Get work item
wi, err := project.WorkItems.Get(ctx, "WI-123")

// Query with automatic pagination
items, err := project.WorkItems.QueryAll(ctx, "type:requirement AND status:open")

// Create work items (automatic batching)
err = project.WorkItems.Create(ctx, item1, item2, item3)

// Update work item
wi.Attributes.Status = "approved"
err = project.WorkItems.Update(ctx, wi)

// Delete work items
err = project.WorkItems.Delete(ctx, "WI-123", "WI-124")
```

[→ Full Work Items Documentation](docs/API-REFERENCE.md#work-items)

### Type-Safe Custom Fields

Define custom work item types with compile-time type checking:

```go
type Requirement struct {
    base             *polarion.WorkItem
    BusinessValue    *string
    TargetRelease    *polarion.DateOnly
    ComplexityPoints *float64
}

// Load from WorkItem
req := &Requirement{}
wi, _ := project.WorkItems.Get(ctx, "REQ-123")
req.LoadFromWorkItem(wi)

// Type-safe access
if req.BusinessValue != nil {
    fmt.Printf("Business Value: %s\n", *req.BusinessValue)
}
```

[→ Custom Work Items Guide](docs/CUSTOM-WORKITEMS.md)

### Code Generation

Automatically generate type-safe structs from your Polarion configuration (requires Polarion 2512+):

```bash
# Install the tool
go install github.com/almnorth/go-polarion/cmd/polarion-codegen@latest

# Generate for a specific work item type
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject \
  --type requirement

# Generate for all work item types
polarion-codegen \
  --url https://polarion.example.com/rest/v1 \
  --token YOUR_TOKEN \
  --project myproject
```

[→ Code Generation Documentation](docs/CODEGEN.md)

### Projects

Manage projects and project templates:

```go
// List all projects
projects, err := client.Projects.List(ctx)

// Create a new project
req := &polarion.CreateProjectRequest{
    ProjectID:   "newproject",
    Name:        "New Project",
    TemplateID:  "template_id",
}
project, err := client.Projects.Create(ctx, req)

// List available templates
templates, err := client.ProjectTemplates.List(ctx)
```

[→ Projects Documentation](docs/API-REFERENCE.md#projects)

### Users and Groups

Manage users, avatars, licenses, and groups:

```go
// Get user
user, err := client.Users.Get(ctx, "user123")

// List all users
users, err := client.Users.List(ctx)

// Update user avatar
avatarData, _ := os.ReadFile("avatar.png")
err = client.Users.UpdateAvatar(ctx, "user123", avatarData, "image/png")

// List user groups
groups, err := client.UserGroups.List(ctx)
```

[→ Users Documentation](docs/API-REFERENCE.md#users)

### Enumerations

Manage project-scoped and global enumerations:

```go
// Get project enumeration
enum, err := project.Enumerations.Get(ctx, "workitem", "status", "requirement")

// Get global enumeration
globalEnum, err := client.GlobalEnumerations.Get(ctx, "workitem", "priority", "task")

// Create custom enumeration
newEnum := &polarion.Enumeration{
    Type: "enumerations",
    Attributes: &polarion.EnumerationAttributes{
        Options: []polarion.EnumerationOption{
            {ID: "new", Name: "New", Default: true, Color: "#00FF00"},
            {ID: "done", Name: "Done", Color: "#0000FF"},
        },
    },
}
err = project.Enumerations.Create(ctx, newEnum)
```

[→ Enumerations Documentation](docs/API-REFERENCE.md#enumerations)

### Metadata and Custom Fields

Access Polarion metadata and manage custom fields (requires Polarion 2512+):

```go
// Get Polarion version
metadata, err := client.Metadata.Get(ctx)
fmt.Printf("Polarion Version: %s\n", metadata.Attributes.Version)

// Get field metadata
fieldsMeta, err := client.FieldsMetadata.Get(ctx, "workitems", "requirement")

// Manage custom fields
config, err := client.GlobalCustomFields.Get(ctx, "workitems", "requirement")
```

[→ Metadata Documentation](docs/API-REFERENCE.md#metadata-api)

## More Features

### Comments

Create and manage threaded comments:

```go
comment := &polarion.WorkItemComment{
    Type: "workitem_comments",
    Attributes: &polarion.WorkItemCommentAttributes{
        Text: polarion.NewHTMLContent("<p>This is a comment</p>"),
    },
}
created, err := project.WorkItemComments.Create(ctx, "WI-123", comment)
```

[→ Comments Documentation](docs/API-REFERENCE.md#work-item-comments)

### Attachments

Upload, download, and manage attachments:

```go
fileData, _ := os.ReadFile("document.pdf")
req := polarion.NewAttachmentCreateRequest("document.pdf", fileData, "application/pdf")
err = project.WorkItemAttachments.Create(ctx, "WI-123", req)
```

[→ Attachments Documentation](docs/API-REFERENCE.md#work-item-attachments)

### Approvals

Manage approvals:

```go
req := polarion.NewApprovalRequest("user-id").WithComment("Please review")
err = project.WorkItemApprovals.Create(ctx, "WI-123", req)
```

[→ Approvals Documentation](docs/API-REFERENCE.md#work-item-approvals)

### Work Records

Track time spent:

```go
timeSpent := polarion.NewTimeSpent(2, 30) // 2 hours 30 minutes
req := polarion.NewWorkRecordRequest("user-id", time.Now(), timeSpent)
err = project.WorkItemWorkRecords.Create(ctx, "WI-123", req)
```

[→ Work Records Documentation](docs/API-REFERENCE.md#work-item-work-records)

### Links

Link work items together:

```go
link := &polarion.WorkItemLink{
    Type: "linkedworkitems",
    Data: &polarion.WorkItemLinkAttributes{
        Role:    "relates_to",
        Suspect: false,
    },
}
err = project.WorkItemLinks.Create(ctx, "WI-123", link)
```

[→ Links Documentation](docs/API-REFERENCE.md#work-item-links)

### Test Parameters

Define and manage test parameter definitions:

```go
param := &polarion.TestParameter{
    Type: "testparameterdefinitions",
    Attributes: &polarion.TestParameterAttributes{
        Name:          "Browser",
        Type:          "enum",
        AllowedValues: []string{"Chrome", "Firefox", "Safari"},
    },
}
err = project.TestParameters.Create(ctx, param)
```

[→ Test Parameters Documentation](docs/API-REFERENCE.md#test-parameters)

## Configuration

The client supports various configuration options:

```go
client, err := polarion.New(
    baseURL,
    bearerToken,
    polarion.WithBatchSize(50),           // Batch size for bulk operations
    polarion.WithPageSize(100),           // Default page size for queries
    polarion.WithTimeout(60*time.Second), // HTTP client timeout
    polarion.WithRetryConfig(polarion.RetryConfig{
        MaxRetries: 3,
        MinWait:    5 * time.Second,
        MaxWait:    15 * time.Second,
    }),
)
```

[→ Configuration Guide](docs/CONFIGURATION.md)

## Error Handling

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
        return
    }

    log.Fatal(err)
}
```

## Examples

Complete working examples are available in the [`examples/`](examples/) directory:

- [`examples/basic/main.go`](examples/basic/main.go) - Comprehensive example showing all major features
- [`examples/custom_workitems/main.go`](examples/custom_workitems/main.go) - Type-safe custom work item types
- [`examples/codegen/main.go`](examples/codegen/main.go) - Code generation usage

To run the basic example:

```bash
export POLARION_URL="https://polarion.example.com/rest/v1"
export POLARION_TOKEN="your-bearer-token"
export POLARION_PROJECT="your-project-id"
cd examples/basic
go run main.go
```

## Documentation

- [API Reference](docs/API-REFERENCE.md) - Complete API documentation
- [API Coverage](API-COVERAGE.md) - Endpoint coverage and implementation status
- [Configuration Guide](docs/CONFIGURATION.md) - Configuration options
- [Custom Work Items](docs/CUSTOM-WORKITEMS.md) - Type-safe custom work item types
- [Code Generation](docs/CODEGEN.md) - Code generation tool
- [Architecture](docs/ARCHITECTURE.md) - Design and architecture

## Design

The library aims to be simple to use while handling complexity internally. It follows Go best practices, supports context-based cancellation, and has zero external dependencies. See the [architecture documentation](docs/ARCHITECTURE.md) for more details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [polarion-rest-api-client-python](https://github.com/dbinfrago/polarion-rest-api-client) - Python client for Polarion REST API

## Support

For issues and questions, please use the [GitHub issue tracker](https://github.com/almnorth/go-polarion/issues).
