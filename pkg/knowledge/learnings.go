package knowledge

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

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Learnings (#264) live on the learnings store collection (#579). Agents
// and the controller record corrections, failures and workarounds; the
// curator reads and settles them.
//
// Records are keyed by the learning id as a decimal string. Ids stay
// integers because the CLI, the RPC and memory edits name learnings by
// id. The record state is the learning status. Summary and fix feed the
// collection's full-text index, which recall reads.

// LearningsSpec is the learnings collection.
var LearningsSpec = store.CollectionSpec{
	Name: "learnings",
	Index: []store.Field{
		{Name: "kind", Path: "kind"},
		{Name: "scope", Path: "scope"},
		{Name: "ticket", Path: "ticket"},
		{Name: "pr", Path: "pr"},
	},
	Search: []store.Field{{Name: "summary", Path: "summary"}, {Name: "fix", Path: "fix"}},
}

// learningValue is the record's JSON; created_at and the status are the
// record's. Project is the project of the ticket, for recall's project
// filter.
type learningValue struct {
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

// all is the List limit of reads that need every match (counts, gate
// windows). Learnings are hundreds of rows.
const all = math.MaxInt32

func openLearnings(ctx context.Context, st Store) (store.Collection, error) {
	return st.Collection(ctx, LearningsSpec)
}

// AddLearning records l (validated by domain.ValidLearning) under the next
// id. project is the project of l.Ticket, or empty.
func AddLearning(ctx context.Context, st Store, l domain.Learning, project string) (domain.Learning, error) {
	if err := domain.ValidLearning(&l); err != nil {
		return domain.Learning{}, err
	}
	col, err := openLearnings(ctx, st)
	if err != nil {
		return domain.Learning{}, err
	}
	r, _, err := putNext(ctx, col, l.Status, "learning", func(id int64) ([]byte, error) {
		v := toLearningValue(l)
		v.ID, v.Project = id, project
		return json.Marshal(v)
	}, nil)
	if err != nil {
		return domain.Learning{}, err
	}
	return learningFromRecord(r)
}

// GetLearning reads one learning; ErrNotFound if missing.
func GetLearning(ctx context.Context, st Store, id int64) (domain.Learning, error) {
	col, err := openLearnings(ctx, st)
	if err != nil {
		return domain.Learning{}, err
	}
	r, err := col.Get(ctx, learningKey(id))
	if err != nil {
		return domain.Learning{}, err
	}
	return learningFromRecord(r)
}

// GetLearnings reads the learnings named by ids, in that order.
func GetLearnings(ctx context.Context, st Store, ids []int64) ([]domain.Learning, error) {
	out := make([]domain.Learning, 0, len(ids))
	for _, id := range ids {
		l, err := GetLearning(ctx, st, id)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// ListLearnings returns learnings matching f, newest first.
func ListLearnings(ctx context.Context, st Store, f store.LearningFilter) ([]domain.Learning, error) {
	rf := store.RecordFilter{Since: f.Since, Order: store.OrderCreatedDesc, Limit: f.Limit, Where: map[string]string{}}
	if f.Status != "" {
		rf.States = []string{f.Status}
	}
	for name, v := range map[string]string{"scope": f.Scope, "kind": f.Kind, "ticket": f.Ticket} {
		if v != "" {
			rf.Where[name] = v
		}
	}
	return listLearnings(ctx, st, rf)
}

// LearningsBetween returns learnings recorded in [since, until), oldest
// first.
func LearningsBetween(ctx context.Context, st Store, since, until time.Time) ([]domain.Learning, error) {
	if until.IsZero() || !since.Before(until) {
		return []domain.Learning{}, nil
	}
	return listLearnings(ctx, st, store.RecordFilter{Since: since, Until: until, Limit: all})
}

// LearningsByPR returns the learnings whose PR is pr, oldest first.
func LearningsByPR(ctx context.Context, st Store, pr string) ([]domain.Learning, error) {
	pr = strings.TrimSpace(pr)
	if pr == "" {
		return []domain.Learning{}, nil
	}
	return listLearnings(ctx, st, store.RecordFilter{Where: map[string]string{"pr": pr}, Limit: all})
}

// CountLearnings counts learnings created at or after since that match f.
// Kind and scope globs are matched here.
func CountLearnings(ctx context.Context, st Store, since time.Time, f domain.LearningCount) (int, error) {
	rf := store.RecordFilter{Since: since, Limit: all}
	if f.Status != "" {
		rf.States = []string{f.Status}
	}
	ls, err := listLearnings(ctx, st, rf)
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

// UpdateLearning sets a learning's status and PR (#273).
func UpdateLearning(ctx context.Context, st Store, id int64, status, pr string) (domain.Learning, error) {
	if !domain.ValidLearningStatus(status) {
		return domain.Learning{}, fmt.Errorf("%w: learning status %q", domain.ErrInvalid, status)
	}
	return updateLearning(ctx, st, id, func(r *store.Record, v *learningValue) {
		r.State, v.PR = status, strings.TrimSpace(pr)
	})
}

// SetLearningGate writes the curator gate's result on a learning (#274).
func SetLearningGate(ctx context.Context, st Store, id int64, gate string) error {
	_, err := updateLearning(ctx, st, id, func(_ *store.Record, v *learningValue) { v.Gate = gate })
	return err
}

func updateLearning(ctx context.Context, st Store, id int64, fn func(*store.Record, *learningValue)) (domain.Learning, error) {
	col, err := openLearnings(ctx, st)
	if err != nil {
		return domain.Learning{}, err
	}
	r, err := casUpdate(ctx, col, learningKey(id), fmt.Sprintf("learning %d", id), func(r *store.Record) error {
		var v learningValue
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return fmt.Errorf("learning %d: %w", id, err)
		}
		fn(r, &v)
		b, err := json.Marshal(v)
		r.Value = b
		return err
	})
	if err != nil {
		return domain.Learning{}, err
	}
	return learningFromRecord(r)
}

// snippetMax bounds a recall snippet.
const snippetMax = 200

// RecallLearnings searches learnings' summary and fix, best match first. A
// learning with no project (global) matches every project, as the old
// recall index did; ticketID narrows to one ticket.
func RecallLearnings(ctx context.Context, st Store, text, project, ticketID string, limit int) ([]store.RecallHit, error) {
	col, err := openLearnings(ctx, st)
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
		var v learningValue
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

func listLearnings(ctx context.Context, st Store, f store.RecordFilter) ([]domain.Learning, error) {
	col, err := openLearnings(ctx, st)
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
		l, err := learningFromRecord(r)
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

func learningKey(id int64) string { return strconv.FormatInt(id, 10) }

func toLearningValue(l domain.Learning) learningValue {
	return learningValue{
		ID: l.ID, Kind: l.Kind, Summary: l.Summary, Fix: l.Fix, Scope: l.Scope, Ticket: l.Ticket,
		Session: l.Session, Source: l.Source, PR: l.PR, Gate: l.Gate,
	}
}

func learningFromRecord(r store.Record) (domain.Learning, error) {
	var v learningValue
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return domain.Learning{}, fmt.Errorf("learning %s: %w", r.Key, err)
	}
	return domain.Learning{
		ID: v.ID, Kind: v.Kind, Summary: v.Summary, Fix: v.Fix, Scope: v.Scope, Ticket: v.Ticket,
		Session: v.Session, Source: v.Source, Status: r.State, PR: v.PR, Gate: v.Gate, CreatedAt: r.CreatedAt,
	}, nil
}

// ExpireLearnings sets expired on the open learnings recorded before
// cutoff that have no PR (JOB-198), and returns the ones it changed. scope
// is exact and kind a name or glob; empty matches all. A learning that
// gained a PR or left open since it was listed is skipped.
func ExpireLearnings(ctx context.Context, st Store, cutoff time.Time, scope, kind string) ([]domain.Learning, error) {
	rf := store.RecordFilter{Until: cutoff, States: []string{domain.LearningOpen}, Order: store.OrderCreatedDesc, Limit: all, Where: map[string]string{}}
	if scope != "" {
		rf.Where["scope"] = scope
	}
	if kind != "" && !strings.ContainsAny(kind, "*?[") {
		rf.Where["kind"] = kind
	}
	cands, err := listLearnings(ctx, st, rf)
	if err != nil {
		return nil, err
	}
	col, err := openLearnings(ctx, st)
	if err != nil {
		return nil, err
	}
	out := []domain.Learning{}
	for _, l := range cands {
		if l.PR != "" || (kind != "" && !domain.Glob(kind, l.Kind)) {
			continue
		}
		r, err := casUpdate(ctx, col, learningKey(l.ID), fmt.Sprintf("learning %d", l.ID), func(r *store.Record) error {
			var v learningValue
			if err := json.Unmarshal(r.Value, &v); err != nil {
				return fmt.Errorf("learning %d: %w", l.ID, err)
			}
			if r.State != domain.LearningOpen || v.PR != "" {
				return errNotExpirable
			}
			r.State = domain.LearningExpired
			return nil
		})
		if errors.Is(err, errNotExpirable) {
			continue
		}
		if err != nil {
			return out, err
		}
		u, err := learningFromRecord(r)
		if err != nil {
			return out, err
		}
		out = append(out, u)
	}
	return out, nil
}

var errNotExpirable = errors.New("learning not expirable")
