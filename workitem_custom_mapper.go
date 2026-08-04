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
//   - []string (for multi-value string and multi-enumeration fields)
//   - *int, *int8, *int16, *int32, *int64 (for integer fields)
//   - *uint, *uint8, *uint16, *uint32, *uint64 (for non-negative integer fields)
//   - *float32, *float64 (for float and currency fields)
//   - *bool (for boolean fields)
//   - *DateOnly (for date fields)
//   - *TimeOnly (for time fields)
//   - *DateTime (for datetime fields)
//   - *Duration (for duration fields)
//   - *TextContent (for text/html fields)
//   - *TableField (for table fields)
//   - *UserRef (for single user reference fields - stored in relationships)
//   - []UserRef (for multi-value user reference fields - stored in relationships)
//
// Named types with one of the kinds above are also supported (e.g. type Position int64,
// type Platform string, []Platform).
//
// Values are converted gracefully: a Polarion integer arrives as a JSON number and is
// accepted by any integer-width field, and a field whose Go type cannot hold the value
// (overflow, or a negative value in an unsigned field) reports an error rather than
// silently truncating.
//
// Note: UserRef fields are stored in Polarion's relationships section, not attributes.
// This function automatically handles loading them from the correct location.
//
// Example:
//
//	type Requirement struct {
//	    BusinessValue        *string           `json:"businessValue"`
//	    TargetRelease        *DateOnly         `json:"targetRelease"`
//	    ComplexityPoints     *float64          `json:"complexityPoints"`
//	    AffectedPlatforms    []string          `json:"affectedPlatforms"` // Multi-enumeration
//	    Purchaser            *polarion.UserRef `json:"purchaser"`   // Single user reference
//	    BoardMembers            []polarion.UserRef `json:"BoardMembers"`  // Multi-value user reference
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

		// Check if this is a UserRef field - these are loaded from relationships
		if isUserRefField(field) {
			if err := loadUserRefField(wi, field, fieldName); err != nil {
				return fmt.Errorf("failed to load user ref field %s: %w", fieldType.Name, err)
			}
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
// Note: UserRef fields are automatically saved to the relationships section, not attributes.
//
// Example:
//
//	req := &Requirement{
//	    BusinessValue: stringPtr("high"),
//	    ComplexityPoints: float64Ptr(13.0),
//	    Purchaser: polarion.NewUserRef("john.doe"),
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

		// Check if this is a UserRef field - these are saved to relationships
		if isUserRefField(field) {
			if err := saveUserRefField(wi, field, fieldName); err != nil {
				return fmt.Errorf("failed to save user ref field %s: %w", fieldType.Name, err)
			}
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
	// Multi-value fields (multi-enumeration, multi-value string) map to string slices
	if isStringSliceType(field.Type()) {
		if vals, ok := cf.GetStringSlice(fieldName); ok {
			field.Set(stringSliceToValue(field.Type(), vals))
		}
		return nil
	}

	if field.Kind() != reflect.Ptr {
		return fmt.Errorf("field must be a pointer type")
	}

	// Get the element type
	elemType := field.Type().Elem()

	// Pointer to a string slice (*[]string) is also accepted for multi-value fields
	if isStringSliceType(elemType) {
		if vals, ok := cf.GetStringSlice(fieldName); ok {
			ptr := reflect.New(elemType)
			ptr.Elem().Set(stringSliceToValue(elemType, vals))
			field.Set(ptr)
		}
		return nil
	}

	switch elemType.Kind() {
	case reflect.String:
		if val, ok := cf.GetString(fieldName); ok {
			ptr := reflect.New(elemType)
			ptr.Elem().SetString(val)
			field.Set(ptr)
		}
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, ok := cf.GetInt64(fieldName); ok {
			ptr := reflect.New(elemType)
			if ptr.Elem().OverflowInt(val) {
				return fmt.Errorf("value %d overflows %s", val, elemType)
			}
			ptr.Elem().SetInt(val)
			field.Set(ptr)
		}
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, ok := cf.GetInt64(fieldName); ok {
			if val < 0 {
				return fmt.Errorf("negative value %d cannot be stored in %s", val, elemType)
			}
			ptr := reflect.New(elemType)
			if ptr.Elem().OverflowUint(uint64(val)) {
				return fmt.Errorf("value %d overflows %s", val, elemType)
			}
			ptr.Elem().SetUint(uint64(val))
			field.Set(ptr)
		}
		return nil

	case reflect.Float32, reflect.Float64:
		if val, ok := cf.GetFloat(fieldName); ok {
			ptr := reflect.New(elemType)
			if ptr.Elem().OverflowFloat(val) {
				return fmt.Errorf("value %v overflows %s", val, elemType)
			}
			ptr.Elem().SetFloat(val)
			field.Set(ptr)
		}
		return nil

	case reflect.Bool:
		if val, ok := cf.GetBool(fieldName); ok {
			ptr := reflect.New(elemType)
			ptr.Elem().SetBool(val)
			field.Set(ptr)
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
	// Multi-value fields (multi-enumeration, multi-value string) map to string slices.
	// A nil slice removes the field; an empty non-nil slice stores an empty array,
	// which clears the field in Polarion.
	if isStringSliceType(field.Type()) {
		if field.IsNil() {
			cf.Delete(fieldName)
			return nil
		}
		cf.SetStringSlice(fieldName, valueToStringSlice(field))
		return nil
	}

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

	// Pointer to a string slice (*[]string) is also accepted for multi-value fields
	if isStringSliceType(elemType) {
		if fieldValue.IsNil() {
			cf.Delete(fieldName)
			return nil
		}
		cf.SetStringSlice(fieldName, valueToStringSlice(fieldValue))
		return nil
	}

	switch elemType.Kind() {
	case reflect.String:
		cf.Set(fieldName, fieldValue.String())
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Stored as int64 so any integer width round-trips without truncation
		cf.Set(fieldName, fieldValue.Int())
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cf.Set(fieldName, fieldValue.Uint())
		return nil

	case reflect.Float32, reflect.Float64:
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

// isStringSliceType reports whether t is a slice whose element kind is string,
// e.g. []string or []MyEnumType. This is how Polarion multi-value custom fields
// (multi-enumeration, multi-value string) are represented in Go.
// []UserRef is excluded because its element kind is struct, not string.
func isStringSliceType(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.String
}

// stringSliceToValue builds a reflect.Value of slice type t from a []string.
// Named string element types (e.g. type Platform string) are supported.
func stringSliceToValue(t reflect.Type, vals []string) reflect.Value {
	slice := reflect.MakeSlice(t, len(vals), len(vals))
	for i, val := range vals {
		slice.Index(i).SetString(val)
	}
	return slice
}

// valueToStringSlice converts a reflect.Value of a string-kinded slice to []string.
func valueToStringSlice(v reflect.Value) []string {
	vals := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		vals[i] = v.Index(i).String()
	}
	return vals
}

// isUserRefField checks if a reflect.Value represents a *UserRef or []UserRef field
func isUserRefField(field reflect.Value) bool {
	fieldType := field.Type()

	// Check for *UserRef
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		if elemType.Kind() == reflect.Struct && elemType.Name() == "UserRef" {
			return true
		}
	}

	// Check for []UserRef
	if fieldType.Kind() == reflect.Slice {
		elemType := fieldType.Elem()
		if elemType.Kind() == reflect.Struct && elemType.Name() == "UserRef" {
			return true
		}
	}

	return false
}

// loadUserRefField loads a UserRef field from the work item's relationships
// Handles both *UserRef (single) and []UserRef (multi-value) fields
func loadUserRefField(wi *WorkItem, field reflect.Value, fieldName string) error {
	if wi.Relationships == nil || wi.Relationships.CustomRelationships == nil {
		return nil
	}

	rel := wi.Relationships.CustomRelationships[fieldName]
	if rel == nil {
		return nil
	}

	fieldType := field.Type()

	// Handle *UserRef (single value)
	if fieldType.Kind() == reflect.Ptr {
		userRef := UserRefFromRelationship(rel)
		if userRef != nil {
			field.Set(reflect.ValueOf(userRef))
		}
		return nil
	}

	// Handle []UserRef (multi-value)
	if fieldType.Kind() == reflect.Slice {
		userRefs := UserRefsFromRelationship(rel)
		if len(userRefs) > 0 {
			field.Set(reflect.ValueOf(userRefs))
		}
		return nil
	}

	return nil
}

// saveUserRefField saves a UserRef field to the work item's relationships
// Handles both *UserRef (single) and []UserRef (multi-value) fields
func saveUserRefField(wi *WorkItem, field reflect.Value, fieldName string) error {
	// Ensure relationships structure exists
	if wi.Relationships == nil {
		wi.Relationships = &WorkItemRelationships{}
	}
	if wi.Relationships.CustomRelationships == nil {
		wi.Relationships.CustomRelationships = make(map[string]*Relationship)
	}

	fieldType := field.Type()

	// Handle *UserRef (single value)
	if fieldType.Kind() == reflect.Ptr {
		// If the field is nil, remove the relationship
		if field.IsNil() {
			delete(wi.Relationships.CustomRelationships, fieldName)
			return nil
		}

		// Get the UserRef value
		userRef := field.Interface().(*UserRef)
		if userRef == nil || userRef.ID == "" {
			delete(wi.Relationships.CustomRelationships, fieldName)
			return nil
		}

		// Set the relationship
		wi.Relationships.CustomRelationships[fieldName] = userRef.ToRelationship()
		return nil
	}

	// Handle []UserRef (multi-value)
	if fieldType.Kind() == reflect.Slice {
		// If the slice is nil or empty, remove the relationship
		if field.IsNil() || field.Len() == 0 {
			delete(wi.Relationships.CustomRelationships, fieldName)
			return nil
		}

		// Convert slice to relationship with array data
		userRefs := field.Interface().([]UserRef)
		wi.Relationships.CustomRelationships[fieldName] = UserRefsToRelationship(userRefs)
		return nil
	}

	return nil
}

// UserRefsFromRelationship extracts all UserRefs from a Relationship.
// This handles multi-value user reference fields that come as arrays.
func UserRefsFromRelationship(rel *Relationship) []UserRef {
	if rel == nil || rel.Data == nil {
		return nil
	}

	var refs []UserRef

	// Handle single data object: {"data": {"type": "users", "id": "john.doe"}}
	if data, ok := rel.Data.(map[string]interface{}); ok {
		if dataType, ok := data["type"].(string); ok && dataType == "users" {
			if id, ok := data["id"].(string); ok && id != "" {
				refs = append(refs, UserRef{ID: id})
			}
		}
		return refs
	}

	// Handle array of data: {"data": [{"type": "users", "id": "john.doe"}, {"type": "users", "id": "jane.doe"}]}
	if dataArray, ok := rel.Data.([]interface{}); ok {
		for _, item := range dataArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if dataType, ok := itemMap["type"].(string); ok && dataType == "users" {
					if id, ok := itemMap["id"].(string); ok && id != "" {
						refs = append(refs, UserRef{ID: id})
					}
				}
			}
		}
	}

	return refs
}

// UserRefsToRelationship converts a slice of UserRefs to a Relationship.
// This creates the proper array structure for multi-value user reference fields.
func UserRefsToRelationship(refs []UserRef) *Relationship {
	if len(refs) == 0 {
		return nil
	}

	data := make([]interface{}, 0, len(refs))
	for _, ref := range refs {
		if ref.ID != "" {
			data = append(data, map[string]interface{}{
				"type": "users",
				"id":   ref.ID,
			})
		}
	}

	if len(data) == 0 {
		return nil
	}

	return &Relationship{
		Data: data,
	}
}
