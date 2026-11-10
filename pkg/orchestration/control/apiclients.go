package control

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/identity"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
)

// apiClients holds one client per identity, shared by every caller in the
// controller: its reads and writes, the pollers, activity, identity check
// and the built-ins. A client keeps no per-call state; a poller's request
// timeout and a callback's observer ride on the context.
type apiClients struct {
	mu     sync.Mutex
	linear map[string]*linear.Client
	// github is keyed by identity and, for a github-app, the installation
	// owner: each installation has its own token.
	github map[string]*github.Client
	// host is the gh CLI's token, for repos no identity covers.
	host *github.Client
}

// linearFor returns identity name's shared Linear client.
func (c *Controller) linearFor(name string) *linear.Client {
	c.clients.mu.Lock()
	defer c.clients.mu.Unlock()
	// The identity's api_url wins; linearURL is the in-process tests' fake.
	url := c.linearURL
	if u := c.config().Identities[name].APIURL; u != "" {
		url = u
	}
	if cl, ok := c.clients.linear[name]; ok && cl.URL == linearBase(url) {
		return cl
	}
	if c.clients.linear == nil {
		c.clients.linear = map[string]*linear.Client{}
	}
	cl := linear.New(url, identitySource{c: c, name: name})
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

// githubFor returns identity name's shared GitHub client for repo: a
// github-app has one per installation owner, a github-token one in all.
func (c *Controller) githubFor(name string, id config.Identity, repo string) *github.Client {
	key, target := name, ""
	if id.Kind == config.IdentityGitHubApp {
		owner, _, _ := strings.Cut(repo, "/")
		key, target = name+"/"+owner, owner+"/"
	}
	c.clients.mu.Lock()
	defer c.clients.mu.Unlock()
	if cl, ok := c.clients.github[key]; ok && cl.Base == apiBase(id.APIURL) {
		return cl
	}
	if c.clients.github == nil {
		c.clients.github = map[string]*github.Client{}
	}
	cl := github.New(id.APIURL, identitySource{c: c, name: name, target: target})
	c.clients.github[key] = cl
	return cl
}

// githubForRepo is the client of the GitHub identity covering repo under
// cfg, or of identity name when set.
func (c *Controller) githubForRepo(cfg *config.Config, name, repo string) (*github.Client, error) {
	if name == "" {
		name, _ = cfg.IdentityFor(cfg.IdentityNames(), repo)
	}
	id, ok := cfg.Identities[name]
	if !ok || !config.IsGitHub(id.Kind) {
		return nil, fmt.Errorf("no GitHub identity covers %s", repo)
	}
	return c.githubFor(name, id, repo), nil
}

// githubOrHost is githubForRepo, else the host's gh token for a repo no
// identity covers (activity, subscriptions and decisions predate
// identities and still serve a host logged in with gh).
func (c *Controller) githubOrHost(cfg *config.Config, name, repo string) *github.Client {
	if cl, err := c.githubForRepo(cfg, name, repo); err == nil {
		return cl
	}
	c.clients.mu.Lock()
	defer c.clients.mu.Unlock()
	if c.clients.host == nil {
		c.clients.host = github.New("", &identity.HostToken{})
	}
	return c.clients.host
}

// githubApp is a github-app identity's client authenticated as the App
// itself (its JWT), for the App endpoints such as hook deliveries.
func (c *Controller) githubApp(id config.Identity) *github.Client {
	return github.New(id.APIURL, appJWT{c: c, id: id})
}

func apiBase(u string) string {
	if u == "" {
		return github.DefaultAPIURL
	}
	return strings.TrimRight(u, "/")
}

func linearBase(u string) string {
	if u == "" {
		return linear.DefaultAPIURL
	}
	return u
}

// appJWT is a fresh App JWT per request; it lives ten minutes.
type appJWT struct {
	c  *Controller
	id config.Identity
}

func (a appJWT) Token(context.Context) (string, error) { return identity.JWTFor(a.id, a.c.now()) }

func (a appJWT) Renew(ctx context.Context, _ string) (string, error) { return a.Token(ctx) }

// identitySource is an identity's token from the minter, under the
// identity as currently configured, so a reload that rotates its key file
// or scopes applies to the next request.
type identitySource struct {
	c      *Controller
	name   string
	target string // owner/ of the installation a github-app mints for
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
