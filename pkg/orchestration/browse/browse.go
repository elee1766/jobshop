// Package browse is Desk's read-only view of Linear and GitHub (JOB-163).
//
// It answers lists, one issue or PR with its discussion, and the Linear
// catalog, as the identities the controller already holds and only for the
// teams and repos it is configured for. The browser never names a URL and
// never sees a credential. Reads are cached, shared between callers that ask
// the same thing at once, and stop at a provider's rate limit until it says
// the limit lifts; a failed read answers with the last good copy marked
// stale. Nothing here writes to a provider.
package browse

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/support/redact"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Bounds. A list page is 50 rows unless asked for up to 100, a discussion
// page 30 messages, and no text field is sent whole past its limit.
const (
	pageSize     = 50
	maxPageSize  = 100
	threadPage   = 30
	maxBody      = 64 << 10
	maxComment   = 16 << 10
	maxQuery     = 200
	maxCursor    = 1024
	requestLimit = 25 * time.Second
)

// Cache timing. A fresh copy answers without asking the provider. After the
// TTL the next read asks again, and a failed ask falls back to the old copy
// for up to staleMax. A forced refresh inside minRefresh of the last read is
// answered from it, so a held key cannot hammer a provider.
const (
	listTTL     = 30 * time.Second
	detailTTL   = 20 * time.Second
	catalogTTL  = 10 * time.Minute
	minRefresh  = 5 * time.Second
	staleMax    = 30 * time.Minute
	maxEntries  = 400
	authBackoff = 30 * time.Second
	downBackoff = 10 * time.Second
	// A rate limit with no reset time blocks for rateDefault; one that
	// names a time never blocks past rateMax.
	rateDefault = 60 * time.Second
	rateMax     = 15 * time.Minute
)

// WritesReason is why Desk's write controls are disabled.
const WritesReason = "Provider writes are not connected yet. Desk reads Linear and GitHub as the service identity and does not act as you. Use the open in provider link to change anything."

// Host is what the service needs from the controller.
type Host interface {
	Config() *config.Config
	Now() time.Time
	// Linear is the client of the identity covering team.
	Linear(team string) (*linear.Client, error)
	// GitHub is the client of the identity covering repo. A repo no
	// identity covers is an error: Desk never reads as the host's own
	// gh login.
	GitHub(repo string) (*github.Client, error)
}

// Service owns the cache and the rate limit state.
type Service struct {
	h   Host
	log *slog.Logger
	rev atomic.Int64

	mu      sync.Mutex
	seq     int64
	cache   map[string]*entry
	epoch   map[string]int64 // block key -> seq of its last invalidation
	blocked map[string]block
	flights map[string]*flight
}

type entry struct {
	value any
	at    time.Time
	seq   int64
	used  time.Time
}

type block struct {
	until time.Time
	code  string
	msg   string
}

type flight struct {
	done   chan struct{}
	value  any
	status rpc.BrowseStatus
}

// New returns a service that reads through h.
func New(h Host, log *slog.Logger) *Service {
	return &Service{
		h: h, log: log,
		cache: map[string]*entry{}, epoch: map[string]int64{},
		blocked: map[string]block{}, flights: map[string]*flight{},
	}
}

// Scope lists what may be read. A Linear team comes from a Linear identity's
// teams ("*" names none), a GitHub repo from the workflows and activity.repos
// that the activity feed already polls.
func (s *Service) Scope() rpc.BrowseScopeResult {
	cfg := s.h.Config()
	res := rpc.BrowseScopeResult{
		Revision: s.rev.Load(),
		Linear:   []rpc.BrowseLinearScope{},
		GitHub:   []rpc.BrowseGitHubScope{},
		Writes:   rpc.BrowseWrites{Reason: WritesReason},
		Limits:   rpc.BrowseLimits{PageSize: pageSize, MaxPageSize: maxPageSize, ThreadPage: threadPage},
	}
	for _, name := range cfg.IdentityNames() {
		id := cfg.Identities[name]
		if !config.IsLinear(id.Kind) {
			continue
		}
		for _, t := range id.Teams {
			if t == "*" || slices.ContainsFunc(res.Linear, func(x rpc.BrowseLinearScope) bool { return x.Team == t }) {
				continue
			}
			res.Linear = append(res.Linear, rpc.BrowseLinearScope{Team: t, Workspace: id.Workspace})
		}
	}
	slices.SortFunc(res.Linear, func(a, b rpc.BrowseLinearScope) int { return strings.Compare(a.Team, b.Team) })
	for _, r := range cfg.ActivityRepos() {
		res.GitHub = append(res.GitHub, rpc.BrowseGitHubScope{Repo: r})
	}
	return res
}

func (s *Service) linearTeam(team string) error {
	if slices.ContainsFunc(s.Scope().Linear, func(x rpc.BrowseLinearScope) bool { return x.Team == team }) {
		return nil
	}
	return fmt.Errorf("%w: linear team %q is not configured for Desk", domain.ErrInvalid, team)
}

func (s *Service) githubRepo(repo string) error {
	if slices.ContainsFunc(s.Scope().GitHub, func(x rpc.BrowseGitHubScope) bool { return x.Repo == repo }) {
		return nil
	}
	return fmt.Errorf("%w: github repo %q is not configured for Desk", domain.ErrInvalid, repo)
}

// Invalidate marks everything cached for one team or repo as out of date and
// moves the revision. The controller calls it when a delivery or a poll shows
// the provider changed; the activity event that goes with it makes Desk's
// live stream send the new revision, and an open view reads again.
func (s *Service) Invalidate(provider, scope string) {
	s.mu.Lock()
	s.seq++
	s.epoch[provider+":"+scope] = s.seq
	s.mu.Unlock()
	s.rev.Add(1)
}

// Revision is the scope revision.
func (s *Service) Revision() int64 { return s.rev.Load() }

// fetch answers key from the cache or load. block names what a provider
// failure blocks: every key of one team or repo shares it, so a rate limit
// stops them all. T is the zero value when the status is an error.
func fetch[T any](ctx context.Context, s *Service, key, blockKey string, ttl time.Duration, refresh bool, load func(context.Context) (T, error)) (T, rpc.BrowseStatus) {
	var zero T
	now := s.h.Now()
	s.mu.Lock()
	e := s.cache[key]
	if e != nil {
		e.used = now
		fresh := e.seq > s.epoch[blockKey] && now.Sub(e.at) < ttl
		if fresh && (!refresh || now.Sub(e.at) < minRefresh) {
			s.mu.Unlock()
			return e.value.(T), rpc.BrowseStatus{State: rpc.BrowseCached, FetchedAt: e.at}
		}
	}
	if b, ok := s.blocked[blockKey]; ok && now.Before(b.until) {
		s.mu.Unlock()
		st := rpc.BrowseStatus{State: rpc.BrowseError, Code: b.code, Message: b.msg, RetryAt: b.until}
		if e != nil && now.Sub(e.at) < staleMax {
			st.State, st.FetchedAt = rpc.BrowseStale, e.at
			return e.value.(T), st
		}
		return zero, st
	}
	if f := s.flights[key]; f != nil {
		s.mu.Unlock()
		select {
		case <-f.done:
		case <-ctx.Done():
			return zero, rpc.BrowseStatus{State: rpc.BrowseError, Code: rpc.BrowseCodeUnavailable, Message: "request cancelled"}
		}
		v, _ := f.value.(T)
		return v, f.status
	}
	f := &flight{done: make(chan struct{})}
	s.flights[key] = f
	s.mu.Unlock()

	rctx, cancel := context.WithTimeout(ctx, requestLimit)
	v, err := load(rctx)
	cancel()
	var st rpc.BrowseStatus
	var out T
	s.mu.Lock()
	if err == nil {
		now = s.h.Now()
		s.seq++
		s.cache[key] = &entry{value: v, at: now, seq: s.seq, used: now}
		s.evict(now)
		delete(s.blocked, blockKey)
		out, st = v, rpc.BrowseStatus{State: rpc.BrowseFresh, FetchedAt: now}
	} else {
		now = s.h.Now()
		c := classify(err, now)
		if c.block > 0 {
			s.blocked[blockKey] = block{until: now.Add(c.block), code: c.code, msg: c.msg}
		}
		st = rpc.BrowseStatus{State: rpc.BrowseError, Code: c.code, Message: c.msg}
		if c.block > 0 {
			st.RetryAt = now.Add(c.block)
		}
		if e = s.cache[key]; e != nil && now.Sub(e.at) < staleMax {
			out = e.value.(T)
			st.State, st.FetchedAt = rpc.BrowseStale, e.at
		}
		if s.log != nil {
			s.log.WarnContext(ctx, "browse: provider read", "key", key, "code", c.code, "err", c.msg)
		}
	}
	f.value, f.status = out, st
	delete(s.flights, key)
	s.mu.Unlock()
	close(f.done)
	return out, st
}

// evict drops entries past staleMax, then the least recently used while the
// cache is over its bound. The caller holds s.mu.
func (s *Service) evict(now time.Time) {
	for k, e := range s.cache {
		if now.Sub(e.at) > staleMax {
			delete(s.cache, k)
		}
	}
	for len(s.cache) > maxEntries {
		oldest, when := "", now.Add(time.Hour)
		for k, e := range s.cache {
			if e.used.Before(when) {
				oldest, when = k, e.used
			}
		}
		delete(s.cache, oldest)
	}
}

// failure is a provider error reduced to what Desk shows.
type failure struct {
	code  string
	msg   string
	block time.Duration
}

// classify turns a provider error into a status. The message never carries
// more than the provider's own short reason, scanned for secrets.
func classify(err error, now time.Time) failure {
	var gh *github.StatusError
	var ln *linear.StatusError
	switch {
	case errors.As(err, &gh):
		switch {
		case gh.RateLimited():
			return limited("GitHub", gh.RetryAt, now)
		case gh.Status == 401:
			return failure{rpc.BrowseCodeAuth, "GitHub rejected the service identity's credentials.", authBackoff}
		case gh.Status == 403:
			return failure{code: rpc.BrowseCodeAuth, msg: "The service identity may not read this on GitHub."}
		case gh.Status == 404:
			return failure{code: rpc.BrowseCodeNotFound, msg: "GitHub has no such item, or the identity cannot see it."}
		case gh.Status >= 500:
			return failure{rpc.BrowseCodeUnavailable, "GitHub is unavailable.", downBackoff}
		}
	case errors.As(err, &ln):
		switch {
		case ln.Code == 429:
			return limited("Linear", ln.RetryAfter, now)
		case ln.Code == 401 || ln.Code == 403:
			return failure{rpc.BrowseCodeAuth, "Linear rejected the service identity's credentials.", authBackoff}
		case ln.Code >= 500:
			return failure{rpc.BrowseCodeUnavailable, "Linear is unavailable.", downBackoff}
		}
	case errors.Is(err, errNotFound):
		return failure{code: rpc.BrowseCodeNotFound, msg: "Not found, or the identity cannot see it."}
	case errors.Is(err, context.DeadlineExceeded):
		return failure{rpc.BrowseCodeUnavailable, "The provider did not answer in time.", downBackoff}
	case linear.Unavailable(err):
		return failure{rpc.BrowseCodeUnavailable, "The provider is unreachable.", downBackoff}
	}
	if u := (*url.Error)(nil); errors.As(err, &u) {
		return failure{rpc.BrowseCodeUnavailable, "The provider is unreachable.", downBackoff}
	}
	// Builtin compiles the rules on first use. Doing it in a package var
	// cost ~0.35s at the start of every process linking this package under
	// -race, including each fake agent a control test spawns (JOB-187).
	msg := redact.Builtin().Redactor(nil).Text(err.Error())
	if len(msg) > 200 {
		msg = strings.ToValidUTF8(msg[:200], "") + "…"
	}
	if strings.Contains(msg, "no identity") {
		return failure{code: rpc.BrowseCodeNotConfigured, msg: msg}
	}
	return failure{code: rpc.BrowseCodeError, msg: msg}
}

func limited(provider string, at, now time.Time) failure {
	wait := rateDefault
	if !at.IsZero() {
		wait = at.Sub(now)
	}
	wait = min(max(wait, time.Second), rateMax)
	return failure{rpc.BrowseCodeRateLimited, provider + " rate limit reached.", wait}
}

// errNotFound is an issue or PR the provider has no record of.
var errNotFound = errors.New("not found")

// clip bounds s to n bytes on a rune boundary and says whether it cut.
func clip(s string, n int) (string, bool) {
	if len(s) <= n {
		return s, false
	}
	return strings.ToValidUTF8(s[:n], ""), true
}

func bound(limit int) int {
	switch {
	case limit <= 0:
		return pageSize
	case limit > maxPageSize:
		return maxPageSize
	}
	return limit
}

func checkCursor(c string) error {
	if len(c) > maxCursor {
		return fmt.Errorf("%w: cursor is too long", domain.ErrInvalid)
	}
	return nil
}

// List reads one page of a team's issues or a repo's PRs or issues.
func (s *Service) List(ctx context.Context, p rpc.BrowseListParams) (rpc.BrowseListResult, error) {
	if err := checkCursor(p.Cursor); err != nil {
		return rpc.BrowseListResult{}, err
	}
	switch p.Provider {
	case rpc.BrowseLinear:
		return s.listLinear(ctx, p)
	case rpc.BrowseGitHub:
		return s.listGitHub(ctx, p)
	}
	return rpc.BrowseListResult{}, fmt.Errorf("%w: provider must be linear or github", domain.ErrInvalid)
}

// Show reads one issue or PR with its discussion.
func (s *Service) Show(ctx context.Context, p rpc.BrowseShowParams) (rpc.BrowseDetail, error) {
	if err := checkCursor(p.Cursor); err != nil {
		return rpc.BrowseDetail{}, err
	}
	switch p.Provider {
	case rpc.BrowseLinear:
		return s.showLinear(ctx, p)
	case rpc.BrowseGitHub:
		return s.showGitHub(ctx, p)
	}
	return rpc.BrowseDetail{}, fmt.Errorf("%w: provider must be linear or github", domain.ErrInvalid)
}

// Catalog reads the lists behind a team's property dropdowns.
func (s *Service) Catalog(ctx context.Context, p rpc.BrowseCatalogParams) (rpc.BrowseCatalogResult, error) {
	if err := s.linearTeam(p.Team); err != nil {
		return rpc.BrowseCatalogResult{}, err
	}
	res, st := fetch(ctx, s, "linear:catalog:"+p.Team, "linear:"+p.Team, catalogTTL, p.Refresh, func(ctx context.Context) (rpc.BrowseCatalogResult, error) {
		cl, err := s.h.Linear(p.Team)
		if err != nil {
			return rpc.BrowseCatalogResult{}, err
		}
		var out rpc.BrowseCatalogResult
		for _, k := range []struct {
			kind string
			into *[]rpc.BrowseOption
		}{{"states", &out.States}, {"projects", &out.Projects}, {"users", &out.Users}, {"labels", &out.Labels}} {
			entries, err := cl.Catalog(ctx, p.Team, k.kind)
			if err != nil {
				return rpc.BrowseCatalogResult{}, err
			}
			*k.into = make([]rpc.BrowseOption, 0, len(entries))
			for _, e := range entries {
				typ := e.Type
				if k.kind == "users" {
					typ = "" // an email is not for the browser
				}
				*k.into = append(*k.into, rpc.BrowseOption{ID: e.ID, Name: e.Name, Type: typ})
			}
		}
		return out, nil
	})
	res = fillCatalog(res)
	res.Status = st
	return res, nil
}

func fillCatalog(r rpc.BrowseCatalogResult) rpc.BrowseCatalogResult {
	for _, l := range []*[]rpc.BrowseOption{&r.States, &r.Projects, &r.Users, &r.Labels} {
		if *l == nil {
			*l = []rpc.BrowseOption{}
		}
	}
	return r
}
