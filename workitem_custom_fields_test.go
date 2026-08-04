// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"
)

func TestCustomFields_GetInt64(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected int64
		expectOK bool
	}{
		{name: "float64 from JSON", value: float64(42), expected: 42, expectOK: true},
		{name: "large float64", value: float64(1234567890123), expected: 1234567890123, expectOK: true},
		{name: "int", value: 42, expected: 42, expectOK: true},
		{name: "int32", value: int32(42), expected: 42, expectOK: true},
		{name: "int64", value: int64(9007199254740993), expected: 9007199254740993, expectOK: true},
		{name: "uint64", value: uint64(42), expected: 42, expectOK: true},
		{name: "negative", value: float64(-7), expected: -7, expectOK: true},
		{name: "truncated float", value: 3.9, expected: 3, expectOK: true},
		{name: "json.Number", value: json.Number("42"), expected: 42, expectOK: true},
		{name: "numeric string", value: "42", expected: 42, expectOK: true},
		{name: "padded numeric string", value: " 42 ", expected: 42, expectOK: true},
		{name: "non-numeric string", value: "high", expected: 0, expectOK: false},
		{name: "uint64 above MaxInt64", value: uint64(math.MaxUint64), expected: 0, expectOK: false},
		{name: "bool", value: true, expected: 0, expectOK: false},
		{name: "nil", value: nil, expected: 0, expectOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := CustomFields{"field": tt.value}
			val, ok := cf.GetInt64("field")
			if ok != tt.expectOK {
				t.Fatalf("ok: expected %t, got %t", tt.expectOK, ok)
			}
			if val != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, val)
			}
		})
	}

	// Missing key
	empty := CustomFields{}
	if _, ok := empty.GetInt64("field"); ok {
		t.Error("expected false for a missing key")
	}
}

func TestCustomFields_GetIntOverflow(t *testing.T) {
	cf := CustomFields{"field": int64(math.MaxInt64)}

	// GetInt64 keeps the value; GetInt refuses to truncate it on 32-bit platforms
	if val, ok := cf.GetInt64("field"); !ok || val != math.MaxInt64 {
		t.Errorf("GetInt64: expected %d, got %d (ok=%t)", int64(math.MaxInt64), val, ok)
	}

	val, ok := cf.GetInt("field")
	if strconv.IntSize == 64 {
		if !ok || int64(val) != math.MaxInt64 {
			t.Errorf("GetInt: expected %d, got %d (ok=%t)", int64(math.MaxInt64), val, ok)
		}
	} else if ok {
		t.Error("GetInt: expected false when the value does not fit an int")
	}
}

func TestCustomFields_GetStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		fields   CustomFields
		key      string
		expected []string
		expectOK bool
	}{
		{
			name:     "array from JSON unmarshaling",
			fields:   CustomFields{"multiEnum": []interface{}{"opt1", "opt2"}},
			key:      "multiEnum",
			expected: []string{"opt1", "opt2"},
			expectOK: true,
		},
		{
			name:     "string slice set by the caller",
			fields:   CustomFields{"multiEnum": []string{"opt1", "opt2"}},
			key:      "multiEnum",
			expected: []string{"opt1", "opt2"},
			expectOK: true,
		},
		{
			name:     "single value as plain string",
			fields:   CustomFields{"multiEnum": "opt1"},
			key:      "multiEnum",
			expected: []string{"opt1"},
			expectOK: true,
		},
		{
			name:     "empty array",
			fields:   CustomFields{"multiEnum": []interface{}{}},
			key:      "multiEnum",
			expected: []string{},
			expectOK: true,
		},
		{
			name:     "empty string",
			fields:   CustomFields{"multiEnum": ""},
			key:      "multiEnum",
			expected: []string{},
			expectOK: true,
		},
		{
			name:     "non-string elements are skipped",
			fields:   CustomFields{"multiEnum": []interface{}{"opt1", 42, nil, "opt2"}},
			key:      "multiEnum",
			expected: []string{"opt1", "opt2"},
			expectOK: true,
		},
		{
			name:     "missing key",
			fields:   CustomFields{},
			key:      "multiEnum",
			expected: nil,
			expectOK: false,
		},
		{
			name:     "nil value",
			fields:   CustomFields{"multiEnum": nil},
			key:      "multiEnum",
			expected: nil,
			expectOK: false,
		},
		{
			name:     "unsupported type",
			fields:   CustomFields{"multiEnum": 42},
			key:      "multiEnum",
			expected: nil,
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := tt.fields.GetStringSlice(tt.key)
			if ok != tt.expectOK {
				t.Fatalf("ok: expected %t, got %t", tt.expectOK, ok)
			}
			if len(val) != len(tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, val)
			}
			for i := range tt.expected {
				if val[i] != tt.expected[i] {
					t.Errorf("index %d: expected %q, got %q", i, tt.expected[i], val[i])
				}
			}
			if tt.expectOK && tt.expected != nil && val == nil {
				t.Error("expected a non-nil slice")
			}
		})
	}
}

func TestCustomFields_GetStringSliceReturnsCopy(t *testing.T) {
	cf := CustomFields{"multiEnum": []string{"opt1", "opt2"}}

	val, ok := cf.GetStringSlice("multiEnum")
	if !ok {
		t.Fatal("expected the field to be present")
	}

	// Mutating the returned slice must not affect the stored value
	val[0] = "mutated"

	stored, _ := cf.GetStringSlice("multiEnum")
	if stored[0] != "opt1" {
		t.Errorf("stored value was mutated: expected 'opt1', got %q", stored[0])
	}
}

func TestCustomFields_GetEnumOnMultiEnum(t *testing.T) {
	cf := CustomFields{"multiEnum": []interface{}{"opt1", "opt2"}}

	// GetEnum/GetString cannot represent multiple values, so they report absence
	if _, ok := cf.GetEnum("multiEnum"); ok {
		t.Error("GetEnum: expected false for a multi-enumeration field")
	}

	if _, ok := cf.GetEnums("multiEnum"); !ok {
		t.Error("GetEnums: expected true for a multi-enumeration field")
	}
}

func TestCustomFields_SetStringSlice(t *testing.T) {
	cf := CustomFields{}

	cf.SetEnums("multiEnum", []string{"opt1", "opt2"})
	val, ok := cf.GetEnums("multiEnum")
	if !ok || len(val) != 2 || val[0] != "opt1" || val[1] != "opt2" {
		t.Errorf("expected [opt1 opt2], got %v (ok=%t)", val, ok)
	}

	// Empty non-nil slice stores an empty array (clears the field in Polarion)
	cf.SetEnums("multiEnum", []string{})
	val, ok = cf.GetEnums("multiEnum")
	if !ok {
		t.Error("expected the field to still be present after setting an empty slice")
	}
	if len(val) != 0 {
		t.Errorf("expected an empty slice, got %v", val)
	}

	// Nil slice removes the field entirely
	cf.SetEnums("multiEnum", nil)
	if cf.Has("multiEnum") {
		t.Error("expected the field to be removed for a nil slice")
	}
}

func TestCustomFields_SetStringSliceStoresCopy(t *testing.T) {
	cf := CustomFields{}
	values := []string{"opt1", "opt2"}

	cf.SetStringSlice("multiEnum", values)

	// Mutating the caller's slice must not affect the stored value
	values[0] = "mutated"

	stored, _ := cf.GetStringSlice("multiEnum")
	if stored[0] != "opt1" {
		t.Errorf("stored value was mutated: expected 'opt1', got %q", stored[0])
	}
}
