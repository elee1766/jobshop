package knowledge

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Memory edits (#274, #592) live on the memory_edits store collection
// (#579). An edit is a merged curator PR or a local memory write; the
// curator gate decides later whether it helped.
//
// Records are keyed by the edit id as a decimal string, not by PR: local
// writes have no PR, and the RPC, the CLI and learnings' "memory:<id>"
// refs name edits by id. The PR is a unique field, so recording the same
// PR twice returns the first edit.

// EditsSpec is the memory_edits collection.
var EditsSpec = store.CollectionSpec{
	Name:   "memory_edits",
	Unique: []store.Field{{Name: "pr", Path: "pr"}},
}

// editValue is the record's JSON. Times are unix ms, as the old table
// kept them; created_at is the record's.
type editValue struct {
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

// AddPREdit records a merged curator PR. A PR already recorded returns its
// edit unchanged, so a repeated record is harmless.
func AddPREdit(ctx context.Context, st Store, e domain.MemoryEdit) (domain.MemoryEdit, error) {
	e.PR = strings.TrimSpace(e.PR)
	if e.PR == "" {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit needs a pr", domain.ErrInvalid)
	}
	if e.MergedAt.IsZero() {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit needs merged_at", domain.ErrInvalid)
	}
	e.Path, e.Before, e.Existed = "", "", false
	return addEdit(ctx, st, e)
}

// AddWriteEdit records a local memory write: Path, Before, Existed,
// Learnings, Pools, Baseline, and MergedAt as the write time. Every
// write is a new edit, so the gate can restore each one's before text.
func AddWriteEdit(ctx context.Context, st Store, e domain.MemoryEdit) (domain.MemoryEdit, error) {
	e.Path = strings.TrimSpace(e.Path)
	if e.Path == "" {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory write needs a path", domain.ErrInvalid)
	}
	if e.MergedAt.IsZero() {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory write needs merged_at", domain.ErrInvalid)
	}
	e.PR = ""
	return addEdit(ctx, st, e)
}

// addEdit inserts e under the next free id. A put that finds its id
// taken by the same PR is a repeat and returns that edit.
func addEdit(ctx context.Context, st Store, e domain.MemoryEdit) (domain.MemoryEdit, error) {
	col, err := st.Collection(ctx, EditsSpec)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	r, _, err := putNext(ctx, col, domain.EditPending, "memory edit", func(id int64) ([]byte, error) {
		e.ID = id
		return json.Marshal(toEditValue(e))
	}, func(r store.Record) (bool, error) {
		out, err := editFromRecord(r)
		return err == nil && e.PR != "" && out.PR == e.PR, err
	})
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	return editFromRecord(r)
}

// GetEdit returns one edit by id.
func GetEdit(ctx context.Context, st Store, id int64) (domain.MemoryEdit, error) {
	col, err := st.Collection(ctx, EditsSpec)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	r, err := col.Get(ctx, strconv.FormatInt(id, 10))
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	return editFromRecord(r)
}

// ListEdits returns the n newest edits (0 means store.DefaultLimit),
// newest merge first; empty state is all.
func ListEdits(ctx context.Context, st Store, state string, n int) ([]domain.MemoryEdit, error) {
	col, err := st.Collection(ctx, EditsSpec)
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
		e, err := editFromRecord(r)
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

// SetEditResult stores the gate's decision on an edit, evaluated at now.
func SetEditResult(ctx context.Context, st Store, id int64, state string, after domain.PoolMetrics, reason string, now time.Time) (domain.MemoryEdit, error) {
	if state != domain.EditHelped && state != domain.EditNotHelped {
		return domain.MemoryEdit{}, fmt.Errorf("%w: memory edit state %q", domain.ErrInvalid, state)
	}
	col, err := st.Collection(ctx, EditsSpec)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	r, err := casUpdate(ctx, col, strconv.FormatInt(id, 10), fmt.Sprintf("memory edit %d", id), func(r *store.Record) error {
		e, err := editFromRecord(*r)
		if err != nil {
			return err
		}
		e.State, e.After, e.Reason, e.EvaluatedAt = state, after, reason, now
		r.State = state
		r.Value, err = json.Marshal(toEditValue(e))
		return err
	})
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	return editFromRecord(r)
}

// EditRef names the edit in a cited learning's gate: the PR, or the edit
// id and path of a local write.
func EditRef(e domain.MemoryEdit) string {
	if e.PR != "" {
		return e.PR
	}
	return fmt.Sprintf("memory edit %d, %s", e.ID, e.Path)
}

const memoryRefPrefix = "memory:"

// MemoryRef is what a local write puts in its learnings' pr field, so the
// curator's settle pass and `learn list` show where each one went. It is
// a stored reference: existing records carry it.
func MemoryRef(id int64) string { return memoryRefPrefix + strconv.FormatInt(id, 10) }

// IsMemoryRef reports whether pr is a local write's reference, not a PR.
func IsMemoryRef(pr string) bool { return strings.HasPrefix(pr, memoryRefPrefix) }

func toEditValue(e domain.MemoryEdit) editValue {
	return editValue{
		ID: e.ID, PR: e.PR, Path: e.Path, Before: e.Before, Existed: e.Existed,
		Learnings: nonNil(e.Learnings), Pools: nonNil(e.Pools), MergedAt: ms(e.MergedAt),
		Baseline: e.Baseline, After: e.After, Reason: e.Reason, EvaluatedAt: ms(e.EvaluatedAt),
	}
}

func editFromRecord(r store.Record) (domain.MemoryEdit, error) {
	var v editValue
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
