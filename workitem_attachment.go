// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "time"

// WorkItemAttachment represents a file attachment on a work item.
// It follows the JSON:API format for workitem_attachments resources.
type WorkItemAttachment struct {
	// Type is always "workitem_attachments" for attachments
	Type string `json:"type,omitempty"`

	// ID is the unique identifier (format: "ProjectID/WorkItemID/AttachmentID")
	ID string `json:"id,omitempty"`

	// Revision is the attachment revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all attachment attributes
	Attributes *WorkItemAttachmentAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *WorkItemAttachmentRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *WorkItemAttachmentLinks `json:"links,omitempty"`

	// Meta contains metadata about the attachment
	Meta *WorkItemAttachmentMeta `json:"meta,omitempty"`
}

// WorkItemAttachmentAttributes contains all attachment attributes.
type WorkItemAttachmentAttributes struct {
	// ID is the attachment identifier (without project/workitem prefix)
	ID string `json:"id,omitempty"`

	// FileName is the name of the attached file
	FileName string `json:"fileName,omitempty"`

	// Title is the attachment title (optional, defaults to filename)
	Title string `json:"title,omitempty"`

	// Length is the file size in bytes
	Length int64 `json:"length,omitempty"`

	// Updated is when the attachment was last updated
	Updated *time.Time `json:"updated,omitempty"`
}

// WorkItemAttachmentRelationships contains relationships to other resources.
type WorkItemAttachmentRelationships struct {
	// Author is the user who created the attachment
	Author *Relationship `json:"author,omitempty"`

	// Project is the project containing the attachment
	Project *Relationship `json:"project,omitempty"`
}

// WorkItemAttachmentLinks contains hypermedia links for the attachment.
type WorkItemAttachmentLinks struct {
	// Self is the link to this attachment resource
	Self string `json:"self,omitempty"`

	// Content is the link to download the attachment content
	Content string `json:"content,omitempty"`
}

// WorkItemAttachmentMeta contains metadata about the attachment.
type WorkItemAttachmentMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// AttachmentCreateRequest represents a request to create an attachment.
// This is used internally for multipart form uploads.
type AttachmentCreateRequest struct {
	// FileName is the name of the file to attach
	FileName string

	// Title is the optional title for the attachment
	Title string

	// Content is the file content as bytes
	Content []byte

	// ContentType is the MIME type of the file
	ContentType string
}

// AttachmentUpdateRequest represents a request to update an attachment.
type AttachmentUpdateRequest struct {
	// AttachmentID is the ID of the attachment to update
	AttachmentID string

	// Title is the new title (optional)
	Title string

	// Content is the new file content (optional)
	Content []byte

	// ContentType is the MIME type if content is provided
	ContentType string

	// FileName is the new filename if content is provided
	FileName string
}

// NewAttachmentCreateRequest creates a new attachment creation request.
func NewAttachmentCreateRequest(fileName string, content []byte, contentType string) *AttachmentCreateRequest {
	return &AttachmentCreateRequest{
		FileName:    fileName,
		Content:     content,
		ContentType: contentType,
	}
}

// WithTitle sets the title for the attachment.
func (r *AttachmentCreateRequest) WithTitle(title string) *AttachmentCreateRequest {
	r.Title = title
	return r
}
