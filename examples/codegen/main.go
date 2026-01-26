// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// Package main demonstrates how to use the codegen package programmatically
// to generate type-safe Go structs for custom work items.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polarion "github.com/almnorth/go-polarion"
	"github.com/almnorth/go-polarion/codegen"
)

func main() {
	// Get configuration from environment variables
	url := os.Getenv("POLARION_URL")
	token := os.Getenv("POLARION_TOKEN")
	projectID := os.Getenv("POLARION_PROJECT")

	if url == "" || token == "" || projectID == "" {
		log.Fatal("Please set POLARION_URL, POLARION_TOKEN, and POLARION_PROJECT environment variables")
	}

	// Create Polarion client
	client, err := polarion.New(url, token)
	if err != nil {
		log.Fatalf("Failed to create Polarion client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Example 1: Generate for all work item types
	fmt.Println("=== Example 1: Generate for all types ===")
	config := &codegen.Config{
		OutputDir: "./generated",
		Package:   "generated",
		TypeID:    "", // Empty means all types
		Refresh:   false,
	}

	gen := codegen.NewGenerator(client, projectID, config)
	if err := gen.Generate(ctx); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	fmt.Println("\n✓ Generated code for all types")

	// Example 2: Generate for a specific type
	fmt.Println("\n=== Example 2: Generate for specific type ===")
	config = &codegen.Config{
		OutputDir: "./generated_requirement",
		Package:   "requirement",
		TypeID:    "requirement",
		Refresh:   false,
	}

	gen = codegen.NewGenerator(client, projectID, config)
	if err := gen.Generate(ctx); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	fmt.Println("\n✓ Generated code for requirement type")

	// Example 3: Refresh existing files
	fmt.Println("\n=== Example 3: Refresh existing files ===")
	config = &codegen.Config{
		OutputDir: "./generated",
		Package:   "generated",
		TypeID:    "",
		Refresh:   true, // Refresh mode
	}

	gen = codegen.NewGenerator(client, projectID, config)
	if err := gen.Generate(ctx); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	fmt.Println("\n✓ Refreshed generated code")

	// Example 4: Integration into build process
	fmt.Println("\n=== Example 4: Build integration ===")
	fmt.Println("You can integrate code generation into your build process:")
	fmt.Println("1. Add a go:generate directive to your code:")
	fmt.Println("   //go:generate polarion-codegen --url $POLARION_URL --token $POLARION_TOKEN --project $POLARION_PROJECT")
	fmt.Println("2. Or create a custom build script that calls the codegen package")
	fmt.Println("3. Or use a Makefile target:")
	fmt.Println("   generate:")
	fmt.Println("       go run ./examples/codegen")
}
