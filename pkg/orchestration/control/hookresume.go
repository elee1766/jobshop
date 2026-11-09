package control

import (
	"context"
	"encoding/json"
	"fmt"

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

// repinRun moves a callback workflow run to the live callback manifest: its
// pin, entries and the reference scope a callback may resolve. A deploy that
// rewrites the callbacks otherwise strands every open run, since the old
// manifest is gone. The caller holds runMu and has checked the admin socket.
// The run's progress (applied batches, completed steps) is kept; only the
// code and authority it runs under change, and the run says so.
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
	if hooks.RunPinned(lc.runtime.Programs, lc.runtime.Config, f.Workflow, a.Digest) && a.OnStart == w.OnStart && a.OnStepCompleted == w.OnStepCompleted {
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
	_, _ = store.AddComment(ctx, c.store, actor, t.ID, fmt.Sprintf("Repinned to the live callback manifest (%s, was %s) by %s.", pin, old, actor))
	c.log.InfoContext(ctx, "run repinned", "run", t.ID, "workflow", f.Workflow, "by", actor)
	return t, f, nil
}
