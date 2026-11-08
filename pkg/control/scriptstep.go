package control

import (
	"context"
	"errors"
	"fmt"

	"github.com/elee1766/jobshop/pkg/control/stepexec"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
)

// Script steps (JOB-124) run in stepexec. What stays here is loading the
// run the controller pinned: it reads the live callback runtime.

// scriptInputs loads the run and step envelopes and the program the run
// pinned. A run whose callback code changed since it started fails the
// attempt rather than running new code against old state.
func (c *Controller) scriptInputs(ctx context.Context, t store.Ticket, sf store.StepFields) (stepexec.ScriptInputs, error) {
	runTicket, err := c.store.Get(ctx, sf.Run)
	if err != nil {
		return stepexec.ScriptInputs{}, err
	}
	rf, err := store.DecodeFields[store.RunFields](runTicket)
	if err != nil {
		return stepexec.ScriptInputs{}, err
	}
	if rf.Automation == nil {
		return stepexec.ScriptInputs{}, errors.New("script step outside a callback workflow")
	}
	runtime := c.live.Load().runtime
	if !hooks.RunPinned(runtime.Programs, runtime.Config, rf.Workflow, rf.Automation.Digest) {
		return stepexec.ScriptInputs{}, errors.New("workflow callback code changed since the run started: restore it or cancel the run")
	}
	program := runtime.Programs[sf.ScriptHook]
	if program == nil {
		return stepexec.ScriptInputs{}, fmt.Errorf("callback %s is not loaded", sf.ScriptHook)
	}
	children, err := store.Children(ctx, c.store, sf.Run)
	if err != nil {
		return stepexec.ScriptInputs{}, err
	}
	run := &hooks.Run{ID: runTicket.ID, Workflow: rf.Workflow, Key: rf.Key, Project: runTicket.Project, Inputs: rf.Inputs, Vars: rf.Vars, State: rf.State, Steps: []hooks.Step{}}
	var step hooks.Step
	for _, child := range children {
		if child.Kind != store.KindStep {
			continue
		}
		in, err := hooks.StepInput(child)
		if err != nil {
			return stepexec.ScriptInputs{}, err
		}
		run.Steps = append(run.Steps, in)
		if child.ID == t.ID {
			step = in
		}
	}
	return stepexec.ScriptInputs{Run: run, Step: &step, Program: program}, nil
}
