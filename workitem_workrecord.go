// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"fmt"
	"time"
)

// WorkRecord represents a time tracking record on a work item.
// It follows the JSON:API format for workrecords resources.
type WorkRecord struct {
	// Type is always "workrecords" for work records
	Type string `json:"type,omitempty"`

	// ID is the unique identifier (format: "ProjectID/WorkItemID/RecordID")
	ID string `json:"id,omitempty"`

	// Revision is the work record revision
	Revision string `json:"revision,omitempty"`

	// Attributes contains all work record attributes
	Attributes *WorkRecordAttributes `json:"attributes,omitempty"`

	// Relationships contains links to related resources
	Relationships *WorkRecordRelationships `json:"relationships,omitempty"`

	// Links contains hypermedia links
	Links *WorkRecordLinks `json:"links,omitempty"`

	// Meta contains metadata about the work record
	Meta *WorkRecordMeta `json:"meta,omitempty"`
}

// WorkRecordAttributes contains all work record attributes.
type WorkRecordAttributes struct {
	// ID is the work record identifier (without project/workitem prefix)
	ID string `json:"id,omitempty"`

	// Date is the date when the work was performed
	Date *time.Time `json:"date,omitempty"`

	// TimeSpent is the time spent in Polarion format (e.g., "2h 30m", "1d 4h")
	TimeSpent string `json:"timeSpent,omitempty"`

	// Comment is an optional comment describing the work
	Comment *TextContent `json:"comment,omitempty"`
}

// WorkRecordRelationships contains relationships to other resources.
type WorkRecordRelationships struct {
	// User is the user who logged the time
	User *Relationship `json:"user,omitempty"`

	// Project is the project containing the work record
	Project *Relationship `json:"project,omitempty"`
}

// WorkRecordLinks contains hypermedia links for the work record.
type WorkRecordLinks struct {
	// Self is the link to this work record resource
	Self string `json:"self,omitempty"`
}

// WorkRecordMeta contains metadata about the work record.
type WorkRecordMeta struct {
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// WorkRecordCreateRequest represents a request to create a work record.
type WorkRecordCreateRequest struct {
	// UserID is the ID of the user logging the time
	UserID string

	// Date is the date when the work was performed
	Date time.Time

	// TimeSpent is the time spent (will be formatted to Polarion format)
	TimeSpent TimeSpent

	// Comment is an optional comment
	Comment string
}

// TimeSpent represents a duration for time tracking.
// It provides helper methods to format time in Polarion's expected format.
type TimeSpent struct {
	// Hours is the number of hours
	Hours int

	// Minutes is the number of minutes
	Minutes int
}

// NewTimeSpent creates a new TimeSpent from hours and minutes.
func NewTimeSpent(hours, minutes int) TimeSpent {
	return TimeSpent{
		Hours:   hours,
		Minutes: minutes,
	}
}

// NewTimeSpentFromDuration creates a TimeSpent from a time.Duration.
func NewTimeSpentFromDuration(d time.Duration) TimeSpent {
	totalMinutes := int(d.Minutes())
	return TimeSpent{
		Hours:   totalMinutes / 60,
		Minutes: totalMinutes % 60,
	}
}

// String returns the Polarion-formatted time string (e.g., "2h 30m", "1h", "45m").
func (ts TimeSpent) String() string {
	if ts.Hours > 0 && ts.Minutes > 0 {
		return fmt.Sprintf("%dh %dm", ts.Hours, ts.Minutes)
	} else if ts.Hours > 0 {
		return fmt.Sprintf("%dh", ts.Hours)
	} else if ts.Minutes > 0 {
		return fmt.Sprintf("%dm", ts.Minutes)
	}
	return "0m"
}

// Duration returns the TimeSpent as a time.Duration.
func (ts TimeSpent) Duration() time.Duration {
	return time.Duration(ts.Hours)*time.Hour + time.Duration(ts.Minutes)*time.Minute
}

// NewWorkRecordRequest creates a new work record creation request.
func NewWorkRecordRequest(userID string, date time.Time, timeSpent TimeSpent) *WorkRecordCreateRequest {
	return &WorkRecordCreateRequest{
		UserID:    userID,
		Date:      date,
		TimeSpent: timeSpent,
	}
}

// WithComment sets the comment for the work record.
func (r *WorkRecordCreateRequest) WithComment(comment string) *WorkRecordCreateRequest {
	r.Comment = comment
	return r
}
