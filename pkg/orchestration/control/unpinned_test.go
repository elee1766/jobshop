package control_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Callback runs follow the live callbacks unless the workflow sets
// pin_callbacks (JOB-188). These tests change the callback source under an
// open run and reload.

const unpinnedV1 = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    if step.id == "a":
        return [start_step(id="b",pool="dev",prompt="second v1",needs=["a"])]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`

// deploy replaces the callback source and reloads, as a pack deploy does.
func deploy(t *testing.T, c *controltest.Controller, root config.Root, source string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), []byte(source), 0o600))
	c.Call(rpc.MethodReload, struct{}{}, nil)
}

func runFieldsOf(t *testing.T, c *controltest.Controller, id string) domain.RunFields {
	t.Helper()
	f, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, id).Run)
	require.NoError(t, err)
	return f
}

func commentsOf(t *testing.T, c *controltest.Controller, id string) []string {
	t.Helper()
	var shown rpc.ShowResult
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: id}, &shown)
	var out []string
	for _, cm := range shown.Comments {
		out = append(out, cm.Body)
	}
	return out
}

func countContaining(bodies []string, sub string) int {
	n := 0
	for _, b := range bodies {
		if strings.Contains(b, sub) {
			n++
		}
	}
	return n
}

// startOpenRun starts a run of the dynamic workflow and waits for its
// first step.
func startOpenRun(t *testing.T, c *controltest.Controller) (runID string, first domain.Ticket) {
	t.Helper()
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &run)
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) >= 1 }, "first step not created")
	return run.Run.ID, hookRun(t, c, run.Run.ID).Steps[0]
}

// By default an open run follows a deploy: the next step completes under the
// new code with no operator action, and the move is one run comment.
func TestUnpinnedRunFollowsLiveCallbacks(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, unpinnedV1)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)
	before := runFieldsOf(t, c, id).Automation.Digest

	deploy(t, c, root, strings.ReplaceAll(unpinnedV1, "second v1", "second v2"))
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 2 }, "run did not continue under the new callbacks")

	steps := hookRun(t, c, id).Steps
	assert.Equal(t, "second v2", steps[1].Body, "the new code made the second step")
	assert.Equal(t, domain.RunRunning, runFieldsOf(t, c, id).State)
	assert.NotEqual(t, before, runFieldsOf(t, c, id).Automation.Digest, "the run moved to the live manifest")

	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: steps[1].ID}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, id).Run.Status == domain.StatusClosed }, "run did not finish")
	assert.Equal(t, 1, countContaining(commentsOf(t, c, id), "Repinned to the live callback manifest"), "one comment for the move")
}

// An unpinned run keeps following: a second deploy is a second move, still
// one comment each.
func TestUnpinnedRunMovesOncePerDeploy(t *testing.T) {
	t.Parallel()
	const script = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    n = {"a": "b", "b": "c"}.get(step.id)
    if n:
        return [start_step(id=n,pool="dev",prompt="next v1",needs=[step.id])]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`
	root := hookTestRoot(t, script)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)
	deploy(t, c, root, strings.ReplaceAll(script, "next v1", "next v2"))
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 2 }, "second step not created")
	deploy(t, c, root, strings.ReplaceAll(script, "next v1", "next v3"))
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: hookRun(t, c, id).Steps[1].ID}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 3 }, "third step not created")
	assert.Equal(t, "next v3", hookRun(t, c, id).Steps[2].Body)
	assert.Equal(t, 2, countContaining(commentsOf(t, c, id), "Repinned to the live callback manifest"))
}

// A script step of an open run starts after the callback code changed and
// runs the new code.
func TestUnpinnedScriptStepRunsLiveCode(t *testing.T) {
	t.Parallel()
	const script = `def begin(run):
    return [
        start_step(id="gate",pool="dev",prompt="gate"),
        start_step(id="check",pool="dev",script=check,retries=0,needs=["gate"]),
    ]
def check(run, step):
    return result(outcome="pass", output={"version": "v1"})
def completed(run, step):
    if step.id == "gate":
        return []
    return [finish_run(outcome=step.outcome, output=step.output)]
def handle(event, args):
    return []
`
	root := hookTestRoot(t, script)
	c := controltest.Start(t, root, control.Options{})
	id, gate := startOpenRun(t, c)
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 2 }, "script step not created")

	deploy(t, c, root, strings.ReplaceAll(script, `"v1"`, `"v2"`))
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: gate.ID}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, id).Run.Status == domain.StatusClosed }, "script step did not run under the new code")

	f := runFieldsOf(t, c, id)
	assert.Equal(t, domain.RunDone, f.State)
	assert.Equal(t, "v2", f.Automation.Output["version"])
	assert.Equal(t, 1, countContaining(commentsOf(t, c, id), "Repinned to the live callback manifest"))
}

// An edit the run's state does not fit fails in the callback: the run goes
// needs_human with the callback's error, not "hook manifest changed", and
// plain resume continues it once the callback is fixed.
func TestUnpinnedIncompatibleEditFailsInCallback(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, unpinnedV1)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)

	broken := strings.Replace(unpinnedV1, `    if step.id == "a":`, `    fail("unknown step " + step.id)
    if step.id == "a":`, 1)
	deploy(t, c, root, broken)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { return runFieldsOf(t, c, id).State == domain.RunNeedsHuman }, "incompatible callback did not stop the run")
	comments := commentsOf(t, c, id)
	assert.Equal(t, 1, countContaining(comments, "unknown step a"), "the callback's error is on the run: %v", comments)
	assert.Zero(t, countContaining(comments, "manifest changed"), "not a pin failure: %v", comments)
	assert.Len(t, hookRun(t, c, id).Steps, 1)

	deploy(t, c, root, unpinnedV1)
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: id}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 2 }, "plain resume did not continue the fixed run")
	assert.Equal(t, domain.RunRunning, runFieldsOf(t, c, id).State)
}

// A retried invocation whose calls no longer match its journal still fails
// with the reuse error: the replay check is not a pin.
func TestUnpinnedReplayStillRefusesChangedCalls(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	const script = `def begin(run):
    response = http.get(run.inputs["url"])
    return [start_step(id="work",pool="dev",prompt_file="retry.md",inputs={"ok":response.json()["ok"]})]
def completed(run, step):
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`
	root := hookTestRoot(t, script)
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte("automation: {http: true}\n")...), 0o600))
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"url": server.URL}}, &run)
	c.PassUntil(func() bool { return runFieldsOf(t, c, run.Run.ID).State == domain.RunNeedsHuman }, "missing prompt did not stop the hook")

	// The edit asks a different URL, so the recorded call does not match.
	deploy(t, c, root, strings.Replace(script, `http.get(run.inputs["url"])`, `http.get(run.inputs["url"] + "/changed")`, 1))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "retry.md"), []byte("HTTP returned {{.ok}}"), 0o600))
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: run.Run.ID}, nil)
	// replayHookCall refuses the recorded call (its input digest differs);
	// the capability layer reports that as a failed recorded-result lookup.
	c.PassUntil(func() bool {
		return countContaining(commentsOf(t, c, run.Run.ID), "recorded-result lookup failed") == 1
	}, "changed calls did not fail with the reuse error")
	assert.Equal(t, domain.RunNeedsHuman, runFieldsOf(t, c, run.Run.ID).State)
	assert.Empty(t, hookRun(t, c, run.Run.ID).Steps)
	assert.EqualValues(t, 1, requests.Load(), "the changed call was not sent")
}

// With pin_callbacks the same deploy stops the run, plain resume refuses and
// the admin socket's --repin continues it.
func TestPinnedRunStopsOnDeploy(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, unpinnedV1)
	pinCallbacks(t, root)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)

	deploy(t, c, root, strings.ReplaceAll(unpinnedV1, "second v1", "second v2"))
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { return runFieldsOf(t, c, id).State == domain.RunNeedsHuman }, "pinned run did not stop")
	comments := commentsOf(t, c, id)
	assert.Equal(t, 1, countContaining(comments, "hook manifest changed"), "%v", comments)
	assert.Zero(t, countContaining(comments, "Repinned"))

	require.Error(t, c.Admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: id}, nil), "plain resume must not adopt the changed manifest")
	assert.Equal(t, domain.RunNeedsHuman, runFieldsOf(t, c, id).State)

	require.NoError(t, c.Admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: id, Repin: true}, nil))
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 2 }, "repinned run did not continue")
	assert.Equal(t, "second v2", hookRun(t, c, id).Steps[1].Body)
}

// A run stranded by a changed manifest before the option existed (here: a
// pinned workflow's run, then the option removed) resumes with plain
// `run resume`.
func TestStrandedRunResumesWithoutRepin(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, unpinnedV1)
	pinCallbacks(t, root)
	c := controltest.Start(t, root, control.Options{})
	id, first := startOpenRun(t, c)
	deploy(t, c, root, strings.ReplaceAll(unpinnedV1, "second v1", "second v2"))
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool { return runFieldsOf(t, c, id).State == domain.RunNeedsHuman }, "pinned run did not stop")

	// Turning the option off releases the run.
	path := filepath.Join(root.Dir, "pack.yaml")
	pack, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(pack), "    pin_callbacks: true\n", "", 1)), 0o600))
	c.Call(rpc.MethodReload, struct{}{}, nil)
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: id}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, id).Steps) == 2 }, "stranded run did not continue")
	assert.Equal(t, domain.RunRunning, runFieldsOf(t, c, id).State)
	assert.Equal(t, "second v2", hookRun(t, c, id).Steps[1].Body)
}

// An unpinned run uses the live host grants and reference scope after a
// reload, widened or narrowed, with no operator action. The run's recorded
// scope moves with it.
func TestUnpinnedRunFollowsLiveAuthority(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	const script = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    if step.id == "a":
        ok = http.get(run.inputs["url"]).json()["ok"]
        return [start_step(id="b",pool=run.inputs["next"],prompt="second",needs=["a"],inputs={"ok":ok})]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`
	pool := func(name string) string {
		return "  " + name + `: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}` + "\n"
	}
	grant := "automation: {http: true}\n"
	edit := func(t *testing.T, root config.Root, file, from, to string) {
		t.Helper()
		path := filepath.Join(root.Dir, file)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Contains(t, string(data), from)
		require.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(data), from, to, 1)), 0o600))
	}
	for name, tc := range map[string]struct {
		next    string
		before  string // pools and grants present when the run starts
		granted bool
		change  func(t *testing.T, root config.Root)
		failure string // empty: the run continues
	}{
		"grant and pool added": {next: "later", change: func(t *testing.T, root config.Root) {
			edit(t, root, "pack.yaml", "pools:\n", "pools:\n"+strings.ReplaceAll(pool("later"), "$ROOT", root.Dir))
			data, err := os.ReadFile(root.Config())
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(root.Config(), append(data, []byte(grant)...), 0o600))
		}},
		"pool removed": {next: "spare", granted: true, before: pool("spare"), change: func(t *testing.T, root config.Root) {
			edit(t, root, "pack.yaml", strings.ReplaceAll(pool("spare"), "$ROOT", root.Dir), "")
		}, failure: "spare"},
		"grant removed": {next: "dev", granted: true, change: func(t *testing.T, root config.Root) {
			edit(t, root, "jobshop.yaml", grant, "")
		}, failure: "http"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := hookTestRoot(t, script)
			edit(t, root, "pack.yaml", "pools:\n", "pools:\n"+strings.ReplaceAll(tc.before, "$ROOT", root.Dir))
			if tc.granted {
				data, err := os.ReadFile(root.Config())
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(root.Config(), append(data, []byte(grant)...), 0o600))
			}
			c := controltest.Start(t, root, control.Options{})
			var run rpc.RunResult
			c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"url": server.URL, "next": tc.next}}, &run)
			c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 1 }, "first step not created")
			first := hookRun(t, c, run.Run.ID).Steps[0]
			startScope := runFieldsOf(t, c, run.Run.ID).Automation.Scope

			tc.change(t, root)
			c.Call(rpc.MethodReload, struct{}{}, nil)
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
			if tc.failure == "" {
				c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 2 }, "run did not use the widened authority")
				got := hookRun(t, c, run.Run.ID).Steps[1]
				assert.Equal(t, "later", got.Pool)
				assert.Contains(t, runFieldsOf(t, c, run.Run.ID).Automation.Scope.Pools, "later", "the recorded scope moved")
				assert.NotContains(t, startScope.Pools, "later")
				return
			}
			c.PassUntil(func() bool { return runFieldsOf(t, c, run.Run.ID).State == domain.RunNeedsHuman }, "narrowed authority did not stop the run")
			comments := commentsOf(t, c, run.Run.ID)
			assert.Equal(t, 1, countContaining(comments, tc.failure), "%v", comments)
			assert.Zero(t, countContaining(comments, "manifest changed"), "%v", comments)
			assert.Len(t, hookRun(t, c, run.Run.ID).Steps, 1)
			if strings.Contains(name, "pool") {
				assert.NotContains(t, runFieldsOf(t, c, run.Run.ID).Automation.Scope.Pools, "spare", "the recorded scope narrowed")
			}
		})
	}
}
