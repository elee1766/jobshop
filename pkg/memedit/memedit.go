// Package memedit stores memory edits (#274, #592) on the memory_edits
// store collection (#579). An edit is a merged curator PR or a local
// memory write; the curator gate decides later whether it helped.
//
// Records are keyed by the edit id as a decimal string, not by PR: local
// writes have no PR, and the RPC, the CLI and learnings' "memory:<id>"
// refs name edits by id. The PR is a unique field, so recording the same
// PR twice returns the first edit.
package memedit

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Spec is the memory_edits collection.
var Spec = store.CollectionSpec{
	Name:   "memory_edits",
	Unique: []store.Field{{Name: "pr", Path: "pr"}},
}

// Store is the part of the store memory edits need.
type Store interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

// value is the record's JSON. Times are unix ms, as the old table kept
// them; created_at is the record's.
type value struct {
	ID          int64              `json:"id"`
	PR          string             `json:"pr"`
	Path        string             `json:"path"`
	Before      string             `json:"before"`
	Existed     bool               `json:"existed"`
	Learnings   []int64            `json:"learnings"`
	Pools       []string           `json:"pools"`
	MergedAt    int64              `json:"merged_at"`
	Baseline    domain.PoolMetrics `json:"baseline"`
	After       domain.PoolMetrics `json:"after"`
	Reason      string             `json:"reason"`
	EvaluatedAt int64              `json:"evaluated_at"`
}

// maxTries bounds the id and compare-and-set loops under races.
const maxTries = 10

// AddPR records a merged curator PR. A PR already recorded returns its
// edit unchanged, so a repeated record is harmless.
func AddPR(ctx context.Context, st Store, e domain.MemoryEdit) (domain.MemoryEdit, error) {
	e.PR = strings.TrimSpace(e.PR)
	if e.PR == "" {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit needs a pr", domain.ErrInvalid)
	}
	if e.MergedAt.IsZero() {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit needs merged_at", domain.ErrInvalid)
	}
	e.Path, e.Before, e.Existed = "", "", false
	return add(ctx, st, e)
}

// AddWrite records a local memory write: Path, Before, Existed,
// Learnings, Pools, Baseline, and MergedAt as the write time. Every
// write is a new edit, so the gate can restore each one's before text.
func AddWrite(ctx context.Context, st Store, e domain.MemoryEdit) (domain.MemoryEdit, error) {
	e.Path = strings.TrimSpace(e.Path)
	if e.Path == "" {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory write needs a path", domain.ErrInvalid)
	}
	if e.MergedAt.IsZero() {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory write needs merged_at", domain.ErrInvalid)
	}
	e.PR = ""
	return add(ctx, st, e)
}

// newest is how many of the newest records add reads for the highest
// id: records created in the same millisecond tie on created_at and
// key order is text order, so the newest one alone may not hold it.
const newest = 20

// add inserts e under the next free id, one past the highest of the
// newest records. A put that finds its id taken lost a race and tries
// the next.
func add(ctx context.Context, st Store, e domain.MemoryEdit) (domain.MemoryEdit, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	last, err := col.List(ctx, store.RecordFilter{Order: store.OrderCreatedDesc, Limit: newest})
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	id := int64(1)
	for _, r := range last {
		if n, err := strconv.ParseInt(r.Key, 10, 64); err == nil && n >= id {
			id = n + 1
		}
	}
	for range maxTries {
		e.ID = id
		v, err := json.Marshal(toValue(e))
		if err != nil {
			return domain.MemoryEdit{}, err
		}
		r, created, err := col.Put(ctx, store.Record{Key: strconv.FormatInt(id, 10), State: domain.EditPending, Value: v})
		if err != nil {
			return domain.MemoryEdit{}, err
		}
		out, err := fromRecord(r)
		if err != nil || created {
			return out, err
		}
		if e.PR != "" && out.PR == e.PR {
			return out, nil // already recorded
		}
		id++ // id taken by a concurrent add
	}
	return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit: lost every id race", domain.ErrConflict)
}

// Get returns one edit by id.
func Get(ctx context.Context, st Store, id int64) (domain.MemoryEdit, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	r, err := col.Get(ctx, strconv.FormatInt(id, 10))
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	return fromRecord(r)
}

// List returns the n newest edits (0 means store.DefaultLimit), newest
// merge first; empty state is all.
func List(ctx context.Context, st Store, state string, n int) ([]domain.MemoryEdit, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return nil, err
	}
	f := store.RecordFilter{Order: store.OrderCreatedDesc, Limit: n}
	if state != "" {
		f.States = []string{state}
	}
	rs, err := col.List(ctx, f)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MemoryEdit, 0, len(rs))
	for _, r := range rs {
		e, err := fromRecord(r)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	slices.SortStableFunc(out, func(a, b domain.MemoryEdit) int {
		return cmp.Or(b.MergedAt.Compare(a.MergedAt), cmp.Compare(b.ID, a.ID))
	})
	return out, nil
}

// SetResult stores the gate's decision on an edit, evaluated at now.
func SetResult(ctx context.Context, st Store, id int64, state string, after domain.PoolMetrics, reason string, now time.Time) (domain.MemoryEdit, error) {
	if state != domain.EditHelped && state != domain.EditNotHelped {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit state %q", domain.ErrInvalid, state)
	}
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	key := strconv.FormatInt(id, 10)
	for range maxTries {
		r, err := col.Get(ctx, key)
		if err != nil {
			return domain.MemoryEdit{}, err
		}
		e, err := fromRecord(r)
		if err != nil {
			return e, err
		}
		e.State, e.After, e.Reason, e.EvaluatedAt = state, after, reason, now
		if r.Value, err = json.Marshal(toValue(e)); err != nil {
			return domain.MemoryEdit{}, err
		}
		r.State = state
		r, err = col.Update(ctx, r)
		if errors.Is(err, domain.ErrConflict) {
			continue
		}
		if err != nil {
			return domain.MemoryEdit{}, err
		}
		return fromRecord(r)
	}
	return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit %d: lost every race", domain.ErrConflict, id)
}

// Ref names the edit in a cited learning's gate: the PR, or the edit id
// and path of a local write.
func Ref(e domain.MemoryEdit) string {
	if e.PR != "" {
		return e.PR
	}
	return fmt.Sprintf("memory edit %d, %s", e.ID, e.Path)
}

func toValue(e domain.MemoryEdit) value {
	return value{
		ID: e.ID, PR: e.PR, Path: e.Path, Before: e.Before, Existed: e.Existed,
		Learnings: nonNil(e.Learnings), Pools: nonNil(e.Pools), MergedAt: ms(e.MergedAt),
		Baseline: e.Baseline, After: e.After, Reason: e.Reason, EvaluatedAt: ms(e.EvaluatedAt),
	}
}

func fromRecord(r store.Record) (domain.MemoryEdit, error) {
	var v value
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return domain.MemoryEdit{}, fmt.Errorf("memory edit %s: %w", r.Key, err)
	}
	return domain.MemoryEdit{
		ID: v.ID, PR: v.PR, Path: v.Path, Before: v.Before, Existed: v.Existed,
		Learnings: nonNil(v.Learnings), Pools: nonNil(v.Pools), MergedAt: fromMS(v.MergedAt),
		Baseline: v.Baseline, State: r.State, After: v.After, Reason: v.Reason,
		EvaluatedAt: fromMS(v.EvaluatedAt), CreatedAt: r.CreatedAt,
	}, nil
}

func ms(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

func fromMS(n int64) time.Time {
	if n == 0 {
		return time.Time{}
	}
	return time.UnixMilli(n).UTC()
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
