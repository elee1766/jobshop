package control

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/hooks"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// The caller holds runMu. The engine rejects unknown or unsafe external outcomes.
func (c *Controller) resumeHook(ctx context.Context, actor string, run domain.Ticket, fields domain.RunFields) (bool, error) {
	if err := runScopeRecorded(fields.Automation); err != nil {
		return false, err
	}
	key, err := c.hooks().Resume(ctx, run.ID, fields.Workflow, fields.Automation.Digest)
	if err != nil || key == "" {
		return false, err
	}
	if _, err := c.setRunState(ctx, actor, run, fields, domain.RunRunning, ""); err != nil {
		return false, err
	}
	_, _ = store.AddComment(ctx, c.store, actor, run.ID, "Resumed automation invocation "+key+" with recorded completed calls.")
	return true, nil
}

// runOnLive reports whether a callback run already runs under the live
// manifest: its pin, entry points and recorded reference scope match the
// active configuration. The scope is not part of the pin, so a reload that
// only adds or removes a pool or workflow reference also moves the run.
func (lc *liveConfig) runOnLive(f domain.RunFields) bool {
	w, ok := lc.cfg.Workflows[f.Workflow]
	a := f.Automation
	if !ok || a == nil || !hooks.RunPinned(lc.runtime.Programs, lc.runtime.Config, f.Workflow, a.Digest) || a.OnStart != w.OnStart || a.OnStepCompleted != w.OnStepCompleted || a.Scope == nil {
		return false
	}
	live := lc.cfg.HookScope("workflows." + f.Workflow)
	return maps.Equal(a.Scope.Pools, live.Pools) && maps.Equal(a.Scope.Workflows, live.Workflows)
}

// followLive moves an open run of a workflow that follows the live
// callbacks (no pin_callbacks, JOB-188) to the live manifest, so a reload
// never strands it. It returns the run as it stands after. A pinned
// workflow, or a run already on the live manifest, is returned as it is.
// The caller holds neither runMu nor any other lock.
func (c *Controller) followLive(ctx context.Context, run domain.Ticket, f domain.RunFields) (domain.Ticket, domain.RunFields, error) {
	lc := c.live.Load()
	if f.Automation == nil || !hooks.FollowsLive(lc.cfg, f.Workflow) || lc.runOnLive(f) {
		return run, f, nil
	}
	c.runMu.Lock()
	defer c.runMu.Unlock()
	// The run may have moved since the caller read it.
	t, err := c.store.Get(ctx, run.ID)
	if err != nil {
		return run, f, err
	}
	cur, err := domain.DecodeFields[domain.RunFields](t)
	if err != nil {
		return t, cur, err
	}
	return c.followLiveLocked(ctx, t, cur)
}

// followLiveLocked is followLive for a caller that holds runMu.
func (c *Controller) followLiveLocked(ctx context.Context, run domain.Ticket, f domain.RunFields) (domain.Ticket, domain.RunFields, error) {
	lc := c.live.Load()
	if f.Automation == nil || run.Status == domain.StatusClosed || !hooks.FollowsLive(lc.cfg, f.Workflow) || lc.runOnLive(f) {
		return run, f, nil
	}
	return c.repinRun(ctx, "controller", run, f)
}

// repinRun moves a callback workflow run to the live callback manifest: its
// pin, entries and the reference scope a callback may resolve. A pinned
// workflow's open runs otherwise stay stranded by a deploy, since the old
// manifest is gone; a workflow that follows the live callbacks calls this
// itself (followLive). The caller holds runMu, and for a pinned workflow has
// checked the admin socket. The run's progress (applied batches, completed
// steps) is kept; only the code and authority it runs under change, and the
// run says so.
func (c *Controller) repinRun(ctx context.Context, actor string, t domain.Ticket, f domain.RunFields) (domain.Ticket, domain.RunFields, error) {
	a := f.Automation
	if a == nil {
		return t, f, fmt.Errorf("%w: run %s is not a callback workflow run", domain.ErrInvalid, t.ID)
	}
	lc := c.live.Load()
	w, ok := lc.cfg.Workflows[f.Workflow]
	pin := lc.cfg.HookRunPins[f.Workflow]
	if !ok || w.OnStart == "" || pin == "" {
		return t, f, fmt.Errorf("%w: workflow %s is no longer a loaded callback workflow, so run %s cannot be repinned: cancel it", domain.ErrInvalid, f.Workflow, t.ID)
	}
	if lc.runOnLive(f) {
		return t, f, nil // already on the live manifest
	}
	old := a.Digest
	next := *a
	next.OnStart, next.OnStepCompleted, next.Digest = w.OnStart, w.OnStepCompleted, pin
	next.Scope = lc.cfg.HookScope("workflows." + f.Workflow)
	f.Automation = &next
	b, err := json.Marshal(f)
	if err != nil {
		return t, f, err
	}
	t, err = c.store.Update(ctx, actor, t.ID, store.Update{Fields: b, IfRevision: t.Revision})
	if err != nil {
		return t, f, err
	}
	detail := fmt.Sprintf("%s, was %s", pin, old)
	if pin == old {
		detail = pin + ", reference scope updated"
	}
	_, _ = store.AddComment(ctx, c.store, actor, t.ID, fmt.Sprintf("Repinned to the live callback manifest (%s) by %s.", detail, actor))
	c.log.InfoContext(ctx, "run repinned", "run", t.ID, "workflow", f.Workflow, "by", actor)
	return t, f, nil
}
