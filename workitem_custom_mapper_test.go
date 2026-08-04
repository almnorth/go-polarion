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
