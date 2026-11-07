package control

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
)

// hookHost exposes controller policy and graph mechanics without exporting it.
type hookHost struct{ c *Controller }

func (c *Controller) hooks() *hooks.Engine {
	c.hooksOnce.Do(func() {
		c.hookEngine = hooks.New(c.store, hookHost{c}, c.root.Dir, c.now, c.log)
	})
	return c.hookEngine
}

func (h hookHost) CurrentHooks() *hooks.Runtime { return h.c.live.Load().runtime }

func (h hookHost) AdmitHook(ctx context.Context, runtime *hooks.Runtime, in hooks.Invocation) (string, error) {
	c := h.c
	if !c.hookReadSetCurrent(ctx, in.ReadSet) {
		return "superseded", nil
	}
	if in.Trigger == "" {
		return "", nil
	}
	tr, ok := runtime.Config.Triggers[in.Trigger]
	if !ok {
		return "superseded", nil
	}
	if in.Event != nil && in.Event.Source == "timer" {
		// The timer slot was accepted before the durable invocation existed.
		tr.When.Cron, tr.Cooldown = "", ""
	}
	st, err := c.evalTrigger(ctx, runtime.Config, in.Trigger, tr)
	if err != nil {
		return "", err
	}
	if !st.holds {
		return "held", nil
	}
	return "", nil
}

func (h hookHost) CommitHook(ctx context.Context, runtime *hooks.Runtime, key string, in hooks.Invocation, journal hooks.JournalPlan) (hooks.Invocation, error) {
	c := h.c
	if !c.hookReadSetCurrent(ctx, in.ReadSet) {
		return in, store.ErrConflict
	}
	lc := c.live.Load()
	if lc.runtime != runtime {
		return in, fmt.Errorf("%w: hook configuration changed", store.ErrConflict)
	}
	plans, err := c.prepareHookRuns(ctx, lc, in)
	if err != nil {
		return in, err
	}
	c.runMu.Lock()
	defer c.runMu.Unlock()
	graph, err := c.planHookActions(ctx, lc, key, in, runtime.Programs[in.Entry], plans)
	if err != nil {
		return in, err
	}
	in.Plan = &graph
	for _, ticket := range graph.Tickets {
		in.Targets = append(in.Targets, ticket.ID)
	}
	for _, plan := range plans {
		if existing, err := store.OpenRun(ctx, c.store, plan.params.Workflow, plan.params.Key); err == nil && !slices.Contains(in.Targets, existing.ID) {
			in.Targets = append(in.Targets, existing.ID)
		}
	}
	if err := journal(in); err != nil {
		return in, err
	}
	return in, c.applyPlannedHook(ctx, key, in)
}

func (h hookHost) ApplyHookPlan(ctx context.Context, key string, in hooks.Invocation) error {
	c := h.c
	c.runMu.Lock()
	defer c.runMu.Unlock()
	return c.applyPlannedHook(ctx, key, in)
}

func (h hookHost) HookFailed(ctx context.Context, key string, in hooks.Invocation) {
	if in.Run == "" {
		return
	}
	c := h.c
	c.runMu.Lock()
	t, err := c.store.Get(ctx, in.Run)
	if err == nil && t.Status != store.StatusClosed {
		f, decodeErr := store.DecodeFields[store.RunFields](t)
		if decodeErr == nil && f.State == store.RunRunning {
			_, _ = c.setRunState(ctx, "controller", t, f, store.RunNeedsHuman, "")
		}
	}
	c.runMu.Unlock()
	_, _ = store.AddComment(ctx, c.store, "controller", in.Run, "Automation invocation "+key+" failed: "+in.Error)
}

func (h hookHost) HookApplied(ctx context.Context, in hooks.Invocation) {
	if in.Run == "" {
		return
	}
	c := h.c
	c.wakeWaitsSubjects(runSubject(in.Run))
	if run, err := c.store.Get(ctx, in.Run); err == nil && run.Status == store.StatusClosed {
		c.removeRunDir(run.ID)
	}
	c.runBatchEvent(ctx, in.Run)
}

func (c *Controller) hookReadSetCurrent(ctx context.Context, readSet map[string]int64) bool {
	for id, revision := range readSet {
		t, err := c.store.Get(ctx, id)
		if err != nil || t.Revision != revision {
			return false
		}
	}
	return true
}

func (c *Controller) applyPlannedHook(ctx context.Context, key string, in hooks.Invocation) error {
	if in.Run != "" {
		run, err := c.store.Get(ctx, in.Run)
		if err != nil {
			return err
		}
		f, err := store.DecodeFields[store.RunFields](run)
		if err != nil {
			return err
		}
		if f.Automation != nil {
			for _, applied := range f.Automation.Applied {
				if applied == key {
					return nil
				}
			}
		}
	} else if len(in.Plan.Tickets) > 0 {
		// Stable IDs were journaled before Apply. The graph is all-or-nothing.
		if _, err := c.store.Get(ctx, in.Plan.Tickets[0].ID); err == nil {
			return nil
		} else if !errors.Is(err, store.ErrNotFound) {
			return err
		}
	}
	lc := c.live.Load()
	program := lc.hooks[in.Entry]
	if program == nil {
		return errors.New("planned hook removed from active configuration")
	}
	expected := program.Digest()
	if in.Run != "" {
		if in.RunInput == nil {
			return errors.New("planned workflow hook has no accepted run input")
		}
		expected = hooks.WorkflowDigest(lc.hooks, lc.cfg.Workflows[in.RunInput.Workflow])
	}
	if expected != in.Digest {
		return errors.New("planned hook manifest changed: restore the accepted configuration before applying actions")
	}
	for _, ticket := range in.Plan.Tickets {
		if ticket.Kind != store.KindRun {
			continue
		}
		var fields store.RunFields
		if err := hooks.DecodeJSON(ticket.Fields, &fields); err != nil {
			return err
		}
		if fields.Key == "" {
			continue
		}
		existing, err := store.OpenRun(ctx, c.store, fields.Workflow, fields.Key)
		if err == nil && existing.ID != ticket.ID {
			return errors.New("a run with the planned workflow and key was created before this batch committed")
		}
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
	}
	_, err := c.store.Apply(ctx, "automation:"+key, *in.Plan)
	if err != nil {
		return fmt.Errorf("apply hook actions: %w", err)
	}
	return nil
}
