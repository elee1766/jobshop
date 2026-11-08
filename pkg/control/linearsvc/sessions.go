package linearsvc

import (
	"context"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
)

// sessionPoll spaces the agent-session sweep. A delegation or mention
// webhook runs it at once; the poll covers missed deliveries.
const sessionPoll = time.Minute

// sessionTimeout bounds one identity's sweep; each request is within the
// request timeout.
const sessionTimeout = 2 * time.Minute

// sessionThrottle throttles the sweep per identity.
type sessionThrottle struct {
	mu   sync.Mutex
	last map[string]time.Time
}

func (p *sessionThrottle) due(name string, now time.Time, force bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !force && now.Before(p.last[name].Add(sessionPoll)) {
		return false
	}
	if p.last == nil {
		p.last = map[string]time.Time{}
	}
	p.last[name] = now
	return true
}

// SweepSessions answers and so ends every open agent session of each
// linear-app identity. Delegating an issue to the app (which is also what
// self-assigning does with app:assignable) opens one; jobshop works
// through runs and comments, so it replies once and closes it.
func (s *Service) SweepSessions(_ context.Context, cfg *config.Config, force bool) {
	for _, name := range cfg.IdentityNames() {
		id := cfg.Identities[name]
		if id.Kind != config.IdentityLinearApp || len(id.Teams) == 0 || !s.host.CanMint() {
			continue
		}
		if !s.sessions.due(name, s.host.Now(), force) {
			continue
		}
		// Outside the tick (JOB-103), one sweep per identity at a time.
		s.host.OffGo("linear-sessions\x00"+name, "linear_sessions", sessionTimeout, func(ctx context.Context) {
			if err := s.endSessions(ctx, name, id); err != nil && s.host.Context().Err() == nil && !Down(err) {
				s.log().WarnContext(ctx, "linear agent sessions", "identity", name, "err", err)
			}
		})
	}
}

func (s *Service) endSessions(ctx context.Context, name string, id config.Identity) error {
	return s.Do(ctx, name, func(ctx context.Context, cl *linear.Client) error {
		open, err := cl.OpenAgentSessions(ctx)
		if err != nil {
			return err
		}
		for _, o := range open {
			if o.Team != "" && !id.CoversTeam(o.Team) {
				continue
			}
			if err := cl.EndAgentSession(ctx, o.ID, linear.AgentSessionReply); err != nil {
				return err
			}
			s.log().InfoContext(ctx, "linear agent session ended", "identity", name, "issue", o.Issue, "was", o.Status)
		}
		return nil
	})
}
