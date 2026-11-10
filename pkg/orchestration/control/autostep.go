package control

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/elee1766/acphost/pkg/proc"
	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/stepexec"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
)

// Command and http steps (#396): the controller claims a ready step that
// has fields.type command or http as holder autoHolder, runs it in a
// goroutine and closes it. No session, no turn. This file keeps the
// claim, lease, record and close; the attempt itself is in stepexec
// (JOB-157).

// autoHolder holds command and http steps while the controller runs them.
const autoHolder = "controller"

// ExitNotYet is the command exit code for "not yet, retry after
// retry_after" (EX_TEMPFAIL). It does not count as a failure.
const ExitNotYet = stepexec.ExitNotYet

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
	held, err := c.store.List(ctx, store.Filter{Kind: domain.KindStep, Holder: autoHolder, Status: domain.StatusInProgress})
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
			if t.Kind != domain.KindStep || t.Pool != "" || t.Holder != "" {
				continue
			}
			c.startAutoStep(ctx, cfg, t, false)
		}
	}
}

// resumeAutoStep handles a step the controller holds with no attempt in
// this process: on_restart retry runs it again under the same claim, so
// the attempt key is unchanged; fail releases it as a failed attempt.
func (c *Controller) resumeAutoStep(ctx context.Context, cfg *config.Config, t domain.Ticket) {
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

// autoStepConfig finds the step's config: in its fields for a script or
// callback command step, else through its run's workflow.
func (c *Controller) autoStepConfig(cfg *config.Config, t domain.Ticket) (config.Step, domain.StepFields, bool) {
	sf, err := domain.DecodeFields[domain.StepFields](t)
	if err != nil || !sf.Auto() {
		return config.Step{}, sf, false
	}
	if sf.Type == domain.StepScript {
		return stepexec.ScriptConfig(sf), sf, true
	}
	if s, ok := stepexec.CommandConfig(sf); ok {
		return s, sf, true
	}
	run, err := c.store.Get(c.ctx, sf.Run)
	if err != nil {
		return config.Step{}, sf, false
	}
	rf, err := domain.DecodeFields[domain.RunFields](run)
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
func (c *Controller) startAutoStep(ctx context.Context, cfg *config.Config, t domain.Ticket, held bool) {
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
func (c *Controller) runAutoAttempt(cfg *config.Config, t domain.Ticket, cl store.Claim, s config.Step, sf domain.StepFields, attemptKey string) {
	ctx := c.ctx
	start := c.now()
	a := stepexec.Attempt{Cfg: cfg, Ticket: t, Step: s, Fields: sf, Key: attemptKey, Store: c.store}
	if s.Type == domain.StepScript {
		a.Engine = c.hooks()
		a.Inputs = func(ctx context.Context) (stepexec.ScriptInputs, error) { return c.scriptInputs(ctx, t, sf) }
	}
	res := c.stepExecutor().Run(ctx, a)
	if ctx.Err() != nil {
		// Shutdown cut the attempt: leave the claim for on_restart.
		return
	}
	if res.Result == stepexec.ResultNotYet && res.RetryAfter <= 0 {
		res.RetryAfter = s.RetryAfterDur()
	}
	attrs := metric.WithAttributes(telemetry.AttrStep.String(s.ID), telemetry.AttrStepType.String(s.Type),
		telemetry.AttrResult.String(res.Result))
	c.metrics.StepDuration.Record(ctx, c.now().Sub(start).Seconds(), attrs)
	c.metrics.StepResults.Add(ctx, 1, attrs)
	c.finishAutoStep(ctx, cfg, t, cl, sf, res)
}

// finishAutoStep writes the attempt's record on the step and closes,
// defers or fails it.
func (c *Controller) finishAutoStep(ctx context.Context, cfg *config.Config, t domain.Ticket, cl store.Claim, sf domain.StepFields, res stepexec.Result) {
	sf.Command, sf.Output, sf.Outcome, sf.Result = res.Command, res.Output, res.Outcome, res.Result
	sf = stepexec.ApplyScript(sf, res)
	b, err := json.Marshal(sf)
	if err == nil {
		_, err = c.store.Update(ctx, autoHolder, t.ID, store.Update{Fields: b})
	}
	if err != nil {
		c.log.ErrorContext(ctx, "step record", "ticket", t.ID, "err", err)
	}
	note := fmt.Sprintf("%s %s: %s", sf.Type, res.Command, res.Result)
	if res.Err != "" {
		note += " (" + res.Err + ")"
	}
	_, _ = store.AddComment(ctx, c.store, autoHolder, t.ID, note)
	switch res.Result {
	case stepexec.ResultOK:
		t, err = store.Close(ctx, c.store, autoHolder, t.ID, &cl, "done")
	case stepexec.ResultNotYet:
		t, err = c.store.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: "not yet", Until: c.now().Add(res.RetryAfter)})
	default:
		t, err = store.Close(ctx, c.store, autoHolder, t.ID, &cl, "failed")
	}
	if err != nil {
		c.log.ErrorContext(ctx, "step finish", "ticket", t.ID, "result", res.Result, "err", err)
		return
	}
	c.log.InfoContext(ctx, "step attempt", "ticket", t.ID, "step", sf.Step, "type", sf.Type, "result", res.Result)
	c.stepEvent(ctx, cfg, t)
}

// stepHost is what stepexec needs from the controller.
type stepHost struct{ c *Controller }

// Run runs a step command isolated as its pool says (#810): the pool's
// sandbox, user and cgroup, as its sessions get.
func (h stepHost) Run(ctx context.Context, cmd stepexec.Cmd) (proc.Exit, error) {
	ic := isolatedCmd{argv: cmd.Argv, dir: cmd.Dir, own: cmd.Own, env: cmd.Env, name: "step", cfg: cmd.Cfg, stdout: cmd.Stdout, stderr: cmd.Stderr}
	if h.c.isolates(cmd.Cfg, cmd.Pool) {
		ic.pool = cmd.Pool
	}
	return h.c.runIsolated(ctx, ic)
}

func (h stepHost) Worktree(repo, ref, dir, reason, what string) error {
	return h.c.addWorktree(repo, ref, dir, reason, what)
}

func (h stepHost) SlotEnv(ctx context.Context, t domain.Ticket) []string {
	return h.c.slotEnv(ctx, t)
}

func (h stepHost) HostEnv() []string {
	exe, err := h.c.hostExe()
	if err != nil {
		return nil
	}
	return []string{
		"JOBSHOP_BIN=" + exe,
		"PATH=" + filepath.Dir(exe) + string(os.PathListSeparator) + os.Getenv("PATH"),
	}
}

// stepExecutor is the executor for one attempt.
func (c *Controller) stepExecutor() stepexec.Executor {
	return stepexec.Executor{Root: c.root.Dir, Host: stepHost{c}, HTTP: stepHTTP, Now: c.now}
}

// removeRunDir deletes a closed run's worktree and scratch dir.
func (c *Controller) removeRunDir(runID string) {
	if _, err := os.Stat(stepexec.RunDir(c.root.Dir, runID)); err != nil {
		return
	}
	// git worktree remove and a large RemoveAll take a while: outside the
	// tick and the close RPC (JOB-103), one per run dir. A dir left behind
	// is removed by a later sweepRunDirs.
	c.offGo("rundir\x00"+runID, "run_dir", runDirRemoveTimeout, false, func(ctx context.Context) any {
		stepexec.RemoveRunDir(ctx, c.log, &c.gitMu, c.root.Dir, runID)
		return nil
	})
}

// runDirRemoveTimeout bounds git worktree remove for one run dir.
const runDirRemoveTimeout = 5 * time.Minute

// sweepRunDirs removes run dirs whose run is closed or gone: a close
// that raced a running attempt, or a restart between close and removal.
func (c *Controller) sweepRunDirs() {
	for _, id := range stepexec.StaleRunDirs(c.ctx, c.store, c.root.Dir) {
		c.removeRunDir(id)
	}
}
