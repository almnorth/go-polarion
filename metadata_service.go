// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Victorien Elvinger
// Copyright (c) 2025 Siemens AG

package polarion

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	internalhttp "github.com/almnorth/go-polarion/internal/http"
)

// MetadataService handles Polarion instance metadata operations.
// This service provides access to Polarion server information including
// version, build details, and REST API configuration.
//
// Requires: Polarion >= 2512
type MetadataService struct {
	client *Client
}

// Get retrieves Polarion instance metadata including version, build,
// and REST API configuration properties.
//
// Endpoint: GET /metadata
// Requires: Polarion >= 2512
//
// Example:
//
//	metadata, err := client.Metadata.Get(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Polarion version: %s\n", metadata.Attributes.Version)
func (s *MetadataService) Get(ctx context.Context, opts ...GetOption) (*Metadata, error) {
	// Apply options
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Build URL
	urlStr := fmt.Sprintf("%s/metadata", s.client.baseURL)

	// Add query parameters
	params := url.Values{}
	if options.fields != nil {
		options.fields.ToQueryParams(params)
	}
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	// Make request with retry
	var metadata Metadata
	err := s.client.retrier.Do(ctx, func() error {
		resp, err := internalhttp.DoRequest(ctx, s.client.httpClient, "GET", urlStr, nil)
		if err != nil {
			return err
		}
		return internalhttp.DecodeDataResponse(resp, &metadata)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return &metadata, nil
}

// GetVersion retrieves and parses version information from the Polarion instance.
// Returns a VersionInfo struct with Major, Minor, and Patch version numbers.
//
// Example:
//
//	version, err := client.Metadata.GetVersion(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Version: %d.%d.%d\n", version.Major, version.Minor, version.Patch)
func (s *MetadataService) GetVersion(ctx context.Context) (*VersionInfo, error) {
	metadata, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}

	return parseVersion(metadata.Attributes.Version)
}

// CheckMinVersion checks if the Polarion version meets the minimum requirement.
// Returns true if the current version is greater than or equal to minVersion.
//
// The minVersion parameter should be in the format "major.minor" or "major.minor.patch".
//
// Example:
//
//	ok, err := client.Metadata.CheckMinVersion(ctx, "25.12")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if !ok {
//	    log.Fatal("Polarion version too old")
//	}
func (s *MetadataService) CheckMinVersion(ctx context.Context, minVersion string) (bool, error) {
	current, err := s.GetVersion(ctx)
	if err != nil {
		return false, err
	}

	required, err := parseVersion(minVersion)
	if err != nil {
		return false, fmt.Errorf("invalid minimum version format: %w", err)
	}

	return compareVersions(current, required) >= 0, nil
}

// parseVersion parses a version string in the format "major.minor.patch" or "major.minor".
// Returns a VersionInfo struct with the parsed version numbers.
func parseVersion(version string) (*VersionInfo, error) {
	if version == "" {
		return nil, fmt.Errorf("empty version string")
	}

	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid version format: %s (expected major.minor or major.minor.patch)", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch := 0
	if len(parts) == 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return &VersionInfo{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// compareVersions compares two version numbers.
// Returns:
//   - negative if v1 < v2
//   - zero if v1 == v2
//   - positive if v1 > v2
func compareVersions(v1, v2 *VersionInfo) int {
	if v1.Major != v2.Major {
		return v1.Major - v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor - v2.Minor
	}
	return v1.Patch - v2.Patch
}
