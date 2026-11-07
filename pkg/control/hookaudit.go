package control

import (
	"context"

	"github.com/elee1766/jobshop/pkg/rpc"
)

func (c *Controller) automationList(ctx context.Context, sc scope, p rpc.AutomationListParams) (any, error) {
	if !sc.admin {
		return nil, forbidden("automation audit requires the admin socket")
	}
	return c.hooks().List(ctx, p)
}

func (c *Controller) automationShow(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if !sc.admin {
		return nil, forbidden("automation audit requires the admin socket")
	}
	return c.hooks().Show(ctx, p)
}
