package control

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Trigger checks run outside the tick (JOB-109). A check is a shell
// command with a timeout of minutes; run inside the pass it held tickMu,
// so one slow check delayed supervise, the sweeps and every reload.
//
// A pass evaluates a trigger as before, but its ctx carries a checkPass:
// where evaluation would run a check, runCheck records the command and
// returns errCheckPending, so nothing fires. The commands then run in a
// goroutine, outside tickMu, at most maxTriggerChecks at a time. When
// they are done the goroutine takes the tick lock and evaluates the
// trigger again with the results in the checkPass: the conditions before
// the check are read again after the wait, so a trigger that stopped
// holding does not fire, and the check still runs only when every other
// condition holds.

// errCheckPending is what runCheck returns while a pass only collects
// the checks a trigger needs. It is not a failure: callers return quietly.
var errCheckPending = errors.New("trigger check has not run yet")

// maxTriggerChecks bounds the check commands running at once.
const maxTriggerChecks = 8

// checkResult is what one check run found: runCheck's three values.
type checkResult struct {
	out, why string
	err      error
}

// checkNeed is a check an evaluation reached.
type checkNeed struct {
	key  string
	when config.TriggerWhen
}

// checkPass is on the ctx of an evaluation under runTriggerChecks. results
// are the finished checks; need collects the ones not found there, in the
// order they were asked.
type checkPass struct {
	results map[string]checkResult
	need    []checkNeed
}

type checkPassKey struct{}

func withCheckPass(ctx context.Context, cp *checkPass) context.Context {
	return context.WithValue(ctx, checkPassKey{}, cp)
}

// checkKey names a check run: the same command with the same timeout is
// one run, whichever signal or trigger asked.
func checkKey(w config.TriggerWhen) string { return w.CheckTimeout + "\x00" + w.Check }

// checkJob is a trigger evaluation that waits for its checks. run
// evaluates the trigger again on the config of the pass it replays in.
type checkJob struct {
	need []checkNeed
	run  func(ctx context.Context, cfg *config.Config)
}

// checkJobs is the background check state of each trigger by name, under
// tickMu: at most one check job per trigger runs at a time.
type checkJobs struct {
	running map[string]bool
	queued  map[string]checkJob // an event that came while the trigger's check ran
}

// runTriggerChecks evaluates one trigger with run, under tickMu. The
// controllers of tests that build one by hand keep the check inline, so a
// direct call to run is still one synchronous evaluation; Run sets
// backgroundChecks.
//
// queue is for an event: it is kept (the newest) while the trigger's
// check runs, because its fields are the check's input. A poll that finds
// the check running is dropped; the next tick polls again.
func (c *Controller) runTriggerChecks(ctx context.Context, cfg *config.Config, name string, queue bool, run func(ctx context.Context, cfg *config.Config)) {
	if !c.backgroundChecks {
		run(ctx, cfg)
		return
	}
	cp := &checkPass{}
	run(withCheckPass(ctx, cp), cfg)
	if len(cp.need) == 0 {
		return
	}
	job := checkJob{need: cp.need, run: run}
	if c.checkJobs.running[name] {
		if queue {
			c.checkJobs.queued[name] = job
		}
		return
	}
	if c.checkJobs.running == nil {
		c.checkJobs.running, c.checkJobs.queued = map[string]bool{}, map[string]checkJob{}
	}
	c.checkJobs.running[name] = true
	c.wg.Go(func() { c.checkLoop(name, job) })
}

// checkLoop runs job's checks without the tick lock, then replays the
// evaluation under it, then does the job queued meanwhile, if any.
func (c *Controller) checkLoop(name string, job checkJob) {
	for {
		results := c.runChecks(c.ctx, name, job.need)
		unlock := c.lockTick(c.ctx, "check")
		if c.ctx.Err() == nil && !c.draining() {
			// One config for the replay, like an event pass (JOB-104).
			job.run(withCheckPass(c.ctx, &checkPass{results: results}), c.config())
		}
		next, more := c.checkJobs.queued[name]
		delete(c.checkJobs.queued, name)
		if !more {
			delete(c.checkJobs.running, name)
		}
		unlock()
		if !more {
			return
		}
		job = next
	}
}

// runChecks runs the checks of one trigger at the same time, within the
// bound of maxTriggerChecks across all triggers. A check that did not get
// to run (the controller is stopping) has no result.
func (c *Controller) runChecks(ctx context.Context, name string, need []checkNeed) map[string]checkResult {
	ctx, span := telemetry.StartSpan(ctx, "trigger checks "+name, attribute.String("jobshop.trigger", name))
	defer telemetry.End(span, nil)
	var mu sync.Mutex
	var wg sync.WaitGroup
	out := make(map[string]checkResult, len(need))
	for _, n := range need {
		wg.Go(func() {
			if !c.acquireCheck(ctx) {
				return
			}
			defer c.releaseCheck()
			key, why, err := c.runCheckNow(ctx, n.when)
			mu.Lock()
			out[n.key] = checkResult{out: key, why: why, err: err}
			mu.Unlock()
		})
	}
	wg.Wait()
	return out
}

func (c *Controller) acquireCheck(ctx context.Context) bool {
	c.checkSlotsOnce.Do(func() { c.checkSlots = make(chan struct{}, maxTriggerChecks) })
	select {
	case c.checkSlots <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *Controller) releaseCheck() { <-c.checkSlots }
