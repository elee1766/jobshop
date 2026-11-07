package control

import (
	"context"
	"errors"

	"github.com/elee1766/jobshop/pkg/store"
)

// The caller holds runMu. Only completed, unchanged external results are reusable.
func (c *Controller) resumeHook(ctx context.Context, actor string, run store.Ticket, fields store.RunFields) (bool, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return false, err
	}
	records, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"run": run.ID}, States: []string{"failed", "retry"}, Order: store.OrderCreatedDesc, Limit: 1})
	if err != nil || len(records) == 0 {
		return false, err
	}
	rec := records[0]
	var in hookInvocation
	if err := decodeHookJSON(rec.Value, &in); err != nil {
		return false, err
	}
	lc := c.live.Load()
	if lc.hooks[in.Entry] == nil || workflowHookDigest(lc.hooks, lc.cfg.Workflows[fields.Workflow]) != fields.Automation.Digest {
		return false, errors.New("restore the pinned hook configuration before resuming")
	}
	calls, err := c.store.Collection(ctx, hookCalls)
	if err != nil {
		return false, err
	}
	observed, err := calls.List(ctx, store.RecordFilter{Where: map[string]string{"invocation": rec.Key}, Limit: 10000})
	if err != nil {
		return false, err
	}
	for _, call := range observed {
		var record hookCallRecord
		if err := decodeHookJSON(call.Value, &record); err != nil {
			return false, err
		}
		if call.State != "completed" || record.Observation.Redacted {
			return false, errors.New("external outcome is unknown, failed, or redacted: reconcile it before retrying this invocation")
		}
	}
	// Retry survives a crash before the run becomes running. Enqueue refreshes its read set.
	rec.State = "retry"
	if _, err := col.Update(ctx, rec); err != nil {
		return false, err
	}
	if _, err := c.setRunState(ctx, actor, run, fields, store.RunRunning, ""); err != nil {
		return false, err
	}
	_, _ = store.AddComment(ctx, c.store, actor, run.ID, "Resumed automation invocation "+rec.Key+" with recorded completed calls.")
	return true, nil
}
