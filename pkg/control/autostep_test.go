package control_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// autoServer is an http step target that records requests.
type autoServer struct {
	mu     sync.Mutex
	bodies []string
	keys   []string
	// block, when set, holds requests until the request context ends.
	block bool
	// seen is signalled once per request.
	seen chan struct{}
}

func (s *autoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	s.mu.Lock()
	s.bodies = append(s.bodies, string(b))
	s.keys = append(s.keys, r.Header.Get("Idempotency-Key"))
	block := s.block
	s.mu.Unlock()
	select {
	case s.seen <- struct{}{}:
	default:
	}
	if block {
		<-r.Context().Done()
		return
	}
	if strings.Contains(string(b), `"later":true`) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	_, _ = fmt.Fprint(w, `{"queued":true,"outcome":"queued"}`)
}

func (s *autoServer) snapshot() ([]string, []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.bodies...), append([]string(nil), s.keys...)
}

// autoRoot writes a root whose pack has workflows mixing agent, command
// and http steps, pointing http steps at url.
func autoRoot(t *testing.T, url string) config.Root {
	t.Helper()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName), []byte(`workflows:
  mixed:
    vars: [issue]
    steps:
      # Not yet on the first attempt (exit 75), then JSON output.
      - id: probe
        type: command
        run: 'if [ ! -f probed ]; then touch probed; exit 75; fi; echo "{\"sha\":\"abc{{.issue}}\",\"outcome\":\"green\"}"'
        retry_after: 1m
      - id: notify
        type: http
        url: `+url+`/notify
        body: '{"sha":"{{.steps.probe.json.sha}}","issue":"{{.issue}}","outcome":"{{.steps.probe.outcome}}"}'
      - id: review
        pool: dev
        title: "Review {{.issue}}"
        prompt: "Review {{.issue}}"
  later:
    vars: [issue]
    steps:
      - id: ask
        type: http
        url: `+url+`/later
        body: '{"later":true}'
  failing:
    vars: [issue]
    escalate_pool: dev
    steps:
      - id: check
        type: command
        retries: 2
        run: [sh, -c, 'echo broken >&2; exit 3']
  slow-retry:
    vars: [issue]
    steps:
      - id: cmd
        type: command
        run: 'echo "$JOBSHOP_STEP_KEY" >> keys; if [ -f go ]; then exit 0; fi; exec sleep 60'
  slow-fail:
    vars: [issue]
    steps:
      - id: cmd
        type: command
        on_restart: fail
        run: 'echo "$JOBSHOP_STEP_KEY" >> keys; if [ -f go ]; then exit 0; fi; exec sleep 60'
  slowhttp:
    vars: [issue]
    steps:
      - id: call
        type: http
        url: `+url+`/slow
`), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "gopk"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gopk", config.PackFileName), []byte(`template: go
workflows:
  gomixed:
    vars: [issue]
    steps:
      - id: probe
        type: command
        run: [sh, -c, 'echo "{\"sha\":\"go{{.issue}}\"}"']
      - id: notify
        type: http
        url: `+url+`/notify
        body: '{"sha":"{{.steps.probe.json.sha}}","issue":"{{.issue}}"}'
`), 0o600))
	cfg := fmt.Sprintf(`packs:
  - path: pk
  - path: gopk
projects:
  js: {}
agents:
  fake:
    command: %q
pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}
`, os.Args[0], dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	return root
}

// startAuto runs a controller on root with a fake clock. stop cancels it
// and waits for it to exit.
func startAuto(t *testing.T, root config.Root, clk clock.Clock) (admin *rpc.Client, stop func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- control.Run(ctx, root, withPass(root.Dir, control.Options{Tick: time.Hour, Clock: clk, Log: testLog(t)}))
	}()
	var once sync.Once
	stop = func() {
		once.Do(func() {
			cancel()
			if err := <-done; !errors.Is(err, context.Canceled) {
				require.NoError(t, err)
			}
		})
	}
	t.Cleanup(stop)
	return dial(t, root.AdminSocket()), stop
}

// passUntilWall runs passes until cond holds. Attempts run in goroutines
// outside the pass, so this waits on the condition, not a pass count.
func passUntilWall(t *testing.T, root config.Root, cond func() bool, msg string) {
	t.Helper()
	require.Eventually(t, func() bool {
		if cond() {
			return true
		}
		pass(t, root.Dir, 1)
		return cond()
	}, 30*time.Second, 10*time.Millisecond, msg)
}

func stepFields(t *testing.T, s domain.Ticket) domain.StepFields {
	t.Helper()
	f, err := domain.DecodeFields[domain.StepFields](s)
	require.NoError(t, err)
	return f
}

// TestAutoStepsMixedRun runs command, http and agent steps in one run: a
// not-yet result defers without a failure, output feeds the next step,
// and the agent step waits for a session.
func TestAutoStepsMixedRun(t *testing.T) {
	t.Parallel()
	srv := &autoServer{seen: make(chan struct{}, 8)}
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	root := autoRoot(t, ts.URL)
	clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	admin, _ := startAuto(t, root, clk)

	r := startRun(t, admin, rpc.RunStartParams{Workflow: "mixed", Vars: map[string]string{"issue": "7"}})
	require.Len(t, r.Steps, 3)
	assert.Empty(t, r.Steps[0].Pool, "command steps have no pool")
	assert.Equal(t, "command probe", r.Steps[0].Title)
	probe := r.Steps[0].ID

	passUntilWall(t, root, func() bool {
		s := showRun(t, admin, r.Run.ID).Steps[0]
		return stepFields(t, s).Result == "not_yet" && s.Status == domain.StatusOpen
	}, "probe deferred")
	s := showRun(t, admin, r.Run.ID).Steps[0]
	assert.Equal(t, 0, stepFields(t, s).Failures, "not yet is not a failure")
	assert.Equal(t, clk.Now().Add(time.Minute), s.DeferUntil.UTC())
	pass(t, root.Dir, 2)
	assert.Equal(t, domain.StatusOpen, showRun(t, admin, r.Run.ID).Steps[0].Status, "deferred until retry_after")

	clk.Advance(time.Minute)
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Steps[1].Status == domain.StatusClosed
	}, "probe and notify done")
	res := showRun(t, admin, r.Run.ID)
	pf := stepFields(t, res.Steps[0])
	assert.Equal(t, "done", res.Steps[0].Resolution)
	assert.Equal(t, "green", pf.Outcome)
	assert.JSONEq(t, `{"sha":"abc7","outcome":"green"}`, pf.Output)
	nf := stepFields(t, res.Steps[1])
	assert.Equal(t, "ok", nf.Result)
	assert.Equal(t, "POST "+ts.URL+"/notify", nf.Command)
	assert.Equal(t, "queued", nf.Outcome)
	bodies, keys := srv.snapshot()
	require.Len(t, bodies, 1)
	assert.JSONEq(t, `{"sha":"abc7","issue":"7","outcome":"green"}`, bodies[0], "probe output fed the request")
	assert.Equal(t, r.Run.ID+"/notify/1", keys[0])

	// The agent step is left for a session.
	assert.Equal(t, []string{res.Steps[2].ID}, readyIDs(t, admin))
	assert.Equal(t, "dev", res.Steps[2].Pool)
	assert.NotEqual(t, probe, res.Steps[2].ID)
	closeStep(t, admin, res.Steps[2].ID, "done")
	passUntilWall(t, root, func() bool { return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunDone }, "run done")

	var orphans []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Orphaned: true}, &orphans))
	assert.Empty(t, orphans, "command and http steps are not orphans")
}

// TestAutoStepGoTemplate passes output between steps in a template: go
// pack through the steps var.
func TestAutoStepGoTemplate(t *testing.T) {
	t.Parallel()
	srv := &autoServer{seen: make(chan struct{}, 8)}
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	root := autoRoot(t, ts.URL)
	admin, _ := startAuto(t, root, clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)))
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "gomixed", Vars: map[string]string{"issue": "9"}})
	passUntilWall(t, root, func() bool { return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunDone }, "run done")
	bodies, _ := srv.snapshot()
	require.Len(t, bodies, 1)
	assert.JSONEq(t, `{"sha":"go9","issue":"9"}`, bodies[0])
}

// TestAutoStepHTTPRetryAfter defers on 503 by the Retry-After header.
func TestAutoStepHTTPRetryAfter(t *testing.T) {
	t.Parallel()
	srv := &autoServer{seen: make(chan struct{}, 8)}
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	root := autoRoot(t, ts.URL)
	clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	admin, _ := startAuto(t, root, clk)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "later", Vars: map[string]string{"issue": "1"}})
	passUntilWall(t, root, func() bool {
		// The result is recorded before the release sets defer_until:
		// wait for both writes.
		s := showRun(t, admin, r.Run.ID).Steps[0]
		return stepFields(t, s).Result == "not_yet" && !s.DeferUntil.IsZero()
	}, "deferred")
	s := showRun(t, admin, r.Run.ID).Steps[0]
	assert.Equal(t, clk.Now().Add(30*time.Second), s.DeferUntil.UTC())
	assert.Equal(t, 0, stepFields(t, s).Failures)
}

// TestAutoStepFailureEscalates counts a nonzero exit against retries and
// escalates the run at the limit.
func TestAutoStepFailureEscalates(t *testing.T) {
	t.Parallel()
	root := autoRoot(t, "http://127.0.0.1:1")
	admin, _ := startAuto(t, root, clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)))
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "failing", Vars: map[string]string{"issue": "1"}})
	devTickets := func() []domain.Ticket {
		var esc []domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Pool: "dev"}, &esc))
		return esc
	}
	// The run goes needs-human before the escalation ticket is written:
	// wait for both.
	passUntilWall(t, root, func() bool {
		return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunNeedsHuman && len(devTickets()) > 0
	}, "needs human with escalation")
	s := showRun(t, admin, r.Run.ID).Steps[0]
	assert.Equal(t, "failed", s.Resolution)
	assert.Equal(t, 2, stepFields(t, s).Failures)
	var sr rpc.ShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: s.ID}, &sr))
	var notes []string
	for _, c := range sr.Comments {
		notes = append(notes, c.Body)
	}
	assert.Contains(t, strings.Join(notes, "\n"), "broken", "stderr tail is on the step")
	assert.Len(t, devTickets(), 1, "one escalation ticket")
}

// keyLines reads the attempt keys a slow command step appended.
func keyLines(t *testing.T, root config.Root) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root.Dir, "keys"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	require.NoError(t, err)
	return strings.Fields(string(b))
}

// TestAutoStepRestart stops the controller during a command step: with
// on_restart retry the next controller runs it again with the same
// attempt key and no failure; with fail it counts one failure.
func TestAutoStepRestart(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{config.OnRestartRetry, config.OnRestartFail} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			root := autoRoot(t, "http://127.0.0.1:1")
			clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
			admin, stop := startAuto(t, root, clk)
			r := startRun(t, admin, rpc.RunStartParams{Workflow: "slow-" + mode, Vars: map[string]string{"issue": "1"}})
			passUntilWall(t, root, func() bool { return len(keyLines(t, root)) == 1 }, "first attempt started")
			stop()

			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
			admin, _ = startAuto(t, root, clk)
			passUntilWall(t, root, func() bool { return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunDone }, "run done")
			keys := keyLines(t, root)
			require.Len(t, keys, 2, "one attempt per controller, none duplicated")
			f := stepFields(t, showRun(t, admin, r.Run.ID).Steps[0])
			if mode == config.OnRestartRetry {
				assert.Equal(t, keys[0], keys[1], "same attempt key after restart")
				assert.Equal(t, 0, f.Failures)
			} else {
				assert.NotEqual(t, keys[0], keys[1], "a new attempt")
				assert.Equal(t, 1, f.Failures)
			}
		})
	}
}

// TestAutoStepHTTPRestart resends an http step cut by a restart with the
// same Idempotency-Key.
func TestAutoStepHTTPRestart(t *testing.T) {
	t.Parallel()
	srv := &autoServer{seen: make(chan struct{}, 8), block: true}
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	root := autoRoot(t, ts.URL)
	clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	admin, stop := startAuto(t, root, clk)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "slowhttp", Vars: map[string]string{"issue": "1"}})
	passUntilWall(t, root, func() bool { _, k := srv.snapshot(); return len(k) == 1 }, "first request")
	stop()

	srv.mu.Lock()
	srv.block = false
	srv.mu.Unlock()
	admin, _ = startAuto(t, root, clk)
	passUntilWall(t, root, func() bool { return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunDone }, "run done")
	_, keys := srv.snapshot()
	require.Len(t, keys, 2)
	assert.Equal(t, keys[0], keys[1], "same Idempotency-Key after restart")
	assert.Equal(t, r.Run.ID+"/call/1", keys[0])
}

// worktreeRoot writes a root whose pool dev has a git repo, and a
// workflow of two command steps sharing the run worktree. sandbox adds
// the bubblewrap sandbox to the pool.
func worktreeRoot(t *testing.T, sandbox bool) (config.Root, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	repo := filepath.Join(dir, "repo")
	require.NoError(t, os.MkdirAll(repo, 0o750))
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName), []byte(`workflows:
  wt:
    vars: [issue]
    steps:
      - id: write
        type: command
        pool: dev
        worktree: true
        run: 'echo "$JOBSHOP_VAR_ISSUE" > note; pwd'
      - id: read
        type: command
        pool: dev
        worktree: true
        run: 'cat note; test ! -e ../../../jobshop.db || echo db-visible'
`), 0o600))
	sb := ""
	if sandbox {
		sb = ", sandbox: {backend: bubblewrap, ro: " + testSystemRO + "}"
	}
	cfg := fmt.Sprintf(`packs:
  - path: pk
projects:
  js: {}
agents:
  fake:
    command: %q
pools:
  dev: {project: js, agent: fake, max_sessions: 0, repo: %q%s}
`, os.Args[0], repo, sb)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	return root, repo
}

// TestCommandStepWorktree runs two command steps in one run worktree
// (#419): the second sees the first's file, and the worktree goes when
// the run closes.
func TestCommandStepWorktree(t *testing.T) {
	t.Parallel()
	testCommandStepWorktree(t, false)
}

// TestCommandStepSandbox is TestCommandStepWorktree under the pool's
// bubblewrap sandbox: the root's jobshop.db is not visible.
func TestCommandStepSandbox(t *testing.T) {
	bwrapWorks(t)
	t.Parallel()
	testCommandStepWorktree(t, true)
}

func testCommandStepWorktree(t *testing.T, sandbox bool) {
	t.Helper()
	root, repo := worktreeRoot(t, sandbox)
	clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	admin, _ := startAuto(t, root, clk)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "wt", Vars: map[string]string{"issue": "9"}})
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Steps[1].Status == domain.StatusClosed
	}, "steps done")
	res := showRun(t, admin, r.Run.ID)
	work := filepath.Join(root.Dir, "runs", r.Run.ID, "work")
	w, rd := stepFields(t, res.Steps[0]), stepFields(t, res.Steps[1])
	require.Equal(t, "ok", w.Result, w.Output)
	assert.Equal(t, work, strings.TrimSpace(w.Output))
	require.Equal(t, "ok", rd.Result, rd.Output)
	if sandbox {
		assert.Equal(t, "9", strings.TrimSpace(rd.Output), "jobshop.db hidden")
	} else {
		assert.Equal(t, "9\ndb-visible", strings.TrimSpace(rd.Output))
	}
	_, err := os.Stat(filepath.Join(repo, "note"))
	assert.True(t, os.IsNotExist(err), "main checkout untouched")

	// State is written before the close that removes the dir.
	passUntilWall(t, root, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "runs", r.Run.ID))
		return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunDone && os.IsNotExist(err)
	}, "run done and run dir removed")
	out, err := exec.Command("git", "-C", repo, "worktree", "list").Output()
	require.NoError(t, err)
	assert.NotContains(t, string(out), work)
}
