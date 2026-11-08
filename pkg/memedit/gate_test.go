package memedit_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/store"
)

var gateT0 = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

func fxStep(id, run, pool string, at time.Time, failures int, failed bool) store.Ticket {
	t := store.Ticket{
		ID: id, Kind: store.KindStep, Pool: pool, ParentID: run, CreatedAt: at,
		Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"a","failures":%d}`, run, failures)),
	}
	if failed {
		t.Status, t.Resolution = store.StatusClosed, "failed"
	}
	return t
}

// TestMetrics: metrics over fixture steps and learnings count only
// the window and the scoped pools; recurrence matches the cited
// learnings by kind and wording.
func TestMetrics(t *testing.T) {
	t.Parallel()
	h := time.Hour
	steps := []store.Ticket{
		fxStep("s1", "r1", "dev", gateT0.Add(h), 2, false),
		fxStep("s2", "r1", "dev", gateT0.Add(2*h), 0, true),
		fxStep("s3", "r2", "dev", gateT0.Add(3*h), 1, false),
		fxStep("s4", "r3", "refinery", gateT0.Add(3*h), 5, true), // other pool
		fxStep("s5", "r4", "dev", gateT0.Add(30*h), 3, false),    // outside window
	}
	cited := []store.Learning{{ID: 1, Kind: store.LearningFailure, Summary: "golangci-lint failed: unused parameter ctx in handler", CreatedAt: gateT0.Add(-h)}}
	ls := []store.Learning{
		cited[0], // before the window
		{ID: 2, Kind: store.LearningFailure, Summary: "golangci-lint failed again: unused parameter ctx in handler", CreatedAt: gateT0.Add(h)},
		{ID: 3, Kind: store.LearningCorrection, Summary: "golangci-lint failed: unused parameter ctx in handler", CreatedAt: gateT0.Add(h)}, // other kind
		{ID: 4, Kind: store.LearningFailure, Summary: "nix build hash mismatch", CreatedAt: gateT0.Add(h)},
		{ID: 5, Kind: store.LearningCorrection, Summary: "review resumed step a (s3): missing test", Ticket: "s3", CreatedAt: gateT0.Add(4 * h)},
		{ID: 6, Kind: store.LearningCorrection, Summary: "review resumed step a (s4): wrong pool", Ticket: "s4", CreatedAt: gateT0.Add(4 * h)},
	}
	m := memedit.Metrics(steps, ls, cited, []string{"dev"}, gateT0, gateT0.Add(24*h), false)
	assert.Equal(t, 2, m.Runs)
	assert.Equal(t, 3, m.Steps)
	assert.Equal(t, 3, m.Retries)
	assert.Equal(t, 1, m.Escalations)
	assert.Equal(t, 1, m.Findings, "only findings on steps of the scoped pools")
	assert.Equal(t, 1, m.Recurrence, "same kind and similar summary")

	all := memedit.Metrics(steps, ls, cited, nil, gateT0, gateT0.Add(24*h), false)
	assert.Equal(t, 4, all.Steps, "no pools: every pool")
	assert.Equal(t, 2, all.Findings)
}

// TestMetricsCitedAfterMerge (#1061): in the window after the merge a
// cited learning is not a recurrence, even one recorded after MergedAt
// by the session that shipped the fix. A similar learning that is not
// cited still is. The baseline window before the merge counts cited ones.
func TestMetricsCitedAfterMerge(t *testing.T) {
	t.Parallel()
	h := time.Hour
	merged := gateT0.Add(10 * h)
	end := merged.Add(24 * h)
	steps := []store.Ticket{fxStep("s1", "r1", "dev", merged.Add(h), 0, false)}
	const text = "curl download failed with a 5xx from the release host"
	cited := []store.Learning{
		{ID: 1, Kind: store.LearningFailure, Summary: text, CreatedAt: merged.Add(-h)},
		// Written after the merge, to describe the fix.
		{ID: 2, Kind: store.LearningWorkaround, Summary: "download from the release host retried on a 5xx failure", CreatedAt: merged.Add(2 * h)},
	}
	ls := []store.Learning{cited[0], cited[1]}

	got := memedit.Metrics(steps, ls, cited, []string{"dev"}, merged, end, true)
	assert.Equal(t, 0, got.Recurrence, "cited learnings recorded after the merge")
	assert.Equal(t, 1, got.Steps)

	// Another cited learning that is similar in wording is not counted either.
	similar := store.Learning{ID: 3, Kind: store.LearningFailure, Summary: text, CreatedAt: merged.Add(3 * h)}
	got = memedit.Metrics(steps, append(ls, similar), append(cited, similar), []string{"dev"}, merged, end, true)
	assert.Equal(t, 0, got.Recurrence, "a third cited learning, similar to the first")

	// The same problem recorded by an uncited learning does count.
	again := store.Learning{ID: 4, Kind: store.LearningFailure, Summary: text, CreatedAt: merged.Add(4 * h)}
	got = memedit.Metrics(steps, append(ls, again), cited, []string{"dev"}, merged, end, true)
	assert.Equal(t, 1, got.Recurrence, "uncited learning that repeats a cited one")

	// Before the merge the cited learnings are the problem: they count.
	base := memedit.Metrics(steps, ls, cited, []string{"dev"}, merged.Add(-24*h), merged, false)
	assert.Equal(t, 1, base.Recurrence)

	// The decision for the reported case: no recurrence, so the edit helped.
	baseline := store.PoolMetrics{Steps: 1000, Recurrence: 1}
	state, reason := memedit.Decide(baseline, memedit.Metrics(steps, ls, cited, []string{"dev"}, merged, end, true))
	assert.Equal(t, store.EditHelped, state, reason)
}

func TestDecide(t *testing.T) {
	t.Parallel()
	base := store.PoolMetrics{Steps: 20, Retries: 4, Escalations: 1, Findings: 1, Recurrence: 3}
	for _, tc := range []struct {
		name  string
		after store.PoolMetrics
		want  string
	}{
		{"no recurrence, same failures", store.PoolMetrics{Steps: 20, Retries: 4, Escalations: 1, Findings: 1}, store.EditHelped},
		{"less recurrence", store.PoolMetrics{Steps: 20, Retries: 3, Recurrence: 1}, store.EditHelped},
		{"no steps since merge", store.PoolMetrics{}, store.EditHelped},
		{"recurs as often", store.PoolMetrics{Steps: 20, Retries: 4, Recurrence: 3}, store.EditNotHelped},
		{"recurs more per step", store.PoolMetrics{Steps: 10, Recurrence: 2}, store.EditNotHelped},
		{"failures rose", store.PoolMetrics{Steps: 20, Retries: 9, Escalations: 2, Findings: 2}, store.EditNotHelped},
		{"failures within noise", store.PoolMetrics{Steps: 20, Retries: 6, Escalations: 1, Findings: 0}, store.EditHelped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := memedit.Decide(base, tc.after)
			assert.Equal(t, tc.want, got, reason)
			assert.NotEmpty(t, reason)
		})
	}
	got, _ := memedit.Decide(store.PoolMetrics{}, store.PoolMetrics{Steps: 5, Recurrence: 1})
	assert.Equal(t, store.EditNotHelped, got, "recurs after none in the baseline")
}
