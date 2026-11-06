package telemetry

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type (
	routeKey   struct{}
	realURLKey struct{}
)

// WithRoute names the route template of the outbound calls made with ctx,
// such as "/repos/{repo}/issues/{n}". It is the only path a span shows:
// real paths carry secrets (webhook URLs, bot tokens), so spans and
// metrics get scheme, host and this template, never the path or query.
func WithRoute(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, routeKey{}, route)
}

// HTTPClient returns a client for outbound calls of the given kind
// ("linear", "github", "hook", ...): each call is a client span under the
// span in the request's context, and its time to response headers goes to
// jobshop_outbound_duration_seconds when ctx carries Metrics (WithMetrics).
// timeout bounds the whole call, body included, and must be positive.
func HTTPClient(kind string, timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: Transport(kind, http.DefaultTransport)}
}

// Transport wraps base as HTTPClient does, for a caller that builds its
// own client.
func Transport(kind string, base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	restore := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// otelhttp has now read its attributes from the safe URL; the wire
		// request goes to the real one.
		r = r.Clone(r.Context())
		if u, ok := r.Context().Value(realURLKey{}).(*url.URL); ok {
			r.URL = u
		}
		return base.RoundTrip(r)
	})
	traced := otelhttp.NewTransport(restore,
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return "HTTP " + r.Method + " " + r.URL.Host + r.URL.Path
		}),
		otelhttp.WithSpanOptions(trace.WithAttributes(AttrOutKind.String(kind))),
	)
	return &outbound{kind: kind, next: traced}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// outbound hides the request's path and query from the span, then
// records the call's duration.
type outbound struct {
	kind string
	next http.RoundTripper
}

func (o *outbound) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	route, _ := r.Context().Value(routeKey{}).(string)
	safe := r.Clone(context.WithValue(r.Context(), realURLKey{}, r.URL))
	safe.URL = &url.URL{Scheme: r.URL.Scheme, Host: r.URL.Host, Path: route}
	resp, err := o.next.RoundTrip(safe)
	if m := MetricsFrom(r.Context()); m != nil {
		m.SinceOutbound(r.Context(), o.kind, r.URL.Host, time.Since(start))
	}
	return resp, err
}
