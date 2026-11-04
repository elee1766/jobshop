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
	waits := map[string]store.ResourceWait{}
	for _, wt := range r.Waits {
		waits[wt.Ticket] = wt
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
		line := fmt.Sprintf("%d  %-12s %s  pool %s  %s  %s  holder %s  failures %d/%d",
			sf.Index, sf.Step, s.ID, s.Pool, s.Status, res, holder, sf.Failures, sf.Retries)
		// The graph (#385): what the step waits on, its gate and outcome.
		if len(sf.Needs) > 0 {
			line += "  needs " + strings.Join(sf.Needs, ",")
		}
		if sf.When != nil {
			line += fmt.Sprintf("  when %s=%s", sf.When.Step, sf.When.Outcome)
		}
		if sf.Outcome != "" {
			line += "  outcome " + sf.Outcome
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
		if len(sf.Resources) > 0 {
			res := "   resources " + strings.Join(sf.Resources, " ")
			if wt, ok := waits[s.ID]; ok {
				// Out of the ready frontier on a full resource (#408).
				res += fmt.Sprintf("  waiting on %s, %d ahead (%d/%d held)", wt.Key, wt.Ahead, wt.Held, wt.Capacity)
			}
			if _, err := fmt.Fprintln(w, res); err != nil {
				return err
			}
		}
		if sf.Auto() && sf.Command != "" {
			// Command and http steps (#396): what ran and what it said.
			if _, err := fmt.Fprintf(w, "   %s %s -> %s\n", sf.Type, sf.Command, sf.Result); err != nil {
				return err
			}
			if out := strings.TrimSpace(sf.Output); out != "" {
				if len(out) > 400 {
					out = out[:400] + "..."
				}
				if _, err := fmt.Fprintf(w, "   output: %s\n", strings.ReplaceAll(out, "\n", "\n           ")); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type batchCmd struct {
	Start batchStartCmd `cmd:"" help:"Start one run per issue under a batch, with ordered cross-run edges."`
	Show  batchShowCmd  `cmd:"" help:"Show a batch's state and one line per member run."`
	List  batchListCmd  `cmd:"" help:"List open batches."`
}

type batchShowCmd struct {
	ID   string `arg:"" help:"Batch id or key."`
	JSON bool   `help:"Print JSON."`
}

func (c batchShowCmd) Run(g *Globals) error {
	p := rpc.IDParams{ID: c.ID}
	if c.JSON {
		return g.call(rpc.MethodBatchShow, p, nil)
	}
	var r rpc.BatchResult
	if err := g.call(rpc.MethodBatchShow, p, &r); err != nil {
		return err
	}
	return printBatch(os.Stdout, r)
}

type batchListCmd struct {
	Project string `help:"Project."`
	Limit   int    `short:"n" help:"Max results (default 500)."`
}

func (c batchListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodBatchList, rpc.BatchListParams{Project: c.Project, Limit: c.Limit}, nil)
}

type batchStartCmd struct {
	Key      string   `short:"k" required:"" help:"Batch key: returns the open batch with this key instead of starting another."`
	Issues   []string `required:"" sep:"," help:"Issue numbers, comma separated (122,123,125)."`
	Order    []string `help:"Order chain a>b>c: b's first step waits on a's last (repeatable)."`
	Workflow string   `help:"Workflow each member runs (default pr-work)."`
	Project  string   `help:"Project (default: the workflow's)."`
	JSON     bool     `help:"Print JSON."`
}

func (c batchStartCmd) params() rpc.BatchStartParams {
	return rpc.BatchStartParams{Key: c.Key, Issues: c.Issues, Order: c.Order, Workflow: c.Workflow, Project: c.Project}
}

func (c batchStartCmd) Run(g *Globals) error {
	if c.JSON {
		return g.call(rpc.MethodBatchStart, c.params(), nil)
	}
	var r rpc.BatchResult
	if err := g.call(rpc.MethodBatchStart, c.params(), &r); err != nil {
		return err
	}
	return printBatch(os.Stdout, r)
}

// printBatch writes the batch line, its order, then one line per member
// run: issue, run id, run state, current step, blocking issues, adopted.
func printBatch(w io.Writer, r rpc.BatchResult) error {
	f, err := store.DecodeFields[store.BatchFields](r.Batch)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("batch %s  key %s  state %s  (%s)", r.Batch.ID, f.Key, f.State, r.Batch.Status)
	if r.Existing {
		line += "  [existing batch for key]"
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	if len(f.Order) > 0 {
		parts := make([]string, len(f.Order))
		for i, p := range f.Order {
			parts[i] = p[0] + ">" + p[1]
		}
		if _, err := fmt.Fprintln(w, "order", strings.Join(parts, " ")); err != nil {
			return err
		}
	}
	for _, run := range r.Runs {
		rf, err := store.DecodeFields[store.RunFields](run.Run)
		if err != nil {
			return err
		}
		extra := ""
		if s := currentStepName(run.Steps); s != "" {
			extra += "  step " + s
		}
		if bl := r.BlockedBy[run.Run.ID]; len(bl) > 0 {
			extra += "  blocked by " + strings.Join(bl, ",")
		}
		if slices.Contains(r.Adopted, run.Run.ID) {
			extra += "  [adopted]"
		}
		if _, err := fmt.Fprintf(w, "%s  run %s  state %s%s\n", rf.Key, run.Run.ID, rf.State, extra); err != nil {
			return err
		}
	}
	return nil
}

// currentStepName is the workflow step name of the first step not closed
// done, or "" when all are done.
func currentStepName(steps []store.Ticket) string {
	for _, s := range steps {
		if store.StepFinished(s) {
			continue
		}
		if sf, err := store.DecodeFields[store.StepFields](s); err == nil {
			return sf.Step
		}
		return s.ID
	}
	return ""
}
