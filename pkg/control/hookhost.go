package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/waitengine"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

// hookHost exposes controller policy and graph mechanics without exporting it.
type hookHost struct{ c *Controller }

func (c *Controller) hooks() *hooks.Engine {
	c.hooksOnce.Do(func() {
		c.hookEngine = hooks.New(c.store, hookHost{c}, c.root.Dir, c.now, c.log)
	})
	return c.hookEngine
}

func (h hookHost) CurrentHooks() *hooks.Runtime { return h.c.live.Load().runtime }

func (h hookHost) AdmitHook(ctx context.Context, runtime *hooks.Runtime, in hooks.Invocation) (string, error) {
	c := h.c
	if !c.hookReadSetCurrent(ctx, in.ReadSet) {
		return "superseded", nil
	}
	if in.Trigger == "" {
		run, err := c.store.Get(ctx, in.Run)
		if err != nil {
			return "", err
		}
		fields, err := domain.DecodeFields[domain.RunFields](run)
		if err != nil {
			return "", err
		}
		// Before evaluation, so a run2 run without its scope runs no code.
		return "", runScopeRecorded(fields.Automation)
	}
	tr, ok := runtime.Config.Triggers[in.Trigger]
	if !ok {
		return "superseded", nil
	}
	if in.Event != nil && in.Event.Source == "timer" {
		// The timer slot was accepted before the durable invocation existed.
		tr.When.Cron, tr.Cooldown = "", ""
	}
	if len(tr.When.Scan) > 0 {
		// evalScan admitted the finding; max_open is checked again on commit.
		return "", nil
	}
	st, err := c.triggerEngine().Eval(ctx, runtime.Config, in.Trigger, tr)
	if err != nil {
		return "", err
	}
	if !st.Holds {
		return "held", nil
	}
	return "", nil
}

func (h hookHost) CommitHook(ctx context.Context, runtime *hooks.Runtime, key string, in hooks.Invocation, journal hooks.JournalPlan) (hooks.Invocation, error) {
	c := h.c
	if !c.hookReadSetCurrent(ctx, in.ReadSet) {
		return in, domain.ErrConflict
	}
	lc := c.live.Load()
	if lc.runtime != runtime {
		return in, fmt.Errorf("%w: hook configuration changed", domain.ErrConflict)
	}
	plans, err := c.prepareHookRuns(ctx, lc, in)
	if err != nil {
		return in, err
	}
	c.runMu.Lock()
	defer c.runMu.Unlock()
	graph, err := c.planHookActions(ctx, lc, key, in, runtime.Programs[in.Entry], plans)
	if err != nil {
		return in, err
	}
	in.Plan = &graph
	for _, ticket := range graph.Tickets {
		in.Targets = append(in.Targets, ticket.ID)
	}
	for _, plan := range plans {
		if existing, err := store.OpenRun(ctx, c.store, plan.params.Workflow, plan.params.Key); err == nil && !slices.Contains(in.Targets, existing.ID) {
			in.Targets = append(in.Targets, existing.ID)
		}
	}
	if err := journal(in); err != nil {
		return in, err
	}
	return in, c.applyPlannedHook(ctx, key, in)
}

func (h hookHost) ApplyHookPlan(ctx context.Context, key string, in hooks.Invocation) error {
	c := h.c
	c.runMu.Lock()
	defer c.runMu.Unlock()
	return c.applyPlannedHook(ctx, key, in)
}

func (h hookHost) HookFailed(ctx context.Context, key string, in hooks.Invocation) {
	if in.Run == "" {
		return
	}
	c := h.c
	c.runMu.Lock()
	t, err := c.store.Get(ctx, in.Run)
	if err == nil && t.Status != domain.StatusClosed {
		f, decodeErr := domain.DecodeFields[domain.RunFields](t)
		if decodeErr == nil && f.State == domain.RunRunning {
			_, _ = c.setRunState(ctx, "controller", t, f, domain.RunNeedsHuman, "")
		}
	}
	c.runMu.Unlock()
	_, _ = store.AddComment(ctx, c.store, "controller", in.Run, "Automation invocation "+key+" failed: "+in.Error)
}

func (h hookHost) HookApplied(ctx context.Context, in hooks.Invocation) {
	if in.Run == "" {
		return
	}
	c := h.c
	c.wakeWaitsSubjects(waitengine.RunSubject(in.Run))
	if run, err := c.store.Get(ctx, in.Run); err == nil && run.Status == domain.StatusClosed {
		c.removeRunDir(run.ID)
	}
	c.runBatchEvent(ctx, in.Run)
}

// GitHubClient serves the github built-in: the shared client of the GitHub
// identity covering repo. An empty repo takes the first github-token
// identity, since an App token needs an owner. There is no fallback to the
// gh CLI.
func (h hookHost) GitHubClient(_ context.Context, repo string) (*github.Client, error) {
	cfg := h.c.config()
	if repo != "" {
		return h.c.githubForRepo(cfg, "", repo)
	}
	for _, n := range cfg.IdentityNames() {
		if id := cfg.Identities[n]; id.Kind == config.IdentityGitHubToken {
			return h.c.githubFor(n, id, ""), nil
		}
	}
	return nil, errors.New("no github-token identity answers a call without repo; pass repo=")
}

// LinearClient serves the linear built-in: the shared client of the Linear
// identity covering team. An empty team takes the only Linear identity.
func (h hookHost) LinearClient(_ context.Context, team string) (*linear.Client, error) {
	cfg := h.c.config()
	var names []string
	for _, n := range cfg.IdentityNames() {
		id := cfg.Identities[n]
		if config.IsLinear(id.Kind) && (team == "" || id.CoversTeam(team)) {
			names = append(names, n)
		}
	}
	switch {
	case len(names) == 0 && team == "":
		return nil, errors.New("no Linear identity configured")
	case len(names) == 0:
		return nil, fmt.Errorf("no Linear identity covers team %s", team)
	case team == "" && len(names) > 1:
		return nil, fmt.Errorf("%d Linear identities; pass team=", len(names))
	}
	return h.c.linearFor(names[0]), nil
}

func (c *Controller) hookReadSetCurrent(ctx context.Context, readSet map[string]int64) bool {
	for id, revision := range readSet {
		t, err := c.store.Get(ctx, id)
		if err != nil || t.Revision != revision {
			return false
		}
	}
	return true
}

func (c *Controller) applyPlannedHook(ctx context.Context, key string, in hooks.Invocation) error {
	var automation *domain.RunAutomation
	switch {
	case in.Run != "":
		run, err := c.store.Get(ctx, in.Run)
		if err != nil {
			return err
		}
		f, err := domain.DecodeFields[domain.RunFields](run)
		if err != nil {
			return err
		}
		automation = f.Automation
		if err := runScopeRecorded(f.Automation); err != nil {
			return err
		}
		if f.Automation != nil {
			for _, applied := range f.Automation.Applied {
				if applied == key {
					return nil
				}
			}
		}
	case len(in.Plan.Tickets) > 0:
		// Stable IDs were journaled before Apply. The graph is all-or-nothing.
		if _, err := c.store.Get(ctx, in.Plan.Tickets[0].ID); err == nil {
			return nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
	case len(in.Plan.Comments) > 0:
		// A comment-only plan has no ticket to find: look for its first
		// comment by this invocation's actor.
		first := in.Plan.Comments[0]
		comments, err := store.Comments(ctx, c.store, first.TicketID)
		if err != nil {
			return err
		}
		for _, cm := range comments {
			if cm.Author == "automation:"+key && cm.Body == first.Body {
				return nil
			}
		}
	}
	lc := c.live.Load()
	program := lc.hooks[in.Entry]
	if program == nil {
		return errors.New("planned hook removed from active configuration")
	}
	pinned := program.Digest() == in.Digest
	if in.Run != "" {
		if in.RunInput == nil {
			return errors.New("planned workflow hook has no accepted run input")
		}
		pinned = hooks.RunPinned(lc.hooks, lc.cfg, in.RunInput.Workflow, in.Digest)
	}
	if !pinned {
		return errors.New("planned hook manifest changed: restore the accepted configuration before applying actions")
	}
	// A run pin survives unrelated reloads, so recheck every pool the plan
	// can route a ticket to: its pool, a step's retry pool and placements.
	for _, ticket := range in.Plan.Tickets {
		pools := []string{ticket.Pool}
		if ticket.Kind == domain.KindStep {
			var f domain.StepFields
			if err := hooks.DecodeJSON(ticket.Fields, &f); err != nil {
				return err
			}
			pools = append(pools, f.RetryPool)
			if f.Placement != nil {
				for _, option := range f.Placement.Options {
					pools = append(pools, option.Pool)
				}
			}
		}
		for _, name := range pools {
			if name == "" {
				continue
			}
			if pool, ok := lc.cfg.Pools[name]; !ok || pool.Project != ticket.Project {
				return fmt.Errorf("planned pool %s changed: restore it before applying actions", name)
			}
		}
	}
	if err := checkPlannedTargets(lc.cfg, in, automation); err != nil {
		return fmt.Errorf("%w: restore the accepted configuration before applying actions", err)
	}
	for _, ticket := range in.Plan.Tickets {
		if ticket.Kind != domain.KindRun {
			continue
		}
		var fields domain.RunFields
		if err := hooks.DecodeJSON(ticket.Fields, &fields); err != nil {
			return err
		}
		if fields.Key == "" {
			continue
		}
		existing, err := store.OpenRun(ctx, c.store, fields.Workflow, fields.Key)
		if err == nil && existing.ID != ticket.ID {
			return errors.New("a run with the planned workflow and key was created before this batch committed")
		}
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
	}
	_, err := c.store.Apply(ctx, "automation:"+key, *in.Plan)
	if err != nil {
		return fmt.Errorf("apply hook actions: %w", err)
	}
	return nil
}

// ReadTickets serves tickets(), runs() and sessions(): a read of the whole
// store, as the controller sees it. Hooks are trusted code; writes still go
// through returned actions.
func (h hookHost) ReadTickets(ctx context.Context, q api.TicketQuery) ([]any, error) {
	f := store.Filter{
		Kind: domain.Kind(q.Kind), Project: q.Project, Pool: q.Pool, Label: q.Label,
		Workflow: q.Workflow, Key: q.Key, State: q.State, ParentID: q.Parent, Holder: q.Holder,
		Order: store.Order(q.Order), Limit: q.Limit,
	}
	for _, s := range q.Statuses {
		f.Statuses = append(f.Statuses, domain.Status(s))
	}
	for _, k := range slices.Sorted(maps.Keys(q.Meta)) {
		pack, key, _ := strings.Cut(k, ".")
		f.Meta = append(f.Meta, store.MetaMatch{Pack: pack, Key: key, Value: q.Meta[k]})
	}
	ts, err := h.c.store.List(ctx, f)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(ts)
	if err != nil {
		return nil, err
	}
	var list []any
	if err := hooks.DecodeJSON(data, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []any{}
	}
	return list, nil
}
