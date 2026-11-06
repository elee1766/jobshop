package control

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// linearIntakePages bounds one Linear read (250 issues a page).
const linearIntakePages = 8

// linearIntake caches the matching issues of each linear_issue_no_run
// entry between polls, so a 30s tick does not read Linear every time.
// Which of them have an open run is decided from the store every tick.
type linearIntake struct {
	mu    sync.Mutex
	cache map[string]linearIntakeRead
}

type linearIntakeRead struct {
	at     time.Time
	issues []linear.Summary
}

// stallLinearIssues is the linear_issue_no_run signal (#287): open issues
// of sig.Team in sig.States with one of sig.Labels and none of
// sig.SkipLabels, minus those an open run carries in meta.linear.issue.
// The subject is the issue key. A Linear read error is logged and the
// signal finds nothing, so an outage does not stop the other signals.
func (c *Controller) stallLinearIssues(ctx context.Context, sig config.StallSignal) ([]stall, error) {
	if sig.Team == "" {
		return nil, nil
	}
	issues, err := c.linearIntakeIssues(ctx, sig)
	if err != nil {
		c.log.WarnContext(ctx, "linear_issue_no_run", "team", sig.Team, "err", err)
		return nil, nil
	}
	busy, err := c.linearIssuesWithOpenRun(ctx)
	if err != nil {
		return nil, err
	}
	grace := linearGrace(sig)
	now := c.now()
	var out []stall
	for _, is := range issues {
		if busy[is.Identifier] {
			continue
		}
		if grace > 0 && !is.UpdatedAt.IsZero() && now.Before(is.UpdatedAt.Add(grace)) {
			continue
		}
		out = append(out, stall{config.StallLinearIssueNoRun, is.Identifier, fmt.Sprintf("%s, no run: %s", is.State.Name, is.Title)})
	}
	return out, nil
}

// linearGrace is sig.Grace; empty is off.
func linearGrace(sig config.StallSignal) time.Duration {
	d, _ := time.ParseDuration(sig.Grace)
	return d
}

// linearIssueHasOpenRun rechecks the store just before a stall-check run
// starts: the planner may have started the issue's run during the scan.
func (c *Controller) linearIssueHasOpenRun(ctx context.Context, key string) (bool, error) {
	busy, err := c.linearIssuesWithOpenRun(ctx)
	if err != nil {
		return false, err
	}
	return busy[strings.ToUpper(key)], nil
}

// linearIntakeIssues returns the matching issues, read from Linear at most
// once per poll interval.
func (c *Controller) linearIntakeIssues(ctx context.Context, sig config.StallSignal) ([]linear.Summary, error) {
	key := strings.Join([]string{sig.Team, sig.States, sig.Labels, sig.SkipLabels}, "|")
	poll := config.Duration(sig.Poll, config.DefaultLinearPoll)
	now := c.now()
	c.linIntake.mu.Lock()
	defer c.linIntake.mu.Unlock()
	if r, ok := c.linIntake.cache[key]; ok && now.Before(r.at.Add(poll)) {
		return r.issues, nil
	}
	issues, err := c.readLinearIntake(ctx, sig)
	if err != nil {
		return nil, err
	}
	if c.linIntake.cache == nil {
		c.linIntake.cache = map[string]linearIntakeRead{}
	}
	c.linIntake.cache[key] = linearIntakeRead{at: now, issues: issues}
	return issues, nil
}

// readLinearIntake lists the team's issues as the identity covering the
// team and keeps the open ones that match. Filtering is done here: the
// list filter takes one state and one label, the vars take several.
func (c *Controller) readLinearIntake(ctx context.Context, sig config.StallSignal) ([]linear.Summary, error) {
	cfg := c.config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), sig.Team)
	if !ok {
		return nil, fmt.Errorf("no identity covers team %s", sig.Team)
	}
	if c.minter == nil {
		return nil, fmt.Errorf("no identity minter")
	}
	tok, err := c.minter.Token(ctx, name, cfg.Identities[name], sig.Team)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cl := c.newLinear(c.linearURL, name, cfg.Identities[name], sig.Team, tok.Token)
	states, labels, skip := config.SplitList(sig.States), config.SplitList(sig.Labels), config.SplitList(sig.SkipLabels)
	f := linear.Filter{Team: sig.Team, Limit: 250}
	var out []linear.Summary
	for range linearIntakePages {
		p, err := cl.List(ctx, f)
		if err != nil {
			return nil, err
		}
		for _, is := range p.Issues {
			if linearIntakeMatch(is, states, labels, skip) && !c.linearIntakeBlocked(ctx, cl, cfg, is.Identifier) {
				out = append(out, is)
			}
		}
		if !p.HasNext {
			return out, nil
		}
		f.After = p.EndCursor
	}
	c.log.WarnContext(ctx, "linear_issue_no_run: team has more issues than one read covers", "team", sig.Team, "pages", linearIntakePages)
	return out, nil
}

// linearIntakeBlocked reports whether a matching issue waits on another:
// an open Linear blocker or an open GitHub issue named by a "Depends on"
// line (JOB-68). The list read carries neither, so each match is read in
// full, once per poll. A read error keeps the issue: intake then behaves
// as before this check.
func (c *Controller) linearIntakeBlocked(ctx context.Context, cl *linear.Client, cfg *config.Config, key string) bool {
	is, err := cl.GetContext(ctx, key, false)
	if err != nil {
		c.log.WarnContext(ctx, "linear_issue_no_run: read issue", "issue", key, "err", err)
		return false
	}
	if len(is.OpenBlockers()) > 0 {
		return true
	}
	if err := is.CheckDeps(ctx, linearRepoOwner(cfg), c.githubIssueOpen); err != nil {
		c.log.WarnContext(ctx, "linear_issue_no_run: dependency", "issue", key, "err", err)
		return false
	}
	return len(is.OpenDeps) > 0
}

// linearRepoOwner is the GitHub owner of the first (by name) Linear
// tracked workflow's repo, for a bare "Depends on repo#n".
func linearRepoOwner(cfg *config.Config) string {
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

// githubIssueOpen reads a GitHub issue's state as the identity covering
// its repo. Tests replace it with c.ghIssueOpen.
func (c *Controller) githubIssueOpen(ctx context.Context, d linear.GitHubDep) (bool, error) {
	if c.ghIssueOpen != nil {
		return c.ghIssueOpen(ctx, d)
	}
	base, tok, err := c.pollerAuth(ctx, c.config(), "", d.Repo)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(telemetry.WithRoute(ctx, "/repos/{repo}/issues/{n}"), http.MethodGet, fmt.Sprintf("%s/repos/%s/issues/%d", base, d.Repo, d.Number), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := githubHTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("github issue %s: %s", d, resp.Status)
	}
	var v struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&v); err != nil {
		return false, err
	}
	return v.State == "open", nil
}

// linearIntakeMatch: open, in one of states (any when empty), with one
// of labels (any when empty), with none of skip. Names compare
// case-insensitively.
func linearIntakeMatch(is linear.Summary, states, labels, skip []string) bool {
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

// linearIssuesWithOpenRun is meta.linear.issue of every open run.
func (c *Controller) linearIssuesWithOpenRun(ctx context.Context) (map[string]bool, error) {
	out := map[string]bool{}
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, Status: st})
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
