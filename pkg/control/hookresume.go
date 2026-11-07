package control

import (
	"context"

	"github.com/elee1766/jobshop/pkg/store"
)

// The caller holds runMu. The engine rejects unknown or unsafe external outcomes.
func (c *Controller) resumeHook(ctx context.Context, actor string, run store.Ticket, fields store.RunFields) (bool, error) {
	key, err := c.hooks().Resume(ctx, run.ID, fields.Workflow, fields.Automation.Digest)
	if err != nil || key == "" {
		return false, err
	}
	if _, err := c.setRunState(ctx, actor, run, fields, store.RunRunning, ""); err != nil {
		return false, err
	}
	_, _ = store.AddComment(ctx, c.store, actor, run.ID, "Resumed automation invocation "+key+" with recorded completed calls.")
	return true, nil
}
