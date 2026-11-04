package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Attribute keys shared by all signals (plan section 9).
const (
	AttrProject    = attribute.Key("jobshop.project")
	AttrPool       = attribute.Key("jobshop.pool")
	AttrSession    = attribute.Key("jobshop.session.id")
	AttrTicket     = attribute.Key("jobshop.ticket.id")
	AttrKind       = attribute.Key("jobshop.kind")
	AttrState      = attribute.Key("jobshop.session.state")
	AttrResolution = attribute.Key("jobshop.resolution")
	AttrMethod     = attribute.Key("rpc.method")
	AttrReason     = attribute.Key("reason")
	AttrStep       = attribute.Key("jobshop.step")
	AttrStepType   = attribute.Key("jobshop.step.type")
	AttrResult     = attribute.Key("jobshop.step.result")
)

// Metrics holds the controller's instruments. Each Metrics has its own
// Prometheus registry so tests and multiple controllers do not collide.
type Metrics struct {
	provider *sdkmetric.MeterProvider
	handler  http.Handler

	Sessions      metric.Int64Gauge
	Restarts      metric.Int64Counter
	Ready         metric.Int64Gauge
	Claims        metric.Int64Counter
	Closes        metric.Int64Counter
	LeaseExpiries metric.Int64Counter
	RequestTime   metric.Float64Histogram
	ConvDropped   metric.Int64Counter
	ConvTruncated metric.Int64Counter
	Turns         metric.Int64Counter
	SessionsEnded metric.Int64Counter
	Orphaned      metric.Int64Gauge
	// RunsNeedsHuman is open runs in state needs_human, by project.
	RunsNeedsHuman metric.Int64Gauge
	// Pool scaling (#281).
	PoolDesired  metric.Int64Gauge
	PoolCordoned metric.Int64Gauge
	ScaleEvents  metric.Int64Counter
	// InjectedBytes is prompt rule bytes per status (#397).
	InjectedBytes metric.Int64Counter
	// Command and http steps (#396): attempt duration and results.
	StepDuration metric.Float64Histogram
	StepResults  metric.Int64Counter
	// Ingest inbox items by state and live external workers (#443).
	Inbox         metric.Int64Gauge
	IngestWorkers metric.Int64Gauge
}

// NewMetrics builds the meter provider, Prometheus exporter and
// instruments.
func NewMetrics() (*Metrics, error) {
	reg := prometheus.NewRegistry()
	exp, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp))
	meter := mp.Meter("github.com/elee1766/jobshop")
	m := &Metrics{provider: mp, handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{})}
	var errs [21]error
	m.Sessions, errs[0] = meter.Int64Gauge("jobshop.sessions", metric.WithDescription("Open sessions by pool and state."))
	m.Restarts, errs[1] = meter.Int64Counter("jobshop.session.restarts", metric.WithDescription("Session agent or host restarts."))
	m.Ready, errs[2] = meter.Int64Gauge("jobshop.tickets.ready", metric.WithDescription("Ready tickets per pool, from the supervisor tick."))
	m.Claims, errs[3] = meter.Int64Counter("jobshop.claims", metric.WithDescription("Successful claims."))
	m.Closes, errs[4] = meter.Int64Counter("jobshop.closes", metric.WithDescription("Ticket closes by resolution."))
	m.LeaseExpiries, errs[5] = meter.Int64Counter("jobshop.lease.expiries", metric.WithDescription("Claims dropped by lease expiry."))
	m.RequestTime, errs[6] = meter.Float64Histogram("jobshop.rpc.duration", metric.WithUnit("s"),
		metric.WithDescription("Socket request latency by method."))
	m.ConvDropped, errs[7] = meter.Int64Counter("jobshop.conversation_export.dropped",
		metric.WithDescription("Conversation export frames dropped by a full host queue."))
	m.ConvTruncated, errs[8] = meter.Int64Counter("jobshop.conversation_export.truncated",
		metric.WithDescription("Conversation export fields cut at the size cap."))
	m.Turns, errs[9] = meter.Int64Counter("jobshop.turns",
		metric.WithDescription("Agent turns by pool and stop reason (error: the prompt failed)."))
	m.SessionsEnded, errs[10] = meter.Int64Counter("jobshop.sessions_ended",
		metric.WithDescription("Sessions ended by pool and resolution (idle, crashed, stuck, ...)."))
	m.Orphaned, errs[11] = meter.Int64Gauge("jobshop.tickets.orphaned",
		metric.WithDescription("Open unclaimed tickets no pool will claim, by reason (no_pool, unknown_pool)."))
	m.RunsNeedsHuman, errs[12] = meter.Int64Gauge("jobshop.runs.needs_human",
		metric.WithDescription("Open runs waiting on a human (a step failed at its retry limit), by project."))
	m.PoolDesired, errs[13] = meter.Int64Gauge("jobshop.pool.sessions.desired",
		metric.WithDescription("Desired sessions per ticket pool, from the supervisor tick (#281)."))
	m.PoolCordoned, errs[14] = meter.Int64Gauge("jobshop.pool.sessions.cordoned",
		metric.WithDescription("Open sessions out of rotation per pool (scale, manual or drain cordon)."))
	m.ScaleEvents, errs[15] = meter.Int64Counter("jobshop.pool.scale_events",
		metric.WithDescription("Pool scale changes by pool and reason (up, cordon, uncordon, manual, auto)."))
	m.InjectedBytes, errs[16] = meter.Int64Counter("jobshop.prompt.injected_bytes", metric.WithUnit("By"),
		metric.WithDescription("Prompt rule bytes by pool and reason: new, changed, always, removed, or unchanged (skipped)."))
	m.StepDuration, errs[17] = meter.Float64Histogram("jobshop.step.duration", metric.WithUnit("s"),
		metric.WithDescription("Command and http step attempts by step, type and result."))
	m.StepResults, errs[18] = meter.Int64Counter("jobshop.step.results",
		metric.WithDescription("Command and http step attempts by step, type and result (ok, not_yet, failed)."))
	m.Inbox, errs[19] = meter.Int64Gauge("jobshop.inbox.items",
		metric.WithDescription("Ingest inbox items by state (#443)."))
	m.IngestWorkers, errs[20] = meter.Int64Gauge("jobshop.ingest.workers",
		metric.WithDescription("External ingest workers with a heartbeat within ingest.worker_missing (#443)."))
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Handler serves /metrics.
func (m *Metrics) Handler() http.Handler { return m.handler }

// Shutdown flushes and stops the meter provider.
func (m *Metrics) Shutdown(ctx context.Context) error { return m.provider.Shutdown(ctx) }

// Since records the seconds elapsed since start for one request.
func (m *Metrics) Since(ctx context.Context, start time.Time, method string) {
	m.RequestTime.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(AttrMethod.String(method)))
}
