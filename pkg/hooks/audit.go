package hooks

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type hookAttempt struct {
	Public      json.RawMessage       `json:"public,omitempty"`
	ReusedCalls []hookCallReuse       `json:"reused_calls,omitempty"`
	Output      string                `json:"output,omitempty"`
	Decisions   []automation.Decision `json:"decisions,omitempty"`
	Actions     []automation.Action   `json:"actions,omitempty"`
	Targets     []string              `json:"targets,omitempty"`
	Input       json.RawMessage       `json:"input"`
	Invocation  string                `json:"invocation"`
	Attempt     int64                 `json:"attempt"`
	StartedAt   time.Time             `json:"started_at"`
	FinishedAt  time.Time             `json:"finished_at,omitzero"`
	Error       string                `json:"error,omitempty"`
}

func (c *Engine) executeHookAttempt(ctx context.Context, col store.Collection, rec store.Record) {
	attempts, err := c.store.Collection(ctx, hookAttempts)
	if err != nil {
		return
	}
	value := hookAttempt{Invocation: rec.Key, Attempt: rec.Attempts, StartedAt: time.Now().UTC(), Input: rec.Value}
	value.Public = c.hookAttemptPublic(value, nil)
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	attempt, _, err := attempts.Put(ctx, store.Record{Key: rec.Key + "/" + strconv.FormatInt(rec.Attempts, 10), State: "UNKNOWN", Value: data})
	if err != nil {
		return
	}
	c.executeHook(ctx, col, rec)
	terminal, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	current, err := col.Get(terminal, rec.Key)
	if err != nil {
		return
	}
	var in Invocation
	if err := DecodeJSON(current.Value, &in); err != nil {
		return
	}
	if attempt, err = attempts.Get(terminal, attempt.Key); err != nil {
		return
	}
	if err := DecodeJSON(attempt.Value, &value); err != nil {
		return
	}
	value.FinishedAt, value.Error = time.Now().UTC(), in.Error
	value.Output, value.Decisions, value.Actions, value.Targets = in.Output, in.Decisions, in.Actions, in.Targets
	value.Public = c.hookAttemptPublic(value, in.Public)
	data, err = json.Marshal(value)
	if err != nil {
		return
	}
	attempt.Value = data
	switch current.State {
	case "done", "failed", "superseded", "held":
		attempt.State = current.State
	}
	_, _ = attempts.Update(terminal, attempt)
}

// A controller holds the root lock before recovery, so these are prior-process leases.
func (c *Engine) recoverHookClaims(ctx context.Context, col store.Collection) error {
	for {
		records, err := col.List(ctx, store.RecordFilter{States: []string{"pending", "planned"}, Limit: 100000})
		if err != nil {
			return err
		}
		changed := false
		for _, rec := range records {
			if rec.Holder == "" {
				continue
			}
			if !rec.LeaseUntil.After(c.now()) {
				rec.Holder = ""
			}
			rec.LeaseUntil = time.Time{}
			if _, err := col.Update(ctx, rec); err != nil && !errors.Is(err, store.ErrConflict) {
				return err
			}
			changed = true
		}
		if !changed {
			return nil
		}
	}
}

// List returns bounded public invocation summaries without private journal inputs.
func (c *Engine) List(ctx context.Context, p rpc.AutomationListParams) (any, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return nil, err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	filter := store.RecordFilter{Where: map[string]string{}, Order: store.OrderCreatedDesc, Limit: min(limit, 100)}
	if p.Run != "" {
		filter.Where["run"] = p.Run
	}
	if p.Trigger != "" {
		filter.Where["trigger"] = p.Trigger
	}
	if p.State != "" {
		filter.States = []string{p.State}
	}
	records, err := col.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i], err = c.publicHookInvocation(records[i], true)
		if err != nil {
			return nil, err
		}
	}
	return records, nil
}

// Show returns public invocation, attempt, and capability evidence in execution order.
func (c *Engine) Show(ctx context.Context, p rpc.IDParams) (any, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return nil, err
	}
	invocation, err := col.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	attempts, err := c.store.Collection(ctx, hookAttempts)
	if err != nil {
		return nil, err
	}
	calls, err := c.store.Collection(ctx, hookCalls)
	if err != nil {
		return nil, err
	}
	filter := store.RecordFilter{Where: map[string]string{"invocation": p.ID}, Order: store.OrderCreatedDesc, Limit: hookAuditAttempts + 1}
	history, err := attempts.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	filter.Limit = hookAuditCalls + 1
	trace, err := calls.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	truncated := len(history) > hookAuditAttempts || len(trace) > hookAuditCalls
	history = history[:min(len(history), hookAuditAttempts)]
	trace = trace[:min(len(trace), hookAuditCalls)]
	slices.SortFunc(trace, func(a, b store.Record) int {
		var left, right hookCallRecord
		_ = json.Unmarshal(a.Value, &left)
		_ = json.Unmarshal(b.Value, &right)
		return cmp.Compare(left.Observation.Index, right.Observation.Index)
	})
	slices.SortFunc(history, func(a, b store.Record) int {
		var left, right hookAttempt
		_ = json.Unmarshal(a.Value, &left)
		_ = json.Unmarshal(b.Value, &right)
		return cmp.Compare(left.Attempt, right.Attempt)
	})
	invocation, err = c.publicHookInvocation(invocation, false)
	if err != nil {
		return nil, err
	}
	for i := range history {
		history[i], err = c.publicHookAttempt(history[i])
		if err != nil {
			return nil, err
		}
	}
	for i := range trace {
		trace[i], err = c.publicHookCall(trace[i])
		if err != nil {
			return nil, err
		}
	}
	return rpc.AutomationResult{Invocation: invocation, Attempts: history, Calls: trace, Truncated: truncated}, nil
}
