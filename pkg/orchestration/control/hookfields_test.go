package control_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// fieldsRoot is hookTestRoot with resources and a second pool, for
// start_step fields that mirror workflow step fields (JOB-182).
func fieldsRoot(t *testing.T, script string) config.Root {
	t.Helper()
	root := controltest.NewRoot(t, `pack: pack.yaml
projects: {js: {}}
agents: {fake: {command: "false"}}
resources:
  emu: {max: 1}
  pr: {keyed: true}
  flagship: {max: 1}
`)
	pack := `pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}
  big: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}
workflows:
  dynamic:
    project: js
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
`
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack.yaml"), []byte(strings.ReplaceAll(pack, "$ROOT", root.Dir)), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), []byte(script), 0o600))
	_, err := config.Load(root)
	require.NoError(t, err)
	return root
}

func hookStep(t *testing.T, c *controltest.Controller, run, name string) domain.Ticket {
	t.Helper()
	for _, s := range hookRun(t, c, run).Steps {
		if stepFields(t, s).Step == name {
			return s
		}
	}
	t.Fatalf("run %s has no step %s", run, name)
	return domain.Ticket{}
}

// TestHookStepNeedsResourcesWaitsOnCap: a Starlark-started step with
// needs_resources is held by the cap like a workflow step (#400): with the
// one emu slot claimed, the second step is not ready until the first closes.
func TestHookStepNeedsResourcesWaitsOnCap(t *testing.T) {
	t.Parallel()
	root := fieldsRoot(t, `def begin(run):
    return [
        start_step(id="first", pool="dev", prompt="one", needs_resources=["emu", "pr:acme#1"]),
        start_step(id="second", pool="dev", prompt="two", needs_resources=["emu"]),
    ]
def completed(run, step):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 2 }, "steps not created")
	first, second := hookStep(t, c, started.Run.ID, "first"), hookStep(t, c, started.Run.ID, "second")
	assert.Equal(t, []string{"emu", "pr:acme#1"}, stepFields(t, first).Resources)

	ready := func() []string {
		var ts []domain.Ticket
		c.Call(rpc.MethodReady, rpc.ReadyParams{Project: "js"}, &ts)
		var ids []string
		for _, x := range ts {
			ids = append(ids, x.ID)
		}
		return ids
	}
	assert.ElementsMatch(t, []string{first.ID, second.ID}, ready(), "both free before a claim")
	c.Call(rpc.MethodClaim, rpc.ClaimParams{ID: first.ID, Holder: "s1"}, nil)
	assert.Empty(t, ready(), "emu (max 1) is held by first")
	waits := hookRun(t, c, started.Run.ID).Waits
	require.Len(t, waits, 1)
	assert.Equal(t, second.ID, waits[0].Ticket)
	assert.Equal(t, "emu", waits[0].Key)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID, Resolution: "done"}, nil)
	assert.Equal(t, []string{second.ID}, ready(), "closing first frees the slot")
}

// TestHookStepResourceErrors: an unknown or wrongly shaped resource key
// fails the whole batch, as it fails a workflow start.
func TestHookStepResourceErrors(t *testing.T) {
	t.Parallel()
	for name, call := range map[string]string{
		"unknown": `start_step(id="a", pool="dev", prompt="x", needs_resources=["nope"])`,
		"keyed":   `start_step(id="a", pool="dev", prompt="x", needs_resources=["pr"])`,
		"counted": `start_step(id="a", pool="dev", prompt="x", needs_resources=["emu:1"])`,
		"placed":  `start_step(id="a", prompt="x", placement={"required": ["nope"], "options": [{"pool": "dev"}]})`,
		"pool":    `start_step(id="a", prompt="x", placement={"options": [{"pool": "missing"}]})`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := fieldsRoot(t, "def begin(run):\n    return "+call+"\ndef completed(run, step):\n    return []\n")
			c := controltest.Start(t, root, control.Options{})
			var started rpc.RunResult
			c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
			c.PassUntil(func() bool {
				f, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, started.Run.ID).Run)
				require.NoError(t, err)
				return f.State == domain.RunNeedsHuman
			}, "invalid step did not fail the batch")
			assert.Empty(t, hookRun(t, c, started.Run.ID).Steps)
		})
	}
}

// TestHookStepPlacement: placement starts the step on its top option and
// the controller falls back to the lower one once prefer_for has passed
// with the top option's resource held (#409).
func TestHookStepPlacement(t *testing.T) {
	t.Parallel()
	root := fieldsRoot(t, `def begin(run):
    return [
        start_step(id="hog", pool="big", prompt="hold flagship", needs_resources=["flagship"]),
        start_step(id="placed", prompt="work", placement={
            "required": ["pr:acme#1"],
            "prefer_for": "10m",
            "options": [{"pool": "dev", "weight": 10}, {"pool": "big", "resources": ["flagship"], "weight": 100}],
        }),
    ]
def completed(run, step):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 2 }, "steps not created")
	hog := hookStep(t, c, started.Run.ID, "hog")
	placed := hookStep(t, c, started.Run.ID, "placed")
	assert.Equal(t, "big", placed.Pool, "starts on the highest weight, not the first listed")
	sf := stepFields(t, placed)
	require.Len(t, sf.Placement.Options, 2)
	assert.Equal(t, []string{"pr:acme#1", "flagship"}, sf.Placement.Options[0].Resources)
	assert.Equal(t, []string{"pr:acme#1"}, sf.Placement.Options[1].Resources)
	assert.Equal(t, (10 * time.Minute).Milliseconds(), sf.Placement.PreferMS)

	c.Call(rpc.MethodClaim, rpc.ClaimParams{ID: hog.ID, Holder: "s1"}, nil)
	c.Pass(2)
	assert.Equal(t, "big", hookStep(t, c, started.Run.ID, "placed").Pool, "holds out during prefer_for")
	c.Clock.Advance(11 * time.Minute)
	c.PassUntil(func() bool { return hookStep(t, c, started.Run.ID, "placed").Pool == "dev" }, "falls back after prefer_for")
	assert.True(t, stepFields(t, hookStep(t, c, started.Run.ID, "placed")).Placement.Fallback())
}

// TestHookStepReviewFields: max_review_rounds and on_request_changes reach
// the step, and an explicit on_request_changes beats the previous step.
func TestHookStepReviewFields(t *testing.T) {
	t.Parallel()
	root := fieldsRoot(t, `def begin(run):
    return [
        start_step(id="impl", pool="dev", prompt="impl"),
        start_step(id="lint", pool="dev", prompt="lint", needs=["impl"]),
        start_step(id="review", pool="big", prompt="review", needs=["impl", "lint"],
                   on_request_changes="impl", max_review_rounds=1),
    ]
def completed(run, step):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 3 }, "steps not created")
	review := stepFields(t, hookStep(t, c, started.Run.ID, "review"))
	assert.Equal(t, "impl", review.OnRequestChanges, "explicit target, not the previous step lint")
	require.NotNil(t, review.MaxReviewRounds)
	assert.Equal(t, 1, review.ReviewRoundLimit())
	lint := stepFields(t, hookStep(t, c, started.Run.ID, "lint"))
	assert.Equal(t, "impl", lint.OnRequestChanges, "default stays the previous step")
	assert.Nil(t, lint.MaxReviewRounds)
}

// TestHookScriptStepRetryAfter: a script step that returns
// result(not_yet=True) is deferred by its retry_after without a failure.
func TestHookScriptStepRetryAfter(t *testing.T) {
	t.Parallel()
	root := fieldsRoot(t, `def begin(run):
    return start_step(id="poll", pool="dev", script=poll, retry_after="30m", retries=0)

def poll(run, step):
    return result(not_yet=True)

def completed(run, step):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &started)
	c.PassUntil(func() bool {
		for _, s := range hookRun(t, c, started.Run.ID).Steps {
			if stepFields(t, s).Result == "not_yet" && s.Status == domain.StatusOpen {
				return true
			}
		}
		return false
	}, "script step did not report not_yet")
	step := hookStep(t, c, started.Run.ID, "poll")
	sf := stepFields(t, step)
	assert.Equal(t, "30m", sf.RetryAfter)
	assert.Zero(t, sf.Failures, "not yet is not a failure")
	assert.Equal(t, domain.StatusOpen, step.Status)
	assert.Equal(t, controltest.Epoch.Add(30*time.Minute), step.DeferUntil.UTC(), "deferred by retry_after, not the 1m default")
}
