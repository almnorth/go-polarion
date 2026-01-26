// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package codegen provides code generation functionality for creating type-safe
// Go structs from Polarion work item types and their custom fields.
package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	polarion "github.com/almnorth/go-polarion"
)

// Generator handles code generation for Polarion work item types
type Generator struct {
	client    *polarion.Client
	projectID string
	config    *Config
}

// Config holds the generator configuration
type Config struct {
	// OutputDir is the directory where generated files will be written
	OutputDir string

	// Package is the Go package name for generated files
	Package string

	// TypeID is the specific work item type to generate (empty for all types)
	TypeID string

	// Refresh indicates whether to refresh existing files
	Refresh bool
}

// NewGenerator creates a new code generator
func NewGenerator(client *polarion.Client, projectID string, config *Config) *Generator {
	return &Generator{
		client:    client,
		projectID: projectID,
		config:    config,
	}
}

// Generate runs the code generation process
func (g *Generator) Generate(ctx context.Context) error {
	fmt.Println("Starting code generation...")
	fmt.Printf("  Project: %s\n", g.projectID)
	fmt.Printf("  Output: %s\n", g.config.OutputDir)
	fmt.Printf("  Package: %s\n", g.config.Package)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get project client
	project := g.client.Project(g.projectID)

	// Determine which types to generate
	var typeIDs []string
	var err error

	if g.config.TypeID != "" {
		// Single type mode
		fmt.Printf("  Mode: Single type (%s)\n\n", g.config.TypeID)
		typeIDs = []string{g.config.TypeID}
	} else {
		// All types mode
		fmt.Println("  Mode: All types\n")
		typeIDs, err = g.discoverWorkItemTypes(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to discover work item types: %w", err)
		}
		fmt.Printf("Discovered %d work item types\n", len(typeIDs))
	}

	// Generate code for each type
	results := make([]GenerationResult, 0, len(typeIDs))
	for _, typeID := range typeIDs {
		fmt.Printf("\nGenerating code for type: %s\n", typeID)
		result, err := g.generateForType(ctx, project, typeID)
		if err != nil {
			return fmt.Errorf("failed to generate code for type %s: %w", typeID, err)
		}
		results = append(results, result)
		fmt.Printf("  ✓ Generated: %s\n", result.FilePath)
		if result.FieldCount > 0 {
			fmt.Printf("    Fields: %d custom fields\n", result.FieldCount)
		}
	}

	// Generate package documentation if generating all types
	if g.config.TypeID == "" {
		if err := g.generatePackageDoc(results); err != nil {
			return fmt.Errorf("failed to generate package documentation: %w", err)
		}
		fmt.Printf("\n  ✓ Generated: %s\n", filepath.Join(g.config.OutputDir, "doc.go"))
	}

	// Print summary
	g.printSummary(results)

	return nil
}

// GenerationResult holds the result of generating code for a single type
type GenerationResult struct {
	TypeID     string
	TypeName   string
	FilePath   string
	FieldCount int
	IsNew      bool
	Changes    []string
}

// discoverWorkItemTypes discovers all work item types in the project
func (g *Generator) discoverWorkItemTypes(ctx context.Context, project *polarion.ProjectClient) ([]string, error) {
	types, err := project.WorkItemTypes.List(ctx)
	if err != nil {
		return nil, err
	}

	typeIDs := make([]string, 0, len(types))
	for _, t := range types {
		if t.ID != "" {
			typeIDs = append(typeIDs, t.ID)
		}
	}

	return typeIDs, nil
}

// generateForType generates code for a specific work item type
func (g *Generator) generateForType(ctx context.Context, project *polarion.ProjectClient, typeID string) (GenerationResult, error) {
	result := GenerationResult{
		TypeID:   typeID,
		TypeName: toTypeName(typeID),
	}

	// Get fields metadata for this type
	metadata, err := project.FieldsMetadata.Get(ctx, "workitems", typeID)
	if err != nil {
		return result, fmt.Errorf("failed to get fields metadata: %w", err)
	}

	// Get custom field definitions for this type (includes table column info)
	customFieldDef, err := project.CustomFields.Get(ctx, "workitems", typeID)
	if err != nil {
		return result, fmt.Errorf("failed to get custom field definitions: %w", err)
	}

	// Discover custom fields
	discoverer := NewDiscoverer(metadata, customFieldDef)
	fields := discoverer.DiscoverFields()
	result.FieldCount = len(fields)

	// Generate file path
	fileName := strings.ToLower(typeID) + ".go"
	filePath := filepath.Join(g.config.OutputDir, fileName)
	result.FilePath = filePath

	// Check if file exists for refresh mode
	var existingFile *ParsedFile
	if g.config.Refresh {
		if _, err := os.Stat(filePath); err == nil {
			parser := NewParser()
			existingFile, err = parser.Parse(filePath)
			if err != nil {
				return result, fmt.Errorf("failed to parse existing file: %w", err)
			}
			result.IsNew = false
		} else {
			result.IsNew = true
		}
	} else {
		result.IsNew = true
	}

	// Generate code
	tmpl := NewTemplate(g.config.Package, g.projectID, typeID, fields)
	code, err := tmpl.Generate()
	if err != nil {
		return result, fmt.Errorf("failed to generate code: %w", err)
	}

	// Merge with existing file if in refresh mode
	if existingFile != nil {
		code, result.Changes = mergeCode(existingFile, code)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return result, fmt.Errorf("failed to write file: %w", err)
	}

	return result, nil
}

// generatePackageDoc generates package documentation
func (g *Generator) generatePackageDoc(results []GenerationResult) error {
	var sb strings.Builder

	sb.WriteString("// Code generated by polarion-codegen. DO NOT EDIT.\n")
	sb.WriteString(fmt.Sprintf("// Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("// Package %s provides type-safe work item structs for Polarion project %s.\n", g.config.Package, g.projectID))
	sb.WriteString("//\n")
	sb.WriteString("// This package contains auto-generated structs for the following work item types:\n")

	for _, result := range results {
		sb.WriteString(fmt.Sprintf("//   - %s (%s)\n", result.TypeName, result.TypeID))
	}

	sb.WriteString("//\n")
	sb.WriteString("// Each struct provides:\n")
	sb.WriteString("//   - Type-safe access to custom fields\n")
	sb.WriteString("//   - LoadFromWorkItem method to populate from a WorkItem\n")
	sb.WriteString("//   - SaveToWorkItem method to save back to a WorkItem\n")
	sb.WriteString("//   - Getter and setter methods for each custom field\n")
	sb.WriteString(fmt.Sprintf("package %s\n", g.config.Package))

	filePath := filepath.Join(g.config.OutputDir, "doc.go")
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

// printSummary prints a summary of the generation results
func (g *Generator) printSummary(results []GenerationResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Generation Summary")
	fmt.Println(strings.Repeat("=", 60))

	newCount := 0
	updatedCount := 0
	totalFields := 0

	for _, result := range results {
		if result.IsNew {
			newCount++
		} else {
			updatedCount++
		}
		totalFields += result.FieldCount
	}

	fmt.Printf("Total types generated: %d\n", len(results))
	if g.config.Refresh {
		fmt.Printf("  New files: %d\n", newCount)
		fmt.Printf("  Updated files: %d\n", updatedCount)
	}
	fmt.Printf("Total custom fields: %d\n", totalFields)
	fmt.Printf("Output directory: %s\n", g.config.OutputDir)

	if g.config.Refresh && updatedCount > 0 {
		fmt.Println("\nChanges detected:")
		for _, result := range results {
			if len(result.Changes) > 0 {
				fmt.Printf("  %s:\n", result.TypeName)
				for _, change := range result.Changes {
					fmt.Printf("    - %s\n", change)
				}
			}
		}
	}
}

// toTypeName converts a type ID to a Go type name
// Examples: "requirement" -> "Requirement", "test_case" -> "TestCase"
func toTypeName(typeID string) string {
	// Replace underscores and hyphens with spaces
	s := strings.ReplaceAll(typeID, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Title case each word
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	// Join without spaces
	return strings.Join(words, "")
}

// mergeCode merges generated code with existing file, preserving custom sections
func mergeCode(existing *ParsedFile, newCode string) (string, []string) {
	// For now, just return the new code
	// TODO: Implement proper merging logic that preserves custom code sections
	changes := []string{"File refreshed with latest metadata"}
	return newCode, changes
}
