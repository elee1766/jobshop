package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type hookRunPlan struct {
	params rpc.RunStartParams
	plan   runPlan
}

func (c *Controller) prepareHookRuns(ctx context.Context, lc *liveConfig, in hookInvocation) (map[int]hookRunPlan, error) {
	plans := map[int]hookRunPlan{}
	owner := "triggers." + in.Trigger
	if in.Trigger == "" {
		owner = "workflows." + in.RunInput.Workflow
	}
	for i, action := range in.Actions {
		if action.Kind != "start_run" {
			continue
		}
		workflow, err := lc.cfg.ResolveHookReference(owner, "workflow", action.Workflow)
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
func (c *Controller) planHookActions(ctx context.Context, lc *liveConfig, key string, in hookInvocation, program *hookProgram, plans map[int]hookRunPlan) (store.Graph, error) {
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
				extra[triggerPack] = map[string]any{"name": in.Trigger, "key": in.Trigger + "/" + key}
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
			pool, err := lc.cfg.ResolveHookReference("workflows."+in.RunInput.Workflow, "pool", action.Pool)
			if err != nil {
				return g, err
			}
			if lc.cfg.Pools[pool].Project != run.Project {
				return g, fmt.Errorf("pool %s is outside run project %s", pool, run.Project)
			}
			retryPool := ""
			if action.RetryPool != "" {
				retryPool, err = lc.cfg.ResolveHookReference("workflows."+in.RunInput.Workflow, "pool", action.RetryPool)
				if err != nil {
					return g, err
				}
				if lc.cfg.Pools[retryPool].Project != run.Project {
					return g, errors.New("retry_pool is outside the run project")
				}
			}
			prompt, err := hookPrompt(program, action)
			if err != nil {
				return g, err
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
			g.Tickets = append(g.Tickets, store.NewTicket{ID: id, Project: run.Project, Kind: store.KindStep, Title: title, Body: prompt, Pool: pool, ParentID: run.ID, Meta: provenance, Fields: data})
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
	if in.Trigger != "" && lc.cfg.Triggers[in.Trigger].MaxOpen > 0 {
		n := 0
		for _, nt := range g.Tickets {
			if nt.Kind == store.KindRun {
				n++
			}
		}
		for _, status := range []store.Status{store.StatusOpen, store.StatusInProgress} {
			tickets, err := c.store.List(ctx, store.Filter{MetaPack: triggerPack, MetaKey: "name", MetaValue: in.Trigger, Status: status, Limit: store.NoLimit})
			if err != nil {
				return g, err
			}
			n += len(tickets)
		}
		if n > lc.cfg.Triggers[in.Trigger].MaxOpen {
			return g, errors.New("returned run batch exceeds trigger max_open")
		}
	}
	return g, nil
}

func hookPrompt(program *hookProgram, action automation.Action) (string, error) {
	if action.PromptFile == "" {
		if strings.TrimSpace(action.Prompt) == "" {
			return "", errors.New("step prompt must not be blank")
		}
		return action.Prompt, nil
	}
	root, err := os.OpenRoot(program.root)
	if err != nil {
		return "", err
	}
	defer func() { _ = root.Close() }()
	file, err := root.Open(action.PromptFile)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("prompt_file must be a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return "", err
	}
	if len(data) > 1<<20 {
		return "", errors.New("prompt_file exceeds 1 MiB")
	}
	return string(data), nil
}
