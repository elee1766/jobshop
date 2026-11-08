package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type hookRunPlan struct {
	params rpc.RunStartParams
	plan   runPlan
}

// hookReference resolves a callback's pool or workflow name in the live
// configuration and, for a workflow callback, within the scope its run
// accepted at start.
func hookReference(cfg *config.Config, in hooks.Invocation, a *store.RunAutomation, kind, name string) (string, error) {
	if in.Trigger != "" {
		return cfg.ResolveHookReference("triggers."+in.Trigger, kind, name)
	}
	if in.RunInput == nil {
		return "", fmt.Errorf("%w: run has no automation", store.ErrConflict)
	}
	if err := runScopeRecorded(a); err != nil {
		return "", err
	}
	return cfg.ResolvePinnedHookReference("workflows."+in.RunInput.Workflow, kind, name, a.Scope)
}

// runScopeRecorded fails closed for a run2 pin whose run lost the scope it
// accepted at start; every callback path checks it, even with no references.
func runScopeRecorded(a *store.RunAutomation) error {
	if a == nil {
		return fmt.Errorf("%w: run has no automation", store.ErrConflict)
	}
	if hooks.ScopedRunPin(a.Digest) && a.Scope == nil {
		return errors.New("run has no accepted reference scope")
	}
	return nil
}

// checkPlannedTargets re-resolves each planned ticket's action before a
// journaled graph is applied, as a reload since planning can narrow scope.
func checkPlannedTargets(cfg *config.Config, in hooks.Invocation, a *store.RunAutomation) error {
	for _, ticket := range in.Plan.Tickets {
		var meta struct {
			Automation *struct {
				Action int `json:"action"`
			} `json:"automation"`
		}
		if len(ticket.Meta) > 0 {
			if err := hooks.DecodeJSON(ticket.Meta, &meta); err != nil {
				return err
			}
		}
		if meta.Automation == nil || meta.Automation.Action < 0 || meta.Automation.Action >= len(in.Actions) {
			if ticket.Kind == store.KindRun {
				return fmt.Errorf("planned run %s has no action", ticket.ID)
			}
			continue // a step of a planned run, checked with that run
		}
		action := in.Actions[meta.Automation.Action]
		want := map[string]string{}
		switch {
		case ticket.Kind == store.KindRun && action.Kind == "start_run":
			var f store.RunFields
			if err := hooks.DecodeJSON(ticket.Fields, &f); err != nil {
				return err
			}
			w := cfg.Workflows[f.Workflow]
			if (f.Automation != nil) != (w.OnStart != "") || f.Automation != nil && cfg.HookRunPins[f.Workflow] != f.Automation.Digest {
				return fmt.Errorf("planned workflow %s changed", f.Workflow)
			}
			want["workflow:"+action.Workflow] = f.Workflow
		case ticket.Kind == store.KindStep && action.Kind == "start_step" && ticket.ParentID == in.Run:
			var f store.StepFields
			if err := hooks.DecodeJSON(ticket.Fields, &f); err != nil {
				return err
			}
			want["pool:"+action.Pool] = ticket.Pool
			if f.Type == store.StepScript {
				want["pool:"+action.Pool] = f.ScriptPool
			}
			if action.RetryPool != "" || f.RetryPool != "" {
				want["pool:"+action.RetryPool] = f.RetryPool
			}
		case ticket.Kind == store.KindTask && action.Kind == "create_ticket" && action.Pool != "":
			want["pool:"+action.Pool] = ticket.Pool
		default:
			continue // a step of a planned static run, from buildRun, or a task without a pool
		}
		for ref, planned := range want {
			kind, name, _ := strings.Cut(ref, ":")
			resolved, err := hookReference(cfg, in, a, kind, name)
			if err != nil {
				return fmt.Errorf("planned %s %q: %w", kind, name, err)
			}
			if resolved != planned {
				return fmt.Errorf("planned %s %q now resolves to %s, not %s", kind, name, resolved, planned)
			}
		}
	}
	return nil
}

func (c *Controller) prepareHookRuns(ctx context.Context, lc *liveConfig, in hooks.Invocation) (map[int]hookRunPlan, error) {
	plans := map[int]hookRunPlan{}
	var automation *store.RunAutomation
	if in.Trigger == "" {
		run, err := c.store.Get(ctx, in.Run)
		if err != nil {
			return nil, err
		}
		fields, err := store.DecodeFields[store.RunFields](run)
		if err != nil {
			return nil, err
		}
		automation = fields.Automation
	}
	for i, action := range in.Actions {
		if action.Kind != "start_run" {
			continue
		}
		workflow, err := hookReference(lc.cfg, in, automation, "workflow", action.Workflow)
		if err != nil {
			return nil, err
		}
		p := rpc.RunStartParams{Workflow: workflow, Key: action.Key, Inputs: action.Inputs, Vars: map[string]string{}}
		for _, name := range lc.cfg.Workflows[workflow].Vars {
			if value, ok := action.Inputs[name]; ok {
				switch value := value.(type) {
				case string:
					p.Vars[name] = value
				case json.Number, float64, bool, int, int64:
					p.Vars[name] = fmt.Sprint(value)
				default:
					return nil, fmt.Errorf("workflow var %s requires a scalar input", name)
				}
			}
		}
		sc := scope{admin: in.Run == "", actor: "automation"}
		if in.RunInput != nil {
			sc.project = in.RunInput.Project
		}
		plan, err := c.planRun(ctx, lc.cfg, sc, p)
		if err != nil {
			return nil, err
		}
		plans[i] = hookRunPlan{params: p, plan: plan}
	}
	return plans, nil
}

// planHookActions validates the complete batch before the first graph write.
// The caller holds runMu only for local reads and the graph commit.
func (c *Controller) planHookActions(ctx context.Context, lc *liveConfig, key string, in hooks.Invocation, program *hooks.Program, plans map[int]hookRunPlan) (store.Graph, error) {
	g := store.Graph{Revisions: in.ReadSet, Children: map[string]int64{}, Fields: map[string]json.RawMessage{}, Close: map[string]string{}}
	if c.live.Load() != lc {
		return g, fmt.Errorf("%w: hook configuration changed", store.ErrConflict)
	}
	var run store.Ticket
	var fields store.RunFields
	steps := map[string]store.Ticket{}
	ids := map[string]string{}
	previous, previousIndex := "", -1
	if in.Run != "" {
		var err error
		run, err = c.store.Get(ctx, in.Run)
		if err != nil {
			return g, err
		}
		fields, err = store.DecodeFields[store.RunFields](run)
		if err != nil {
			return g, err
		}
		if run.Status == store.StatusClosed || fields.State != store.RunRunning || fields.Automation == nil {
			return g, fmt.Errorf("%w: run no longer running", store.ErrConflict)
		}
		if err := runScopeRecorded(fields.Automation); err != nil {
			return g, err
		}
		children, err := store.Children(ctx, c.store, in.Run)
		if err != nil {
			return g, err
		}
		g.Children[in.Run] = int64(len(children))
		if !c.hookReadSetCurrent(ctx, in.ReadSet) || len(children)+1 != len(in.ReadSet) {
			return g, fmt.Errorf("%w: hook input changed", store.ErrConflict)
		}
		for _, step := range children {
			if step.Kind != store.KindStep {
				continue
			}
			f, err := store.DecodeFields[store.StepFields](step)
			if err != nil {
				return g, err
			}
			steps[f.Step], ids[f.Step] = step, step.ID
			if f.Index > previousIndex {
				previous, previousIndex = f.Step, f.Index
			}
		}
	}
	finished := false
	newSteps := 0
	started := map[string]string{}
	created := map[string]bool{}
	for i, action := range in.Actions {
		switch action.Kind {
		case "start_run":
			rp := plans[i]
			dedup := rp.params.Workflow + "/" + rp.params.Key
			if _, ok := started[dedup]; ok {
				continue
			}
			existing, err := store.OpenRun(ctx, c.store, rp.params.Workflow, rp.params.Key)
			if err == nil {
				started[dedup] = existing.ID
				continue
			}
			if !errors.Is(err, store.ErrNotFound) {
				return g, err
			}
			extra := map[string]any{"automation": map[string]any{"invocation": key, "action": i}}
			if in.Trigger != "" {
				// trigger/<run key>, as a static run: action keys dedup
				// across invocations and finding history reads them.
				m := map[string]any{"name": in.Trigger, "key": in.Trigger + "/" + rp.params.Key}
				if in.Event != nil && in.Event.Source == hooks.ScanSource {
					// The #704 hold and max_runs read the finding id on its runs.
					m[findingMetaKey] = in.Event.Delivery
				}
				extra[triggerPack] = m
			}
			graph, err := c.buildRun(lc.cfg, rp.plan, rp.params, extra, "")
			if err != nil {
				return g, err
			}
			if action.Title != "" {
				graph.Tickets[0].Title = action.Title
			}
			started[dedup] = graph.Tickets[0].ID
			g.Tickets = append(g.Tickets, graph.Tickets...)
			g.Edges = append(g.Edges, graph.Edges...)
		case "start_step":
			if in.Run == "" {
				return g, errors.New("start_step requires a workflow invocation")
			}
			if finished {
				return g, errors.New("cannot start a step after finish_run")
			}
			if !config.OutcomePattern.MatchString(action.ID) {
				return g, fmt.Errorf("invalid step id %q", action.ID)
			}
			if ids[action.ID] != "" {
				return g, fmt.Errorf("step %q already exists", action.ID)
			}
			pool, err := hookReference(lc.cfg, in, fields.Automation, "pool", action.Pool)
			if err != nil {
				return g, err
			}
			if lc.cfg.Pools[pool].Project != run.Project {
				return g, fmt.Errorf("pool %s is outside run project %s", pool, run.Project)
			}
			retryPool := ""
			if action.RetryPool != "" {
				retryPool, err = hookReference(lc.cfg, in, fields.Automation, "pool", action.RetryPool)
				if err != nil {
					return g, err
				}
				if lc.cfg.Pools[retryPool].Project != run.Project {
					return g, errors.New("retry_pool is outside the run project")
				}
			}
			prompt := ""
			if action.Script == "" {
				prompt, err = program.Prompt(action)
				if err != nil {
					return g, err
				}
			}
			if action.PromptFile != "" {
				vars := maps.Clone(fields.Inputs)
				if vars == nil {
					vars = map[string]any{}
				}
				maps.Copy(vars, action.Inputs)
				workflow := lc.cfg.Workflows[fields.Workflow]
				prompt, err = workflow.Render(prompt, vars)
				if err != nil {
					return g, err
				}
			}
			retries := config.DefaultStepRetries
			if action.RetriesSet {
				retries = action.Retries
			}
			if err := checkHookSchema(action.OutputSchema, nil, false); err != nil {
				return g, err
			}
			id, err := c.store.NewID(run.Project)
			if err != nil {
				return g, err
			}
			title := action.Title
			if title == "" {
				title = action.ID
			}
			f := store.StepFields{Run: run.ID, Step: action.ID, Index: len(steps) + newSteps, Retries: retries, RetryPool: retryPool, Needs: action.Needs, Inputs: action.Inputs, OutputSchema: action.OutputSchema}
			ticketPool := pool
			if action.Script != "" {
				if action.Worktree && lc.cfg.Pools[pool].Repo == "" {
					return g, fmt.Errorf("step %s: worktree needs a pool with a repo, %s has none", action.ID, pool)
				}
				// No ticket pool: the controller runs it, no session claims it.
				f.Type, f.Script, f.ScriptHook, f.ScriptPool = store.StepScript, action.Script, in.Entry, pool
				f.Worktree, f.Timeout, ticketPool = action.Worktree, action.Timeout, ""
			}
			if slices.Contains(action.Needs, previous) {
				f.OnRequestChanges = previous
			}
			data, err := json.Marshal(f)
			if err != nil {
				return g, err
			}
			meta := map[string]any{}
			if len(run.Meta) > 0 {
				if err := json.Unmarshal(run.Meta, &meta); err != nil {
					return g, err
				}
			}
			meta["automation"] = map[string]any{"invocation": key, "action": i}
			provenance, err := json.Marshal(meta)
			if err != nil {
				return g, err
			}
			g.Tickets = append(g.Tickets, store.NewTicket{ID: id, Project: run.Project, Kind: store.KindStep, Title: title, Body: prompt, Pool: ticketPool, ParentID: run.ID, Meta: provenance, Fields: data})
			for _, need := range action.Needs {
				target := ids[need]
				if target == "" {
					return g, fmt.Errorf("step %s needs unknown or later step %s", action.ID, need)
				}
				g.Edges = append(g.Edges, store.Edge{From: id, To: target, Kind: store.EdgeBlocks})
			}
			if len(action.Needs) == 0 {
				for _, after := range fields.Automation.After {
					g.Edges = append(g.Edges, store.Edge{From: id, To: after, Kind: store.EdgeBlocks})
				}
			}
			ids[action.ID] = id
			previous = action.ID
			newSteps++
		case "finish_run":
			if in.Run == "" {
				return g, errors.New("finish_run requires a workflow invocation")
			}
			if finished || newSteps > 0 {
				return g, errors.New("finish_run cannot be combined with start_step or another finish_run")
			}
			for _, step := range steps {
				if !store.StepFinished(step) {
					return g, fmt.Errorf("cannot finish run with unfinished step %s", step.ID)
				}
			}
			if !config.OutcomePattern.MatchString(action.Outcome) {
				return g, errors.New("invalid finish_run outcome")
			}
			fields.State = store.RunDone
			fields.Automation.Output = action.Output
			g.Close[run.ID] = action.Outcome
			finished = true
		case "create_ticket":
			nt, skip, err := c.planHookTicket(ctx, lc.cfg, key, i, in, fields.Automation, action, created)
			if err != nil {
				return g, err
			}
			if !skip {
				created[action.Key] = true
				g.Tickets = append(g.Tickets, nt)
			}
		case "comment":
			if _, err := c.store.Get(ctx, action.Ticket); err != nil {
				return g, fmt.Errorf("comment on %s: %w", action.Ticket, err)
			}
			g.Comments = append(g.Comments, store.GraphComment{TicketID: action.Ticket, Body: action.Body})
		default:
			return g, fmt.Errorf("unsupported action %q", action.Kind)
		}
	}
	if in.Run != "" {
		a := fields.Automation
		a.Version++
		a.Applied = append(a.Applied, key)
		if a.Completed == nil {
			a.Completed = map[string]int64{}
		}
		if in.Step != "" {
			a.Completed[in.Step] = in.StepRevision
		}
		data, err := json.Marshal(fields)
		if err != nil {
			return g, err
		}
		g.Fields[run.ID] = data
	}
	if tr := lc.cfg.Triggers[in.Trigger]; in.Trigger != "" && (tr.MaxOpen > 0 || len(tr.When.Scan) > 0) {
		limit := tr.MaxOpen
		if len(tr.When.Scan) > 0 {
			limit = tr.ScanMaxOpen()
		}
		n := 0
		for _, nt := range g.Tickets {
			if nt.Kind == store.KindRun || nt.Kind == store.KindTask {
				n++
			}
		}
		open, err := c.triggerOpen(ctx, in.Trigger)
		if err != nil {
			return g, err
		}
		n += open
		if n > limit {
			return g, errors.New("returned batch exceeds trigger max_open")
		}
	}
	return g, nil
}

// triggerOpen counts the open runs and tasks a trigger filed. Steps of its
// runs carry no meta.trigger, but are excluded by kind in case one does.
func (c *Controller) triggerOpen(ctx context.Context, name string) (int, error) {
	ts, err := c.store.List(ctx, store.Filter{
		MetaPack: triggerPack, MetaKey: "name", MetaValue: name,
		Kinds:    []store.Kind{store.KindRun, store.KindTask},
		Statuses: []store.Status{store.StatusOpen, store.StatusInProgress}, Limit: store.NoLimit,
	})
	return len(ts), err
}

// planHookTicket builds a create_ticket action's task. Its key dedups per
// trigger or workflow: an open ticket with the same key, or one earlier in
// the batch, makes the action a no-op (skip). The project is the pool's,
// else the parent's, else the run's. A trigger's ticket carries
// meta.trigger, so it holds the trigger like its runs (max_open).
func (c *Controller) planHookTicket(ctx context.Context, cfg *config.Config, key string, i int, in hooks.Invocation, a *store.RunAutomation, action automation.Action, created map[string]bool) (store.NewTicket, bool, error) {
	owner := "workflows."
	if in.Trigger != "" {
		owner = "triggers." + in.Trigger
	} else if in.RunInput != nil {
		owner += in.RunInput.Workflow
	}
	dedup := owner + "/" + action.Key
	if created[action.Key] {
		return store.NewTicket{}, true, nil
	}
	open, err := c.store.List(ctx, store.Filter{
		Meta:     []store.MetaMatch{{Pack: "automation", Key: "key", Value: dedup}},
		Statuses: []store.Status{store.StatusOpen, store.StatusInProgress}, Limit: 1,
	})
	if err != nil {
		return store.NewTicket{}, false, err
	}
	if len(open) > 0 {
		return store.NewTicket{}, true, nil
	}
	nt := store.NewTicket{Kind: store.KindTask, Title: action.Title, Body: action.Body, Priority: action.Priority, Labels: action.Labels, ParentID: action.Parent}
	if action.Pool != "" {
		pool, err := hookReference(cfg, in, a, "pool", action.Pool)
		if err != nil {
			return store.NewTicket{}, false, err
		}
		nt.Pool, nt.Project = pool, cfg.Pools[pool].Project
	}
	if action.Parent != "" {
		parent, err := c.store.Get(ctx, action.Parent)
		if err != nil {
			return store.NewTicket{}, false, fmt.Errorf("create_ticket parent %s: %w", action.Parent, err)
		}
		if nt.Project == "" {
			nt.Project = parent.Project
		}
	}
	if nt.Project == "" && in.RunInput != nil {
		nt.Project = in.RunInput.Project
	}
	if nt.Project == "" {
		return store.NewTicket{}, false, errors.New("create_ticket needs a pool or parent to pick its project")
	}
	m := map[string]any{"automation": map[string]any{"invocation": key, "action": i, "key": dedup}}
	if in.Trigger != "" {
		m[triggerPack] = map[string]any{"name": in.Trigger, "key": in.Trigger + "/" + action.Key}
	}
	if nt.Meta, err = json.Marshal(m); err != nil {
		return store.NewTicket{}, false, err
	}
	if nt.ID, err = c.store.NewID(nt.Project); err != nil {
		return store.NewTicket{}, false, err
	}
	return nt, false, nil
}
