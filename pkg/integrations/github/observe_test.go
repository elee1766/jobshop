package github_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/integrations/github"
)

func TestNewestRun(t *testing.T) {
	t.Parallel()
	rs := []github.WorkflowRun{
		{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml"},
		{ID: 3, Name: "CI", Path: ".github/workflows/ci.yml"},
		{ID: 9, Name: "release", Path: ".github/workflows/release.yml"},
	}
	for _, want := range []string{"ci", "CI", "ci.yml", ".github/workflows/ci.yml"} {
		r, ok := github.NewestRun(rs, want)
		assert.True(t, ok, want)
		assert.Equal(t, int64(3), r.ID, want)
	}
	_, ok := github.NewestRun(rs, "lint")
	assert.False(t, ok)
}

func TestNeedsRebase(t *testing.T) {
	t.Parallel()
	for state, want := range map[string][2]bool{
		"dirty": {true, true}, "BEHIND": {true, true}, "clean": {false, true},
		"blocked": {false, true}, "unknown": {false, false}, "": {false, false},
	} {
		needs, known := github.NeedsRebase(state)
		assert.Equal(t, want, [2]bool{needs, known}, state)
	}
}

func TestPRRefs(t *testing.T) {
	t.Parallel()
	got := github.PRRefs("PR https://github.com/a/b/pull/12 head x, again https://github.com/c/d.e/pull/3")
	assert.Equal(t, []github.Thread{{Repo: "a/b", Number: 12, PR: true}, {Repo: "c/d.e", Number: 3, PR: true}}, got)
	assert.Empty(t, github.PRRefs("https://github.com/a/b/issues/12"))
}

func TestRefNumber(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 7, github.RefNumber(float64(7)))
	assert.Equal(t, 7, github.RefNumber("7"))
	assert.Equal(t, 7, github.RefNumber("#7"))
	assert.Equal(t, 0, github.RefNumber(nil))
	assert.Equal(t, 0, github.RefNumber("x"))
}

func TestCheckBucket(t *testing.T) {
	t.Parallel()
	for _, c := range [][3]string{
		{"queued", "", "pending"},
		{"completed", "success", "pass"},
		{"completed", "neutral", "pass"},
		{"completed", "skipped", "skipping"},
		{"completed", "cancelled", "cancel"},
		{"completed", "failure", "fail"},
	} {
		assert.Equal(t, c[2], github.CheckBucket(c[0], c[1]), c[:2])
	}
}
