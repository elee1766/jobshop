package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/orchestration/linearsvc"
	"github.com/elee1766/jobshop/pkg/orchestration/threads"
	"github.com/elee1766/jobshop/pkg/orchestration/workflow"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Waits (#381). A task or step with fields.wait is out of the ready
// frontier. The controller evaluates its conditions on the subscription
// poll interval and when the ingest drain moves a webhook delivery, then
// clears the wait: the ticket becomes ready, pre-assigned to the session
// that parked it when resume is set. A PR that merges, closes or moves
// its head invalidates review and checks conditions bound to the old
// head; that also clears the wait, with the reason, so the agent decides
// again. Past its timeout a wait is labeled needs-human once.

// linearWaitSource reads a Linear issue for linear.state and
// linear.label. Tests replace it.
type linearWaitSource interface {
	IssueContext(ctx context.Context, key string) (linear.Context, error)
}

// liveLinearWait reads through the identity covering the issue's team.
type liveLinearWait struct{ c *Controller }

func (l liveLinearWait) IssueContext(ctx context.Context, key string) (linear.Context, error) {
	var is linear.Context
	err := l.c.linearSvc().DoKey(ctx, key, func(ctx context.Context, cl *linear.Client) error {
		var err error
		is, err = cl.GetContext(ctx, key, false)
		return err
	})
	return is, err
}

// waitEvery is the poll interval: the subscriptions interval when set.
func (c *Controller) waitEvery() time.Duration {
	if s := c.config().Subscriptions; s != nil {
		return config.Duration(s.Interval, threads.DefaultInterval)
	}
	return threads.DefaultInterval
}

func (c *Controller) waitSrc() workflow.Source {
	if c.waitSource != nil {
		return c.waitSource
	}
	return github.CLI{}
}

func (c *Controller) linearWaitSrc() linearWaitSource {
	if c.linearWait != nil {
		return c.linearWait
	}
	return liveLinearWait{c: c}
}

// waitOf returns fields.wait of a task or step, nil for none.
func waitOf(t domain.Ticket) *domain.Wait { return domain.WaitOf(t) }

// withWait returns t's fields with wait set (nil removes it), keeping
// the other fields.
func withWait(t domain.Ticket, w *domain.Wait) (json.RawMessage, error) {
	m := map[string]json.RawMessage{}
	if len(t.Fields) > 0 {
		if err := json.Unmarshal(t.Fields, &m); err != nil {
			return nil, err
		}
	}
	if w == nil {
		delete(m, "wait")
	} else {
		b, err := json.Marshal(w)
		if err != nil {
			return nil, err
		}
		m["wait"] = b
	}
	return json.Marshal(m)
}

// waitAdd parks a ticket. A held ticket is released as a deferral, not a
// failed attempt; with resume the wait remembers the session.
func (c *Controller) waitAdd(ctx context.Context, sc scope, p rpc.WaitAddParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindTask && t.Kind != domain.KindStep {
		return nil, fmt.Errorf("%w: only tasks and steps wait (%s is a %s)", domain.ErrInvalid, id, t.Kind)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: %s", domain.ErrClosed, id)
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("%s is in another project", id)
	}
	var held *store.Claim
	if t.Status == domain.StatusInProgress {
		cl, err := c.heldClaim(ctx, sc, id)
		if err != nil {
			return nil, err
		}
		held = &cl
	}
	now := c.now().UTC()
	w := &domain.Wait{Conds: p.Conds, Reason: strings.TrimSpace(p.Reason), Owner: p.Owner, Resume: p.Resume, Since: now, By: sc.actor}
	if p.Any {
		w.Mode = domain.WaitAny
	}
	if p.Timeout != nil {
		if !p.Timeout.After(now) {
			return nil, fmt.Errorf("%w: timeout must be in the future", domain.ErrInvalid)
		}
		w.Timeout = p.Timeout.UTC()
	}
	if p.Resume {
		w.Session = t.Holder
		if w.Session == "" {
			w.Session = sc.session
		}
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	for _, cd := range w.Conds {
		if cd.Type == domain.WaitDeploy && c.opt.Commit == "" {
			return nil, fmt.Errorf("%w: deploy wait: this controller reports no build commit, so it cannot tell what is deployed", domain.ErrInvalid)
		}
	}
	fields, err := withWait(t, w)
	if err != nil {
		return nil, err
	}
	if t, err = c.store.Update(ctx, sc.actor, id, store.Update{Fields: fields, IfRevision: t.Revision}); err != nil {
		return nil, err
	}
	if held != nil {
		held.Fence = t.Fence
		// A release with until is a deferral: not a failed attempt.
		if t, err = c.store.Release(ctx, "", held.TicketID, held, store.Outcome{Reason: "waiting: " + w.Reason, Until: now}); err != nil {
			return nil, err
		}
	}
	c.wakeWaits()
	return t, nil
}

// waitList lists waiting tickets.
func (c *Controller) waitList(ctx context.Context, sc scope, p rpc.WaitListParams) (any, error) {
	proj := p.Project
	if !sc.admin {
		proj = sc.project
	}
	ts, err := c.waiting(ctx, proj)
	if err != nil {
		return nil, err
	}
	out := []rpc.WaitItem{}
	for _, t := range ts {
		out = append(out, rpc.WaitItem{ID: t.ID, Project: t.Project, Title: t.Title, Wait: *waitOf(t)})
	}
	return out, nil
}

// waitClear removes a wait by hand.
func (c *Controller) waitClear(ctx context.Context, sc scope, p rpc.WaitClearParams) (any, error) {
	if p.ID == "" {
		return nil, fmt.Errorf("%w: ticket id is required", domain.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("%s is in another project", p.ID)
	}
	w := waitOf(t)
	if w == nil {
		return t, nil
	}
	why := "cleared by " + sc.actor
	if p.Reason != "" {
		why += ": " + p.Reason
	}
	return c.clearWait(ctx, t, w, why)
}

// waiting is every open task or step with a wait, in project (empty: all).
func (c *Controller) waiting(ctx context.Context, project string) ([]domain.Ticket, error) {
	var out []domain.Ticket
	for _, k := range []domain.Kind{domain.KindTask, domain.KindStep} {
		ts, err := c.store.List(ctx, store.Filter{Project: project, Kind: k, Status: domain.StatusOpen, Limit: math.MaxInt32})
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			if waitOf(t) != nil {
				out = append(out, t)
			}
		}
	}
	return out, nil
}

// clearWait removes the wait, notes why on the ticket and, with resume,
// pre-assigns it to the session that parked it while that is open and
// idle. A busy one leaves the ticket in its pool (JOB-120).
func (c *Controller) clearWait(ctx context.Context, t domain.Ticket, w *domain.Wait, why string) (domain.Ticket, error) {
	fields, err := withWait(t, nil)
	if err != nil {
		return t, err
	}
	// One Update so no reader sees the wait cleared without its comment
	// (JOB-39).
	u := store.Update{
		Fields:     fields,
		IfRevision: t.Revision,
		Comment:    fmt.Sprintf("Wait (%s) ended: %s", w.Reason, why),
	}
	if w.TimedOut && slices.Contains(t.Labels, needsHumanLabel) {
		u.RemoveLabels = []string{needsHumanLabel}
	}
	if w.Resume && w.Session != "" && t.Holder == "" {
		switch holder, busy := c.resumeHolder(ctx, w.Session); {
		case holder != "":
			u.Holder = &holder
		case busy != "":
			u.Comment += fmt.Sprintf("\n\nSession %s is busy on %s, so this goes back to pool %s for any idle session.", w.Session, busy, t.Pool)
		}
	}
	if t, err = c.store.Update(ctx, "controller", t.ID, u); err != nil {
		return t, err
	}
	c.log.InfoContext(ctx, "wait cleared", "ticket", t.ID, "why", why)
	return t, nil
}

// resumeHolder is the session a resumed ticket is pre-assigned to: the
// parking session, while it is open and holds no claimed ticket. A session
// claims its in-progress work before anything pre-assigned, so a busy one
// would leave the resumed ticket behind unrelated work for as long as that
// takes (JOB-120). Then holder is empty and busy names the work; the
// ticket stays in its pool, and the PR and the ticket's comments carry
// the context to whichever session claims it.
func (c *Controller) resumeHolder(ctx context.Context, id string) (holder, busy string) {
	s, err := c.store.Get(ctx, id)
	if err != nil || s.Kind != domain.KindSession || s.Status == domain.StatusClosed {
		return "", ""
	}
	held, err := c.store.List(ctx, store.Filter{Holder: id, Status: domain.StatusInProgress, Limit: 1})
	if err != nil {
		return "", "" // unknown: the pool is never wrong, only slower
	}
	if len(held) > 0 {
		return "", held[0].ID
	}
	return id, ""
}

// pollWaits runs one pass over every waiting ticket.
func (c *Controller) pollWaits(ctx context.Context, src workflow.Source, ths threads.Source) {
	c.waitMu.Lock()
	defer c.waitMu.Unlock()
	ts, err := c.waiting(ctx, "")
	if err != nil {
		if ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: list", "err", err)
		}
		return
	}
	// The pass starts after the list: every observation it caches is newer
	// than the snapshots it is applied to (JOB-144).
	pass := c.newWaitPass(src, ths)
	for _, t := range ts {
		if err := c.evalWait(ctx, pass, t); err != nil && ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: evaluate", "ticket", t.ID, "err", err)
		}
	}
}

// evalWait evaluates one ticket's wait with the engine and applies the
// decision. t is the snapshot the decision was made from: every write
// below is at t's revision, so a wait that was cleared, replaced or
// re-parked since the snapshot (by a person, a release, an agent) makes
// the write fail with store.ErrConflict instead of acting on a stale
// observation. The next pass or wake evaluates the current wait.
func (c *Controller) evalWait(ctx context.Context, pass *workflow.Pass, t domain.Ticket) error {
	w := waitOf(t)
	if w == nil {
		return nil
	}
	d := workflow.Evaluate(ctx, pass, *w, c.now().UTC())
	if d.Clear {
		_, err := c.clearWait(ctx, t, w, d.Why)
		return err
	}
	var errs []error
	for _, err := range d.Errs {
		// A Linear outage is logged once by linearCall, not per wait.
		if !linearsvc.Down(err) {
			errs = append(errs, err)
		}
	}
	if d.Save {
		fields, err := withWait(t, &d.Wait)
		if err != nil {
			return err
		}
		if t, err = c.store.Update(ctx, "controller", t.ID, store.Update{Fields: fields, IfRevision: t.Revision}); err != nil {
			return err
		}
	}
	if d.TimedOut {
		w = &d.Wait
		if err := store.AddLabel(ctx, c.store, "controller", t.ID, needsHumanLabel); err != nil {
			return err
		}
		owner := w.Owner
		if owner == "" {
			owner = "unset"
		}
		_, _ = store.AddComment(ctx, c.store, "controller", t.ID, fmt.Sprintf("Wait (%s) passed its timeout %s; owner %s. Clear it with `jobshop wait clear %s`.", w.Reason, w.Timeout.Format(time.RFC3339), owner, t.ID))
		c.log.WarnContext(ctx, "wait timed out", "ticket", t.ID, "reason", w.Reason, "owner", w.Owner)
		c.holdChat(ctx, t.ID)
	}
	return errors.Join(errs...)
}

// waitSources is what a pass observes through: the gh and Linear
// adapters, this controller's runs and subscription filters, and its
// build commit.
func (c *Controller) waitSources(src workflow.Source, ths threads.Source) workflow.Sources {
	return workflow.Sources{
		GitHub:  src,
		Linear:  c.linearWaitSrc(),
		Runs:    runWaitSource{c: c},
		Replies: threads.Replies{E: c.threadEngine(), Src: ths},
		Commit:  c.opt.Commit,
	}
}

func (c *Controller) newWaitPass(src workflow.Source, ths threads.Source) *workflow.Pass {
	return workflow.NewPass(c.waitSources(src, ths))
}

// runWaitSource reads run tickets for the run condition.
type runWaitSource struct{ c *Controller }

func (r runWaitSource) Run(ctx context.Context, id string) (workflow.RunState, error) {
	t, err := r.c.store.Get(ctx, id)
	if err != nil {
		return workflow.RunState{}, err
	}
	f, err := domain.DecodeFields[domain.RunFields](t)
	if err != nil {
		return workflow.RunState{}, err
	}
	return workflow.RunState{Closed: t.Status == domain.StatusClosed, State: f.State}, nil
}
