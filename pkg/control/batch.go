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
	if t, err := store.OpenBatch(ctx, c.store, proj, p.Key); err == nil {
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
		run, err := store.OpenRun(ctx, c.store, workflow, params[i].Key)
		switch {
		case err == nil:
			if run.ParentID != "" && run.ParentID != batchID {
				if parent, err := c.store.Get(ctx, run.ParentID); err == nil && parent.Kind == store.KindBatch && parent.Status != store.StatusClosed {
					return nil, fmt.Errorf("%w: issue %s: run %s is already in batch %s", store.ErrInvalid, n, run.ID, parent.ID)
				}
			}
			steps, err := store.Children(ctx, c.store, run.ID)
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
	kids, err := store.Children(ctx, c.store, b.ID)
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

// deriveBatch sets the batch fields.state from its member runs, the way
// runs derive theirs from steps (#170): failed if a member run closed
// without done (cancelled), done when every member is done, needs_human if
// any member needs a human, running once any member step was claimed.
// A batch that reaches done or failed is closed with that resolution. It
// files no escalation of its own: needs_human points at the member run.
func (c *Controller) deriveBatch(ctx context.Context, batchID string) {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	b, err := c.store.Get(ctx, batchID)
	if err != nil || b.Kind != store.KindBatch || b.Status == store.StatusClosed {
		return
	}
	f, err := store.DecodeFields[store.BatchFields](b)
	if err != nil {
		return
	}
	kids, err := store.Children(ctx, c.store, batchID)
	if err != nil {
		return
	}
	state, err := c.batchState(ctx, f.State, kids)
	if err != nil || state == f.State {
		return
	}
	f.State = state
	raw, err := json.Marshal(f)
	if err != nil {
		return
	}
	if b, err = c.store.Update(ctx, "controller", b.ID, store.Update{Fields: raw, IfRevision: b.Revision}); err != nil {
		c.log.Error("batch state", "batch", batchID, "err", err)
		return
	}
	if state == store.BatchDone || state == store.BatchFailed {
		if _, err := store.Close(ctx, c.store, "controller", b.ID, nil, state); err != nil {
			c.log.Error("close batch", "batch", batchID, "err", err)
			return
		}
	}
	c.log.Info("batch state", "batch", batchID, "key", f.Key, "state", state)
}

// batchState computes the state from the member runs. cur is the stored
// state: a batch that left open does not return to it (a resumed step is
// open again, but the batch has been running).
func (c *Controller) batchState(ctx context.Context, cur string, kids []store.Ticket) (string, error) {
	members, done, needsHuman, started := 0, 0, false, cur != store.BatchOpen
	for _, k := range kids {
		if k.Kind != store.KindRun {
			continue
		}
		members++
		rf, err := store.DecodeFields[store.RunFields](k)
		if err != nil {
			return "", err
		}
		switch {
		case k.Status == store.StatusClosed && rf.State == store.RunDone:
			done++
			started = true
			continue
		case k.Status == store.StatusClosed:
			return store.BatchFailed, nil
		case rf.State == store.RunNeedsHuman:
			needsHuman = true
		}
		if started {
			continue
		}
		steps, err := store.Children(ctx, c.store, k.ID)
		if err != nil {
			return "", err
		}
		for _, s := range steps {
			if s.Status != store.StatusOpen || s.Holder != "" {
				started = true
				break
			}
		}
	}
	switch {
	case members == 0:
		return cur, nil
	case done == members:
		return store.BatchDone, nil
	case needsHuman:
		return store.BatchNeedsHuman, nil
	case started:
		return store.BatchRunning, nil
	}
	return store.BatchOpen, nil
}

// runBatchEvent derives the state of the batch a run belongs to, if any.
func (c *Controller) runBatchEvent(ctx context.Context, runID string) {
	run, err := c.store.Get(ctx, runID)
	if err != nil || run.ParentID == "" {
		return
	}
	c.deriveBatch(ctx, run.ParentID)
}

// sweepBatches re-derives every open batch each tick, catching changes
// made by paths that do not call runBatchEvent.
func (c *Controller) sweepBatches() {
	bs, err := c.store.List(c.ctx, store.Filter{Kind: store.KindBatch, Status: store.StatusOpen})
	if err != nil {
		return
	}
	for _, b := range bs {
		c.deriveBatch(c.ctx, b.ID)
	}
}

func (c *Controller) batchShow(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	b, err := c.store.Get(ctx, p.ID)
	if errors.Is(err, store.ErrNotFound) {
		proj, perr := c.project(sc, "")
		if perr != nil {
			return nil, err
		}
		b, err = store.OpenBatch(ctx, c.store, proj, p.ID)
	}
	if err != nil {
		return nil, err
	}
	if b.Kind != store.KindBatch {
		return nil, fmt.Errorf("%w: %s is a %s, not a batch", store.ErrInvalid, p.ID, b.Kind)
	}
	if !sc.admin && b.Project != sc.project {
		return nil, forbidden("session is scoped to project %s", sc.project)
	}
	res, err := c.batchResult(ctx, b)
	if err != nil {
		return nil, err
	}
	return res, c.batchBlocked(ctx, &res)
}

// batchBlocked fills BlockedBy: for each member run, the issues of other
// members whose open steps block the run's current step.
func (c *Controller) batchBlocked(ctx context.Context, res *rpc.BatchResult) error {
	issueOf := map[string]string{} // step id -> issue
	for _, r := range res.Runs {
		rf, err := store.DecodeFields[store.RunFields](r.Run)
		if err != nil {
			return err
		}
		for _, s := range r.Steps {
			issueOf[s.ID] = strings.TrimPrefix(rf.Key, "issue-")
		}
	}
	open := map[string]bool{}
	for _, r := range res.Runs {
		for _, s := range r.Steps {
			open[s.ID] = s.Status != store.StatusClosed
		}
	}
	res.BlockedBy = map[string][]string{}
	for _, r := range res.Runs {
		cur := currentStep(r.Steps)
		if cur == nil || r.Run.Status == store.StatusClosed {
			continue
		}
		edges, err := store.Edges(ctx, c.store, cur.ID)
		if err != nil {
			return err
		}
		for _, e := range edges {
			if e.From != cur.ID || !e.Kind.Blocking() || !open[e.To] {
				continue
			}
			if n, ok := issueOf[e.To]; ok && !slices.Contains(res.BlockedBy[r.Run.ID], n) {
				res.BlockedBy[r.Run.ID] = append(res.BlockedBy[r.Run.ID], n)
			}
		}
	}
	return nil
}

// currentStep is the first step not closed done or skipped, or nil.
func currentStep(steps []store.Ticket) *store.Ticket {
	for i := range steps {
		if !store.StepFinished(steps[i]) {
			return &steps[i]
		}
	}
	return nil
}

func (c *Controller) batchList(ctx context.Context, sc scope, p rpc.BatchListParams) (any, error) {
	if !sc.admin {
		p.Project = sc.project
	}
	return c.store.List(ctx, store.Filter{Project: p.Project, Kind: store.KindBatch, Status: store.StatusOpen, Order: store.OrderNewest, Limit: p.Limit})
}
