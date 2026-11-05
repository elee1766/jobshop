package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Runs (docs/plan.md section 6 "Runs and workflows", docs/notes/89-runs.md). A run is a
// kind run ticket with no pool; its steps are kind step children, each
// blocked on the previous one. Failure counting and needs_human are
// separate (#90).

func (c *Controller) runStart(ctx context.Context, sc scope, p rpc.RunStartParams) (any, error) {
	return c.startRun(ctx, sc, p, nil)
}

// startRun is run start for the RPC and for run: triggers. extra is added
// to the run ticket's meta (the trigger link); steps do not get it.
func (c *Controller) startRun(ctx context.Context, sc scope, p rpc.RunStartParams, extra map[string]any) (any, error) {
	plan, err := c.planRun(ctx, sc, p)
	if err != nil {
		return nil, err
	}
	// Serialize starts so two starts with one key cannot both miss.
	c.runMu.Lock()
	defer c.runMu.Unlock()
	if p.Key != "" {
		t, err := store.OpenRun(ctx, c.store, p.Workflow, p.Key)
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
	g, err := c.buildRun(plan, p, extra, "")
	if err != nil {
		return nil, err
	}
	out, err := c.store.CreateGraph(ctx, sc.actor, g.Tickets, g.Edges)
	if err != nil {
		return nil, err
	}
	c.log.Info("run started", "run", out[0].ID, "workflow", p.Workflow, "key", p.Key, "steps", len(out)-1)
	return rpc.RunResult{Run: out[0], Steps: out[1:]}, nil
}

// runPlan is a validated run start: workflow, project, filled vars and meta.
type runPlan struct {
	w    config.Workflow
	proj string
	vars map[string]any
	meta json.RawMessage
	// placements are the resolved placements by step index (#409).
	placements map[int]*store.StepPlacement
}

// planRun checks p against the config and the caller's scope.
func (c *Controller) planRun(ctx context.Context, sc scope, p rpc.RunStartParams) (runPlan, error) {
	cfg := c.config()
	w, ok := cfg.Workflows[p.Workflow]
	if !ok {
		return runPlan{}, fmt.Errorf("%w: unknown workflow %q", store.ErrInvalid, p.Workflow)
	}
	vars := map[string]any{}
	for k, v := range w.PackVars {
		vars[k] = v
	}
	for k, v := range p.Vars {
		if !slices.Contains(w.Vars, k) {
			return runPlan{}, fmt.Errorf("%w: workflow %s has no var %q (vars: %v)", store.ErrInvalid, p.Workflow, k, w.Vars)
		}
		// An empty issue or PR var renders "Implement #" with no issue to
		// work, and its key (issue-) joins every later empty start to one
		// broken run (#496). Other vars may be empty (a stall's detail).
		if (k == w.Issue || k == w.PR) && strings.TrimSpace(v) == "" {
			return runPlan{}, fmt.Errorf("%w: workflow %s var %s is empty", store.ErrInvalid, p.Workflow, k)
		}
		vars[k] = v
	}
	for _, k := range w.Vars {
		if _, ok := vars[k]; !ok {
			return runPlan{}, fmt.Errorf("%w: workflow %s needs var %s (--var %s=...)", store.ErrInvalid, p.Workflow, k, k)
		}
	}
	// A step pool may name a run var (#384): resolve every one before
	// any ticket exists, so a bad pool fails the start.
	w.Steps = slices.Clone(w.Steps)
	placements := map[int]*store.StepPlacement{}
	for i, s := range w.Steps {
		if s.Placement != nil {
			// Placement (#409): the step starts on its top option.
			sp, err := resolvePlacement(cfg, &w, s.Placement, vars)
			if err != nil {
				return runPlan{}, fmt.Errorf("%w: workflow %s step %s: %w", store.ErrInvalid, p.Workflow, s.ID, err)
			}
			placements[i] = sp
			w.Steps[i].Pool = sp.Options[0].Pool
			w.Steps[i].NeedsResources = sp.Options[0].Resources
			continue
		}
		if s.Auto() && s.Pool == "" {
			continue
		}
		name, err := cfg.ResolvePool(&w, s.Pool, vars)
		if err != nil {
			return runPlan{}, fmt.Errorf("%w: workflow %s step %s: %w", store.ErrInvalid, p.Workflow, s.ID, err)
		}
		w.Steps[i].Pool = name
		// Resource keys too (#400): an unknown resource fails the start.
		rs, err := cfg.ResolveResources(&w, s.NeedsResources, vars)
		if err != nil {
			return runPlan{}, fmt.Errorf("%w: workflow %s step %s: %w", store.ErrInvalid, p.Workflow, s.ID, err)
		}
		w.Steps[i].NeedsResources = rs
	}
	// Steps live in their pools' projects; the run goes with the first.
	proj := stepProject(cfg, w.Steps[0])
	if p.Project != "" && p.Project != proj {
		return runPlan{}, fmt.Errorf("%w: workflow %s runs in project %s", store.ErrInvalid, p.Workflow, proj)
	}
	if !sc.admin && sc.project != proj {
		return runPlan{}, forbidden("session is scoped to project %s", sc.project)
	}
	meta, err := runMeta(w, vars)
	if err != nil {
		return runPlan{}, err
	}
	for _, id := range p.After {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return runPlan{}, fmt.Errorf("after %s: %w", id, err)
		}
		if !sc.admin && t.Project != sc.project {
			return runPlan{}, forbidden("after %s: session is scoped to project %s", id, sc.project)
		}
	}
	plan := runPlan{w: w, proj: proj, vars: vars, meta: meta, placements: placements}
	if err := c.intakeRun(ctx, plan); err != nil {
		return runPlan{}, err
	}
	return plan, nil
}

// buildRun returns the run ticket (first), its steps in order and their
// edges, without writing. parent is the run's parent_id (a batch, or "").
func (c *Controller) buildRun(plan runPlan, p rpc.RunStartParams, extra map[string]any, parent string) (store.Graph, error) {
	cfg := c.config()
	w, proj, vars, meta := plan.w, plan.proj, plan.vars, plan.meta
	var g store.Graph
	runID, err := c.store.NewID(proj)
	if err != nil {
		return g, err
	}
	pack := ""
	if s := cfg.Sources["workflows."+p.Workflow]; len(s) > 0 {
		pack = s[0]
	}
	rf, err := json.Marshal(store.RunFields{Workflow: p.Workflow, Pack: pack, Vars: p.Vars, Key: p.Key, State: store.RunRunning})
	if err != nil {
		return g, err
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
				return g, err
			}
		}
		maps.Copy(m, extra)
		if runMetaJSON, err = json.Marshal(m); err != nil {
			return g, err
		}
	}
	nts := []store.NewTicket{{ID: runID, Project: proj, Kind: store.KindRun, Title: title, ParentID: parent, Fields: rf, Meta: runMetaJSON}}
	var edges []store.Edge
	stepIDs := map[string]string{} // workflow step id -> ticket id
	for i, s := range w.Steps {
		title, err := w.Render(s.Title, vars)
		if err != nil {
			return g, fmt.Errorf("%w: step %s title: %w", store.ErrInvalid, s.ID, err)
		}
		body, err := w.Render(s.Prompt, vars)
		if err != nil {
			return g, fmt.Errorf("%w: step %s prompt: %w", store.ErrInvalid, s.ID, err)
		}
		id, err := c.store.NewID(stepProject(cfg, s))
		if err != nil {
			return g, err
		}
		needs := w.StepNeeds(i)
		stepF := store.StepFields{Run: runID, Step: s.ID, Index: i, Retries: s.RetryLimit(), Resources: s.NeedsResources, Needs: needs, Placement: plan.placements[i]}
		pool := s.Pool
		if s.Auto() {
			// No pool on the ticket: no session may claim it; the
			// controller runs it (#396).
			stepF.Type, pool = s.Type, ""
		}
		nt := store.NewTicket{
			ID: id, Project: stepProject(cfg, s), Kind: store.KindStep, Title: title, Body: body,
			Pool: pool, ParentID: runID, Meta: meta,
		}
		if s.When != nil {
			// Held until checkRun decides: open it, or close it skipped.
			stepF.When = &store.StepWhen{Step: s.When.Step, Outcome: s.When.Outcome}
			nt.DeferUntil = gateHold
		}
		if nt.Fields, err = json.Marshal(stepF); err != nil {
			return g, err
		}
		nts = append(nts, nt)
		for _, n := range needs {
			edges = append(edges, store.Edge{From: id, To: stepIDs[n], Kind: store.EdgeBlocks})
		}
		if len(needs) == 0 {
			// Same graph apply as the step: no window where it is
			// claimable without its outside blockers (#149).
			for _, a := range p.After {
				edges = append(edges, store.Edge{From: id, To: a, Kind: store.EdgeBlocks})
			}
		}
		stepIDs[s.ID] = id
	}
	return store.Graph{Tickets: nts, Edges: edges}, nil
}

// stepProject is the step's pool's project, or the default project for
// a command or http step with no pool.
func stepProject(cfg *config.Config, s config.Step) string {
	if p, ok := cfg.Pools[s.Pool]; ok {
		return p.Project
	}
	return cfg.DefaultProject
}

// gateHold is the defer_until of a step with a when condition until its
// needs finish (#385): far enough out that nothing claims it.
var gateHold = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

// setStepOutcome records `close --outcome` on an open step ticket.
func (c *Controller) setStepOutcome(ctx context.Context, sc scope, id, outcome string) error {
	if !config.OutcomePattern.MatchString(outcome) {
		return fmt.Errorf("%w: outcome %q must match %s", store.ErrInvalid, outcome, config.OutcomePattern)
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if t.Kind != store.KindStep {
		return fmt.Errorf("%w: --outcome is for workflow steps; %s is a %s", store.ErrInvalid, id, t.Kind)
	}
	f, err := store.DecodeFields[store.StepFields](t)
	if err != nil {
		return err
	}
	f.Outcome = outcome
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	_, err = c.store.Update(ctx, sc.actor, id, store.Update{Fields: b, IfRevision: t.Revision})
	return err
}

// gateSteps opens or skips gated steps whose needs have all finished,
// repeating while a skip unblocks more. It reports whether it changed
// anything.
func (c *Controller) gateSteps(ctx context.Context, runID string) bool {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	changed := false
	for range 64 { // bound: each pass finishes at least one step
		steps, err := store.Children(ctx, c.store, runID)
		if err != nil {
			return changed
		}
		byID := map[string]store.Ticket{}
		fields := map[string]store.StepFields{}
		for _, s := range steps {
			if f, err := store.DecodeFields[store.StepFields](s); err == nil {
				byID[f.Step], fields[f.Step] = s, f
			}
		}
		acted := false
		for name, f := range fields {
			s := byID[name]
			if f.When == nil || s.Status == store.StatusClosed || !s.DeferUntil.Equal(gateHold) {
				continue
			}
			ready := true
			for _, n := range f.Needs {
				if !store.StepFinished(byID[n]) {
					ready = false
				}
			}
			if !ready {
				continue
			}
			gate := byID[f.When.Step]
			if gate.Resolution == "done" && fields[f.When.Step].Outcome == f.When.Outcome {
				var zero time.Time
				if _, err := c.store.Update(ctx, "controller", s.ID, store.Update{DeferUntil: &zero}); err != nil {
					c.log.Error("open gated step", "step", s.ID, "err", err)
					return changed
				}
				c.log.Info("gated step opened", "run", runID, "step", name)
			} else {
				if _, err := store.Close(ctx, c.store, "controller", s.ID, nil, store.StepSkipped); err != nil {
					c.log.Error("skip step", "step", s.ID, "err", err)
					return changed
				}
				_, _ = c.store.AddComment(ctx, "controller", s.ID, fmt.Sprintf("Skipped: step %s outcome %q, wanted %q.",
					f.When.Step, fields[f.When.Step].Outcome, f.When.Outcome))
				c.log.Info("step skipped", "run", runID, "step", name)
			}
			acted, changed = true, true
		}
		if !acted {
			return changed
		}
	}
	return changed
}

// runMeta is meta.github = {repo, issue|pr} when the workflow names an
// issue or pr var (#183). A numeric value is stored as a number.
func runMeta(w config.Workflow, vars map[string]any) (json.RawMessage, error) {
	key, name := "issue", w.Issue
	if w.PR != "" {
		key, name = "pr", w.PR
	}
	if name == "" {
		return nil, nil
	}
	if w.Tracker == config.TrackerLinear {
		if w.LinearWorkspace == "" {
			return nil, fmt.Errorf("%w: workflow has tracker linear but no linear_workspace (set the pack's linear_workspace var)", store.ErrInvalid)
		}
		// The issue is a Linear key; meta.github keeps only the code repo,
		// so GitHub readers (subscriptions, issue_no_run) do not take
		// ENG-123 for a GitHub issue (#201).
		key := strings.ToUpper(strings.TrimSpace(fmt.Sprint(vars[name])))
		if !linear.KeyPattern.MatchString(key) {
			return nil, fmt.Errorf("%w: %s %q is not a Linear issue key like ENG-123", store.ErrInvalid, name, key)
		}
		lm := map[string]any{"workspace": w.LinearWorkspace, "issue": key}
		// Who questions and escalations mention on the issue (#293).
		if h, _ := w.PackVars["linear_human"].(string); strings.TrimSpace(h) != "" {
			lm["human"] = strings.TrimSpace(h)
		}
		return json.Marshal(map[string]map[string]any{
			"linear": lm,
			"github": {"repo": w.Repo},
		})
	}
	var ref any = strings.TrimPrefix(fmt.Sprint(vars[name]), "#")
	if n, err := strconv.Atoi(ref.(string)); err == nil {
		ref = n
	}
	return json.Marshal(map[string]map[string]any{"github": {"repo": w.Repo, key: ref}})
}

func (c *Controller) runResult(ctx context.Context, run store.Ticket) (rpc.RunResult, error) {
	steps, err := store.Children(ctx, c.store, run.ID)
	if err != nil {
		return rpc.RunResult{}, err
	}
	waits, err := c.runWaits(ctx, steps)
	if err != nil {
		return rpc.RunResult{}, err
	}
	return rpc.RunResult{Run: run, Steps: steps, Waits: waits}, nil
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
	return c.store.List(ctx, store.Filter{Kind: store.KindRun, Project: p.Project, State: p.State, Order: store.OrderNewest, Limit: p.Limit})
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
	if err == nil {
		// A run wait reads the state: wake the waits on this run (#679).
		c.wakeWaitsSubjects(runSubject(t.ID))
	}
	if err != nil || resolution == "" {
		return t, err
	}
	t, err = store.Close(ctx, c.store, actor, t.ID, nil, resolution)
	if err == nil {
		c.removeRunDir(t.ID)
	}
	return t, err
}

func (c *Controller) runCancel(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	res, err := c.cancelRun(ctx, sc, p)
	if err == nil {
		c.runBatchEvent(ctx, p.ID)
	}
	return res, err
}

func (c *Controller) cancelRun(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	t, f, err := c.getRun(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: run %s", store.ErrClosed, t.ID)
	}
	steps, err := store.Children(ctx, c.store, t.ID)
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		if s.Status == store.StatusClosed {
			continue
		}
		// Administrative close: a session holding the step loses it.
		if _, err := store.Close(ctx, c.store, sc.actor, s.ID, nil, store.RunCancelled); err != nil {
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
	res, err := c.resumeRun(ctx, sc, p)
	if err == nil {
		c.runBatchEvent(ctx, p.ID)
	}
	return res, err
}

func (c *Controller) resumeRun(ctx context.Context, sc scope, p rpc.RunResumeParams) (any, error) {
	c.runMu.Lock()
	defer c.runMu.Unlock()
	t, f, err := c.getRun(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: run %s", store.ErrClosed, t.ID)
	}
	steps, err := store.Children(ctx, c.store, t.ID)
	if err != nil {
		return nil, err
	}
	var step *store.Ticket
	for i := range steps {
		sf, err := store.DecodeFields[store.StepFields](steps[i])
		if err != nil {
			return nil, err
		}
		done := store.StepFinished(steps[i])
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
	if step.Status == store.StatusClosed && step.Resolution == "done" {
		// Reopening a finished step is a review finding against it.
		c.recordReviewFinding(ctx, sc, *step)
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

// completeRun closes a running run done once every step is closed done
// or skipped.
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
	steps, err := store.Children(ctx, c.store, runID)
	if err != nil || len(steps) == 0 {
		return
	}
	for _, s := range steps {
		if !store.StepFinished(s) {
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
	c.sweepRunDirs()
	for _, state := range []string{store.RunRunning, store.RunNeedsHuman} {
		runs, err := c.store.List(c.ctx, store.Filter{Kind: store.KindRun, State: state, Order: store.OrderNewest})
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
