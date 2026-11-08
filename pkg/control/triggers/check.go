package triggers

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
)

// Trigger checks run outside the tick (JOB-109). A check is a shell
// command with a timeout of minutes; run inside the pass it held the tick
// lock, so one slow check delayed supervise, the sweeps and every reload.
//
// A pass evaluates a trigger as before, but its ctx carries a Pass:
// where evaluation would run a check, runCheck records the command and
// returns ErrCheckPending, so nothing fires. The commands then run in a
// goroutine, outside the tick lock, at most maxChecks at a time. When
// they are done the goroutine takes the tick lock and evaluates the
// trigger again with the results in the Pass: the conditions before
// the check are read again after the wait, so a trigger that stopped
// holding does not fire, and the check still runs only when every other
// condition holds.

// ErrCheckPending is what a call returns while a pass only collects the
// checks a trigger needs. It is not a failure: callers return quietly.
var ErrCheckPending = errors.New("trigger check has not run yet")

// maxChecks bounds the check commands running at once.
const maxChecks = 8

// maxRounds bounds the replays of one check job: a check, then the intake
// read of the run it starts, then a spare.
const maxRounds = 3

// Result is what one check run found: runCheck's three values. A call
// that is not a check command (a Linear read, intake hooks) keeps its
// value in Val.
type Result struct {
	Out, Why string
	Err      error
	Val      any
}

// Need is a call an evaluation reached: a check command (When), or any
// other call off the tick (Run, JOB-103). Poller names it in the timeout
// log and metric.
type Need struct {
	Key     string
	When    config.TriggerWhen
	Poller  string
	Timeout time.Duration
	Run     func(ctx context.Context) Result
}

// Pass is on the ctx of an evaluation under RunChecked. Results are the
// finished checks; Need collects the ones not found there, in the order
// they were asked.
type Pass struct {
	Results map[string]Result
	Need    []Need
}

type passKey struct{}

// WithPass returns ctx carrying p.
func WithPass(ctx context.Context, p *Pass) context.Context {
	return context.WithValue(ctx, passKey{}, p)
}

// OffTick returns the result of call n for the evaluation on ctx: run now
// when no Pass is there (trigger.list, a run.start RPC, tests), the
// replay's result, or ErrCheckPending and a note of the call to make
// outside the tick (JOB-103).
func OffTick(ctx context.Context, n Need) (Result, error) {
	cp, ok := ctx.Value(passKey{}).(*Pass)
	if !ok {
		ctx, cancel := context.WithTimeout(ctx, n.Timeout)
		defer cancel()
		return n.Run(ctx), nil
	}
	if r, ok := cp.Results[n.Key]; ok {
		return r, nil
	}
	cp.note(n)
	return Result{}, ErrCheckPending
}

// note records a call the evaluation needs, once. A replay notes the calls
// it reached past the first round's (an intake read at the run start
// after the check), and checkLoop runs them in another round.
func (cp *Pass) note(n Need) {
	if !slices.ContainsFunc(cp.Need, func(m Need) bool { return m.Key == n.Key }) {
		cp.Need = append(cp.Need, n)
	}
}

// checkKey names a check run: the same command with the same timeout is
// one run, whichever signal or trigger asked.
func checkKey(w config.TriggerWhen) string { return w.CheckTimeout + "\x00" + w.Check }

// checkJob is a trigger evaluation that waits for its checks. run
// evaluates the trigger again on the config of the pass it replays in.
type checkJob struct {
	need []Need
	run  func(ctx context.Context, cfg *config.Config)
}

// checkJobs is the background check state of each trigger by name, under
// the tick lock: at most one check job per trigger runs at a time.
type checkJobs struct {
	running map[string]bool
	queued  map[string]checkJob // an event that came while the trigger's check ran
}

// RunChecked evaluates one trigger with run, under the tick lock. A host
// that is not Background (a controller built by hand in a test) keeps the
// check inline, so a direct call to run is still one synchronous
// evaluation.
//
// queue is for an event: it is kept (the newest) while the trigger's
// check runs, because its fields are the check's input. A poll that finds
// the check running is dropped; the next tick polls again.
func (e *Engine) RunChecked(ctx context.Context, cfg *config.Config, name string, queue bool, run func(ctx context.Context, cfg *config.Config)) {
	if !e.host.Background() {
		run(ctx, cfg)
		return
	}
	cp := &Pass{}
	run(WithPass(ctx, cp), cfg)
	if len(cp.Need) == 0 {
		return
	}
	job := checkJob{need: cp.Need, run: run}
	if e.jobs.running[name] {
		if queue {
			e.jobs.queued[name] = job
		}
		return
	}
	if e.jobs.running == nil {
		e.jobs.running, e.jobs.queued = map[string]bool{}, map[string]checkJob{}
	}
	e.jobs.running[name] = true
	e.host.Go(func() { e.checkLoop(name, job) })
}

// checkLoop runs job's checks without the tick lock, then replays the
// evaluation under it, then does the job queued meanwhile, if any.
func (e *Engine) checkLoop(name string, job checkJob) {
	for {
		// A root span per round (the goroutine outlives the pass that
		// started it): the commands, the wait for the tick lock and the
		// replay nest under it, and a waiter behind the replay links it.
		ctx, span := e.host.RootSpan(e.host.Context(), "trigger checks "+name, attribute.String("jobshop.trigger", name))
		results := map[string]Result{}
		need := job.need
		var unlock func()
		for round := 1; ; round++ {
			maps.Copy(results, e.RunChecks(ctx, need))
			unlock = e.host.LockTick(ctx, "check", name)
			if ctx.Err() != nil || e.host.Draining() {
				break
			}
			// One config for the replay, like an event pass (JOB-104).
			cp := &Pass{Results: results}
			job.run(WithPass(ctx, cp), e.host.Config())
			if len(cp.Need) == 0 {
				break
			}
			if round == maxRounds {
				e.log().WarnContext(ctx, "trigger still waits on calls after its check rounds", "trigger", name, "rounds", round)
				break
			}
			// The replay reached a call outside the tick (an intake
			// read when it starts a run, JOB-103): run it and replay again.
			unlock()
			need = cp.Need
		}
		next, more := e.jobs.queued[name]
		delete(e.jobs.queued, name)
		if !more {
			delete(e.jobs.running, name)
		}
		unlock()
		span.End()
		if !more {
			return
		}
		job = next
	}
}

// RunChecks runs the checks of one trigger at the same time, within the
// bound of maxChecks across all triggers. A check that did not get to run
// (the controller is stopping) has no result.
func (e *Engine) RunChecks(ctx context.Context, need []Need) map[string]Result {
	var mu sync.Mutex
	var wg sync.WaitGroup
	out := make(map[string]Result, len(need))
	for _, n := range need {
		wg.Go(func() {
			if !e.acquire(ctx) {
				return
			}
			defer e.release()
			var r Result
			if n.Run != nil {
				rctx, cancel := context.WithTimeout(ctx, n.Timeout)
				r = n.Run(rctx)
				if errors.Is(rctx.Err(), context.DeadlineExceeded) {
					e.host.TimedOut(n.Poller, n.Timeout, "key", strings.ReplaceAll(n.Key, "\x00", " "))
				}
				cancel()
			} else {
				r.Out, r.Why, r.Err = e.runCheckNow(ctx, n.When)
				if errors.Is(r.Err, context.DeadlineExceeded) && ctx.Err() == nil {
					e.host.TimedOut("check", config.Duration(n.When.CheckTimeout, time.Minute), "check", n.When.Check)
				}
			}
			mu.Lock()
			out[n.Key] = r
			mu.Unlock()
		})
	}
	wg.Wait()
	return out
}

func (e *Engine) acquire(ctx context.Context) bool {
	e.slotsOnce.Do(func() { e.slots = make(chan struct{}, maxChecks) })
	select {
	case e.slots <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (e *Engine) release() { <-e.slots }
