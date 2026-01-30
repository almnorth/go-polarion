// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"encoding/json"
	"strconv"
)

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

// RelationshipType represents the type of a relationship reference in Polarion.
// These are the standard types used in Polarion's REST API.
type RelationshipType string

const (
	// RelationshipTypeUsers represents a user reference
	RelationshipTypeUsers RelationshipType = "users"

	// RelationshipTypeWorkItems represents a work item reference
	RelationshipTypeWorkItems RelationshipType = "workitems"

	// RelationshipTypeDocuments represents a document reference
	RelationshipTypeDocuments RelationshipType = "documents"

	// RelationshipTypeCategories represents a category reference
	RelationshipTypeCategories RelationshipType = "categories"

	// RelationshipTypePlans represents a plan reference
	RelationshipTypePlans RelationshipType = "plans"

	// RelationshipTypeCollections represents a collection reference
	RelationshipTypeCollections RelationshipType = "collections"

	// RelationshipTypeWorkItemComments represents a work item comment reference
	RelationshipTypeWorkItemComments RelationshipType = "workitem_comments"

	// RelationshipTypeWorkItemAttachments represents a work item attachment reference
	RelationshipTypeWorkItemAttachments RelationshipType = "workitem_attachments"

	// RelationshipTypeProjects represents a project reference
	RelationshipTypeProjects RelationshipType = "projects"

	// RelationshipTypeLinkedWorkItems represents a linked work item reference
	RelationshipTypeLinkedWorkItems RelationshipType = "linkedworkitems"
)

// UserRef represents a user reference field in Polarion.
// This type handles the JSON marshaling/unmarshaling of user reference custom fields
// which are stored as relationships with type "users".
//
// UserRef can be used in custom field structs for type-safe access:
//
//	type MyScopeItem struct {
//	    ResponsiblePurchaser *polarion.UserRef `json:"Chairman,omitempty"`
//	}
//
// When unmarshaling from Polarion API responses, it handles the relationship structure:
//
//	{"data": {"type": "users", "id": "john.doe"}}
//
// When marshaling for API requests, it produces the same structure.
// For simple access, use the ID field directly or the String() method.
type UserRef struct {
	// ID is the user identifier (e.g., "john.doe", "SSV005")
	ID string
}

// NewUserRef creates a new UserRef with the given user ID.
// Returns nil if the userID is empty.
func NewUserRef(userID string) *UserRef {
	if userID == "" {
		return nil
	}
	return &UserRef{ID: userID}
}

// String returns the user ID as a string.
func (u UserRef) String() string {
	return u.ID
}

// IsEmpty returns true if the UserRef has no ID set.
func (u UserRef) IsEmpty() bool {
	return u.ID == ""
}

// MarshalJSON implements json.Marshaler for UserRef.
// It produces the relationship structure that Polarion expects:
// {"data": {"type": "users", "id": "john.doe"}}
func (u UserRef) MarshalJSON() ([]byte, error) {
	if u.ID == "" {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"type": "users",
			"id":   u.ID,
		},
	})
}

// UnmarshalJSON implements json.Unmarshaler for UserRef.
// It handles multiple formats:
// - Relationship structure: {"data": {"type": "users", "id": "john.doe"}}
// - Array of relationships: {"data": [{"type": "users", "id": "john.doe"}]}
// - Simple string: "john.doe"
func (u *UserRef) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		u.ID = ""
		return nil
	}

	// Try to unmarshal as a simple string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		u.ID = str
		return nil
	}

	// Try to unmarshal as a relationship structure
	var rel struct {
		Data interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &rel); err != nil {
		return err
	}

	// Handle single data object: {"data": {"type": "users", "id": "john.doe"}}
	if dataMap, ok := rel.Data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			u.ID = id
			return nil
		}
	}

	// Handle array of data: {"data": [{"type": "users", "id": "john.doe"}]}
	if dataArray, ok := rel.Data.([]interface{}); ok && len(dataArray) > 0 {
		if firstItem, ok := dataArray[0].(map[string]interface{}); ok {
			if id, ok := firstItem["id"].(string); ok {
				u.ID = id
				return nil
			}
		}
	}

	return nil
}

// ToRelationship converts the UserRef to a Relationship structure
// suitable for use in WorkItemRelationships.CustomRelationships.
func (u *UserRef) ToRelationship() *Relationship {
	if u == nil || u.ID == "" {
		return nil
	}
	return &Relationship{
		Data: map[string]interface{}{
			"type": "users",
			"id":   u.ID,
		},
	}
}

// ToRelationshipReference converts the UserRef to a RelationshipReference.
func (u *UserRef) ToRelationshipReference() *RelationshipReference {
	if u == nil || u.ID == "" {
		return nil
	}
	return &RelationshipReference{
		Type: RelationshipTypeUsers,
		ID:   u.ID,
	}
}

// UserRefFromRelationship creates a UserRef from a Relationship.
// Returns nil if the relationship is nil or doesn't contain a valid user reference.
func UserRefFromRelationship(rel *Relationship) *UserRef {
	if rel == nil || rel.Data == nil {
		return nil
	}

	// Handle map[string]interface{} from JSON unmarshaling (single value)
	if data, ok := rel.Data.(map[string]interface{}); ok {
		if dataType, ok := data["type"].(string); ok && dataType == "users" {
			if id, ok := data["id"].(string); ok {
				return &UserRef{ID: id}
			}
		}
	}

	// Handle []interface{} from JSON unmarshaling (array - return first)
	if dataArray, ok := rel.Data.([]interface{}); ok && len(dataArray) > 0 {
		if firstItem, ok := dataArray[0].(map[string]interface{}); ok {
			if dataType, ok := firstItem["type"].(string); ok && dataType == "users" {
				if id, ok := firstItem["id"].(string); ok {
					return &UserRef{ID: id}
				}
			}
		}
	}

	return nil
}

// RelationshipReference represents a reference to another resource in Polarion.
// This is used for custom fields that reference other resources (users, work items, etc.).
// The structure matches Polarion's JSON:API relationship format.
type RelationshipReference struct {
	// Type is the type of the referenced resource (e.g., "users", "workitems")
	Type RelationshipType `json:"type"`

	// ID is the identifier of the referenced resource
	// For users: just the user ID (e.g., "john.doe")
	// For work items: project/workitem format (e.g., "myProject/WI-123")
	// For categories: project/category format (e.g., "myProject/interface")
	ID string `json:"id"`

	// Revision is an optional revision for the reference (used for versioned references)
	Revision string `json:"revision,omitempty"`
}

// NewRelationshipReference creates a new RelationshipReference with the given type and ID.
func NewRelationshipReference(refType RelationshipType, id string) *RelationshipReference {
	return &RelationshipReference{
		Type: refType,
		ID:   id,
	}
}

// NewUserReference creates a new RelationshipReference for a user.
func NewUserReference(userID string) *RelationshipReference {
	return NewRelationshipReference(RelationshipTypeUsers, userID)
}

// NewWorkItemReference creates a new RelationshipReference for a work item.
// The id should be in the format "projectId/workItemId" (e.g., "myProject/WI-123").
func NewWorkItemReference(id string) *RelationshipReference {
	return NewRelationshipReference(RelationshipTypeWorkItems, id)
}

// NewCategoryReference creates a new RelationshipReference for a category.
// The id should be in the format "projectId/categoryId" (e.g., "myProject/interface").
func NewCategoryReference(id string) *RelationshipReference {
	return NewRelationshipReference(RelationshipTypeCategories, id)
}

// ToRelationship converts the RelationshipReference to a Relationship structure
// suitable for use in WorkItemRelationships.CustomRelationships.
func (r *RelationshipReference) ToRelationship() *Relationship {
	if r == nil || r.ID == "" {
		return nil
	}
	data := map[string]interface{}{
		"type": string(r.Type),
		"id":   r.ID,
	}
	if r.Revision != "" {
		data["revision"] = r.Revision
	}
	return &Relationship{
		Data: data,
	}
}

// ToRelationshipData returns the data portion of a relationship for use in API requests.
// This is useful when you need to set a single-value relationship.
func (r *RelationshipReference) ToRelationshipData() map[string]interface{} {
	if r == nil || r.ID == "" {
		return nil
	}
	data := map[string]interface{}{
		"type": string(r.Type),
		"id":   r.ID,
	}
	if r.Revision != "" {
		data["revision"] = r.Revision
	}
	return data
}

// GetRelationshipReference safely retrieves a relationship reference custom field.
// Relationship reference fields in Polarion are stored as relationships with type and id.
// This method handles both single references and extracts the first item from arrays.
// Returns the RelationshipReference and true if the field exists and contains a valid reference,
// otherwise returns nil and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if ref, ok := cf.GetRelationshipReference("responsiblePurchaser"); ok {
//	    fmt.Printf("Type: %s, ID: %s\n", ref.Type, ref.ID)
//	}
func (cf CustomFields) GetRelationshipReference(key string) (*RelationshipReference, bool) {
	val, exists := cf[key]
	if !exists {
		return nil, false
	}

	// Handle nil value
	if val == nil {
		return nil, false
	}

	// If it's already a string (ID extracted from relationship), we can't determine the type
	// This shouldn't happen in normal usage, but handle it gracefully
	if str, ok := val.(string); ok {
		// Assume it's a user reference if it's just a string (backward compatibility)
		return &RelationshipReference{Type: RelationshipTypeUsers, ID: str}, true
	}

	// Handle RelationshipReference directly
	if ref, ok := val.(*RelationshipReference); ok {
		return ref, true
	}
	if ref, ok := val.(RelationshipReference); ok {
		return &ref, true
	}

	// Handle relationship structure from JSON unmarshaling
	// Structure: {"data": {"type": "users", "id": "user123"}} or {"data": [{"type": "users", "id": "user123"}]}
	if m, ok := val.(map[string]interface{}); ok {
		return extractRelationshipReferenceFromMap(m)
	}

	return nil, false
}

// extractRelationshipReferenceFromMap extracts a RelationshipReference from a map structure.
// Handles both single data objects and arrays (returns first element).
func extractRelationshipReferenceFromMap(m map[string]interface{}) (*RelationshipReference, bool) {
	data := m["data"]
	if data == nil {
		return nil, false
	}

	// Handle single data object: {"data": {"type": "users", "id": "user123"}}
	if dataMap, ok := data.(map[string]interface{}); ok {
		return extractRelationshipReferenceFromData(dataMap)
	}

	// Handle array of data: {"data": [{"type": "users", "id": "user123"}]}
	if dataArray, ok := data.([]interface{}); ok && len(dataArray) > 0 {
		if firstItem, ok := dataArray[0].(map[string]interface{}); ok {
			return extractRelationshipReferenceFromData(firstItem)
		}
	}

	return nil, false
}

// extractRelationshipReferenceFromData extracts a RelationshipReference from a data map.
func extractRelationshipReferenceFromData(data map[string]interface{}) (*RelationshipReference, bool) {
	refType, hasType := data["type"].(string)
	id, hasID := data["id"].(string)

	if !hasType || !hasID {
		return nil, false
	}

	ref := &RelationshipReference{
		Type: RelationshipType(refType),
		ID:   id,
	}

	// Extract optional revision
	if revision, ok := data["revision"].(string); ok {
		ref.Revision = revision
	}

	return ref, true
}

// SetRelationshipReference sets a relationship reference custom field.
// This creates the proper relationship structure that Polarion expects.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	cf.SetRelationshipReference("responsiblePurchaser", polarion.NewUserReference("john.doe"))
func (cf CustomFields) SetRelationshipReference(key string, ref *RelationshipReference) {
	if ref == nil || ref.ID == "" {
		delete(cf, key)
		return
	}
	// Store as relationship structure that Polarion expects
	data := map[string]interface{}{
		"type": string(ref.Type),
		"id":   ref.ID,
	}
	if ref.Revision != "" {
		data["revision"] = ref.Revision
	}
	cf[key] = map[string]interface{}{
		"data": data,
	}
}

// GetUserReference safely retrieves a user reference custom field.
// This is a convenience method that wraps GetRelationshipReference for user references.
// Returns the user ID and true if the field exists and contains a valid user reference,
// otherwise returns empty string and false.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	if userID, ok := cf.GetUserReference("responsiblePurchaser"); ok {
//	    fmt.Printf("Responsible Purchaser: %s\n", userID)
//	}
func (cf CustomFields) GetUserReference(key string) (string, bool) {
	ref, ok := cf.GetRelationshipReference(key)
	if !ok {
		return "", false
	}
	// Verify it's a user reference
	if ref.Type != RelationshipTypeUsers {
		return "", false
	}
	return ref.ID, true
}

// SetUserReference sets a user reference custom field.
// This is a convenience method that wraps SetRelationshipReference for user references.
//
// Example:
//
//	cf := CustomFields(workItem.Attributes.CustomFields)
//	cf.SetUserReference("responsiblePurchaser", "john.doe")
func (cf CustomFields) SetUserReference(key string, userID string) {
	if userID == "" {
		delete(cf, key)
		return
	}
	cf.SetRelationshipReference(key, NewUserReference(userID))
}

// RelationshipReferenceFromRelationship extracts a RelationshipReference from a Relationship.
// Returns the RelationshipReference and true if the relationship contains a valid reference,
// otherwise returns nil and false.
func RelationshipReferenceFromRelationship(rel *Relationship) (*RelationshipReference, bool) {
	if rel == nil || rel.Data == nil {
		return nil, false
	}

	// Handle map[string]interface{} from JSON unmarshaling (single value)
	if data, ok := rel.Data.(map[string]interface{}); ok {
		return extractRelationshipReferenceFromData(data)
	}

	// Handle []interface{} from JSON unmarshaling (array - return first)
	if dataArray, ok := rel.Data.([]interface{}); ok && len(dataArray) > 0 {
		if firstItem, ok := dataArray[0].(map[string]interface{}); ok {
			return extractRelationshipReferenceFromData(firstItem)
		}
	}

	return nil, false
}

// RelationshipReferencesFromRelationship extracts all RelationshipReferences from a Relationship.
// This is useful for multi-value relationships like assignees.
// Returns a slice of RelationshipReferences.
func RelationshipReferencesFromRelationship(rel *Relationship) []*RelationshipReference {
	if rel == nil || rel.Data == nil {
		return nil
	}

	var refs []*RelationshipReference

	// Handle map[string]interface{} from JSON unmarshaling (single value)
	if data, ok := rel.Data.(map[string]interface{}); ok {
		if ref, ok := extractRelationshipReferenceFromData(data); ok {
			refs = append(refs, ref)
		}
		return refs
	}

	// Handle []interface{} from JSON unmarshaling (array)
	if dataArray, ok := rel.Data.([]interface{}); ok {
		for _, item := range dataArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if ref, ok := extractRelationshipReferenceFromData(itemMap); ok {
					refs = append(refs, ref)
				}
			}
		}
	}

	return refs
}
