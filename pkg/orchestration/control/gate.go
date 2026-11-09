package control

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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

// The judgement (metrics over a window, helped or not) and the work of
// recording, measuring and settling an edit are in pkg/knowledge; this file
// supplies the thresholds, the pools and the clock.

// editPools are the pools cited learnings point at: a pool scope, or the
// pool of the ticket (or its run's steps) they are about. Empty: all.
func (c *Controller) editPools(ctx context.Context, ls []domain.Learning) []string {
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
		if t.Kind == domain.KindRun {
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
func (c *Controller) recordEdit(ctx context.Context, pr string, ids []int64, mergedAt time.Time) (domain.MemoryEdit, error) {
	cited, err := knowledge.CitedForPR(ctx, c.store, pr, ids)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	if mergedAt.IsZero() {
		mergedAt = c.now()
	}
	return knowledge.RecordPREdit(ctx, c.store, pr, cited, c.editPools(ctx, cited), mergedAt, gateBaseline)
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
	edits, err := knowledge.ListEdits(ctx, c.store, domain.EditPending, 0)
	if err != nil {
		return nil, err
	}
	now := c.now()
	out := []domain.MemoryEdit{}
	for _, e := range edits {
		if p.ID != 0 && e.ID != p.ID {
			continue
		}
		due, after, err := knowledge.EditDue(ctx, c.store, e, runs, age, now)
		if err != nil {
			return nil, err
		}
		if !due && p.ID == 0 {
			continue
		}
		u, err := knowledge.SettleEdit(ctx, c.store, e, after, now)
		if err != nil {
			return nil, err
		}
		// A local write that did not help is undone here (#592); a PR edit
		// is reverted by the evaluate step.
		if u.State == domain.EditNotHelped {
			if err := c.restoreEdit(ctx, u); err != nil {
				return nil, err
			}
		}
		out = append(out, u)
	}
	if p.ID != 0 && len(out) == 0 {
		return nil, fmt.Errorf("%w: no pending memory edit %d", domain.ErrNotFound, p.ID)
	}
	return out, nil
}

func (c *Controller) learnEdits(ctx context.Context, _ scope, p rpc.LearnEditsParams) (any, error) {
	return knowledge.ListEdits(ctx, c.store, p.State, p.Limit)
}

// editsHold returns "" when a pending memory edit is due for evaluation.
func (c *Controller) editsHold(ctx context.Context, w config.EditsWhen, now time.Time) (string, error) {
	edits, err := knowledge.ListEdits(ctx, c.store, domain.EditPending, 0)
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
		due, _, err := knowledge.EditDue(ctx, c.store, e, runs, age, now)
		if err != nil || due {
			return "", err
		}
	}
	return fmt.Sprintf("%d pending memory edits, none due", len(edits)), nil
}
