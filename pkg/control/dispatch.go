package control

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Ingest dispatch (#442, docs/notes/442-ingest-dispatch.md). Pollers only
// write the inbox; admission (#348) moves rows to admitted, denied or
// held; this stage lease-claims admitted rows and hands each to its
// consumer using the admitted payload, then marks it dispatched. Denied
// and held rows are never claimed here. Consumers:
//   - subscription poller comments (poll:github:..., poll:linear:...):
//     the thread's owner, through linearAnswers and deliverThread;
//   - activity poller items (poll:activity:...): the activity feed and
//     revert learnings;
//   - webhook rows: no consumer yet; webhooks only wake a poll, so they
//     are marked dispatched.

const (
	dispatchHolder = "dispatch"
	dispatchLease  = 5 * time.Minute
	dispatchBatch  = 100
	// dispatchKeyOp is the write key op that makes a comment delivery
	// happen once per delivery id, even if a lease expires mid-hand-off.
	dispatchKeyOp = "inbox.dispatch"
)

// ingestPass releases answered holds, admits every pending row, then
// dispatches every admitted row. Pollers call it after writing so their
// finds are delivered in the same pass; the admission pump runs it too.
func (c *Controller) ingestPass(ctx context.Context) {
	if err := c.releaseAnswered(ctx); err != nil && ctx.Err() == nil {
		c.log.Error("admission release", "err", err)
	}
	for range 1000 {
		ok, err := c.admitNext(ctx)
		if err != nil && ctx.Err() == nil {
			c.log.Error("admission", "err", err)
		}
		if !ok || err != nil {
			break
		}
	}
	for range 100 {
		n, err := c.dispatchNext(ctx)
		if err != nil && ctx.Err() == nil {
			c.log.Error("dispatch", "err", err)
		}
		if n == 0 || err != nil {
			break
		}
	}
}

// dispatchNext claims up to dispatchBatch admitted rows and hands them
// off. It returns how many it claimed.
func (c *Controller) dispatchNext(ctx context.Context) (int, error) {
	c.dispatchMu.Lock()
	defer c.dispatchMu.Unlock()
	var rows []store.InboxItem
	for range dispatchBatch {
		it, err := c.store.InboxClaim(ctx, []string{store.InboxAdmitted}, dispatchHolder, dispatchLease)
		if errors.Is(err, store.ErrNotFound) {
			break
		}
		if err != nil {
			return len(rows), err
		}
		rows = append(rows, it)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	var comments, activity, other []store.InboxItem
	for _, it := range rows {
		switch {
		case strings.HasPrefix(it.DeliveryID, activityPrefix):
			activity = append(activity, it)
		case strings.HasPrefix(it.DeliveryID, "poll:"):
			comments = append(comments, it)
		default:
			other = append(other, it)
		}
	}
	c.done(ctx, other)
	c.dispatchActivity(ctx, activity)
	c.dispatchComments(ctx, comments)
	return len(rows), nil
}

// done marks rows dispatched. A lost lease is logged: the row is then
// claimed again, and the consumers are idempotent.
func (c *Controller) done(ctx context.Context, rows []store.InboxItem) {
	for _, it := range rows {
		if _, err := c.store.InboxSetState(ctx, it.ID, dispatchHolder, store.InboxUpdate{State: store.InboxDispatched}); err != nil {
			c.log.Warn("dispatch: mark dispatched", "delivery", it.DeliveryID, "err", err)
		}
	}
}

// dispatchComments delivers admitted poller comments, grouped by thread
// so a thread's news arrives as one message, as the poller used to send.
func (c *Controller) dispatchComments(ctx context.Context, rows []store.InboxItem) {
	if len(rows) == 0 {
		return
	}
	cfg := c.config()
	if cfg.Subscriptions == nil {
		c.done(ctx, rows)
		return
	}
	now := c.now()
	window := config.Duration(cfg.Subscriptions.ResumeWindow, defaultSubResume)
	_, byThread, _, err := c.followers(ctx, now)
	if err != nil {
		c.log.Warn("dispatch: followers", "err", err)
		return // leases expire; retried
	}
	type group struct {
		th   thread
		cs   []ghComment
		keys []string
		rows []store.InboxItem
	}
	groups := map[string]*group{}
	var order []string
	var skip []store.InboxItem
	for _, it := range rows {
		var p pollComment
		if json.Unmarshal(it.Admitted, &p) != nil || (p.Thread.Linear == "" && p.Thread.Repo == "") {
			skip = append(skip, it) // not a comment row (older format)
			continue
		}
		cm := p.Comment.comment()
		k := p.Thread.key()
		if c.ignored(cm) || len(byThread[k]) == 0 {
			skip = append(skip, it)
			continue
		}
		key := "dispatch/" + it.DeliveryID
		wk, err := c.store.ClaimWriteKey(ctx, key, dispatchKeyOp, "controller")
		if err != nil {
			c.log.Warn("dispatch: write key", "delivery", it.DeliveryID, "err", err)
			continue
		}
		if wk.Done {
			skip = append(skip, it) // delivered before a lost lease
			continue
		}
		g := groups[k]
		if g == nil {
			th, _ := threadOf(byThread[k], k)
			g = &group{th: th}
			groups[k] = g
			order = append(order, k)
		}
		g.cs = append(g.cs, cm)
		g.keys = append(g.keys, key)
		g.rows = append(g.rows, it)
	}
	c.done(ctx, skip)
	for _, k := range order {
		g := groups[k]
		if err := c.deliverThread(ctx, g.th, g.cs, byThread[k], now, window); err != nil {
			c.log.Warn("subscriptions: deliver", "thread", k, "err", err)
			continue // leases expire; retried
		}
		for _, key := range g.keys {
			if _, err := c.store.FinishWriteKey(ctx, key, "delivered"); err != nil {
				c.log.Warn("dispatch: finish write key", "key", key, "err", err)
			}
		}
		c.done(ctx, g.rows)
	}
}

// dispatchActivity adds admitted activity items to the feed and checks
// merges for reverts.
func (c *Controller) dispatchActivity(ctx context.Context, rows []store.InboxItem) {
	if len(rows) == 0 {
		return
	}
	cfg := c.config()
	if cfg.Activity == nil {
		c.done(ctx, rows)
		return
	}
	c.activityMu.Lock()
	if c.activity == nil {
		c.activity = c.loadActivity()
	}
	s := c.activity
	c.activityMu.Unlock()
	var items []rpc.ActivityItem
	for _, it := range rows {
		var a rpc.ActivityItem
		if json.Unmarshal(it.Admitted, &a) == nil && a.ID != "" {
			items = append(items, a)
		}
	}
	now := c.now()
	maxPer := cfg.Activity.MaxItems
	if maxPer == 0 {
		maxPer = defaultActivityMax
	}
	added := s.add(items, now, config.Duration(cfg.Activity.Window, defaultActivityWindow), maxPer)
	if rs, ok := c.activitySource().(revertSource); ok {
		c.checkReverts(ctx, rs, items)
	}
	if err := s.save(); err != nil {
		c.activityErr(s, "cache", err)
	}
	if added > 0 && c.events != nil {
		c.events.publish([]rpc.Event{{Kind: rpc.EventKindActivity, Time: now}})
	}
	c.done(ctx, rows)
}

// activitySource is the configured activity source: the test fake, else
// the live APIs.
func (c *Controller) activitySource() activitySource {
	if c.activitySrc != nil {
		return c.activitySrc
	}
	return &liveActivity{c: c}
}
