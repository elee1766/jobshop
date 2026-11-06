package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/writekey"
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
		c.log.ErrorContext(ctx, "admission release", "err", err)
	}
	// With ingest.workers external, `jobshop ingest work` admits (#443).
	for range admitBudget(c.config().Ingest.External()) {
		ok, err := c.admitNext(ctx)
		if err != nil && ctx.Err() == nil {
			c.log.ErrorContext(ctx, "admission", "err", err)
		}
		if !ok || err != nil {
			break
		}
	}
	if c.config().Ingest.ExternalDispatch() {
		return // `jobshop ingest work --dispatch` dispatches (#633)
	}
	if _, err := c.dispatchAll(ctx, dispatchHolder); err != nil && ctx.Err() == nil {
		c.log.ErrorContext(ctx, "dispatch", "err", err)
	}
}

// dispatchNext claims up to dispatchBatch admitted rows and hands them
// off. It returns how many it claimed.
func (c *Controller) dispatchNext(ctx context.Context, holder string) (int, error) {
	c.dispatchMu.Lock()
	defer c.dispatchMu.Unlock()
	var rows []store.InboxItem
	for range dispatchBatch {
		it, err := ingest.InboxClaim(ctx, c.store, []string{store.InboxAdmitted}, holder, dispatchLease)
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
	c.done(ctx, holder, other)
	c.dispatchActivity(ctx, holder, activity)
	c.dispatchComments(ctx, holder, comments)
	return len(rows), nil
}

// done marks rows dispatched. A lost lease is logged: the row is then
// claimed again, and the consumers are idempotent.
func (c *Controller) done(ctx context.Context, holder string, rows []store.InboxItem) {
	for _, it := range rows {
		if _, err := ingest.InboxSetState(ctx, c.store, it.ID, holder, store.InboxUpdate{State: store.InboxDispatched}); err != nil {
			c.log.WarnContext(ctx, "dispatch: mark dispatched", "delivery", it.DeliveryID, "err", err)
		}
	}
}

// dispatchComments delivers admitted poller comments, grouped by thread
// so a thread's news arrives as one message, as the poller used to send.
func (c *Controller) dispatchComments(ctx context.Context, holder string, rows []store.InboxItem) {
	if len(rows) == 0 {
		return
	}
	cfg := c.config()
	if cfg.Subscriptions == nil {
		c.done(ctx, holder, rows)
		return
	}
	now := c.now()
	window := config.Duration(cfg.Subscriptions.ResumeWindow, defaultSubResume)
	_, byThread, _, err := c.followers(ctx, now)
	if err != nil {
		c.log.WarnContext(ctx, "dispatch: followers", "err", err)
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
		wk, err := writekey.Claim(ctx, c.store, key, dispatchKeyOp, holder)
		if err != nil {
			c.log.WarnContext(ctx, "dispatch: write key", "delivery", it.DeliveryID, "err", err)
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
	c.done(ctx, holder, skip)
	for _, k := range order {
		g := groups[k]
		if err := c.deliverThread(ctx, g.th, g.cs, byThread[k], now, window); err != nil {
			c.log.WarnContext(ctx, "subscriptions: deliver", "thread", k, "err", err)
			continue // leases expire; retried
		}
		for _, key := range g.keys {
			if _, err := writekey.Finish(ctx, c.store, key, "delivered"); err != nil {
				c.log.WarnContext(ctx, "dispatch: finish write key", "key", key, "err", err)
			}
		}
		c.done(ctx, holder, g.rows)
	}
}

// dispatchActivity adds admitted activity items to the feed and checks
// merges for reverts.
func (c *Controller) dispatchActivity(ctx context.Context, holder string, rows []store.InboxItem) {
	if len(rows) == 0 {
		return
	}
	cfg := c.config()
	if cfg.Activity == nil {
		c.done(ctx, holder, rows)
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
	c.done(ctx, holder, rows)
}

// activitySource is the configured activity source: the test fake, else
// the live APIs.
func (c *Controller) activitySource() activitySource {
	if c.activitySrc != nil {
		return c.activitySrc
	}
	return &liveActivity{c: c}
}

// admitBudget is how many items one pass admits in process: none when
// external workers admit.
func admitBudget(external bool) int {
	if external {
		return 0
	}
	return 1000
}

// ingestDispatchMethod dispatches admitted rows for `jobshop ingest work
// --dispatch` (#633). The rows are claimed under the worker's own holder,
// so its leases fence it against the controller and other workers like
// any dispatcher; the dispatch/<row id> write key keeps a row whose lease
// expired mid-hand-off from being delivered twice.
func (c *Controller) ingestDispatchMethod(ctx context.Context, sc scope, p rpc.IngestClaimParams) (any, error) {
	if err := requireAdmin(sc, "ingest dispatch"); err != nil {
		return nil, err
	}
	if p.Worker == "" {
		return nil, fmt.Errorf("%w: worker is required", store.ErrInvalid)
	}
	c.beat(p, false, false)
	n, err := c.dispatchAll(ctx, ingestWorkerPrefix+p.Worker)
	if err != nil {
		return nil, err
	}
	return rpc.IngestDispatchResult{Dispatched: n}, nil
}

// dispatchAll runs dispatchNext as holder until nothing admitted is
// left, bounded, returning how many rows it claimed.
func (c *Controller) dispatchAll(ctx context.Context, holder string) (int, error) {
	total := 0
	for range 100 {
		n, err := c.dispatchNext(ctx, holder)
		total += n
		if n == 0 || err != nil {
			return total, err
		}
	}
	return total, nil
}
