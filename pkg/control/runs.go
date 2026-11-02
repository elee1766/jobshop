package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Runs (docs/plan.md section 6 "Runs and workflows", 7.15). A run is a
// kind run ticket with no pool; its steps are kind step children, each
// blocked on the previous one. Failure counting and needs_human are
// separate (#90).

func (c *Controller) runStart(ctx context.Context, sc scope, p rpc.RunStartParams) (any, error) {
	return c.startRun(ctx, sc, p, nil)
}

// startRun is run start for the RPC and for run: triggers. extra is added
// to the run ticket's meta (the trigger link); steps do not get it.
func (c *Controller) startRun(ctx context.Context, sc scope, p rpc.RunStartParams, extra map[string]any) (any, error) {
	cfg := c.config()
	w, ok := cfg.Workflows[p.Workflow]
	if !ok {
		return nil, fmt.Errorf("%w: unknown workflow %q", store.ErrInvalid, p.Workflow)
	}
	// Steps live in their pools' projects; the run goes with the first.
	proj := cfg.Pools[w.Steps[0].Pool].Project
	if p.Project != "" && p.Project != proj {
		return nil, fmt.Errorf("%w: workflow %s runs in project %s", store.ErrInvalid, p.Workflow, proj)
	}
	if !sc.admin && sc.project != proj {
		return nil, forbidden("session is scoped to project %s", sc.project)
	}
	vars := map[string]any{}
	for k, v := range w.PackVars {
		vars[k] = v
	}
	for k, v := range p.Vars {
		if !slices.Contains(w.Vars, k) {
			return nil, fmt.Errorf("%w: workflow %s has no var %q (vars: %v)", store.ErrInvalid, p.Workflow, k, w.Vars)
		}
		vars[k] = v
	}
	for _, k := range w.Vars {
		if _, ok := vars[k]; !ok {
			return nil, fmt.Errorf("%w: workflow %s needs var %s (--var %s=...)", store.ErrInvalid, p.Workflow, k, k)
		}
	}
	meta, err := runMeta(w, vars)
	if err != nil {
		return nil, err
	}

	// Serialize starts so two starts with one key cannot both miss.
	c.runMu.Lock()
	defer c.runMu.Unlock()
	if p.Key != "" {
		t, err := c.store.FindOpenRunByKey(ctx, p.Workflow, p.Key)
		if err == nil {
			res, err := c.runResult(ctx, t)
			if err != nil {
				return nil, err
			}
			res.Existing = true
			return res, nil
		}
		if !errors.Is(err, store.ErrNotFound) {
			return nil, err
		}
	}

	runID, err := c.store.NewID(proj)
	if err != nil {
		return nil, err
	}
	pack := ""
	if s := cfg.Sources["workflows."+p.Workflow]; len(s) > 0 {
		pack = s[0]
	}
	rf, err := json.Marshal(store.RunFields{Workflow: p.Workflow, Pack: pack, Vars: p.Vars, Key: p.Key, State: store.RunRunning})
	if err != nil {
		return nil, err
	}
	title := "Run " + p.Workflow
	if p.Key != "" {
		title += " " + p.Key
	}
	runMetaJSON := meta
	if extra != nil {
		m := map[string]any{}
		if meta != nil {
			if err := json.Unmarshal(meta, &m); err != nil {
				return nil, err
			}
		}
		maps.Copy(m, extra)
		if runMetaJSON, err = json.Marshal(m); err != nil {
			return nil, err
		}
	}
	nts := []store.NewTicket{{ID: runID, Project: proj, Kind: store.KindRun, Title: title, Fields: rf, Meta: runMetaJSON}}
	var edges []store.Edge
	prev := ""
	for i, s := range w.Steps {
		title, err := config.FillTemplate(s.Title, vars)
		if err != nil {
			return nil, fmt.Errorf("%w: step %s title: %w", store.ErrInvalid, s.ID, err)
		}
		body, err := config.FillTemplate(s.Prompt, vars)
		if err != nil {
			return nil, fmt.Errorf("%w: step %s prompt: %w", store.ErrInvalid, s.ID, err)
		}
		id, err := c.store.NewID(cfg.Pools[s.Pool].Project)
		if err != nil {
			return nil, err
		}
		sf, err := json.Marshal(store.StepFields{Run: runID, Step: s.ID, Index: i, Retries: s.RetryLimit()})
		if err != nil {
			return nil, err
		}
		nts = append(nts, store.NewTicket{
			ID: id, Project: cfg.Pools[s.Pool].Project, Kind: store.KindStep, Title: title, Body: body,
			Pool: s.Pool, ParentID: runID, Fields: sf, Meta: meta,
		})
		if prev != "" {
			edges = append(edges, store.Edge{From: id, To: prev, Kind: store.EdgeBlocks})
		}
		prev = id
	}
	out, err := c.store.CreateGraph(ctx, sc.actor, nts, edges)
	if err != nil {
		return nil, err
	}
	c.log.Info("run started", "run", runID, "workflow", p.Workflow, "key", p.Key, "steps", len(w.Steps))
	return rpc.RunResult{Run: out[0], Steps: out[1:]}, nil
}

// runMeta is meta.github.issue when the workflow names an issue var. A
// numeric value is stored as a number.
func runMeta(w config.Workflow, vars map[string]any) (json.RawMessage, error) {
	if w.Issue == "" {
		return nil, nil
	}
	var issue any = fmt.Sprint(vars[w.Issue])
	if n, err := strconv.Atoi(issue.(string)); err == nil {
		issue = n
	}
	return json.Marshal(map[string]map[string]any{"github": {"issue": issue}})
}

func (c *Controller) runResult(ctx context.Context, run store.Ticket) (rpc.RunResult, error) {
	steps, err := c.store.ListChildren(ctx, run.ID)
	if err != nil {
		return rpc.RunResult{}, err
	}
	return rpc.RunResult{Run: run, Steps: steps}, nil
}

// getRun loads a run the caller may see.
func (c *Controller) getRun(ctx context.Context, sc scope, id string) (store.Ticket, store.RunFields, error) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return t, store.RunFields{}, err
	}
	if t.Kind != store.KindRun {
		return t, store.RunFields{}, fmt.Errorf("%w: %s is a %s, not a run", store.ErrInvalid, id, t.Kind)
	}
	if !sc.admin && t.Project != sc.project {
		return t, store.RunFields{}, forbidden("session is scoped to project %s", sc.project)
	}
	f, err := store.DecodeFields[store.RunFields](t)
	return t, f, err
}

func (c *Controller) runShow(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	t, _, err := c.getRun(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	return c.runResult(ctx, t)
}

func (c *Controller) runList(ctx context.Context, sc scope, p rpc.RunListParams) (any, error) {
	if !sc.admin {
		p.Project = sc.project
	}
	return c.store.ListRuns(ctx, p.Project, p.State, p.Limit)
}

// setRunState writes fields.state, closing the run with resolution when
// it is not empty.
func (c *Controller) setRunState(ctx context.Context, actor string, t store.Ticket, f store.RunFields, state, resolution string) (store.Ticket, error) {
	f.State = state
	b, err := json.Marshal(f)
	if err != nil {
		return t, err
	}
	t, err = c.store.Update(ctx, actor, t.ID, store.Update{Fields: b, IfRevision: t.Revision})
	if err != nil || resolution == "" {
		return t, err
	}
	return c.store.CloseTicket(ctx, actor, t.ID, nil, resolution)
}

func (c *Controller) runCancel(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	t, f, err := c.getRun(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: run %s", store.ErrClosed, t.ID)
	}
	steps, err := c.store.ListChildren(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		if s.Status == store.StatusClosed {
			continue
		}
		// Administrative close: a session holding the step loses it.
		if _, err := c.store.CloseTicket(ctx, sc.actor, s.ID, nil, store.RunCancelled); err != nil {
			return nil, err
		}
	}
	if _, err := c.setRunState(ctx, sc.actor, t, f, store.RunCancelled, store.RunCancelled); err != nil {
		return nil, err
	}
	c.log.Info("run cancelled", "run", t.ID, "by", sc.actor)
	t, err = c.store.Get(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	return c.runResult(ctx, t)
}

func (c *Controller) runResume(ctx context.Context, sc scope, p rpc.RunResumeParams) (any, error) {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	t, f, err := c.getRun(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: run %s", store.ErrClosed, t.ID)
	}
	steps, err := c.store.ListChildren(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	var step *store.Ticket
	for i := range steps {
		sf, err := store.DecodeFields[store.StepFields](steps[i])
		if err != nil {
			return nil, err
		}
		done := steps[i].Status == store.StatusClosed && steps[i].Resolution == "done"
		if (p.Step != "" && sf.Step == p.Step) || (p.Step == "" && !done) {
			step = &steps[i]
			break
		}
	}
	if step == nil {
		if p.Step != "" {
			return nil, fmt.Errorf("%w: run %s has no step %q", store.ErrInvalid, t.ID, p.Step)
		}
		return nil, fmt.Errorf("%w: run %s has no unfinished step", store.ErrInvalid, t.ID)
	}
	sf, err := store.DecodeFields[store.StepFields](*step)
	if err != nil {
		return nil, err
	}
	sf.Failures = 0
	b, err := json.Marshal(sf)
	if err != nil {
		return nil, err
	}
	if step.Status == store.StatusClosed {
		if _, err := c.store.Reopen(ctx, sc.actor, step.ID); err != nil {
			return nil, err
		}
	}
	if _, err := c.store.Update(ctx, sc.actor, step.ID, store.Update{Fields: b}); err != nil {
		return nil, err
	}
	if f.State != store.RunRunning {
		if _, err := c.setRunState(ctx, sc.actor, t, f, store.RunRunning, ""); err != nil {
			return nil, err
		}
	}
	_, _ = c.store.AddComment(ctx, sc.actor, step.ID, fmt.Sprintf("Resumed by %s; failures reset.", sc.actor))
	c.log.Info("run resumed", "run", t.ID, "step", sf.Step, "by", sc.actor)
	t, err = c.store.Get(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	return c.runResult(ctx, t)
}

// completeRun closes a running run done once every step is closed done.
// Called after a step closes and on each tick (a close by another path,
// or a controller restart between the two writes, is caught there).
func (c *Controller) completeRun(ctx context.Context, runID string) {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	t, err := c.store.Get(ctx, runID)
	if err != nil || t.Kind != store.KindRun || t.Status == store.StatusClosed {
		return
	}
	f, err := store.DecodeFields[store.RunFields](t)
	if err != nil || f.State != store.RunRunning {
		return
	}
	steps, err := c.store.ListChildren(ctx, runID)
	if err != nil || len(steps) == 0 {
		return
	}
	for _, s := range steps {
		if s.Status != store.StatusClosed || s.Resolution != "done" {
			return
		}
	}
	if _, err := c.setRunState(ctx, "controller", t, f, store.RunDone, "done"); err != nil {
		c.log.Error("complete run", "run", runID, "err", err)
		return
	}
	c.log.Info("run done", "run", runID, "workflow", f.Workflow)
}

// sweepRuns completes or escalates running runs and repairs escalations
// for needs_human runs, then sets the needs_human gauge.
func (c *Controller) sweepRuns() {
	defer c.recordRunGauges(c.ctx)
	for _, state := range []string{store.RunRunning, store.RunNeedsHuman} {
		runs, err := c.store.ListRuns(c.ctx, "", state, 0)
		if err != nil {
			return
		}
		for _, r := range runs {
			if r.Status != store.StatusClosed {
				c.checkRun(c.ctx, r.ID)
			}
		}
	}
}
