package control

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
)

// enqueueHookEvents runs before acknowledging ingress. Re-delivery is harmless.
func (c *Controller) enqueueHookEvents(ctx context.Context, items []store.InboxItem) error {
	lc := c.live.Load()
	for _, it := range items {
		ev := eventOf(it)
		for _, name := range slices.Sorted(maps.Keys(lc.cfg.Triggers)) {
			tr := lc.cfg.Triggers[name]
			if tr.Handler == "" || !triggerOn(tr, ev) {
				continue
			}
			program := lc.hooks[tr.Handler]
			if program == nil {
				return fmt.Errorf("trigger %s has no loaded hook", name)
			}
			payload := it.Admitted
			if len(payload) == 0 || string(payload) == "null" {
				payload = it.Payload
			}
			var object map[string]any
			if err := hooks.DecodeJSON(payload, &object); err != nil {
				return fmt.Errorf("trigger %s: payload must be a JSON object: %w", name, err)
			}
			event := &hooks.Event{Source: it.Source, Event: it.Kind, Delivery: it.DeliveryID, Repository: it.RepoOrTeam, Trigger: name, Payload: object, Fields: ev.vars()}
			if err := c.hooks().Put(ctx, "event/"+name+"/"+it.Source+"/"+it.DeliveryID, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: program.Digest(), Event: event, Args: tr.Args}); err != nil {
				return err
			}
		}
	}
	c.kickHooks()
	return nil
}

func stepHookInput(t store.Ticket) (hooks.Step, error) {
	f, err := store.DecodeFields[store.StepFields](t)
	return hooks.Step{ID: f.Step, TicketID: t.ID, Pool: t.Pool, Status: string(t.Status), Resolution: t.Resolution, Outcome: f.Outcome, Output: f.StructuredOutput, Inputs: f.Inputs, Failures: f.Failures, Retries: f.Retries}, err
}

func (c *Controller) enqueueRunHooks(ctx context.Context) error {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunRunning, Limit: store.NoLimit})
	if err != nil {
		return err
	}
	for _, run := range runs {
		f, err := store.DecodeFields[store.RunFields](run)
		if err != nil {
			return err
		}
		a := f.Automation
		if a == nil || run.Status == store.StatusClosed {
			continue
		}
		children, err := store.Children(ctx, c.store, run.ID)
		if err != nil {
			return err
		}
		input := &hooks.Run{ID: run.ID, Workflow: f.Workflow, Key: f.Key, Project: run.Project, Inputs: f.Inputs, Vars: f.Vars, State: f.State, Steps: []hooks.Step{}}
		readSet := map[string]int64{run.ID: run.Revision}
		for _, child := range children {
			readSet[child.ID] = child.Revision
			if child.Kind != store.KindStep {
				continue
			}
			step, err := stepHookInput(child)
			if err != nil {
				return err
			}
			input.Steps = append(input.Steps, step)
		}
		in := hooks.Invocation{Run: run.ID, RunInput: input, ReadSet: readSet, Digest: a.Digest, ParentInvocation: metaString(run.Meta, "automation", "invocation")}
		key := run.ID + "/start"
		if a.Version == 0 {
			in.Entry = a.OnStart
		} else {
			for _, child := range children {
				if child.Kind != store.KindStep || !store.StepFinished(child) || a.Completed[child.ID] == child.Revision {
					continue
				}
				step, err := stepHookInput(child)
				if err != nil {
					return err
				}
				in.Entry, in.Step, in.StepRevision, in.StepInput = a.OnStepCompleted, child.ID, child.Revision, &step
				key = run.ID + "/step/" + child.ID + "/" + strconv.FormatInt(child.Revision, 10) + "/" + strconv.FormatInt(a.Version, 10)
				break
			}
		}
		if in.Entry != "" {
			if err := c.hooks().Put(ctx, key, in); err != nil {
				return err
			}
		}
	}
	return nil
}

// One bounded worker evaluates hooks without tickMu, runMu, or controller mu.
func (c *Controller) kickHooks() {
	if c.ctx == nil || c.ctx.Err() != nil {
		return
	}
	c.hookPending.Store(true)
	if !c.hookRunning.CompareAndSwap(false, true) {
		return
	}
	c.wg.Go(func() {
		c.hookPending.Store(false)
		defer func() {
			c.hookRunning.Store(false)
			if c.hookPending.Swap(false) {
				c.kickHooks()
			}
		}()
		if err := c.enqueueRunHooks(c.ctx); err != nil {
			c.log.ErrorContext(c.ctx, "enqueue workflow hooks", "err", err)
			return
		}
		for range 64 {
			processed, err := c.hooks().Process(c.ctx)
			if err != nil {
				c.log.ErrorContext(c.ctx, "process hook", "err", err)
				return
			}
			if !processed {
				return
			}
			if err := c.enqueueRunHooks(c.ctx); err != nil {
				return
			}
		}
		// Continue draining when the bounded pass exhausted its budget.
		c.hookPending.Store(true)
	})
}
