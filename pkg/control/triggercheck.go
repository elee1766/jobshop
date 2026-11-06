package control

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

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

// checkResult is what one check run found: runCheck's three values. A
// call that is not a check command (a Linear read, intake hooks) keeps
// its value in val.
type checkResult struct {
	out, why string
	err      error
	val      any
}

// checkNeed is a call an evaluation reached: a check command (when), or
// any other call off the tick (run, JOB-103). poller names it in the
// timeout log and metric.
type checkNeed struct {
	key     string
	when    config.TriggerWhen
	poller  string
	timeout time.Duration
	run     func(ctx context.Context) checkResult
}

// checkPass is on the ctx of an evaluation under runTriggerChecks. results
// are the finished checks; need collects the ones not found there, in the
// order they were asked.
type checkPass struct {
	results map[string]checkResult
	need    []checkNeed
}

// offTick returns the result of call n for the evaluation on ctx: run
// now when no checkPass is there (trigger.list, a run.start RPC, tests),
// the replay's result, or errCheckPending and a note of the call to make
// outside the tick (JOB-103).
func offTick(ctx context.Context, n checkNeed) (checkResult, error) {
	cp, ok := ctx.Value(checkPassKey{}).(*checkPass)
	if !ok {
		ctx, cancel := context.WithTimeout(ctx, n.timeout)
		defer cancel()
		return n.run(ctx), nil
	}
	if r, ok := cp.results[n.key]; ok {
		return r, nil
	}
	cp.note(n)
	return checkResult{}, errCheckPending
}

// note records a call the evaluation needs, once. A replay notes the calls
// it reached past the first round's (an intake read at the run start
// after the check), and checkLoop runs them in another round.
func (cp *checkPass) note(n checkNeed) {
	if !slices.ContainsFunc(cp.need, func(m checkNeed) bool { return m.key == n.key }) {
		cp.need = append(cp.need, n)
	}
}

// maxCheckRounds bounds the replays of one check job: a check, then the
// intake read of the run it starts, then a spare.
const maxCheckRounds = 3

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
		// A root span per round (the goroutine outlives the pass that
		// started it): the commands, the wait for the tick lock and the
		// replay nest under it, and a waiter behind the replay links it.
		ctx, span := c.rootSpan(c.ctx, "trigger checks "+name, attribute.String("jobshop.trigger", name))
		results := map[string]checkResult{}
		need := job.need
		var unlock func()
		for round := 1; ; round++ {
			maps.Copy(results, c.runChecks(ctx, need))
			unlock = c.lockTickFor(ctx, "check", name)
			if ctx.Err() != nil || c.draining() {
				break
			}
			// One config for the replay, like an event pass (JOB-104).
			cp := &checkPass{results: results}
			job.run(withCheckPass(ctx, cp), c.config())
			if len(cp.need) == 0 {
				break
			}
			if round == maxCheckRounds {
				c.log.WarnContext(ctx, "trigger still waits on calls after its check rounds", "trigger", name, "rounds", round)
				break
			}
			// The replay reached a call outside the tick (an intake
			// read when it starts a run, JOB-103): run it and replay again.
			unlock()
			need = cp.need
		}
		next, more := c.checkJobs.queued[name]
		delete(c.checkJobs.queued, name)
		if !more {
			delete(c.checkJobs.running, name)
		}
		unlock()
		span.End()
		if !more {
			return
		}
		job = next
	}
}

// runChecks runs the checks of one trigger at the same time, within the
// bound of maxTriggerChecks across all triggers. A check that did not get
// to run (the controller is stopping) has no result.
func (c *Controller) runChecks(ctx context.Context, need []checkNeed) map[string]checkResult {
	var mu sync.Mutex
	var wg sync.WaitGroup
	out := make(map[string]checkResult, len(need))
	for _, n := range need {
		wg.Go(func() {
			if !c.acquireCheck(ctx) {
				return
			}
			defer c.releaseCheck()
			var r checkResult
			if n.run != nil {
				rctx, cancel := context.WithTimeout(ctx, n.timeout)
				r = n.run(rctx)
				c.offTimedOut(rctx, n.poller, n.key, n.timeout)
				cancel()
			} else {
				r.out, r.why, r.err = c.runCheckNow(ctx, n.when)
				if errors.Is(r.err, context.DeadlineExceeded) && ctx.Err() == nil {
					c.checkTimedOut(n)
				}
			}
			mu.Lock()
			out[n.key] = r
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

// checkTimedOut logs and counts a check command that ran out its
// check_timeout (JOB-103).
func (c *Controller) checkTimedOut(n checkNeed) {
	timeout := config.Duration(n.when.CheckTimeout, time.Minute)
	c.log.WarnContext(c.ctx, "off-pass call exceeded its timeout", "poller", "check", "timeout", timeout, "check", n.when.Check)
	if c.metrics != nil {
		c.metrics.PollerTimeouts.Add(c.ctx, 1, metric.WithAttributes(telemetry.AttrPoller.String("check")))
	}
}
