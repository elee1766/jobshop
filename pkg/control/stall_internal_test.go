package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

const stallYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
workflows:
  stall-check:
    vars: [signal, subject, detail]
    escalate_pool: plan
    steps:
      - {id: diagnose, pool: plan, retries: 1, title: "Stall {{.signal}} {{.subject}}", prompt: "Check {{.subject}}: {{.detail}}"}
  work:
    vars: [n]
    issue: n
    repo: acme/widget
    steps:
      - {id: a, pool: dev, retries: 1, title: "A {{.n}}", prompt: a}
      - {id: b, pool: dev, title: "B {{.n}}", prompt: b}
triggers:
  stall:
    when:
      stall:
        - {signal: ready_unclaimed, after: 1h}
        - {signal: pool_quarantined}
        - {signal: step_stalled, after: 1h}
        - {signal: session_silent, after: 30m}
        - {signal: needs_human, after: 2h}
        - {signal: blocked, after: 1h}
        - {signal: issue_no_run, repo: acme/widget, check: "cat %[1]s/issues"}
    run:
      workflow: stall-check
      vars: {signal: "{{.stall_signal}}", subject: "{{.stall_subject}}", detail: "{{.stall_detail}}"}
    cooldown: 10m
    max_runs: 2
`

// stallClock is a clock.Fake with the lower-case helpers these tests use.
type stallClock struct{ *clock.Fake }

func newStallClock(t time.Time) *stallClock { return &stallClock{clock.NewFake(t)} }

func (c *stallClock) now() time.Time { return c.Now() }

func (c *stallClock) advance(d time.Duration) { c.Advance(d) }

// stallController builds a controller without serving: the test calls the
// stall functions directly with a fake clock.
func stallController(t *testing.T) (*Controller, *stallClock, string) {
	t.Helper()
	return stallControllerYAML(t, stallYAML)
}

// stallControllerYAML is stallController with a config whose %[1]s is
// the temp dir.
func stallControllerYAML(t *testing.T, yaml string) (*Controller, *stallClock, string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, yaml, dir), 0o600))
	root := config.Root{Dir: dir}
	cfg, err := config.Load(root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	live := new(atomic.Pointer[liveConfig])
	live.Store(lc)
	clk := newStallClock(time.Date(2026, 11, 2, 8, 0, 0, 0, time.UTC))
	st, err := sqlite.Open(t.Context(), filepath.Join(dir, "db"), sqlite.Options{Now: clk.now, Prefix: func(p string) string { return p }})
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	c := &Controller{
		root: root, live: live, store: st, opt: Options{Clock: clk}, log: slog.New(slog.DiscardHandler),
		sessions: map[string]context.CancelFunc{}, runners: map[string]*runner{},
		drained: map[string]bool{}, poolCrashLog: map[string][]time.Time{}, poolPausedUntil: map[string]time.Time{},
	}
	c.ctx = t.Context()
	return c, clk, dir
}

func stallKeys(t *testing.T, c *Controller) []string {
	t.Helper()
	found, err := c.detectStalls(t.Context(), c.config().Triggers["stall"], nil)
	require.NoError(t, err)
	var out []string
	for _, s := range found {
		out = append(out, s.key())
	}
	return out
}

func mkTicket(t *testing.T, c *Controller, nt store.NewTicket) store.Ticket {
	t.Helper()
	if nt.Project == "" {
		nt.Project = "js"
	}
	if nt.Kind == "" {
		nt.Kind = store.KindTask
	}
	tk, err := store.Create(t.Context(), c.store, "admin", nt)
	require.NoError(t, err)
	return tk
}

func TestStallReadyUnclaimed(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	tk := mkTicket(t, c, store.NewTicket{Title: "x", Pool: "dev"})
	mkTicket(t, c, store.NewTicket{Title: "orphan", Pool: "gone"})
	assert.Empty(t, stallKeys(t, c), "fresh ticket")
	clk.advance(30 * time.Minute)
	_, err := store.AddComment(t.Context(), c.store, "admin", tk.ID, "touch")
	require.NoError(t, err)
	clk.advance(50 * time.Minute)
	assert.Empty(t, stallKeys(t, c), "a comment resets the clock")
	clk.advance(11 * time.Minute)
	assert.Equal(t, []string{"stall:ready_unclaimed:" + tk.ID}, stallKeys(t, c), "orphans are not ready_unclaimed")
}

func TestStallReadyUnclaimedSinceUnblock(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	ctx := t.Context()
	blocked := func() (store.Ticket, store.Ticket) {
		step := mkTicket(t, c, store.NewTicket{Title: "step", Pool: "dev"})
		b := mkTicket(t, c, store.NewTicket{Title: "blocker"})
		require.NoError(t, store.AddEdge(ctx, c.store, "admin", store.Edge{From: step.ID, To: b.ID, Kind: store.EdgeBlocks}))
		return step, b
	}
	// Blocker closed 2h ago: the step is stalled.
	old, ob := blocked()
	_, err := store.Close(ctx, c.store, "admin", ob.ID, nil, "done")
	require.NoError(t, err)
	// Created 2h ago, blocker closes 1m ago: not stalled.
	fresh, fb := blocked()
	clk.advance(2*time.Hour - time.Minute)
	_, err = store.Close(ctx, c.store, "admin", fb.ID, nil, "done")
	require.NoError(t, err)
	clk.advance(time.Minute)
	keys := stallKeys(t, c)
	assert.Contains(t, keys, "stall:ready_unclaimed:"+old.ID)
	assert.NotContains(t, keys, "stall:ready_unclaimed:"+fresh.ID)
}

func TestStallPoolQuarantined(t *testing.T) {
	t.Parallel()
	c, _, _ := stallController(t)
	assert.Empty(t, stallKeys(t, c))
	for range poolCrashes {
		c.poolCrashed("dev", "")
	}
	assert.Equal(t, []string{"stall:pool_quarantined:dev@20261102T081500Z"}, stallKeys(t, c))
}

// TestPoolPauseNamesSpawnError (JOB-27): a crash-loop pause carries the
// last session error and the failed preflight checks into attention and
// the stall detail.
func TestPoolPauseNamesSpawnError(t *testing.T) {
	t.Parallel()
	c, _, _ := stallControllerYAML(t, strings.Replace(stallYAML, "dev: {agent: fake, dir: %[1]q}",
		"dev: {agent: fake, dir: %[1]q, sandbox: {backend: bubblewrap, network: pasta}}", 1))
	c.opt.Preflight = func(host.Spec, ...host.PreflightOption) (host.PreflightReport, error) {
		return host.PreflightReport{Checks: []host.PreflightCheck{
			{Name: "sandbox.bwrap", OK: true},
			{Name: "network.nft", Reason: "nft not found", Hint: "install nftables"},
		}}, nil
	}
	spawnErr := `sandbox: deny: exec: "nft": executable file not found`
	for range poolCrashes {
		c.poolCrashed("dev", spawnErr)
	}
	found, err := c.detectStalls(t.Context(), c.config().Triggers["stall"], nil)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Contains(t, found[0].detail, spawnErr)
	assert.Contains(t, found[0].detail, "FAIL network.nft: nft not found")
	assert.Contains(t, found[0].detail, "fix: install nftables")
	assert.NotContains(t, found[0].detail, "sandbox.bwrap", "passing checks are left out")

	items, err := c.attentionItems(t.Context())
	require.NoError(t, err)
	var pool []rpc.AttentionItem
	for _, it := range items {
		if it.Kind == rpc.AttentionPool {
			pool = append(pool, it)
		}
	}
	require.Len(t, pool, 1)
	assert.Equal(t, "dev", pool[0].ID)
	assert.Contains(t, pool[0].Title, spawnErr)
	assert.Contains(t, pool[0].Summary, spawnErr)

	c.unquarantinePools("dev")
	items, err = c.attentionItems(t.Context())
	require.NoError(t, err)
	for _, it := range items {
		assert.NotEqual(t, rpc.AttentionPool, it.Kind)
	}
}

func TestStallStepAndSession(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	res, err := c.startRun(t.Context(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}, Key: "w1"}, nil)
	require.NoError(t, err)
	steps := res.(rpc.RunResult).Steps
	sess := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, _, err = c.store.Claim(t.Context(), steps[0].ID, sess.ID, time.Hour)
	require.NoError(t, err)

	clk.advance(31 * time.Minute)
	assert.Equal(t, []string{"stall:session_silent:" + sess.ID}, stallKeys(t, c), "no journal yet")

	// Journal output resets session_silent; the step still ages.
	sd := c.root.SessionDir(sess.ID)
	require.NoError(t, os.MkdirAll(sd, 0o700))
	line := fmt.Sprintf(`{"seq":1,"t":%d}`+"\n", clk.now().UnixMilli())
	require.NoError(t, os.WriteFile(filepath.Join(sd, host.JournalFile), []byte(line), 0o600))
	clk.advance(28 * time.Minute)
	assert.Empty(t, stallKeys(t, c))
	clk.advance(3 * time.Minute)
	assert.ElementsMatch(t, []string{"stall:step_stalled:" + steps[0].ID, "stall:session_silent:" + sess.ID}, stallKeys(t, c))
}

func TestStallNeedsHumanAndBlocked(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	res, err := c.startRun(t.Context(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}}, nil)
	require.NoError(t, err)
	run := res.(rpc.RunResult).Run
	f, err := store.DecodeFields[store.RunFields](run)
	require.NoError(t, err)
	_, err = c.setRunState(t.Context(), "controller", run, f, store.RunNeedsHuman, "")
	require.NoError(t, err)
	_, err = store.Close(t.Context(), c.store, "admin", res.(rpc.RunResult).Steps[0].ID, nil, "failed")
	require.NoError(t, err)
	// An escalation ticket is covered by its run.
	mkTicket(t, c, store.NewTicket{Title: "Escalate", Labels: []string{"needs-human"}, Meta: json.RawMessage(`{"run":{"id":"` + run.ID + `"}}`)})
	nh := mkTicket(t, c, store.NewTicket{Title: "waits", Pool: "plan", Labels: []string{"needs-human"}, DeferUntil: clk.now().Add(5 * time.Hour)})
	bl := mkTicket(t, c, store.NewTicket{Title: "blocked", Pool: "plan", Labels: []string{"blocked"}, DeferUntil: clk.now().Add(5 * time.Hour)})

	clk.advance(2*time.Hour + time.Minute)
	assert.Equal(t, []string{"stall:needs_human:" + run.ID}, stallKeys(t, c), "deferred tickets wait from their defer time")
	clk.advance(5 * time.Hour)
	assert.ElementsMatch(t, []string{"stall:needs_human:" + run.ID, "stall:needs_human:" + nh.ID, "stall:blocked:" + bl.ID}, stallKeys(t, c))
}

func TestStallCheckSignal(t *testing.T) {
	t.Parallel()
	c, _, dir := stallController(t)
	assert.Empty(t, stallKeys(t, c), "check fails without the file: no stall")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "issues"), []byte("issue-7 open, no run\n\nissue-9\n"), 0o600))
	found, err := c.detectStalls(t.Context(), c.config().Triggers["stall"], nil)
	require.NoError(t, err)
	require.Len(t, found, 2)
	assert.Equal(t, stall{"issue_no_run", "issue-7", "open, no run"}, found[0])
	assert.Equal(t, "issue-9", found[1].subject)

	// An issue with an open run in the same repo (meta.github) is not a
	// stall; the same number in another repo still is (#183).
	require.NoError(t, os.WriteFile(filepath.Join(dir, "issues"), []byte("#12\n13\n"), 0o600))
	mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"work","state":"running"}`), Meta: json.RawMessage(`{"github":{"repo":"acme/widget","issue":12}}`)})
	mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "r2", Fields: json.RawMessage(`{"workflow":"work","state":"running"}`), Meta: json.RawMessage(`{"github":{"repo":"other/repo","issue":13}}`)})
	assert.Equal(t, []string{"stall:issue_no_run:13"}, stallKeys(t, c))
}

// TestStallFire covers the run per key, its exclusion from detection,
// cooldown and max_runs.
func TestStallFire(t *testing.T) {
	t.Parallel()
	c, clk, dir := stallController(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "issues"), []byte("issue-7 no run\n"), 0o600))
	runs := func() []store.Ticket {
		rs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindRun, Order: store.OrderNewest})
		require.NoError(t, err)
		var out []store.Ticket
		for _, r := range rs {
			if f, _ := store.DecodeFields[store.RunFields](r); f.Workflow == "stall-check" {
				out = append(out, r)
			}
		}
		return out
	}
	c.runTriggers()
	rs := runs()
	require.Len(t, rs, 1)
	f, err := store.DecodeFields[store.RunFields](rs[0])
	require.NoError(t, err)
	assert.Equal(t, "stall:issue_no_run:issue-7", f.Key)
	assert.Equal(t, map[string]string{"signal": "issue_no_run", "subject": "issue-7", "detail": "no run"}, f.Vars)
	steps, err := store.Children(t.Context(), c.store, rs[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Stall issue_no_run issue-7", steps[0].Title)

	// Its own step sits ready and unclaimed: not a stall of its own.
	clk.advance(2 * time.Hour)
	assert.Equal(t, []string{"stall:issue_no_run:issue-7"}, stallKeys(t, c))
	c.runTriggers()
	assert.Len(t, runs(), 1, "open run holds the key")

	closeRun := func() {
		_, err := store.Close(t.Context(), c.store, "admin", steps[0].ID, nil, "done")
		require.NoError(t, err)
		c.completeRun(t.Context(), rs[0].ID)
	}
	closeRun()
	c.runTriggers()
	assert.Len(t, runs(), 1, "cooldown")
	clk.advance(11 * time.Minute)
	c.runTriggers()
	rs = runs()
	require.Len(t, rs, 2, "stall still there after cooldown: second run")
	steps, err = store.Children(t.Context(), c.store, rs[0].ID)
	require.NoError(t, err)
	closeRun()
	clk.advance(11 * time.Minute)
	c.runTriggers()
	assert.Len(t, runs(), 2, "max_runs 2")

	var infos []rpc.TriggerInfo
	res, err := c.triggerList(t.Context(), scope{admin: true}, struct{}{})
	require.NoError(t, err)
	infos = res.([]rpc.TriggerInfo)
	require.Len(t, infos, 1)
	assert.False(t, infos[0].Holds)
	assert.Contains(t, infos[0].Why, "max_runs 2 reached")
}

func TestStallMaxOpen(t *testing.T) {
	t.Parallel()
	c, _, dir := stallController(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "issues"), []byte("a\nb\nc\nd\ne\n"), 0o600))
	c.runTriggers()
	c.runTriggers()
	rs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindRun, Order: store.OrderNewest})
	require.NoError(t, err)
	assert.Len(t, rs, stallMaxOpen)
}

func TestStallConfig(t *testing.T) {
	t.Parallel()
	base := `projects: {js: {}}
agents: {a: {command: x}}
pools: {plan: {agent: a, dir: /w}}
workflows:
  sc:
    vars: [s]
    steps: [{id: x, pool: plan, title: t, prompt: p}]
triggers:
  st:
`
	for name, tc := range map[string]struct{ yaml, err string }{
		"ok":              {"    when: {stall: [{signal: blocked, after: 1h}, {signal: x, check: 'true'}]}\n    run: {workflow: sc, vars: {s: '{{.stall_subject}}'}}\n", ""},
		"no after":        {"    when: {stall: [{signal: blocked}]}\n    run: {workflow: sc, vars: {s: x}}\n", "needs after"},
		"no check":        {"    when: {stall: [{signal: pr_unmerged, after: 1h}]}\n    run: {workflow: sc, vars: {s: x}}\n", "needs a check"},
		"check":           {"    when: {stall: [{signal: blocked, after: 1h, check: 'true'}]}\n    run: {workflow: sc, vars: {s: x}}\n", "takes no check"},
		"ticket":          {"    when: {stall: [{signal: blocked, after: 1h}]}\n    ticket: {title: x}\n", "needs run"},
		"key":             {"    when: {stall: [{signal: blocked, after: 1h}]}\n    run: {workflow: sc, vars: {s: x}, key: 'pr-{{.stall_subject}}'}\n", ""},
		"bad key":         {"    when: {stall: [{signal: blocked, after: 1h}]}\n    run: {workflow: sc, vars: {s: x}, key: '{{.check_output}}'}\n", "key uses"},
		"bad var":         {"    when: {stall: [{signal: blocked, after: 1h}]}\n    run: {workflow: sc, vars: {s: '{{.check_output}}'}}\n", "only"},
		"mixed":           {"    when: {ready_below: 1, stall: [{signal: blocked, after: 1h}]}\n    run: {workflow: sc, vars: {s: x}}\n", "no other condition"},
		"max_runs":        {"    when: {cron: '0 6 * * *'}\n    ticket: {title: x}\n    max_runs: 2\n", "max_runs needs"},
		"duplicated":      {"    when: {stall: [{signal: blocked, after: 1h}, {signal: blocked, after: 2h}]}\n    run: {workflow: sc, vars: {s: x}}\n", "twice"},
		"on":              {"    on: [{github: {event: pull_request}}]\n    fallback: 15m\n    when: {stall: [{signal: x, check: 'gh pr view {{.event.pr}}'}]}\n    run: {workflow: sc, vars: {s: x}}\n", ""},
		"on bad fallback": {"    on: [{github: {event: pull_request}}]\n    fallback: soon\n    when: {stall: [{signal: x, check: 'true'}]}\n    run: {workflow: sc, vars: {s: x}}\n", "fallback"},
		"on bad source":   {"    on: [{}]\n    when: {stall: [{signal: x, check: 'true'}]}\n    run: {workflow: sc, vars: {s: x}}\n", "exactly one"},
		"on check var":    {"    on: [{github: {event: pull_request}}]\n    when: {stall: [{signal: x, check: 'echo {{.stall_subject}}'}]}\n    run: {workflow: sc, vars: {s: x}}\n", "only"},
	} {
		_, err := config.Parse([]byte(base + tc.yaml))
		if tc.err == "" {
			assert.NoError(t, err, name)
		} else if assert.Error(t, err, name) {
			assert.Contains(t, err.Error(), tc.err, name)
		}
	}
}

// Two packs' stall triggers watch needs_human with their own workflows.
const stallCascadeYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
workflows:
  check-a:
    vars: [subject]
    steps: [{id: diagnose, pool: plan, retries: 0, title: "A {{.subject}}", prompt: p}]
  check-b:
    vars: [subject]
    steps: [{id: diagnose, pool: plan, retries: 0, title: "B {{.subject}}", prompt: p}]
  work:
    vars: [n]
    steps: [{id: a, pool: dev, retries: 0, title: "W {{.n}}", prompt: a}]
triggers:
  a:
    when: {stall: [{signal: needs_human, after: 2h}]}
    run: {workflow: check-a, vars: {subject: "{{.stall_subject}}"}}
  b:
    when: {stall: [{signal: needs_human, after: 2h}]}
    run: {workflow: check-b, vars: {subject: "{{.stall_subject}}"}}
`

// TestStallNoCascade: a needs_human run gets one stall check across both
// triggers, and that check going needs_human starts no check of its own
// (#704).
func TestStallNoCascade(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallControllerYAML(t, stallCascadeYAML)
	ctx := t.Context()
	res, err := c.startRun(ctx, scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}}, nil)
	require.NoError(t, err)
	failStep := func(step store.Ticket) {
		t.Helper()
		_, err := store.Close(ctx, c.store, "admin", step.ID, nil, "failed")
		require.NoError(t, err)
		require.True(t, c.escalateRun(ctx, step.ParentID))
	}
	failStep(res.(rpc.RunResult).Steps[0])
	checks := func() []store.Ticket {
		rs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun})
		require.NoError(t, err)
		var out []store.Ticket
		for _, r := range rs {
			if isStallRun(r) {
				out = append(out, r)
			}
		}
		return out
	}
	clk.advance(2*time.Hour + time.Minute)
	c.runTriggers()
	cs := checks()
	require.Len(t, cs, 1, "one check per subject across triggers")

	steps, err := store.Children(ctx, c.store, cs[0].ID)
	require.NoError(t, err)
	failStep(steps[0])
	for range 3 {
		clk.advance(5 * time.Hour)
		c.runTriggers()
	}
	assert.Len(t, checks(), 1, "a needs_human check starts no check, from either trigger")
}

// TestStallOwnInstance (JOB-17): with two pack instances, a needs_human
// run is checked by its own instance's trigger only, and that check
// escalating starts no further check.
func TestStallOwnInstance(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallControllerYAML(t, stallCascadeYAML)
	cfg := *c.config()
	// work belongs to b's instance; trigger a sorts first and would win the
	// shared stall key without the instance check.
	cfg.Instances = []config.PackInstance{{Path: "a", Workflows: []string{"check-a"}}, {Path: "b", Workflows: []string{"check-b", "work"}}}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	ctx := t.Context()
	res, err := c.startRun(ctx, scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}}, nil)
	require.NoError(t, err)
	failStep := func(step store.Ticket) {
		t.Helper()
		_, err := store.Close(ctx, c.store, "admin", step.ID, nil, "failed")
		require.NoError(t, err)
		require.True(t, c.escalateRun(ctx, step.ParentID))
	}
	failStep(res.(rpc.RunResult).Steps[0])
	checks := func() []store.Ticket {
		rs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun})
		require.NoError(t, err)
		var out []store.Ticket
		for _, r := range rs {
			if isStallRun(r) {
				out = append(out, r)
			}
		}
		return out
	}
	for range 3 {
		clk.advance(2*time.Hour + time.Minute)
		c.runTriggers()
	}
	cs := checks()
	require.Len(t, cs, 1)
	assert.Equal(t, "b", metaString(cs[0].Meta, triggerPack, "name"), "the run's own instance checks it")

	steps, err := store.Children(ctx, c.store, cs[0].ID)
	require.NoError(t, err)
	failStep(steps[0])
	for range 3 {
		clk.advance(5 * time.Hour)
		c.runTriggers()
	}
	assert.Len(t, checks(), 1, "the check escalating starts no other check")
}
