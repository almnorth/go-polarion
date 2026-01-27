// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"fmt"
	"reflect"
	"strings"
)

// LoadCustomFields automatically loads custom fields from a WorkItem into a struct
// using reflection and JSON struct tags. Fields should be tagged with `json:"fieldName"`
// to specify the custom field name in Polarion (the same field ID used in the API).
//
// Supported field types:
//   - *string (for string and enum fields)
//   - *int (for integer fields)
//   - *float64 (for float fields)
//   - *bool (for boolean fields)
//   - *DateOnly (for date fields)
//   - *TimeOnly (for time fields)
//   - *DateTime (for datetime fields)
//   - *Duration (for duration fields)
//   - *TextContent (for text/html fields)
//   - *TableField (for table fields)
//
// Example:
//
//	type Requirement struct {
//	    BusinessValue    *string   `json:"businessValue"`
//	    TargetRelease    *DateOnly `json:"targetRelease"`
//	    ComplexityPoints *float64  `json:"complexityPoints"`
//	}
//
//	req := &Requirement{}
//	err := polarion.LoadCustomFields(wi, req)
func LoadCustomFields(wi *WorkItem, target interface{}) error {
	if wi == nil {
		return fmt.Errorf("work item is nil")
	}
	if wi.Attributes == nil {
		return fmt.Errorf("work item attributes are nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	targetElem := targetValue.Elem()
	if targetElem.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	cf := CustomFields(wi.Attributes.CustomFields)
	targetType := targetElem.Type()

	for i := 0; i < targetElem.NumField(); i++ {
		field := targetElem.Field(i)
		fieldType := targetType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the json tag
		tag := fieldType.Tag.Get("json")
		if tag == "" {
			continue
		}

		// Parse tag (support "fieldName" or "fieldName,omitempty")
		fieldName := strings.Split(tag, ",")[0]
		if fieldName == "" || fieldName == "-" {
			continue
		}

		// Load the field based on its type
		if err := loadField(cf, field, fieldName); err != nil {
			return fmt.Errorf("failed to load field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// SaveCustomFields automatically saves struct fields to a WorkItem's custom fields
// using reflection and JSON struct tags. Fields should be tagged with `json:"fieldName"`
// to specify the custom field name in Polarion (the same field ID used in the API).
//
// Example:
//
//	req := &Requirement{
//	    BusinessValue: stringPtr("high"),
//	    ComplexityPoints: float64Ptr(13.0),
//	}
//	err := polarion.SaveCustomFields(wi, req)
func SaveCustomFields(wi *WorkItem, source interface{}) error {
	if wi == nil {
		return fmt.Errorf("work item is nil")
	}
	if wi.Attributes == nil {
		return fmt.Errorf("work item attributes are nil")
	}

	sourceValue := reflect.ValueOf(source)
	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}
	if sourceValue.Kind() != reflect.Struct {
		return fmt.Errorf("source must be a struct or pointer to struct")
	}

	if wi.Attributes.CustomFields == nil {
		wi.Attributes.CustomFields = make(map[string]interface{})
	}

	cf := CustomFields(wi.Attributes.CustomFields)
	sourceType := sourceValue.Type()

	for i := 0; i < sourceValue.NumField(); i++ {
		field := sourceValue.Field(i)
		fieldType := sourceType.Field(i)

		// Get the json tag
		tag := fieldType.Tag.Get("json")
		if tag == "" {
			continue
		}

		// Parse tag
		fieldName := strings.Split(tag, ",")[0]
		if fieldName == "" || fieldName == "-" {
			continue
		}

		// Save the field based on its type
		if err := saveField(cf, field, fieldName); err != nil {
			return fmt.Errorf("failed to save field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// loadField loads a single field from custom fields based on its type
func loadField(cf CustomFields, field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.Ptr {
		return fmt.Errorf("field must be a pointer type")
	}

	// Get the element type
	elemType := field.Type().Elem()

	switch elemType.Kind() {
	case reflect.String:
		if val, ok := cf.GetString(fieldName); ok {
			field.Set(reflect.ValueOf(&val))
		}
		return nil

	case reflect.Int:
		if val, ok := cf.GetInt(fieldName); ok {
			field.Set(reflect.ValueOf(&val))
		}
		return nil

	case reflect.Float64:
		if val, ok := cf.GetFloat(fieldName); ok {
			field.Set(reflect.ValueOf(&val))
		}
		return nil

	case reflect.Bool:
		if val, ok := cf.GetBool(fieldName); ok {
			field.Set(reflect.ValueOf(&val))
		}
		return nil

	case reflect.Struct:
		// Handle custom types
		switch elemType.Name() {
		case "DateOnly":
			if val, ok := cf.GetDateOnly(fieldName); ok {
				field.Set(reflect.ValueOf(&val))
			}
			return nil

		case "TimeOnly":
			if val, ok := cf.GetTimeOnly(fieldName); ok {
				field.Set(reflect.ValueOf(&val))
			}
			return nil

		case "DateTime":
			if val, ok := cf.GetDateTime(fieldName); ok {
				field.Set(reflect.ValueOf(&val))
			}
			return nil

		case "Duration":
			if val, ok := cf.GetDuration(fieldName); ok {
				field.Set(reflect.ValueOf(&val))
			}
			return nil

		case "TextContent":
			if val, ok := cf.GetText(fieldName); ok {
				field.Set(reflect.ValueOf(val))
			}
			return nil

		case "TableField":
			if val, ok := cf.GetTable(fieldName); ok {
				field.Set(reflect.ValueOf(val))
			}
			return nil

		default:
			return fmt.Errorf("unsupported struct type: %s", elemType.Name())
		}

	default:
		return fmt.Errorf("unsupported field type: %s", elemType.Kind())
	}
}

// saveField saves a single field to custom fields based on its type
func saveField(cf CustomFields, field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.Ptr {
		return fmt.Errorf("field must be a pointer type")
	}

	// If the field is nil, delete it from custom fields
	if field.IsNil() {
		cf.Delete(fieldName)
		return nil
	}

	// Dereference the pointer
	fieldValue := field.Elem()
	elemType := field.Type().Elem()

	switch elemType.Kind() {
	case reflect.String:
		cf.Set(fieldName, fieldValue.String())
		return nil

	case reflect.Int:
		cf.Set(fieldName, int(fieldValue.Int()))
		return nil

	case reflect.Float64:
		cf.Set(fieldName, fieldValue.Float())
		return nil

	case reflect.Bool:
		cf.Set(fieldName, fieldValue.Bool())
		return nil

	case reflect.Struct:
		// Handle custom types
		switch elemType.Name() {
		case "DateOnly":
			dateOnly := fieldValue.Interface().(DateOnly)
			cf.Set(fieldName, dateOnly.String())
			return nil

		case "TimeOnly":
			timeOnly := fieldValue.Interface().(TimeOnly)
			cf.Set(fieldName, timeOnly.String())
			return nil

		case "DateTime":
			dateTime := fieldValue.Interface().(DateTime)
			cf.Set(fieldName, dateTime.String())
			return nil

		case "Duration":
			duration := fieldValue.Interface().(Duration)
			cf.Set(fieldName, duration.String())
			return nil

		case "TextContent":
			textContent := fieldValue.Interface().(TextContent)
			cf.Set(fieldName, &textContent)
			return nil

		case "TableField":
			tableField := fieldValue.Interface().(TableField)
			cf.Set(fieldName, &tableField)
			return nil

		default:
			return fmt.Errorf("unsupported struct type: %s", elemType.Name())
		}

	default:
		return fmt.Errorf("unsupported field type: %s", elemType.Kind())
	}
}
