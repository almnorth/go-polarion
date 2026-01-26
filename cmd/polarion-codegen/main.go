// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

// polarion-codegen is a CLI tool that leverages the Polarion metadata discovery APIs
// to auto-generate Go structs for custom work items.
//
// Usage:
//
//	polarion-codegen --url <url> --token <token> --project <project> [options]
//
// Options:
//
//	--url          Polarion REST API URL (required)
//	--token        Authentication token (required)
//	--project      Project ID (required)
//	--type         Work item type to generate (optional, if omitted generates all types)
//	--output       Output directory path (default: "./generated")
//	--package      Package name (default: "generated")
//	--refresh      Refresh existing files instead of creating new
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	polarion "github.com/almnorth/go-polarion"
	"github.com/almnorth/go-polarion/codegen"
)

func main() {
	// Parse command-line flags
	var (
		url       string
		token     string
		projectID string
		typeID    string
		outputDir string
		pkgName   string
		refresh   bool
	)

	flag.StringVar(&url, "url", "", "Polarion REST API URL (required)")
	flag.StringVar(&token, "token", "", "Authentication token (required)")
	flag.StringVar(&projectID, "project", "", "Project ID (required)")
	flag.StringVar(&typeID, "type", "", "Work item type to generate (optional, if omitted generates all types)")
	flag.StringVar(&outputDir, "output", "./generated", "Output directory path")
	flag.StringVar(&pkgName, "package", "generated", "Package name")
	flag.BoolVar(&refresh, "refresh", false, "Refresh existing files instead of creating new")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: polarion-codegen [options]\n\n")
		fmt.Fprintf(os.Stderr, "polarion-codegen is a CLI tool that leverages the Polarion metadata discovery APIs\n")
		fmt.Fprintf(os.Stderr, "to auto-generate Go structs for custom work items.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate for a specific work item type\n")
		fmt.Fprintf(os.Stderr, "  polarion-codegen --url https://polarion.example.com/rest/v1 \\\n")
		fmt.Fprintf(os.Stderr, "    --token YOUR_TOKEN --project myproject --type requirement\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate for all work item types in the project\n")
		fmt.Fprintf(os.Stderr, "  polarion-codegen --url https://polarion.example.com/rest/v1 \\\n")
		fmt.Fprintf(os.Stderr, "    --token YOUR_TOKEN --project myproject\n\n")
		fmt.Fprintf(os.Stderr, "  # Refresh existing generated files\n")
		fmt.Fprintf(os.Stderr, "  polarion-codegen --url https://polarion.example.com/rest/v1 \\\n")
		fmt.Fprintf(os.Stderr, "    --token YOUR_TOKEN --project myproject --refresh\n")
	}

	flag.Parse()

	// Validate required flags
	if url == "" {
		log.Fatal("Error: --url is required")
	}
	if token == "" {
		log.Fatal("Error: --token is required")
	}
	if projectID == "" {
		log.Fatal("Error: --project is required")
	}
	if outputDir == "" {
		log.Fatal("Error: --output cannot be empty")
	}
	if pkgName == "" {
		log.Fatal("Error: --package cannot be empty")
	}

	// Create Polarion client
	client, err := polarion.New(url, token)
	if err != nil {
		log.Fatalf("Failed to create Polarion client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create generator configuration
	config := &codegen.Config{
		OutputDir: outputDir,
		Package:   pkgName,
		TypeID:    typeID,
		Refresh:   refresh,
	}

	// Create generator
	gen := codegen.NewGenerator(client, projectID, config)

	// Run generation
	if err := gen.Generate(ctx); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	fmt.Println("\n✓ Code generation completed successfully!")
}
