package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/store"
)

type (
	hookCallContextKey struct{}
	hookCallContext    struct {
		c          *Controller
		invocation string
		attempt    int64
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
	col, err := trace.c.store.Collection(ctx, hookCalls)
	if err != nil {
		return err
	}
	key := trace.invocation + "/" + strconv.Itoa(observation.Index)
	now := time.Now().UTC()
	if observation.Phase == "started" {
		call := hookCallRecord{Invocation: trace.invocation, Attempt: trace.attempt, StartedAt: now, Observation: observation}
		data, err := trace.c.hookCallData(call, key)
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
	if err := decodeHookJSON(rec.Value, &call); err != nil {
		return err
	}
	call.FinishedAt, call.ElapsedNS, call.Observation = now, now.Sub(call.StartedAt).Nanoseconds(), observation
	data, err := trace.c.hookCallData(call, key)
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

func (c *Controller) hookCallData(call hookCallRecord, key string) ([]byte, error) {
	call.Public = c.hookCallPublic(call)
	data, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}
	if len(data) <= 64<<10 {
		return data, nil
	}
	dir := filepath.Join(c.root.Dir, "automation", "artifacts")
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
	col, err := trace.c.store.Collection(ctx, hookCalls)
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
	if err := decodeHookJSON(rec.Value, &call); err != nil {
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
		if err := decodeHookJSON(data, &call); err != nil {
			return nil, false, err
		}
	}
	if call.Observation.Redacted || call.Observation.Capability != observation.Capability || call.Observation.InputDigest != observation.InputDigest {
		return nil, false, errors.New("recorded capability result cannot safely be reused")
	}
	attempts, err := trace.c.store.Collection(ctx, hookAttempts)
	if err != nil {
		return nil, false, err
	}
	attempt, err := attempts.Get(ctx, trace.invocation+"/"+strconv.FormatInt(trace.attempt, 10))
	if err != nil {
		return nil, false, err
	}
	var value hookAttempt
	if err := decodeHookJSON(attempt.Value, &value); err != nil {
		return nil, false, err
	}
	value.ReusedCalls = append(value.ReusedCalls, hookCallReuse{Key: key, Index: observation.Index, SourceAttempt: call.Attempt})
	value.Public = trace.c.hookAttemptPublic(value, nil)
	attempt.Value, err = json.Marshal(value)
	if err != nil {
		return nil, false, err
	}
	if _, err := attempts.Update(ctx, attempt); err != nil {
		return nil, false, err
	}
	return call.Observation.Result, true, nil
}

func (c *Controller) executeHook(ctx context.Context, col store.Collection, rec store.Record) {
	var in hookInvocation
	if err := decodeHookJSON(rec.Value, &in); err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	lc := c.live.Load()
	program := lc.hooks[in.Entry]
	if in.Plan != nil {
		c.applyHookPlan(ctx, col, rec, in)
		return
	}
	if program == nil {
		c.hookFailure(ctx, col, rec, in, errors.New("hook removed from active configuration"))
		return
	}
	expected := program.digest
	if in.Run != "" {
		expected = workflowHookDigest(lc.hooks, lc.cfg.Workflows[in.RunInput.Workflow])
	}
	if expected != in.Digest {
		c.hookFailure(ctx, col, rec, in, errors.New("hook manifest changed: restore the pinned configuration before resuming"))
		return
	}
	if !c.hookReadSetCurrent(ctx, in.ReadSet) {
		_, _ = c.saveHook(ctx, col, rec, in, "superseded")
		return
	}
	if in.Trigger != "" {
		tr, ok := lc.cfg.Triggers[in.Trigger]
		if !ok {
			_, _ = c.saveHook(ctx, col, rec, in, "superseded")
			return
		}
		if in.Event != nil && in.Event.Source == "timer" {
			// The timer slot was accepted before this durable work item existed.
			tr.When.Cron, tr.Cooldown = "", ""
		}
		// Legacy checks are synchronous here, but this worker holds no controller lock.
		st, err := c.evalTrigger(ctx, lc.cfg, in.Trigger, tr)
		if err != nil {
			c.hookFailure(ctx, col, rec, in, err)
			return
		}
		if !st.holds {
			_, _ = c.saveHook(ctx, col, rec, in, "held")
			return
		}
	}
	invokeCtx := context.WithValue(ctx, hookCallContextKey{}, hookCallContext{c: c, invocation: rec.Key, attempt: rec.Attempts})
	args := []any{in.RunInput}
	if in.StepInput != nil {
		args = append(args, in.StepInput)
	}
	if in.Event != nil {
		args = []any{in.Event, in.Args}
	}
	result, err := program.engine.Invoke(invokeCtx, program.entry, args...)
	in.Actions, in.Output, in.Decisions = result.Actions, result.Output, result.Decisions
	in.Public = c.hookInvocationPublic(in)
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	if len(result.Actions) == 0 {
		values, ok := result.Value.([]any)
		if !ok || len(values) != 0 {
			c.hookFailure(ctx, col, rec, in, errors.New("hook must return an action or a list of actions"))
			return
		}
	}
	if !c.hookReadSetCurrent(ctx, in.ReadSet) {
		_, _ = c.saveHook(ctx, col, rec, in, "superseded")
		return
	}
	plans, err := c.prepareHookRuns(ctx, lc, in)
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	c.runMu.Lock()
	graph, err := c.planHookActions(ctx, lc, rec.Key, in, program, plans)
	if err != nil {
		c.runMu.Unlock()
		if errors.Is(err, store.ErrConflict) {
			_, _ = c.saveHook(ctx, col, rec, in, "superseded")
			return
		}
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	in.Plan = &graph
	for _, ticket := range graph.Tickets {
		in.Targets = append(in.Targets, ticket.ID)
	}
	for _, plan := range plans {
		if existing, err := store.OpenRun(ctx, c.store, plan.params.Workflow, plan.params.Key); err == nil && !slices.Contains(in.Targets, existing.ID) {
			in.Targets = append(in.Targets, existing.ID)
		}
	}
	in.Public = c.hookInvocationPublic(in)
	data, err := json.Marshal(in)
	if err == nil {
		rec.Value, rec.State = data, "planned"
		rec, err = col.Update(ctx, rec)
	}
	if err == nil {
		err = c.applyPlannedHook(ctx, rec.Key, in)
	}
	c.runMu.Unlock()
	if errors.Is(err, store.ErrConflict) {
		_, _ = c.saveHook(ctx, col, rec, in, "superseded")
		return
	}
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	_, _ = c.saveHook(ctx, col, rec, in, "done")
	c.hookApplied(ctx, in)
	c.kickHooks()
}

func (c *Controller) hookReadSetCurrent(ctx context.Context, readSet map[string]int64) bool {
	for id, revision := range readSet {
		t, err := c.store.Get(ctx, id)
		if err != nil || t.Revision != revision {
			return false
		}
	}
	return true
}

func (c *Controller) applyHookPlan(ctx context.Context, col store.Collection, rec store.Record, in hookInvocation) {
	c.runMu.Lock()
	err := c.applyPlannedHook(ctx, rec.Key, in)
	c.runMu.Unlock()
	if errors.Is(err, store.ErrConflict) {
		_, _ = c.saveHook(ctx, col, rec, in, "superseded")
		return
	}
	if err != nil {
		c.hookFailure(ctx, col, rec, in, err)
		return
	}
	_, _ = c.saveHook(ctx, col, rec, in, "done")
	c.hookApplied(ctx, in)
}

func (c *Controller) hookApplied(ctx context.Context, in hookInvocation) {
	if in.Run == "" {
		return
	}
	c.wakeWaitsSubjects(runSubject(in.Run))
	if run, err := c.store.Get(ctx, in.Run); err == nil && run.Status == store.StatusClosed {
		c.removeRunDir(run.ID)
	}
	c.runBatchEvent(ctx, in.Run)
}

func (c *Controller) applyPlannedHook(ctx context.Context, key string, in hookInvocation) error {
	if in.Run != "" {
		run, err := c.store.Get(ctx, in.Run)
		if err != nil {
			return err
		}
		f, err := store.DecodeFields[store.RunFields](run)
		if err != nil {
			return err
		}
		if f.Automation != nil {
			for _, applied := range f.Automation.Applied {
				if applied == key {
					return nil
				}
			}
		}
	} else if len(in.Plan.Tickets) > 0 {
		// Stable IDs were journaled before Apply. The graph is all-or-nothing.
		if _, err := c.store.Get(ctx, in.Plan.Tickets[0].ID); err == nil {
			return nil
		} else if !errors.Is(err, store.ErrNotFound) {
			return err
		}
	}
	lc := c.live.Load()
	program := lc.hooks[in.Entry]
	if program == nil {
		return errors.New("planned hook removed from active configuration")
	}
	expected := program.digest
	if in.Run != "" {
		if in.RunInput == nil {
			return errors.New("planned workflow hook has no accepted run input")
		}
		expected = workflowHookDigest(lc.hooks, lc.cfg.Workflows[in.RunInput.Workflow])
	}
	if expected != in.Digest {
		return errors.New("planned hook manifest changed: restore the accepted configuration before applying actions")
	}
	for _, ticket := range in.Plan.Tickets {
		if ticket.Kind != store.KindRun {
			continue
		}
		var fields store.RunFields
		if err := decodeHookJSON(ticket.Fields, &fields); err != nil {
			return err
		}
		if fields.Key == "" {
			continue
		}
		existing, err := store.OpenRun(ctx, c.store, fields.Workflow, fields.Key)
		if err == nil && existing.ID != ticket.ID {
			return errors.New("a run with the planned workflow and key was created before this batch committed")
		}
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
	}
	_, err := c.store.Apply(ctx, "automation:"+key, *in.Plan)
	if err != nil {
		return fmt.Errorf("apply hook actions: %w", err)
	}
	return nil
}
