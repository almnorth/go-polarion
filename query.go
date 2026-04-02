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

	// Include specifies related resources to embed inline in the response.
	// For example, []string{"linkedWorkItems"} requests that linked work item
	// entries are included in the response's "included" array, avoiding
	// separate follow-up API calls.
	Include []string
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

	// BacklinkedWorkItems specifies which backlinked work item fields to include
	BacklinkedWorkItems string

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
		BacklinkedWorkItems: "@all",
		WorkItemAttachments: "@all",
	}

	// FieldsDefault requests basic fields plus essential relationship data
	FieldsDefault = &FieldSelector{
		WorkItems:           "@basic",
		LinkedWorkItems:     "id,role,suspect",
		BacklinkedWorkItems: "id,role,suspect",
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

// WithBacklinkedWorkItemFields sets the backlinked work item fields to include.
func (fs *FieldSelector) WithBacklinkedWorkItemFields(fields string) *FieldSelector {
	fs.BacklinkedWorkItems = fields
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
	linkedWorkItemFields := fs.LinkedWorkItems
	if linkedWorkItemFields == "" {
		// Included backlinks are returned as linkedworkitems resources in the
		// general work item endpoints, so they share the same sparse field key.
		linkedWorkItemFields = fs.BacklinkedWorkItems
	}
	if linkedWorkItemFields != "" {
		params.Set("fields[linkedworkitems]", linkedWorkItemFields)
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
	include    []string
}

// defaultQueryOptions returns default query options.
// By default, we request all fields to ensure custom fields are included.
func defaultQueryOptions() queryOptions {
	return queryOptions{
		pageSize: 100,
		fields:   FieldsAll,
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

// WithInclude requests that the given related resources are embedded inline
// in the response "included" array.  For example:
//
//	polarion.WithInclude("linkedWorkItems")
//
// causes the API to return linked-work-item entries alongside each work item,
// which are then surfaced as WorkItem.LinkedWorkItemsInline without requiring
// separate follow-up calls.
func WithInclude(includes ...string) QueryOption {
	return func(o *queryOptions) {
		o.include = append(o.include, includes...)
	}
}

// GetOption is a functional option for Get operations.
type GetOption func(*getOptions)

// getOptions holds internal get configuration.
type getOptions struct {
	fields   *FieldSelector
	revision string
	include  []string
}

// defaultGetOptions returns default get options.
// By default, we request all fields to ensure custom fields are included.
func defaultGetOptions() getOptions {
	return getOptions{
		fields: FieldsAll,
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

// WithGetInclude requests that the given related resources are embedded inline
// in the response "included" array for a Get operation.
func WithGetInclude(includes ...string) GetOption {
	return func(o *getOptions) {
		o.include = append(o.include, includes...)
	}
}
