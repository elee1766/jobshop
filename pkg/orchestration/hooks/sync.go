package hooks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/gfx-labs/jobshop/pkg/execution/automation"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Sync handlers (guard hooks, desk commands) run while a caller waits.
// Each call gets an invocation record so automation show lists its
// capability calls; nothing replays or retries it.
const (
	SyncRunning = "sync_running"
	SyncDone    = "sync_done"
	SyncFailed  = "sync_failed"
)

// Sync runs program's handler for name (for example "hooks.action" or
// "desk_commands.deploy-state") with args and journals the call.
func (c *Engine) Sync(ctx context.Context, name string, p *Program, timeout time.Duration, args ...any) (automation.Result, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return automation.Result{}, err
	}
	nonce := make([]byte, 8)
	_, _ = rand.Read(nonce)
	in := Invocation{AcceptedAt: c.now(), Sync: name, Entry: p.entry, Digest: p.Digest()}
	in.Public = c.hookInvocationPublic(in)
	data, err := json.Marshal(in)
	if err != nil {
		return automation.Result{}, err
	}
	key := hookKey("sync/" + name + "/" + in.AcceptedAt.Format(time.RFC3339Nano) + "/" + hex.EncodeToString(nonce))
	rec, _, err := col.Put(ctx, store.Record{Key: key, State: SyncRunning, Value: data})
	if err != nil {
		return automation.Result{}, err
	}
	callCtx := context.WithValue(ctx, hookCallContextKey{}, hookCallContext{engine: c, invocation: rec.Key, attempt: rec.Attempts, now: in.AcceptedAt})
	result, runErr := p.engine.Call(callCtx, p.entry, timeout, args...)
	in.Output, in.Decisions = result.Logs, result.Decisions
	in.ResultMetadata, in.ResultOutput = &result.Metadata, result.Output
	state := SyncDone
	if runErr == nil && len(result.Actions) > 0 {
		runErr = errSyncActions
	}
	if runErr != nil {
		state, in.Error = SyncFailed, c.hookAuditText(runErr.Error())
	}
	in.Public = c.hookInvocationPublic(in)
	terminal, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if current, err := col.Get(terminal, rec.Key); err == nil {
		_, _ = c.saveHook(terminal, col, current, in, state)
	}
	return result, runErr
}

var errSyncActions = errors.New("a guard or desk handler returns data, not actions")

// PruneSync deletes sync invocations last updated before cutoff (a
// running one that old was cut short by a restart),
// with their calls and call artifacts, at most limit at a time. Callback
// and step journals stay: they back replay and run history.
func (c *Engine) PruneSync(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return 0, err
	}
	calls, err := c.store.Collection(ctx, hookCalls)
	if err != nil {
		return 0, err
	}
	records, err := col.List(ctx, store.RecordFilter{States: []string{SyncRunning, SyncDone, SyncFailed}, UpdatedBefore: cutoff, Limit: limit})
	if err != nil || len(records) == 0 {
		return 0, err
	}
	keys := make([]string, 0, len(records))
	for _, rec := range records {
		recorded, err := calls.List(ctx, store.RecordFilter{Where: map[string]string{"invocation": rec.Key}, Limit: hookAuditCalls * 5})
		if err != nil {
			return 0, err
		}
		for _, call := range recorded {
			var v hookCallRecord
			if DecodeJSON(call.Value, &v) == nil && v.Artifact != "" {
				_ = os.Remove(v.Artifact)
			}
		}
		if _, err := calls.Delete(ctx, store.RecordFilter{Where: map[string]string{"invocation": rec.Key}}); err != nil {
			return 0, err
		}
		keys = append(keys, rec.Key)
	}
	return col.Delete(ctx, store.RecordFilter{Keys: keys})
}
