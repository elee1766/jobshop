package control

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/ingestqueue"
	"github.com/elee1766/jobshop/pkg/control/triggers"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Ingest split mode (#443, docs/notes/443-ingest-workers.md). With
// ingest.workers external the controller does not admit; `jobshop ingest
// work` processes claim pending items over the admin socket, run the
// hooks themselves and report the run. The controller keeps every store
// write: the claim, the needs-human ticket for a hold and the state. The
// worker loop, heartbeats and hook runner are pkg/control/ingestqueue;
// the heartbeats are in memory, so after a controller restart the missing
// worker clock starts at the restart.

// ingestClaim leases one pending item to an external worker.
func (c *Controller) ingestClaim(ctx context.Context, sc scope, p rpc.IngestClaimParams) (any, error) {
	if err := requireAdmin(sc, "ingest claim"); err != nil {
		return nil, err
	}
	if p.Worker == "" {
		return nil, fmt.Errorf("%w: worker is required", domain.ErrInvalid)
	}
	cfg := c.config()
	if !cfg.Ingest.External() {
		c.queue().Beat(p, false, false)
		return nil, fmt.Errorf("%w: ingest.workers is not external; the controller admits", domain.ErrConflict)
	}
	hooks := ingestqueue.Hooks(cfg)
	lease := ingestqueue.Lease(hooks)
	it, err := ingest.InboxClaim(ctx, c.store, []string{domain.InboxPending}, ingestqueue.WorkerPrefix+p.Worker, lease)
	if errors.Is(err, domain.ErrNotFound) {
		c.queue().Beat(p, false, false)
		return rpc.IngestClaimResult{}, nil
	}
	if err != nil {
		return nil, err
	}
	c.queue().Beat(p, true, false)
	_, answered := c.humanAnswer(ctx, it.DeliveryID)
	return rpc.IngestClaimResult{Item: &it, Hooks: hooks, Project: cfg.DefaultProject, Answered: answered, LeaseMS: lease.Milliseconds()}, nil
}

// ingestAdmit stores a worker's run for an item it still leases.
func (c *Controller) ingestAdmit(ctx context.Context, sc scope, p rpc.IngestAdmitParams) (any, error) {
	if err := requireAdmin(sc, "ingest admit"); err != nil {
		return nil, err
	}
	if p.Worker == "" {
		return nil, fmt.Errorf("%w: worker is required", domain.ErrInvalid)
	}
	var run ingestqueue.Run
	if err := json.Unmarshal(p.Run, &run); err != nil {
		return nil, fmt.Errorf("%w: run: %w", domain.ErrInvalid, err)
	}
	it, err := ingest.InboxGet(ctx, c.store, p.ID)
	if err != nil {
		return nil, err
	}
	// The worker may not move the item to another delivery.
	run.Event.On, run.Event.DeliveryID, run.Event.Source = ingestqueue.On, it.DeliveryID, it.Source
	if run.Results == nil {
		run.Results = []ingestqueue.Result{}
	}
	if err := c.finishAdmission(ctx, it, ingestqueue.WorkerPrefix+p.Worker, run); err != nil {
		return nil, err
	}
	c.queue().Beat(rpc.IngestClaimParams{Worker: p.Worker}, false, true)
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
	return rpc.IngestWorkersResult{Mode: mode, Workers: c.queue().Workers(), Counts: counts}, nil
}

// lastWorkerBeat is the newest heartbeat, or the controller's start when
// no worker has called since.
func (c *Controller) lastWorkerBeat() time.Time { return c.queue().LastBeat(c.started) }

// ingestAttention lists a missing external worker and an admission
// backlog. The same conditions are the ingest_worker and
// ingest_backlog scan signals.
func (c *Controller) ingestAttention(ctx context.Context, cfg *config.Config) ([]rpc.AttentionItem, error) {
	var items []rpc.AttentionItem
	if s, ok := c.missingWorker(cfg); ok {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionIngest, ID: "ingest-worker", Title: s.Detail, Since: c.lastWorkerBeat()})
	}
	s, ok, err := c.ingestBacklog(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if ok {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionIngest, ID: "ingest-backlog", Title: s.Detail, Since: c.now().Add(-cfg.Ingest.BacklogAge())})
	}
	return items, nil
}

// missingWorker: external mode and no worker heartbeat for
// ingest.worker_missing.
func (c *Controller) missingWorker(cfg *config.Config) (triggers.Finding, bool) {
	in := cfg.Ingest
	if !in.External() && !in.ExternalDispatch() {
		return triggers.Finding{}, false
	}
	last := c.lastWorkerBeat()
	if c.now().Sub(last) < in.WorkerMissingAfter() {
		return triggers.Finding{}, false
	}
	return triggers.Finding{
		Signal: config.SignalIngestWorker, Subject: "ingest-worker@" + last.UTC().Format("20060102T150405Z"),
		Detail: fmt.Sprintf("ingest.workers or ingest.dispatch is external and no `jobshop ingest work` has called for %s; pending items are not admitted or admitted ones not dispatched", triggers.Ago(c.now(), last)),
	}, true
}

// ingestBacklog: more than ingest.backlog_max pending items older than
// ingest.backlog_after.
func (c *Controller) ingestBacklog(ctx context.Context, cfg *config.Config) (triggers.Finding, bool, error) {
	in := cfg.Ingest
	n, err := ingest.InboxCountOlder(ctx, c.store, domain.InboxPending, c.now().Add(-in.BacklogAge()))
	if err != nil || n <= int64(in.BacklogMax) {
		return triggers.Finding{}, false, err
	}
	return triggers.Finding{
		Signal: config.SignalIngestBacklog, Subject: "ingest-backlog",
		Detail: fmt.Sprintf("%d pending inbox items older than %s (ingest.backlog_max %d)", n, in.BacklogAge(), in.BacklogMax),
	}, true, nil
}

func (c *Controller) signalIngest(ctx context.Context, cfg *config.Config, signal string) ([]triggers.Finding, error) {
	if signal == config.SignalIngestWorker {
		if s, ok := c.missingWorker(cfg); ok {
			return []triggers.Finding{s}, nil
		}
		return nil, nil
	}
	s, ok, err := c.ingestBacklog(ctx, cfg)
	if !ok {
		return nil, err
	}
	return []triggers.Finding{s}, nil
}

// recordInboxGauges records inbox counts by state and live workers.
func (c *Controller) recordInboxGauges(ctx context.Context, cfg *config.Config) {
	counts, err := ingest.InboxCounts(ctx, c.store)
	if err != nil {
		return
	}
	for st, n := range counts {
		c.metrics.Inbox.Record(ctx, n, metric.WithAttributes(telemetry.AttrState.String(st)))
	}
	c.metrics.IngestWorkers.Record(ctx, c.queue().Live(c.now().Add(-cfg.Ingest.WorkerMissingAfter())))
}
