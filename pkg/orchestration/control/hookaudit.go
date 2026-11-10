package control

import (
	"context"

	"github.com/gfx-labs/jobshop/pkg/orchestration/hooks"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func (c *Controller) automationList(ctx context.Context, sc scope, p rpc.AutomationListParams) (any, error) {
	if !sc.admin {
		return nil, forbidden("automation audit requires the admin socket")
	}
	return c.hooks().List(ctx, hooks.ListFilter{Run: p.Run, Trigger: p.Trigger, State: p.State, Limit: p.Limit})
}

func (c *Controller) automationShow(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if !sc.admin {
		return nil, forbidden("automation audit requires the admin socket")
	}
	a, err := c.hooks().Show(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return rpc.AutomationResult{Invocation: a.Invocation, Attempts: a.Attempts, Calls: a.Calls, Truncated: a.Truncated}, nil
}
