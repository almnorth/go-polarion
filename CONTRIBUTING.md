# Contributing to go-polarion

Thank you for your interest in contributing to the Polarion Go client! This document provides guidelines and instructions for contributing.

## How to Contribute

### Reporting Issues

- Check if the issue already exists before creating a new one
- Provide a clear description of the problem
- Include steps to reproduce the issue
- Specify your Go version and operating system
- Include relevant code snippets or error messages

### Suggesting Features

- Open an issue to discuss the feature before implementing it
- Explain the use case and benefits
- Consider backward compatibility

### Pull Requests

1. **Fork the repository** and create a new branch from `main`
2. **Make your changes** following the code style guidelines
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Run tests** and ensure they pass
6. **Submit a pull request** with a clear description

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/go-polarion.git
cd go-polarion

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linters
go vet ./...
go fmt ./...
```

## Code Style Guidelines

### General Principles

- Follow standard Go conventions and idioms
- Keep functions small and focused
- Write clear, self-documenting code
- Add comments for exported types and functions
- Use meaningful variable and function names

### Formatting

- Use `go fmt` to format all code
- Use `go vet` to check for common mistakes
- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines

### Documentation

- All exported types, functions, and methods must have doc comments
- Doc comments should start with the name of the item being documented
- Include examples in doc comments where appropriate
- Keep documentation up to date with code changes

Example:

```go
// Get retrieves a work item by its ID.
// It returns an error if the work item is not found or if the request fails.
//
// Example:
//
//	wi, err := project.WorkItems.Get(ctx, "WI-123")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *WorkItemService) Get(ctx context.Context, id string) (*WorkItem, error) {
    // Implementation
}
```

### Error Handling

- Return errors rather than panicking
- Wrap errors with context using `fmt.Errorf` with `%w`
- Use custom error types for specific error conditions
- Check errors immediately after they occur

### Testing

- Write unit tests for all new functionality
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for high test coverage
- Test error cases as well as success cases

Example test:

```go
func TestWorkItemService_Get(t *testing.T) {
    tests := []struct {
        name    string
        id      string
        want    *WorkItem
        wantErr bool
    }{
        {
            name: "valid work item",
            id:   "WI-123",
            want: &WorkItem{ID: "WI-123"},
        },
        {
            name:    "not found",
            id:      "WI-999",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Project Structure

```
go-polarion/
├── internal/          # Internal packages (not part of public API)
│   └── http/         # HTTP client and retry logic
├── examples/         # Example programs
│   └── basic/       # Basic usage example
├── *.go             # Public API files
├── *_test.go        # Test files
├── doc.go           # Package documentation
├── go.mod           # Go module file
├── LICENSE          # Apache 2.0 license
└── README.md        # Project readme
```

## Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in the imperative mood (e.g., "Add", "Fix", "Update")
- Keep the first line under 72 characters
- Add a blank line followed by a detailed description if needed

Examples:

```
Add support for work item attachments

Implement methods to upload, download, and delete work item attachments.
Includes comprehensive tests and documentation.
```

```
Fix retry logic for rate limit errors

The retry logic was not correctly handling 429 status codes.
Now properly retries with exponential backoff.
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestWorkItemService_Get
```

### Writing Tests

- Place tests in `*_test.go` files
- Use the `testing` package
- Mock external dependencies
- Test both success and error cases
- Use table-driven tests for multiple scenarios

## Documentation

### Package Documentation

- Update `doc.go` when adding major features
- Include examples in documentation
- Keep the README up to date

### API Documentation

- All exported symbols must have doc comments
- Follow Go documentation conventions
- Include usage examples where helpful

## Release Process

Releases are managed by project maintainers. The process includes:

1. Update version in documentation
2. Update CHANGELOG (if present)
3. Create a git tag
4. Push the tag to trigger release automation

## Questions?

If you have questions about contributing, please open an issue for discussion.

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.
