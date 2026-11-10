package knowledge_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/knowledge"
)

var gateT0 = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

func fxStep(id, run, pool string, at time.Time, failures int, failed bool) domain.Ticket {
	t := domain.Ticket{
		ID: id, Kind: domain.KindStep, Pool: pool, ParentID: run, CreatedAt: at,
		Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"a","failures":%d}`, run, failures)),
	}
	if failed {
		t.Status, t.Resolution = domain.StatusClosed, "failed"
	}
	return t
}

// TestMetrics: metrics over fixture steps and learnings count only
// the window and the scoped pools; recurrence matches the cited
// learnings by kind and wording.
func TestMetrics(t *testing.T) {
	t.Parallel()
	h := time.Hour
	steps := []domain.Ticket{
		fxStep("s1", "r1", "dev", gateT0.Add(h), 2, false),
		fxStep("s2", "r1", "dev", gateT0.Add(2*h), 0, true),
		fxStep("s3", "r2", "dev", gateT0.Add(3*h), 1, false),
		fxStep("s4", "r3", "refinery", gateT0.Add(3*h), 5, true), // other pool
		fxStep("s5", "r4", "dev", gateT0.Add(30*h), 3, false),    // outside window
	}
	cited := []domain.Learning{{ID: 1, Kind: domain.LearningFailure, Summary: "golangci-lint failed: unused parameter ctx in handler", CreatedAt: gateT0.Add(-h)}}
	ls := []domain.Learning{
		cited[0], // before the window
		{ID: 2, Kind: domain.LearningFailure, Summary: "golangci-lint failed again: unused parameter ctx in handler", CreatedAt: gateT0.Add(h)},
		{ID: 3, Kind: domain.LearningCorrection, Summary: "golangci-lint failed: unused parameter ctx in handler", CreatedAt: gateT0.Add(h)}, // other kind
		{ID: 4, Kind: domain.LearningFailure, Summary: "nix build hash mismatch", CreatedAt: gateT0.Add(h)},
		{ID: 5, Kind: domain.LearningCorrection, Summary: "review resumed step a (s3): missing test", Ticket: "s3", CreatedAt: gateT0.Add(4 * h)},
		{ID: 6, Kind: domain.LearningCorrection, Summary: "review resumed step a (s4): wrong pool", Ticket: "s4", CreatedAt: gateT0.Add(4 * h)},
	}
	m := knowledge.EditMetrics(steps, ls, cited, []string{"dev"}, gateT0, gateT0.Add(24*h), false)
	assert.Equal(t, 2, m.Runs)
	assert.Equal(t, 3, m.Steps)
	assert.Equal(t, 3, m.Retries)
	assert.Equal(t, 1, m.Escalations)
	assert.Equal(t, 1, m.Findings, "only findings on steps of the scoped pools")
	assert.Equal(t, 1, m.Recurrence, "same kind and similar summary")

	all := knowledge.EditMetrics(steps, ls, cited, nil, gateT0, gateT0.Add(24*h), false)
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
	steps := []domain.Ticket{fxStep("s1", "r1", "dev", merged.Add(h), 0, false)}
	const text = "curl download failed with a 5xx from the release host"
	cited := []domain.Learning{
		{ID: 1, Kind: domain.LearningFailure, Summary: text, CreatedAt: merged.Add(-h)},
		// Written after the merge, to describe the fix.
		{ID: 2, Kind: domain.LearningWorkaround, Summary: "download from the release host retried on a 5xx failure", CreatedAt: merged.Add(2 * h)},
	}
	ls := []domain.Learning{cited[0], cited[1]}

	got := knowledge.EditMetrics(steps, ls, cited, []string{"dev"}, merged, end, true)
	assert.Equal(t, 0, got.Recurrence, "cited learnings recorded after the merge")
	assert.Equal(t, 1, got.Steps)

	// Another cited learning that is similar in wording is not counted either.
	similar := domain.Learning{ID: 3, Kind: domain.LearningFailure, Summary: text, CreatedAt: merged.Add(3 * h)}
	got = knowledge.EditMetrics(steps, append(ls, similar), append(cited, similar), []string{"dev"}, merged, end, true)
	assert.Equal(t, 0, got.Recurrence, "a third cited learning, similar to the first")

	// The same problem recorded by an uncited learning does count.
	again := domain.Learning{ID: 4, Kind: domain.LearningFailure, Summary: text, CreatedAt: merged.Add(4 * h)}
	got = knowledge.EditMetrics(steps, append(ls, again), cited, []string{"dev"}, merged, end, true)
	assert.Equal(t, 1, got.Recurrence, "uncited learning that repeats a cited one")

	// Before the merge the cited learnings are the problem: they count.
	base := knowledge.EditMetrics(steps, ls, cited, []string{"dev"}, merged.Add(-24*h), merged, false)
	assert.Equal(t, 1, base.Recurrence)

	// The decision for the reported case: no recurrence, so the edit helped.
	baseline := domain.PoolMetrics{Steps: 1000, Recurrence: 1}
	state, reason := knowledge.DecideEdit(baseline, knowledge.EditMetrics(steps, ls, cited, []string{"dev"}, merged, end, true))
	assert.Equal(t, domain.EditHelped, state, reason)
}

func TestDecide(t *testing.T) {
	t.Parallel()
	base := domain.PoolMetrics{Steps: 20, Retries: 4, Escalations: 1, Findings: 1, Recurrence: 3}
	for _, tc := range []struct {
		name  string
		after domain.PoolMetrics
		want  string
	}{
		{"no recurrence, same failures", domain.PoolMetrics{Steps: 20, Retries: 4, Escalations: 1, Findings: 1}, domain.EditHelped},
		{"less recurrence", domain.PoolMetrics{Steps: 20, Retries: 3, Recurrence: 1}, domain.EditHelped},
		{"no steps since merge", domain.PoolMetrics{}, domain.EditHelped},
		{"recurs as often", domain.PoolMetrics{Steps: 20, Retries: 4, Recurrence: 3}, domain.EditNotHelped},
		{"recurs more per step", domain.PoolMetrics{Steps: 10, Recurrence: 2}, domain.EditNotHelped},
		{"failures rose", domain.PoolMetrics{Steps: 20, Retries: 9, Escalations: 2, Findings: 2}, domain.EditNotHelped},
		{"failures within noise", domain.PoolMetrics{Steps: 20, Retries: 6, Escalations: 1, Findings: 0}, domain.EditHelped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := knowledge.DecideEdit(base, tc.after)
			assert.Equal(t, tc.want, got, reason)
			assert.NotEmpty(t, reason)
		})
	}
	got, _ := knowledge.DecideEdit(domain.PoolMetrics{}, domain.PoolMetrics{Steps: 5, Recurrence: 1})
	assert.Equal(t, domain.EditNotHelped, got, "recurs after none in the baseline")
}
