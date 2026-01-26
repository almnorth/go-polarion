// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FieldKind represents the Polarion field type kind.
// These constants map directly to Polarion's field type system and are used
// to determine how field values should be parsed, validated, and serialized.
type FieldKind string

const (
	// FieldKindString represents a simple string field
	FieldKindString FieldKind = "string"

	// FieldKindText represents a multi-line text field
	FieldKindText FieldKind = "text"

	// FieldKindTextHTML represents an HTML-formatted text field
	FieldKindTextHTML FieldKind = "text/html"

	// FieldKindInteger represents an integer number field
	FieldKindInteger FieldKind = "integer"

	// FieldKindFloat represents a floating-point number field
	FieldKindFloat FieldKind = "float"

	// FieldKindTime represents a time-only field (HH:MM:SS format)
	FieldKindTime FieldKind = "time"

	// FieldKindDate represents a date-only field (YYYY-MM-DD format)
	FieldKindDate FieldKind = "date"

	// FieldKindDateTime represents a date and time field (ISO 8601 format)
	FieldKindDateTime FieldKind = "date-time"

	// FieldKindDuration represents a time duration field (e.g., "1h", "2d 3h")
	FieldKindDuration FieldKind = "duration"

	// FieldKindBoolean represents a boolean field (true/false)
	FieldKindBoolean FieldKind = "boolean"

	// FieldKindEnumeration represents an enumeration field with predefined values
	FieldKindEnumeration FieldKind = "enumeration"

	// FieldKindRelationship represents a relationship to another work item or resource
	FieldKindRelationship FieldKind = "relationship"

	// FieldKindCode represents a code field with syntax highlighting
	FieldKindCode FieldKind = "code"

	// FieldKindStructure represents a structured data field (JSON/XML)
	FieldKindStructure FieldKind = "structure"

	// FieldKindCurrency represents a currency/monetary value field
	FieldKindCurrency FieldKind = "currency"

	// FieldKindTable !! NOT IN THE POLARION API (why?) !! WE JUST GUESS BASED ON THE STRUCURE IN GO. represents a table field with rows and columns
	FieldKindTable FieldKind = "table"
)

// TimeOnly represents a Polarion time field (HH:MM:SS format).
// Used for time-only fields without date information.
//
// Example:
//
//	t := polarion.NewTimeOnly(21, 12, 0)
//	fmt.Println(t.String()) // Output: 21:12:00
//
// JSON marshaling:
//
//	{"timeField": "21:12:00"}
type TimeOnly struct {
	Hour   int
	Minute int
	Second int
}

// NewTimeOnly creates a new TimeOnly instance with validation.
// Returns an error if the time values are out of valid ranges:
// - Hour: 0-23
// - Minute: 0-59
// - Second: 0-59
func NewTimeOnly(hour, minute, second int) (TimeOnly, error) {
	if hour < 0 || hour > 23 {
		return TimeOnly{}, fmt.Errorf("invalid hour: %d (must be 0-23)", hour)
	}
	if minute < 0 || minute > 59 {
		return TimeOnly{}, fmt.Errorf("invalid minute: %d (must be 0-59)", minute)
	}
	if second < 0 || second > 59 {
		return TimeOnly{}, fmt.Errorf("invalid second: %d (must be 0-59)", second)
	}
	return TimeOnly{
		Hour:   hour,
		Minute: minute,
		Second: second,
	}, nil
}

// ParseTimeOnly parses a time string in HH:MM:SS format.
// Returns an error if the format is invalid or values are out of range.
//
// Example:
//
//	t, err := polarion.ParseTimeOnly("21:12:00")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseTimeOnly(s string) (TimeOnly, error) {
	if s == "" {
		return TimeOnly{}, fmt.Errorf("empty time string")
	}

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return TimeOnly{}, fmt.Errorf("invalid time format: %s (expected HH:MM:SS)", s)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return TimeOnly{}, fmt.Errorf("invalid hour in time: %s", s)
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return TimeOnly{}, fmt.Errorf("invalid minute in time: %s", s)
	}

	second, err := strconv.Atoi(parts[2])
	if err != nil {
		return TimeOnly{}, fmt.Errorf("invalid second in time: %s", s)
	}

	return NewTimeOnly(hour, minute, second)
}

// String returns the time in HH:MM:SS format.
func (t TimeOnly) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

// MarshalJSON implements json.Marshaler for TimeOnly.
func (t TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements json.Unmarshaler for TimeOnly.
func (t *TimeOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseTimeOnly(s)
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

// DateOnly represents a Polarion date field (YYYY-MM-DD format).
// Used for date-only fields without time information.
// The time component is normalized to midnight UTC.
//
// Example:
//
//	d := polarion.NewDateOnly(time.Now())
//	fmt.Println(d.String()) // Output: 2026-01-26
//
// JSON marshaling:
//
//	{"dateField": "2026-01-26"}
type DateOnly struct {
	time.Time
}

// NewDateOnly creates a new DateOnly instance, normalizing the time to midnight UTC.
func NewDateOnly(t time.Time) DateOnly {
	year, month, day := t.Date()
	return DateOnly{
		Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
}

// ParseDateOnly parses a date string in YYYY-MM-DD format.
// Returns an error if the format is invalid.
//
// Example:
//
//	d, err := polarion.ParseDateOnly("2026-01-26")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseDateOnly(s string) (DateOnly, error) {
	if s == "" {
		return DateOnly{}, fmt.Errorf("empty date string")
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return DateOnly{}, fmt.Errorf("invalid date format: %w", err)
	}

	return NewDateOnly(t), nil
}

// String returns the date in YYYY-MM-DD format.
func (d DateOnly) String() string {
	return d.Time.Format("2006-01-02")
}

// MarshalJSON implements json.Marshaler for DateOnly.
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler for DateOnly.
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseDateOnly(s)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// DateTime represents a Polarion date-time field (ISO 8601 format).
// Used for fields that require both date and time information.
//
// Example:
//
//	dt := polarion.NewDateTime(time.Now())
//	fmt.Println(dt.String()) // Output: 2026-01-26T19:23:30Z
//
// JSON marshaling:
//
//	{"dateTimeField": "2026-01-26T19:23:30Z"}
type DateTime struct {
	time.Time
}

// NewDateTime creates a new DateTime instance.
func NewDateTime(t time.Time) DateTime {
	return DateTime{Time: t}
}

// ParseDateTime parses a date-time string in ISO 8601 format (RFC3339).
// Returns an error if the format is invalid.
//
// Example:
//
//	dt, err := polarion.ParseDateTime("2026-01-26T19:23:30Z")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseDateTime(s string) (DateTime, error) {
	if s == "" {
		return DateTime{}, fmt.Errorf("empty datetime string")
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return DateTime{}, fmt.Errorf("invalid datetime format: %w", err)
	}

	return NewDateTime(t), nil
}

// String returns the date-time in RFC3339 format (ISO 8601).
func (dt DateTime) String() string {
	return dt.Time.Format(time.RFC3339)
}

// MarshalJSON implements json.Marshaler for DateTime.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dt.String())
}

// UnmarshalJSON implements json.Unmarshaler for DateTime.
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseDateTime(s)
	if err != nil {
		return err
	}

	*dt = parsed
	return nil
}

// Duration represents a Polarion duration field.
// Supports Polarion's duration format with units: d (days), h (hours), m (minutes), s (seconds).
//
// Example:
//
//	d := polarion.NewDuration(2*24*time.Hour + 3*time.Hour)
//	fmt.Println(d.String()) // Output: 2d 3h
//
// JSON marshaling:
//
//	{"durationField": "2d 3h"}
type Duration struct {
	time.Duration
}

// NewDuration creates a new Duration instance.
func NewDuration(d time.Duration) Duration {
	return Duration{Duration: d}
}

// ParseDuration parses a duration string in Polarion format.
// Supports units: d (days), h (hours), m (minutes), s (seconds).
// Multiple units can be combined with spaces (e.g., "2d 3h 30m").
//
// Example:
//
//	d, err := polarion.ParseDuration("2d 3h 30m")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseDuration(s string) (Duration, error) {
	if s == "" {
		return Duration{}, fmt.Errorf("empty duration string")
	}

	// Regular expression to match duration components
	// Matches patterns like "2d", "3h", "30m", "45s"
	re := regexp.MustCompile(`(\d+)\s*([dhms])`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) == 0 {
		return Duration{}, fmt.Errorf("invalid duration format: %s", s)
	}

	var total time.Duration

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		value, err := strconv.Atoi(match[1])
		if err != nil {
			return Duration{}, fmt.Errorf("invalid duration value: %s", match[1])
		}

		unit := match[2]
		switch unit {
		case "d":
			total += time.Duration(value) * 24 * time.Hour
		case "h":
			total += time.Duration(value) * time.Hour
		case "m":
			total += time.Duration(value) * time.Minute
		case "s":
			total += time.Duration(value) * time.Second
		default:
			return Duration{}, fmt.Errorf("unknown duration unit: %s", unit)
		}
	}

	return NewDuration(total), nil
}

// String returns the duration in Polarion format.
// Formats as a combination of days, hours, minutes, and seconds.
// Only non-zero components are included.
//
// Examples:
//   - 2 days, 3 hours: "2d 3h"
//   - 30 minutes: "30m"
//   - 1 day, 30 minutes: "1d 30m"
func (d Duration) String() string {
	if d.Duration == 0 {
		return "0s"
	}

	total := d.Duration
	var parts []string

	// Days
	days := total / (24 * time.Hour)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
		total -= days * 24 * time.Hour
	}

	// Hours
	hours := total / time.Hour
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
		total -= hours * time.Hour
	}

	// Minutes
	minutes := total / time.Minute
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		total -= minutes * time.Minute
	}

	// Seconds
	seconds := total / time.Second
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, " ")
}

// MarshalJSON implements json.Marshaler for Duration.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler for Duration.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseDuration(s)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// TableField represents a Polarion table field.
// Tables have column keys and rows of cells, where each cell contains typed content.
//
// Example structure:
//
//	{
//	  "keys": ["Column1", "Column2"],
//	  "rows": [
//	    {"values": [{"type": "text/html", "value": "Cell 1"}, {"type": "text/html", "value": "Cell 2"}]},
//	    {"values": [{"type": "text/html", "value": "Cell 3"}, {"type": "text/html", "value": "Cell 4"}]}
//	  ]
//	}
type TableField struct {
	// Keys are the column headers
	Keys []string `json:"keys,omitempty"`

	// Rows contains the table data
	Rows []TableRow `json:"rows,omitempty"`
}

// TableRow represents a single row in a table field.
type TableRow struct {
	// Values contains the cell values for this row
	Values []TextContent `json:"values,omitempty"`
}

// GetCell returns the cell value at the specified row and column index.
// Returns an error if the indices are out of bounds.
func (t *TableField) GetCell(row, col int) (*TextContent, error) {
	if row < 0 || row >= len(t.Rows) {
		return nil, fmt.Errorf("row index %d out of bounds (table has %d rows)", row, len(t.Rows))
	}
	if col < 0 || col >= len(t.Rows[row].Values) {
		return nil, fmt.Errorf("column index %d out of bounds (row %d has %d columns)", col, row, len(t.Rows[row].Values))
	}
	return &t.Rows[row].Values[col], nil
}

// GetCellByKey returns the cell value at the specified row and column key.
// Returns an error if the row index is out of bounds or the column key is not found.
func (t *TableField) GetCellByKey(row int, key string) (*TextContent, error) {
	if row < 0 || row >= len(t.Rows) {
		return nil, fmt.Errorf("row index %d out of bounds (table has %d rows)", row, len(t.Rows))
	}

	// Find column index by key
	colIndex := -1
	for i, k := range t.Keys {
		if k == key {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return nil, fmt.Errorf("column key %q not found in table", key)
	}

	return t.GetCell(row, colIndex)
}

// RowCount returns the number of rows in the table.
func (t *TableField) RowCount() int {
	return len(t.Rows)
}

// ColumnCount returns the number of columns in the table.
func (t *TableField) ColumnCount() int {
	return len(t.Keys)
}

// GetHeaders returns the column headers (keys) of the table.
func (t *TableField) GetHeaders() []string {
	return t.Keys
}

// GetRows returns all rows in the table.
func (t *TableField) GetRows() []TableRow {
	return t.Rows
}

// GetRow returns all cells in the specified row.
// Returns an error if the row index is out of bounds.
func (t *TableField) GetRow(row int) ([]TextContent, error) {
	if row < 0 || row >= len(t.Rows) {
		return nil, fmt.Errorf("row index %d out of bounds (table has %d rows)", row, len(t.Rows))
	}
	return t.Rows[row].Values, nil
}

// GetRowAsMap returns a row as a map with column names as keys.
// This provides convenient access to cells by column name.
// Returns an error if the row index is out of bounds.
//
// Example:
//
//	rowMap, err := table.GetRowAsMap(0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	firstNameCell := rowMap["firstName"]
//	fmt.Printf("First Name: %s\n", firstNameCell.Value)
func (t *TableField) GetRowAsMap(row int) (map[string]TextContent, error) {
	if row < 0 || row >= len(t.Rows) {
		return nil, fmt.Errorf("row index %d out of bounds (table has %d rows)", row, len(t.Rows))
	}

	result := make(map[string]TextContent)
	for i, key := range t.Keys {
		if i < len(t.Rows[row].Values) {
			result[key] = t.Rows[row].Values[i]
		}
	}
	return result, nil
}

// GetAllRowsAsMap returns all rows as a slice of maps with column names as keys.
// This provides convenient access to all table data with named columns.
//
// Example:
//
//	rows := table.GetAllRowsAsMap()
//	for i, row := range rows {
//	    fmt.Printf("Row %d - First Name: %s, Last Name: %s\n",
//	        i, row["firstName"].Value, row["lastName"].Value)
//	}
func (t *TableField) GetAllRowsAsMap() []map[string]TextContent {
	result := make([]map[string]TextContent, len(t.Rows))
	for i := range t.Rows {
		rowMap, _ := t.GetRowAsMap(i) // Error already checked in loop bounds
		result[i] = rowMap
	}
	return result
}

// GetColumn returns all cells in the specified column by index.
// Returns an error if the column index is out of bounds.
func (t *TableField) GetColumn(col int) ([]TextContent, error) {
	if col < 0 || col >= len(t.Keys) {
		return nil, fmt.Errorf("column index %d out of bounds (table has %d columns)", col, len(t.Keys))
	}

	result := make([]TextContent, len(t.Rows))
	for i, row := range t.Rows {
		if col < len(row.Values) {
			result[i] = row.Values[col]
		}
	}
	return result, nil
}

// GetColumnByKey returns all cells in the specified column by key.
// Returns an error if the column key is not found.
func (t *TableField) GetColumnByKey(key string) ([]TextContent, error) {
	// Find column index by key
	colIndex := -1
	for i, k := range t.Keys {
		if k == key {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return nil, fmt.Errorf("column key %q not found in table", key)
	}

	return t.GetColumn(colIndex)
}

// AddRow adds a new row to the table.
// The number of values must match the number of columns.
func (t *TableField) AddRow(values []TextContent) error {
	if len(values) != len(t.Keys) {
		return fmt.Errorf("row has %d values but table has %d columns", len(values), len(t.Keys))
	}
	t.Rows = append(t.Rows, TableRow{Values: values})
	return nil
}

// SetCell sets the value of a cell at the specified row and column index.
// Returns an error if the indices are out of bounds.
func (t *TableField) SetCell(row, col int, value TextContent) error {
	if row < 0 || row >= len(t.Rows) {
		return fmt.Errorf("row index %d out of bounds (table has %d rows)", row, len(t.Rows))
	}
	if col < 0 || col >= len(t.Rows[row].Values) {
		return fmt.Errorf("column index %d out of bounds (row %d has %d columns)", col, row, len(t.Rows[row].Values))
	}
	t.Rows[row].Values[col] = value
	return nil
}
