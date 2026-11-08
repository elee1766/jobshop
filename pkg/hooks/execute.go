package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/automation/api"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

type (
	hookCallContextKey struct{}
	hookCallContext    struct {
		engine     *Engine
		invocation string
		attempt    int64
		now        time.Time
	}
)

type hookCallRecord struct {
	Public      json.RawMessage        `json:"public,omitempty"`
	Invocation  string                 `json:"invocation"`
	Attempt     int64                  `json:"attempt"`
	StartedAt   time.Time              `json:"started_at"`
	FinishedAt  time.Time              `json:"finished_at,omitzero"`
	ElapsedNS   int64                  `json:"elapsed_ns,omitempty"`
	Observation automation.Observation `json:"observation"`
	Artifact    string                 `json:"artifact,omitempty"`
}

func observeHookCall(ctx context.Context, observation automation.Observation) error {
	trace, ok := ctx.Value(hookCallContextKey{}).(hookCallContext)
	if !ok {
		return errors.New("capability call has no durable invocation")
	}
	if observation.Phase != "started" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
	}
	col, err := trace.engine.store.Collection(ctx, hookCalls)
	if err != nil {
		return err
	}
	key := trace.invocation + "/" + strconv.Itoa(observation.Index)
	now := time.Now().UTC()
	if observation.Phase == "started" {
		call := hookCallRecord{Invocation: trace.invocation, Attempt: trace.attempt, StartedAt: now, Observation: observation}
		data, err := trace.engine.hookCallData(call, key)
		if err != nil {
			return err
		}
		_, created, err := col.Put(ctx, store.Record{Key: key, State: "UNKNOWN", Value: data})
		if err != nil {
			return err
		}
		if !created {
			return errors.New("capability already observed: reconcile the recorded external outcome before retrying")
		}
		return nil
	}
	rec, err := col.Get(ctx, key)
	if err != nil {
		return err
	}
	var call hookCallRecord
	if err := DecodeJSON(rec.Value, &call); err != nil {
		return err
	}
	call.FinishedAt, call.ElapsedNS, call.Observation = now, now.Sub(call.StartedAt).Nanoseconds(), observation
	data, err := trace.engine.hookCallData(call, key)
	if err != nil {
		return err
	}
	rec.Value, rec.State = data, observation.Phase
	if observation.Phase == "unknown" {
		rec.State = "UNKNOWN"
	}
	_, err = col.Update(ctx, rec)
	return err
}

func (c *Engine) hookCallData(call hookCallRecord, key string) ([]byte, error) {
	call.Public = c.hookCallPublic(call)
	data, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}
	if len(data) <= 64<<10 {
		return data, nil
	}
	dir := filepath.Join(c.root, "automation", "artifacts")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, hookKey(key)+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, err
	}
	call.Artifact = path
	call.Observation.Result = nil
	call.Observation.Inputs = nil
	return json.Marshal(call)
}

func replayHookCall(ctx context.Context, observation automation.Observation) (any, bool, error) {
	trace, ok := ctx.Value(hookCallContextKey{}).(hookCallContext)
	if !ok {
		return nil, false, errors.New("missing durable invocation")
	}
	col, err := trace.engine.store.Collection(ctx, hookCalls)
	if err != nil {
		return nil, false, err
	}
	key := trace.invocation + "/" + strconv.Itoa(observation.Index)
	rec, err := col.Get(ctx, key)
	if errors.Is(err, store.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var call hookCallRecord
	if err := DecodeJSON(rec.Value, &call); err != nil {
		return nil, false, err
	}
	if rec.State != "completed" {
		return nil, false, errors.New("recorded external outcome is unknown or failed")
	}
	if call.Artifact != "" {
		file, err := os.Open(call.Artifact)
		if err != nil {
			return nil, false, err
		}
		data, readErr := io.ReadAll(io.LimitReader(file, 4<<20))
		_ = file.Close()
		if readErr != nil {
			return nil, false, readErr
		}
		if err := DecodeJSON(data, &call); err != nil {
			return nil, false, err
		}
	}
	if call.Observation.Redacted || call.Observation.Capability != observation.Capability || call.Observation.InputDigest != observation.InputDigest {
		return nil, false, errors.New("recorded capability result cannot safely be reused")
	}
	attempts, err := trace.engine.store.Collection(ctx, hookAttempts)
	if err != nil {
		return nil, false, err
	}
	attempt, err := attempts.Get(ctx, trace.invocation+"/"+strconv.FormatInt(trace.attempt, 10))
	if err != nil {
		return nil, false, err
	}
	var value hookAttempt
	if err := DecodeJSON(attempt.Value, &value); err != nil {
		return nil, false, err
	}
	value.ReusedCalls = append(value.ReusedCalls, hookCallReuse{Key: key, Index: observation.Index, SourceAttempt: call.Attempt})
	value.Public = trace.engine.hookAttemptPublic(value, nil)
	attempt.Value, err = json.Marshal(value)
	if err != nil {
		return nil, false, err
	}
	if _, err := attempts.Update(ctx, attempt); err != nil {
		return nil, false, err
	}
	return call.Observation.Result, true, nil
}

func (c *Engine) executeHook(ctx context.Context, col store.Collection, rec store.Record) {
	var in Invocation
	if err := DecodeJSON(rec.Value, &in); err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	if in.Plan != nil {
		c.applyHookPlan(ctx, col, rec, in)
		return
	}
	runtime := c.host.CurrentHooks()
	program := runtime.Programs[in.Entry]
	if program == nil {
		c.hookFailure(ctx, col, rec, in, errors.New("hook removed from active configuration"))
		return
	}
	pinned := program.digest == in.Digest
	if in.Run != "" {
		pinned = in.RunInput != nil && RunPinned(runtime.Programs, runtime.Config, in.RunInput.Workflow, in.Digest)
	}
	if !pinned {
		c.hookFailure(ctx, col, rec, in, errors.New("hook manifest changed: restore the pinned configuration before resuming"))
		return
	}
	state, err := c.host.AdmitHook(ctx, runtime, in)
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	if state != "" {
		_, _ = c.saveHook(ctx, col, rec, in, state)
		return
	}
	invokeCtx := context.WithValue(ctx, hookCallContextKey{}, hookCallContext{engine: c, invocation: rec.Key, attempt: rec.Attempts, now: in.AcceptedAt})
	args := []any{in.RunInput}
	if in.StepInput != nil {
		args = append(args, in.StepInput)
	}
	if in.Event != nil {
		args = []any{in.Event, in.Args}
	}
	result, err := program.engine.Invoke(invokeCtx, program.entry, args...)
	in.Actions, in.Output, in.Decisions = result.Actions, result.Logs, result.Decisions
	in.ResultMetadata, in.ResultOutput = &result.Metadata, result.Output
	in.Public = c.hookInvocationPublic(in)
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	evidence := in.Public
	in, err = c.host.CommitHook(ctx, runtime, rec.Key, in, func(planned Invocation) error {
		planned.Public = c.hookInvocationPublic(planned)
		evidence = planned.Public
		data, err := json.Marshal(planned)
		if err != nil {
			return err
		}
		rec.Value, rec.State = data, "planned"
		rec, err = col.Update(ctx, rec)
		return err
	})
	in.Public = evidence
	if errors.Is(err, store.ErrConflict) {
		_, _ = c.saveHook(ctx, col, rec, in, "superseded")
		return
	}
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	_, _ = c.saveHook(ctx, col, rec, in, "done")
	c.host.HookApplied(ctx, in)
}

func (c *Engine) applyHookPlan(ctx context.Context, col store.Collection, rec store.Record, in Invocation) {
	err := c.host.ApplyHookPlan(ctx, rec.Key, in)
	if errors.Is(err, store.ErrConflict) {
		_, _ = c.saveHook(ctx, col, rec, in, "superseded")
		return
	}
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	_, _ = c.saveHook(ctx, col, rec, in, "done")
	c.host.HookApplied(ctx, in)
}

// hookGitHubToken asks the host for a GitHub token on behalf of the
// invocation the call belongs to, so tokens are minted per call.
func hookGitHubToken(ctx context.Context, repo string) (api.Token, error) {
	trace, ok := ctx.Value(hookCallContextKey{}).(hookCallContext)
	if !ok {
		return api.Token{}, errors.New("capability call has no durable invocation")
	}
	return trace.engine.host.GitHubToken(ctx, repo)
}

// hookLinearClient asks the host for the Linear client on behalf of the
// invocation the call belongs to.
func hookLinearClient(ctx context.Context, team string) (*linear.Client, error) {
	trace, ok := ctx.Value(hookCallContextKey{}).(hookCallContext)
	if !ok {
		return nil, errors.New("capability call has no durable invocation")
	}
	return trace.engine.host.LinearClient(ctx, team)
}

// hookNow is time.now() in a callback: when the invocation was accepted,
// so a resumed attempt sees the same time and tests follow the controller
// clock.
func hookNow(ctx context.Context) time.Time {
	if trace, ok := ctx.Value(hookCallContextKey{}).(hookCallContext); ok && !trace.now.IsZero() {
		return trace.now
	}
	return time.Now()
}

// hookTickets serves tickets(), runs() and sessions().
func hookTickets(ctx context.Context, q api.TicketQuery) ([]any, error) {
	trace, ok := ctx.Value(hookCallContextKey{}).(hookCallContext)
	if !ok {
		return nil, errors.New("capability call has no durable invocation")
	}
	return trace.engine.host.ReadTickets(ctx, q)
}
