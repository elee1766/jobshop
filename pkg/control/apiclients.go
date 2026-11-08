package control

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
)

// apiClients holds one Linear client per identity, shared by every caller
// in the controller: its reads and writes, the pollers, identity check and
// the linear built-in. A client keeps no per-call state; a poller's request
// timeout and a callback's observer ride on the context.
type apiClients struct {
	mu     sync.Mutex
	linear map[string]*linear.Client
}

// linearFor returns identity name's shared Linear client.
func (c *Controller) linearFor(name string) *linear.Client {
	c.clients.mu.Lock()
	defer c.clients.mu.Unlock()
	if cl, ok := c.clients.linear[name]; ok {
		return cl
	}
	if c.clients.linear == nil {
		c.clients.linear = map[string]*linear.Client{}
	}
	cl := linear.New(c.linearURL, identitySource{c: c, name: name})
	c.clients.linear[name] = cl
	return cl
}

// linearForTeam is linearFor for the identity covering team under cfg.
func (c *Controller) linearForTeam(cfg *config.Config, team string) (*linear.Client, error) {
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), team)
	if !ok {
		return nil, fmt.Errorf("no identity covers team %s", team)
	}
	return c.linearFor(name), nil
}

// identitySource is an identity's token from the minter, under the
// identity as currently configured, so a reload that rotates its key file
// or scopes applies to the next request.
type identitySource struct {
	c      *Controller
	name   string
	target string // the repo whose installation a github-app mints for
}

func (s identitySource) identity() (config.Identity, error) {
	if s.c.minter == nil {
		return config.Identity{}, errors.New("no identity minter")
	}
	id, ok := s.c.config().Identities[s.name]
	if !ok {
		return config.Identity{}, fmt.Errorf("identity %s is not configured", s.name)
	}
	return id, nil
}

func (s identitySource) Token(ctx context.Context) (string, error) {
	id, err := s.identity()
	if err != nil {
		return "", err
	}
	t, err := s.c.minter.Token(ctx, s.name, id, s.target)
	return t.Token, err
}

// Renew mints past a refused token once, however many callers saw it
// refused (#866).
func (s identitySource) Renew(ctx context.Context, stale string) (string, error) {
	id, err := s.identity()
	if err != nil {
		return "", err
	}
	t, err := s.c.minter.Renew(ctx, s.name, id, s.target, stale)
	return t.Token, err
}
