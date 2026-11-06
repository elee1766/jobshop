package config

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"time"
)

// checkA2AGateway checks a network: a2a gateway (#547). It has no rooms
// or senders: its only traffic is escalations it sends to the peer.
func (c *Config) checkA2AGateway(g Gateway) error {
	u, err := url.Parse(g.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("url %q must be an http(s) URL", g.URL)
	}
	if u.User != nil {
		return errors.New("url must not hold credentials; use token_file")
	}
	if g.TokenFile.IsZero() {
		return errors.New("token_file is required")
	}
	p, ok := c.Pools[g.Pool]
	if !ok {
		return fmt.Errorf("unknown pool %q", g.Pool)
	}
	if !p.Chat() {
		return fmt.Errorf("pool %s is not a chat pool", g.Pool)
	}
	for _, k := range g.Escalations {
		if k != "all" && !slices.Contains(EscalationCauses, k) {
			return fmt.Errorf("escalations: %q must be all or one of %v", k, EscalationCauses)
		}
	}
	if g.Poll != "" {
		if d, err := time.ParseDuration(g.Poll); err != nil || d < time.Second {
			return fmt.Errorf("poll %q must be a duration of at least 1s", g.Poll)
		}
	}
	return nil
}

// PollDuration is Poll parsed, or 30s.
func (g Gateway) PollDuration() time.Duration {
	if d, err := time.ParseDuration(g.Poll); err == nil && d > 0 {
		return d
	}
	return 30 * time.Second
}
