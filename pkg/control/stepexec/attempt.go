package stepexec

import (
	"context"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
)

// Attempt is one attempt of a step.
type Attempt struct {
	Cfg    *config.Config
	Ticket store.Ticket
	Step   config.Step
	Fields store.StepFields
	// Key is the attempt key: run/step/fence.
	Key string
	// Store loads the run's vars and earlier steps' outputs.
	Store store.Store
	// Engine and Inputs are for script steps.
	Engine *hooks.Engine
	Inputs func(context.Context) (ScriptInputs, error)
}

// Run runs the attempt under its step timeout and returns the result. It
// does not record it: the caller owns claims, leases and the step record,
// and drops the result when shutdown cut the attempt.
func (e Executor) Run(ctx context.Context, a Attempt) Result {
	tpl, err := LoadTemplates(ctx, a.Store, a.Cfg, a.Fields)
	if err != nil {
		return Result{Result: ResultFailed, Err: err.Error()}
	}
	actx, cancel := context.WithTimeout(ctx, a.Step.TimeoutDur())
	defer cancel()
	switch a.Step.Type {
	case config.StepCommand:
		return e.command(actx, a.Cfg, a.Ticket, a.Step, a.Fields, tpl, a.Key)
	case config.StepHTTP:
		return e.httpStep(actx, a.Cfg, a.Step, tpl, a.Key)
	case store.StepScript:
		return e.script(actx, a.Cfg, a.Ticket, a.Step, a.Fields, tpl, a.Key, a.Engine, a.Inputs)
	}
	// Not reachable: autoStepConfig only returns known types. The empty
	// result is recorded as a failure.
	return Result{}
}
