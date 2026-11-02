package cli

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type runCmd struct {
	Start  runStartCmd  `cmd:"" help:"Start a workflow run: a run ticket and one step ticket per workflow step."`
	Show   runShowCmd   `cmd:"" help:"Show a run's state, vars and steps."`
	List   runListCmd   `cmd:"" help:"List runs, newest first."`
	Cancel runCancelCmd `cmd:"" help:"Close a run and its open steps as cancelled."`
	Resume runResumeCmd `cmd:"" help:"Reopen a run's step (default: the first unfinished) with failures reset."`
}

type runStartCmd struct {
	Workflow string            `arg:"" help:"Workflow name from the config."`
	Var      map[string]string `short:"v" help:"Workflow var, k=v (repeatable)."`
	Key      string            `short:"k" help:"Dedup key: returns the open run with this workflow and key instead of starting another."`
	Project  string            `help:"Project (default: the first step's pool project)."`
	After    []string          `help:"Ticket or run id the first step waits on (repeatable); the edge is created with the run."`
	JSON     bool              `help:"Print JSON."`
}

func (c runStartCmd) params() rpc.RunStartParams {
	return rpc.RunStartParams{Workflow: c.Workflow, Vars: c.Var, Key: c.Key, Project: c.Project, After: c.After}
}

func (c runStartCmd) Run(g *Globals) error {
	return g.runCall(rpc.MethodRunStart, c.params(), c.JSON)
}

type runShowCmd struct {
	ID   string `arg:"" help:"Run id."`
	JSON bool   `help:"Print JSON."`
}

func (c runShowCmd) Run(g *Globals) error {
	return g.runCall(rpc.MethodRunShow, rpc.IDParams{ID: c.ID}, c.JSON)
}

type runListCmd struct {
	Project string `help:"Project."`
	State   string `short:"s" enum:",running,needs_human,done,cancelled" default:"" help:"Run state."`
	Limit   int    `short:"n" help:"Max results (default 500)."`
}

func (c runListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodRunList, rpc.RunListParams{Project: c.Project, State: c.State, Limit: c.Limit}, nil)
}

type runCancelCmd struct {
	ID   string `arg:"" help:"Run id."`
	JSON bool   `help:"Print JSON."`
}

func (c runCancelCmd) Run(g *Globals) error {
	return g.runCall(rpc.MethodRunCancel, rpc.IDParams{ID: c.ID}, c.JSON)
}

type runResumeCmd struct {
	ID   string `arg:"" help:"Run id."`
	Step string `help:"Step id to reopen (default: the first step not closed done)."`
	JSON bool   `help:"Print JSON."`
}

func (c runResumeCmd) params() rpc.RunResumeParams {
	return rpc.RunResumeParams{ID: c.ID, Step: c.Step}
}

func (c runResumeCmd) Run(g *Globals) error {
	return g.runCall(rpc.MethodRunResume, c.params(), c.JSON)
}

// runCall calls a method returning rpc.RunResult and prints it as text,
// or as JSON with asJSON.
func (g *Globals) runCall(method string, params any, asJSON bool) error {
	if asJSON {
		return g.call(method, params, nil)
	}
	var r rpc.RunResult
	if err := g.call(method, params, &r); err != nil {
		return err
	}
	return printRun(os.Stdout, r)
}

// printRun writes the run line, its vars, then one line per step: index,
// id, pool, status, resolution, holder, failures/retries.
func printRun(w io.Writer, r rpc.RunResult) error {
	f, err := store.DecodeFields[store.RunFields](r.Run)
	if err != nil {
		return err
	}
	status := string(r.Run.Status)
	if r.Run.Resolution != "" {
		status += " " + r.Run.Resolution
	}
	line := fmt.Sprintf("run %s  workflow %s  state %s  (%s)", r.Run.ID, f.Workflow, f.State, status)
	if f.Key != "" {
		line += "  key " + f.Key
	}
	if r.Existing {
		line += "  [existing run for key]"
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	if len(f.Vars) > 0 {
		keys := make([]string, 0, len(f.Vars))
		for k := range f.Vars {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + "=" + f.Vars[k]
		}
		if _, err := fmt.Fprintln(w, "vars", strings.Join(parts, " ")); err != nil {
			return err
		}
	}
	for _, s := range r.Steps {
		sf, err := store.DecodeFields[store.StepFields](s)
		if err != nil {
			return err
		}
		res, holder := s.Resolution, s.Holder
		if res == "" {
			res = "-"
		}
		if holder == "" {
			holder = "-"
		}
		if _, err := fmt.Fprintf(w, "%d  %-12s %s  pool %s  %s  %s  holder %s  failures %d/%d\n",
			sf.Index, sf.Step, s.ID, s.Pool, s.Status, res, holder, sf.Failures, sf.Retries); err != nil {
			return err
		}
	}
	return nil
}
