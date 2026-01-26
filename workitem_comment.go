// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "time"

// WorkItemComment represents a comment on a Polarion work item following the JSON:API format.
type WorkItemComment struct {
	// Type is always "workitem_comments" for work item comments
	Type string `json:"type,omitempty"`

	// ID is the unique identifier of the comment
	ID string `json:"id,omitempty"`

	// Revision is the comment revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all comment attributes
	Attributes *WorkItemCommentAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *WorkItemCommentRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *WorkItemCommentLinks `json:"links,omitempty"`

	// Meta contains metadata about the comment
	Meta *WorkItemCommentMeta `json:"meta,omitempty"`
}

// WorkItemCommentAttributes contains all work item comment attributes.
type WorkItemCommentAttributes struct {
	// Text is the comment text content
	Text *TextContent `json:"text,omitempty"`

	// Title is the comment title
	Title string `json:"title,omitempty"`

	// Created is when the comment was created
	Created *time.Time `json:"created,omitempty"`

	// Updated is when the comment was last updated
	Updated *time.Time `json:"updated,omitempty"`

	// Resolved indicates if the comment is resolved
	Resolved bool `json:"resolved,omitempty"`

	// ResolvedOn is when the comment was resolved
	ResolvedOn *time.Time `json:"resolvedOn,omitempty"`

	// ChildCommentIds contains IDs of child comments (for threaded comments)
	ChildCommentIds []string `json:"childCommentIds,omitempty"`
}

// WorkItemCommentRelationships contains relationships to other resources.
type WorkItemCommentRelationships struct {
	// Author is the relationship to the comment author
	Author *Relationship `json:"author,omitempty"`

	// ParentComment is the relationship to the parent comment (for threaded comments)
	ParentComment *Relationship `json:"parentComment,omitempty"`

	// ChildComments is the relationship to child comments
	ChildComments *Relationship `json:"childComments,omitempty"`

	// WorkItem is the relationship to the work item
	WorkItem *Relationship `json:"workItem,omitempty"`

	// Project is the relationship to the project
	Project *Relationship `json:"project,omitempty"`
}

// WorkItemCommentLinks contains hypermedia links for the comment.
type WorkItemCommentLinks struct {
	Self string `json:"self,omitempty"`
}

// WorkItemCommentMeta contains metadata about the comment.
type WorkItemCommentMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}
