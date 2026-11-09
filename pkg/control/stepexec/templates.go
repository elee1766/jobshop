package stepexec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Templates is what a command or http step's templates are filled
// from: the run's workflow and vars, and earlier steps' outputs.
type Templates struct {
	W     config.Workflow
	Vars  map[string]any
	Steps map[string]any // id -> {output, outcome, json} for template: go
	Out   map[string]domain.StepFields
}

// Fill renders one template.
func (t Templates) Fill(s string) (string, error) {
	return t.W.FillStepTemplate(s, t.Vars, t.Steps)
}

// LoadTemplates loads the run's vars and earlier steps' outputs.
func LoadTemplates(ctx context.Context, st store.Store, cfg *config.Config, sf domain.StepFields) (Templates, error) {
	var t Templates
	run, err := st.Get(ctx, sf.Run)
	if err != nil {
		return t, err
	}
	rf, err := domain.DecodeFields[domain.RunFields](run)
	if err != nil {
		return t, err
	}
	t.Vars = map[string]any{}
	if w, ok := cfg.Workflows[rf.Workflow]; ok {
		t.W = w
		for k, v := range w.PackVars {
			t.Vars[k] = v
		}
	}
	for k, v := range rf.Vars {
		t.Vars[k] = v
	}
	steps, err := store.Children(ctx, st, sf.Run)
	if err != nil {
		return t, err
	}
	t.Out = map[string]domain.StepFields{}
	t.Steps = map[string]any{}
	for _, s := range steps {
		f, err := domain.DecodeFields[domain.StepFields](s)
		if err != nil || f.Index >= sf.Index {
			continue
		}
		t.Out[f.Step] = f
		m := map[string]any{"output": f.Output, "outcome": f.Outcome}
		var j any
		if json.Unmarshal([]byte(f.Output), &j) == nil {
			m["json"] = j
		}
		t.Steps[f.Step] = m
	}
	return t, nil
}

// Env is what a command step gets on top of the controller's env: the
// run vars as JOBSHOP_VAR_<NAME> and the attempt key.
func Env(vars map[string]any, attemptKey string, t domain.Ticket) []string {
	var env []string
	for k, v := range vars {
		env = append(env, "JOBSHOP_VAR_"+strings.ToUpper(k)+"="+fmt.Sprint(v))
	}
	return append(env, "JOBSHOP_STEP_KEY="+attemptKey, "JOBSHOP_TICKET="+t.ID)
}
