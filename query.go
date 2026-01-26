// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "net/url"

// QueryOptions defines parameters for querying work items.
type QueryOptions struct {
	// Query is the Lucene query string (e.g., "type:requirement AND status:open")
	Query string

	// PageSize is the number of items per page
	PageSize int

	// PageNumber is the page number (1-indexed)
	PageNumber int

	// Fields specifies which fields to include in the response (sparse fields)
	Fields *FieldSelector

	// Revision specifies a specific revision to query
	Revision string
}

// PageResult contains paginated query results.
type PageResult struct {
	// Items contains the work items in this page
	Items []WorkItem

	// HasNext indicates if there are more pages available
	HasNext bool

	// TotalCount is the total number of items (if available)
	TotalCount int
}

// FieldSelector defines sparse field selection for queries.
// This allows you to request only specific fields to reduce response size.
type FieldSelector struct {
	// WorkItems specifies which work item fields to include
	// Use "@basic" for basic fields, "@all" for all fields, or comma-separated field names
	WorkItems string

	// LinkedWorkItems specifies which linked work item fields to include
	LinkedWorkItems string

	// WorkItemAttachments specifies which attachment fields to include
	WorkItemAttachments string
}

// Predefined field selectors for common use cases.
var (
	// FieldsBasic requests only basic work item fields
	FieldsBasic = &FieldSelector{
		WorkItems: "@basic",
	}

	// FieldsAll requests all available fields
	FieldsAll = &FieldSelector{
		WorkItems:           "@all",
		LinkedWorkItems:     "@all",
		WorkItemAttachments: "@all",
	}

	// FieldsDefault requests basic fields plus essential relationship data
	FieldsDefault = &FieldSelector{
		WorkItems:           "@basic",
		LinkedWorkItems:     "id,role,suspect",
		WorkItemAttachments: "@basic",
	}
)

// NewFieldSelector creates a new empty field selector.
func NewFieldSelector() *FieldSelector {
	return &FieldSelector{}
}

// WithWorkItemFields sets the work item fields to include.
func (fs *FieldSelector) WithWorkItemFields(fields string) *FieldSelector {
	fs.WorkItems = fields
	return fs
}

// WithLinkedWorkItemFields sets the linked work item fields to include.
func (fs *FieldSelector) WithLinkedWorkItemFields(fields string) *FieldSelector {
	fs.LinkedWorkItems = fields
	return fs
}

// WithAttachmentFields sets the attachment fields to include.
func (fs *FieldSelector) WithAttachmentFields(fields string) *FieldSelector {
	fs.WorkItemAttachments = fields
	return fs
}

// ToQueryParams converts the field selector to URL query parameters.
func (fs *FieldSelector) ToQueryParams(params url.Values) {
	if fs.WorkItems != "" {
		params.Set("fields[workitems]", fs.WorkItems)
	}
	if fs.LinkedWorkItems != "" {
		params.Set("fields[linkedworkitems]", fs.LinkedWorkItems)
	}
	if fs.WorkItemAttachments != "" {
		params.Set("fields[workitem_attachments]", fs.WorkItemAttachments)
	}
}

// QueryOption is a functional option for configuring queries.
type QueryOption func(*queryOptions)

// queryOptions holds internal query configuration.
type queryOptions struct {
	query      string
	pageSize   int
	pageNumber int
	fields     *FieldSelector
	revision   string
}

// defaultQueryOptions returns default query options.
func defaultQueryOptions() queryOptions {
	return queryOptions{
		pageSize: 100,
		fields:   FieldsDefault,
	}
}

// WithFields sets the field selector for a query.
func WithFields(fields *FieldSelector) QueryOption {
	return func(o *queryOptions) {
		o.fields = fields
	}
}

// WithQueryPageSize sets the page size for a query.
func WithQueryPageSize(size int) QueryOption {
	return func(o *queryOptions) {
		o.pageSize = size
	}
}

// WithPageNumber sets the page number for a query.
func WithPageNumber(number int) QueryOption {
	return func(o *queryOptions) {
		o.pageNumber = number
	}
}

// WithRevision sets the revision for a query.
func WithRevision(revision string) QueryOption {
	return func(o *queryOptions) {
		o.revision = revision
	}
}

// WithQuery sets the query string for filtering.
func WithQuery(query string) QueryOption {
	return func(o *queryOptions) {
		o.query = query
	}
}

// GetOption is a functional option for Get operations.
type GetOption func(*getOptions)

// getOptions holds internal get configuration.
type getOptions struct {
	fields   *FieldSelector
	revision string
}

// defaultGetOptions returns default get options.
func defaultGetOptions() getOptions {
	return getOptions{
		fields: FieldsDefault,
	}
}

// WithGetFields sets the field selector for a Get operation.
func WithGetFields(fields *FieldSelector) GetOption {
	return func(o *getOptions) {
		o.fields = fields
	}
}

// WithGetRevision sets the revision for a Get operation.
func WithGetRevision(revision string) GetOption {
	return func(o *getOptions) {
		o.revision = revision
	}
}
