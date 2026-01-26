// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package http

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Retrier defines the interface for retry logic.
type Retrier interface {
	Do(ctx context.Context, fn func() error) error
}

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	MaxRetries int
	MinWait    time.Duration
	MaxWait    time.Duration
	RetryIf    func(error) bool
}

// retrier implements exponential backoff retry logic with jitter.
type retrier struct {
	config RetryConfig
}

// NewRetrier creates a new retrier with the given configuration.
func NewRetrier(config RetryConfig) Retrier {
	return &retrier{config: config}
}

// Do executes the given function with retry logic.
// It will retry the function up to maxRetries times if it returns an error
// that satisfies the retryIf condition. Between retries, it waits for an
// exponentially increasing duration with jitter.
func (r *retrier) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return err
		}

		// Execute function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if r.config.RetryIf != nil && !r.config.RetryIf(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		// Calculate backoff with jitter
		backoff := r.calculateBackoff(attempt)

		select {
		case <-time.After(backoff):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateBackoff calculates exponential backoff with jitter.
// The backoff duration is: min * 2^attempt, capped at max, with ±25% jitter.
func (r *retrier) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: min * 2^attempt
	backoff := r.config.MinWait * time.Duration(1<<uint(attempt))

	// Cap at max wait
	if backoff > r.config.MaxWait {
		backoff = r.config.MaxWait
	}

	// Add jitter (±25%)
	// This helps prevent thundering herd problems
	jitterRange := backoff / 2 // 50% of backoff
	jitter := time.Duration(rand.Int63n(int64(jitterRange)))

	// Apply jitter: backoff - 25% + random(0, 50%)
	return backoff - backoff/4 + jitter
}

// noRetrier is a retrier that never retries.
type noRetrier struct{}

// Do executes the function once without retrying.
func (n *noRetrier) Do(ctx context.Context, fn func() error) error {
	return fn()
}

// NewNoRetrier creates a retrier that never retries.
func NewNoRetrier() Retrier {
	return &noRetrier{}
}
