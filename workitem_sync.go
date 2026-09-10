// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// SyncLogger is the printf-style logger the sync engine reports progress to.
// *zap.SugaredLogger satisfies it without an adapter, as does any logger with
// the same four methods. A nil Sync.Logger disables logging entirely.
type SyncLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// SyncKeySeparator joins the parts of a composite key built by KeyByFields.
const SyncKeySeparator = "|"

// KeyFunc derives the sync key of an existing Polarion work item.
// Returning ok=false means the item cannot be indexed and is ignored, so a row
// carrying that key is treated as new.
type KeyFunc func(wi *WorkItem) (key string, ok bool)

// KeyByFields returns a KeyFunc reading the given string custom fields and
// joining them with SyncKeySeparator. Every field must be present and non-empty,
// otherwise the item is not indexed.
//
// Example:
//
//	polarion.KeyByFields("logwisId")                              // "4711"
//	polarion.KeyByFields("purchaseOrderNumber", "orderItem")      // "4500001234|10"
func KeyByFields(fields ...string) KeyFunc {
	return func(wi *WorkItem) (string, bool) {
		if len(fields) == 0 || wi == nil || wi.Attributes == nil || wi.Attributes.CustomFields == nil {
			return "", false
		}
		cf := CustomFields(wi.Attributes.CustomFields)
		parts := make([]string, 0, len(fields))
		for _, f := range fields {
			v, ok := cf.GetString(f)
			if !ok || v == "" {
				return "", false
			}
			parts = append(parts, v)
		}
		return strings.Join(parts, SyncKeySeparator), true
	}
}

// IndexWorkItems indexes items by the key derived from each item. Items whose
// KeyFunc returns ok=false are omitted. On a duplicate key the first item wins
// and the key is reported in dups, so the caller can flag ambiguous data.
func IndexWorkItems(items []WorkItem, key KeyFunc) (index map[string]*WorkItem, dups []string) {
	index = make(map[string]*WorkItem, len(items))
	for i := range items {
		item := &items[i]
		k, ok := key(item)
		if !ok {
			continue
		}
		if _, exists := index[k]; exists {
			dups = append(dups, k)
			continue
		}
		index[k] = item
	}
	return index, dups
}

// SyncResult reports what Flush did, or in DryRun mode what it would have done.
//
// Created, Updated, Unchanged and Deleted count work items. Skipped and Errors
// count rows: a row is skipped when its Key returns ok=false, when a pass with
// Create=false finds no matching item, or when a Delete row matches nothing.
// The distinction matters in a multi-pass sync, where several rows can
// contribute to a single work item.
type SyncResult struct {
	Created   int
	Updated   int
	Unchanged int
	Deleted   int
	Skipped   int
	Errors    int

	// CreatedItems holds the work items Create accepted, each carrying the ID
	// Polarion assigned. Empty in DryRun mode.
	CreatedItems []*WorkItem
	// UpdatedItems holds the staged items that were sent as updates.
	UpdatedItems []*WorkItem
	// DeletedIDs holds the IDs that were deleted.
	DeletedIDs []string
}

// syncStore is the subset of *WorkItemService the sync engine needs. It exists
// so tests can substitute a recording fake; *WorkItemService satisfies it.
type syncStore interface {
	Create(ctx context.Context, items ...*WorkItem) error
	UpdateBatchWithOldValues(ctx context.Context, pairs ...UpdatePair) error
	Delete(ctx context.Context, ids ...string) error
	Equals(a, b *WorkItem) bool
	EqualsWithDiff(a, b *WorkItem) *WorkItemAttributes
}

// Sync is a stage-then-flush upsert session for one work item type.
//
// The lifecycle is:
//
//  1. NewSync (or NewSyncFromItems) queries Polarion once and indexes the result
//     by sync key.
//  2. Stage applies one or more passes of source rows. Passes may carry different
//     row types, and a later pass sees the items staged by earlier ones —
//     including items staged for creation — so several sources can build up a
//     single work item before anything is written.
//  3. Flush diffs each staged item once against what Polarion served, then issues
//     one batched Create, one batched update and the deletes, in staging order.
//
// A Sync is not safe for concurrent use.
type Sync struct {
	// Logger receives progress, one line with a JSON diff per changed item, and
	// errors. Leaving it nil disables logging.
	Logger SyncLogger
	// DryRun logs every would-be create, update and delete and returns the
	// projected SyncResult without calling Polarion.
	DryRun bool

	store    syncStore
	existing map[string]*WorkItem
	staged   map[string]*WorkItem
	order    []string
	deletes  []string
	deleted  map[string]bool
	rows     SyncResult
}

// NewSyncFromItems starts a session over work items the caller already queried.
// Use it when the query needs options the engine does not set, for example
// WithInclude("linkedWorkItems") so link caches can be built from the served
// items (Clone does not copy inline links).
func NewSyncFromItems(svc *WorkItemService, items []WorkItem, key KeyFunc) *Sync {
	index, dups := IndexWorkItems(items, key)
	s := &Sync{
		store:    svc,
		existing: index,
		staged:   make(map[string]*WorkItem),
		deleted:  make(map[string]bool),
	}
	for _, d := range dups {
		s.warnf("sync: duplicate key %q in Polarion; keeping the first item", d)
	}
	return s
}

// NewSync queries Polarion with svc.QueryAll and starts a session over the result.
//
// Example:
//
//	s, err := polarion.NewSync(ctx, project.WorkItems,
//	    "type:scopeItem AND NOT HAS_VALUE:resolution",
//	    polarion.KeyByFields("bridgeItemId"))
func NewSync(ctx context.Context, svc *WorkItemService, query string, key KeyFunc, opts ...QueryOption) (*Sync, error) {
	items, err := svc.QueryAll(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("sync: query %q: %w", query, err)
	}
	return NewSyncFromItems(svc, items, key), nil
}

// Pass describes one source of rows applied to a Sync.
type Pass[Row any] struct {
	// Key returns the sync key of a row. Returning ok=false skips the row and
	// counts it in SyncResult.Skipped.
	Key func(row *Row) (key string, ok bool)

	// Populate maps one row onto wi and should end with SaveCustomFields(wi, m),
	// where m is a struct whose json tags are the Polarion custom field IDs.
	//
	// For a new item wi.ID is empty and wi.Attributes is allocated with an empty
	// CustomFields map; set Attributes.Type, Status and Title. For an existing
	// item wi is a Clone carrying every custom field Polarion served, so assign
	// only the fields this source owns — SaveCustomFields leaves keys no struct
	// tag names untouched.
	//
	// Note that a nil mapping field does not clear the Polarion value on update:
	// the update diff only considers keys present in the staged item.
	Populate func(wi *WorkItem, row *Row) error

	// Create allows staging new work items for rows with no existing match.
	// Leaving it false makes the pass enrichment-only: unmatched rows are skipped.
	Create bool

	// Delete, when non-nil and returning true for a row, marks the matching
	// existing item for deletion instead of populating it. Anything staged for
	// that key is discarded and later passes will not resurrect it.
	Delete func(row *Row) bool
}

// Stage applies a pass to rows, in order.
//
// Populate errors are logged and counted in SyncResult.Errors. A row that fails
// while being staged for the first time leaves nothing behind; a row that fails
// on an item an earlier pass staged leaves that earlier state intact.
func Stage[Row any](s *Sync, rows []Row, p Pass[Row]) {
	if s == nil || p.Key == nil || p.Populate == nil {
		panic("polarion: Stage requires a Sync with Pass.Key and Pass.Populate set")
	}

	for i := range rows {
		row := &rows[i]

		key, ok := p.Key(row)
		if !ok || key == "" {
			s.rows.Skipped++
			continue
		}

		if s.deleted[key] {
			// An earlier row already marked this item for deletion.
			s.rows.Skipped++
			continue
		}

		if p.Delete != nil && p.Delete(row) {
			existing, exists := s.existing[key]
			if !exists {
				s.rows.Skipped++
				continue
			}
			s.markDeleted(key, existing.ID)
			continue
		}

		wi, fresh := s.stagedFor(key, p.Create)
		if wi == nil {
			// No existing item and this pass may not create one.
			s.rows.Skipped++
			continue
		}

		if err := p.Populate(wi, row); err != nil {
			s.rows.Errors++
			s.errorf("sync: populate %q: %v", key, err)
			if !fresh {
				// An earlier pass staged this item; keep what it built.
				continue
			}
			s.unstage(key)
			continue
		}

		s.commitStaged(key, wi)
	}
}

// stagedFor returns the work item to populate for key: the already staged one,
// a clone of what Polarion served, or a brand new item when create is allowed.
// It returns nil when there is nothing to populate. The item is not recorded in
// the staging order until commitStaged; fresh reports that this call built it,
// so a failing Populate can discard it without losing an earlier pass's work.
func (s *Sync) stagedFor(key string, create bool) (wi *WorkItem, fresh bool) {
	if staged, ok := s.staged[key]; ok {
		return staged, false
	}
	if existing, ok := s.existing[key]; ok {
		return existing.Clone(), true
	}
	if !create {
		return nil, false
	}
	return &WorkItem{
		Type: "workitems",
		Attributes: &WorkItemAttributes{
			CustomFields: make(map[string]interface{}),
		},
	}, true
}

// commitStaged records wi under key, appending to the flush order the first time.
func (s *Sync) commitStaged(key string, wi *WorkItem) {
	if _, ok := s.staged[key]; !ok {
		s.order = append(s.order, key)
	}
	s.staged[key] = wi
}

// unstage removes a key staged by the current call only.
func (s *Sync) unstage(key string) {
	if _, ok := s.staged[key]; !ok {
		return
	}
	delete(s.staged, key)
	for i, k := range s.order {
		if k == key {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
}

// markDeleted queues an existing item for deletion and drops anything staged.
func (s *Sync) markDeleted(key, id string) {
	s.unstage(key)
	s.deleted[key] = true
	s.deletes = append(s.deletes, id)
}

// Item returns the work item for key: the staged one if the key was staged,
// otherwise the one Polarion served, otherwise nil. After Flush, items that were
// created carry their new ID, which makes this the way to resolve IDs when
// linking work items after a sync.
func (s *Sync) Item(key string) *WorkItem {
	if wi, ok := s.staged[key]; ok {
		return wi
	}
	if wi, ok := s.existing[key]; ok {
		return wi
	}
	return nil
}

// Items returns the staged items overlaid on the ones Polarion served, keyed by
// sync key. Deleted keys are omitted. The map is a fresh copy; the work items
// are not.
func (s *Sync) Items() map[string]*WorkItem {
	out := make(map[string]*WorkItem, len(s.existing)+len(s.staged))
	for k, wi := range s.existing {
		if !s.deleted[k] {
			out[k] = wi
		}
	}
	for k, wi := range s.staged {
		out[k] = wi
	}
	return out
}

// Flush diffs every staged item against what Polarion served and writes the
// changes: one batched Create, one batched update and the deletes, all in
// staging order so repeated runs produce identical requests.
//
// The returned SyncResult is valid whether or not an error is returned; the
// error joins the failures of the individual batch calls.
func (s *Sync) Flush(ctx context.Context) (SyncResult, error) {
	res := s.rows
	res.CreatedItems = nil
	res.UpdatedItems = nil
	res.DeletedIDs = nil

	var toCreate []*WorkItem
	var toUpdate []UpdatePair

	for _, key := range s.order {
		staged := s.staged[key]
		if staged == nil {
			continue
		}
		original, exists := s.existing[key]

		if !exists {
			// Create validates every item up front, so one item without a title
			// would fail the whole batch. Reject it here instead.
			if staged.Attributes == nil || staged.Attributes.Title == "" {
				res.Errors++
				s.errorf("sync: not creating %q: title is required", key)
				continue
			}
			toCreate = append(toCreate, staged)
			continue
		}

		if s.store.Equals(original, staged) {
			res.Unchanged++
			continue
		}

		// Equals also covers custom relationships; EqualsWithDiff reports only
		// attribute changes, so a nil diff here means relationships changed.
		if diff := s.store.EqualsWithDiff(original, staged); diff != nil {
			if encoded, err := json.Marshal(diff); err == nil {
				s.infof("sync: %s (%s) changed: %s", original.ID, key, encoded)
			} else {
				s.infof("sync: %s (%s) changed", original.ID, key)
			}
		} else {
			s.infof("sync: %s (%s) relationships changed", original.ID, key)
		}
		toUpdate = append(toUpdate, UpdatePair{Original: original, Updated: staged})
	}

	if s.DryRun {
		res.Created = len(toCreate)
		res.Updated = len(toUpdate)
		res.Deleted = len(s.deletes)
		s.infof("sync: dry run — would create %d, update %d, delete %d (unchanged %d, skipped %d, errors %d)",
			res.Created, res.Updated, res.Deleted, res.Unchanged, res.Skipped, res.Errors)
		return res, nil
	}

	var errs []error

	if len(toCreate) > 0 {
		s.infof("sync: creating %d work items", len(toCreate))
		if err := s.store.Create(ctx, toCreate...); err != nil {
			errs = append(errs, fmt.Errorf("create: %w", err))
			s.errorf("sync: create failed: %v", err)
		}
		// Create assigns IDs to the items it sent. Anything still without one was
		// dropped as oversize or belonged to a batch that never ran.
		for _, wi := range toCreate {
			if wi.ID != "" {
				res.Created++
				res.CreatedItems = append(res.CreatedItems, wi)
				continue
			}
			res.Errors++
			s.warnf("sync: work item %q was not created", wi.Attributes.Title)
		}
	}

	if len(toUpdate) > 0 {
		s.infof("sync: updating %d work items", len(toUpdate))
		if err := s.store.UpdateBatchWithOldValues(ctx, toUpdate...); err != nil {
			errs = append(errs, fmt.Errorf("update: %w", err))
			s.errorf("sync: update failed: %v", err)
			res.Errors += len(toUpdate)
		} else {
			res.Updated = len(toUpdate)
			for _, pair := range toUpdate {
				res.UpdatedItems = append(res.UpdatedItems, pair.Updated)
			}
		}
	}

	if len(s.deletes) > 0 {
		s.infof("sync: deleting %d work items", len(s.deletes))
		if err := s.store.Delete(ctx, s.deletes...); err != nil {
			errs = append(errs, fmt.Errorf("delete: %w", err))
			s.errorf("sync: delete failed: %v", err)
			res.Errors += len(s.deletes)
		} else {
			res.Deleted = len(s.deletes)
			res.DeletedIDs = append(res.DeletedIDs, s.deletes...)
		}
	}

	return res, errors.Join(errs...)
}

func (s *Sync) infof(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Infof(format, args...)
	}
}

func (s *Sync) warnf(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Warnf(format, args...)
	}
}

func (s *Sync) errorf(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Errorf(format, args...)
	}
}
