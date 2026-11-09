// Package activity is the activity feed (#255): recent GitHub and Linear
// activity for every repo and team the host works. The Service polls the
// sources, keeps a bounded window in <root>/activity.json and answers
// listings. The cache is a file and not a store table: the data is
// external and can be refetched, and store writes would add history rows
// and change events for every item.
//
// What stays in pkg/orchestration/control (JOB-149): the RPC grant checks, the
// ticket and run link map computed from the store at listing time, the
// inbox and ingest admission that new items pass through, and revert
// detection. The Service reaches them through Host.
package activity

import (
	"context"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

const (
	defaultInterval = time.Minute
	defaultWindow   = 7 * 24 * time.Hour
	defaultMax      = 1000
	defaultListMax  = 200
)

// Source fetches one repo or team. Tests replace it.
type Source interface {
	GitHub(ctx context.Context, repo string, st *RepoPoll) ([]rpc.ActivityItem, error)
	Linear(ctx context.Context, team string, since time.Time) ([]rpc.ActivityItem, error)
}

// Host is what the Service needs from the controller.
type Host interface {
	Config() *config.Config
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	// Inbox records newly polled items in the controller's inbox (#442).
	// Admission hooks see them there; admitted ones come back to Admit.
	Inbox(ctx context.Context, items []rpc.ActivityItem)
	// Ingest runs the controller's ingest pass, which admits inbox rows.
	Ingest(ctx context.Context)
	// Admitted sees the items an admission pass let through, before the
	// cache is saved (revert detection, JOB-149 keeps it in the controller).
	Admitted(ctx context.Context, items []rpc.ActivityItem)
	// Changed tells subscribers the feed changed at now.
	Changed(now time.Time)
}

// Service owns the activity cache and its poll loop.
type Service struct {
	h    Host
	log  *slog.Logger
	path string

	mu sync.Mutex
	st *State // loaded on first use with activity configured
}

// New returns a Service whose cache lives in dir.
func New(h Host, log *slog.Logger, dir string) *Service {
	return &Service{h: h, log: log, path: filepath.Join(dir, File)}
}

func (s *Service) state() *State {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.st == nil {
		s.st = loadState(s.path, s.log)
	}
	return s.st
}

// Load reads the cache if it has not been read yet.
func (s *Service) Load() { s.state() }

// Loaded reports whether the cache has been read yet. Before the first
// poll there is nothing to list.
func (s *Service) Loaded() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.st != nil
}

// setErr logs a source's error only when it changes, so a broken source
// does not log every poll; a later success clears it.
func (s *Service) setErr(st *State, src string, err error) {
	if err == nil {
		if _, ok := st.errs[src]; ok {
			s.log.Info("activity: source recovered", "source", src)
			delete(st.errs, src)
		}
		return
	}
	if st.errs[src] == err.Error() {
		return
	}
	st.errs[src] = err.Error()
	s.log.Warn("activity: poll", "source", src, "err", err)
}

// Poll runs one pass over every repo and team due. New items go to the
// inbox; the dispatcher adds the admitted ones to the feed through Admit
// (#442). It returns how many this pass added.
func (s *Service) Poll(ctx context.Context, src Source) int {
	cfg := s.h.Config()
	a := cfg.Activity
	if a == nil {
		return 0
	}
	st := s.state()
	now := s.h.Now()
	window := config.Duration(a.Window, defaultWindow)
	st.mu.Lock()
	before := st.added
	st.mu.Unlock()
	for _, repo := range cfg.ActivityRepos() {
		rp := st.Repos[repo]
		if rp == nil {
			rp = &RepoPoll{}
			st.Repos[repo] = rp
		}
		if now.Before(rp.Next) {
			continue
		}
		items, err := src.GitHub(ctx, repo, rp)
		s.setErr(st, "github:"+repo, err)
		s.h.Inbox(ctx, st.unseen(items))
	}
	for _, team := range cfg.ActivityTeams() {
		since, ok := st.Teams[team]
		if !ok || now.Sub(since) > window {
			since = now.Add(-window)
		}
		items, err := src.Linear(ctx, team, since)
		s.setErr(st, "linear:"+team, err)
		if err == nil {
			st.Teams[team] = now
		}
		s.h.Inbox(ctx, st.unseen(items))
	}
	// Drop poll state for repos and teams no longer configured.
	for r := range st.Repos {
		if !slices.Contains(cfg.ActivityRepos(), r) {
			delete(st.Repos, r)
		}
	}
	for t := range st.Teams {
		if !slices.Contains(cfg.ActivityTeams(), t) {
			delete(st.Teams, t)
		}
	}
	if err := st.save(); err != nil {
		s.setErr(st, "cache", err)
	}
	s.h.Ingest(ctx)
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.added - before
}

// Run polls until ctx ends. The cache is read on the first pass with
// activity configured.
func (s *Service) Run(ctx context.Context, src Source) {
	for {
		every := defaultInterval
		if a := s.h.Config().Activity; a != nil {
			s.Poll(ctx, src)
			every = config.Duration(a.Interval, defaultInterval)
		}
		select {
		case <-ctx.Done():
			return
		case <-s.h.After(every):
		}
	}
}

// bounds are the configured window and per-source bound.
func bounds(a *config.Activity) (time.Duration, int) {
	maxPer := a.MaxItems
	if maxPer == 0 {
		maxPer = defaultMax
	}
	return config.Duration(a.Window, defaultWindow), maxPer
}

// Admit adds items an admission pass let through, saves the cache and
// tells subscribers. It returns how many were new. With activity off it
// does nothing.
func (s *Service) Admit(ctx context.Context, items []rpc.ActivityItem) int {
	a := s.h.Config().Activity
	if a == nil {
		return 0
	}
	st := s.state()
	now := s.h.Now()
	window, maxPer := bounds(a)
	added := st.add(items, now, window, maxPer)
	s.h.Admitted(ctx, items)
	if err := st.save(); err != nil {
		s.setErr(st, "cache", err)
	}
	if added > 0 {
		s.h.Changed(now)
	}
	return added
}

// Record appends one controller-made item (a recorded `jobshop gh` call,
// a redaction) to the feed, if one is configured, and tells subscribers.
// Nothing is written when the item is already known.
func (s *Service) Record(it rpc.ActivityItem) {
	a := s.h.Config().Activity
	if a == nil {
		return
	}
	st := s.state()
	window, maxPer := bounds(a)
	if st.add([]rpc.ActivityItem{it}, it.At, window, maxPer) > 0 {
		if err := st.save(); err != nil {
			s.setErr(st, "cache", err)
		}
		s.h.Changed(it.At)
	}
}

// LinkKey is the key an item has in the link map List takes: "repo#n"
// for GitHub, the issue key for Linear, "" for an item with no ref.
func LinkKey(it rpc.ActivityItem) string {
	if it.Source == rpc.ActivityLinear {
		return it.Ref
	}
	if it.Ref == "" {
		return ""
	}
	return it.Repo + "#" + it.Ref
}

// List returns the cached items matching p, newest first, each with the
// ticket ids links holds for its LinkKey. p.Bots must already be valid.
func (s *Service) List(p rpc.ActivityListParams, links map[string][]string, now time.Time) []rpc.ActivityItem {
	st := s.state()
	limit := p.Limit
	if limit <= 0 {
		limit = defaultListMax
	}
	out := []rpc.ActivityItem{}
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, it := range st.Items {
		if p.Since > 0 && now.Sub(it.At) > p.Since {
			continue
		}
		if (p.Source != "" && it.Source != p.Source) || (p.Repo != "" && !strings.EqualFold(it.Repo, p.Repo)) ||
			(p.Kind != "" && it.Kind != p.Kind) || (p.Actor != "" && !strings.EqualFold(it.Actor, p.Actor)) ||
			(p.Bots == "humans" && it.Bot) || (p.Bots == "bots" && !it.Bot) {
			continue
		}
		it.Tickets = links[LinkKey(it)]
		if p.Ticket != "" && !slices.Contains(it.Tickets, p.Ticket) {
			continue
		}
		out = append(out, it)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// ValidBots reports whether v is an accepted Bots filter.
func ValidBots(v string) bool {
	switch v {
	case "", "all", "humans", "bots":
		return true
	}
	return false
}
