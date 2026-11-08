package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Command and http steps (#396): the controller claims a ready step that
// has fields.type command or http as holder autoHolder, runs it in a
// goroutine and closes it. No session, no turn.

// autoHolder holds command and http steps while the controller runs them.
const autoHolder = "controller"

// ExitNotYet is the command exit code for "not yet, retry after
// retry_after" (EX_TEMPFAIL). It does not count as a failure.
const ExitNotYet = 75

// autoOutputMax bounds the stored output of one attempt.
const autoOutputMax = 64 << 10

// Step results.
const (
	resultOK     = "ok"
	resultNotYet = "not_yet"
	resultFailed = "failed"
)

// autoResult is one attempt's outcome.
type autoResult struct {
	result     string
	output     string
	outcome    string
	command    string
	retryAfter time.Duration
	err        string
	// structured is a script step's result(output=...).
	structured map[string]any
}

// autoSteps tracks attempts the controller is running, by step ticket id,
// and how many run per workflow step.
type autoSteps struct {
	running map[string]bool
	per     map[string]int // workflow/step -> running attempts
}

// runAutoSteps starts ready command and http steps, and picks up ones it
// held when it stopped. Called every tick.
func (c *Controller) runAutoSteps(ctx context.Context, cfg *config.Config) {
	// Held by the controller but not running here: a restart cut the
	// attempt short.
	held, err := c.store.List(ctx, store.Filter{Kind: store.KindStep, Holder: autoHolder, Status: store.StatusInProgress})
	if err != nil {
		return
	}
	for _, t := range held {
		c.resumeAutoStep(ctx, cfg, t)
	}
	if c.draining() {
		return
	}
	for proj := range cfg.Projects {
		ready, err := c.store.List(ctx, store.Filter{Project: proj, Ready: true})
		if err != nil {
			continue
		}
		for _, t := range ready {
			if t.Kind != store.KindStep || t.Pool != "" || t.Holder != "" {
				continue
			}
			c.startAutoStep(ctx, cfg, t, false)
		}
	}
}

// resumeAutoStep handles a step the controller holds with no attempt in
// this process: on_restart retry runs it again under the same claim, so
// the attempt key is unchanged; fail releases it as a failed attempt.
func (c *Controller) resumeAutoStep(ctx context.Context, cfg *config.Config, t store.Ticket) {
	c.autoMu.Lock()
	running := c.auto.running[t.ID]
	c.autoMu.Unlock()
	if running {
		return
	}
	s, _, ok := c.autoStepConfig(cfg, t)
	if ok && s.OnRestart == config.OnRestartFail {
		cl := store.Claim{TicketID: t.ID, Holder: autoHolder, Fence: t.Fence}
		if t2, err := c.store.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: "controller restarted during the attempt (on_restart: fail)"}); err == nil {
			c.log.WarnContext(ctx, "step attempt lost on restart", "ticket", t.ID, "step", s.ID)
			c.stepEvent(ctx, cfg, t2)
		}
		return
	}
	c.startAutoStep(ctx, cfg, t, true)
}

// autoStepConfig finds the step's config through its run's workflow.
func (c *Controller) autoStepConfig(cfg *config.Config, t store.Ticket) (config.Step, store.StepFields, bool) {
	sf, err := store.DecodeFields[store.StepFields](t)
	if err != nil || !sf.Auto() {
		return config.Step{}, sf, false
	}
	if sf.Type == store.StepScript {
		return scriptStep(sf), sf, true
	}
	run, err := c.store.Get(c.ctx, sf.Run)
	if err != nil {
		return config.Step{}, sf, false
	}
	rf, err := store.DecodeFields[store.RunFields](run)
	if err != nil {
		return config.Step{}, sf, false
	}
	w, ok := cfg.Workflows[rf.Workflow]
	if !ok {
		return config.Step{}, sf, false
	}
	for _, s := range w.Steps {
		if s.ID == sf.Step && s.Type == sf.Type {
			return s, sf, true
		}
	}
	return config.Step{}, sf, false
}

// startAutoStep claims t (unless held already) and runs one attempt.
func (c *Controller) startAutoStep(ctx context.Context, cfg *config.Config, t store.Ticket, held bool) {
	s, sf, ok := c.autoStepConfig(cfg, t)
	if !ok {
		return
	}
	key := sf.Run + "/" + s.ID
	c.autoMu.Lock()
	if c.auto.running == nil {
		c.auto = autoSteps{running: map[string]bool{}, per: map[string]int{}}
	}
	// The limit is per step id across runs of one workflow.
	limKey := t.Project + "/" + s.ID + "/" + s.Type
	if c.auto.running[t.ID] || c.auto.per[limKey] >= s.ConcurrencyLimit() {
		c.autoMu.Unlock()
		return
	}
	c.auto.running[t.ID] = true
	c.auto.per[limKey]++
	c.autoMu.Unlock()
	done := func() {
		c.autoMu.Lock()
		delete(c.auto.running, t.ID)
		c.auto.per[limKey]--
		c.autoMu.Unlock()
	}
	// The lease covers the attempt plus a margin; the controller renews
	// nothing, so an attempt that outlives it is a lost attempt.
	lease := s.TimeoutDur() + time.Minute
	cl := store.Claim{TicketID: t.ID, Holder: autoHolder, Fence: t.Fence}
	if held {
		var err error
		if cl, err = c.store.Heartbeat(ctx, cl, lease); err != nil {
			done()
			return
		}
	} else {
		var err error
		if t, cl, err = c.store.Claim(ctx, t.ID, autoHolder, lease); err != nil {
			done()
			return
		}
	}
	attemptKey := fmt.Sprintf("%s/%d", key, cl.Fence)
	c.wg.Go(func() {
		defer done()
		// The attempt keeps the config of the pass that started it.
		c.runAutoAttempt(cfg, t, cl, s, sf, attemptKey)
	})
}

// runAutoAttempt runs one attempt and records it.
func (c *Controller) runAutoAttempt(cfg *config.Config, t store.Ticket, cl store.Claim, s config.Step, sf store.StepFields, attemptKey string) {
	ctx := c.ctx
	start := c.now()
	st, err := c.autoStepVars(ctx, cfg, sf)
	var res autoResult
	if err != nil {
		res = autoResult{result: resultFailed, err: err.Error()}
	} else {
		actx, cancel := context.WithTimeout(ctx, s.TimeoutDur())
		switch s.Type {
		case config.StepCommand:
			res = c.runCommandStep(actx, cfg, t, s, sf, st, attemptKey)
		case config.StepHTTP:
			res = c.runHTTPStep(actx, cfg, s, st, attemptKey)
		case store.StepScript:
			res = c.runScriptStep(actx, cfg, t, s, sf, st, attemptKey)
		}
		cancel()
	}
	if ctx.Err() != nil {
		// Shutdown cut the attempt: leave the claim for on_restart.
		return
	}
	if res.result == resultNotYet && res.retryAfter <= 0 {
		res.retryAfter = s.RetryAfterDur()
	}
	attrs := metric.WithAttributes(telemetry.AttrStep.String(s.ID), telemetry.AttrStepType.String(s.Type),
		telemetry.AttrResult.String(res.result))
	c.metrics.StepDuration.Record(ctx, c.now().Sub(start).Seconds(), attrs)
	c.metrics.StepResults.Add(ctx, 1, attrs)
	c.finishAutoStep(ctx, cfg, t, cl, sf, res)
}

// finishAutoStep writes the attempt's record on the step and closes,
// defers or fails it.
func (c *Controller) finishAutoStep(ctx context.Context, cfg *config.Config, t store.Ticket, cl store.Claim, sf store.StepFields, res autoResult) {
	sf.Command, sf.Output, sf.Outcome, sf.Result = res.command, res.output, res.outcome, res.result
	sf = scriptStepFields(sf, res)
	b, err := json.Marshal(sf)
	if err == nil {
		_, err = c.store.Update(ctx, autoHolder, t.ID, store.Update{Fields: b})
	}
	if err != nil {
		c.log.ErrorContext(ctx, "step record", "ticket", t.ID, "err", err)
	}
	note := fmt.Sprintf("%s %s: %s", sf.Type, res.command, res.result)
	if res.err != "" {
		note += " (" + res.err + ")"
	}
	_, _ = store.AddComment(ctx, c.store, autoHolder, t.ID, note)
	switch res.result {
	case resultOK:
		t, err = store.Close(ctx, c.store, autoHolder, t.ID, &cl, "done")
	case resultNotYet:
		t, err = c.store.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: "not yet", Until: c.now().Add(res.retryAfter)})
	default:
		t, err = store.Close(ctx, c.store, autoHolder, t.ID, &cl, "failed")
	}
	if err != nil {
		c.log.ErrorContext(ctx, "step finish", "ticket", t.ID, "result", res.result, "err", err)
		return
	}
	c.log.InfoContext(ctx, "step attempt", "ticket", t.ID, "step", sf.Step, "type", sf.Type, "result", res.result)
	c.stepEvent(ctx, cfg, t)
}

// stepTemplates is what a command or http step's templates are filled
// from: the run's workflow and vars, and earlier steps' outputs.
type stepTemplates struct {
	w     config.Workflow
	vars  map[string]any
	steps map[string]any // id -> {output, outcome, json} for template: go
	out   map[string]store.StepFields
}

// fill renders one template.
func (st stepTemplates) fill(s string) (string, error) {
	return st.w.FillStepTemplate(s, st.vars, st.steps)
}

// autoStepVars loads the run's vars and earlier steps' outputs.
func (c *Controller) autoStepVars(ctx context.Context, cfg *config.Config, sf store.StepFields) (stepTemplates, error) {
	var st stepTemplates
	run, err := c.store.Get(ctx, sf.Run)
	if err != nil {
		return st, err
	}
	rf, err := store.DecodeFields[store.RunFields](run)
	if err != nil {
		return st, err
	}
	st.vars = map[string]any{}
	if w, ok := cfg.Workflows[rf.Workflow]; ok {
		st.w = w
		for k, v := range w.PackVars {
			st.vars[k] = v
		}
	}
	for k, v := range rf.Vars {
		st.vars[k] = v
	}
	steps, err := store.Children(ctx, c.store, sf.Run)
	if err != nil {
		return st, err
	}
	st.out = map[string]store.StepFields{}
	st.steps = map[string]any{}
	for _, s := range steps {
		f, err := store.DecodeFields[store.StepFields](s)
		if err != nil || f.Index >= sf.Index {
			continue
		}
		st.out[f.Step] = f
		m := map[string]any{"output": f.Output, "outcome": f.Outcome}
		var j any
		if json.Unmarshal([]byte(f.Output), &j) == nil {
			m["json"] = j
		}
		st.steps[f.Step] = m
	}
	return st, nil
}

// stepEnv is what a command step gets on top of the controller's env:
// the run vars as JOBSHOP_VAR_<NAME> and the attempt key.
func stepEnv(vars map[string]any, attemptKey string, t store.Ticket) []string {
	var env []string
	for k, v := range vars {
		env = append(env, "JOBSHOP_VAR_"+strings.ToUpper(k)+"="+fmt.Sprint(v))
	}
	return append(env, "JOBSHOP_STEP_KEY="+attemptKey, "JOBSHOP_TICKET="+t.ID)
}

// runDir is a run's scratch dir for its command steps: the sandbox's own
// writable dir and the parent of the run worktree (#419).
func (c *Controller) runDir(runID string) string {
	return filepath.Join(c.root.Dir, "runs", runID)
}

// stepPlace is where and how a command step runs.
type stepPlace struct {
	pool string
	dir  string
	// own is the run dir when the step is sandboxed or has a worktree.
	own string
}

// stepPlace resolves a command step's pool and working dir: the run
// worktree when the step asks for one, else the pool's dir, else its
// repo, else the root. It makes the worktree.
func (c *Controller) stepPlace(cfg *config.Config, s config.Step, sf store.StepFields, st stepTemplates) (stepPlace, error) {
	pl := stepPlace{dir: c.root.Dir}
	if s.Pool == "" {
		return pl, nil
	}
	name, err := cfg.ResolvePool(&st.w, s.Pool, st.vars)
	if err != nil {
		return pl, err
	}
	pl.pool = name
	p := cfg.Pools[name]
	switch {
	case p.Dir != "":
		pl.dir = p.Dir
	case p.Repo != "":
		pl.dir = p.Repo
	}
	if s.Worktree || cfg.SandboxFor(name).Backend != "" {
		pl.own = c.runDir(sf.Run)
		if err := os.MkdirAll(pl.own, 0o700); err != nil {
			return pl, err
		}
	}
	if s.Worktree {
		if p.Repo == "" {
			return pl, fmt.Errorf("worktree: pool %s has no repo", name)
		}
		pl.dir = filepath.Join(pl.own, "work")
		if err := c.addWorktree(p.Repo, p.Ref, pl.dir, "jobshop run "+sf.Run, "pool "+name); err != nil {
			return pl, err
		}
	}
	return pl, nil
}

// removeRunDir deletes a closed run's worktree and scratch dir.
func (c *Controller) removeRunDir(runID string) {
	dir := c.runDir(runID)
	if _, err := os.Stat(dir); err != nil {
		return
	}
	// git worktree remove and a large RemoveAll take a while: outside the
	// tick and the close RPC (JOB-103), one per run dir. A dir left behind
	// is removed by a later sweepRunDirs.
	c.offGo("rundir\x00"+runID, "run_dir", runDirRemoveTimeout, false, func(ctx context.Context) any {
		work := filepath.Join(dir, "work")
		if gd := gitCommonDir(work); gd != "" {
			c.gitMu.Lock()
			// Locked at add; two --force override the lock.
			if out, err := exec.CommandContext(ctx, "git", "--git-dir", gd, "worktree", "remove", "--force", "--force", work).CombinedOutput(); err != nil {
				c.log.WarnContext(ctx, "git worktree remove", "run", runID, "err", err, "out", strings.TrimSpace(string(out)))
			}
			c.gitMu.Unlock()
		}
		if err := os.RemoveAll(dir); err != nil {
			c.log.WarnContext(ctx, "remove run dir", "run", runID, "err", err)
		}
		return nil
	})
}

// runDirRemoveTimeout bounds git worktree remove for one run dir.
const runDirRemoveTimeout = 5 * time.Minute

// sweepRunDirs removes run dirs whose run is closed or gone: a close
// that raced a running attempt, or a restart between close and removal.
func (c *Controller) sweepRunDirs() {
	ents, err := os.ReadDir(filepath.Join(c.root.Dir, "runs"))
	if err != nil {
		return
	}
	for _, e := range ents {
		t, err := c.store.Get(c.ctx, e.Name())
		if errors.Is(err, store.ErrNotFound) || (err == nil && t.Status == store.StatusClosed) {
			c.removeRunDir(e.Name())
		}
	}
}

func (c *Controller) runCommandStep(ctx context.Context, cfg *config.Config, t store.Ticket, s config.Step, sf store.StepFields, st stepTemplates, attemptKey string) autoResult {
	argv, err := s.Argv()
	if err != nil {
		return autoResult{result: resultFailed, err: err.Error()}
	}
	for i, a := range argv {
		if argv[i], err = st.fill(a); err != nil {
			return autoResult{result: resultFailed, err: err.Error()}
		}
	}
	extra := append(stepEnv(st.vars, attemptKey, t), c.slotEnv(ctx, t)...)
	// The controller's own binary comes first on PATH and is JOBSHOP_BIN,
	// as for sessions: units run with a PATH that has no jobshop (#669).
	if exe, err := c.hostExe(); err == nil {
		extra = append(extra, "JOBSHOP_BIN="+exe,
			"PATH="+filepath.Dir(exe)+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	env := append(withoutEnv(os.Environ(), "PATH"), extra...)
	red := cfg.Redact.Scanner().Redactor(env)
	res := autoResult{command: red.Text(strings.Join(argv, " "))}
	pl, err := c.stepPlace(cfg, s, sf, st)
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	ic := isolatedCmd{argv: argv, dir: pl.dir, own: pl.own, env: extra, name: "step", cfg: cfg}
	if c.isolates(cfg, pl.pool) {
		// The pool's sandbox, user and cgroup, as its sessions get (#810).
		ic.pool = pl.pool
	}
	var out, errOut limitedBuffer
	ic.stdout, ic.stderr = &out, &errOut
	ex, runErr := c.runIsolated(ctx, ic)
	res.output = red.Text(out.String())
	res.outcome = outcomeOf(res.output)
	switch {
	case ctx.Err() != nil:
		res.result, res.err = resultFailed, "timed out"
	case runErr == nil && ex.ExitCode == 0 && ex.Signal == "":
		res.result = resultOK
	case runErr == nil && ex.Signal == "" && ex.ExitCode == ExitNotYet:
		res.result = resultNotYet
	default:
		res.result = resultFailed
		if runErr != nil {
			res.err = runErr.Error()
		} else {
			res.err = "exit " + exitText(ex)
		}
		if tail := strings.TrimSpace(red.Text(errOut.String())); tail != "" {
			res.err += ": " + lastLines(tail, 5)
		}
	}
	return res
}

func (c *Controller) runHTTPStep(ctx context.Context, cfg *config.Config, s config.Step, st stepTemplates, attemptKey string) autoResult {
	fill := st.fill
	url, err := fill(s.URL)
	if err != nil {
		return autoResult{result: resultFailed, err: err.Error()}
	}
	res := autoResult{command: s.Method + " " + url}
	body, err := fill(s.Body)
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	req, err := http.NewRequestWithContext(telemetry.WithRoute(ctx, "/step"), s.Method, url, strings.NewReader(body))
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		return res
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range s.Headers {
		if v, err = fill(v); err != nil {
			res.result, res.err = resultFailed, err.Error()
			return res
		}
		req.Header.Set(k, v)
	}
	var secrets []string
	for k, f := range s.HeaderFiles {
		b, err := os.ReadFile(f)
		if err != nil {
			res.result, res.err = resultFailed, fmt.Sprintf("header %s: %v", k, err)
			return res
		}
		v := strings.TrimSpace(string(b))
		req.Header.Set(k, v)
		secrets = append(secrets, "HEADER_SECRET="+v)
	}
	req.Header.Set("Idempotency-Key", attemptKey)
	resp, err := stepHTTP.Do(req)
	if err != nil {
		res.result, res.err = resultFailed, err.Error()
		if ctx.Err() != nil {
			res.err = "timed out"
		}
		return res
	}
	defer func() { _ = resp.Body.Close() }()
	var out limitedBuffer
	_, _ = io.Copy(&out, resp.Body)
	res.output = cfg.Redact.Scanner().Redactor(secrets).Text(out.String())
	res.outcome = outcomeOf(res.output)
	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:
		res.result = resultOK
	case code == http.StatusConflict || code == http.StatusTooEarly || code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable:
		res.result = resultNotYet
		res.retryAfter = hooks.RetryAfter(resp.Header.Get("Retry-After"), c.now())
	default:
		res.result, res.err = resultFailed, resp.Status
	}
	return res
}

// outcomeOf is the outcome word an attempt set: a top-level JSON
// "outcome" string, or the last stdout line "outcome: <word>". For
// conditional steps (#385).
func outcomeOf(out string) string {
	var m map[string]any
	if json.Unmarshal([]byte(out), &m) == nil {
		if s, ok := m["outcome"].(string); ok {
			return strings.TrimSpace(s)
		}
		return ""
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if w, ok := strings.CutPrefix(lines[len(lines)-1], "outcome:"); ok {
		return strings.TrimSpace(w)
	}
	return ""
}

func lastLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}

// limitedBuffer keeps the first autoOutputMax bytes and drops the rest,
// so a chatty command cannot grow the step ticket.
type limitedBuffer struct {
	b         bytes.Buffer
	truncated bool
}

func (l *limitedBuffer) Write(p []byte) (int, error) {
	if room := autoOutputMax - l.b.Len(); room < len(p) {
		l.truncated = true
		if room > 0 {
			l.b.Write(p[:room])
		}
		return len(p), nil
	}
	return l.b.Write(p)
}

func (l *limitedBuffer) String() string {
	if l.truncated {
		return l.b.String() + "\n[truncated]"
	}
	return l.b.String()
}
