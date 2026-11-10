package triggers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// Two triggers that share a script and signal but not a repo or args make
// separate calls: the self-dev pack's issue_no_run read took the github
// pack's cached answer, so TestPacks intermittently started a run for an
// issue of another repo.
func TestScriptSignalKey(t *testing.T) {
	t.Parallel()
	sig := config.ScanSignal{Signal: config.SignalIssueNoRun, Script: "triggers.star:issue_no_run", Repo: "acme/widget", Args: map[string]any{"b": "2", "a": "1"}}
	same := sig
	same.Args = map[string]any{"a": "1", "b": "2"}
	assert.Equal(t, scriptSignalKey(sig), scriptSignalKey(same), "map order does not matter")
	otherRepo := sig
	otherRepo.Repo = "acme/other"
	assert.NotEqual(t, scriptSignalKey(sig), scriptSignalKey(otherRepo))
	otherArgs := sig
	otherArgs.Args = map[string]any{"a": "1", "b": "3"}
	assert.NotEqual(t, scriptSignalKey(sig), scriptSignalKey(otherArgs))
}

func TestFindingID(t *testing.T) {
	t.Parallel()
	f := Finding{Signal: "blocked", Subject: "js-1"}
	assert.Equal(t, "blocked:js-1", f.ID())
	assert.Equal(t, "scan:blocked:js-1", f.KeyWith("scan"), "the id is the same under any key prefix")
	assert.Equal(t, "review:blocked:js-1", f.KeyWith("review"))
}

func TestEventKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "abc", Event{"head_sha": "abc", "delivery": "d"}.Key(), "the commit wins")
	assert.Equal(t, "d", Event{"delivery": "d"}.Key(), "else the delivery")
}

// An event value a shell could misread is left empty for a signal check,
// as on a poll (JOB-72).
func TestEventShellVars(t *testing.T) {
	t.Parallel()
	ev := Event{"repo": "acme/app", "branch": "x; rm -rf /", "head_sha": "abc"}
	v := ev.ShellVars()
	assert.Equal(t, "acme/app", v["repo"])
	assert.Equal(t, "abc", v["head_sha"])
	assert.Equal(t, "", v["branch"])
	assert.Equal(t, "x; rm -rf /", ev.Vars()["branch"], "Vars keeps it for templates")
}

func TestPollDue(t *testing.T) {
	t.Parallel()
	e := New(nil)
	// A host is needed only for the clock of a fallback poll, so these
	// cases cover the ones that decide without it.
	assert.True(t, e.pollDue("a", config.Trigger{}), "no on, no fallback: every tick")
	assert.False(t, e.pollDue("b", config.Trigger{On: []config.TriggerEvent{{Source: "github"}}}), "on without when conditions never polls")
}

// pollClock is a Host that has only a clock, which is all pollDue reads.
type pollClock struct {
	Host
	now time.Time
}

func (p *pollClock) Now() time.Time { return p.now }

// TestPollDueFallbackWithoutOn: a trigger with fallback and no on: polls
// once per fallback, not every tick (JOB-191: selfdev's evaluate trigger).
func TestPollDueFallbackWithoutOn(t *testing.T) {
	t.Parallel()
	h := &pollClock{now: time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)}
	e := New(h)
	tr := config.Trigger{Fallback: "10m", Handler: "triggers.star:start"}
	assert.True(t, e.pollDue("evaluate", tr), "first tick polls")
	h.now = h.now.Add(50 * time.Millisecond)
	assert.False(t, e.pollDue("evaluate", tr), "the next tick does not")
	h.now = h.now.Add(9 * time.Minute)
	assert.False(t, e.pollDue("evaluate", tr), "nor before the fallback")
	h.now = h.now.Add(time.Minute)
	assert.True(t, e.pollDue("evaluate", tr), "at the fallback")
	assert.True(t, e.pollDue("other", tr), "each trigger has its own clock")
}
