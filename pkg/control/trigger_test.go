package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// newClock is the fake clock the trigger tests start at.
func newClock() *clock.Fake { return clock.NewFake(time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)) }

// Pools have max_sessions 0, so nothing claims: tests see raw trigger output.
const triggerYAML = `projects:
  js: {}
agents:
  fake:
    command: false
pools:
  dev:
    agent: fake
    dir: %[1]q
  plan:
    agent: fake
    dir: %[1]q
triggers:
  replan:
    when:
      pool: dev
      ready_below: 1
    ticket:
      pool: plan
      title: Plan next step
      priority: 1
      labels: [plan]
    cooldown: 10m
  nightly:
    when:
      cron: "0 6 * * *"
    ticket:
      title: Nightly review
`

// triggerController runs a controller on root with clock clk until the
// returned stop is called.
func triggerController(t *testing.T, root config.Root, clk *clock.Fake) func() {
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- control.Run(ctx, root, withPass(root.Dir, control.Options{Tick: time.Hour, Clock: clk}))
	}()
	return func() { cancel(); require.NoError(t, <-done) }
}

func TestTriggers(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, triggerYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	clk := newClock()
	stop := triggerController(t, root, clk)

	admin := dial(t, root.AdminSocket())
	byTrigger := func(name string) []store.Ticket {
		var all []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{}, &all))
		var out []store.Ticket
		for _, tk := range all {
			if string(tk.Meta) == `{"trigger":{"name":"`+name+`"}}` {
				out = append(out, tk)
			}
		}
		return out
	}
	settle := func() { pass(t, root.Dir, 3) }

	// dev is empty: replan fires once, then waits on its open ticket.
	passUntil(t, root.Dir, func() bool { return len(byTrigger("replan")) == 1 }, "trigger")
	settle()
	plans := byTrigger("replan")
	require.Len(t, plans, 1, "no second ticket while the first is open")
	assert.Equal(t, "plan", plans[0].Pool)
	assert.Equal(t, "Plan next step", plans[0].Title)
	assert.Equal(t, []string{"plan"}, plans[0].Labels)
	assert.Empty(t, byTrigger("nightly"), "cron not due before 06:00")

	var infos []rpc.TriggerInfo
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTriggerList, nil, &infos))
	require.Len(t, infos, 2)
	assert.Equal(t, "nightly", infos[0].Name)
	assert.False(t, infos[0].Holds)
	require.NotNil(t, infos[0].NextCron)
	assert.Equal(t, time.Date(2026, 11, 1, 6, 0, 0, 0, time.UTC), infos[0].NextCron.UTC())
	assert.Equal(t, "replan", infos[1].Name)
	assert.Equal(t, plans[0].ID, infos[1].Open)
	require.NotNil(t, infos[1].LastFired)

	// Closing it starts the cooldown.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: plans[0].ID}, nil))
	clk.Advance(9 * time.Minute)
	settle()
	assert.Len(t, byTrigger("replan"), 1, "still cooling down")
	clk.Advance(2 * time.Minute)
	passUntil(t, root.Dir, func() bool { return len(byTrigger("replan")) == 2 }, "trigger")

	// Work in dev: the condition no longer holds after the ticket closes.
	var work store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "w", Pool: "dev"}, &work))
	for _, tk := range byTrigger("replan") {
		_ = admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: tk.ID}, nil)
	}
	clk.Advance(time.Hour)
	settle()
	assert.Len(t, byTrigger("replan"), 2, "dev has ready work")

	// Cron: fires once at 06:00, not again in the same slot.
	clk.Set(time.Date(2026, 11, 1, 6, 0, 30, 0, time.UTC))
	passUntil(t, root.Dir, func() bool { return len(byTrigger("nightly")) == 1 }, "trigger")
	nightly := byTrigger("nightly")[0]
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: nightly.ID}, nil))
	clk.Advance(time.Hour)
	settle()
	assert.Len(t, byTrigger("nightly"), 1, "one ticket per slot")

	// Downtime across three slots: one ticket on startup, not three.
	stop()
	clk.Set(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	stop = triggerController(t, root, clk)
	defer stop()
	admin = dial(t, root.AdminSocket())
	passUntil(t, root.Dir, func() bool { return len(byTrigger("nightly")) == 2 }, "trigger")
	for _, tk := range byTrigger("nightly") {
		_ = admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: tk.ID}, nil)
	}
	settle()
	assert.Len(t, byTrigger("nightly"), 2, "missed slots collapse into one")
}

// The check reads ./sha from the root: missing means green.
const checkYAML = `projects:
  js: {}
agents:
  fake:
    command: false
pools:
  dev:
    agent: fake
    dir: %[1]q
triggers:
  red:
    when:
      check: cat sha
    ticket:
      pool: dev
      title: Fix master
`

func TestTriggerCheck(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, checkYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	clk := newClock()
	stop := triggerController(t, root, clk)
	defer stop()
	admin := dial(t, root.AdminSocket())
	list := func() []store.Ticket {
		var all []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{}, &all))
		return all
	}
	settle := func() { pass(t, root.Dir, 3) }
	setSHA := func(s string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sha"), []byte(s+"\n"), 0o600))
	}

	settle()
	assert.Empty(t, list(), "check exits non-zero")
	var infos []rpc.TriggerInfo
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTriggerList, nil, &infos))
	require.Len(t, infos, 1)
	assert.Equal(t, "check exited 1", infos[0].Why)

	setSHA("abc")
	passUntil(t, root.Dir, func() bool { return len(list()) == 1 }, "trigger")
	tk := list()[0]
	assert.JSONEq(t, `{"trigger":{"name":"red","key":"red/abc"}}`, string(tk.Meta))
	assert.Contains(t, tk.Body, "abc")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: tk.ID}, nil))
	settle()
	assert.Len(t, list(), 1, "same key does not fire again")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTriggerList, nil, &infos))
	assert.Equal(t, "key abc already filed as "+tk.ID, infos[0].Why)

	setSHA("def")
	passUntil(t, root.Dir, func() bool { return len(list()) == 2 }, "trigger")
}

// TestTicketTrace builds trigger -> plan ticket -> dev ticket the way the
// selfdev loop does and checks that trace links all of it.
func TestTicketTrace(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, triggerYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	clk := newClock()
	stop := triggerController(t, root, clk)
	defer stop()
	admin := dial(t, root.AdminSocket())
	ctx := t.Context()

	var plan store.Ticket
	passUntil(t, root.Dir, func() bool {
		var l []store.Ticket
		require.NoError(t, admin.Call(ctx, rpc.MethodTicketList, rpc.ListParams{Pool: "plan"}, &l))
		if len(l) == 0 {
			return false
		}
		plan = l[0]
		return true
	}, "plan ticket filed")

	// Planner session claims the plan ticket and files a dev ticket.
	var ps rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "fake", Pool: "plan"}, &ps))
	p := dial(t, ps.Socket)
	require.NoError(t, p.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: plan.ID}, nil))
	var dev store.Ticket
	require.NoError(t, p.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{
		Title: "Resolve #70", Pool: "dev", Body: "https://github.com/elee1766/jobshop/issues/70",
	}, &dev))
	require.NoError(t, p.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{Body: "filed https://github.com/elee1766/jobshop/issues/70"}, nil))
	require.NoError(t, p.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{}, nil))
	clk.Advance(time.Second)
	// Created by the planner after it closed the plan ticket: not a child.
	require.NoError(t, p.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "later", Pool: "dev"}, nil))

	// Dev session claims it and records the PR.
	var ds rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "fake", Pool: "dev"}, &ds))
	d := dial(t, ds.Socket)
	require.NoError(t, d.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: dev.ID}, nil))
	require.NoError(t, d.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{Body: "PR https://github.com/elee1766/jobshop/pull/71 head abc"}, nil))

	var tr rpc.TraceNode
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketTrace, rpc.TraceParams{ID: plan.ID}, &tr))
	assert.Equal(t, "replan", tr.Trigger)
	assert.Equal(t, "controller", tr.CreatedBy)
	assert.Equal(t, store.StatusClosed, tr.Status)
	assert.Equal(t, []string{ps.Ticket.ID}, tr.Holders)
	assert.Equal(t, []string{"https://github.com/elee1766/jobshop/issues/70"}, tr.Links)
	require.Len(t, tr.Children, 1, "only the ticket filed while holding the plan")
	c := tr.Children[0]
	assert.Equal(t, dev.ID, c.ID)
	assert.Equal(t, ps.Ticket.ID, c.CreatedBy)
	assert.Equal(t, []string{ds.Ticket.ID}, c.Holders)
	assert.Equal(t, []string{
		"https://github.com/elee1766/jobshop/issues/70",
		"https://github.com/elee1766/jobshop/pull/71",
	}, c.Links)

	// A session traces its current ticket with no id.
	var own rpc.TraceNode
	require.NoError(t, d.Call(ctx, rpc.MethodTicketTrace, rpc.TraceParams{}, &own))
	assert.Equal(t, dev.ID, own.ID)
}

const runTriggerYAML = `projects:
  js: {}
agents:
  fake:
    command: false
pools:
  dev:
    agent: fake
    dir: %[1]q
workflows:
  fix:
    vars: [sha]
    steps:
      - {id: fix, pool: dev, title: "Fix {{.sha}}", prompt: "Fix master at {{.sha}}"}
triggers:
  red:
    when:
      check: cat sha
    run:
      workflow: fix
      vars: {sha: "{{.check_output}}"}
      key: "red-{{.check_output}}"
`

// A check trigger with run: starts one run per check output.
func TestTriggerRun(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, runTriggerYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	clk := newClock()
	stop := triggerController(t, root, clk)
	defer stop()
	admin := dial(t, root.AdminSocket())
	runs := func() []store.Ticket {
		var out []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{}, &out))
		return out
	}
	settle := func() { pass(t, root.Dir, 3) }
	setSHA := func(s string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sha"), []byte(s+"\nextra\n"), 0o600))
	}

	setSHA("abc")
	passUntil(t, root.Dir, func() bool { return len(runs()) == 1 }, "trigger")
	var r rpc.RunResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunShow, rpc.IDParams{ID: runs()[0].ID}, &r))
	assert.Contains(t, string(r.Run.Meta), `"key":"red/abc`)
	require.Len(t, r.Steps, 1)
	assert.Equal(t, "Fix abc", r.Steps[0].Title, "first line of the output")
	var f store.RunFields
	require.NoError(t, json.Unmarshal(r.Run.Fields, &f))
	assert.Equal(t, "red-abc", f.Key)

	settle()
	assert.Len(t, runs(), 1, "open run holds the trigger")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunCancel, rpc.IDParams{ID: r.Run.ID}, nil))
	settle()
	assert.Len(t, runs(), 1, "same output does not start another run")

	setSHA("def")
	passUntil(t, root.Dir, func() bool { return len(runs()) == 2 }, "trigger")
}

const curateTriggerYAML = `projects:
  js: {}
agents:
  fake:
    command: false
pools:
  learn:
    agent: fake
    dir: %[1]q
workflows:
  curate:
    steps:
      - {id: curate, pool: learn, title: "Curate", prompt: "Curate"}
triggers:
  curate:
    when:
      learnings: {min: 3}
    run:
      workflow: curate
      key: curate
`

// TestTriggerLearnings: the curate trigger (#273) fires once open
// learnings reach the threshold, not again while its run is open, and
// after the run closes counts only learnings newer than that run.
func TestTriggerLearnings(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, curateTriggerYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	clk := newClock()
	stop := triggerController(t, root, clk)
	defer stop()
	admin := dial(t, root.AdminSocket())
	runs := func() []store.Ticket {
		var out []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{}, &out))
		return out
	}
	learn := func(n int) {
		for i := range n {
			require.NoError(t, admin.Call(t.Context(), rpc.MethodLearnAdd, rpc.LearnParams{Kind: store.LearningFailure, Summary: fmt.Sprint("l", i)}, nil))
		}
	}
	settle := func() { pass(t, root.Dir, 3) }

	learn(2)
	settle()
	assert.Empty(t, runs(), "below the threshold")
	learn(1)
	passUntil(t, root.Dir, func() bool { return len(runs()) == 1 }, "trigger")
	learn(5)
	settle()
	assert.Len(t, runs(), 1, "open run holds the trigger")

	// Learnings recorded before the run started do not count again.
	clk.Advance(time.Minute)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRunCancel, rpc.IDParams{ID: runs()[0].ID}, nil))
	settle()
	assert.Len(t, runs(), 1, "no new learnings since the last pass")
	learn(3)
	passUntil(t, root.Dir, func() bool { return len(runs()) == 2 }, "trigger")
}

const scopedCurateYAML = `projects:
  js: {}
agents:
  fake:
    command: false
pools:
  learn:
    agent: fake
    dir: %[1]q
learnings:
  kinds: [friction, "review-*"]
workflows:
  curate:
    steps:
      - {id: curate, pool: learn, title: "Curate", prompt: "Curate"}
triggers:
  curate:
    when:
      learnings: {min: 2, kinds: ["review-*", failure], scope: ["repo:acme/widget", "pool:learn"]}
    run:
      workflow: curate
      key: curate
`

// TestTriggerLearningsFiltered: with kinds and scope (#524) the curate
// trigger counts only matching learnings; another repo's and other kinds
// do not fire it. learnings.kinds admits custom and glob kinds.
func TestTriggerLearningsFiltered(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, scopedCurateYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	stop := triggerController(t, root, newClock())
	defer stop()
	admin := dial(t, root.AdminSocket())
	runs := func() []store.Ticket {
		var out []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodRunList, rpc.RunListParams{}, &out))
		return out
	}
	learn := func(kind, scope string) error {
		return admin.Call(t.Context(), rpc.MethodLearnAdd, rpc.LearnParams{Kind: kind, Summary: kind + " " + scope, Scope: scope}, nil)
	}
	err = learn("blocker", "global")
	require.ErrorContains(t, err, "friction, review-*", "unlisted kind names the patterns")
	require.ErrorIs(t, learn("workaround", "global"), store.ErrInvalid, "set kinds replace the defaults")
	require.NoError(t, learn("correction", "repo:acme/widget"), "the controller's kinds stay allowed")
	require.NoError(t, learn("review-nit", "repo:other/repo"), "another repo")
	require.NoError(t, learn("friction", "repo:acme/widget"), "kind not in the trigger")
	require.NoError(t, learn("failure", "global"), "scope not in the trigger")
	pass(t, root.Dir, 3)
	assert.Empty(t, runs(), "nothing matching yet")
	require.NoError(t, learn("review-nit", "repo:acme/widget"))
	require.NoError(t, learn("failure", "pool:learn"))
	passUntil(t, root.Dir, func() bool { return len(runs()) == 1 }, "trigger")
}
