package github_test

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func packRoot(t *testing.T, prefix, tracker string, callbacks bool) config.Root {
	t.Helper()
	root := controltest.NewRoot(t, "pack: pack.cue\n")
	assets := filepath.Join(root.Dir, "packs", "github")
	require.NoError(t, os.MkdirAll(assets, 0o700))
	require.NoError(t, os.CopyFS(assets, os.DirFS(".")))
	write(t, root.Dir, "cue.mod/module.cue", "module: \"github.com/elee1766/jobshop@v0\"\nlanguage: version: \"v0.17.0\"\n")
	write(t, root.Dir, "pack.cue", fmt.Sprintf(`package acceptance
import githubpack "github.com/elee1766/jobshop/packs/github"
default_project: "js"
projects: js: {}
agents: local: {command: "false"}
packs: github: githubpack.#Pack & {
    prefix: %q
    vars: {
        repo: "acme/widget"
        checkout: %q
        agent: "local"
        human: "operator"
        tracker: %q
        linear_workspace: "acme"
        starlark_workflows: %t
    }
}
`, prefix, root.Dir, tracker, callbacks))
	write(t, root.Dir, config.FileName, "pack: pack.cue\n")
	return root
}

func write(t *testing.T, root, name, data string) {
	t.Helper()
	file := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o700))
	require.NoError(t, os.WriteFile(file, []byte(data), 0o600))
}

// This comparison uses the shipped YAML as the migration contract, not a copied fixture.
func TestNativePackStaticParity(t *testing.T) {
	for _, tracker := range []string{"github", "linear"} {
		for _, prefix := range []string{"", "qa-"} {
			t.Run(tracker+"/"+prefix, func(t *testing.T) {
				root := packRoot(t, prefix, tracker, false)
				native, err := config.Load(root)
				require.NoError(t, err)
				legacy := loadLegacy(t, root, prefix, tracker)
				for name, workflow := range native.Workflows {
					workflow.PackVars = maps.Clone(workflow.PackVars)
					delete(workflow.PackVars, "starlark_workflows")
					delete(workflow.PackVars, "asset_prefix")
					native.Workflows[name] = workflow
				}
				assert.Equal(t, legacy.Pools, native.Pools, "pool capacities, models, permissions and rendered prompts")
				assert.Equal(t, legacy.Workflows, native.Workflows, "workflow graph, conditions, retries and prompt templates")
				assert.Equal(t, legacy.Triggers, native.Triggers, "source filters, checks, keys, timers and history caps")
				assert.Equal(t, legacy.PromptRules, native.PromptRules, "prompt rule order, filters and bytes")
			})
		}
	}
}

func loadLegacy(t *testing.T, root config.Root, prefix, tracker string) *config.Config {
	t.Helper()
	write(t, root.Dir, "legacy.yaml", fmt.Sprintf(`default_project: js
projects: {js: {}}
agents: {local: {command: "false"}}
packs:
  - path: packs/github
    prefix: %q
    vars:
      repo: acme/widget
      checkout: %q
      agent: local
      human: operator
      tracker: %q
      linear_workspace: acme
`, prefix, root.Dir, tracker))
	cfg, err := config.LoadFile(filepath.Join(root.Dir, "legacy.yaml"))
	require.NoError(t, err)
	return cfg
}

func offlineController(t *testing.T) (*controltest.Controller, *config.Config) {
	t.Helper()
	root := packRoot(t, "qa-", "github", true)
	legacy := loadLegacy(t, root, "qa-", "github")
	write(t, root.Dir, "offline.cue", `package acceptance
pools: {
    for name in ["dev", "dev-sonnet", "plan", "stall", "refinery", "deploy-check"] {
        "qa-\(name)": max_sessions: 0
    }
}
triggers: {
    for name in ["replan", "master-red", "release-pr", "stall", "pr-unmerged"] {
        "qa-\(name)": enabled: false
    }
}
`)
	return controltest.Start(t, root, control.Options{}), legacy
}

func runSnapshot(c *controltest.Controller, id string) rpc.RunResult {
	var result rpc.RunResult
	c.Call(rpc.MethodRunShow, rpc.IDParams{ID: id}, &result)
	return result
}

func waitStep(t *testing.T, c *controltest.Controller, run, id string) store.Ticket {
	t.Helper()
	var found store.Ticket
	c.PassUntil(func() bool {
		for _, step := range runSnapshot(c, run).Steps {
			fields, err := store.DecodeFields[store.StepFields](step)
			require.NoError(t, err)
			if fields.Step == id {
				found = step
				return true
			}
		}
		return false
	}, "native workflow did not create "+id)
	return found
}

func checkStep(t *testing.T, cfg *config.Config, workflow string, index int, actual store.Ticket, vars map[string]any) {
	t.Helper()
	w := cfg.Workflows[workflow]
	inputs := maps.Clone(w.PackVars)
	maps.Copy(inputs, vars)
	want := w.Steps[index]
	pool, err := cfg.ResolvePool(&w, want.Pool, inputs)
	require.NoError(t, err)
	assert.Equal(t, pool, actual.Pool)
	title, err := w.Render(want.Title, inputs)
	require.NoError(t, err)
	assert.Equal(t, title, actual.Title)
	body, err := w.Render(want.Prompt, inputs)
	require.NoError(t, err)
	assert.Equal(t, body, actual.Body, "the native action must render the original prompt bytes")
	fields, err := store.DecodeFields[store.StepFields](actual)
	require.NoError(t, err)
	assert.Equal(t, want.RetryLimit(), fields.Retries)
	assert.Equal(t, w.StepNeeds(index), fields.Needs)
	assert.Equal(t, w.RequestChangesTarget(index), fields.OnRequestChanges)
}

func TestNativeWorkflowLifecycle(t *testing.T) {
	for _, tc := range []struct {
		workflow string
		vars     map[string]string
	}{
		{"ci-repair", map[string]string{"sha": "1234567890abcdef"}},
		{"release-merge", map[string]string{"pr": "46"}},
		{"stall-check", map[string]string{"signal": "blocked", "subject": "js-47", "detail": "waiting for a dependency"}},
	} {
		t.Run(tc.workflow, func(t *testing.T) {
			c, legacy := offlineController(t)
			name := "qa-" + tc.workflow
			var run rpc.RunResult
			c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: name, Key: tc.workflow, Vars: tc.vars}, &run)
			inputs := map[string]any{}
			for key, value := range tc.vars {
				inputs[key] = value
			}
			for index, expected := range legacy.Workflows[name].Steps {
				step := waitStep(t, c, run.Run.ID, expected.ID)
				checkStep(t, legacy, name, index, step, inputs)
				c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: step.ID}, nil)
			}
			c.PassUntil(func() bool { return runSnapshot(c, run.Run.ID).Run.Status == store.StatusClosed }, "native workflow did not finish")
		})
	}
	for _, finalOutcome := range []string{"pr", "done"} {
		t.Run("retry and review findings/"+finalOutcome, func(t *testing.T) {
			c, legacy := offlineController(t)
			var run rpc.RunResult
			vars := map[string]string{"issue": "42", "dev_pool": "qa-dev-sonnet"}
			c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "qa-pr-work", Key: "issue-42", Vars: vars}, &run)
			implement := waitStep(t, c, run.Run.ID, "implement")
			inputs := map[string]any{"issue": "42", "dev_pool": "qa-dev-sonnet"}
			checkStep(t, legacy, "qa-pr-work", 0, implement, inputs)
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: implement.ID, Resolution: "failed"}, nil)
			implement = waitStep(t, c, run.Run.ID, "implement")
			assert.Equal(t, store.StatusOpen, implement.Status)
			assert.Equal(t, "qa-dev", implement.Pool, "failed cheap attempts use the strong retry pool")
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: implement.ID, Outcome: "pr"}, nil)
			review := waitStep(t, c, run.Run.ID, "review-merge")
			checkStep(t, legacy, "qa-pr-work", 1, review, inputs)
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: review.ID, Resolution: "request_changes", Comment: "Cover the missing case"}, nil)
			implement = waitStep(t, c, run.Run.ID, "implement")
			assert.Equal(t, store.StatusOpen, implement.Status)
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: implement.ID, Outcome: finalOutcome}, nil)
			c.Pass(2)
			assert.Len(t, runSnapshot(c, run.Run.ID).Steps, 2, "review recovery must not duplicate the reviewer")
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: review.ID}, nil)
			if finalOutcome != "pr" {
				c.PassUntil(func() bool { return runSnapshot(c, run.Run.ID).Run.Status == store.StatusClosed }, "recovered non-PR run did not finish")
				assert.Len(t, runSnapshot(c, run.Run.ID).Steps, 2, "post-merge remains gated on the latest implement outcome")
				return
			}
			post := waitStep(t, c, run.Run.ID, "post-merge")
			checkStep(t, legacy, "qa-pr-work", 2, post, inputs)
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: post.ID}, nil)
			c.PassUntil(func() bool { return runSnapshot(c, run.Run.ID).Run.Status == store.StatusClosed }, "shipped native run did not finish")
			assert.Equal(t, "done", runSnapshot(c, run.Run.ID).Run.Resolution)
		})
	}
	t.Run("no PR finishes without refinery", func(t *testing.T) {
		c, _ := offlineController(t)
		var run rpc.RunResult
		c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "qa-pr-work", Key: "issue-43", Vars: map[string]string{"issue": "43"}}, &run)
		implement := waitStep(t, c, run.Run.ID, "implement")
		assert.Equal(t, "qa-dev", implement.Pool, "the default pool remains inside the instance")
		c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: implement.ID, Outcome: "done"}, nil)
		c.PassUntil(func() bool { return runSnapshot(c, run.Run.ID).Run.Status == store.StatusClosed }, "no-PR run did not finish")
		assert.Len(t, runSnapshot(c, run.Run.ID).Steps, 1)
	})
	t.Run("held PR does not deploy", func(t *testing.T) {
		c, legacy := offlineController(t)
		var run rpc.RunResult
		c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "qa-pr-review", Key: "pr-44", Vars: map[string]string{"pr": "44"}}, &run)
		review := waitStep(t, c, run.Run.ID, "review-merge")
		checkStep(t, legacy, "qa-pr-review", 0, review, map[string]any{"pr": "44"})
		c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: review.ID, Outcome: "held"}, nil)
		c.PassUntil(func() bool { return runSnapshot(c, run.Run.ID).Run.Status == store.StatusClosed }, "held review did not finish")
		assert.Len(t, runSnapshot(c, run.Run.ID).Steps, 1)
	})
	t.Run("drill exhausts immediately", func(t *testing.T) {
		c, legacy := offlineController(t)
		var run rpc.RunResult
		c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "qa-drill", Key: "drill-45", Vars: map[string]string{"issue": "45"}}, &run)
		step := waitStep(t, c, run.Run.ID, "fail")
		checkStep(t, legacy, "qa-drill", 0, step, map[string]any{"issue": "45"})
		c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: step.ID, Resolution: "failed"}, nil)
		c.PassUntil(func() bool {
			fields, err := store.DecodeFields[store.RunFields](runSnapshot(c, run.Run.ID).Run)
			require.NoError(t, err)
			return fields.State == store.RunNeedsHuman
		}, "exhausted native drill did not escalate")
	})
}

// Linear intake is deliberately not invoked: these cases execute the real hooks
// with recorded run inputs, without an external tracker or provider substitute.
func TestNativeLinearWorkflow(t *testing.T) {
	for _, tc := range []struct{ tracker, workflow, issue string }{
		{"linear", "pr-work", "ENG-48"},
		{"github", "pr-work-linear", "ENG-49"},
	} {
		t.Run(tc.workflow, func(t *testing.T) {
			root := packRoot(t, "qa-", tc.tracker, true)
			native, err := config.Load(root)
			require.NoError(t, err)
			legacy := loadLegacy(t, root, "qa-", tc.tracker)
			name := "qa-" + tc.workflow
			workflow := native.Workflows[name]
			before := legacy.Workflows[name]
			file, start, _ := strings.Cut(workflow.OnStart, ":")
			_, completed, _ := strings.Cut(workflow.OnStepCompleted, ":")
			entries := []string{filepath.Base(file) + ":" + start, filepath.Base(file) + ":" + completed}
			engine, err := automation.Load(t.Context(), filepath.Dir(file), entries, automation.Options{})
			require.NoError(t, err)
			inputs := maps.Clone(workflow.PackVars)
			inputs["issue"] = tc.issue
			run := hooks.Run{Workflow: name, Inputs: inputs, Steps: []hooks.Step{}}
			result, err := engine.Invoke(t.Context(), entries[0], run)
			require.NoError(t, err)
			for index, expected := range before.Steps {
				require.Len(t, result.Actions, 1)
				action := result.Actions[0]
				assert.Equal(t, "start_step", action.Kind)
				assert.Equal(t, expected.ID, action.ID)
				pool, err := native.ResolveHookReference("workflows."+name, "pool", action.Pool)
				require.NoError(t, err)
				wantPool, err := legacy.ResolvePool(&before, expected.Pool, inputs)
				require.NoError(t, err)
				assert.Equal(t, wantPool, pool)
				title, err := before.Render(expected.Title, inputs)
				require.NoError(t, err)
				assert.Equal(t, title, action.Title)
				raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), action.PromptFile))
				require.NoError(t, err)
				prompt, err := workflow.Render(string(raw), inputs)
				require.NoError(t, err)
				wantPrompt, err := before.Render(expected.Prompt, inputs)
				require.NoError(t, err)
				assert.Equal(t, wantPrompt, prompt)
				assert.Equal(t, before.StepNeeds(index), action.Needs)
				assert.Equal(t, expected.RetryLimit(), action.Retries)
				step := hooks.Step{ID: action.ID, Status: "closed", Resolution: "done", Outcome: "pr"}
				run.Steps = append(run.Steps, step)
				result, err = engine.Invoke(t.Context(), entries[1], run, step)
				require.NoError(t, err)
			}
			require.Len(t, result.Actions, 1)
			assert.Equal(t, "finish_run", result.Actions[0].Kind)
			assert.Equal(t, "done", result.Actions[0].Outcome)
		})
	}
}
