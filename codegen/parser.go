// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package codegen

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParsedFile represents a parsed generated file
type ParsedFile struct {
	// FilePath is the path to the file
	FilePath string

	// CustomCode contains custom code sections (outside generation markers)
	CustomCode map[string]string

	// GeneratedSections contains the generated code sections
	GeneratedSections map[string]string
}

// Parser parses existing generated files to extract custom code
type Parser struct{}

// NewParser creates a new file parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an existing generated file
func (p *Parser) Parse(filePath string) (*ParsedFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	parsed := &ParsedFile{
		FilePath:          filePath,
		CustomCode:        make(map[string]string),
		GeneratedSections: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	var currentSection strings.Builder
	var inGeneratedSection bool
	var sectionName string

	for scanner.Scan() {
		line := scanner.Text()

		// Check for generation markers
		if strings.Contains(line, "GENERATED_FIELDS_START") {
			inGeneratedSection = true
			sectionName = "fields"
			currentSection.Reset()
			continue
		} else if strings.Contains(line, "GENERATED_FIELDS_END") {
			if inGeneratedSection && sectionName == "fields" {
				parsed.GeneratedSections[sectionName] = currentSection.String()
			}
			inGeneratedSection = false
			sectionName = ""
			currentSection.Reset()
			continue
		} else if strings.Contains(line, "GENERATED_METHODS_START") {
			inGeneratedSection = true
			sectionName = "methods"
			currentSection.Reset()
			continue
		} else if strings.Contains(line, "GENERATED_METHODS_END") {
			if inGeneratedSection && sectionName == "methods" {
				parsed.GeneratedSections[sectionName] = currentSection.String()
			}
			inGeneratedSection = false
			sectionName = ""
			currentSection.Reset()
			continue
		}

		// Accumulate lines
		if inGeneratedSection {
			currentSection.WriteString(line)
			currentSection.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parsed, nil
}

// ExtractCustomCode extracts custom code sections from the file
// This is a placeholder for future enhancement to preserve custom code
func (p *Parser) ExtractCustomCode(filePath string) (map[string]string, error) {
	// For now, return empty map
	// In a full implementation, this would parse the file and extract
	// code sections that are outside the generation markers
	return make(map[string]string), nil
}
