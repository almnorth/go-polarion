// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Test struct with various field types
type TestCustomWorkItem struct {
	StringField   *string      `json:"stringField"`
	IntField      *int         `json:"intField"`
	FloatField    *float64     `json:"floatField"`
	BoolField     *bool        `json:"boolField"`
	DateField     *DateOnly    `json:"dateField"`
	TimeField     *TimeOnly    `json:"timeField"`
	DateTimeField *DateTime    `json:"dateTimeField"`
	DurationField *Duration    `json:"durationField"`
	TextField     *TextContent `json:"textField"`
	IgnoredField  *string      // No JSON tag - should be ignored
	SkippedField  *string      `json:"-"` // Explicitly skipped
}

// testPlatform is a named string type, used to verify that multi-value fields
// support named element types (e.g. generated enum constants).
type testPlatform string

// testPosition is a named int64 type, used to verify that named numeric types work.
type testPosition int64

// Test struct covering the numeric widths a Polarion integer field can be mapped to
type TestNumericWorkItem struct {
	Int64Field    *int64        `json:"int64Field"`
	Int32Field    *int32        `json:"int32Field"`
	Int16Field    *int16        `json:"int16Field"`
	Int8Field     *int8         `json:"int8Field"`
	IntField      *int          `json:"intField"`
	UintField     *uint         `json:"uintField"`
	Uint64Field   *uint64       `json:"uint64Field"`
	Float32Field  *float32      `json:"float32Field"`
	Float64Field  *float64      `json:"float64Field"`
	PositionField *testPosition `json:"positionField"`
}

// Test struct for multi-value fields (multi-enumeration, multi-value string)
type TestMultiValueWorkItem struct {
	MultiEnumField  []string       `json:"multiEnumField"`
	NamedEnumField  []testPlatform `json:"namedEnumField"`
	PointerToSlice  *[]string      `json:"pointerToSlice"`
	SingleEnumField *string        `json:"singleEnumField"`
}

func TestLoadCustomFields(t *testing.T) {
	// Create a work item with custom fields
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"stringField":   "test value",
				"intField":      42,
				"floatField":    3.14,
				"boolField":     true,
				"dateField":     "2026-06-15",
				"timeField":     "14:30:00",
				"dateTimeField": "2026-06-15T14:30:00Z",
				"durationField": "2d 3h 30m",
				"textField": map[string]interface{}{
					"type":  "text/html",
					"value": "<p>Test content</p>",
				},
			},
		},
	}

	// Load into custom struct
	custom := &TestCustomWorkItem{}
	err := LoadCustomFields(wi, custom)
	if err != nil {
		t.Fatalf("LoadCustomFields failed: %v", err)
	}

	// Verify string field
	if custom.StringField == nil || *custom.StringField != "test value" {
		t.Errorf("StringField: expected 'test value', got %v", custom.StringField)
	}

	// Verify int field
	if custom.IntField == nil || *custom.IntField != 42 {
		t.Errorf("IntField: expected 42, got %v", custom.IntField)
	}

	// Verify float field
	if custom.FloatField == nil || *custom.FloatField != 3.14 {
		t.Errorf("FloatField: expected 3.14, got %v", custom.FloatField)
	}

	// Verify bool field
	if custom.BoolField == nil || *custom.BoolField != true {
		t.Errorf("BoolField: expected true, got %v", custom.BoolField)
	}

	// Verify date field
	if custom.DateField == nil {
		t.Error("DateField: expected non-nil value")
	} else {
		expected := "2026-06-15"
		if custom.DateField.String() != expected {
			t.Errorf("DateField: expected %s, got %s", expected, custom.DateField.String())
		}
	}

	// Verify time field
	if custom.TimeField == nil {
		t.Error("TimeField: expected non-nil value")
	} else {
		expected := "14:30:00"
		if custom.TimeField.String() != expected {
			t.Errorf("TimeField: expected %s, got %s", expected, custom.TimeField.String())
		}
	}

	// Verify datetime field
	if custom.DateTimeField == nil {
		t.Error("DateTimeField: expected non-nil value")
	}

	// Verify duration field
	if custom.DurationField == nil {
		t.Error("DurationField: expected non-nil value")
	}

	// Verify text field
	if custom.TextField == nil {
		t.Error("TextField: expected non-nil value")
	} else {
		if custom.TextField.Type != "text/html" {
			t.Errorf("TextField.Type: expected 'text/html', got %s", custom.TextField.Type)
		}
		if custom.TextField.Value != "<p>Test content</p>" {
			t.Errorf("TextField.Value: expected '<p>Test content</p>', got %s", custom.TextField.Value)
		}
	}

	// Verify ignored fields are nil
	if custom.IgnoredField != nil {
		t.Error("IgnoredField: expected nil (no json tag)")
	}
	if custom.SkippedField != nil {
		t.Error("SkippedField: expected nil (json:\"-\")")
	}
}

func TestLoadCustomFields_NumericWidths(t *testing.T) {
	// Polarion integers arrive as JSON numbers, which unmarshal to float64
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"int64Field":    float64(9007199254740992),
				"int32Field":    float64(2147483647),
				"int16Field":    float64(-32768),
				"int8Field":     float64(127),
				"intField":      float64(42),
				"uintField":     float64(7),
				"uint64Field":   float64(18446744073709551615 >> 12),
				"float32Field":  float64(3.5),
				"float64Field":  float64(3.14),
				"positionField": float64(1234567890123),
			},
		},
	}

	custom := &TestNumericWorkItem{}
	if err := LoadCustomFields(wi, custom); err != nil {
		t.Fatalf("LoadCustomFields failed: %v", err)
	}

	if custom.Int64Field == nil || *custom.Int64Field != 9007199254740992 {
		t.Errorf("Int64Field: expected 9007199254740992, got %v", custom.Int64Field)
	}
	if custom.Int32Field == nil || *custom.Int32Field != 2147483647 {
		t.Errorf("Int32Field: expected 2147483647, got %v", custom.Int32Field)
	}
	if custom.Int16Field == nil || *custom.Int16Field != -32768 {
		t.Errorf("Int16Field: expected -32768, got %v", custom.Int16Field)
	}
	if custom.Int8Field == nil || *custom.Int8Field != 127 {
		t.Errorf("Int8Field: expected 127, got %v", custom.Int8Field)
	}
	if custom.IntField == nil || *custom.IntField != 42 {
		t.Errorf("IntField: expected 42, got %v", custom.IntField)
	}
	if custom.UintField == nil || *custom.UintField != 7 {
		t.Errorf("UintField: expected 7, got %v", custom.UintField)
	}
	if custom.Uint64Field == nil {
		t.Error("Uint64Field: expected non-nil value")
	}
	if custom.Float32Field == nil || *custom.Float32Field != 3.5 {
		t.Errorf("Float32Field: expected 3.5, got %v", custom.Float32Field)
	}
	if custom.Float64Field == nil || *custom.Float64Field != 3.14 {
		t.Errorf("Float64Field: expected 3.14, got %v", custom.Float64Field)
	}
	if custom.PositionField == nil || *custom.PositionField != 1234567890123 {
		t.Errorf("PositionField: expected 1234567890123, got %v", custom.PositionField)
	}
}

func TestLoadCustomFields_NumericOverflow(t *testing.T) {
	// A value that does not fit the target type must error rather than truncate
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"int8Field": float64(300),
			},
		},
	}

	if err := LoadCustomFields(wi, &TestNumericWorkItem{}); err == nil {
		t.Error("expected an overflow error for 300 into *int8")
	}

	// Same for a negative value in an unsigned field
	wi.Attributes.CustomFields = map[string]interface{}{"uintField": float64(-1)}
	if err := LoadCustomFields(wi, &TestNumericWorkItem{}); err == nil {
		t.Error("expected an error for -1 into *uint")
	}
}

func TestSaveCustomFields_NumericWidths(t *testing.T) {
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title:        "Test Work Item",
			CustomFields: make(map[string]interface{}),
		},
	}

	int64Val := int64(9007199254740993) // Not representable as float64
	uintVal := uint(7)
	float32Val := float32(3.5)
	positionVal := testPosition(1234567890123)

	custom := &TestNumericWorkItem{
		Int64Field:    &int64Val,
		UintField:     &uintVal,
		Float32Field:  &float32Val,
		PositionField: &positionVal,
	}

	if err := SaveCustomFields(wi, custom); err != nil {
		t.Fatalf("SaveCustomFields failed: %v", err)
	}

	cf := CustomFields(wi.Attributes.CustomFields)

	// Integers are stored as int64, so exact large values survive the round-trip
	if val, ok := cf.GetInt64("int64Field"); !ok || val != int64Val {
		t.Errorf("int64Field: expected %d, got %d (ok=%t)", int64Val, val, ok)
	}
	if val, ok := cf.GetInt64("uintField"); !ok || val != 7 {
		t.Errorf("uintField: expected 7, got %d (ok=%t)", val, ok)
	}
	if val, ok := cf.GetFloat("float32Field"); !ok || val != 3.5 {
		t.Errorf("float32Field: expected 3.5, got %v (ok=%t)", val, ok)
	}
	if val, ok := cf.GetInt64("positionField"); !ok || val != 1234567890123 {
		t.Errorf("positionField: expected 1234567890123, got %d (ok=%t)", val, ok)
	}

	// The payload must carry plain JSON numbers
	data, err := json.Marshal(wi.Attributes)
	if err != nil {
		t.Fatalf("marshaling attributes failed: %v", err)
	}
	if !strings.Contains(string(data), `"int64Field":9007199254740993`) {
		t.Errorf("expected int64Field to marshal as an exact JSON number, got: %s", data)
	}
}

func TestLoadCustomFields_MultiValue(t *testing.T) {
	// Values as they arrive from the API: JSON arrays unmarshal to []interface{}
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"multiEnumField":  []interface{}{"opt1", "opt2"},
				"namedEnumField":  []interface{}{"linux", "windows"},
				"pointerToSlice":  []string{"a", "b"},
				"singleEnumField": "opt1",
			},
		},
	}

	custom := &TestMultiValueWorkItem{}
	if err := LoadCustomFields(wi, custom); err != nil {
		t.Fatalf("LoadCustomFields failed: %v", err)
	}

	if len(custom.MultiEnumField) != 2 || custom.MultiEnumField[0] != "opt1" || custom.MultiEnumField[1] != "opt2" {
		t.Errorf("MultiEnumField: expected [opt1 opt2], got %v", custom.MultiEnumField)
	}

	if len(custom.NamedEnumField) != 2 || custom.NamedEnumField[0] != "linux" || custom.NamedEnumField[1] != "windows" {
		t.Errorf("NamedEnumField: expected [linux windows], got %v", custom.NamedEnumField)
	}

	if custom.PointerToSlice == nil {
		t.Error("PointerToSlice: expected non-nil value")
	} else if len(*custom.PointerToSlice) != 2 || (*custom.PointerToSlice)[0] != "a" {
		t.Errorf("PointerToSlice: expected [a b], got %v", *custom.PointerToSlice)
	}

	if custom.SingleEnumField == nil || *custom.SingleEnumField != "opt1" {
		t.Errorf("SingleEnumField: expected 'opt1', got %v", custom.SingleEnumField)
	}
}

func TestLoadCustomFields_MultiValueSingleString(t *testing.T) {
	// A multi-value field may come back as a plain string when it holds one value
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"multiEnumField": "opt1",
			},
		},
	}

	custom := &TestMultiValueWorkItem{}
	if err := LoadCustomFields(wi, custom); err != nil {
		t.Fatalf("LoadCustomFields failed: %v", err)
	}

	if len(custom.MultiEnumField) != 1 || custom.MultiEnumField[0] != "opt1" {
		t.Errorf("MultiEnumField: expected [opt1], got %v", custom.MultiEnumField)
	}
}

func TestSaveCustomFields_MultiValue(t *testing.T) {
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title:        "Test Work Item",
			CustomFields: make(map[string]interface{}),
		},
	}

	singleVal := "opt1"
	pointerSlice := []string{"a", "b"}
	custom := &TestMultiValueWorkItem{
		MultiEnumField:  []string{"opt1", "opt2"},
		NamedEnumField:  []testPlatform{"linux"},
		PointerToSlice:  &pointerSlice,
		SingleEnumField: &singleVal,
	}

	if err := SaveCustomFields(wi, custom); err != nil {
		t.Fatalf("SaveCustomFields failed: %v", err)
	}

	cf := CustomFields(wi.Attributes.CustomFields)

	if val, ok := cf.GetEnums("multiEnumField"); !ok || len(val) != 2 || val[0] != "opt1" || val[1] != "opt2" {
		t.Errorf("multiEnumField: expected [opt1 opt2], got %v", val)
	}

	if val, ok := cf.GetStringSlice("namedEnumField"); !ok || len(val) != 1 || val[0] != "linux" {
		t.Errorf("namedEnumField: expected [linux], got %v", val)
	}

	if val, ok := cf.GetStringSlice("pointerToSlice"); !ok || len(val) != 2 || val[0] != "a" {
		t.Errorf("pointerToSlice: expected [a b], got %v", val)
	}

	// Verify the stored value marshals to a JSON array, which is what Polarion expects
	data, err := json.Marshal(wi.Attributes)
	if err != nil {
		t.Fatalf("marshaling attributes failed: %v", err)
	}
	if !strings.Contains(string(data), `"multiEnumField":["opt1","opt2"]`) {
		t.Errorf("expected multiEnumField to marshal as a JSON array, got: %s", data)
	}
}

func TestSaveCustomFields_MultiValueNilAndEmpty(t *testing.T) {
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"multiEnumField": []string{"opt1"},
				"namedEnumField": []string{"linux"},
			},
		},
	}

	// nil slice removes the field; empty non-nil slice clears it in Polarion
	custom := &TestMultiValueWorkItem{
		MultiEnumField: nil,
		NamedEnumField: []testPlatform{},
	}

	if err := SaveCustomFields(wi, custom); err != nil {
		t.Fatalf("SaveCustomFields failed: %v", err)
	}

	cf := CustomFields(wi.Attributes.CustomFields)

	if cf.Has("multiEnumField") {
		t.Error("multiEnumField: expected to be deleted for a nil slice")
	}

	val, ok := cf.GetStringSlice("namedEnumField")
	if !ok {
		t.Error("namedEnumField: expected to be present for an empty slice")
	}
	if len(val) != 0 {
		t.Errorf("namedEnumField: expected empty slice, got %v", val)
	}
}

func TestSaveCustomFields(t *testing.T) {
	// Create a work item
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title:        "Test Work Item",
			CustomFields: make(map[string]interface{}),
		},
	}

	// Create custom struct with values
	stringVal := "test value"
	intVal := 42
	floatVal := 3.14
	boolVal := true
	dateVal := NewDateOnly(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	timeVal, _ := NewTimeOnly(14, 30, 0)
	dateTimeVal := NewDateTime(time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC))
	durationVal := NewDuration(2*24*time.Hour + 3*time.Hour + 30*time.Minute)
	textVal := &TextContent{Type: "text/html", Value: "<p>Test content</p>"}

	custom := &TestCustomWorkItem{
		StringField:   &stringVal,
		IntField:      &intVal,
		FloatField:    &floatVal,
		BoolField:     &boolVal,
		DateField:     &dateVal,
		TimeField:     &timeVal,
		DateTimeField: &dateTimeVal,
		DurationField: &durationVal,
		TextField:     textVal,
	}

	// Save to work item
	err := SaveCustomFields(wi, custom)
	if err != nil {
		t.Fatalf("SaveCustomFields failed: %v", err)
	}

	cf := CustomFields(wi.Attributes.CustomFields)

	// Verify string field
	if val, ok := cf.GetString("stringField"); !ok || val != "test value" {
		t.Errorf("stringField: expected 'test value', got %v", val)
	}

	// Verify int field
	if val, ok := cf.GetInt("intField"); !ok || val != 42 {
		t.Errorf("intField: expected 42, got %v", val)
	}

	// Verify float field
	if val, ok := cf.GetFloat("floatField"); !ok || val != 3.14 {
		t.Errorf("floatField: expected 3.14, got %v", val)
	}

	// Verify bool field
	if val, ok := cf.GetBool("boolField"); !ok || val != true {
		t.Errorf("boolField: expected true, got %v", val)
	}

	// Verify date field
	if val, ok := cf.GetDateOnly("dateField"); !ok {
		t.Error("dateField: expected to be set")
	} else if val.String() != "2026-06-15" {
		t.Errorf("dateField: expected '2026-06-15', got %s", val.String())
	}

	// Verify time field
	if val, ok := cf.GetTimeOnly("timeField"); !ok {
		t.Error("timeField: expected to be set")
	} else if val.String() != "14:30:00" {
		t.Errorf("timeField: expected '14:30:00', got %s", val.String())
	}

	// Verify datetime field
	if _, ok := cf.GetDateTime("dateTimeField"); !ok {
		t.Error("dateTimeField: expected to be set")
	}

	// Verify duration field
	if _, ok := cf.GetDuration("durationField"); !ok {
		t.Error("durationField: expected to be set")
	}

	// Verify text field
	if val, ok := cf.GetText("textField"); !ok {
		t.Error("textField: expected to be set")
	} else {
		if val.Type != "text/html" {
			t.Errorf("textField.Type: expected 'text/html', got %s", val.Type)
		}
		if val.Value != "<p>Test content</p>" {
			t.Errorf("textField.Value: expected '<p>Test content</p>', got %s", val.Value)
		}
	}
}

func TestSaveCustomFields_NilValues(t *testing.T) {
	// Create a work item with existing custom fields
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"stringField": "existing value",
				"intField":    42,
			},
		},
	}

	// Create custom struct with nil values (should delete fields)
	custom := &TestCustomWorkItem{
		StringField: nil,
		IntField:    nil,
	}

	// Save to work item
	err := SaveCustomFields(wi, custom)
	if err != nil {
		t.Fatalf("SaveCustomFields failed: %v", err)
	}

	cf := CustomFields(wi.Attributes.CustomFields)

	// Verify fields are deleted
	if cf.Has("stringField") {
		t.Error("stringField: expected to be deleted")
	}
	if cf.Has("intField") {
		t.Error("intField: expected to be deleted")
	}
}

func TestLoadCustomFields_MissingFields(t *testing.T) {
	// Create a work item with no custom fields
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title:        "Test Work Item",
			CustomFields: make(map[string]interface{}),
		},
	}

	// Load into custom struct
	custom := &TestCustomWorkItem{}
	err := LoadCustomFields(wi, custom)
	if err != nil {
		t.Fatalf("LoadCustomFields failed: %v", err)
	}

	// Verify all fields are nil
	if custom.StringField != nil {
		t.Error("StringField: expected nil")
	}
	if custom.IntField != nil {
		t.Error("IntField: expected nil")
	}
	if custom.FloatField != nil {
		t.Error("FloatField: expected nil")
	}
	if custom.BoolField != nil {
		t.Error("BoolField: expected nil")
	}
}

func TestLoadCustomFields_InvalidInput(t *testing.T) {
	wi := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title:        "Test Work Item",
			CustomFields: make(map[string]interface{}),
		},
	}

	// Test with non-pointer
	custom := TestCustomWorkItem{}
	err := LoadCustomFields(wi, custom)
	if err == nil {
		t.Error("Expected error when passing non-pointer")
	}

	// Test with nil work item
	err = LoadCustomFields(nil, &custom)
	if err == nil {
		t.Error("Expected error when passing nil work item")
	}

	// Test with nil attributes
	wiNoAttrs := &WorkItem{ID: "TEST-123", Type: "workitems"}
	err = LoadCustomFields(wiNoAttrs, &custom)
	if err == nil {
		t.Error("Expected error when work item has nil attributes")
	}
}

func TestSaveCustomFields_InvalidInput(t *testing.T) {
	custom := &TestCustomWorkItem{}

	// Test with nil work item
	err := SaveCustomFields(nil, custom)
	if err == nil {
		t.Error("Expected error when passing nil work item")
	}

	// Test with nil attributes
	wiNoAttrs := &WorkItem{ID: "TEST-123", Type: "workitems"}
	err = SaveCustomFields(wiNoAttrs, custom)
	if err == nil {
		t.Error("Expected error when work item has nil attributes")
	}
}

func TestRoundTrip(t *testing.T) {
	// Create original work item with custom fields
	original := &WorkItem{
		ID:   "TEST-123",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title: "Test Work Item",
			CustomFields: map[string]interface{}{
				"stringField": "test value",
				"intField":    42,
				"floatField":  3.14,
				"boolField":   true,
			},
		},
	}

	// Load into custom struct
	custom := &TestCustomWorkItem{}
	err := LoadCustomFields(original, custom)
	if err != nil {
		t.Fatalf("LoadCustomFields failed: %v", err)
	}

	// Create new work item and save
	newWI := &WorkItem{
		ID:   "TEST-456",
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			Title:        "New Work Item",
			CustomFields: make(map[string]interface{}),
		},
	}

	err = SaveCustomFields(newWI, custom)
	if err != nil {
		t.Fatalf("SaveCustomFields failed: %v", err)
	}

	// Verify fields match
	cf := CustomFields(newWI.Attributes.CustomFields)

	if val, ok := cf.GetString("stringField"); !ok || val != "test value" {
		t.Errorf("stringField: expected 'test value', got %v", val)
	}
	if val, ok := cf.GetInt("intField"); !ok || val != 42 {
		t.Errorf("intField: expected 42, got %v", val)
	}
	if val, ok := cf.GetFloat("floatField"); !ok || val != 3.14 {
		t.Errorf("floatField: expected 3.14, got %v", val)
	}
	if val, ok := cf.GetBool("boolField"); !ok || val != true {
		t.Errorf("boolField: expected true, got %v", val)
	}
}
