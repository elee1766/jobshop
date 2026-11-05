package control

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Ingest split mode (#443, docs/notes/443-ingest-workers.md). With
// ingest.workers external the controller does not admit; `jobshop ingest
// work` processes claim pending items over the admin socket, run the
// hooks themselves and report the run. The controller keeps every store
// write: the claim, the needs-human ticket for a hold and the state.
// Heartbeats are kept in memory: after a controller restart the missing
// worker clock starts at the restart.

// ingestWorkerPrefix makes a worker's inbox holder distinct from the
// in-process admitter's.
const ingestWorkerPrefix = "worker:"

func (c *Controller) beat(w rpc.IngestClaimParams, claimed, admitted bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ingestWorkers == nil {
		c.ingestWorkers = map[string]*rpc.IngestWorker{}
	}
	hb := c.ingestWorkers[w.Worker]
	if hb == nil {
		hb = &rpc.IngestWorker{ID: w.Worker}
		c.ingestWorkers[w.Worker] = hb
	}
	if w.Host != "" {
		hb.Host = w.Host
	}
	hb.Seen = c.now()
	if claimed {
		hb.LastClaim = hb.Seen
	}
	if admitted {
		hb.Admitted++
	}
}

// ingestClaim leases one pending item to an external worker.
func (c *Controller) ingestClaim(ctx context.Context, sc scope, p rpc.IngestClaimParams) (any, error) {
	if err := requireAdmin(sc, "ingest claim"); err != nil {
		return nil, err
	}
	if p.Worker == "" {
		return nil, fmt.Errorf("%w: worker is required", store.ErrInvalid)
	}
	cfg := c.config()
	if !cfg.Ingest.External() {
		c.beat(p, false, false)
		return nil, fmt.Errorf("%w: ingest.workers is not external; the controller admits", store.ErrConflict)
	}
	hooks := admissionHooks(cfg)
	it, err := ingest.InboxClaim(ctx, c.store, []string{store.InboxPending}, ingestWorkerPrefix+p.Worker, admissionLease(hooks))
	if errors.Is(err, store.ErrNotFound) {
		c.beat(p, false, false)
		return rpc.IngestClaimResult{}, nil
	}
	if err != nil {
		return nil, err
	}
	c.beat(p, true, false)
	_, answered := c.humanAnswer(ctx, it.DeliveryID)
	return rpc.IngestClaimResult{Item: &it, Hooks: hooks, Project: cfg.DefaultProject, Answered: answered, LeaseMS: admissionLease(hooks).Milliseconds()}, nil
}

// ingestAdmit stores a worker's run for an item it still leases.
func (c *Controller) ingestAdmit(ctx context.Context, sc scope, p rpc.IngestAdmitParams) (any, error) {
	if err := requireAdmin(sc, "ingest admit"); err != nil {
		return nil, err
	}
	if p.Worker == "" {
		return nil, fmt.Errorf("%w: worker is required", store.ErrInvalid)
	}
	var run admissionRun
	if err := json.Unmarshal(p.Run, &run); err != nil {
		return nil, fmt.Errorf("%w: run: %w", store.ErrInvalid, err)
	}
	it, err := ingest.InboxGet(ctx, c.store, p.ID)
	if err != nil {
		return nil, err
	}
	// The worker may not move the item to another delivery.
	run.Event.On, run.Event.DeliveryID, run.Event.Source = admissionOn, it.DeliveryID, it.Source
	if run.Results == nil {
		run.Results = []admissionResult{}
	}
	if err := c.finishAdmission(ctx, it, ingestWorkerPrefix+p.Worker, run); err != nil {
		return nil, err
	}
	c.beat(rpc.IngestClaimParams{Worker: p.Worker}, false, true)
	got, err := ingest.InboxGet(ctx, c.store, p.ID)
	if err != nil {
		return nil, err
	}
	return got, nil
}

// ingestWorkersMethod lists the external workers' heartbeats.
func (c *Controller) ingestWorkersMethod(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "attention", "ingest workers"); err != nil {
		return nil, err
	}
	counts, err := ingest.InboxCounts(ctx, c.store)
	if err != nil {
		return nil, err
	}
	mode := cmp.Or(c.config().Ingest.Workers, config.IngestWorkersInternal)
	return rpc.IngestWorkersResult{Mode: mode, Workers: c.workerList(), Counts: counts}, nil
}

func (c *Controller) workerList() []rpc.IngestWorker {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]rpc.IngestWorker, 0, len(c.ingestWorkers))
	for _, k := range slices.Sorted(maps.Keys(c.ingestWorkers)) {
		out = append(out, *c.ingestWorkers[k])
	}
	return out
}

// lastWorkerBeat is the newest heartbeat, or the controller's start when
// no worker has called since.
func (c *Controller) lastWorkerBeat() time.Time {
	last := c.started
	for _, w := range c.workerList() {
		if w.Seen.After(last) {
			last = w.Seen
		}
	}
	return last
}

// ingestAttention lists a missing external worker and an admission
// backlog. The same conditions are the ingest_worker and
// ingest_backlog stall signals.
func (c *Controller) ingestAttention(ctx context.Context) ([]rpc.AttentionItem, error) {
	var items []rpc.AttentionItem
	if s, ok := c.missingWorker(); ok {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionIngest, ID: "ingest-worker", Title: s.detail, Since: c.lastWorkerBeat()})
	}
	s, ok, err := c.ingestBacklog(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionIngest, ID: "ingest-backlog", Title: s.detail, Since: c.now().Add(-c.config().Ingest.BacklogAge())})
	}
	return items, nil
}

// missingWorker: external mode and no worker heartbeat for
// ingest.worker_missing.
func (c *Controller) missingWorker() (stall, bool) {
	in := c.config().Ingest
	if !in.External() && !in.ExternalDispatch() {
		return stall{}, false
	}
	last := c.lastWorkerBeat()
	if c.now().Sub(last) < in.WorkerMissingAfter() {
		return stall{}, false
	}
	return stall{
		signal: config.StallIngestWorker, subject: "ingest-worker@" + last.UTC().Format("20060102T150405Z"),
		detail: fmt.Sprintf("ingest.workers or ingest.dispatch is external and no `jobshop ingest work` has called for %s; pending items are not admitted or admitted ones not dispatched", ago(c.now(), last)),
	}, true
}

// ingestBacklog: more than ingest.backlog_max pending items older than
// ingest.backlog_after.
func (c *Controller) ingestBacklog(ctx context.Context) (stall, bool, error) {
	in := c.config().Ingest
	n, err := ingest.InboxCountOlder(ctx, c.store, store.InboxPending, c.now().Add(-in.BacklogAge()))
	if err != nil || n <= int64(in.BacklogMax) {
		return stall{}, false, err
	}
	return stall{
		signal: config.StallIngestBacklog, subject: "ingest-backlog",
		detail: fmt.Sprintf("%d pending inbox items older than %s (ingest.backlog_max %d)", n, in.BacklogAge(), in.BacklogMax),
	}, true, nil
}

func (c *Controller) stallIngest(ctx context.Context, signal string) ([]stall, error) {
	if signal == config.StallIngestWorker {
		if s, ok := c.missingWorker(); ok {
			return []stall{s}, nil
		}
		return nil, nil
	}
	s, ok, err := c.ingestBacklog(ctx)
	if !ok {
		return nil, err
	}
	return []stall{s}, nil
}

// recordInboxGauges records inbox counts by state and live workers.
func (c *Controller) recordInboxGauges(ctx context.Context) {
	counts, err := ingest.InboxCounts(ctx, c.store)
	if err != nil {
		return
	}
	for st, n := range counts {
		c.metrics.Inbox.Record(ctx, n, metric.WithAttributes(telemetry.AttrState.String(st)))
	}
	live := int64(0)
	cutoff := c.now().Add(-c.config().Ingest.WorkerMissingAfter())
	for _, w := range c.workerList() {
		if w.Seen.After(cutoff) {
			live++
		}
	}
	c.metrics.IngestWorkers.Record(ctx, live)
}
