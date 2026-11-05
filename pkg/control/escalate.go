package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Escalation (#90). The store counts step failures inside the failing
// transaction and closes a step failed at its retry limit. The controller
// then moves the run to needs_human and files one escalation task in the
// workflow's escalate_pool. Both happen on the triggering call and on
// every tick, so a crash between the two writes is repaired.

// runMetaPack holds the escalation ticket's link to its run.
const runMetaPack = "run"

// escalationComments is how many recent comments per step the escalation
// ticket body carries.
const escalationComments = 3

// stepEvent runs after a call that may have counted a step failure.
func (c *Controller) stepEvent(ctx context.Context, step store.Ticket) {
	if step.Kind != store.KindStep || step.ParentID == "" {
		return
	}
	c.checkRun(ctx, step.ParentID)
}

// stepEvents is stepEvent for ticket ids (release-all and lease sweep).
func (c *Controller) stepEvents(ctx context.Context, ids []string) {
	for _, id := range ids {
		if t, err := c.store.Get(ctx, id); err == nil {
			c.stepEvent(ctx, t)
		}
	}
}

// checkRun escalates a running run with an exhausted step, then completes
// it when every step closed done.
func (c *Controller) checkRun(ctx context.Context, runID string) {
	// Gated steps are decided first: a skip may finish the run.
	c.gateSteps(ctx, runID)
	if !c.escalateRun(ctx, runID) {
		c.completeRun(ctx, runID)
	}
	c.runBatchEvent(ctx, runID)
}

// exhaustedStep is a step closed failed at its retry limit.
func exhaustedStep(s store.Ticket) bool {
	if s.Status != store.StatusClosed || s.Resolution != "failed" {
		return false
	}
	f, err := store.DecodeFields[store.StepFields](s)
	return err == nil && f.Failures >= f.Retries
}

// escalateRun sets needs_human and files the escalation ticket when a step
// of a running run is exhausted. It reports whether the run needs a human.
func (c *Controller) escalateRun(ctx context.Context, runID string) bool {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	run, err := c.store.Get(ctx, runID)
	if err != nil || run.Kind != store.KindRun || run.Status == store.StatusClosed {
		return false
	}
	f, err := store.DecodeFields[store.RunFields](run)
	if err != nil {
		return false
	}
	if f.State == store.RunNeedsHuman {
		// Already escalated; the ticket may still be missing after a crash.
		c.ensureEscalation(ctx, run, f)
		return true
	}
	if f.State != store.RunRunning {
		return false
	}
	steps, err := store.Children(ctx, c.store, runID)
	if err != nil {
		return false
	}
	var failed *store.Ticket
	for i := range steps {
		if exhaustedStep(steps[i]) {
			failed = &steps[i]
			break
		}
	}
	if failed == nil {
		return false
	}
	f.Escalations++
	if run, err = c.setRunState(ctx, "controller", run, f, store.RunNeedsHuman, ""); err != nil {
		c.log.Error("run needs_human", "run", runID, "err", err)
		return false
	}
	f.State = store.RunNeedsHuman
	c.log.Warn("run needs a human", "run", runID, "workflow", f.Workflow, "step", failed.ID)
	c.ensureEscalation(ctx, run, f)
	return true
}

// ensureEscalation creates the escalation ticket for the run's current
// needs_human episode unless it exists, so retries after a crash and
// repeated sweeps file exactly one.
func (c *Controller) ensureEscalation(ctx context.Context, run store.Ticket, f store.RunFields) {
	key := fmt.Sprintf("%s/%d", run.ID, f.Escalations)
	seen, err := c.store.List(ctx, store.Filter{MetaPack: runMetaPack, MetaKey: "escalation", MetaValue: key, Limit: 1})
	if err != nil || len(seen) > 0 {
		return
	}
	cfg := c.config()
	pool := ""
	if w, ok := cfg.Workflows[f.Workflow]; ok {
		pool = w.EscalatePool
	}
	body, err := c.escalationBody(ctx, run, f)
	if err != nil {
		c.log.Error("escalation body", "run", run.ID, "err", err)
		return
	}
	meta := map[string]json.RawMessage{}
	if len(run.Meta) > 0 {
		_ = json.Unmarshal(run.Meta, &meta)
	}
	meta[runMetaPack], _ = json.Marshal(map[string]string{"id": run.ID, "workflow": f.Workflow, "escalation": key})
	mb, err := json.Marshal(meta)
	if err != nil {
		return
	}
	prio := 1
	t, err := c.store.Create(ctx, "controller", store.NewTicket{
		Project: run.Project, Kind: store.KindTask, Title: "Escalate run " + run.ID,
		Body: body, Priority: &prio, Pool: pool, Labels: []string{"needs-human"}, Meta: mb,
	})
	if err != nil {
		c.log.Error("create escalation", "run", run.ID, "err", err)
		return
	}
	c.log.Info("escalation filed", "run", run.ID, "ticket", t.ID, "pool", pool)
	c.record(ctx, store.LearningFailure, run, fmt.Sprintf("run %s (workflow %s) escalated to a human: a step failed at its retry limit (escalation %s)", run.ID, f.Workflow, t.ID), "")
	c.ensureEscalationChat(ctx, t)
}

// escalationBody lists the run and, per step, its attempts, resolution and
// last comments, so the escalation agent need not dig.
func (c *Controller) escalationBody(ctx context.Context, run store.Ticket, f store.RunFields) (string, error) {
	steps, err := store.Children(ctx, c.store, run.ID)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Run %s (workflow %s) needs a human: a step failed at its retry limit.\n", run.ID, f.Workflow)
	if len(f.Vars) > 0 {
		fmt.Fprintf(&b, "Vars: %v\n", f.Vars)
	}
	fmt.Fprintf(&b, "Inspect with `jobshop run show %s`; after fixing the blocker, `jobshop run resume %s`.\n", run.ID, run.ID)
	for _, s := range steps {
		sf, err := store.DecodeFields[store.StepFields](s)
		if err != nil {
			return "", err
		}
		res := s.Resolution
		if res == "" {
			res = string(s.Status)
		}
		// Controller-run steps have no ticket pool, so no session claims
		// them (#396); name what runs them instead of an empty pool (#669).
		where := "pool " + s.Pool
		if s.Pool == "" && sf.Type != "" {
			where = sf.Type + " step, run by the controller"
		}
		fmt.Fprintf(&b, "\n## Step %d %s (%s, %s): %s, failures %d/%d\n", sf.Index, sf.Step, s.ID, where, res, sf.Failures, sf.Retries)
		comments, err := c.store.Comments(ctx, s.ID)
		if err != nil {
			return "", err
		}
		if len(comments) > escalationComments {
			comments = comments[len(comments)-escalationComments:]
		}
		for _, cm := range comments {
			body := cm.Body
			if len(body) > promptCommentSize {
				body = body[:promptCommentSize] + " [truncated]"
			}
			fmt.Fprintf(&b, "- [%s] %s: %s\n", cm.CreatedAt.UTC().Format("2006-01-02 15:04"), cm.Author, body)
		}
	}
	return b.String(), nil
}

// recordRunGauges sets jobshop.runs.needs_human per project.
func (c *Controller) recordRunGauges(ctx context.Context) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunNeedsHuman, Order: store.OrderNewest})
	if err != nil {
		return
	}
	counts := map[string]int64{}
	for name := range c.config().Projects {
		counts[name] = 0
	}
	for _, r := range runs {
		if r.Status != store.StatusClosed {
			counts[r.Project]++
		}
	}
	for p, n := range counts {
		c.metrics.RunsNeedsHuman.Record(ctx, n, metric.WithAttributes(telemetry.AttrProject.String(p)))
	}
}
