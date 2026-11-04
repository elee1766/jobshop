package control

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

func TestAutoDesired(t *testing.T) {
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
			var hist []scaleSample
			var got int
			for _, s := range tc.steps {
				got, hist = autoDesired(tc.pool, s.demand, t0.Add(s.at), hist)
				assert.Equal(t, s.want, got, "at %s demand %d", s.at, s.demand)
			}
		})
	}
	// Same history, max lowered by reload: clamped without waiting.
	hist := []scaleSample{{at: t0, n: 8}}
	lowered := p
	lowered.MaxSessions = 3
	got, _ := autoDesired(lowered, 0, t0.Add(time.Minute), hist)
	assert.Equal(t, 3, got)
}

func TestPlanPool(t *testing.T) {
	t0 := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	at := func(m int) time.Time { return t0.Add(time.Duration(m) * time.Minute) }
	for _, tc := range []struct {
		name     string
		desired  int
		sessions []scaleSession
		want     scalePlan
	}{
		{"scale up starts the shortfall", 3, []scaleSession{{ID: "a"}}, scalePlan{Start: 2}},
		{
			"scale down picks idle first, never a working session", 1,
			[]scaleSession{
				{ID: "w1", LastClaim: at(1)},
				{ID: "i1", Idle: true, LastClaim: at(5)},
				{ID: "w2", LastClaim: at(2)},
				{ID: "i2", Idle: true, LastClaim: at(3)},
			},
			scalePlan{Cordon: []string{"i2", "i1", "w1"}},
		},
		{
			"least recent claim, then oldest host", 1,
			[]scaleSession{
				{ID: "new", Idle: true, LastClaim: at(1), Created: at(1)},
				{ID: "old", Idle: true, LastClaim: at(1), Created: at(0)},
			},
			scalePlan{Cordon: []string{"old"}},
		},
		{
			"cordoned sessions do not count; scale cordons are lifted before starting", 3,
			[]scaleSession{
				{ID: "a"},
				{ID: "m", Cordon: store.CordonManual},
				{ID: "s1", Cordon: store.CordonScale, Idle: true},
				{ID: "s2", Cordon: store.CordonScale},
			},
			scalePlan{Uncordon: []string{"s2", "s1"}},
		},
		{"manual cordon is never lifted", 2, []scaleSession{{ID: "m", Cordon: store.CordonManual}}, scalePlan{Start: 2}},
		{"steady state", 2, []scaleSession{{ID: "a"}, {ID: "b"}}, scalePlan{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, planPool(tc.desired, tc.sessions))
		})
	}
}

func TestCapStarts(t *testing.T) {
	reqs := func() []capRequest {
		return []capRequest{
			{Pool: "dev", Priority: 2, Start: 5},
			{Pool: "refinery", Priority: 1, Start: 2},
			{Pool: "deploy", Priority: 1, Start: 1},
		}
	}
	assert.Equal(t, map[string]int{"dev": 5, "refinery": 2, "deploy": 1}, capStarts(reqs(), 10, 0), "no cap")
	assert.Equal(t, map[string]int{"deploy": 1, "refinery": 2, "dev": 1}, capStarts(reqs(), 6, 10), "shared by priority")
	assert.Equal(t, map[string]int{"deploy": 1, "refinery": 1}, capStarts(reqs(), 8, 10))
	assert.Empty(t, capStarts(reqs(), 12, 10), "over the cap starts nothing")
}
