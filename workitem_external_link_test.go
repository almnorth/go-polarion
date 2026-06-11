// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"net/url"
	"testing"
)

func TestParseExternalLinkID(t *testing.T) {
	t.Parallel()

	projectID, workItemID, roleID, hostname, targetProjectID, linkedWorkItemID, err := ParseExternalLinkID(
		"MYPROJ/WI-123/relates_to/remote-host.example.com/OTHER/WI-456",
	)
	if err != nil {
		t.Fatalf("ParseExternalLinkID returned error: %v", err)
	}

	if projectID != "MYPROJ" {
		t.Fatalf("projectID = %q, want %q", projectID, "MYPROJ")
	}
	if workItemID != "WI-123" {
		t.Fatalf("workItemID = %q, want %q", workItemID, "WI-123")
	}
	if roleID != "relates_to" {
		t.Fatalf("roleID = %q, want %q", roleID, "relates_to")
	}
	if hostname != "remote-host.example.com" {
		t.Fatalf("hostname = %q, want %q", hostname, "remote-host.example.com")
	}
	if targetProjectID != "OTHER" {
		t.Fatalf("targetProjectID = %q, want %q", targetProjectID, "OTHER")
	}
	if linkedWorkItemID != "WI-456" {
		t.Fatalf("linkedWorkItemID = %q, want %q", linkedWorkItemID, "WI-456")
	}
}

func TestParseExternalLinkID_Invalid(t *testing.T) {
	t.Parallel()

	if _, _, _, _, _, _, err := ParseExternalLinkID("not-enough-parts"); err == nil {
		t.Fatal("ParseExternalLinkID should reject malformed IDs")
	}
}

func TestBuildExternalLinkID(t *testing.T) {
	t.Parallel()

	got := BuildExternalLinkID("MYPROJ", "WI-123", "relates_to", "remote-host", "OTHER", "WI-456")
	want := "MYPROJ/WI-123/relates_to/remote-host/OTHER/WI-456"
	if got != want {
		t.Fatalf("BuildExternalLinkID = %q, want %q", got, want)
	}
}

func TestNewWorkItemExternalLink(t *testing.T) {
	t.Parallel()

	link := NewWorkItemExternalLink("relates_to", "https://remote.example.com/polarion/#/project/OTHER/workitem?id=WI-456")
	if link.Type != "externallylinkedworkitems" {
		t.Fatalf("Type = %q, want %q", link.Type, "externallylinkedworkitems")
	}
	if link.Attributes == nil {
		t.Fatal("Attributes should be initialized")
	}
	if link.Attributes.Role != "relates_to" {
		t.Fatalf("Role = %q, want %q", link.Attributes.Role, "relates_to")
	}
	if link.Attributes.WorkItemURI == "" {
		t.Fatal("WorkItemURI should be set")
	}
}

func TestFieldSelectorToQueryParams_ExternalLinks(t *testing.T) {
	t.Parallel()

	params := url.Values{}
	NewFieldSelector().
		WithWorkItemFields("@basic").
		WithExternallyLinkedWorkItemFields("id,role,workItemURI").
		ToQueryParams(params)

	if got := params.Get("fields[workitems]"); got != "@basic" {
		t.Fatalf("fields[workitems] = %q, want %q", got, "@basic")
	}
	if got := params.Get("fields[externallylinkedworkitems]"); got != "id,role,workItemURI" {
		t.Fatalf("fields[externallylinkedworkitems] = %q, want %q", got, "id,role,workItemURI")
	}
}
