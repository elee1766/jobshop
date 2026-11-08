package control

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Curator gate (#274). A memory edit is recorded with a baseline: the
// metrics of the scoped pools over the window before it. Local writes
// (#592, memory.write) record it at write time; older curator PRs at
// `learn update --status applied`. After enough runs in those pools, or
// enough days, the evaluate workflow calls `learn evaluate`, which
// measures the same window after and decides helped or not_helped. A
// not_helped local write gets its previous text back here; a not_helped
// PR edit is reverted by the evaluate step.

// Gate defaults; the trigger and the CLI can override them.
const (
	gateRuns     = 20
	gateAge      = 72 * time.Hour
	gateBaseline = 7 * 24 * time.Hour
)

// The judgement itself (metrics over a window, helped or not) is
// memedit.Metrics and memedit.Decide; this file reads the store and
// applies the result.

// measure reads the store and computes metrics over [from, to): the
// window after the merge when after is true, else the baseline before.
func (c *Controller) measure(ctx context.Context, cited []store.Learning, pools []string, from, to time.Time, after bool) (store.PoolMetrics, error) {
	steps, err := c.store.List(ctx, store.Filter{Kind: store.KindStep, Created: &store.TimeRange{Since: from, Until: to}, Order: store.OrderOldest, Limit: store.NoLimit})
	if err != nil {
		return store.PoolMetrics{}, err
	}
	ls, err := learning.Between(ctx, c.store, from, to)
	if err != nil {
		return store.PoolMetrics{}, err
	}
	return memedit.Metrics(steps, ls, cited, pools, from, to, after), nil
}

// editPools are the pools cited learnings point at: a pool scope, or the
// pool of the ticket (or its run's steps) they are about. Empty: all.
func (c *Controller) editPools(ctx context.Context, ls []store.Learning) []string {
	var out []string
	add := func(p string) {
		if p != "" && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	for _, l := range ls {
		if p, ok := strings.CutPrefix(l.Scope, "pool:"); ok {
			add(p)
			continue
		}
		if l.Ticket == "" {
			continue
		}
		t, err := c.store.Get(ctx, l.Ticket)
		if err != nil {
			continue
		}
		if t.Kind == store.KindRun {
			if steps, err := store.Children(ctx, c.store, t.ID); err == nil {
				for _, s := range steps {
					add(s.Pool)
				}
			}
			continue
		}
		add(t.Pool)
	}
	slices.Sort(out)
	return out
}

// recordEdit records a merged curator PR with its baseline. ids empty:
// the learnings whose pr is pr.
func (c *Controller) recordEdit(ctx context.Context, pr string, ids []int64, mergedAt time.Time) (store.MemoryEdit, error) {
	var cited []store.Learning
	if len(ids) == 0 {
		ls, err := learning.ByPR(ctx, c.store, pr)
		if err != nil {
			return store.MemoryEdit{}, err
		}
		cited = ls
	}
	for _, id := range ids {
		l, err := learning.Get(ctx, c.store, id)
		if err != nil {
			return store.MemoryEdit{}, err
		}
		cited = append(cited, l)
	}
	if mergedAt.IsZero() {
		mergedAt = c.now()
	}
	pools := c.editPools(ctx, cited)
	base, err := c.measure(ctx, cited, pools, mergedAt.Add(-gateBaseline), mergedAt, false)
	if err != nil {
		return store.MemoryEdit{}, err
	}
	e := store.MemoryEdit{PR: pr, Pools: pools, MergedAt: mergedAt, Baseline: base}
	for _, l := range cited {
		e.Learnings = append(e.Learnings, l.ID)
	}
	return memedit.AddPR(ctx, c.store, e)
}

// editDue reports whether a pending edit has enough evidence: runs runs
// in its pools since the merge, or age since the merge.
func (c *Controller) editDue(ctx context.Context, e store.MemoryEdit, runs int, age time.Duration, now time.Time) (bool, store.PoolMetrics, error) {
	cited, err := c.citedLearnings(ctx, e)
	if err != nil {
		return false, store.PoolMetrics{}, err
	}
	after, err := c.measure(ctx, cited, e.Pools, e.MergedAt, now, true)
	if err != nil {
		return false, after, err
	}
	return after.Runs >= runs || !now.Before(e.MergedAt.Add(age)), after, nil
}

func (c *Controller) citedLearnings(ctx context.Context, e store.MemoryEdit) ([]store.Learning, error) {
	out := make([]store.Learning, 0, len(e.Learnings))
	for _, id := range e.Learnings {
		l, err := learning.Get(ctx, c.store, id)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// learnEvaluate decides every pending edit that is due (or the named one,
// due or not) and records the result on the edit and its learnings.
func (c *Controller) learnEvaluate(ctx context.Context, _ scope, p rpc.LearnEvaluateParams) (any, error) {
	runs, age := p.Runs, p.Age
	if runs <= 0 {
		runs = gateRuns
	}
	if age <= 0 {
		age = gateAge
	}
	edits, err := memedit.List(ctx, c.store, store.EditPending, 0)
	if err != nil {
		return nil, err
	}
	now := c.now()
	out := []store.MemoryEdit{}
	for _, e := range edits {
		if p.ID != 0 && e.ID != p.ID {
			continue
		}
		due, after, err := c.editDue(ctx, e, runs, age, now)
		if err != nil {
			return nil, err
		}
		if !due && p.ID == 0 {
			continue
		}
		state, reason := memedit.Decide(e.Baseline, after)
		u, err := memedit.SetResult(ctx, c.store, e.ID, state, after, reason, now)
		if err != nil {
			return nil, err
		}
		// The edit's result is the record; the gate line on each cited
		// learning is a copy for `learn list`, so a failure here leaves
		// the edit decided and the next evaluate does not redo it.
		gate := fmt.Sprintf("%s: %s (%s)", state, reason, memedit.Ref(u))
		for _, l := range u.Learnings {
			if err := learning.SetGate(ctx, c.store, l, gate); err != nil {
				return nil, err
			}
		}
		// A local write that did not help is undone here (#592); a PR edit
		// is reverted by the evaluate step.
		if state == store.EditNotHelped {
			if err := c.restoreEdit(ctx, u); err != nil {
				return nil, err
			}
		}
		out = append(out, u)
	}
	if p.ID != 0 && len(out) == 0 {
		return nil, fmt.Errorf("%w: no pending memory edit %d", store.ErrNotFound, p.ID)
	}
	return out, nil
}

func (c *Controller) learnEdits(ctx context.Context, _ scope, p rpc.LearnEditsParams) (any, error) {
	return memedit.List(ctx, c.store, p.State, p.Limit)
}

// editsHold returns "" when a pending memory edit is due for evaluation.
func (c *Controller) editsHold(ctx context.Context, w config.EditsWhen, now time.Time) (string, error) {
	edits, err := memedit.List(ctx, c.store, store.EditPending, 0)
	if err != nil {
		return "", err
	}
	if len(edits) == 0 {
		return "no pending memory edits", nil
	}
	runs := w.Runs
	if runs <= 0 {
		runs = gateRuns
	}
	age := config.Duration(w.Age, gateAge)
	for _, e := range edits {
		due, _, err := c.editDue(ctx, e, runs, age, now)
		if err != nil || due {
			return "", err
		}
	}
	return fmt.Sprintf("%d pending memory edits, none due", len(edits)), nil
}
