package control

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

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
	// gateWorse is how much the failure rate may grow before an edit
	// counts as harmful: noise between windows of a few dozen steps.
	gateWorse = 1.25
	// similarMin is the word overlap (Jaccard) at which a learning
	// repeats a cited one.
	similarMin = 0.5
)

// reviewFindingPrefix starts the correction recordReviewFinding writes.
const reviewFindingPrefix = "review resumed "

// gateMetrics computes the metrics of steps in pools (empty: all) and
// learnings recorded in [from, to). Recurrence counts learnings that
// repeat one of cited (same kind, similar summary). Before the merge
// (after false) the cited learnings are the problem itself and count.
// After it (after true) a cited learning is not a recurrence, however
// it matches: the session that ships the fix records a learning about
// it later, and that is not the problem coming back (#1061).
func gateMetrics(steps []store.Ticket, ls []store.Learning, cited []store.Learning, pools []string, from, to time.Time, after bool) store.PoolMetrics {
	m := store.PoolMetrics{From: from, To: to}
	inWin := func(t time.Time) bool { return !t.Before(from) && t.Before(to) }
	runs := map[string]bool{}
	ids := map[string]bool{}
	for _, s := range steps {
		if s.Kind != store.KindStep || !inWin(s.CreatedAt) || (len(pools) > 0 && !slices.Contains(pools, s.Pool)) {
			continue
		}
		m.Steps++
		ids[s.ID] = true
		if s.ParentID != "" {
			runs[s.ParentID] = true
		}
		if sf, err := store.DecodeFields[store.StepFields](s); err == nil {
			m.Retries += sf.Failures
		}
		if s.Status == store.StatusClosed && s.Resolution == "failed" {
			m.Escalations++
		}
	}
	m.Runs = len(runs)
	words := make([]map[string]bool, len(cited))
	citedID := map[int64]bool{}
	for i, c := range cited {
		words[i] = wordSet(c.Summary)
		citedID[c.ID] = true
	}
	for _, l := range ls {
		if !inWin(l.CreatedAt) {
			continue
		}
		if l.Kind == store.LearningCorrection && strings.HasPrefix(l.Summary, reviewFindingPrefix) && ids[l.Ticket] {
			m.Findings++
		}
		if after && citedID[l.ID] {
			continue
		}
		w := wordSet(l.Summary)
		for i, c := range cited {
			if c.ID == l.ID || (c.Kind == l.Kind && jaccard(words[i], w) >= similarMin) {
				m.Recurrence++
				break
			}
		}
	}
	return m
}

// gateDecide compares before and after. The edit did not help when the
// cited problem recurs at least as often per step as before, or when
// retries, escalations and review findings per step grew by more than
// gateWorse. No steps after the merge is no evidence: it helped.
func gateDecide(base, after store.PoolMetrics) (string, string) {
	if after.Steps == 0 {
		return store.EditHelped, "no steps in the scoped pools since the merge"
	}
	rb, ra := base.Rate(base.Recurrence), after.Rate(after.Recurrence)
	if after.Recurrence > 0 && ra >= rb {
		return store.EditNotHelped, fmt.Sprintf("cited problem still recurs: %d in %d steps after, %d in %d before",
			after.Recurrence, after.Steps, base.Recurrence, base.Steps)
	}
	fb := base.Rate(base.Retries + base.Escalations + base.Findings)
	fa := after.Rate(after.Retries + after.Escalations + after.Findings)
	if fa > fb*gateWorse && fa > 0 && after.Retries+after.Escalations+after.Findings > 1 {
		return store.EditNotHelped, fmt.Sprintf("failures per step rose from %.2f to %.2f (retries %d, escalations %d, findings %d in %d steps)",
			fb, fa, after.Retries, after.Escalations, after.Findings, after.Steps)
	}
	return store.EditHelped, fmt.Sprintf("recurrence %d in %d steps (before %d in %d), failures per step %.2f (before %.2f)",
		after.Recurrence, after.Steps, base.Recurrence, base.Steps, fa, fb)
}

func wordSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if len(w) >= 4 {
			out[w] = true
		}
	}
	return out
}

func jaccard(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := 0
	for w := range a {
		if b[w] {
			n++
		}
	}
	return float64(n) / float64(len(a)+len(b)-n)
}

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
	return gateMetrics(steps, ls, cited, pools, from, to, after), nil
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
		state, reason := gateDecide(e.Baseline, after)
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
