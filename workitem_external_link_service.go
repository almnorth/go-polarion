// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// WorkItemExternalLinkService provides operations for externally linked work items.
type WorkItemExternalLinkService struct {
	project *ProjectClient
}

// newWorkItemExternalLinkService creates a new external link service.
func newWorkItemExternalLinkService(project *ProjectClient) *WorkItemExternalLinkService {
	return &WorkItemExternalLinkService{
		project: project,
	}
}

// Get retrieves a specific externally linked work item by its ID.
//
// Example:
//
//	link, err := project.WorkItemExternalLinks.Get(ctx, "myproject/WI-123/relates_to/remote-host/otherproject/WI-456")
func (s *WorkItemExternalLinkService) Get(ctx context.Context, linkID string, opts ...GetOption) (*WorkItemExternalLink, error) {
	projectID, primaryWorkItemID, roleID, hostname, targetProjectID, linkedWorkItemID, err := ParseExternalLinkID(linkID)
	if err != nil {
		return nil, err
	}
	if projectID != s.project.projectID {
		return nil, fmt.Errorf("external link %q belongs to project %q, not %q", linkID, projectID, s.project.projectID)
	}

	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/externallylinkedworkitems/%s/%s/%s/%s",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(primaryWorkItemID),
		url.PathEscape(roleID),
		url.PathEscape(hostname),
		url.PathEscape(targetProjectID),
		url.PathEscape(linkedWorkItemID))

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

	var link WorkItemExternalLink
	err = s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &link)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get external link %s: %w", linkID, err)
	}

	return &link, nil
}

// List retrieves all externally linked work items for a specific work item.
//
// Example:
//
//	links, err := project.WorkItemExternalLinks.List(ctx, "WI-123")
func (s *WorkItemExternalLinkService) List(ctx context.Context, workItemID string, opts ...QueryOption) ([]WorkItemExternalLink, error) {
	options := defaultQueryOptions()
	options.pageSize = s.project.client.config.pageSize
	for _, opt := range opts {
		opt(&options)
	}

	cleanWorkItemID := extractWorkItemID(workItemID)
	var allLinks []WorkItemExternalLink
	pageNum := 1

	for {
		urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/externallylinkedworkitems",
			s.project.client.baseURL,
			url.PathEscape(s.project.projectID),
			url.PathEscape(cleanWorkItemID))

		params := url.Values{}
		pageSize := options.pageSize
		if pageSize <= 0 {
			pageSize = s.project.client.config.pageSize
		}
		params.Set("page[size]", strconv.Itoa(pageSize))
		params.Set("page[number]", strconv.Itoa(pageNum))

		if options.fields != nil {
			options.fields.ToQueryParams(params)
		}
		if options.revision != "" {
			params.Set("revision", options.revision)
		}
		if len(options.include) > 0 {
			params.Set("include", strings.Join(options.include, ","))
		}

		urlStr += "?" + params.Encode()

		var response struct {
			Data  []WorkItemExternalLink `json:"data"`
			Links struct {
				Next string `json:"next,omitempty"`
			} `json:"links"`
		}

		err := s.project.client.retrier.Do(ctx, func() error {
			resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "GET", urlStr, nil)
			if err != nil {
				return err
			}
			return internalhttp.DecodeResponse(resp, &response)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list external links for %s: %w", workItemID, err)
		}

		allLinks = append(allLinks, response.Data...)
		if response.Links.Next == "" {
			break
		}
		pageNum++
	}

	return allLinks, nil
}

// Create creates one or more externally linked work items.
//
// Example:
//
//	link := polarion.NewWorkItemExternalLink("relates_to", "https://remote.example.com/polarion/#/project/OTHER/workitem?id=WI-456")
//	err := project.WorkItemExternalLinks.Create(ctx, "WI-123", link)
func (s *WorkItemExternalLinkService) Create(ctx context.Context, primaryWorkItemID string, links ...*WorkItemExternalLink) error {
	if len(links) == 0 {
		return nil
	}

	for i, link := range links {
		if err := s.validateLink(link); err != nil {
			return fmt.Errorf("validation failed for external link %d: %w", i, err)
		}
	}

	cleanWorkItemID := extractWorkItemID(primaryWorkItemID)
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/externallylinkedworkitems",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID))

	requestData := make([]interface{}, len(links))
	for i, link := range links {
		requestData[i] = map[string]interface{}{
			"type": "externallylinkedworkitems",
			"attributes": map[string]interface{}{
				"role":        link.Attributes.Role,
				"workItemURI": link.Attributes.WorkItemURI,
			},
		}
	}

	body := map[string]interface{}{
		"data": requestData,
	}

	var response struct {
		Data []WorkItemExternalLink `json:"data"`
	}
	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "POST", urlStr, body)
		if err != nil {
			return err
		}
		return internalhttp.DecodeResponse(resp, &response)
	})
	if err != nil {
		return fmt.Errorf("failed to create external links: %w", err)
	}

	for i, created := range response.Data {
		if i < len(links) {
			links[i].ID = created.ID
			if created.Links != nil {
				links[i].Links = created.Links
			}
		}
	}

	return nil
}

// Delete deletes one or more externally linked work items by ID.
//
// Example:
//
//	err := project.WorkItemExternalLinks.Delete(ctx, "myproject/WI-123/relates_to/remote-host/otherproject/WI-456")
func (s *WorkItemExternalLinkService) Delete(ctx context.Context, linkIDs ...string) error {
	if len(linkIDs) == 0 {
		return nil
	}

	linksByWorkItem := make(map[string][]string)
	for _, linkID := range linkIDs {
		projectID, workItemID, _, _, _, _, err := ParseExternalLinkID(linkID)
		if err != nil {
			return err
		}
		if projectID != s.project.projectID {
			return fmt.Errorf("external link %q belongs to project %q, not %q", linkID, projectID, s.project.projectID)
		}

		primaryWorkItemID := projectID + "/" + workItemID
		linksByWorkItem[primaryWorkItemID] = append(linksByWorkItem[primaryWorkItemID], linkID)
	}

	for primaryWorkItemID, ids := range linksByWorkItem {
		if err := s.deleteBatch(ctx, primaryWorkItemID, ids); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkItemExternalLinkService) deleteBatch(ctx context.Context, primaryWorkItemID string, linkIDs []string) error {
	cleanWorkItemID := extractWorkItemID(primaryWorkItemID)
	urlStr := fmt.Sprintf("%s/projects/%s/workitems/%s/externallylinkedworkitems",
		s.project.client.baseURL,
		url.PathEscape(s.project.projectID),
		url.PathEscape(cleanWorkItemID))

	linkData := make([]map[string]interface{}, len(linkIDs))
	for i, linkID := range linkIDs {
		linkData[i] = map[string]interface{}{
			"type": "externallylinkedworkitems",
			"id":   linkID,
		}
	}

	body := map[string]interface{}{
		"data": linkData,
	}

	err := s.project.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.project.client.httpClient, "DELETE", urlStr, body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete external links: %w", err)
	}

	return nil
}

func (s *WorkItemExternalLinkService) validateLink(link *WorkItemExternalLink) error {
	if link == nil {
		return NewValidationError("link", "external work item link cannot be nil")
	}
	if link.Attributes == nil {
		return NewValidationError("attributes", "external work item link attributes cannot be nil")
	}
	if link.Attributes.Role == "" {
		return NewValidationError("role", "external work item link role is required")
	}
	if link.Attributes.WorkItemURI == "" {
		return NewValidationError("workItemURI", "external work item link URI is required")
	}
	if link.Type == "" {
		link.Type = "externallylinkedworkitems"
	}

	return nil
}
