// Package learning stores learnings (#264) on the learnings store
// collection (#579). Agents and the controller record corrections,
// failures and workarounds; the curator reads and settles them.
//
// Records are keyed by the learning id as a decimal string. Ids stay
// integers because the CLI, the RPC and memory edits name learnings by
// id. The record state is the learning status. Summary and fix feed the
// collection's full-text index, which recall reads.
package learning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Spec is the learnings collection.
var Spec = store.CollectionSpec{
	Name: "learnings",
	Index: []store.Field{
		{Name: "kind", Path: "kind"},
		{Name: "scope", Path: "scope"},
		{Name: "ticket", Path: "ticket"},
		{Name: "pr", Path: "pr"},
	},
	Search: []store.Field{{Name: "summary", Path: "summary"}, {Name: "fix", Path: "fix"}},
}

// Store is the part of the store learnings need.
type Store interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

// value is the record's JSON; created_at and the status are the record's.
// Project is the project of the ticket, for recall's project filter.
type value struct {
	ID      int64  `json:"id"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Fix     string `json:"fix"`
	Scope   string `json:"scope"`
	Ticket  string `json:"ticket"`
	Session string `json:"session"`
	Source  string `json:"source"`
	PR      string `json:"pr"`
	Gate    string `json:"gate"`
	Project string `json:"project"`
}

const (
	// maxTries bounds the id and compare-and-set loops under races.
	maxTries = 10
	// newest is how many of the newest records Add reads for the highest
	// id: records in the same millisecond tie on created_at.
	newest = 20
	// all is the List limit of reads that need every match (counts,
	// gate windows). Learnings are hundreds of rows.
	all = math.MaxInt32
)

func open(ctx context.Context, st Store) (store.Collection, error) {
	return st.Collection(ctx, Spec)
}

// Add records l (validated by store.ValidLearning) under the next id.
// project is the project of l.Ticket, or empty.
func Add(ctx context.Context, st Store, l domain.Learning, project string) (domain.Learning, error) {
	if err := domain.ValidLearning(&l); err != nil {
		return domain.Learning{}, err
	}
	col, err := open(ctx, st)
	if err != nil {
		return domain.Learning{}, err
	}
	last, err := col.List(ctx, store.RecordFilter{Order: store.OrderCreatedDesc, Limit: newest})
	if err != nil {
		return domain.Learning{}, err
	}
	id := int64(1)
	for _, r := range last {
		if n, err := strconv.ParseInt(r.Key, 10, 64); err == nil && n >= id {
			id = n + 1
		}
	}
	for range maxTries {
		v := toValue(l)
		v.ID, v.Project = id, project
		b, err := json.Marshal(v)
		if err != nil {
			return domain.Learning{}, err
		}
		r, created, err := col.Put(ctx, store.Record{Key: key(id), State: l.Status, Value: b})
		if err != nil {
			return domain.Learning{}, err
		}
		if created {
			return fromRecord(r)
		}
		id++ // taken by a concurrent add
	}
	return domain.Learning{}, fmt.Errorf("%w: learning: lost every id race", domain.ErrConflict)
}

// Get reads one learning; ErrNotFound if missing.
func Get(ctx context.Context, st Store, id int64) (domain.Learning, error) {
	col, err := open(ctx, st)
	if err != nil {
		return domain.Learning{}, err
	}
	r, err := col.Get(ctx, key(id))
	if err != nil {
		return domain.Learning{}, err
	}
	return fromRecord(r)
}

// List returns learnings matching f, newest first.
func List(ctx context.Context, st Store, f store.LearningFilter) ([]domain.Learning, error) {
	rf := store.RecordFilter{Since: f.Since, Order: store.OrderCreatedDesc, Limit: f.Limit, Where: map[string]string{}}
	if f.Status != "" {
		rf.States = []string{f.Status}
	}
	for name, v := range map[string]string{"scope": f.Scope, "kind": f.Kind, "ticket": f.Ticket} {
		if v != "" {
			rf.Where[name] = v
		}
	}
	return list(ctx, st, rf)
}

// Between returns learnings recorded in [since, until), oldest first.
func Between(ctx context.Context, st Store, since, until time.Time) ([]domain.Learning, error) {
	if until.IsZero() || !since.Before(until) {
		return []domain.Learning{}, nil
	}
	return list(ctx, st, store.RecordFilter{Since: since, Until: until, Limit: all})
}

// ByPR returns the learnings whose PR is pr, oldest first.
func ByPR(ctx context.Context, st Store, pr string) ([]domain.Learning, error) {
	pr = strings.TrimSpace(pr)
	if pr == "" {
		return []domain.Learning{}, nil
	}
	return list(ctx, st, store.RecordFilter{Where: map[string]string{"pr": pr}, Limit: all})
}

// Count counts learnings created at or after since that match f. Kind
// and scope globs are matched here.
func Count(ctx context.Context, st Store, since time.Time, f domain.LearningCount) (int, error) {
	rf := store.RecordFilter{Since: since, Limit: all}
	if f.Status != "" {
		rf.States = []string{f.Status}
	}
	ls, err := list(ctx, st, rf)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, l := range ls {
		if f.Matches(l.Kind, l.Scope, l.Source) {
			n++
		}
	}
	return n, nil
}

// Update sets a learning's status and PR (#273).
func Update(ctx context.Context, st Store, id int64, status, pr string) (domain.Learning, error) {
	if status != domain.LearningOpen && status != domain.LearningApplied && status != domain.LearningRejected {
		return domain.Learning{}, fmt.Errorf("%w: learning status %q", domain.ErrInvalid, status)
	}
	return update(ctx, st, id, func(r *store.Record, v *value) {
		r.State, v.PR = status, strings.TrimSpace(pr)
	})
}

// SetGate writes the curator gate's result on a learning (#274).
func SetGate(ctx context.Context, st Store, id int64, gate string) error {
	_, err := update(ctx, st, id, func(_ *store.Record, v *value) { v.Gate = gate })
	return err
}

func update(ctx context.Context, st Store, id int64, fn func(*store.Record, *value)) (domain.Learning, error) {
	col, err := open(ctx, st)
	if err != nil {
		return domain.Learning{}, err
	}
	for range maxTries {
		r, err := col.Get(ctx, key(id))
		if err != nil {
			return domain.Learning{}, err
		}
		var v value
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return domain.Learning{}, fmt.Errorf("learning %d: %w", id, err)
		}
		fn(&r, &v)
		if r.Value, err = json.Marshal(v); err != nil {
			return domain.Learning{}, err
		}
		r, err = col.Update(ctx, r)
		if errors.Is(err, domain.ErrConflict) {
			continue
		}
		if err != nil {
			return domain.Learning{}, err
		}
		return fromRecord(r)
	}
	return domain.Learning{}, fmt.Errorf("%w: learning %d: lost every race", domain.ErrConflict, id)
}

// snippetMax bounds a recall snippet.
const snippetMax = 200

// Recall searches learnings' summary and fix, best match first. A
// learning with no project (global) matches every project, as the old
// recall index did; ticketID narrows to one ticket.
func Recall(ctx context.Context, st Store, text, project, ticketID string, limit int) ([]store.RecallHit, error) {
	col, err := open(ctx, st)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	// Filters run here, after the search: fetch more than asked.
	rs, err := col.Search(ctx, text, limit*10)
	if err != nil {
		return nil, err
	}
	out := []store.RecallHit{}
	for _, r := range rs {
		var v value
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return nil, fmt.Errorf("learning %s: %w", r.Key, err)
		}
		if project != "" && v.Project != "" && v.Project != project {
			continue
		}
		if ticketID != "" && v.Ticket != ticketID {
			continue
		}
		snip := strings.TrimSpace(v.Summary + "\n" + v.Fix)
		if len(snip) > snippetMax {
			snip = snip[:snippetMax] + "..."
		}
		out = append(out, store.RecallHit{
			Kind: store.RecallLearning, Ref: r.Key, TicketID: v.Ticket, Project: v.Project,
			At: r.CreatedAt, Snippet: snip, Score: float64(len(out)),
		})
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func list(ctx context.Context, st Store, f store.RecordFilter) ([]domain.Learning, error) {
	col, err := open(ctx, st)
	if err != nil {
		return nil, err
	}
	if len(f.Where) == 0 {
		f.Where = nil
	}
	rs, err := col.List(ctx, f)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Learning, 0, len(rs))
	for _, r := range rs {
		l, err := fromRecord(r)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	// created_at ties order by key as text; ids order numerically.
	desc := f.Order == store.OrderCreatedDesc
	slices.SortStableFunc(out, func(a, b domain.Learning) int {
		c := a.CreatedAt.Compare(b.CreatedAt)
		if c == 0 {
			c = cmpInt(a.ID, b.ID)
		}
		if desc {
			return -c
		}
		return c
	})
	return out, nil
}

func cmpInt(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func key(id int64) string { return strconv.FormatInt(id, 10) }

func toValue(l domain.Learning) value {
	return value{
		ID: l.ID, Kind: l.Kind, Summary: l.Summary, Fix: l.Fix, Scope: l.Scope, Ticket: l.Ticket,
		Session: l.Session, Source: l.Source, PR: l.PR, Gate: l.Gate,
	}
}

func fromRecord(r store.Record) (domain.Learning, error) {
	var v value
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return domain.Learning{}, fmt.Errorf("learning %s: %w", r.Key, err)
	}
	return domain.Learning{
		ID: v.ID, Kind: v.Kind, Summary: v.Summary, Fix: v.Fix, Scope: v.Scope, Ticket: v.Ticket,
		Session: v.Session, Source: v.Source, Status: r.State, PR: v.PR, Gate: v.Gate, CreatedAt: r.CreatedAt,
	}, nil
}
