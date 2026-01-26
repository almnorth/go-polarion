// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"fmt"
	"net/http"
	"time"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// Config holds client configuration options.
type Config struct {
	bearerToken    string
	batchSize      int
	pageSize       int
	maxContentSize int
	retryConfig    internalhttp.RetryConfig
	httpClient     *http.Client
}

// RetryConfig defines retry behavior for failed requests.
type RetryConfig struct {
	MaxRetries int
	MinWait    time.Duration
	MaxWait    time.Duration
	RetryIf    func(error) bool
}

// Option is a functional option for configuring the client.
type Option func(*Config) error

// defaultConfig returns a Config with sensible defaults.
func defaultConfig() *Config {
	return &Config{
		batchSize:      100,
		pageSize:       100,
		maxContentSize: 2 * 1024 * 1024, // 2MB
		retryConfig: internalhttp.RetryConfig{
			MaxRetries: 1,
			MinWait:    5 * time.Second,
			MaxWait:    15 * time.Second,
			RetryIf:    IsRetryable,
		},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithBatchSize sets the batch size for bulk operations.
// The batch size determines how many work items are sent in a single request
// when creating multiple items.
func WithBatchSize(size int) Option {
	return func(c *Config) error {
		if size <= 0 {
			return fmt.Errorf("batch size must be positive, got %d", size)
		}
		c.batchSize = size
		return nil
	}
}

// WithPageSize sets the default page size for queries.
// This controls how many items are returned per page when querying work items.
func WithPageSize(size int) Option {
	return func(c *Config) error {
		if size <= 0 {
			return fmt.Errorf("page size must be positive, got %d", size)
		}
		c.pageSize = size
		return nil
	}
}

// WithMaxContentSize sets the maximum request body size in bytes.
// Requests exceeding this size will be split into multiple batches.
func WithMaxContentSize(size int) Option {
	return func(c *Config) error {
		if size <= 0 {
			return fmt.Errorf("max content size must be positive, got %d", size)
		}
		c.maxContentSize = size
		return nil
	}
}

// WithRetryConfig sets the retry configuration for failed requests.
// This controls exponential backoff behavior when requests fail.
func WithRetryConfig(rc RetryConfig) Option {
	return func(c *Config) error {
		if rc.MaxRetries < 0 {
			return fmt.Errorf("max retries must be non-negative, got %d", rc.MaxRetries)
		}
		if rc.MinWait < 0 {
			return fmt.Errorf("min wait must be non-negative, got %v", rc.MinWait)
		}
		if rc.MaxWait < rc.MinWait {
			return fmt.Errorf("max wait (%v) must be >= min wait (%v)", rc.MaxWait, rc.MinWait)
		}
		c.retryConfig = internalhttp.RetryConfig{
			MaxRetries: rc.MaxRetries,
			MinWait:    rc.MinWait,
			MaxWait:    rc.MaxWait,
			RetryIf:    rc.RetryIf,
		}
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
// Use this to customize transport, TLS configuration, or other HTTP client settings.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Config) error {
		if httpClient == nil {
			return fmt.Errorf("http client cannot be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithTimeout sets the HTTP client timeout.
// This is a convenience method that creates or modifies the HTTP client's timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) error {
		if timeout < 0 {
			return fmt.Errorf("timeout must be non-negative, got %v", timeout)
		}
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
		return nil
	}
}

// BatchSize returns the configured batch size.
func (c *Config) BatchSize() int {
	return c.batchSize
}

// PageSize returns the configured page size.
func (c *Config) PageSize() int {
	return c.pageSize
}

// MaxContentSize returns the configured maximum content size.
func (c *Config) MaxContentSize() int {
	return c.maxContentSize
}

// RetryConfig returns the configured retry configuration.
func (c *Config) RetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: c.retryConfig.MaxRetries,
		MinWait:    c.retryConfig.MinWait,
		MaxWait:    c.retryConfig.MaxWait,
		RetryIf:    c.retryConfig.RetryIf,
	}
}

// HTTPClient returns the configured HTTP client.
func (c *Config) HTTPClient() *http.Client {
	return c.httpClient
}
