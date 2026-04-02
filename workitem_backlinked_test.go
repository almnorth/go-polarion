// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestFieldSelectorToQueryParamsIncludesBacklinkedWorkItems(t *testing.T) {
	params := url.Values{}

	NewFieldSelector().
		WithWorkItemFields("@basic").
		WithBacklinkedWorkItemFields("id,role,suspect").
		WithAttachmentFields("@basic").
		ToQueryParams(params)

	if got := params.Get("fields[linkedworkitems]"); got != "id,role,suspect" {
		t.Fatalf("expected backlinked work item fields to map to linkedworkitems, got %q", got)
	}
	if got := params.Get("fields[backlinkedworkitems]"); got != "" {
		t.Fatalf("expected no backlinkedworkitems sparse field parameter, got %q", got)
	}
}

func TestWorkItemRelationshipsUnmarshalJSONHandlesBacklinkedWorkItems(t *testing.T) {
	data := []byte(`{
		"linkedWorkItems": {
			"links": {
				"related": "https://example.test/linked"
			}
		},
		"backlinkedWorkItems": {
			"links": {
				"related": "https://example.test/backlinked"
			}
		},
		"customReviewer": {
			"data": {
				"type": "users",
				"id": "alice"
			}
		}
	}`)

	var relationships WorkItemRelationships
	if err := json.Unmarshal(data, &relationships); err != nil {
		t.Fatalf("unmarshal relationships: %v", err)
	}

	if relationships.BacklinkedWorkItems == nil {
		t.Fatal("expected BacklinkedWorkItems to be unmarshaled")
	}
	if got := relationships.BacklinkedWorkItems.Links.Related; got != "https://example.test/backlinked" {
		t.Fatalf("unexpected backlink related URL: %q", got)
	}
	if relationships.GetCustomRelationship("backlinkedWorkItems") != nil {
		t.Fatal("backlinkedWorkItems should not be treated as a custom relationship")
	}
	if relationships.GetCustomRelationship("customReviewer") == nil {
		t.Fatal("expected customReviewer to remain in CustomRelationships")
	}
}

func TestWorkItemLinkBacklinkHelpers(t *testing.T) {
	tests := []struct {
		name                string
		link                WorkItemLink
		wantSourceID        string
		wantSourceShortID   string
		wantSourceProjectID string
		wantParentID        string
		wantParentShortID   string
		wantParentProjectID string
	}{
		{
			name: "pre-2512 included backlink without sourceWorkItem",
			link: WorkItemLink{
				ID: "MYPROJ/WI-1/parent/MYPROJ/WI-2",
				Relationships: &LinkedWorkItemRelationships{
					WorkItem: &Relationship{
						Data: map[string]interface{}{
							"type": "workitems",
							"id":   "MYPROJ/WI-1",
						},
					},
				},
			},
			wantSourceID:        "MYPROJ/WI-1",
			wantSourceShortID:   "WI-1",
			wantSourceProjectID: "MYPROJ",
			wantParentID:        "MYPROJ/WI-2",
			wantParentShortID:   "WI-2",
			wantParentProjectID: "MYPROJ",
		},
		{
			name: "2512+ included backlink with sourceWorkItem",
			link: WorkItemLink{
				ID: "MYPROJ/WI-1/parent/MYPROJ/WI-2",
				Relationships: &LinkedWorkItemRelationships{
					WorkItem: &Relationship{
						Data: map[string]interface{}{
							"type": "workitems",
							"id":   "MYPROJ/WI-2",
						},
					},
					SourceWorkItem: &Relationship{
						Data: map[string]interface{}{
							"type": "workitems",
							"id":   "MYPROJ/WI-1",
						},
					},
				},
			},
			wantSourceID:        "MYPROJ/WI-1",
			wantSourceShortID:   "WI-1",
			wantSourceProjectID: "MYPROJ",
			wantParentID:        "MYPROJ/WI-2",
			wantParentShortID:   "WI-2",
			wantParentProjectID: "MYPROJ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.link.GetBacklinkSourceWorkItemID(); got != tt.wantSourceID {
				t.Fatalf("GetBacklinkSourceWorkItemID() = %q, want %q", got, tt.wantSourceID)
			}
			if got := tt.link.GetBacklinkSourceWorkItemIDShort(); got != tt.wantSourceShortID {
				t.Fatalf("GetBacklinkSourceWorkItemIDShort() = %q, want %q", got, tt.wantSourceShortID)
			}
			if got := tt.link.GetBacklinkSourceProjectID(); got != tt.wantSourceProjectID {
				t.Fatalf("GetBacklinkSourceProjectID() = %q, want %q", got, tt.wantSourceProjectID)
			}
			if got := tt.link.GetBacklinkParentWorkItemID(); got != tt.wantParentID {
				t.Fatalf("GetBacklinkParentWorkItemID() = %q, want %q", got, tt.wantParentID)
			}
			if got := tt.link.GetBacklinkParentWorkItemIDShort(); got != tt.wantParentShortID {
				t.Fatalf("GetBacklinkParentWorkItemIDShort() = %q, want %q", got, tt.wantParentShortID)
			}
			if got := tt.link.GetBacklinkParentProjectID(); got != tt.wantParentProjectID {
				t.Fatalf("GetBacklinkParentProjectID() = %q, want %q", got, tt.wantParentProjectID)
			}
		})
	}
}

func TestWorkItemServiceQueryIncludesBacklinkedWorkItemsInlineUsesLinkID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/projects/MYPROJ/workitems" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("include"); got != "backlinkedWorkItems" {
			t.Fatalf("unexpected include parameter: %q", got)
		}
		if got := r.URL.Query().Get("fields[linkedworkitems]"); got != "@all" {
			t.Fatalf("unexpected backlink field selection: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{
					"type": "workitems",
					"id": "MYPROJ/WI-2",
					"attributes": {
						"title": "Target work item"
					}
				}
			],
			"included": [
				{
					"type": "linkedworkitems",
					"id": "MYPROJ/WI-1/relates_to/MYPROJ/WI-2",
					"attributes": {
						"role": "relates_to"
					},
					"relationships": {
						"workItem": {
							"data": {
								"type": "workitems",
								"id": "MYPROJ/WI-1"
							}
						},
						"sourceWorkItem": {
							"data": {
								"type": "workitems",
								"id": "MYPROJ/WI-1"
							}
						}
					}
				}
			],
			"meta": {
				"totalCount": 1
			}
		}`))
	}))
	defer server.Close()

	client, err := New(
		server.URL+"/rest/v1",
		"token",
		WithHTTPClient(server.Client()),
		WithRetryConfig(RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	result, err := client.Project("MYPROJ").WorkItems.Query(context.Background(), QueryOptions{
		Query:   "id:WI-2",
		Fields:  FieldsAll,
		Include: []string{"backlinkedWorkItems"},
	})
	if err != nil {
		t.Fatalf("query work items: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected one work item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if len(item.BacklinkedWorkItemsInline) != 1 {
		t.Fatalf("expected one backlink, got %d", len(item.BacklinkedWorkItemsInline))
	}
	if got := item.BacklinkedWorkItemsInline[0].GetSecondaryWorkItemID(); got != "MYPROJ/WI-1" {
		t.Fatalf("expected pre-2512 workItem semantics in payload, got %q", got)
	}
}

func TestWorkItemServiceGetIncludesBacklinkedWorkItemsInline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/projects/MYPROJ/workitems/WI-2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("include"); got != "backlinkedWorkItems" {
			t.Fatalf("unexpected include parameter: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"type": "workitems",
				"id": "MYPROJ/WI-2",
				"attributes": {
					"title": "Target work item"
				}
			},
			"included": [
				{
					"type": "linkedworkitems",
					"id": "MYPROJ/WI-1/parent/MYPROJ/WI-2",
					"attributes": {
						"role": "parent"
					},
					"relationships": {
						"workItem": {
							"data": {
								"type": "workitems",
								"id": "MYPROJ/WI-2"
							}
						},
						"sourceWorkItem": {
							"data": {
								"type": "workitems",
								"id": "MYPROJ/WI-1"
							}
						}
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client, err := New(
		server.URL+"/rest/v1",
		"token",
		WithHTTPClient(server.Client()),
		WithRetryConfig(RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	item, err := client.Project("MYPROJ").WorkItems.Get(
		context.Background(),
		"MYPROJ/WI-2",
		WithGetInclude("backlinkedWorkItems"),
	)
	if err != nil {
		t.Fatalf("get work item: %v", err)
	}

	if len(item.BacklinkedWorkItemsInline) != 1 {
		t.Fatalf("expected one backlink, got %d", len(item.BacklinkedWorkItemsInline))
	}
	if got := item.BacklinkedWorkItemsInline[0].Relationships.SourceWorkItem; got == nil {
		t.Fatal("expected sourceWorkItem relationship to be available")
	}
}

func TestWorkItemServiceGetRelationshipsAcceptsFullWorkItemID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/projects/MYPROJ/workitems/WI-2/relationships/backlinkedWorkItems" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": []
		}`))
	}))
	defer server.Close()

	client, err := New(
		server.URL+"/rest/v1",
		"token",
		WithHTTPClient(server.Client()),
		WithRetryConfig(RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = client.Project("MYPROJ").WorkItems.GetRelationships(
		context.Background(),
		"MYPROJ/WI-2",
		"backlinkedWorkItems",
	)
	if err != nil {
		t.Fatalf("get relationships: %v", err)
	}
}
