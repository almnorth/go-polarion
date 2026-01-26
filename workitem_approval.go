// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "time"

// WorkItemApproval represents an approval on a work item.
// It follows the JSON:API format for workitem_approvals resources.
type WorkItemApproval struct {
	// Type is always "workitem_approvals" for approvals
	Type string `json:"type,omitempty"`

	// ID is the unique identifier (format: "ProjectID/WorkItemID/UserID")
	ID string `json:"id,omitempty"`

	// Revision is the approval revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all approval attributes
	Attributes *WorkItemApprovalAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *WorkItemApprovalRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *WorkItemApprovalLinks `json:"links,omitempty"`

	// Meta contains metadata about the approval
	Meta *WorkItemApprovalMeta `json:"meta,omitempty"`
}

// WorkItemApprovalAttributes contains all approval attributes.
type WorkItemApprovalAttributes struct {
	// Status is the approval status (approved, disapproved, waiting)
	Status ApprovalStatus `json:"status,omitempty"`

	// Comment is an optional comment for the approval
	Comment string `json:"comment,omitempty"`

	// Date is when the approval was made
	Date *time.Time `json:"date,omitempty"`
}

// ApprovalStatus represents the status of an approval.
type ApprovalStatus string

// Approval status constants.
const (
	// ApprovalStatusApproved indicates the work item is approved
	ApprovalStatusApproved ApprovalStatus = "approved"

	// ApprovalStatusDisapproved indicates the work item is disapproved/rejected
	ApprovalStatusDisapproved ApprovalStatus = "disapproved"

	// ApprovalStatusWaiting indicates the approval is pending
	ApprovalStatusWaiting ApprovalStatus = "waiting"
)

// WorkItemApprovalRelationships contains relationships to other resources.
type WorkItemApprovalRelationships struct {
	// User is the user who provided the approval
	User *Relationship `json:"user,omitempty"`

	// Project is the project containing the approval
	Project *Relationship `json:"project,omitempty"`
}

// WorkItemApprovalLinks contains hypermedia links for the approval.
type WorkItemApprovalLinks struct {
	// Self is the link to this approval resource
	Self string `json:"self,omitempty"`
}

// WorkItemApprovalMeta contains metadata about the approval.
type WorkItemApprovalMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// ApprovalCreateRequest represents a request to create an approval.
type ApprovalCreateRequest struct {
	// UserID is the ID of the user to request approval from
	UserID string

	// Status is the initial status (typically "waiting")
	Status ApprovalStatus

	// Comment is an optional comment
	Comment string
}

// ApprovalUpdateRequest represents a request to update an approval.
type ApprovalUpdateRequest struct {
	// UserID is the ID of the user whose approval to update
	UserID string

	// Status is the new approval status
	Status ApprovalStatus

	// Comment is an optional comment
	Comment string
}

// NewApprovalRequest creates a new approval request for a user.
func NewApprovalRequest(userID string) *ApprovalCreateRequest {
	return &ApprovalCreateRequest{
		UserID: userID,
		Status: ApprovalStatusWaiting,
	}
}

// WithComment sets the comment for the approval request.
func (r *ApprovalCreateRequest) WithComment(comment string) *ApprovalCreateRequest {
	r.Comment = comment
	return r
}

// NewApprovalUpdate creates a new approval update request.
func NewApprovalUpdate(userID string, status ApprovalStatus) *ApprovalUpdateRequest {
	return &ApprovalUpdateRequest{
		UserID: userID,
		Status: status,
	}
}

// WithUpdateComment sets the comment for the approval update.
func (r *ApprovalUpdateRequest) WithUpdateComment(comment string) *ApprovalUpdateRequest {
	r.Comment = comment
	return r
}
