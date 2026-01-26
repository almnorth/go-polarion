// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package polarion provides a Go client for the Polarion REST API.
// It supports work item CRUD operations, querying with pagination,
// sparse field selection, automatic batching, and retry logic.
package polarion

import (
	"fmt"
	"strings"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// Client is the main Polarion API client.
// It provides access to project-scoped operations through the Project method
// and global operations through service fields.
type Client struct {
	baseURL    string
	httpClient internalhttp.Client
	config     *Config
	retrier    internalhttp.Retrier

	// Users provides access to user management operations
	Users *UserService

	// UserGroups provides access to user group operations
	UserGroups *UserGroupService
}

// New creates a new Polarion API client.
// The baseURL should be the base URL of the Polarion REST API (e.g., "https://polarion.example.com/rest/v1").
// The bearerToken is used for authentication.
// Additional options can be provided to customize the client behavior.
//
// Example:
//
//	client, err := polarion.New(
//	    "https://polarion.example.com/rest/v1",
//	    "your-bearer-token",
//	    polarion.WithBatchSize(50),
//	    polarion.WithPageSize(100),
//	)
func New(baseURL, bearerToken string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if bearerToken == "" {
		return nil, fmt.Errorf("bearerToken cannot be empty")
	}

	// Remove trailing slash from baseURL
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Create default config
	config := defaultConfig()
	config.bearerToken = bearerToken

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create HTTP client
	httpClient := internalhttp.NewClient(config.httpClient, bearerToken)

	// Create retrier
	var retrier internalhttp.Retrier
	if config.retryConfig.MaxRetries > 0 {
		retrier = internalhttp.NewRetrier(config.retryConfig)
	} else {
		retrier = internalhttp.NewNoRetrier()
	}

	client := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		config:     config,
		retrier:    retrier,
	}

	// Initialize global services
	client.Users = newUserService(client)
	client.UserGroups = newUserGroupService(client)

	return client, nil
}

// Project creates a project-scoped client for the given project ID.
// The project ID is used to scope all operations to a specific project.
//
// Example:
//
//	project := client.Project("my-project")
//	wi, err := project.WorkItems.Get(ctx, "WI-123")
func (c *Client) Project(projectID string) *ProjectClient {
	return newProjectClient(c, projectID)
}

// BaseURL returns the base URL of the Polarion API.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Config returns the client configuration.
func (c *Client) Config() *Config {
	return c.config
}
