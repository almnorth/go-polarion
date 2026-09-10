// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- test doubles ---------------------------------------------------------

// fakeStore records what the engine asks for. It embeds a zero WorkItemService
// so Equals/EqualsWithDiff run the library's real comparison code, which never
// touches HTTP.
type fakeStore struct {
	*WorkItemService

	created [][]*WorkItem
	updated [][]UpdatePair
	deleted [][]string

	createErr error
	updateErr error
	deleteErr error

	// createIDsUpTo limits how many items get an ID assigned, to simulate a
	// partially applied multi-batch create.
	createIDsUpTo int
	nextID        int
}

func newFakeStore() *fakeStore {
	return &fakeStore{WorkItemService: &WorkItemService{}, createIDsUpTo: -1}
}

func (f *fakeStore) Create(_ context.Context, items ...*WorkItem) error {
	batch := append([]*WorkItem(nil), items...)
	f.created = append(f.created, batch)
	for i, wi := range items {
		if f.createIDsUpTo >= 0 && i >= f.createIDsUpTo {
			break
		}
		f.nextID++
		wi.ID = fmt.Sprintf("MYPROJ/WI-%d", f.nextID)
	}
	return f.createErr
}

func (f *fakeStore) UpdateBatchWithOldValues(_ context.Context, pairs ...UpdatePair) error {
	f.updated = append(f.updated, append([]UpdatePair(nil), pairs...))
	return f.updateErr
}

func (f *fakeStore) Delete(_ context.Context, ids ...string) error {
	f.deleted = append(f.deleted, append([]string(nil), ids...))
	return f.deleteErr
}

// recordingLogger captures log lines so tests can assert on diff output.
type recordingLogger struct{ lines []string }

func (l *recordingLogger) Debugf(f string, a ...interface{}) { l.add("DEBUG", f, a...) }
func (l *recordingLogger) Infof(f string, a ...interface{})  { l.add("INFO", f, a...) }
func (l *recordingLogger) Warnf(f string, a ...interface{})  { l.add("WARN", f, a...) }
func (l *recordingLogger) Errorf(f string, a ...interface{}) { l.add("ERROR", f, a...) }
func (l *recordingLogger) add(level, f string, a ...interface{}) {
	l.lines = append(l.lines, level+" "+fmt.Sprintf(f, a...))
}
func (l *recordingLogger) contains(sub string) bool {
	for _, line := range l.lines {
		if strings.Contains(line, sub) {
			return true
		}
	}
	return false
}

// --- fixtures -------------------------------------------------------------

type testRow struct {
	Key   string
	Title string
	Value string
	Gone  bool
}

type testMapping struct {
	ExtID *string `json:"extId,omitempty"`
	Value *string `json:"value,omitempty"`
}

func rowKey(r *testRow) (string, bool) { return r.Key, r.Key != "" }

func populateRow(wi *WorkItem, r *testRow) error {
	if wi.ID == "" {
		wi.Attributes.Type = "task"
		wi.Attributes.Status = "open"
	}
	wi.Attributes.Title = r.Title
	m := &testMapping{}
	if r.Key != "" {
		k := r.Key
		m.ExtID = &k
	}
	if r.Value != "" {
		v := r.Value
		m.Value = &v
	}
	return SaveCustomFields(wi, m)
}

// existingItem builds a served work item with the given custom fields.
func existingItem(id, title string, fields map[string]interface{}) WorkItem {
	cf := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		cf[k] = v
	}
	return WorkItem{
		Type: "workitems",
		ID:   id,
		Attributes: &WorkItemAttributes{
			Type:         "task",
			Title:        title,
			Status:       "open",
			CustomFields: cf,
		},
	}
}

// newTestSync wires a Sync onto a fakeStore without going through the network.
func newTestSync(f *fakeStore, items []WorkItem, key KeyFunc) *Sync {
	index, _ := IndexWorkItems(items, key)
	return &Sync{
		store:    f,
		existing: index,
		staged:   make(map[string]*WorkItem),
		deleted:  make(map[string]bool),
	}
}

// --- KeyByFields / IndexWorkItems ----------------------------------------

func TestKeyByFields(t *testing.T) {
	wi := existingItem("MYPROJ/WI-1", "t", map[string]interface{}{
		"po":   "4500001234",
		"item": "10",
		"num":  42,
	})

	tests := []struct {
		name    string
		fields  []string
		want    string
		wantOK  bool
		useItem *WorkItem
	}{
		{name: "single", fields: []string{"po"}, want: "4500001234", wantOK: true},
		{name: "composite", fields: []string{"po", "item"}, want: "4500001234|10", wantOK: true},
		{name: "missing field", fields: []string{"po", "nope"}, wantOK: false},
		{name: "no fields", fields: nil, wantOK: false},
		{name: "non-string field", fields: []string{"num"}, wantOK: false},
		{name: "nil item", fields: []string{"po"}, wantOK: false, useItem: &WorkItem{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &wi
			if tt.useItem != nil {
				target = tt.useItem
			}
			got, ok := KeyByFields(tt.fields...)(target)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Fatalf("key = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIndexWorkItemsReportsDuplicates(t *testing.T) {
	items := []WorkItem{
		existingItem("MYPROJ/WI-1", "first", map[string]interface{}{"extId": "A"}),
		existingItem("MYPROJ/WI-2", "second", map[string]interface{}{"extId": "A"}),
		existingItem("MYPROJ/WI-3", "third", map[string]interface{}{"extId": "B"}),
		existingItem("MYPROJ/WI-4", "no key", nil),
	}

	index, dups := IndexWorkItems(items, KeyByFields("extId"))

	if len(index) != 2 {
		t.Fatalf("index size = %d, want 2", len(index))
	}
	if index["A"].ID != "MYPROJ/WI-1" {
		t.Fatalf("duplicate key: first item should win, got %s", index["A"].ID)
	}
	if len(dups) != 1 || dups[0] != "A" {
		t.Fatalf("dups = %v, want [A]", dups)
	}
}

// --- create / update / unchanged -----------------------------------------

func TestSyncCreatesNewRow(t *testing.T) {
	f := newFakeStore()
	s := newTestSync(f, nil, KeyByFields("extId"))

	Stage(s, []testRow{{Key: "A", Title: "New item", Value: "v1"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})

	res, err := s.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Created != 1 || res.Updated != 0 || res.Unchanged != 0 {
		t.Fatalf("result = %+v, want Created 1", res)
	}
	if len(f.created) != 1 || len(f.created[0]) != 1 {
		t.Fatalf("expected one create batch with one item, got %v", f.created)
	}
	if got := f.created[0][0].Attributes.CustomFields["value"]; got != "v1" {
		t.Fatalf("custom field value = %v, want v1", got)
	}
	if res.CreatedItems[0].ID == "" {
		t.Fatal("created item has no ID")
	}
	if s.Item("A").ID != res.CreatedItems[0].ID {
		t.Fatal("Item() should resolve the created item's new ID")
	}
}

func TestSyncUpdatesChangedRowAndLogsDiff(t *testing.T) {
	f := newFakeStore()
	log := &recordingLogger{}
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Old title", map[string]interface{}{
		"extId": "A", "value": "old",
	})}
	s := newTestSync(f, items, KeyByFields("extId"))
	s.Logger = log

	Stage(s, []testRow{{Key: "A", Title: "New title", Value: "new"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})

	res, err := s.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Updated != 1 || res.Created != 0 {
		t.Fatalf("result = %+v, want Updated 1", res)
	}
	if len(f.updated) != 1 || len(f.updated[0]) != 1 {
		t.Fatalf("expected one update batch with one pair, got %v", f.updated)
	}
	if f.updated[0][0].Original.ID != "MYPROJ/WI-1" {
		t.Fatalf("update pair carries the wrong original: %s", f.updated[0][0].Original.ID)
	}
	if !log.contains(`"value":"new"`) {
		t.Fatalf("expected the diff to be logged, got %v", log.lines)
	}
}

func TestSyncLeavesIdenticalRowUnchanged(t *testing.T) {
	f := newFakeStore()
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Same", map[string]interface{}{
		"extId": "A", "value": "v",
	})}
	s := newTestSync(f, items, KeyByFields("extId"))

	Stage(s, []testRow{{Key: "A", Title: "Same", Value: "v"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})

	res, err := s.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Unchanged != 1 || res.Updated != 0 || res.Created != 0 {
		t.Fatalf("result = %+v, want Unchanged 1", res)
	}
	if len(f.created) != 0 || len(f.updated) != 0 {
		t.Fatal("an unchanged item must not produce any API call")
	}
}

func TestSyncSkipsRowWithoutKey(t *testing.T) {
	f := newFakeStore()
	s := newTestSync(f, nil, KeyByFields("extId"))

	Stage(s, []testRow{{Key: "", Title: "No key"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})

	res, _ := s.Flush(context.Background())
	if res.Skipped != 1 || res.Created != 0 {
		t.Fatalf("result = %+v, want Skipped 1", res)
	}
}

// --- enrichment-only passes ----------------------------------------------

func TestSyncEnrichmentOnlyPass(t *testing.T) {
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Existing", map[string]interface{}{"extId": "A"})}

	t.Run("unmatched row is skipped", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, items, KeyByFields("extId"))
		Stage(s, []testRow{{Key: "ZZZ", Title: "Nope"}}, Pass[testRow]{
			Key: rowKey, Populate: populateRow, // Create defaults to false
		})
		res, _ := s.Flush(context.Background())
		if res.Skipped != 1 || res.Created != 0 || res.Updated != 0 {
			t.Fatalf("result = %+v, want Skipped 1", res)
		}
		if len(f.created) != 0 {
			t.Fatal("an enrichment-only pass must never create")
		}
	})

	t.Run("matched row updates", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, items, KeyByFields("extId"))
		Stage(s, []testRow{{Key: "A", Title: "Enriched", Value: "v"}}, Pass[testRow]{
			Key: rowKey, Populate: populateRow,
		})
		res, _ := s.Flush(context.Background())
		if res.Updated != 1 {
			t.Fatalf("result = %+v, want Updated 1", res)
		}
	})
}

// --- multi-pass staging ---------------------------------------------------

// secondRow is a different row type, to prove passes can carry different shapes.
type secondRow struct {
	Key   string
	Extra string
}

type secondMapping struct {
	Extra *string `json:"extra,omitempty"`
}

func populateSecond(wi *WorkItem, r *secondRow) error {
	e := r.Extra
	return SaveCustomFields(wi, &secondMapping{Extra: &e})
}

func TestSyncSecondPassEnrichesItemStagedForCreation(t *testing.T) {
	f := newFakeStore()
	s := newTestSync(f, nil, KeyByFields("extId"))

	Stage(s, []testRow{{Key: "A", Title: "Created here", Value: "v1"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})
	Stage(s, []secondRow{{Key: "A", Extra: "enriched"}}, Pass[secondRow]{
		Key:      func(r *secondRow) (string, bool) { return r.Key, r.Key != "" },
		Populate: populateSecond, // Create false: must still find the staged item
	})

	res, err := s.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Created != 1 {
		t.Fatalf("result = %+v, want exactly one item created", res)
	}
	if len(f.created[0]) != 1 {
		t.Fatalf("both passes must build one item, got %d", len(f.created[0]))
	}
	cf := f.created[0][0].Attributes.CustomFields
	if cf["value"] != "v1" || cf["extra"] != "enriched" {
		t.Fatalf("created item missing a pass's fields: %v", cf)
	}
}

func TestSyncSecondPassUpdatesOtherwiseUnchangedItem(t *testing.T) {
	f := newFakeStore()
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Same", map[string]interface{}{
		"extId": "A", "value": "v",
	})}
	s := newTestSync(f, items, KeyByFields("extId"))

	// Pass 1 changes nothing.
	Stage(s, []testRow{{Key: "A", Title: "Same", Value: "v"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})
	// Pass 2 adds a field, so the item must now be updated.
	Stage(s, []secondRow{{Key: "A", Extra: "added"}}, Pass[secondRow]{
		Key:      func(r *secondRow) (string, bool) { return r.Key, r.Key != "" },
		Populate: populateSecond,
	})

	res, _ := s.Flush(context.Background())
	if res.Updated != 1 || res.Unchanged != 0 {
		t.Fatalf("result = %+v, want Updated 1", res)
	}
}

// --- deletion -------------------------------------------------------------

func TestSyncDelete(t *testing.T) {
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Doomed", map[string]interface{}{"extId": "A"})}
	deletePass := Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
		Delete: func(r *testRow) bool { return r.Gone },
	}

	t.Run("existing item is deleted", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, items, KeyByFields("extId"))
		Stage(s, []testRow{{Key: "A", Title: "Doomed", Gone: true}}, deletePass)
		res, err := s.Flush(context.Background())
		if err != nil {
			t.Fatalf("Flush: %v", err)
		}
		if res.Deleted != 1 || len(f.deleted) != 1 || f.deleted[0][0] != "MYPROJ/WI-1" {
			t.Fatalf("result = %+v, deleted = %v", res, f.deleted)
		}
		if _, ok := s.Items()["A"]; ok {
			t.Fatal("a deleted key must not appear in Items()")
		}
	})

	t.Run("unmatched delete is skipped", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, items, KeyByFields("extId"))
		Stage(s, []testRow{{Key: "ZZZ", Title: "Ghost", Gone: true}}, deletePass)
		res, _ := s.Flush(context.Background())
		if res.Skipped != 1 || res.Deleted != 0 {
			t.Fatalf("result = %+v, want Skipped 1", res)
		}
	})

	t.Run("staged then deleted is neither created nor updated", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, items, KeyByFields("extId"))
		Stage(s, []testRow{
			{Key: "A", Title: "Changed first", Value: "v"},
			{Key: "A", Title: "Doomed", Gone: true},
		}, deletePass)
		res, _ := s.Flush(context.Background())
		if res.Deleted != 1 || res.Updated != 0 || res.Created != 0 {
			t.Fatalf("result = %+v, want only Deleted 1", res)
		}
		if len(f.updated) != 0 {
			t.Fatal("a deleted item must not also be updated")
		}
	})

	t.Run("later row cannot resurrect a deleted key", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, items, KeyByFields("extId"))
		Stage(s, []testRow{
			{Key: "A", Title: "Doomed", Gone: true},
			{Key: "A", Title: "Back again", Value: "v"},
		}, deletePass)
		res, _ := s.Flush(context.Background())
		if res.Deleted != 1 || res.Updated != 0 {
			t.Fatalf("result = %+v, want only Deleted 1", res)
		}
	})
}

// --- populate errors ------------------------------------------------------

func TestSyncPopulateError(t *testing.T) {
	boom := errors.New("boom")

	t.Run("on a new row nothing is staged", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, nil, KeyByFields("extId"))
		Stage(s, []testRow{{Key: "A", Title: "x"}}, Pass[testRow]{
			Key:      rowKey,
			Populate: func(*WorkItem, *testRow) error { return boom },
			Create:   true,
		})
		res, _ := s.Flush(context.Background())
		if res.Errors != 1 || res.Created != 0 {
			t.Fatalf("result = %+v, want Errors 1", res)
		}
		if s.Item("A") != nil {
			t.Fatal("a failed first populate must leave nothing staged")
		}
	})

	t.Run("keeps an earlier pass's work", func(t *testing.T) {
		f := newFakeStore()
		s := newTestSync(f, nil, KeyByFields("extId"))
		Stage(s, []testRow{{Key: "A", Title: "Good", Value: "v1"}}, Pass[testRow]{
			Key: rowKey, Populate: populateRow, Create: true,
		})
		Stage(s, []secondRow{{Key: "A", Extra: "x"}}, Pass[secondRow]{
			Key:      func(r *secondRow) (string, bool) { return r.Key, true },
			Populate: func(*WorkItem, *secondRow) error { return boom },
		})
		res, _ := s.Flush(context.Background())
		if res.Errors != 1 || res.Created != 1 {
			t.Fatalf("result = %+v, want Errors 1 and Created 1", res)
		}
		if f.created[0][0].Attributes.CustomFields["value"] != "v1" {
			t.Fatal("the first pass's fields must survive a later failure")
		}
	})
}

// --- field preservation ---------------------------------------------------

func TestSyncPreservesFieldsTheMappingDoesNotDeclare(t *testing.T) {
	f := newFakeStore()
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Item", map[string]interface{}{
		"extId":      "A",
		"value":      "v",
		"ownedByYou": "do not touch",
	})}
	s := newTestSync(f, items, KeyByFields("extId"))

	// The mapping declares extId and value only.
	Stage(s, []testRow{{Key: "A", Title: "Item", Value: "v"}}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
	})

	res, _ := s.Flush(context.Background())
	if res.Unchanged != 1 {
		t.Fatalf("result = %+v, want Unchanged 1 — a foreign field must not read as a change", res)
	}
	if got := s.Item("A").Attributes.CustomFields["ownedByYou"]; got != "do not touch" {
		t.Fatalf("foreign field was lost: %v", got)
	}
}

// --- relationship-only change --------------------------------------------

func TestSyncDetectsRelationshipOnlyChange(t *testing.T) {
	f := newFakeStore()
	items := []WorkItem{existingItem("MYPROJ/WI-1", "Item", map[string]interface{}{"extId": "A"})}
	s := newTestSync(f, items, KeyByFields("extId"))
	log := &recordingLogger{}
	s.Logger = log

	// Populate changes only a user reference, which lives in relationships.
	Stage(s, []testRow{{Key: "A", Title: "Item"}}, Pass[testRow]{
		Key: rowKey, Create: true,
		Populate: func(wi *WorkItem, r *testRow) error {
			wi.Attributes.Title = r.Title
			wi.SetUserReferenceField("reviewer", "jdoe")
			return nil
		},
	})

	res, _ := s.Flush(context.Background())
	if res.Updated != 1 {
		t.Fatalf("result = %+v, want Updated 1 (Equals covers relationships)", res)
	}
	if !log.contains("relationships changed") {
		t.Fatalf("expected the relationship-only log line, got %v", log.lines)
	}
}

// --- determinism ----------------------------------------------------------

func TestSyncFlushesInStagingOrder(t *testing.T) {
	keys := []string{"K1", "K2", "K3", "K4", "K5", "K6", "K7", "K8"}

	run := func() []string {
		f := newFakeStore()
		s := newTestSync(f, nil, KeyByFields("extId"))
		rows := make([]testRow, 0, len(keys))
		for _, k := range keys {
			rows = append(rows, testRow{Key: k, Title: "T-" + k})
		}
		Stage(s, rows, Pass[testRow]{Key: rowKey, Populate: populateRow, Create: true})
		if _, err := s.Flush(context.Background()); err != nil {
			t.Fatalf("Flush: %v", err)
		}
		var got []string
		for _, wi := range f.created[0] {
			got = append(got, wi.Attributes.CustomFields["extId"].(string))
		}
		return got
	}

	first := run()
	for i, k := range keys {
		if first[i] != k {
			t.Fatalf("create batch order = %v, want staging order %v", first, keys)
		}
	}
	// Repeat: map iteration order varies per run, the batch must not.
	for i := 0; i < 5; i++ {
		if got := run(); strings.Join(got, ",") != strings.Join(first, ",") {
			t.Fatalf("run %d order = %v, want %v", i, got, first)
		}
	}
}

// --- dry run --------------------------------------------------------------

func TestSyncDryRunMakesNoCalls(t *testing.T) {
	f := newFakeStore()
	items := []WorkItem{
		existingItem("MYPROJ/WI-1", "Old", map[string]interface{}{"extId": "A", "value": "old"}),
		existingItem("MYPROJ/WI-2", "Doomed", map[string]interface{}{"extId": "B"}),
	}
	s := newTestSync(f, items, KeyByFields("extId"))
	s.DryRun = true

	Stage(s, []testRow{
		{Key: "A", Title: "New", Value: "new"},
		{Key: "B", Title: "Doomed", Gone: true},
		{Key: "C", Title: "Fresh"},
	}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
		Delete: func(r *testRow) bool { return r.Gone },
	})

	res, err := s.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Created != 1 || res.Updated != 1 || res.Deleted != 1 {
		t.Fatalf("result = %+v, want 1 create, 1 update, 1 delete projected", res)
	}
	if len(f.created)+len(f.updated)+len(f.deleted) != 0 {
		t.Fatal("dry run must not call the API")
	}
	if len(res.CreatedItems) != 0 {
		t.Fatal("dry run must not report created items")
	}
}

// --- create edge cases ----------------------------------------------------

func TestSyncRejectsNewItemWithoutTitle(t *testing.T) {
	f := newFakeStore()
	s := newTestSync(f, nil, KeyByFields("extId"))

	Stage(s, []testRow{
		{Key: "A", Title: ""},
		{Key: "B", Title: "Fine"},
	}, Pass[testRow]{Key: rowKey, Populate: populateRow, Create: true})

	res, err := s.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Errors != 1 || res.Created != 1 {
		t.Fatalf("result = %+v, want Errors 1 and Created 1", res)
	}
	if len(f.created[0]) != 1 {
		t.Fatal("the untitled item must be kept out of the batch")
	}
}

func TestSyncCountsPartialCreateFailure(t *testing.T) {
	f := newFakeStore()
	f.createIDsUpTo = 1 // only the first item gets an ID
	f.createErr = errors.New("second batch failed")
	s := newTestSync(f, nil, KeyByFields("extId"))

	Stage(s, []testRow{
		{Key: "A", Title: "First"},
		{Key: "B", Title: "Second"},
		{Key: "C", Title: "Third"},
	}, Pass[testRow]{Key: rowKey, Populate: populateRow, Create: true})

	res, err := s.Flush(context.Background())
	if err == nil {
		t.Fatal("expected the create error to be returned")
	}
	if res.Created != 1 || res.Errors != 2 {
		t.Fatalf("result = %+v, want Created 1 and Errors 2", res)
	}
}

func TestSyncCountsUpdateAndDeleteFailures(t *testing.T) {
	f := newFakeStore()
	f.updateErr = errors.New("patch failed")
	f.deleteErr = errors.New("delete failed")
	items := []WorkItem{
		existingItem("MYPROJ/WI-1", "Old", map[string]interface{}{"extId": "A"}),
		existingItem("MYPROJ/WI-2", "Doomed", map[string]interface{}{"extId": "B"}),
	}
	s := newTestSync(f, items, KeyByFields("extId"))

	Stage(s, []testRow{
		{Key: "A", Title: "Changed"},
		{Key: "B", Title: "Doomed", Gone: true},
	}, Pass[testRow]{
		Key: rowKey, Populate: populateRow, Create: true,
		Delete: func(r *testRow) bool { return r.Gone },
	})

	res, err := s.Flush(context.Background())
	if err == nil {
		t.Fatal("expected the batch errors to be returned")
	}
	if res.Updated != 0 || res.Deleted != 0 || res.Errors != 2 {
		t.Fatalf("result = %+v, want no successes and Errors 2", res)
	}
}

// --- end to end over HTTP -------------------------------------------------

func TestSyncEndToEnd(t *testing.T) {
	var patched []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/v1/projects/MYPROJ/workitems") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"data": [
					{"type":"workitems","id":"MYPROJ/WI-1",
					 "attributes":{"type":"task","title":"Existing","status":"open",
					               "extId":"A","value":"old","foreign":"keep me"}}
				],
				"meta": {"totalCount": 1}
			}`))
		case http.MethodPost:
			var body struct {
				Data []map[string]interface{} `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode POST: %v", err)
			}
			out := make([]map[string]interface{}, 0, len(body.Data))
			for i := range body.Data {
				out = append(out, map[string]interface{}{
					"type": "workitems",
					"id":   fmt.Sprintf("MYPROJ/WI-%d", 100+i),
				})
			}
			resp, _ := json.Marshal(map[string]interface{}{"data": out})
			_, _ = w.Write(resp)
		case http.MethodPatch:
			var body struct {
				Data []map[string]interface{} `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode PATCH: %v", err)
			}
			patched = append(patched, body.Data...)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := New(server.URL+"/rest/v1", "token",
		WithHTTPClient(server.Client()),
		WithRetryConfig(RetryConfig{MaxRetries: 0}))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	project := client.Project("MYPROJ")
	ctx := context.Background()

	s, err := NewSync(ctx, project.WorkItems, "type:task", KeyByFields("extId"))
	if err != nil {
		t.Fatalf("NewSync: %v", err)
	}

	Stage(s, []testRow{
		{Key: "A", Title: "Existing", Value: "new"},
		{Key: "B", Title: "Brand new", Value: "v"},
	}, Pass[testRow]{Key: rowKey, Populate: populateRow, Create: true})

	res, err := s.Flush(ctx)
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if res.Created != 1 || res.Updated != 1 {
		t.Fatalf("result = %+v, want 1 created and 1 updated", res)
	}
	if res.CreatedItems[0].ID != "MYPROJ/WI-100" {
		t.Fatalf("created ID = %q, want MYPROJ/WI-100", res.CreatedItems[0].ID)
	}

	if len(patched) != 1 {
		t.Fatalf("expected one patched item, got %d", len(patched))
	}
	attrs, _ := patched[0]["attributes"].(map[string]interface{})
	if attrs["value"] != "new" {
		t.Fatalf("PATCH should carry the changed field, got %v", attrs)
	}
	if _, sent := attrs["foreign"]; sent {
		t.Fatalf("PATCH must not resend the untouched foreign field: %v", attrs)
	}
}
