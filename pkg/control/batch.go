package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Batches (#148 design, #169 build). A batch is a kind batch ticket with no
// pool; its member runs are its children. batch.start writes the batch,
// new runs, re-parents of adopted runs and the cross-run edges in one
// transaction, so no step is claimable before its edges exist.

// defaultBatchWorkflow is the workflow each member run starts.
const defaultBatchWorkflow = "pr-work"

// parseOrder turns "a>b>c" chains into [a b] [b c] pairs. Every issue named
// must be in issues.
func parseOrder(order, issues []string) ([][2]string, error) {
	var pairs [][2]string
	for _, chain := range order {
		parts := strings.Split(chain, ">")
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: order %q: want a>b", store.ErrInvalid, chain)
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			if !slices.Contains(issues, parts[i]) {
				return nil, fmt.Errorf("%w: order %q: issue %q is not in the batch", store.ErrInvalid, chain, parts[i])
			}
		}
		for i := 1; i < len(parts); i++ {
			if parts[i-1] == parts[i] {
				return nil, fmt.Errorf("%w: order %q: issue %s after itself", store.ErrInvalid, chain, parts[i])
			}
			p := [2]string{parts[i-1], parts[i]}
			if !slices.Contains(pairs, p) {
				pairs = append(pairs, p)
			}
		}
	}
	return pairs, nil
}

// batchMember is one issue's run while a start is being planned.
type batchMember struct {
	issue   string
	run     store.Ticket   // adopted run, or the new run ticket
	steps   []store.Ticket // adopted steps
	newRun  store.Graph    // new run graph; empty when adopted
	adopted bool
}

func (m batchMember) firstStep() string {
	if m.adopted {
		return m.steps[0].ID
	}
	return m.newRun.Tickets[1].ID
}

func (m batchMember) lastStep() string {
	if m.adopted {
		return m.steps[len(m.steps)-1].ID
	}
	return m.newRun.Tickets[len(m.newRun.Tickets)-1].ID
}

func (c *Controller) batchStart(ctx context.Context, sc scope, p rpc.BatchStartParams) (any, error) {
	// Scope is checked per member by planRun, as for run.start: the
	// planner is a session and starts batches in its own project.
	if p.Key == "" {
		return nil, fmt.Errorf("%w: batch: key is required", store.ErrInvalid)
	}
	if len(p.Issues) == 0 {
		return nil, fmt.Errorf("%w: batch: issues are required", store.ErrInvalid)
	}
	issues := make([]string, len(p.Issues))
	for i, n := range p.Issues {
		n = strings.TrimPrefix(strings.TrimSpace(n), "#")
		if n == "" || slices.Contains(issues[:i], n) {
			return nil, fmt.Errorf("%w: batch: empty or repeated issue %q", store.ErrInvalid, p.Issues[i])
		}
		issues[i] = n
	}
	pairs, err := parseOrder(p.Order, issues)
	if err != nil {
		return nil, err
	}
	workflow := p.Workflow
	if workflow == "" {
		workflow = defaultBatchWorkflow
	}
	issueVar := c.config().Workflows[workflow].Issue
	if issueVar == "" {
		issueVar = "issue"
	}
	params := make([]rpc.RunStartParams, len(issues))
	plans := make([]runPlan, len(issues))
	for i, n := range issues {
		params[i] = rpc.RunStartParams{Workflow: workflow, Vars: map[string]string{issueVar: n}, Key: "issue-" + n, Project: p.Project}
		if plans[i], err = c.planRun(ctx, sc, params[i]); err != nil {
			return nil, fmt.Errorf("issue %s: %w", n, err)
		}
	}
	proj := plans[0].proj

	c.runMu.Lock()
	defer c.runMu.Unlock()
	if t, err := c.store.FindOpenBatchByKey(ctx, proj, p.Key); err == nil {
		res, err := c.batchResult(ctx, t)
		if err != nil {
			return nil, err
		}
		res.Existing = true
		return res, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	batchID, err := c.store.NewID(proj)
	if err != nil {
		return nil, err
	}
	bf, err := json.Marshal(store.BatchFields{Key: p.Key, Issues: issues, Order: pairs, State: store.BatchOpen})
	if err != nil {
		return nil, err
	}
	g := store.Graph{
		Tickets:  []store.NewTicket{{ID: batchID, Project: proj, Kind: store.KindBatch, Title: "Batch " + p.Key, Fields: bf}},
		Reparent: map[string]string{},
	}
	members := map[string]*batchMember{}
	var adopted []string
	for i, n := range issues {
		m := &batchMember{issue: n}
		members[n] = m
		run, err := c.store.FindOpenRunByKey(ctx, workflow, params[i].Key)
		switch {
		case err == nil:
			if run.ParentID != "" && run.ParentID != batchID {
				if parent, err := c.store.Get(ctx, run.ParentID); err == nil && parent.Kind == store.KindBatch && parent.Status != store.StatusClosed {
					return nil, fmt.Errorf("%w: issue %s: run %s is already in batch %s", store.ErrInvalid, n, run.ID, parent.ID)
				}
			}
			steps, err := c.store.ListChildren(ctx, run.ID)
			if err != nil {
				return nil, err
			}
			if len(steps) == 0 {
				return nil, fmt.Errorf("%w: issue %s: run %s has no steps", store.ErrInvalid, n, run.ID)
			}
			m.run, m.steps, m.adopted = run, steps, true
			g.Reparent[run.ID] = batchID
			adopted = append(adopted, run.ID)
		case errors.Is(err, store.ErrNotFound):
			if m.newRun, err = c.buildRun(plans[i], params[i], nil, batchID); err != nil {
				return nil, fmt.Errorf("issue %s: %w", n, err)
			}
			g.Tickets = append(g.Tickets, m.newRun.Tickets...)
			g.Edges = append(g.Edges, m.newRun.Edges...)
		default:
			return nil, err
		}
	}
	for _, pr := range pairs {
		a, b := members[pr[0]], members[pr[1]]
		first := b.firstStep()
		if b.adopted {
			// An edge on a claimed step does not stop the session working
			// it, so refuse instead of reporting an order that will not hold.
			s := b.steps[0]
			if s.Status != store.StatusOpen || s.Holder != "" {
				return nil, fmt.Errorf("%w: issue %s must wait for issue %s, but its run %s step %s is already %s (holder %q)",
					store.ErrNotHolder, b.issue, a.issue, b.run.ID, s.ID, s.Status, s.Holder)
			}
			if !slices.Contains(g.Unclaimed, first) {
				g.Unclaimed = append(g.Unclaimed, first)
			}
		}
		g.Edges = append(g.Edges, store.Edge{From: first, To: a.lastStep(), Kind: store.EdgeBlocks})
	}
	out, err := c.store.ApplyGraph(ctx, sc.actor, g)
	if err != nil {
		return nil, err
	}
	c.log.Info("batch started", "batch", batchID, "key", p.Key, "issues", issues, "adopted", adopted)
	res, err := c.batchResult(ctx, out[0])
	if err != nil {
		return nil, err
	}
	res.Adopted = adopted
	return res, nil
}

// batchResult loads the batch's member runs in fields.issues order.
func (c *Controller) batchResult(ctx context.Context, b store.Ticket) (rpc.BatchResult, error) {
	res := rpc.BatchResult{Batch: b, Runs: []rpc.RunResult{}}
	kids, err := c.store.ListChildren(ctx, b.ID)
	if err != nil {
		return res, err
	}
	f, err := store.DecodeFields[store.BatchFields](b)
	if err != nil {
		return res, err
	}
	rank := func(t store.Ticket) int {
		if rf, err := store.DecodeFields[store.RunFields](t); err == nil {
			if i := slices.Index(f.Issues, strings.TrimPrefix(rf.Key, "issue-")); i >= 0 {
				return i
			}
		}
		return len(f.Issues)
	}
	slices.SortStableFunc(kids, func(x, y store.Ticket) int { return rank(x) - rank(y) })
	for _, k := range kids {
		if k.Kind != store.KindRun {
			continue
		}
		r, err := c.runResult(ctx, k)
		if err != nil {
			return res, err
		}
		res.Runs = append(res.Runs, r)
	}
	return res, nil
}
