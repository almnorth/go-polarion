# Local Development Setup

This guide explains how to work with the go-polarion module locally for debugging and development.

## Method 1: Using the `replace` directive (Recommended)

The most common and recommended approach is to use the `replace` directive in your `go.mod` file. This tells Go to use the local version instead of the remote version.

### Step 1: Add the replace directive

In your project's `go.mod` file, add a `replace` directive:

```go
module your-project

go 1.25.5

require github.com/almnorth/go-polarion v0.0.0

replace github.com/almnorth/go-polarion => ../go-polarion
```

### Step 2: Run go mod tidy

After adding the replace directive, run:

```bash
go mod tidy
```

This will update your go.sum file and ensure all dependencies are resolved correctly.

### Step 3: Start developing

Now you can make changes to the go-polarion module in the `go-polarion/` directory, and your project will use the local version. Go will automatically pick up your changes.

### Step 4: When you're done

When you're ready to publish your changes, remove the `replace` directive and run `go mod tidy` again.

## Method 2: Using a local import path

You can also use a relative path directly in your import statements:

```go
import (
    "github.com/almnorth/go-polarion"
    // OR
    "path/to/your/workspace/go-polarion"
)
```

However, this requires you to have the module in a specific location relative to your project.

## Method 3: Using Go Workspaces (for multiple projects)

If you're working on multiple projects that depend on go-polarion, you can use Go workspaces:

### Step 1: Create a go.work file

```bash
go work init
go work use ./go-polarion
```

### Step 2: Use the workspace

Now you can use the module from any project in the workspace:

```go
import "github.com/almnorth/go-polarion"
```

## Practical Example

Here's a complete example of how to set this up:

### Project structure:
```
/home/victorien/Documents/
├── my-app/
│   ├── go.mod
│   └── main.go
└── go-polarion/          # The module you're developing
    ├── go.mod
    └── *.go
```

### In my-app/go.mod:

```go
module github.com/yourusername/my-app

go 1.25.5

require github.com/almnorth/go-polarion v0.0.0

replace github.com/almnorth/go-polarion => ../go-polarion
```

### In my-app/main.go:

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
    client, err := polarion.New(
        "https://polarion.example.com/rest/v1",
        "your-bearer-token",
        polarion.WithTimeout(60*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    project := client.Project("my-project")

    // Your code here
    fmt.Println("Using local go-polarion module")
}
```

### Running your app:

```bash
cd my-app
go run main.go
```

Any changes you make to the go-polarion module will be immediately available in your app.

## Debugging

When you make changes to the go-polarion module:

1. **Hot reload**: Go will automatically pick up your changes when you run your app
2. **Debugging**: You can set breakpoints in the go-polarion source files and debug your app
3. **Testing**: Run tests in the go-polarion directory to verify your changes:

```bash
cd go-polarion
go test ./...
```

## Best Practices

1. **Use replace directive**: It's the cleanest and most maintainable approach
2. **Commit the replace directive**: If you're sharing your project with others, commit the replace directive so they can work with the local version
3. **Remove before publishing**: When you're ready to publish your changes, remove the replace directive
4. **Run go mod tidy**: Always run `go mod tidy` after adding or removing the replace directive
5. **Test locally**: Make sure your changes work with your app before publishing

## Troubleshooting

### Issue: "module not found"

**Solution**: Make sure the replace directive is correct and the path points to the go-polarion directory.

### Issue: Changes not reflected

**Solution**: 
- Make sure you're running `go mod tidy` after adding the replace directive
- Try running `go clean -modcache` and then `go mod tidy` again
- Check that the go-polarion directory has a valid go.mod file

### Issue: Version conflicts

**Solution**: Ensure the version in the require directive matches what you're replacing. You can use `v0.0.0` or a specific version.

## Additional Resources

- [Go Modules Documentation](https://go.dev/ref/mod#go-mod-edit)
- [Go Workspaces](https://go.dev/ref/mod#workspaces)