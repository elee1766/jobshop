package control

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// ghCallSeq makes activity ids unique within one controller run.
var ghCallSeq atomic.Int64

// activityRecord adds one `jobshop gh` mutating call to the activity
// feed (#349), allowed or not, so Desk shows what agents did and tried.
// Any session may record its own calls; the actor is the caller.
func (c *Controller) activityRecord(_ context.Context, sc scope, p rpc.ActivityRecordParams) (any, error) {
	if p.Action == "" || p.Decision == "" {
		return nil, fmt.Errorf("%w: action and decision are required", store.ErrInvalid)
	}
	cfg := c.config()
	if cfg.Activity == nil {
		return struct{}{}, nil // no feed configured: nothing to record into
	}
	now := c.now()
	actor := sc.session
	if actor == "" {
		actor = sc.actor
	}
	_, ref, _ := strings.Cut(p.Target, "#")
	title := fmt.Sprintf("gh %s %s (%s, exit %d)", p.Action, p.Target, p.Decision, p.ExitCode)
	it := rpc.ActivityItem{
		ID:     fmt.Sprintf("gh-call:%s:%d:%d", actor, now.UnixNano(), ghCallSeq.Add(1)),
		Source: rpc.ActivityGitHub, Repo: p.Repo, Kind: "gh-" + p.Action,
		Actor: actor, Bot: true, At: now, Title: strings.TrimSpace(title),
		Body: truncate(p.Command), Ref: ref,
	}
	c.addActivity(it)
	return struct{}{}, nil
}
