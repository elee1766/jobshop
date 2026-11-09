package control

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/hooks"
	"github.com/elee1766/jobshop/pkg/orchestration/triggers"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// enqueueHookEvents runs before acknowledging ingress. Re-delivery is harmless.
func (c *Controller) enqueueHookEvents(ctx context.Context, items []domain.InboxItem) error {
	lc := c.live.Load()
	for _, it := range items {
		ev := triggers.EventOf(it)
		for _, name := range slices.Sorted(maps.Keys(lc.cfg.Triggers)) {
			tr := lc.cfg.Triggers[name]
			if tr.Handler == "" || !triggers.On(tr, ev) {
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
			event := &hooks.Event{Source: it.Source, Event: it.Kind, Delivery: it.DeliveryID, Repository: it.RepoOrTeam, Trigger: name, Payload: object, Fields: ev.Vars()}
			if err := c.hooks().Put(ctx, "event/"+name+"/"+it.Source+"/"+it.DeliveryID, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: program.Digest(), Event: event, Args: tr.Args}); err != nil {
				return err
			}
		}
	}
	c.kickHooks()
	return nil
}

func (c *Controller) enqueueRunHooks(ctx context.Context) error {
	runs, err := c.store.List(ctx, store.Filter{Kind: domain.KindRun, State: domain.RunRunning, Limit: store.NoLimit})
	if err != nil {
		return err
	}
	for _, run := range runs {
		f, err := domain.DecodeFields[domain.RunFields](run)
		if err != nil {
			return err
		}
		if f.Automation == nil || run.Status == domain.StatusClosed {
			continue
		}
		children, err := store.Children(ctx, c.store, run.ID)
		if err != nil {
			return err
		}
		// What is due and its key are hooks.RunDue's; the controller only
		// reads the run and hands the invocation to the journal.
		key, in, due, err := hooks.RunDue(run, f, children, metaString(run.Meta, "automation", "invocation"))
		if err != nil {
			return err
		}
		if due {
			if err := c.hooks().Put(ctx, key, in); err != nil {
				return err
			}
		}
	}
	return nil
}

// kickHooks starts the bounded hook drain (hooks.Worker) without tickMu,
// runMu or the controller mu.
func (c *Controller) kickHooks() { c.hookWorker().Kick() }

func (c *Controller) hookWorker() *hooks.Worker {
	c.hookWorkerOnce.Do(func() {
		c.hookDrain = &hooks.Worker{
			Go:      func(f func()) { c.wg.Go(f) },
			Context: func() context.Context { return c.ctx },
			Enqueue: c.enqueueRunHooks,
			Process: func(ctx context.Context) (bool, error) { return c.hooks().Process(ctx) },
			Log:     c.log,
		}
	})
	return c.hookDrain
}
