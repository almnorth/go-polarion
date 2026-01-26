# API Reference

Complete reference for all API operations supported by go-polarion.

## Table of Contents

- [Work Items](#work-items)
- [Work Item Comments](#work-item-comments)
- [Work Item Attachments](#work-item-attachments)
- [Work Item Approvals](#work-item-approvals)
- [Work Item Work Records](#work-item-work-records)
- [Work Item Links](#work-item-links)
- [Work Item Types](#work-item-types)
- [Work Item Relationships](#work-item-relationships)
- [Projects](#projects)
- [Project Templates](#project-templates)
- [Test Parameters](#test-parameters)
- [Users](#users)
- [User Groups](#user-groups)
- [Enumerations](#enumerations)
- [Metadata API](#metadata-api)
- [Fields Metadata API](#fields-metadata-api)
- [Custom Fields API](#custom-fields-api)

## Work Items

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

## Work Item Comments

### Get Comments

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
```

### Create Comments

```go
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
```

### Update Comments

```go
// Update a comment
comment.Attributes.Text = polarion.NewHTMLContent("<p>Updated comment</p>")
err = project.WorkItemComments.Update(ctx, "WI-123", comment)

// Mark comment as resolved
comment.Attributes.Resolved = true
err = project.WorkItemComments.Update(ctx, "WI-123", comment)
```

## Work Item Attachments

### List Attachments

```go
// List all attachments for a work item
attachments, hasNext, err := project.WorkItemAttachments.List(ctx, "WI-123",
    polarion.WithQueryPageSize(50))
if err != nil {
    log.Fatal(err)
}

for _, attachment := range attachments {
    fmt.Printf("Attachment: %s (%d bytes)\n",
        attachment.Attributes.FileName, attachment.Attributes.Length)
}
```

### Upload Attachments

```go
// Upload a file attachment
fileData, _ := os.ReadFile("document.pdf")
req := polarion.NewAttachmentCreateRequest("document.pdf", fileData, "application/pdf").
    WithTitle("Requirements Document")
err = project.WorkItemAttachments.Create(ctx, "WI-123", req)
```

### Download Attachments

```go
// Download attachment content
content, err := project.WorkItemAttachments.GetContent(ctx, "WI-123", "attachment-id")
if err != nil {
    log.Fatal(err)
}
defer content.Close()
data, _ := io.ReadAll(content)
```

### Update Attachments

```go
// Update attachment metadata
updateReq := &polarion.AttachmentUpdateRequest{
    AttachmentID: "attachment-id",
    Title:        "Updated Title",
}
err = project.WorkItemAttachments.Update(ctx, "WI-123", updateReq)
```

### Delete Attachments

```go
// Delete attachments
err = project.WorkItemAttachments.Delete(ctx, "WI-123", "attachment-id-1", "attachment-id-2")
```

## Work Item Approvals

### List Approvals

```go
// List all approvals for a work item
approvals, hasNext, err := project.WorkItemApprovals.List(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

for _, approval := range approvals {
    fmt.Printf("Approval: %s - Status: %s\n",
        approval.Relationships.User, approval.Attributes.Status)
}
```

### Request Approvals

```go
// Request approval from users
req := polarion.NewApprovalRequest("user-id").WithComment("Please review")
err = project.WorkItemApprovals.Create(ctx, "WI-123", req)
```

### Update Approvals

```go
// Update approval status
update := polarion.NewApprovalUpdate("user-id", polarion.ApprovalStatusApproved).
    WithUpdateComment("Looks good!")
err = project.WorkItemApprovals.Update(ctx, "WI-123", update)

// Batch update approvals
updates := []*polarion.ApprovalUpdateRequest{
    polarion.NewApprovalUpdate("user1", polarion.ApprovalStatusApproved),
    polarion.NewApprovalUpdate("user2", polarion.ApprovalStatusApproved),
}
err = project.WorkItemApprovals.UpdateBatch(ctx, "WI-123", updates...)
```

### Delete Approvals

```go
// Delete approvals
err = project.WorkItemApprovals.Delete(ctx, "WI-123", "user-id-1", "user-id-2")
```

## Work Item Work Records

### List Work Records

```go
// List all work records for a work item
records, hasNext, err := project.WorkItemWorkRecords.List(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

for _, record := range records {
    fmt.Printf("Work Record: %s - %s\n",
        record.Attributes.TimeSpent, record.Attributes.Comment)
}
```

### Log Time

```go
// Log time on a work item
timeSpent := polarion.NewTimeSpent(2, 30) // 2 hours 30 minutes
req := polarion.NewWorkRecordRequest("user-id", time.Now(), timeSpent).
    WithComment("Implemented feature X")
err = project.WorkItemWorkRecords.Create(ctx, "WI-123", req)

// Create from duration
duration := 3 * time.Hour
timeSpent = polarion.NewTimeSpentFromDuration(duration)
req = polarion.NewWorkRecordRequest("user-id", time.Now(), timeSpent)
err = project.WorkItemWorkRecords.Create(ctx, "WI-123", req)
```

### Delete Work Records

```go
// Delete work records
err = project.WorkItemWorkRecords.Delete(ctx, "WI-123", "record-id-1", "record-id-2")
```

## Work Item Links

### List Links

```go
// List all links for a work item
links, err := project.WorkItemLinks.List(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

for _, link := range links {
    fmt.Printf("Link: %s (suspect: %v)\n", link.Data.Role, link.Data.Suspect)
}
```

### Create Links

```go
// Create a link between work items
link := &polarion.WorkItemLink{
    Type: "linkedworkitems",
    Data: &polarion.WorkItemLinkAttributes{
        Role:    "relates_to",
        Suspect: false,
    },
}
err = project.WorkItemLinks.Create(ctx, "WI-123", link)
```

### Update Links

```go
// Update link (e.g., mark as suspect)
link.Data.Suspect = true
err = project.WorkItemLinks.Update(ctx, link)
```

### Delete Links

```go
// Delete links
err = project.WorkItemLinks.Delete(ctx,
    "myproject/WI-123/relates_to/myproject/WI-456")
```

## Work Item Types

### Get Type Information

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
```

### Get Field Definitions

```go
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

## Work Item Relationships

### Get Relationships

```go
// Get relationships for a work item
relationships, err := project.WorkItems.GetRelationships(ctx, "WI-123", "linkedWorkItems")
if err != nil {
    log.Fatal(err)
}
```

### Create Relationships

```go
// Create relationships
newRelationships := []map[string]interface{}{
    {"type": "workitems", "id": "MyProject/WI-456"},
    {"type": "workitems", "id": "MyProject/WI-789"},
}
err = project.WorkItems.CreateRelationships(ctx, "WI-123", "linkedWorkItems", newRelationships...)
```

### Update Relationships

```go
// Update relationships
err = project.WorkItems.UpdateRelationships(ctx, "WI-123", "linkedWorkItems", newRelationships...)
```

### Delete Relationships

```go
// Delete relationships
err = project.WorkItems.DeleteRelationships(ctx, "WI-123", "linkedWorkItems")
```

### Workflow Actions

```go
// Get available workflow actions
actions, err := project.WorkItems.GetWorkflowActions(ctx, "WI-123")
if err != nil {
    log.Fatal(err)
}

for _, action := range actions {
    fmt.Printf("Available action: %v\n", action)
}
```

### Document Operations

```go
// Move work item to a document
err = project.WorkItems.MoveToDocument(ctx, "WI-123", "DOC-456", 5)

// Remove work item from document
err = project.WorkItems.MoveFromDocument(ctx, "WI-123")
```

## Projects

### Get Projects

```go
// Get a specific project
project, err := client.Projects.Get(ctx, "myproject")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Project: %s - %s\n", project.ID, project.Attributes.Name)

// List all projects
projects, err := client.Projects.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, proj := range projects {
    fmt.Printf("Project: %s - %s (Active: %v)\n",
        proj.ID, proj.Attributes.Name, proj.Attributes.Active)
}
```

### Create Projects

```go
// Create a new project
req := &polarion.CreateProjectRequest{
    ProjectID:   "newproject",
    Name:        "New Project",
    Description: "Project description",
    TemplateID:  "template_id",
}
project, err = client.Projects.Create(ctx, req)
```

### Update Projects

```go
// Update project
project.Attributes.Description = "Updated description"
updated, err := client.Projects.Update(ctx, project)
```

### Move Projects

```go
// Move project to a different location
err = client.Projects.Move(ctx, "myproject", &polarion.MoveProjectRequest{
    NewLocation: "/new/location",
})
```

### Mark/Unmark Projects

```go
// Mark project as favorite
err = client.Projects.Mark(ctx, "myproject")

// Unmark project
err = client.Projects.Unmark(ctx, "myproject")
```

### Delete Projects

```go
// Delete project
err = client.Projects.Delete(ctx, "myproject")
```

## Project Templates

### List Templates

```go
// List available project templates
templates, err := client.ProjectTemplates.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, template := range templates {
    fmt.Printf("Template: %s - %s\n",
        template.ID, template.Attributes.Name)
}

// Use template when creating a project
req := &polarion.CreateProjectRequest{
    ProjectID:  "newproject",
    Name:       "New Project",
    TemplateID: templates[0].ID, // Use first available template
}
project, err := client.Projects.Create(ctx, req)
```

## Test Parameters

### List Parameters

```go
// Get project client
proj := client.Project("myproject")

// List test parameter definitions
params, err := proj.TestParameters.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, param := range params {
    fmt.Printf("Parameter: %s (%s)\n",
        param.Attributes.Name, param.Attributes.Type)
}
```

### Get Parameter

```go
// Get specific parameter
param, err := proj.TestParameters.Get(ctx, "browser")
if err != nil {
    log.Fatal(err)
}
```

### Create Parameters

```go
// Create test parameter
newParam := &polarion.TestParameter{
    Type: "testparameterdefinitions",
    Attributes: &polarion.TestParameterAttributes{
        Name:          "Browser",
        Type:          "enum",
        AllowedValues: []string{"Chrome", "Firefox", "Safari"},
        DefaultValue:  "Chrome",
        Required:      true,
    },
}
err = proj.TestParameters.Create(ctx, newParam)

// Create multiple parameters
params = []*polarion.TestParameter{
    {
        Type: "testparameterdefinitions",
        Attributes: &polarion.TestParameterAttributes{
            Name: "OS",
            Type: "enum",
            AllowedValues: []string{"Windows", "Linux", "macOS"},
        },
    },
    {
        Type: "testparameterdefinitions",
        Attributes: &polarion.TestParameterAttributes{
            Name: "Version",
            Type: "string",
        },
    },
}
err = proj.TestParameters.Create(ctx, params...)
```

### Delete Parameters

```go
// Delete parameter
err = proj.TestParameters.Delete(ctx, "browser")

// Delete multiple parameters
err = proj.TestParameters.DeleteBatch(ctx, "browser", "os", "version")
```

## Users

### Get Users

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
```

### Create Users

```go
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
```

### Update Users

```go
// Update user
user.Attributes.Name = "Updated Name"
err = client.Users.Update(ctx, user)
```

### User Avatars

```go
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
```

### User Licenses

```go
// Set user license
license := &polarion.License{
    Type: "licenses",
    ID:   "developer",
}
err = client.Users.SetLicense(ctx, "user123", license)
```

## User Groups

### Get Groups

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

## Enumerations

### Project-Scoped Enumerations

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

### Global Enumerations

```go
// Get a specific global enumeration
globalEnum, err := client.GlobalEnumerations.Get(ctx, "workitem", "status", "requirement")
if err != nil {
    log.Fatal(err)
}

for _, option := range globalEnum.Attributes.Options {
    fmt.Printf("Option: %s - %s (default: %v)\n",
        option.ID, option.Name, option.Default)
}

// Use EnumerationID helper
enumID := polarion.NewEnumerationID("workitem", "priority", "task")
globalEnum, err = client.GlobalEnumerations.GetByID(ctx, enumID)

// List all global enumerations
globalEnums, err := client.GlobalEnumerations.List(ctx)
if err != nil {
    log.Fatal(err)
}

// Create a global enumeration
newGlobalEnum := &polarion.Enumeration{
    Type: "enumerations",
    ID:   "enum/workitem/customPriority/requirement",
    Attributes: &polarion.EnumerationAttributes{
        Options: []polarion.EnumerationOption{
            {ID: "critical", Name: "Critical", Default: false, Color: "#FF0000"},
            {ID: "high", Name: "High", Default: false, Color: "#FF8800"},
            {ID: "medium", Name: "Medium", Default: true, Color: "#FFFF00"},
            {ID: "low", Name: "Low", Default: false, Color: "#00FF00"},
        },
    },
}
err = client.GlobalEnumerations.Create(ctx, newGlobalEnum)

// Update global enumeration
globalEnum.Attributes.Options = append(globalEnum.Attributes.Options,
    polarion.EnumerationOption{ID: "urgent", Name: "Urgent", Color: "#CC0000"})
err = client.GlobalEnumerations.Update(ctx, globalEnum)

// Delete global enumeration
err = client.GlobalEnumerations.Delete(ctx, "workitem", "customPriority", "requirement")

// Or delete using EnumerationID
err = client.GlobalEnumerations.DeleteByID(ctx, enumID)
```

## Metadata API

Requires Polarion >= 2512

### Get Instance Metadata

```go
// Get Polarion instance metadata
metadata, err := client.Metadata.Get(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Polarion Version: %s\n", metadata.Attributes.Version)
fmt.Printf("Build: %s\n", metadata.Attributes.Build)
fmt.Printf("Timezone: %s\n", metadata.Attributes.Timezone)

// Get API configuration limits
if metadata.Attributes.APIProperties != nil {
    fmt.Printf("Max Page Size: %d\n", metadata.Attributes.APIProperties.MaxPageSize)
    fmt.Printf("Body Size Limit: %d bytes\n", metadata.Attributes.APIProperties.BodySizeLimit)
}
```

### Version Checking

```go
// Get parsed version information
version, err := client.Metadata.GetVersion(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Version: %d.%d.%d\n", version.Major, version.Minor, version.Patch)

// Check if Polarion version meets minimum requirement
ok, err := client.Metadata.CheckMinVersion(ctx, "25.12")
if err != nil {
    log.Fatal(err)
}
if !ok {
    log.Fatal("Polarion version too old")
}
```

## Fields Metadata API

Requires Polarion >= 2512

### Get Field Metadata

```go
// Get fields metadata for work items of type "requirement" (global context)
metadata, err := client.FieldsMetadata.Get(ctx, "workitems", "requirement")
if err != nil {
    log.Fatal(err)
}

// Iterate through attribute fields
for fieldID, field := range metadata.Data.Attributes {
    fmt.Printf("Field: %s - %s (Type: %s)\n",
        fieldID, field.Label, field.Type.Kind)
}

// Iterate through relationship fields
for fieldID, field := range metadata.Data.Relationships {
    fmt.Printf("Relationship: %s - %s\n", fieldID, field.Label)
    if len(field.Type.TargetResourceTypes) > 0 {
        fmt.Printf("  Target Types: %v\n", field.Type.TargetResourceTypes)
    }
}

// Get fields metadata in project context
proj := client.Project("myproject")
projectMetadata, err := proj.FieldsMetadata.Get(ctx, "workitems", "requirement")
if err != nil {
    log.Fatal(err)
}

// Get fields for all work items (no specific type)
allFieldsMetadata, err := client.FieldsMetadata.Get(ctx, "workitems", "~")
```

## Custom Fields API

Requires Polarion >= 2512

### Global Custom Fields

```go
// Get custom fields configuration for work items of type "requirement"
config, err := client.GlobalCustomFields.Get(ctx, "workitems", "requirement")
if err != nil {
    log.Fatal(err)
}

for _, field := range config.Attributes.Fields {
    fmt.Printf("Custom Field: %s - %s (Type: %s)\n",
        field.ID, field.Name, field.Type.Kind)
    if field.Required {
        fmt.Printf("  Required: yes\n")
    }
    if field.Type.Kind == "enumeration" {
        fmt.Printf("  Enumeration: %s\n", field.Type.EnumName)
    }
}

// Create custom fields configuration
newConfig := polarion.NewCustomFieldsConfig("workitems", "feature")
newConfig.Attributes.Fields = []polarion.CustomFieldDefinition{
    {
        ID:   "businessValue",
        Name: "Business Value",
        Type: polarion.CustomFieldType{Kind: "enumeration", EnumName: "businessValue"},
        Required: true,
    },
    {
        ID:   "targetRelease",
        Name: "Target Release",
        Type: polarion.CustomFieldType{Kind: "date"},
    },
    {
        ID:   "complexityPoints",
        Name: "Complexity Points",
        Type: polarion.CustomFieldType{Kind: "float"},
    },
}
created, err := client.GlobalCustomFields.Create(ctx, newConfig)

// Update custom fields configuration
config.Attributes.Fields = append(config.Attributes.Fields, polarion.CustomFieldDefinition{
    ID:   "securityReviewed",
    Name: "Security Reviewed",
    Type: polarion.CustomFieldType{Kind: "boolean"},
})
err = client.GlobalCustomFields.Update(ctx, "workitems", "feature", config)

// Delete custom fields configuration
err = client.GlobalCustomFields.Delete(ctx, "workitems", "feature")

// Use CustomFieldID helper
id := polarion.CustomFieldID{
    ResourceType: "workitems",
    TargetType:   "requirement",
}
config, err = client.GlobalCustomFields.GetByID(ctx, id)
err = client.GlobalCustomFields.DeleteByID(ctx, id)
```

### Project Custom Fields

```go
// Get custom fields configuration in project context
proj := client.Project("myproject")
config, err := proj.CustomFields.Get(ctx, "workitems", "requirement")
if err != nil {
    log.Fatal(err)
}

// Create project-specific custom fields
newConfig := polarion.NewCustomFieldsConfig("workitems", "task")
newConfig.Attributes.Fields = []polarion.CustomFieldDefinition{
    {
        ID:   "estimatedHours",
        Name: "Estimated Hours",
        Type: polarion.CustomFieldType{Kind: "float"},
    },
    {
        ID:   "assignedTeam",
        Name: "Assigned Team",
        Type: polarion.CustomFieldType{Kind: "string"},
    },
}
created, err := proj.CustomFields.Create(ctx, newConfig)

// Update project custom fields
config.Attributes.Fields = append(config.Attributes.Fields, polarion.CustomFieldDefinition{
    ID:   "priority",
    Name: "Priority",
    Type: polarion.CustomFieldType{Kind: "enumeration", EnumName: "priority"},
    Required: true,
})
err = proj.CustomFields.Update(ctx, "workitems", "task", config)

// Delete project custom fields
err = proj.CustomFields.Delete(ctx, "workitems", "task")

// Use CustomFieldID helper
id := polarion.CustomFieldID{
    ResourceType: "workitems",
    TargetType:   "requirement",
}
config, err = proj.CustomFields.GetByID(ctx, id)
err = proj.CustomFields.UpdateByID(ctx, id, config)
```

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
        for _, detail := range apiErr.Details {
            fmt.Printf("  Detail: %s\n", detail.Detail)
        }
        return
    }

    log.Fatal(err)
}
```

## Context and Cancellation

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

## See Also

- [Configuration](CONFIGURATION.md) - Client configuration options
- [Custom Work Items](CUSTOM-WORKITEMS.md) - Type-safe custom work items
- [Code Generation](CODEGEN.md) - Automatic code generation tool
- [Architecture](ARCHITECTURE.md) - Design principles and architecture
