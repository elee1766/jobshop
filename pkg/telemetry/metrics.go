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
	"go.opentelemetry.io/otel/sdk/resource"
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
	AttrStopReason = attribute.Key("jobshop.turn.stop_reason")
	AttrModel      = attribute.Key("jobshop.session.model")
	AttrEffort     = attribute.Key("jobshop.session.effort")
	AttrPoller     = attribute.Key("jobshop.poller")
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
	// Journal shipping to VictoriaLogs (#530), by pool.
	JournalFrames metric.Int64Counter
	JournalBytes  metric.Int64Counter
	JournalErrors metric.Int64Counter
	JournalLag    metric.Float64Gauge
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
	// Session cgroups of isolated pools (#388), summed per pool each tick.
	CgroupMemory   metric.Int64Gauge
	CgroupCPU      metric.Int64Gauge
	CgroupOOMKills metric.Int64Gauge
	// TickStep is the duration of each supervisor tick step, so a slow
	// step shows on a dashboard without opening a trace (JOB-105).
	TickStep metric.Float64Histogram
	// Outbound is the time to response headers of each outbound HTTP call,
	// by kind (linear, github, hook, ...) and target host.
	Outbound metric.Float64Histogram
	// LockWait is time spent waiting for the tick lock, by waiter.
	LockWait metric.Float64Histogram
	// PollerTimeouts counts calls run off the pass (checks, Linear reads,
	// provider calls, hooks) that ran out their timeout (JOB-103).
	PollerTimeouts metric.Int64Counter
}

// NewMetrics builds the meter provider, Prometheus exporter and
// instruments with the minimal jobshop resource (tests).
func NewMetrics() (*Metrics, error) { return NewMetricsFor(nil) }

// NewMetricsFor builds the meter provider, Prometheus exporter and
// instruments. res is the process resource (NewResource): the exporter
// serves it as target_info. Nil: service.name=jobshop only.
func NewMetricsFor(res *resource.Resource) (*Metrics, error) {
	reg := prometheus.NewRegistry()
	exp, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithResource(orDefault(res)), sdkmetric.WithReader(exp))
	meter := mp.Meter("github.com/elee1766/jobshop")
	m := &Metrics{provider: mp, handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{})}
	var errs [30]error
	m.Sessions, errs[0] = meter.Int64Gauge("jobshop.sessions", metric.WithDescription("Open sessions by pool and state."))
	m.Restarts, errs[1] = meter.Int64Counter("jobshop.session.restarts", metric.WithDescription("Session agent or host restarts."))
	m.Ready, errs[2] = meter.Int64Gauge("jobshop.tickets.ready", metric.WithDescription("Ready tickets per pool, from the supervisor tick."))
	m.Claims, errs[3] = meter.Int64Counter("jobshop.claims", metric.WithDescription("Successful claims."))
	m.Closes, errs[4] = meter.Int64Counter("jobshop.closes", metric.WithDescription("Ticket closes by resolution."))
	m.LeaseExpiries, errs[5] = meter.Int64Counter("jobshop.lease.expiries", metric.WithDescription("Claims dropped by lease expiry."))
	m.RequestTime, errs[6] = meter.Float64Histogram("jobshop.rpc.duration", metric.WithUnit("s"),
		metric.WithDescription("Socket request latency by method."))
	m.JournalFrames, errs[7] = meter.Int64Counter("jobshop.journal.ship.frames",
		metric.WithDescription("Journal frames shipped to VictoriaLogs, by pool (#530)."))
	// No unit: the exporter would name it _bytes_bytes_total.
	m.JournalBytes, errs[8] = meter.Int64Counter("jobshop.journal.ship.bytes",
		metric.WithDescription("Journal row bytes shipped to VictoriaLogs, by pool."))
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
	m.CgroupMemory, errs[21] = meter.Int64Gauge("jobshop.pool.cgroup.memory", metric.WithUnit("By"),
		metric.WithDescription("memory.current summed over a pool's live session cgroups (#388)."))
	m.CgroupCPU, errs[22] = meter.Int64Gauge("jobshop.pool.cgroup.cpu", metric.WithUnit("us"),
		metric.WithDescription("cpu.stat usage_usec summed over a pool's live session cgroups (#388)."))
	m.CgroupOOMKills, errs[23] = meter.Int64Gauge("jobshop.pool.cgroup.oom_kills",
		metric.WithDescription("memory.events oom_kill summed over a pool's live session cgroups (#388)."))
	m.JournalErrors, errs[24] = meter.Int64Counter("jobshop.journal.ship.errors",
		metric.WithDescription("Failed journal ship batches, by pool; the batch is retried."))
	m.JournalLag, errs[25] = meter.Float64Gauge("jobshop.journal.ship.lag", metric.WithUnit("s"),
		metric.WithDescription("Age of the oldest unshipped journal frame, by pool; 0 when caught up."))
	m.TickStep, errs[26] = meter.Float64Histogram("jobshop.tick.step.duration", metric.WithUnit("s"),
		metric.WithDescription("Supervisor tick step duration by step (JOB-105)."))
	m.Outbound, errs[27] = meter.Float64Histogram("jobshop.outbound.duration", metric.WithUnit("s"),
		metric.WithDescription("Outbound HTTP call time to response headers by kind and target host (JOB-105)."))
	m.LockWait, errs[28] = meter.Float64Histogram("jobshop.lock.wait", metric.WithUnit("s"),
		metric.WithDescription("Time waiting for the tick lock by lock and waiter (JOB-105)."))
	m.PollerTimeouts, errs[29] = meter.Int64Counter("jobshop.poller.timeouts",
		metric.WithDescription("Calls run off the supervisor pass that ran out their timeout, by poller (JOB-103)."))
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
