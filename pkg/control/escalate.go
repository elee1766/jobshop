package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/chatlife"
	"github.com/elee1766/jobshop/pkg/domain"
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
func (c *Controller) stepEvent(ctx context.Context, cfg *config.Config, step domain.Ticket) {
	if step.Kind != domain.KindStep || step.ParentID == "" {
		return
	}
	c.checkRun(ctx, cfg, step.ParentID)
}

// stepEvents is stepEvent for ticket ids (release-all and lease sweep).
func (c *Controller) stepEvents(ctx context.Context, cfg *config.Config, ids []string) {
	for _, id := range ids {
		if t, err := c.store.Get(ctx, id); err == nil {
			c.stepEvent(ctx, cfg, t)
		}
	}
}

// checkRun escalates a running run with an exhausted step, then completes
// it when every step closed done.
func (c *Controller) checkRun(ctx context.Context, cfg *config.Config, runID string) {
	// Gated steps are decided first: a skip may finish the run.
	c.gateSteps(ctx, runID)
	c.escalateFailedSteps(ctx, cfg, runID)
	if !c.escalateRun(ctx, cfg, runID) {
		c.completeRun(ctx, runID)
	}
	c.runBatchEvent(ctx, runID)
	c.kickHooks()
}

// escalateRun sets needs_human and files the escalation ticket when a step
// of a running run is exhausted. It reports whether the run needs a human.
func (c *Controller) escalateRun(ctx context.Context, cfg *config.Config, runID string) bool {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	run, err := c.store.Get(ctx, runID)
	if err != nil || run.Kind != domain.KindRun || run.Status == domain.StatusClosed {
		return false
	}
	f, err := domain.DecodeFields[domain.RunFields](run)
	if err != nil {
		return false
	}
	if f.State == domain.RunNeedsHuman {
		// Already escalated; the ticket may still be missing after a crash.
		c.ensureEscalation(ctx, cfg, run, f)
		return true
	}
	if f.State != domain.RunRunning {
		return false
	}
	steps, err := store.Children(ctx, c.store, runID)
	if err != nil {
		return false
	}
	var failed *domain.Ticket
	for i := range steps {
		if domain.StepExhausted(steps[i]) {
			failed = &steps[i]
			break
		}
	}
	if failed == nil {
		return false
	}
	f.Escalations++
	if run, err = c.setRunState(ctx, "controller", run, f, domain.RunNeedsHuman, ""); err != nil {
		c.log.ErrorContext(ctx, "run needs_human", "run", runID, "err", err)
		return false
	}
	f.State = domain.RunNeedsHuman
	c.log.WarnContext(ctx, "run needs a human", "run", runID, "workflow", f.Workflow, "step", failed.ID)
	c.ensureEscalation(ctx, cfg, run, f)
	return true
}

// ensureEscalation creates the escalation ticket for the run's current
// needs_human episode unless it exists, so retries after a crash and
// repeated sweeps file exactly one.
func (c *Controller) ensureEscalation(ctx context.Context, cfg *config.Config, run domain.Ticket, f domain.RunFields) {
	key := fmt.Sprintf("%s/%d", run.ID, f.Escalations)
	seen, err := c.store.List(ctx, store.Filter{MetaPack: runMetaPack, MetaKey: "escalation", MetaValue: key, Limit: 1})
	if err != nil || len(seen) > 0 {
		return
	}
	pool := ""
	if w, ok := cfg.Workflows[f.Workflow]; ok {
		pool = w.EscalatePool
	}
	body, err := c.escalationBody(ctx, cfg, run, f)
	if err != nil {
		c.log.ErrorContext(ctx, "escalation body", "run", run.ID, "err", err)
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
	t, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: run.Project, Kind: domain.KindTask, Title: "Escalate run " + run.ID,
		Body: body, Priority: &prio, Pool: pool, Labels: []string{"needs-human"}, Meta: mb,
	})
	if err != nil {
		c.log.ErrorContext(ctx, "create escalation", "run", run.ID, "err", err)
		return
	}
	c.log.InfoContext(ctx, "escalation filed", "run", run.ID, "ticket", t.ID, "pool", pool)
	c.record(ctx, domain.LearningFailure, run, fmt.Sprintf("run %s (workflow %s) escalated to a human: a step failed at its retry limit (escalation %s)", run.ID, f.Workflow, t.ID), "")
	c.ensureEscalationChat(ctx, cfg, t)
}

// escalationBody lists the run and, per step, its attempts, resolution and
// last comments, so the escalation agent need not dig.
func (c *Controller) escalationBody(ctx context.Context, cfg *config.Config, run domain.Ticket, f domain.RunFields) (string, error) {
	steps, err := store.Children(ctx, c.store, run.ID)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Run %s (workflow %s) needs a human: %s.\n", run.ID, f.Workflow, chatlife.Reason(steps))
	if len(f.Vars) > 0 {
		fmt.Fprintf(&b, "Vars: %v\n", f.Vars)
	}
	fmt.Fprintf(&b, "Inspect with `jobshop run show %s`; after fixing the blocker, `jobshop run resume %s`.\n", run.ID, run.ID)
	if f.Automation != nil {
		// Legacy full-manifest pins also differ here, which is right: a
		// repin moves them to the run pin.
		if pin := cfg.HookRunPins[f.Workflow]; pin != "" && pin != f.Automation.Digest {
			fmt.Fprintf(&b, "The run's pinned callbacks no longer match the loaded configuration (a deploy or reload changed them). To continue it under the live callbacks, `jobshop run resume %s --repin` on the admin socket; or cancel it.\n", run.ID)
		}
	}
	for _, s := range steps {
		sf, err := domain.DecodeFields[domain.StepFields](s)
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
		fmt.Fprintf(&b, "\n## Step %d %s (%s, %s): %s, failures %d/%d", sf.Index, sf.Step, s.ID, where, res, sf.Failures, sf.Retries)
		if sf.ReviewRounds > 0 {
			fmt.Fprintf(&b, ", review rounds %d/%d", sf.ReviewRounds, sf.ReviewRoundLimit())
		}
		b.WriteString("\n")
		comments, err := store.Comments(ctx, c.store, s.ID)
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
func (c *Controller) recordRunGauges(ctx context.Context, cfg *config.Config) {
	runs, err := c.store.List(ctx, store.Filter{Kind: domain.KindRun, State: domain.RunNeedsHuman, Order: store.OrderNewest})
	if err != nil {
		return
	}
	counts := map[string]int64{}
	for name := range cfg.Projects {
		counts[name] = 0
	}
	for _, r := range runs {
		if r.Status != domain.StatusClosed {
			counts[r.Project]++
		}
	}
	for p, n := range counts {
		c.metrics.RunsNeedsHuman.Record(ctx, n, metric.WithAttributes(telemetry.AttrProject.String(p)))
	}
}
