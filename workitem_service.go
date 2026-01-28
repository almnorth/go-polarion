// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// WorkItemService provides operations for work items.
type WorkItemService struct {
	project *ProjectClient
}

// newWorkItemService creates a new work item service.
func newWorkItemService(project *ProjectClient) *WorkItemService {
	return &WorkItemService{
		project: project,
	}
}

// Get retrieves a single work item by ID.
// The ID should be in the format "PROJECT_ID/WORK_ITEM_ID" (e.g., "myproject/WI-123")
// or just "WORK_ITEM_ID" if the project is already scoped.
//
// Example:
//
//	wi, err := project.WorkItems.Get(ctx, "WI-123")
func (s *WorkItemService) Get(ctx context.Context, id string, opts ...GetOption) (*WorkItem, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Extract work item ID from full ID if needed (e.g., "test/TEST-122" -> "TEST-122")
	workItemID := id
	if strings.Contains(workItemID, "/") {
		parts := strings.Split(workItemID, "/")
		workItemID = parts[len(parts)-1]
	}

	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if options.revision != "" {
		params.Set("revision", options.revision)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var wi WorkItem
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &wi)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get work item %s: %w", id, err)
	}

	return &wi, nil
}

// Query retrieves work items matching a query with pagination.
// Returns a single page of results.
//
// Example:
//
//	result, err := project.WorkItems.Query(ctx, polarion.QueryOptions{
//	    Query:      "type:requirement AND status:open",
//	    PageSize:   50,
//	    PageNumber: 1,
//	})
func (s *WorkItemService) Query(ctx context.Context, opts QueryOptions) (*PageResult, error) {
	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems", s.project.client.baseURL, url.PathEscape(s.project.projectID))

	// Build query parameters
	params := url.Values{}
	if opts.Query != "" {
		params.Set("query", opts.Query)
	}

	// Set page size (use default if not specified)
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = s.project.client.config.pageSize
	}
	params.Set("page[size]", strconv.Itoa(pageSize))

	// Set page number (default to 1)
	pageNumber := opts.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	params.Set("page[number]", strconv.Itoa(pageNumber))

	// Add field selection (default to all fields if not specified)
	fields := opts.Fields
	if fields == nil {
		fields = FieldsAll
	}
	fields.ToQueryParams(params)

	// Add revision if specified
	if opts.Revision != "" {
		params.Set("revision", opts.Revision)
	}

	urlStr += "?" + params.Encode()

	// Make request with retry
	var response struct {
		Data  []WorkItem `json:"data"`
		Links struct {
			Next string `json:"next,omitempty"`
		} `json:"links"`
		Meta struct {
			TotalCount int `json:"totalCount,omitempty"`
		} `json:"meta"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &PageResult{
		Items:      response.Data,
		HasNext:    response.Links.Next != "",
		TotalCount: response.Meta.TotalCount,
	}, nil
}

// QueryAll retrieves all work items matching a query with automatic pagination.
// This method handles pagination automatically and returns all matching items.
//
// Example:
//
//	items, err := project.WorkItems.QueryAll(ctx, "type:requirement")
func (s *WorkItemService) QueryAll(ctx context.Context, query string, opts ...QueryOption) ([]WorkItem, error) {
	// Apply options
	options := defaultQueryOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var allItems []WorkItem
	pageNum := 1

	for {
		result, err := s.Query(ctx, QueryOptions{
			Query:      query,
			PageSize:   options.pageSize,
			PageNumber: pageNum,
			Fields:     options.fields,
			Revision:   options.revision,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query page %d: %w", pageNum, err)
		}

		allItems = append(allItems, result.Items...)

		if !result.HasNext {
			break
		}
		pageNum++
	}

	return allItems, nil
}

// Create creates one or more work items with automatic batching.
// The work items will be split into batches based on the configured batch size
// and maximum content size.
//
// Example:
//
//	wi := &polarion.WorkItem{
//	    Type: "workitems",
//	    Attributes: &polarion.WorkItemAttributes{
//	        Title:  "New Requirement",
//	        Status: "open",
//	    },
//	}
//	err := project.WorkItems.Create(ctx, wi)
func (s *WorkItemService) Create(ctx context.Context, items ...*WorkItem) error {
	if len(items) == 0 {
		return nil
	}

	// Validate all items first
	for i, item := range items {
		if err := s.validateWorkItem(item); err != nil {
			return fmt.Errorf("validation failed for item %d: %w", i, err)
		}
	}

	// Split into batches
	batches := s.splitIntoBatches(items)

	// Process each batch
	for i, batch := range batches {
		if err := s.createBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to create batch %d: %w", i, err)
		}
	}

	return nil
}

// Update updates a work item directly without comparison.
// The work item must have an ID set.
// All modifiable fields in the work item will be sent to the API.
// Read-only fields (type, created, updated, resolvedOn) are automatically excluded.
//
// Example:
//
//	wi.Attributes.Status = "approved"
//	err := project.WorkItems.Update(ctx, wi)
func (s *WorkItemService) Update(ctx context.Context, item *WorkItem) error {
	if item.ID == "" {
		return NewValidationError("ID", "work item ID is required for update")
	}

	// Extract work item ID from full ID if needed
	workItemID := item.ID
	if strings.Contains(workItemID, "/") {
		parts := strings.Split(workItemID, "/")
		workItemID = parts[len(parts)-1]
	}

	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Prepare request body - exclude read-only fields
	// Read-only fields: type, created, updated, resolvedOn
	cleanAttrs := &WorkItemAttributes{
		Title:             item.Attributes.Title,
		Description:       item.Attributes.Description,
		Status:            item.Attributes.Status,
		Resolution:        item.Attributes.Resolution,
		Priority:          item.Attributes.Priority,
		Severity:          item.Attributes.Severity,
		DueDate:           item.Attributes.DueDate,
		PlannedStart:      item.Attributes.PlannedStart,
		PlannedEnd:        item.Attributes.PlannedEnd,
		InitialEstimate:   item.Attributes.InitialEstimate,
		RemainingEstimate: item.Attributes.RemainingEstimate,
		TimeSpent:         item.Attributes.TimeSpent,
		OutlineNumber:     item.Attributes.OutlineNumber,
		Hyperlinks:        item.Attributes.Hyperlinks,
		CustomFields:      item.Attributes.CustomFields,
		// Explicitly NOT including: Type, Created, Updated, ResolvedOn
	}

	updateItem := &WorkItem{
		Type:       "workitems",
		ID:         item.ID,
		Attributes: cleanAttrs,
	}

	body := map[string]interface{}{
		"data": updateItem,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		// PATCH may return 204 No Content (empty body) or 200 OK with updated data
		if resp.StatusCode == 204 {
			resp.Body.Close()
			return nil
		}
		// Update the item with the response if body is present
		return internalhttp.DecodeDataResponse(resp, item)
	})

	if err != nil {
		return fmt.Errorf("failed to update work item %s: %w", item.ID, err)
	}

	return nil
}

// UpdateWithOldValue updates a work item by comparing it with the original state.
// Only changed fields are sent to the API.
// The original parameter should be the work item as fetched from the server.
// The updated parameter should be the work item with modifications.
//
// Example:
//
//	original, err := project.WorkItems.Get(ctx, "WI-123")
//	if err != nil {
//	    return err
//	}
//	updated := original
//	updated.Attributes.Status = "approved"
//	err = project.WorkItems.UpdateWithOldValue(ctx, original, updated)
func (s *WorkItemService) UpdateWithOldValue(ctx context.Context, original, updated *WorkItem) error {
	if updated.ID == "" {
		return NewValidationError("ID", "work item ID is required for update")
	}

	// Extract work item ID from full ID if needed
	workItemID := updated.ID
	if strings.Contains(workItemID, "/") {
		parts := strings.Split(workItemID, "/")
		workItemID = parts[len(parts)-1]
	}

	// Compare and get only changed fields
	changedAttrs := s.compareAttributes(original.Attributes, updated.Attributes)

	// If no fields changed, nothing to update
	if changedAttrs == nil {
		return nil
	}

	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Create update item with only changed attributes
	updateItem := &WorkItem{
		Type:       "workitems",
		ID:         updated.ID,
		Attributes: changedAttrs,
	}

	body := map[string]interface{}{
		"data": updateItem,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		// PATCH may return 204 No Content (empty body) or 200 OK with updated data
		if resp.StatusCode == 204 {
			resp.Body.Close()
			return nil
		}
		// Update the item with the response if body is present
		return internalhttp.DecodeDataResponse(resp, updated)
	})

	if err != nil {
		return fmt.Errorf("failed to update work item %s: %w", updated.ID, err)
	}

	return nil
}

// Equals checks if two work items are equal by comparing their attributes.
// Returns true if the work items have identical attributes, false otherwise.
// This uses the same comparison logic as UpdateWithOldValue.
//
// Example:
//
//	if project.WorkItems.Equals(original, updated) {
//	    fmt.Println("No changes detected")
//	}
func (s *WorkItemService) Equals(a, b *WorkItem) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Use the same comparison logic as UpdateWithOldValue
	changedAttrs := s.compareAttributes(a.Attributes, b.Attributes)
	return changedAttrs == nil
}

// EqualsWithDiff checks if two work items are equal and returns the changed attributes if not.
// Returns nil if the work items are equal, otherwise returns the changed attributes.
// This is useful for debugging to see exactly what fields are different.
//
// Example:
//
//	if diff := project.WorkItems.EqualsWithDiff(original, updated); diff != nil {
//	    fmt.Printf("Changed fields: %+v\n", diff)
//	}
func (s *WorkItemService) EqualsWithDiff(a, b *WorkItem) *WorkItemAttributes {
	if a == nil && b == nil {
		return nil
	}
	if a == nil || b == nil {
		// Return a marker to indicate one is nil
		return &WorkItemAttributes{Title: "ONE_IS_NIL"}
	}
	return s.compareAttributes(a.Attributes, b.Attributes)
}

// compareAttributes compares two WorkItemAttributes and returns a new WorkItemAttributes
// containing only the fields that have changed. Returns nil if no changes detected.
func (s *WorkItemService) compareAttributes(current, updated *WorkItemAttributes) *WorkItemAttributes {
	if current == nil || updated == nil {
		return updated
	}

	changed := &WorkItemAttributes{
		CustomFields: make(map[string]interface{}),
	}
	hasChanges := false

	// Compare standard fields
	if updated.Title != "" && updated.Title != current.Title {
		changed.Title = updated.Title
		hasChanges = true
	}

	if !areTextContentsEqual(current.Description, updated.Description) {
		changed.Description = updated.Description
		hasChanges = true
	}

	if updated.Status != "" && updated.Status != current.Status {
		changed.Status = updated.Status
		hasChanges = true
	}

	if updated.Resolution != "" && updated.Resolution != current.Resolution {
		changed.Resolution = updated.Resolution
		hasChanges = true
	}

	if updated.Priority != "" && updated.Priority != current.Priority {
		changed.Priority = updated.Priority
		hasChanges = true
	}

	if updated.Severity != "" && updated.Severity != current.Severity {
		changed.Severity = updated.Severity
		hasChanges = true
	}

	if updated.DueDate != "" && updated.DueDate != current.DueDate {
		changed.DueDate = updated.DueDate
		hasChanges = true
	}

	if !areTimesEqual(current.PlannedStart, updated.PlannedStart) {
		changed.PlannedStart = updated.PlannedStart
		hasChanges = true
	}

	if !areTimesEqual(current.PlannedEnd, updated.PlannedEnd) {
		changed.PlannedEnd = updated.PlannedEnd
		hasChanges = true
	}

	if updated.InitialEstimate != "" && updated.InitialEstimate != current.InitialEstimate {
		changed.InitialEstimate = updated.InitialEstimate
		hasChanges = true
	}

	if updated.RemainingEstimate != "" && updated.RemainingEstimate != current.RemainingEstimate {
		changed.RemainingEstimate = updated.RemainingEstimate
		hasChanges = true
	}

	if updated.TimeSpent != "" && updated.TimeSpent != current.TimeSpent {
		changed.TimeSpent = updated.TimeSpent
		hasChanges = true
	}

	if updated.OutlineNumber != "" && updated.OutlineNumber != current.OutlineNumber {
		changed.OutlineNumber = updated.OutlineNumber
		hasChanges = true
	}

	if !areHyperlinksEqual(current.Hyperlinks, updated.Hyperlinks) {
		changed.Hyperlinks = updated.Hyperlinks
		hasChanges = true
	}

	// Compare custom fields
	for key, updatedValue := range updated.CustomFields {
		currentValue, exists := current.CustomFields[key]
		if !exists || !areCustomFieldValuesEqual(currentValue, updatedValue) {
			changed.CustomFields[key] = updatedValue
			hasChanges = true
		}
	}

	// If no changes detected, return nil
	if !hasChanges {
		return nil
	}

	return changed
}

// areTextContentsEqual compares two TextContent pointers for equality.
func areTextContentsEqual(a, b *TextContent) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Type == b.Type && a.Value == b.Value
}

// areTimesEqual compares two time.Time pointers for equality.
func areTimesEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

// areHyperlinksEqual compares two Hyperlink slices for equality.
func areHyperlinksEqual(a, b []Hyperlink) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}

	// Create a simple comparison - order matters
	for i := range a {
		if a[i].URI != b[i].URI || a[i].Role != b[i].Role {
			return false
		}
	}
	return true
}

// areCustomFieldValuesEqual compares two custom field values for equality.
func areCustomFieldValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Convert both to JSON for deep comparison
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

// Delete deletes one or more work items by ID.
//
// Example:
//
//	err := project.WorkItems.Delete(ctx, "WI-123", "WI-124")
func (s *WorkItemService) Delete(ctx context.Context, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	// Delete each work item
	for _, id := range ids {
		// Extract work item ID from full ID if needed (e.g., "test/TEST-122" -> "TEST-122")
		workItemID := id
		if strings.Contains(workItemID, "/") {
			parts := strings.Split(workItemID, "/")
			workItemID = parts[len(parts)-1]
		}

		urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s",
			s.project.client.baseURL,
			url.PathEscape(s.project.projectID),
			url.PathEscape(workItemID))

		err := s.project.client.retrier.Do(ctx, func() error {
			resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, nil)
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to delete work item %s: %w", id, err)
		}
	}

	return nil
}

// validateWorkItem validates a work item before creation or update.
func (s *WorkItemService) validateWorkItem(item *WorkItem) error {
	if item == nil {
		return NewValidationError("item", "work item cannot be nil")
	}

	if item.Attributes == nil {
		return NewValidationError("attributes", "work item attributes cannot be nil")
	}

	if item.Attributes.Title == "" {
		return NewValidationError("title", "work item title is required")
	}

	// Set type if not set
	if item.Type == "" {
		item.Type = "workitems"
	}

	return nil
}

// splitIntoBatches splits work items into batches based on size and count limits.
func (s *WorkItemService) splitIntoBatches(items []*WorkItem) [][]*WorkItem {
	var batches [][]*WorkItem
	var currentBatch []*WorkItem
	currentSize := 0

	minRequestSize := len(`{"data":[]}`)

	for _, item := range items {
		itemJSON, _ := json.Marshal(item)
		itemSize := len(itemJSON)

		// Check if single item is too large
		if itemSize+minRequestSize > s.project.client.config.maxContentSize {
			// Skip this item or log warning
			continue
		}

		projectedSize := currentSize + itemSize
		if len(currentBatch) > 0 {
			projectedSize += 2 // comma and space
		}

		// Start new batch if size or count limit reached
		if projectedSize >= s.project.client.config.maxContentSize ||
			len(currentBatch) >= s.project.client.config.batchSize {
			batches = append(batches, currentBatch)
			currentBatch = []*WorkItem{item}
			currentSize = minRequestSize + itemSize
		} else {
			currentBatch = append(currentBatch, item)
			currentSize = projectedSize
		}
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}

// createBatch creates a single batch of work items.
func (s *WorkItemService) createBatch(ctx context.Context, items []*WorkItem) error {
	// Build URL
	urlStr := fmt.Sprintf("%s/projects/%s/workitems", s.project.client.baseURL, url.PathEscape(s.project.projectID))

	// Prepare request body
	body := map[string]interface{}{
		"data": items,
	}

	// Make request with retry
	var response struct {
		Data []WorkItem `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return err
	}

	// Update items with created IDs and data
	for i, created := range response.Data {
		if i < len(items) {
			items[i].ID = created.ID
			items[i].Revision = created.Revision
			if created.Links != nil {
				items[i].Links = created.Links
			}
		}
	}

	return nil
}

// GetRelationships retrieves relationships for a work item.
// The relationshipID specifies which relationship to retrieve (e.g., "linkedWorkItems", "attachments").
//
// Example:
//
//	relationships, err := project.WorkItems.GetRelationships(ctx, "WI-123", "linkedWorkItems")
func (s *WorkItemService) GetRelationships(ctx context.Context, workItemID, relationshipID string) (interface{}, error) {
	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/relationships/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(relationshipID))

	// Make request with retry
	var result interface{}
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &result)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get relationships %s for work item %s: %w", relationshipID, workItemID, err)
	}

	return result, nil
}

// CreateRelationships creates relationships for a work item.
//
// Example:
//
//	relationships := []map[string]interface{}{
//	    {"type": "workitems", "id": "MyProject/WI-456"},
//	}
//	err := project.WorkItems.CreateRelationships(ctx, "WI-123", "linkedWorkItems", relationships...)
func (s *WorkItemService) CreateRelationships(ctx context.Context, workItemID, relationshipID string, relationships ...interface{}) error {
	if len(relationships) == 0 {
		return nil
	}

	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/relationships/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(relationshipID))

	// Prepare request body
	body := map[string]interface{}{
		"data": relationships,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create relationships %s for work item %s: %w", relationshipID, workItemID, err)
	}

	return nil
}

// UpdateRelationships updates relationships for a work item.
//
// Example:
//
//	relationships := []map[string]interface{}{
//	    {"type": "workitems", "id": "MyProject/WI-456"},
//	}
//	err := project.WorkItems.UpdateRelationships(ctx, "WI-123", "linkedWorkItems", relationships...)
func (s *WorkItemService) UpdateRelationships(ctx context.Context, workItemID, relationshipID string, relationships ...interface{}) error {
	if len(relationships) == 0 {
		return nil
	}

	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/relationships/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(relationshipID))

	// Prepare request body
	body := map[string]interface{}{
		"data": relationships,
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "PATCH", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update relationships %s for work item %s: %w", relationshipID, workItemID, err)
	}

	return nil
}

// DeleteRelationships deletes relationships for a work item.
//
// Example:
//
//	err := project.WorkItems.DeleteRelationships(ctx, "WI-123", "linkedWorkItems")
func (s *WorkItemService) DeleteRelationships(ctx context.Context, workItemID, relationshipID string) error {
	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/relationships/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID),
		url.PathEscape(relationshipID))

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, nil)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete relationships %s for work item %s: %w", relationshipID, workItemID, err)
	}

	return nil
}

// GetWorkflowActions retrieves available workflow actions for a work item.
//
// Example:
//
//	actions, err := project.WorkItems.GetWorkflowActions(ctx, "WI-123")
func (s *WorkItemService) GetWorkflowActions(ctx context.Context, workItemID string) ([]interface{}, error) {
	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/actions",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Make request with retry
	var response struct {
		Data []interface{} `json:"data"`
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get workflow actions for work item %s: %w", workItemID, err)
	}

	return response.Data, nil
}

// MoveToDocument moves a work item to a specific position in a document.
//
// Example:
//
//	err := project.WorkItems.MoveToDocument(ctx, "WI-123", "DOC-456", 5)
func (s *WorkItemService) MoveToDocument(ctx context.Context, workItemID, documentID string, position int) error {
	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/actions/moveToDocument",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Prepare request body
	fullID := s.buildWorkItemID(workItemID)
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "workitems",
			"id":   fullID,
			"attributes": map[string]interface{}{
				"targetDocument": documentID,
				"position":       position,
			},
		},
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to move work item %s to document %s: %w", workItemID, documentID, err)
	}

	return nil
}

// MoveFromDocument removes a work item from its current document.
//
// Example:
//
//	err := project.WorkItems.MoveFromDocument(ctx, "WI-123")
func (s *WorkItemService) MoveFromDocument(ctx context.Context, workItemID string) error {
	// Build URL - use the project-scoped endpoint
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/actions/moveFromDocument",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(workItemID))

	// Prepare request body
	fullID := s.buildWorkItemID(workItemID)
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "workitems",
			"id":   fullID,
		},
	}

	// Make request with retry
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to move work item %s from document: %w", workItemID, err)
	}

	return nil
}

// buildWorkItemID builds a full work item ID with project prefix if needed.
func (s *WorkItemService) buildWorkItemID(id string) string {
	// If ID already contains project prefix, return as-is
	if strings.Contains(id, "/") {
		return id
	}
	// Otherwise, prepend project ID
	return fmt.Sprintf("%s/%s", s.project.projectID, id)
}
