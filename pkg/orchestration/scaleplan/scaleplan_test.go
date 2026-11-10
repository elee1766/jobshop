package scaleplan_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/scaleplan"
)

func TestAutoDesired(t *testing.T) {
	t.Parallel()
	t0 := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	p := config.Pool{MaxSessions: 8, MinSessions: 1, ScaleDownAfter: "10m"}
	type step struct {
		at     time.Duration
		demand int
		want   int
	}
	for _, tc := range []struct {
		name  string
		pool  config.Pool
		steps []step
	}{
		{"burst up is immediate and capped", p, []step{{0, 0, 1}, {time.Minute, 20, 8}}},
		{"min kept warm with no work", p, []step{{0, 0, 1}, {time.Hour, 0, 1}}},
		{"stabilization holds the peak for the window", p, []step{
			{0, 5, 5}, {time.Minute, 0, 5}, {9 * time.Minute, 1, 5}, {10 * time.Minute, 1, 1},
		}},
		{"reload lowering max applies at once", p, []step{{0, 8, 8}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hist []scaleplan.Sample
			var got int
			for _, s := range tc.steps {
				got, hist = scaleplan.Desired(tc.pool, s.demand, t0.Add(s.at), hist)
				assert.Equal(t, s.want, got, "at %s demand %d", s.at, s.demand)
			}
		})
	}
	// Same history, max lowered by reload: clamped without waiting.
	hist := []scaleplan.Sample{{At: t0, N: 8}}
	lowered := p
	lowered.MaxSessions = 3
	got, _ := scaleplan.Desired(lowered, 0, t0.Add(time.Minute), hist)
	assert.Equal(t, 3, got)
}

func TestPlanPool(t *testing.T) {
	t.Parallel()
	t0 := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	at := func(m int) time.Time { return t0.Add(time.Duration(m) * time.Minute) }
	for _, tc := range []struct {
		name     string
		desired  int
		sessions []scaleplan.Session
		want     scaleplan.Plan
	}{
		{"scale up starts the shortfall", 3, []scaleplan.Session{{ID: "a"}}, scaleplan.Plan{Start: 2}},
		{
			"scale down picks idle first, never a working session", 1,
			[]scaleplan.Session{
				{ID: "w1", LastClaim: at(1)},
				{ID: "i1", Idle: true, LastClaim: at(5)},
				{ID: "w2", LastClaim: at(2)},
				{ID: "i2", Idle: true, LastClaim: at(3)},
			},
			scaleplan.Plan{Cordon: []string{"i2", "i1", "w1"}},
		},
		{
			"least recent claim, then oldest host", 1,
			[]scaleplan.Session{
				{ID: "new", Idle: true, LastClaim: at(1), Created: at(1)},
				{ID: "old", Idle: true, LastClaim: at(1), Created: at(0)},
			},
			scaleplan.Plan{Cordon: []string{"old"}},
		},
		{
			"cordoned sessions do not count; scale cordons are lifted before starting", 3,
			[]scaleplan.Session{
				{ID: "a"},
				{ID: "m", Cordon: domain.CordonManual},
				{ID: "s1", Cordon: domain.CordonScale, Idle: true},
				{ID: "s2", Cordon: domain.CordonScale},
			},
			scaleplan.Plan{Uncordon: []string{"s2", "s1"}},
		},
		{"manual cordon is never lifted", 2, []scaleplan.Session{{ID: "m", Cordon: domain.CordonManual}}, scaleplan.Plan{Start: 2}},
		{"steady state", 2, []scaleplan.Session{{ID: "a"}, {ID: "b"}}, scaleplan.Plan{}},
		{
			"a resumed session over the cap is not cordoned and nothing starts (JOB-189)", 1,
			[]scaleplan.Session{{ID: "a"}, {ID: "r", Exempt: true}},
			scaleplan.Plan{},
		},
		{
			"a resumed session fills a slot, so one fewer starts (JOB-189)", 2,
			[]scaleplan.Session{{ID: "r", Exempt: true}},
			scaleplan.Plan{Start: 1},
		},
		{
			"a cordoned resumed session is planned like any cordoned one", 1,
			[]scaleplan.Session{{ID: "r", Exempt: true, Cordon: domain.CordonScale}},
			scaleplan.Plan{Uncordon: []string{"r"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, scaleplan.ForPool(tc.desired, tc.sessions))
		})
	}
}

func TestCapStarts(t *testing.T) {
	t.Parallel()
	reqs := func() []scaleplan.CapRequest {
		return []scaleplan.CapRequest{
			{Pool: "dev", Priority: 2, Start: 5},
			{Pool: "refinery", Priority: 1, Start: 2},
			{Pool: "deploy", Priority: 1, Start: 1},
		}
	}
	assert.Equal(t, map[string]int{"dev": 5, "refinery": 2, "deploy": 1}, scaleplan.CapStarts(reqs(), 10, 0), "no cap")
	assert.Equal(t, map[string]int{"deploy": 1, "refinery": 2, "dev": 1}, scaleplan.CapStarts(reqs(), 6, 10), "shared by priority")
	assert.Equal(t, map[string]int{"deploy": 1, "refinery": 1}, scaleplan.CapStarts(reqs(), 8, 10))
	assert.Empty(t, scaleplan.CapStarts(reqs(), 12, 10), "over the cap starts nothing")
}
