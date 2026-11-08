package control

import (
	"context"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
)

// linearSessionPoll spaces the agent-session sweep. A delegation or
// mention webhook runs it at once; the poll covers missed deliveries.
const linearSessionPoll = time.Minute

// linearSessions throttles endLinearAgentSessions per identity.
type linearSessions struct {
	mu   sync.Mutex
	last map[string]time.Time
}

// sweepLinearAgentSessions answers and so ends every open agent session
// of each linear-app identity. Delegating an issue to the app (which is
// also what self-assigning does with app:assignable) opens one; jobshop
// works through runs and comments, so it replies once and closes it.
func (c *Controller) sweepLinearAgentSessions(_ context.Context, cfg *config.Config, force bool) {
	for _, name := range cfg.IdentityNames() {
		id := cfg.Identities[name]
		if id.Kind != config.IdentityLinearApp || len(id.Teams) == 0 || c.minter == nil {
			continue
		}
		if !c.linSessions.due(name, c.now(), force) {
			continue
		}
		// Outside the tick (JOB-103), one sweep per identity at a time.
		c.offGo("linear-sessions\x00"+name, "linear_sessions", linearSessionTimeout, false, func(ctx context.Context) any {
			if err := c.endLinearAgentSessions(ctx, name, id); err != nil && c.ctx.Err() == nil {
				c.log.WarnContext(ctx, "linear agent sessions", "identity", name, "err", err)
			}
			return nil
		})
	}
}

// linearSessionTimeout bounds one identity's sweep; each request is
// within linearRequestTimeout.
const linearSessionTimeout = 2 * time.Minute

func (s *linearSessions) due(name string, now time.Time, force bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !force && now.Before(s.last[name].Add(linearSessionPoll)) {
		return false
	}
	if s.last == nil {
		s.last = map[string]time.Time{}
	}
	s.last[name] = now
	return true
}

func (c *Controller) endLinearAgentSessions(ctx context.Context, name string, id config.Identity) error {
	ctx = pollerCtx(ctx)
	cl := c.linearFor(name)
	open, err := cl.OpenAgentSessions(ctx)
	if err != nil {
		return err
	}
	for _, s := range open {
		if s.Team != "" && !id.CoversTeam(s.Team) {
			continue
		}
		if err := cl.EndAgentSession(ctx, s.ID, linear.AgentSessionReply); err != nil {
			return err
		}
		c.log.InfoContext(ctx, "linear agent session ended", "identity", name, "issue", s.Issue, "was", s.Status)
	}
	return nil
}
