package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/ingest"
	"github.com/elee1766/jobshop/pkg/orchestration/ingest/queue"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Ingest admission (#348): the controller's half. The hook runner, the
// event and result types and the state decision live in
// pkg/orchestration/ingest/queue; this file claims pending items, stores a
// run's outcome and files the hold ticket, all store writes.

// pumpAdmission admits pending items and dispatches admitted ones (#442)
// until the controller stops. The queue wakes it on an inbox change
// (#679) and on a backstop timer.
func (c *Controller) pumpAdmission() {
	c.queue().RunPass(c.ctx, c.ingestPass)
}

// releaseAnswered moves held items whose needs-human ticket is closed
// back to pending (#353). Admission only claims pending items; on the
// next claim admit reads the answer through askedHuman: done admits,
// any other resolution denies. Held is not pruned, so without this an
// answered item would stay held forever.
func (c *Controller) releaseAnswered(ctx context.Context) error {
	held, err := ingest.InboxList(ctx, c.store, store.InboxFilter{State: domain.InboxHeld, Limit: 200})
	if err != nil {
		return err
	}
	for _, it := range held {
		a, ok := c.askedHuman(ctx, "admission/"+it.DeliveryID)
		if !ok || a.Decision == domain.GuardAskHuman {
			continue
		}
		if _, err := ingest.InboxRequeue(ctx, c.store, it.ID, domain.InboxHeld, domain.InboxPending); err != nil {
			if errors.Is(err, domain.ErrConflict) {
				continue // moved or leased meanwhile
			}
			return err
		}
		c.log.InfoContext(ctx, "inbox admission released", "delivery", it.DeliveryID, "ticket", a.Ticket, "answer", a.Decision)
	}
	return nil
}

// admitNext claims one pending item and admits it. false: none pending.
func (c *Controller) admitNext(ctx context.Context) (bool, error) {
	cfg := c.config()
	hooks := queue.Hooks(cfg)
	it, err := ingest.InboxClaim(ctx, c.store, []string{domain.InboxPending}, queue.Holder, queue.Lease(hooks))
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, answered := c.humanAnswer(ctx, it.DeliveryID)
	run := c.admissionRunner().Run(ctx, cfg.DefaultProject, hooks, it, answered)
	return true, c.finishAdmission(ctx, it, queue.Holder, run)
}

// humanAnswer is a human's closed answer to an earlier hold of delivery.
func (c *Controller) humanAnswer(ctx context.Context, delivery string) (rpc.GuardResult, bool) {
	a, ok := c.askedHuman(ctx, "admission/"+delivery)
	if !ok || a.Decision == domain.GuardAskHuman {
		return rpc.GuardResult{}, false
	}
	return a, true
}

// admissionRunner runs the in-process hooks in the controller's root.
func (c *Controller) admissionRunner() queue.Runner {
	return queue.Runner{Dir: c.root.Dir, After: c.after}
}

// finishAdmission stores a run's outcome for an item holder leases,
// filing the needs-human ticket for a hold. A human's closed answer to
// an earlier hold overrides the validating hooks.
func (c *Controller) finishAdmission(ctx context.Context, it domain.InboxItem, holder string, run queue.Run) error {
	a, answered := c.humanAnswer(ctx, it.DeliveryID)
	state, admitted, results := queue.Settle(run, a, answered)
	if state == domain.InboxHeld {
		c.holdForHuman(ctx, "admission/"+it.DeliveryID, run.Event, strings.Join(run.Asks, "\n"))
	}
	raw, err := json.Marshal(results)
	if err != nil {
		return err
	}
	u := store.InboxUpdate{State: state, HookResults: raw}
	if state == domain.InboxAdmitted {
		u.Admitted = admitted
	}
	if _, err := ingest.InboxSetState(ctx, c.store, it.ID, holder, u); err != nil {
		return err
	}
	if state == domain.InboxAdmitted {
		// Dispatch-only workers long-poll for admitted items (#679).
		c.queue().Advance()
	}
	c.log.InfoContext(ctx, "inbox admission", "delivery", it.DeliveryID, "state", state, "by", holder)
	return nil
}

// holdForHuman files one needs-human ticket per held delivery, like a
// refused intake (#288). Closing it done admits the item when it is
// next claimed; any other resolution denies it.
func (c *Controller) holdForHuman(ctx context.Context, key string, ev queue.Event, reason string) {
	open, err := c.store.List(ctx, store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: key, Order: store.OrderNewest, Limit: 1})
	if err == nil && len(open) > 0 && open[0].Status != domain.StatusClosed {
		return
	}
	meta, err := json.Marshal(map[string]map[string]string{hookMetaPack: {"key": key}})
	if err != nil {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "An admission hook asks a human about %s %s delivery %s", ev.Source, ev.Kind, ev.DeliveryID)
	if ev.RepoOrTeam != "" {
		fmt.Fprintf(&b, " in %s", ev.RepoOrTeam)
		if ev.Number != 0 {
			fmt.Fprintf(&b, "#%d", ev.Number)
		}
	}
	if ev.Actor != "" {
		fmt.Fprintf(&b, " by %s", ev.Actor)
	}
	fmt.Fprintf(&b, ".\n\nReason:\n%s\n\nQuestion: may this event reach an agent?\n", reason)
	b.WriteString("Close done to allow it, close with any other resolution to keep it out.")
	project := ev.Project
	if project == "" {
		project = c.config().DefaultProject
	}
	prio := 1
	if _, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: project, Kind: domain.KindTask, Title: "Admission held " + ev.Source + " delivery " + ev.DeliveryID,
		Body: b.String(), Priority: &prio, Labels: []string{needsHumanLabel, questionLabel}, Meta: meta,
	}); err != nil {
		c.log.WarnContext(ctx, "admission: file ticket", "delivery", ev.DeliveryID, "err", err)
	}
}
