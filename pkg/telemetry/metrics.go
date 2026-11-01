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
	var errs [7]error
	m.Sessions, errs[0] = meter.Int64Gauge("jobshop.sessions", metric.WithDescription("Open sessions by pool and state."))
	m.Restarts, errs[1] = meter.Int64Counter("jobshop.session.restarts", metric.WithDescription("Session agent or host restarts."))
	m.Ready, errs[2] = meter.Int64Gauge("jobshop.tickets.ready", metric.WithDescription("Ready tickets per pool, from the supervisor tick."))
	m.Claims, errs[3] = meter.Int64Counter("jobshop.claims", metric.WithDescription("Successful claims."))
	m.Closes, errs[4] = meter.Int64Counter("jobshop.closes", metric.WithDescription("Ticket closes by resolution."))
	m.LeaseExpiries, errs[5] = meter.Int64Counter("jobshop.lease.expiries", metric.WithDescription("Claims dropped by lease expiry."))
	m.RequestTime, errs[6] = meter.Float64Histogram("jobshop.rpc.duration", metric.WithUnit("s"),
		metric.WithDescription("Socket request latency by method."))
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
