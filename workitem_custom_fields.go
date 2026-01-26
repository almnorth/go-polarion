// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "strconv"

// CustomFields provides type-safe access to custom fields in WorkItemAttributes.
// It wraps the map[string]interface{} to provide convenient accessor methods
// that handle Polarion's data quirks (missing keys, type conversions).
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if priority, ok := cf.GetString("priority"); ok {
//	    fmt.Printf("Priority: %s\n", priority)
//	}
//	if dueDate, ok := cf.GetDateOnly("dueDate"); ok {
//	    fmt.Printf("Due: %s\n", dueDate.String())
//	}
type CustomFields map[string]interface{}

// GetString safely retrieves a string custom field (kind: string, enumeration).
// Returns the value and true if the field exists and is a string, otherwise returns empty string and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if title, ok := cf.GetString("customTitle"); ok {
//	    fmt.Printf("Custom Title: %s\n", title)
//	}
func (cf CustomFields) GetString(key string) (string, bool) {
	val, exists := cf[key]
	if !exists {
		return "", false
	}

	// Handle nil value
	if val == nil {
		return "", false
	}

	// Try direct string conversion
	if str, ok := val.(string); ok {
		return str, true
	}

	return "", false
}

// GetInt safely retrieves an integer custom field (kind: integer).
// Handles both int and float64 from JSON unmarshaling.
// Returns the value and true if the field exists and can be converted to int, otherwise returns 0 and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if count, ok := cf.GetInt("itemCount"); ok {
//	    fmt.Printf("Item Count: %d\n", count)
//	}
func (cf CustomFields) GetInt(key string) (int, bool) {
	val, exists := cf[key]
	if !exists {
		return 0, false
	}

	// Handle nil value
	if val == nil {
		return 0, false
	}

	// Handle different numeric types from JSON unmarshaling
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

// GetFloat safely retrieves a float custom field (kind: float, currency).
// Handles float64, int, and string (for currency fields) from JSON unmarshaling.
// Returns the value and true if the field exists and can be converted to float64, otherwise returns 0.0 and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if score, ok := cf.GetFloat("qualityScore"); ok {
//	    fmt.Printf("Quality Score: %.2f\n", score)
//	}
func (cf CustomFields) GetFloat(key string) (float64, bool) {
	val, exists := cf[key]
	if !exists {
		return 0.0, false
	}

	// Handle nil value
	if val == nil {
		return 0.0, false
	}

	// Handle different numeric types from JSON unmarshaling
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		// Handle currency fields which come as strings from Polarion API
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
		return 0.0, false
	default:
		return 0.0, false
	}
}

// GetBool safely retrieves a boolean custom field (kind: boolean).
// Returns the value and true if the field exists and is a bool, otherwise returns false and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if isActive, ok := cf.GetBool("isActive"); ok {
//	    fmt.Printf("Is Active: %t\n", isActive)
//	}
func (cf CustomFields) GetBool(key string) (bool, bool) {
	val, exists := cf[key]
	if !exists {
		return false, false
	}

	// Handle nil value
	if val == nil {
		return false, false
	}

	// Try direct bool conversion
	if b, ok := val.(bool); ok {
		return b, true
	}

	return false, false
}

// GetText safely retrieves a text custom field (kind: text, text/html).
// Returns TextContent with type and value.
// Handles both TextContent objects and map[string]interface{} from JSON unmarshaling.
// Returns the value and true if the field exists, otherwise returns nil and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if description, ok := cf.GetText("customDescription"); ok {
//	    fmt.Printf("Description Type: %s\n", description.Type)
//	    fmt.Printf("Description Value: %s\n", description.Value)
//	}
func (cf CustomFields) GetText(key string) (*TextContent, bool) {
	val, exists := cf[key]
	if !exists {
		return nil, false
	}

	// Handle nil value
	if val == nil {
		return nil, false
	}

	// Handle TextContent object directly
	if tc, ok := val.(*TextContent); ok {
		return tc, true
	}

	// Handle non-pointer TextContent
	if tc, ok := val.(TextContent); ok {
		return &tc, true
	}

	// Handle map from JSON unmarshaling
	if m, ok := val.(map[string]interface{}); ok {
		tc := &TextContent{}
		if t, ok := m["type"].(string); ok {
			tc.Type = t
		}
		if v, ok := m["value"].(string); ok {
			tc.Value = v
		}
		return tc, true
	}

	return nil, false
}

// GetTimeOnly safely retrieves a time custom field (kind: time).
// Parses the string value in HH:MM:SS format.
// Returns the value and true if the field exists and can be parsed, otherwise returns zero value and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if startTime, ok := cf.GetTimeOnly("startTime"); ok {
//	    fmt.Printf("Start Time: %s\n", startTime.String())
//	}
func (cf CustomFields) GetTimeOnly(key string) (TimeOnly, bool) {
	str, ok := cf.GetString(key)
	if !ok {
		return TimeOnly{}, false
	}

	t, err := ParseTimeOnly(str)
	if err != nil {
		return TimeOnly{}, false
	}

	return t, true
}

// GetDateOnly safely retrieves a date custom field (kind: date).
// Parses the string value in YYYY-MM-DD format.
// Returns the value and true if the field exists and can be parsed, otherwise returns zero value and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if dueDate, ok := cf.GetDateOnly("dueDate"); ok {
//	    fmt.Printf("Due Date: %s\n", dueDate.String())
//	}
func (cf CustomFields) GetDateOnly(key string) (DateOnly, bool) {
	str, ok := cf.GetString(key)
	if !ok {
		return DateOnly{}, false
	}

	d, err := ParseDateOnly(str)
	if err != nil {
		return DateOnly{}, false
	}

	return d, true
}

// GetDateTime safely retrieves a datetime custom field (kind: date-time).
// Parses the string value in ISO 8601 format.
// Returns the value and true if the field exists and can be parsed, otherwise returns zero value and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if createdAt, ok := cf.GetDateTime("customCreatedAt"); ok {
//	    fmt.Printf("Created At: %s\n", createdAt.String())
//	}
func (cf CustomFields) GetDateTime(key string) (DateTime, bool) {
	str, ok := cf.GetString(key)
	if !ok {
		return DateTime{}, false
	}

	dt, err := ParseDateTime(str)
	if err != nil {
		return DateTime{}, false
	}

	return dt, true
}

// GetDuration safely retrieves a duration custom field (kind: duration).
// Parses the string value in Polarion format (e.g., "1h", "2d 3h").
// Returns the value and true if the field exists and can be parsed, otherwise returns zero value and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if estimate, ok := cf.GetDuration("timeEstimate"); ok {
//	    fmt.Printf("Time Estimate: %s\n", estimate.String())
//	}
func (cf CustomFields) GetDuration(key string) (Duration, bool) {
	str, ok := cf.GetString(key)
	if !ok {
		return Duration{}, false
	}

	d, err := ParseDuration(str)
	if err != nil {
		return Duration{}, false
	}

	return d, true
}

// GetTable safely retrieves a table custom field (kind: table).
// Handles map[string]interface{} from JSON unmarshaling and converts it to TableField.
// Returns the value and true if the field exists and can be converted, otherwise returns nil and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if table, ok := cf.GetTable("dataTable"); ok {
//	    headers := table.GetHeaders()
//	    for i, row := range table.GetAllRowsAsMap() {
//	        fmt.Printf("Row %d: %v\n", i, row)
//	    }
//	}
func (cf CustomFields) GetTable(key string) (*TableField, bool) {
	val, exists := cf[key]
	if !exists {
		return nil, false
	}

	// Handle nil value
	if val == nil {
		return nil, false
	}

	// Handle TableField object directly
	if table, ok := val.(*TableField); ok {
		return table, true
	}

	// Handle non-pointer TableField
	if table, ok := val.(TableField); ok {
		return &table, true
	}

	// Handle map from JSON unmarshaling
	if m, ok := val.(map[string]interface{}); ok {
		table := &TableField{}

		// Extract keys
		if keysRaw, ok := m["keys"].([]interface{}); ok {
			table.Keys = make([]string, len(keysRaw))
			for i, k := range keysRaw {
				if str, ok := k.(string); ok {
					table.Keys[i] = str
				}
			}
		}

		// Extract rows
		if rowsRaw, ok := m["rows"].([]interface{}); ok {
			table.Rows = make([]TableRow, len(rowsRaw))
			for i, rowRaw := range rowsRaw {
				if rowMap, ok := rowRaw.(map[string]interface{}); ok {
					if valuesRaw, ok := rowMap["values"].([]interface{}); ok {
						row := TableRow{
							Values: make([]TextContent, len(valuesRaw)),
						}
						for j, cellRaw := range valuesRaw {
							if cellMap, ok := cellRaw.(map[string]interface{}); ok {
								cell := TextContent{}
								if t, ok := cellMap["type"].(string); ok {
									cell.Type = t
								}
								if v, ok := cellMap["value"].(string); ok {
									cell.Value = v
								}
								row.Values[j] = cell
							}
						}
						table.Rows[i] = row
					}
				}
			}
		}

		return table, true
	}

	return nil, false
}

// GetEnum safely retrieves an enum custom field (kind: enumeration).
// This is an alias for GetString but makes the intent clearer for enumeration fields.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if status, ok := cf.GetEnum("customStatus"); ok {
//	    fmt.Printf("Custom Status: %s\n", status)
//	}
func (cf CustomFields) GetEnum(key string) (string, bool) {
	return cf.GetString(key)
}

// Set sets a custom field value.
// The value can be any type that is JSON-serializable.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	cf.Set("priority", "high")
//	cf.Set("itemCount", 42)
//	cf.Set("isActive", true)
func (cf CustomFields) Set(key string, value interface{}) {
	cf[key] = value
}

// Has checks if a custom field exists (key is present in the map).
// Returns true if the key exists, even if the value is nil.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if cf.Has("priority") {
//	    fmt.Println("Priority field exists")
//	}
func (cf CustomFields) Has(key string) bool {
	_, exists := cf[key]
	return exists
}

// Delete removes a custom field from the map.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	cf.Delete("obsoleteField")
func (cf CustomFields) Delete(key string) {
	delete(cf, key)
}
