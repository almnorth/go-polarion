// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package codegen

import (
	"fmt"
	"strings"

	polarion "github.com/almnorth/go-polarion"
)

// FieldInfo represents discovered field information
type FieldInfo struct {
	// ID is the field identifier (e.g., "businessValue")
	ID string

	// Name is the human-readable field name (e.g., "Business Value")
	Name string

	// GoName is the Go struct field name (e.g., "BusinessValue")
	GoName string

	// GoType is the Go type for this field (e.g., "*string", "*polarion.DateOnly")
	GoType string

	// Kind is the Polarion field kind
	Kind polarion.FieldKind

	// Description is the field description for documentation
	Description string

	// EnumName is the enumeration name for enum fields
	EnumName string

	// EnumValues are the valid values for enum fields (if available)
	EnumValues []string

	// IsRequired indicates if the field is required
	IsRequired bool

	// TableColumns contains the column definitions for table fields
	TableColumns []TableColumn
}

// TableColumn represents a column definition for a table field
type TableColumn struct {
	// Key is the column key/identifier
	Key string

	// Name is the column name
	Name string

	// Title is the column title
	Title string
}

// Discoverer discovers custom fields from metadata
type Discoverer struct {
	metadata       *polarion.FieldsMetadata
	customFieldDef *polarion.CustomFieldsConfig
}

// NewDiscoverer creates a new field discoverer
func NewDiscoverer(metadata *polarion.FieldsMetadata, customFieldDef *polarion.CustomFieldsConfig) *Discoverer {
	return &Discoverer{
		metadata:       metadata,
		customFieldDef: customFieldDef,
	}
}

// DiscoverFields discovers all custom fields from the metadata
func (d *Discoverer) DiscoverFields() []FieldInfo {
	var fields []FieldInfo

	// Process attribute fields (primitive types)
	if d.metadata.Data.Attributes != nil {
		for fieldID, fieldMeta := range d.metadata.Data.Attributes {
			// Skip standard fields (we only want custom fields)
			if isStandardField(fieldID) {
				continue
			}

			field := d.convertField(fieldID, fieldMeta)
			fields = append(fields, field)
		}
	}

	// Note: We skip relationship fields for now as they require more complex handling
	// They could be added in a future enhancement

	return fields
}

// convertField converts a FieldMetadata to FieldInfo
func (d *Discoverer) convertField(fieldID string, meta polarion.FieldMetadata) FieldInfo {
	kind := polarion.FieldKind(meta.Type.Kind)

	// Check if this is a table field (structure with structureName "Table")
	if kind == polarion.FieldKindStructure && meta.Type.StructureName == "Table" {
		kind = polarion.FieldKindTable
	}

	field := FieldInfo{
		ID:          fieldID,
		Name:        meta.Label,
		GoName:      toGoFieldName(fieldID),
		Kind:        kind,
		Description: meta.Label,
		EnumName:    meta.Type.EnumName,
	}

	// Map Polarion field kind to Go type
	field.GoType = mapFieldKindToGoType(kind)

	// If this is a table field, extract column definitions from custom field config
	if kind == polarion.FieldKindTable && d.customFieldDef != nil {
		for _, customField := range d.customFieldDef.Attributes.Fields {
			if customField.ID == fieldID {
				// Extract table columns from parameters
				for _, param := range customField.Parameters {
					field.TableColumns = append(field.TableColumns, TableColumn{
						Key:   param.Key,
						Name:  param.Name,
						Title: param.Title,
					})
				}
				break
			}
		}
	}

	return field
}

// mapFieldKindToGoType maps a Polarion field kind to a Go type
func mapFieldKindToGoType(kind polarion.FieldKind) string {
	switch kind {
	case polarion.FieldKindString:
		return "*string"
	case polarion.FieldKindText, polarion.FieldKindTextHTML:
		return "*polarion.TextContent"
	case polarion.FieldKindInteger:
		return "*int"
	case polarion.FieldKindFloat:
		return "*float64"
	case polarion.FieldKindTime:
		return "*polarion.TimeOnly"
	case polarion.FieldKindDate:
		return "*polarion.DateOnly"
	case polarion.FieldKindDateTime:
		return "*polarion.DateTime"
	case polarion.FieldKindDuration:
		return "*polarion.Duration"
	case polarion.FieldKindBoolean:
		return "*bool"
	case polarion.FieldKindEnumeration:
		return "*string" // Enums are represented as strings with validation
	case polarion.FieldKindRelationship:
		return "*string" // Relationships are represented as IDs
	case polarion.FieldKindCode:
		return "*polarion.TextContent" // Code fields are text with syntax highlighting
	case polarion.FieldKindStructure:
		return "*string" // Structure fields contain structured data (JSON/XML)
	case polarion.FieldKindCurrency:
		return "*float64" // Currency fields are numeric values
	case polarion.FieldKindTable:
		return "*polarion.TableField" // Table fields with rows and columns
	default:
		// Log warning for unknown field types
		fmt.Printf("  ⚠ Warning: Unknown field kind '%s', defaulting to *string\n", kind)
		return "*string" // Default to string for unknown types
	}
}

// toGoFieldName converts a field ID to a Go field name
// Examples: "businessValue" -> "BusinessValue", "target_release" -> "TargetRelease"
func toGoFieldName(fieldID string) string {
	// Replace underscores and hyphens with spaces
	s := strings.ReplaceAll(fieldID, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Title case each word
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			// Handle camelCase: split on uppercase letters
			if i == 0 && strings.ToLower(word) != word {
				// First word might be camelCase
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			} else {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
	}

	// Join without spaces
	result := strings.Join(words, "")

	// Handle camelCase input (e.g., "businessValue" -> "BusinessValue")
	if len(result) > 0 && result[0] >= 'a' && result[0] <= 'z' {
		result = strings.ToUpper(result[:1]) + result[1:]
	}

	return result
}

// isStandardField checks if a field is a standard Polarion field
// Standard fields are already part of the WorkItem struct
func isStandardField(fieldID string) bool {
	standardFields := map[string]bool{
		"id":                true,
		"title":             true,
		"type":              true,
		"status":            true,
		"priority":          true,
		"severity":          true,
		"author":            true,
		"created":           true,
		"updated":           true,
		"assignee":          true,
		"description":       true,
		"dueDate":           true,
		"plannedStart":      true,
		"plannedEnd":        true,
		"timePoint":         true,
		"initialEstimate":   true,
		"remainingEstimate": true,
		"timeSpent":         true,
		"resolution":        true,
		"approvalStatus":    true,
		"outlineNumber":     true,
		"project":           true,
	}

	return standardFields[fieldID]
}
