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
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
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
}

// autoSteps tracks attempts the controller is running, by step ticket id,
// and how many run per workflow step.
type autoSteps struct {
	running map[string]bool
	per     map[string]int // workflow/step -> running attempts
}

// runAutoSteps starts ready command and http steps, and picks up ones it
// held when it stopped. Called every tick.
func (c *Controller) runAutoSteps() {
	ctx := c.ctx
	cfg := c.config()
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
		ready, err := c.store.Ready(ctx, proj, "", 0)
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
		if t2, err := c.store.Release(ctx, cl, "controller restarted during the attempt (on_restart: fail)", time.Time{}); err == nil {
			c.log.Warn("step attempt lost on restart", "ticket", t.ID, "step", s.ID)
			c.stepEvent(ctx, t2)
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
		c.runAutoAttempt(t, cl, s, sf, attemptKey)
	})
}

// runAutoAttempt runs one attempt and records it.
func (c *Controller) runAutoAttempt(t store.Ticket, cl store.Claim, s config.Step, sf store.StepFields, attemptKey string) {
	ctx := c.ctx
	start := c.now()
	st, err := c.autoStepVars(ctx, sf)
	var res autoResult
	if err != nil {
		res = autoResult{result: resultFailed, err: err.Error()}
	} else {
		actx, cancel := context.WithTimeout(ctx, s.TimeoutDur())
		switch s.Type {
		case config.StepCommand:
			res = c.runCommandStep(actx, t, s, st, attemptKey)
		case config.StepHTTP:
			res = c.runHTTPStep(actx, s, st, attemptKey)
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
	c.finishAutoStep(ctx, t, cl, sf, res)
}

// finishAutoStep writes the attempt's record on the step and closes,
// defers or fails it.
func (c *Controller) finishAutoStep(ctx context.Context, t store.Ticket, cl store.Claim, sf store.StepFields, res autoResult) {
	sf.Command, sf.Output, sf.Outcome, sf.Result = res.command, res.output, res.outcome, res.result
	b, err := json.Marshal(sf)
	if err == nil {
		_, err = c.store.Update(ctx, autoHolder, t.ID, store.Update{Fields: b})
	}
	if err != nil {
		c.log.Error("step record", "ticket", t.ID, "err", err)
	}
	note := fmt.Sprintf("%s %s: %s", sf.Type, res.command, res.result)
	if res.err != "" {
		note += " (" + res.err + ")"
	}
	_, _ = c.store.AddComment(ctx, autoHolder, t.ID, note)
	switch res.result {
	case resultOK:
		t, err = c.store.CloseTicket(ctx, autoHolder, t.ID, &cl, "done")
	case resultNotYet:
		t, err = c.store.Release(ctx, cl, "not yet", c.now().Add(res.retryAfter))
	default:
		t, err = c.store.CloseTicket(ctx, autoHolder, t.ID, &cl, "failed")
	}
	if err != nil {
		c.log.Error("step finish", "ticket", t.ID, "result", res.result, "err", err)
		return
	}
	c.log.Info("step attempt", "ticket", t.ID, "step", sf.Step, "type", sf.Type, "result", res.result)
	c.stepEvent(ctx, t)
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
	return st.w.FillStepTemplate(s, st.vars, st.steps, st.look)
}

// look resolves ${steps.<id>.<field>}.
func (st stepTemplates) look(id, field string) (string, bool) {
	f, ok := st.out[id]
	if !ok {
		return "", false
	}
	switch field {
	case "output":
		return f.Output, true
	case "outcome":
		return f.Outcome, true
	}
	path, _ := strings.CutPrefix(field, "json.")
	return jsonPath(f.Output, path)
}

// autoStepVars loads the run's vars and earlier steps' outputs.
func (c *Controller) autoStepVars(ctx context.Context, sf store.StepFields) (stepTemplates, error) {
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
	if w, ok := c.config().Workflows[rf.Workflow]; ok {
		st.w = w
		for k, v := range w.PackVars {
			st.vars[k] = v
		}
	}
	for k, v := range rf.Vars {
		st.vars[k] = v
	}
	steps, err := c.store.ListChildren(ctx, sf.Run)
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

// jsonPath reads a dotted path (a.b.0.c) from JSON text. Strings come back
// bare, other values as JSON.
func jsonPath(text, path string) (string, bool) {
	var v any
	if json.Unmarshal([]byte(text), &v) != nil {
		return "", false
	}
	for p := range strings.SplitSeq(path, ".") {
		switch x := v.(type) {
		case map[string]any:
			n, ok := x[p]
			if !ok {
				return "", false
			}
			v = n
		case []any:
			i, err := strconv.Atoi(p)
			if err != nil || i < 0 || i >= len(x) {
				return "", false
			}
			v = x[i]
		default:
			return "", false
		}
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	b, err := json.Marshal(v)
	return string(b), err == nil
}

// stepEnv is the env of a command step: the controller's, plus the run
// vars as JOBSHOP_VAR_<NAME> and the attempt key.
func stepEnv(vars map[string]any, attemptKey string, t store.Ticket) []string {
	env := os.Environ()
	for k, v := range vars {
		env = append(env, "JOBSHOP_VAR_"+strings.ToUpper(k)+"="+fmt.Sprint(v))
	}
	return append(env, "JOBSHOP_STEP_KEY="+attemptKey, "JOBSHOP_TICKET="+t.ID)
}

// stepDir is where a command step runs: its pool's dir, else its repo,
// else the root.
func (c *Controller) stepDir(s config.Step, st stepTemplates) string {
	cfg := c.config()
	if s.Pool != "" {
		if name, err := cfg.ResolvePool(&st.w, s.Pool, st.vars); err == nil {
			p := cfg.Pools[name]
			if p.Dir != "" {
				return p.Dir
			}
			if p.Repo != "" {
				return p.Repo
			}
		}
	}
	return c.root.Dir
}

func (c *Controller) runCommandStep(ctx context.Context, t store.Ticket, s config.Step, st stepTemplates, attemptKey string) autoResult {
	argv, err := s.Argv()
	if err != nil {
		return autoResult{result: resultFailed, err: err.Error()}
	}
	for i, a := range argv {
		if argv[i], err = st.fill(a); err != nil {
			return autoResult{result: resultFailed, err: err.Error()}
		}
	}
	env := stepEnv(st.vars, attemptKey, t)
	res := autoResult{command: telemetry.RedactText(strings.Join(argv, " "), env)}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = c.stepDir(s, st)
	cmd.Env = env
	var out, errOut limitedBuffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	cmd.WaitDelay = 5 * time.Second
	runErr := cmd.Run()
	res.output = telemetry.RedactText(out.String(), env)
	res.outcome = outcomeOf(res.output)
	var exit *exec.ExitError
	switch {
	case runErr == nil:
		res.result = resultOK
	case ctx.Err() != nil:
		res.result, res.err = resultFailed, "timed out"
	case errors.As(runErr, &exit) && exit.ExitCode() == ExitNotYet:
		res.result = resultNotYet
	default:
		res.result = resultFailed
		res.err = runErr.Error()
		if tail := strings.TrimSpace(telemetry.RedactText(errOut.String(), env)); tail != "" {
			res.err += ": " + lastLines(tail, 5)
		}
	}
	return res
}

func (c *Controller) runHTTPStep(ctx context.Context, s config.Step, st stepTemplates, attemptKey string) autoResult {
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
	req, err := http.NewRequestWithContext(ctx, s.Method, url, strings.NewReader(body))
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
	resp, err := http.DefaultClient.Do(req)
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
	res.output = telemetry.RedactText(out.String(), secrets)
	res.outcome = outcomeOf(res.output)
	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:
		res.result = resultOK
	case code == http.StatusConflict || code == http.StatusTooEarly || code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable:
		res.result = resultNotYet
		res.retryAfter = retryAfter(resp.Header.Get("Retry-After"), c.now())
	default:
		res.result, res.err = resultFailed, resp.Status
	}
	return res
}

// retryAfter parses a Retry-After header: seconds or an HTTP date.
func retryAfter(h string, now time.Time) time.Duration {
	if h == "" {
		return 0
	}
	if n, err := strconv.Atoi(strings.TrimSpace(h)); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
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
