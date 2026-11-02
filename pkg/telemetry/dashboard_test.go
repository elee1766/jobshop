package telemetry

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

// The dashboard may only reference series and labels that /metrics really
// exports, so renaming an instrument breaks this test instead of a panel.
func TestGrafanaDashboardMetrics(t *testing.T) {
	b, err := os.ReadFile("../../deploy/grafana/jobshop.json")
	require.NoError(t, err)
	var dash struct {
		Templating struct {
			List []struct {
				Query any `json:"query"`
			} `json:"list"`
		} `json:"templating"`
		Panels []struct {
			Title   string `json:"title"`
			Targets []struct {
				Expr string `json:"expr"`
			} `json:"targets"`
		} `json:"panels"`
	}
	require.NoError(t, json.Unmarshal(b, &dash))
	require.NotEmpty(t, dash.Panels)

	known := exportedNames(t)
	tok := regexp.MustCompile(`\b(jobshop_[a-z_]+|rpc_method)\b`)
	check := func(where, expr string) {
		for _, name := range tok.FindAllString(expr, -1) {
			assert.Contains(t, known, name, "%s: %s", where, expr)
		}
	}
	for _, p := range dash.Panels {
		require.NotEmpty(t, p.Targets, p.Title)
		for _, tg := range p.Targets {
			require.NotEmpty(t, tg.Expr, p.Title)
			check(p.Title, tg.Expr)
		}
	}
	for _, v := range dash.Templating.List {
		if s, ok := v.Query.(string); ok {
			check("templating", s)
		}
	}
}

// exportedNames records every instrument once, scrapes /metrics and returns
// the series and label names it saw.
func exportedNames(t *testing.T) map[string]bool {
	m, err := NewMetrics()
	require.NoError(t, err)
	defer func() { _ = m.Shutdown(context.Background()) }()
	ctx := context.Background()
	all := metric.WithAttributes(AttrProject.String("p"), AttrPool.String("x"), AttrState.String("idle"),
		AttrKind.String("task"), AttrResolution.String("done"), AttrStopReason.String("end_turn"))
	m.Sessions.Record(ctx, 1, all)
	m.Restarts.Add(ctx, 1, all)
	m.Ready.Record(ctx, 1, all)
	m.Claims.Add(ctx, 1, all)
	m.Closes.Add(ctx, 1, all)
	m.Turns.Add(ctx, 1, all)
	m.SessionsEnded.Add(ctx, 1, all)
	m.Orphaned.Record(ctx, 1, metric.WithAttributes(AttrReason.String("no_pool")))
	m.LeaseExpiries.Add(ctx, 1)
	m.Since(ctx, time.Now(), "ping")

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	names := map[string]bool{}
	for _, mm := range regexp.MustCompile(`(?m)^([a-zA-Z_:][a-zA-Z0-9_:]*)`).FindAllStringSubmatch(string(body), -1) {
		names[mm[1]] = true
	}
	for _, mm := range regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)="`).FindAllStringSubmatch(string(body), -1) {
		names[mm[1]] = true
	}
	return names
}
