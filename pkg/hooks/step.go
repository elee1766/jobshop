package hooks

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/automation/api"
	"github.com/elee1766/jobshop/pkg/store"
)

// Script step attempts (JOB-124) are journaled like callbacks: one
// invocation record per attempt, and each capability call in hook_calls,
// so automation show lists a step's http, jev and shell calls. Their
// states keep them out of the callback worker (pending, planned) and out
// of Resume (failed, retry).
const (
	StepRunning     = "step_running"
	StepDone        = "step_done"
	StepFailed      = "step_failed"
	StepInterrupted = "step_interrupted"
)

// StepAttempt runs one script step attempt with a journal. key must be
// unique per attempt. An earlier attempt of the same step still marked
// running was cut short by a restart; it is marked interrupted, and its
// calls stay as they were (UNKNOWN when they had started).
func (c *Engine) StepAttempt(ctx context.Context, key string, p *Program, entry string, timeout time.Duration, exec api.Exec, run *Run, step *Step) (automation.Result, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return automation.Result{}, err
	}
	c.interruptStepAttempts(ctx, col, run.ID, step.TicketID)
	in := Invocation{AcceptedAt: c.now(), Run: run.ID, Step: step.TicketID, Entry: entry, Digest: p.Digest(), RunInput: run, StepInput: step}
	in.Public = c.hookInvocationPublic(in)
	data, err := json.Marshal(in)
	if err != nil {
		return automation.Result{}, err
	}
	rec, _, err := col.Put(ctx, store.Record{Key: hookKey(key), State: StepRunning, Value: data})
	if err != nil {
		return automation.Result{}, err
	}
	callCtx := context.WithValue(ctx, hookCallContextKey{}, hookCallContext{engine: c, invocation: rec.Key, attempt: rec.Attempts})
	result, runErr := p.engine.Step(callCtx, entry, timeout, exec, run, step)
	in.Output, in.Decisions = result.Logs, result.Decisions
	in.ResultMetadata, in.ResultOutput = &result.Metadata, result.Output
	state := StepDone
	if runErr != nil {
		state, in.Error = StepFailed, c.hookAuditText(runErr.Error())
	}
	in.Public = c.hookInvocationPublic(in)
	terminal, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if current, err := col.Get(terminal, rec.Key); err == nil {
		_, _ = c.saveHook(terminal, col, current, in, state)
	}
	return result, runErr
}

func (c *Engine) interruptStepAttempts(ctx context.Context, col store.Collection, run, ticket string) {
	records, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"run": run}, States: []string{StepRunning}, Limit: 1000})
	if err != nil {
		return
	}
	for _, rec := range records {
		var in Invocation
		if DecodeJSON(rec.Value, &in) != nil || in.Step != ticket {
			continue
		}
		in.Error = "controller stopped during the attempt"
		in.Public = c.hookPublicError(in.Public, in.Error)
		_, _ = c.saveHook(ctx, col, rec, in, StepInterrupted)
	}
}

// StepAttemptKey is a script step attempt's journal key: the step, its
// claim fence and a start time, since a restart reruns the same fence.
func StepAttemptKey(ticket string, fence int64, start time.Time) string {
	return strings.Join([]string{"step", ticket, strconv.FormatInt(fence, 10), strconv.FormatInt(start.UnixNano(), 10)}, "/")
}
