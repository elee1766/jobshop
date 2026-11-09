package linearsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// intakePages bounds one Linear read (250 issues a page).
const intakePages = 8

// intakeTimeout bounds one intake read: up to intakePages list pages and a
// context read per match, each request within the request timeout.
const intakeTimeout = 2 * time.Minute

// intakeCache caches the matching issues of each linear_issue_no_run
// entry between polls, so a 30s tick does not read Linear every time.
// Which of them have an open run is decided from the store every tick.
type intakeCache struct {
	mu    sync.Mutex
	cache map[string]intakeRead
}

type intakeRead struct {
	at     time.Time
	issues []linear.Summary
}

// Finding is an issue the linear_issue_no_run signal found.
type Finding struct {
	Subject, Detail string
}

// Signal is the linear_issue_no_run signal (#287): open issues of
// sig.Team in sig.States with one of sig.Labels and none of
// sig.SkipLabels, minus those an open run carries in meta.linear.issue.
// The subject is the issue key. A Linear read error is logged and the
// signal finds nothing, so an outage does not stop the other signals.
// pending is true while the read has not run outside the tick yet: the
// caller treats it as a check that is still running.
func (s *Service) Signal(ctx context.Context, cfg *config.Config, sig config.ScanSignal) (found []Finding, pending bool, err error) {
	if sig.Team == "" {
		return nil, false, nil
	}
	issues, pending, err := s.intakeIssues(ctx, cfg, sig)
	if pending {
		return nil, true, nil
	}
	if err != nil {
		if !Down(err) { // logged once per outage
			s.log().WarnContext(ctx, "linear_issue_no_run", "team", sig.Team, "err", err)
		}
		return nil, false, nil
	}
	busy, err := s.issuesWithOpenRun(ctx)
	if err != nil {
		return nil, false, err
	}
	grace := graceOf(sig)
	now := s.host.Now()
	var out []Finding
	for _, is := range issues {
		if busy[is.Identifier] {
			continue
		}
		if grace > 0 && !is.UpdatedAt.IsZero() && now.Before(is.UpdatedAt.Add(grace)) {
			continue
		}
		out = append(out, Finding{is.Identifier, fmt.Sprintf("%s, no run: %s", is.State.Name, is.Title)})
	}
	return out, false, nil
}

// graceOf is sig.Grace; empty is off.
func graceOf(sig config.ScanSignal) time.Duration {
	d, _ := time.ParseDuration(sig.Grace)
	return d
}

// HasOpenRun rechecks the store just before a triage run starts: the
// planner may have started the issue's run during the scan.
func (s *Service) HasOpenRun(ctx context.Context, key string) (bool, error) {
	busy, err := s.issuesWithOpenRun(ctx)
	if err != nil {
		return false, err
	}
	return busy[strings.ToUpper(key)], nil
}

// intakeIssues returns the matching issues, read from Linear at most once
// per poll interval. In a pass the read runs outside the tick like a
// check (JOB-103): pending is true and the replay has the issues.
func (s *Service) intakeIssues(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]linear.Summary, bool, error) {
	key := strings.Join([]string{sig.Team, sig.States, sig.Labels, sig.SkipLabels}, "|")
	poll := config.Duration(sig.Poll, config.DefaultLinearPoll)
	now := s.host.Now()
	s.intake.mu.Lock()
	r, ok := s.intake.cache[key]
	s.intake.mu.Unlock()
	if ok && now.Before(r.at.Add(poll)) {
		return r.issues, false, nil
	}
	val, pending, err := s.host.OffTick(ctx, "linear\x00"+key, "linear_issue_no_run", intakeTimeout, func(ctx context.Context) (any, error) {
		return s.readIntake(ctx, cfg, sig)
	})
	if pending || err != nil {
		return nil, pending, err
	}
	issues, _ := val.([]linear.Summary)
	s.intake.mu.Lock()
	defer s.intake.mu.Unlock()
	if s.intake.cache == nil {
		s.intake.cache = map[string]intakeRead{}
	}
	s.intake.cache[key] = intakeRead{at: now, issues: issues}
	return issues, false, nil
}

// InvalidateIntake drops cached linear_issue_no_run reads of team, so the
// next tick reads Linear again.
func (s *Service) InvalidateIntake(team string) {
	s.intake.mu.Lock()
	defer s.intake.mu.Unlock()
	for k := range s.intake.cache {
		if t, _, _ := strings.Cut(k, "|"); strings.EqualFold(t, team) {
			delete(s.intake.cache, k)
		}
	}
}

// readIntake lists the team's issues as the identity covering the team and
// keeps the open ones that match. Filtering is done here: the list filter
// takes one state and one label, the vars take several.
func (s *Service) readIntake(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]linear.Summary, error) {
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), sig.Team)
	if !ok {
		return nil, fmt.Errorf("no identity covers team %s", sig.Team)
	}
	var out []linear.Summary
	err := s.Do(ctx, name, func(ctx context.Context, cl *linear.Client) error {
		var err error
		out, err = s.readIntakeWith(ctx, cl, cfg, sig)
		return err
	})
	return out, err
}

func (s *Service) readIntakeWith(ctx context.Context, cl *linear.Client, cfg *config.Config, sig config.ScanSignal) ([]linear.Summary, error) {
	states, labels, skip := config.SplitList(sig.States), config.SplitList(sig.Labels), config.SplitList(sig.SkipLabels)
	f := linear.Filter{Team: sig.Team, Limit: 250}
	var out []linear.Summary
	for range intakePages {
		p, err := cl.List(ctx, f)
		if err != nil {
			return nil, err
		}
		for _, is := range p.Issues {
			if !intakeMatch(is, states, labels, skip) {
				continue
			}
			blocked, err := s.intakeBlocked(ctx, cl, cfg, is.Identifier)
			if err != nil {
				return nil, err
			}
			if !blocked {
				out = append(out, is)
			}
		}
		if !p.HasNext {
			return out, nil
		}
		f.After = p.EndCursor
	}
	s.log().WarnContext(ctx, "linear_issue_no_run: team has more issues than one read covers", "team", sig.Team, "pages", intakePages)
	return out, nil
}

// intakeBlocked reports whether a matching issue waits on another: an open
// Linear blocker or an open GitHub issue named by a "Depends on" line
// (JOB-68). The list read carries neither, so each match is read in full,
// once per poll. A read error keeps the issue: intake then behaves as
// before this check. Linear not answering (JOB-118) ends the read instead,
// so a hung API costs one request timeout, not one per issue.
func (s *Service) intakeBlocked(ctx context.Context, cl *linear.Client, cfg *config.Config, key string) (bool, error) {
	is, err := cl.GetContext(ctx, key, false)
	if linear.Unavailable(err) {
		return false, err
	}
	if err != nil {
		s.log().WarnContext(ctx, "linear_issue_no_run: read issue", "issue", key, "err", err)
		return false, nil
	}
	if len(is.OpenBlockers()) > 0 {
		return true, nil
	}
	issueOpen := func(ctx context.Context, d linear.GitHubDep) (bool, error) {
		return s.host.GitHubIssueOpen(ctx, cfg, d)
	}
	if err := is.CheckDeps(ctx, repoOwner(cfg), issueOpen); err != nil {
		s.log().WarnContext(ctx, "linear_issue_no_run: dependency", "issue", key, "err", err)
		return false, nil
	}
	return len(is.OpenDeps) > 0, nil
}

// repoOwner is the GitHub owner of the first (by name) Linear tracked
// workflow's repo, for a bare "Depends on repo#n".
func repoOwner(cfg *config.Config) string {
	names := slices.Sorted(maps.Keys(cfg.Workflows))
	for _, n := range names {
		w := cfg.Workflows[n]
		if w.Tracker == config.TrackerLinear {
			if owner, _, ok := strings.Cut(w.Repo, "/"); ok {
				return owner
			}
		}
	}
	return ""
}

// intakeMatch: open, in one of states (any when empty), with one of labels
// (any when empty), with none of skip. Names compare case-insensitively.
func intakeMatch(is linear.Summary, states, labels, skip []string) bool {
	if !linear.OpenStateType(is.State.Type) {
		return false
	}
	has := func(names []string, s string) bool {
		return slices.ContainsFunc(names, func(n string) bool { return strings.EqualFold(n, s) })
	}
	if len(states) > 0 && !has(states, is.State.Name) {
		return false
	}
	if slices.ContainsFunc(is.Labels, func(l string) bool { return has(skip, l) }) {
		return false
	}
	return len(labels) == 0 || slices.ContainsFunc(is.Labels, func(l string) bool { return has(labels, l) })
}

// issuesWithOpenRun is meta.linear.issue of every open run.
func (s *Service) issuesWithOpenRun(ctx context.Context) (map[string]bool, error) {
	out := map[string]bool{}
	for _, st := range []domain.Status{domain.StatusOpen, domain.StatusInProgress} {
		runs, err := s.host.List(ctx, store.Filter{Kind: domain.KindRun, Status: st})
		if err != nil {
			return nil, err
		}
		for _, r := range runs {
			var m struct {
				Linear struct {
					Issue string `json:"issue"`
				} `json:"linear"`
			}
			if json.Unmarshal(r.Meta, &m) == nil && m.Linear.Issue != "" {
				out[strings.ToUpper(m.Linear.Issue)] = true
			}
		}
	}
	return out, nil
}

// IntakeStates is linear.intake (JOB-121): whether a linear_issue_no_run
// signal watches team, and the states they take (empty: any), so `linear
// create` can say when a new issue lands outside them. It reads config
// only.
func IntakeStates(cfg *config.Config, team string) (watched bool, states []string) {
	anyState := false
	for _, name := range slices.Sorted(maps.Keys(cfg.Triggers)) {
		for _, sig := range cfg.Triggers[name].When.Scan {
			if sig.Signal != config.SignalLinearIssueNoRun || !strings.EqualFold(sig.Team, team) {
				continue
			}
			watched = true
			ss := config.SplitList(sig.States)
			if len(ss) == 0 {
				anyState = true
			}
			for _, st := range ss {
				if !slices.ContainsFunc(states, func(x string) bool { return strings.EqualFold(x, st) }) {
					states = append(states, st)
				}
			}
		}
	}
	if anyState {
		states = nil
	}
	return watched, states
}
