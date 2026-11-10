package hooks

import (
	"context"
	"errors"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Resume marks only safely replayable work for retry. The host then resumes its run.
func (c *Engine) Resume(ctx context.Context, run, workflow, digest string) (string, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return "", err
	}
	records, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"run": run}, States: []string{"failed", "retry"}, Order: store.OrderCreatedDesc, Limit: 1})
	if err != nil || len(records) == 0 {
		return "", err
	}
	rec := records[0]
	var in Invocation
	if err := DecodeJSON(rec.Value, &in); err != nil {
		return "", err
	}
	runtime := c.host.CurrentHooks()
	if runtime.Programs[LiveEntry(runtime.Config, in)] == nil || !RunAllowed(runtime.Programs, runtime.Config, workflow, digest) {
		return "", errors.New("restore the pinned hook configuration before resuming")
	}
	calls, err := c.store.Collection(ctx, hookCalls)
	if err != nil {
		return "", err
	}
	observed, err := calls.List(ctx, store.RecordFilter{Where: map[string]string{"invocation": rec.Key}, Limit: 10000})
	if err != nil {
		return "", err
	}
	for _, call := range observed {
		var record hookCallRecord
		if err := DecodeJSON(call.Value, &record); err != nil {
			return "", err
		}
		if call.State != "completed" || record.Observation.Redacted {
			return "", errors.New("external outcome is unknown, failed, or redacted: reconcile it before retrying this invocation")
		}
	}
	// Retry survives a crash before the run resumes. Enqueue refreshes its read set.
	rec.State = "retry"
	if _, err := col.Update(ctx, rec); err != nil {
		return "", err
	}
	return rec.Key, nil
}
