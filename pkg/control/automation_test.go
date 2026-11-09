package control_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func hookTestRoot(t *testing.T, script string) config.Root {
	t.Helper()
	root := controltest.NewRoot(t, `pack: pack.yaml
projects: {js: {}}
agents: {fake: {command: "false"}}
`)
	pack := `pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}
workflows:
  dynamic:
    project: js
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
triggers:
  custom:
    on: [{source: generic, event: tasks}]
    handler: hooks.star:handle
`
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack.yaml"), []byte(strings.ReplaceAll(pack, "$ROOT", root.Dir)), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), []byte(script), 0o600))
	_, err := config.Load(root)
	require.NoError(t, err)
	return root
}

func TestHookCapabilitiesOutsideControllerLocksAndCancelDiscardsActions(t *testing.T) {
	t.Parallel()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		_, _ = w.Write([]byte(`{"route":"small"}`))
	}))
	defer server.Close()
	defer unblock()
	root := hookTestRoot(t, `def begin(run):
    response = http.get(run.inputs["url"])
    return [start_step(id="work",pool="dev",prompt="Must never start after cancellation")]
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	// The timeout is wall time. 10s ran out under a loaded parallel run
	// before the cancel landed, failing the hook instead (JOB-117).
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte("automation: {http: true, timeout: 1m}\n")...), 0o600))
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"url": server.URL}}, &run)
	select {
	case <-entered:
	case <-time.After(controltest.DialWait):
		t.Fatal("hook did not issue its HTTP request")
	}
	c.Pass(1)
	c.Call(rpc.MethodRunCancel, rpc.IDParams{ID: run.Run.ID}, nil)
	unblock()
	var audit []store.Record
	c.PassUntil(func() bool {
		c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &audit)
		return len(audit) == 1 && audit[0].State == "superseded"
	}, "cancelled hook applied actions or did not finish")
	assert.Empty(t, hookRun(t, c, run.Run.ID).Steps)
	var trace rpc.AutomationResult
	c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: audit[0].Key}, &trace)
	require.Len(t, trace.Calls, 1)
	assert.Equal(t, "completed", trace.Calls[0].State)
	var call map[string]any
	require.NoError(t, json.Unmarshal(trace.Calls[0].Value, &call))
	assert.Equal(t, audit[0].Key, call["invocation"])
	observation, ok := call["observation"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http.get", observation["capability"])
}

const routingHooks = `def begin(run):
    return [start_step(id="work", pool="dev", prompt="Work " + run.inputs["task"], inputs=run.inputs, retries=2, output_schema={"type":"object", "required":["route"], "properties":{"route":{"enum":["small","large"]}}, "additionalProperties":False})]

def completed(run, step):
    if step.id == "work":
        return [start_step(id=step.output["route"], pool="dev", prompt="Review selected route", needs=["work"])]
    return [finish_run(outcome="done", output={"selected":step.id, "data":run.inputs.get("data", {})})]

def handle(event, args):
    return [start_run(workflow="dynamic", key=event.payload["key"], inputs={"task":event.payload["task"]})]
`

func hookRun(t *testing.T, c *controltest.Controller, id string) rpc.RunResult {
	t.Helper()
	var result rpc.RunResult
	c.Call(rpc.MethodRunShow, rpc.IDParams{ID: id}, &result)
	return result
}

func TestHookWorkflowRoutesValidatedOutputAndSurvivesRestart(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, routingHooks)
	c := controltest.Start(t, root, control.Options{})
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Key: "task-1", Inputs: map[string]any{"task": "nested input", "data": map[string]any{"count": int64(9007199254740993)}}}, &started)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 1 }, "start hook did not create work")
	run := hookRun(t, c, started.Run.ID)
	assert.Equal(t, domain.StatusOpen, run.Run.Status)
	assert.Equal(t, "Work nested input", run.Steps[0].Body)
	work := run.Steps[0]
	require.Contains(t, string(work.Fields), "9007199254740993", "structured inputs must survive durable invocation and action planning")
	var invalid domain.Ticket
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: work.ID, Output: map[string]any{"route": "invalid"}}, &invalid)
	assert.Equal(t, domain.StatusOpen, invalid.Status)
	f, err := domain.DecodeFields[domain.StepFields](invalid)
	require.NoError(t, err)
	assert.Equal(t, 1, f.Failures)
	assert.Nil(t, f.StructuredOutput)
	c.Pass(2)
	require.Len(t, hookRun(t, c, started.Run.ID).Steps, 1)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: work.ID, Outcome: "selected", Output: map[string]any{"route": "small"}}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, started.Run.ID).Steps) == 2 }, "completion hook did not route")
	run = hookRun(t, c, started.Run.ID)
	require.Equal(t, domain.StatusOpen, run.Run.Status)
	var selected domain.Ticket
	for _, step := range run.Steps {
		sf, err := domain.DecodeFields[domain.StepFields](step)
		require.NoError(t, err)
		assert.NotEqual(t, "large", sf.Step)
		if sf.Step == "small" {
			selected = step
		}
	}
	require.NotEmpty(t, selected.ID)
	c.Stop()
	c = controltest.Start(t, root, control.Options{})
	c.Pass(2)
	require.Len(t, hookRun(t, c, started.Run.ID).Steps, 2)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: selected.ID, Output: map[string]any{"summary": "reviewed"}}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, started.Run.ID).Run.Status == domain.StatusClosed }, "explicit finish action did not close run")
	run = hookRun(t, c, started.Run.ID)
	rf, err := domain.DecodeFields[domain.RunFields](run.Run)
	require.NoError(t, err)
	assert.Equal(t, "small", rf.Automation.Output["selected"])
	require.Contains(t, string(run.Run.Fields), "9007199254740993", "structured data must remain exact through restart and finish actions")
	assert.Len(t, rf.Automation.Applied, 3)
	var audit []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &audit)
	require.Len(t, audit, 3)
	for _, record := range audit {
		assert.Equal(t, "done", record.State)
		var trace rpc.AutomationResult
		c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: record.Key}, &trace)
		require.NotEmpty(t, trace.Attempts)
	}
}

func TestHookInvalidLaterActionDoesNotCreateEarlierStep(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return [start_step(id="valid",pool="dev",prompt="Do work"), start_step(id="invalid",pool="missing",prompt="Must reject whole batch")]
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &run)
	c.PassUntil(func() bool {
		f, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, run.Run.ID).Run)
		require.NoError(t, err)
		return f.State == domain.RunNeedsHuman
	}, "invalid batch did not fail")
	assert.Empty(t, hookRun(t, c, run.Run.ID).Steps)
	var audit []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &audit)
	require.Len(t, audit, 1)
	assert.Equal(t, "failed", audit[0].State)
	var value map[string]any
	require.NoError(t, json.Unmarshal(audit[0].Value, &value))
	assert.Contains(t, value["error"], "missing")
}

func TestHookResumeReusesCompletedCallAfterPromptRepair(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	root := hookTestRoot(t, `def begin(run):
    response = http.get(run.inputs["url"])
    return [start_step(id="work",pool="dev",prompt_file="retry.md",inputs={"ok":response.json()["ok"]})]
def completed(run, step):
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`)
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte("automation: {http: true}\n")...), 0o600))
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"url": server.URL}}, &run)
	c.PassUntil(func() bool {
		fields, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, run.Run.ID).Run)
		require.NoError(t, err)
		return fields.State == domain.RunNeedsHuman
	}, "missing prompt did not stop hook")
	c.Stop()
	c = controltest.Start(t, root, control.Options{})
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "retry.md"), []byte("HTTP returned {{.ok}}"), 0o600))
	c.Call(rpc.MethodRunResume, rpc.RunResumeParams{ID: run.Run.ID}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 1 }, "resumed hook did not create step")
	assert.EqualValues(t, 1, requests.Load())
	var records []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &records)
	require.Len(t, records, 1)
	assert.Equal(t, "done", records[0].State)
	var audit rpc.AutomationResult
	c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: records[0].Key}, &audit)
	assert.Len(t, audit.Attempts, 2)
	assert.Len(t, audit.Calls, 1)
	var attempt struct {
		ReusedCalls []struct {
			SourceAttempt int64 `json:"source_attempt"`
			Index         int   `json:"index"`
		} `json:"reused_calls"`
	}
	require.NoError(t, json.Unmarshal(audit.Attempts[1].Value, &attempt))
	require.Len(t, attempt.ReusedCalls, 1)
	assert.EqualValues(t, 1, attempt.ReusedCalls[0].SourceAttempt)
	assert.Equal(t, 1, attempt.ReusedCalls[0].Index)
}

func TestHookInterruptedExternalCallRemainsUnknownAcrossRestart(t *testing.T) {
	t.Parallel()
	entered := make(chan struct{})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			close(entered)
		}
		<-r.Context().Done()
	}))
	defer server.Close()
	root := hookTestRoot(t, `def begin(run):
    http.get(run.inputs["url"])
    return [start_step(id="work",pool="dev",prompt="Must not repeat unknown effect")]
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte("automation: {http: true, timeout: 1m}\n")...), 0o600))
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"url": server.URL}}, &run)
	select {
	case <-entered:
	case <-time.After(controltest.DialWait):
		t.Fatal("hook did not begin external effect")
	}
	c.Stop()
	c = controltest.Start(t, root, control.Options{})
	c.PassUntil(func() bool {
		fields, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, run.Run.ID).Run)
		require.NoError(t, err)
		return fields.State == domain.RunNeedsHuman
	}, "unknown effect was not surfaced after restart")
	assert.Empty(t, hookRun(t, c, run.Run.ID).Steps)
	assert.EqualValues(t, 1, requests.Load())
	var records []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &records)
	require.Len(t, records, 1)
	var audit rpc.AutomationResult
	c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: records[0].Key}, &audit)
	require.Len(t, audit.Calls, 1)
	assert.Equal(t, "UNKNOWN", audit.Calls[0].State)
	err = c.Admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: run.Run.ID}, nil)
	require.ErrorContains(t, err, "reconcile")
	assert.EqualValues(t, 1, requests.Load())
}

func TestHookAdmittedGenericDeliveryDeduplicatesAcrossRestart(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, routingHooks)
	c := controltest.Start(t, root, control.Options{})
	spoolPath := filepath.Join(root.Dir, "events.db")
	spool, err := ingest.OpenSpool(t.Context(), spoolPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = spool.Close() })
	endpoint := config.IngestEndpoint{Name: "tasks", Source: "generic", Map: config.IngestMap{Kind: "event"}}
	payload := []byte(`{"event":"tasks","key":"widget-1","task":"Add a verbose flag"}`)
	put := func() {
		_, err := ingest.Put(t.Context(), spool, endpoint, "delivery-1", "", payload, controltest.Epoch)
		require.NoError(t, err)
		c.Call(rpc.MethodIngestDrain, rpc.IngestDrainParams{Spool: spoolPath}, nil)
	}
	put()
	var runs []domain.Ticket
	c.PassUntil(func() bool {
		c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
		return len(runs) == 1 && len(hookRun(t, c, runs[0].ID).Steps) == 1
	}, "admitted generic event did not start workflow")
	c.Stop()
	c = controltest.Start(t, root, control.Options{})
	put()
	c.Pass(3)
	c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
	require.Len(t, runs, 1)
	var invocations []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Trigger: "custom"}, &invocations)
	require.Len(t, invocations, 1)
	assert.Equal(t, "done", invocations[0].State)
	var provenance map[string]map[string]any
	require.NoError(t, json.Unmarshal(runs[0].Meta, &provenance))
	assert.Equal(t, invocations[0].Key, provenance["automation"]["invocation"])
}

func TestHookDocumentedCUEWorkflowFixture(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, "")
	require.NoError(t, os.Remove(root.Config()))
	require.NoError(t, os.CopyFS(root.Dir, os.DirFS("../config/testdata/cue-starlark-example")))
	_, err := config.Load(root)
	require.NoError(t, err)
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "demo-hooked", Vars: map[string]string{"task": "Add a verbose flag", "pool": "demo-dev"}}, &run)
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 1 }, "documented fixture did not create implementation step")
	step := hookRun(t, c, run.Run.ID).Steps[0]
	assert.Contains(t, step.Body, "Add a verbose flag")
	assert.NotContains(t, step.Body, "{{")
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: step.ID, Outcome: "implemented", Output: map[string]any{"summary": "Verbose flag added"}}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, run.Run.ID).Run.Status == domain.StatusClosed }, "documented completion callback did not finish run")
	fields, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, run.Run.ID).Run)
	require.NoError(t, err)
	assert.Equal(t, "Verbose flag added", fields.Automation.Output["summary"])
}

func TestHookTimerEmptyActionsRetainScheduleAcrossRestart(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return []
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	path := filepath.Join(root.Dir, "pack.yaml")
	pack, err := os.ReadFile(path)
	require.NoError(t, err)
	pack = []byte(strings.Replace(string(pack), "    on: [{source: generic, event: tasks}]", "    when: {cron: '*/5 * * * *'}", 1))
	require.NoError(t, os.WriteFile(path, pack, 0o600))
	c := controltest.Start(t, root, control.Options{})
	c.Clock.Advance(5 * time.Minute)
	var records []store.Record
	read := func() { c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Trigger: "custom"}, &records) }
	c.PassUntil(func() bool { read(); return len(records) == 1 && records[0].State == "done" }, "timer did not invoke empty hook")
	c.Clock.Advance(time.Minute)
	c.Pass(2)
	read()
	require.Len(t, records, 1)
	clock := c.Clock
	c.Stop()
	c = controltest.Start(t, root, control.Options{Clock: clock})
	c.Pass(2)
	read()
	require.Len(t, records, 1)
	clock.Advance(4 * time.Minute)
	c.PassUntil(func() bool { read(); return len(records) == 2 && records[0].State == "done" }, "restart lost timer cadence")
}

func TestHookBatchOrderBlocksDynamicSourceSteps(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return [start_step(id="work",pool="dev",prompt="Issue " + run.inputs["issue"])]
def completed(run, step):
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`)
	path := filepath.Join(root.Dir, "pack.yaml")
	pack, err := os.ReadFile(path)
	require.NoError(t, err)
	pack = []byte(strings.Replace(string(pack), "  dynamic:\n", "  dynamic:\n    vars: [issue]\n", 1))
	require.NoError(t, os.WriteFile(path, pack, 0o600))
	c := controltest.Start(t, root, control.Options{})
	var batch rpc.BatchResult
	c.Call(rpc.MethodBatchStart, rpc.BatchStartParams{Workflow: "dynamic", Key: "ordered", Issues: []string{"1", "2"}, Order: []string{"1>2"}}, &batch)
	require.Len(t, batch.Runs, 2)
	c.PassUntil(func() bool {
		return len(hookRun(t, c, batch.Runs[0].Run.ID).Steps) == 1 && len(hookRun(t, c, batch.Runs[1].Run.ID).Steps) == 1
	}, "batch hooks did not create source steps")
	first := hookRun(t, c, batch.Runs[0].Run.ID).Steps[0]
	second := hookRun(t, c, batch.Runs[1].Run.ID).Steps[0]
	var ready []domain.Ticket
	c.Call(rpc.MethodReady, rpc.ReadyParams{Pool: "dev"}, &ready)
	require.Len(t, ready, 1)
	assert.Equal(t, first.ID, ready[0].ID)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	c.PassUntil(func() bool {
		c.Call(rpc.MethodReady, rpc.ReadyParams{Pool: "dev"}, &ready)
		return len(ready) == 1 && ready[0].ID == second.ID
	}, "explicit completion did not release next run")
	err = c.Admin.Call(t.Context(), rpc.MethodBatchStart, rpc.BatchStartParams{Workflow: "dynamic", Key: "cycle", Issues: []string{"3", "4"}, Order: []string{"3>4", "4>3"}}, nil)
	require.Error(t, err)
}

func TestHookReviewLoopReusesStepsAndClearsPriorOutput(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return [start_step(id="work",pool="dev",prompt="Implement")]
def completed(run, step):
    if step.id == "work":
        for existing in run.steps:
            if existing.id == "review":
                return []
        return [start_step(id="review",pool="dev",prompt="Review",needs=["work"])]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &run)
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 1 }, "work step missing")
	work := hookRun(t, c, run.Run.ID).Steps[0]
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: work.ID, Output: map[string]any{"summary": "first attempt"}}, nil)
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 2 }, "review step missing")
	review := hookRun(t, c, run.Run.ID).Steps[1]
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: review.ID, Resolution: domain.ResolutionRequestChanges, Comment: "Add the missing validation"}, nil)
	current := hookRun(t, c, run.Run.ID)
	assert.Equal(t, domain.StatusOpen, current.Steps[0].Status)
	assert.Equal(t, domain.StatusOpen, current.Steps[1].Status)
	fields, err := domain.DecodeFields[domain.StepFields](current.Steps[0])
	require.NoError(t, err)
	assert.Nil(t, fields.StructuredOutput)
	require.NotNil(t, fields.Changes)
	assert.Contains(t, fields.Changes.Findings, "missing validation")
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: work.ID, Output: map[string]any{"summary": "validation added"}}, nil)
	c.PassUntil(func() bool {
		var records []store.Record
		c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID, State: "done"}, &records)
		return len(records) == 3
	}, "revised work callback did not complete")
	assert.Len(t, hookRun(t, c, run.Run.ID).Steps, 2)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: review.ID}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, run.Run.ID).Run.Status == domain.StatusClosed }, "review did not finish run")
}

// A callback run pins its code, grants and the references accepted at start,
// not the rest of the configuration, which a live pack checkout reloads often.
func TestHookRunPinSurvivesUnrelatedReloadOnly(t *testing.T) {
	t.Parallel()
	const script = `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    if step.id == "a":
        return [start_step(id="b",pool=run.inputs["next"],prompt="second",needs=["a"])]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`
	pool := func(name string) string {
		return "  " + name + `: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}` + "\n"
	}
	for name, tc := range map[string]struct {
		next   string
		before string // pools present when the run starts
		edit   func(t *testing.T, root config.Root, pack string) string
		done   bool
		// pinChanged: resume is refused. Otherwise it retries and the
		// pinned reference scope rejects the action again.
		pinChanged bool
	}{
		"unrelated pool and workflow": {next: "dev", done: true, edit: func(_ *testing.T, _ config.Root, pack string) string {
			pack = strings.Replace(pack, "pools:\n", "pools:\n"+pool("other"), 1)
			return strings.Replace(pack, "workflows:\n", "workflows:\n  other-flow:\n    steps: [{id: x, pool: other, title: x, prompt: x}]\n", 1)
		}},
		"pool added after start": {next: "later", edit: func(_ *testing.T, _ config.Root, pack string) string {
			return strings.Replace(pack, "pools:\n", "pools:\n"+pool("later"), 1)
		}},
		"pool removed after start": {next: "spare", before: pool("spare"), edit: func(_ *testing.T, root config.Root, pack string) string {
			return strings.Replace(pack, strings.ReplaceAll(pool("spare"), "$ROOT", root.Dir), "", 1)
		}},
		"callback source": {next: "dev", pinChanged: true, edit: func(t *testing.T, root config.Root, pack string) string {
			path := filepath.Join(root.Dir, "hooks.star")
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, append(data, []byte("# changed\n")...), 0o600))
			return pack
		}},
		"capability grant": {next: "dev", pinChanged: true, edit: func(t *testing.T, root config.Root, pack string) string {
			data, err := os.ReadFile(root.Config())
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(root.Config(), append(data, []byte("automation: {http: true}\n")...), 0o600))
			return pack
		}},
	} {
		t.Run(name, func(t *testing.T) {
			root := hookTestRoot(t, script)
			path := filepath.Join(root.Dir, "pack.yaml")
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			pack := strings.Replace(string(data), "pools:\n", "pools:\n"+strings.ReplaceAll(tc.before, "$ROOT", root.Dir), 1)
			require.NoError(t, os.WriteFile(path, []byte(pack), 0o600))
			c := controltest.Start(t, root, control.Options{})
			var run rpc.RunResult
			c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"next": tc.next}}, &run)
			c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 1 }, "first step not created")
			first := hookRun(t, c, run.Run.ID).Steps[0]
			pack = strings.ReplaceAll(tc.edit(t, root, pack), "$ROOT", root.Dir)
			require.NoError(t, os.WriteFile(path, []byte(pack), 0o600))
			c.Call(rpc.MethodReload, struct{}{}, nil)
			c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
			if tc.done {
				c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 2 }, "second step not created after reload")
				c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: hookRun(t, c, run.Run.ID).Steps[1].ID}, nil)
				c.PassUntil(func() bool { return hookRun(t, c, run.Run.ID).Run.Status == domain.StatusClosed }, "run did not finish after unrelated reload")
				return
			}
			needsHuman := func() bool {
				f, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, run.Run.ID).Run)
				require.NoError(t, err)
				return f.State == domain.RunNeedsHuman
			}
			c.PassUntil(needsHuman, "changed authority did not stop the run")
			assert.Len(t, hookRun(t, c, run.Run.ID).Steps, 1)
			err = c.Admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: run.Run.ID}, nil)
			if tc.pinChanged {
				assert.Error(t, err, "resume must not adopt the changed pin")
				return
			}
			require.NoError(t, err)
			c.PassUntil(needsHuman, "retried action outside the accepted scope did not fail again")
			assert.Len(t, hookRun(t, c, run.Run.ID).Steps, 1)
		})
	}
}

// TestHookGitHubUsesIdentity: the github built-in authenticates with the
// identity covering the repo, read per call, and journals the call.
func TestHookGitHubUsesIdentity(t *testing.T) {
	t.Parallel()
	var auth atomic.Value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.Store(r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"workflow_runs":[{"head_sha":"abc123","conclusion":"failure"}]}`))
	}))
	defer server.Close()
	root := hookTestRoot(t, `def begin(run):
    return []
def completed(run, step):
    return []
def handle(event, args):
    runs = github.get("/repos/acme/widget/actions/runs", params={"branch": "master", "per_page": 1}).json()["workflow_runs"]
    return [start_run(workflow="dynamic", key="red-" + runs[0]["head_sha"])]
`)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "gh-token"), []byte("ghp_first\n"), 0o600))
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	extra := "automation: {github: true}\nidentities:\n  bot: {kind: github-token, token_file: " + filepath.Join(root.Dir, "gh-token") + ", api_url: " + server.URL + ", repos: [acme/widget]}\n"
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte(extra)...), 0o600))
	c := controltest.Start(t, root, control.Options{})
	spoolPath := filepath.Join(root.Dir, "events.db")
	spool, err := ingest.OpenSpool(t.Context(), spoolPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = spool.Close() })
	endpoint := config.IngestEndpoint{Name: "tasks", Source: "generic", Map: config.IngestMap{Kind: "event"}}
	put := func(id string) {
		_, err := ingest.Put(t.Context(), spool, endpoint, id, "", []byte(`{"event":"tasks"}`), controltest.Epoch)
		require.NoError(t, err)
		c.Call(rpc.MethodIngestDrain, rpc.IngestDrainParams{Spool: spoolPath}, nil)
	}
	put("d1")
	var runs []domain.Ticket
	c.PassUntil(func() bool {
		c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
		return len(runs) == 1
	}, "handler did not start a run from the GitHub response")
	rf, err := domain.DecodeFields[domain.RunFields](runs[0])
	require.NoError(t, err)
	assert.Equal(t, "red-abc123", rf.Key)
	assert.Equal(t, "Bearer ghp_first", auth.Load())
	// A rotated token file is picked up by the next call.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "gh-token"), []byte("ghp_second\n"), 0o600))
	put("d2")
	c.PassUntil(func() bool { return auth.Load() == "Bearer ghp_second" }, "token was not read per call")
	var invocations []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Trigger: "custom"}, &invocations)
	require.NotEmpty(t, invocations)
	var trace rpc.AutomationResult
	c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: invocations[len(invocations)-1].Key}, &trace)
	require.Len(t, trace.Calls, 1)
	assert.Contains(t, string(trace.Calls[0].Value), `"capability":"github.get"`)
	assert.NotContains(t, string(trace.Calls[0].Value), "ghp_")
}

// TestHookReadsAndTicketActions: a handler reads the store, files a ticket
// once per key and comments on an existing one, and time.now() is the
// invocation's accepted time.
func TestHookReadsAndTicketActions(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return []
def completed(run, step):
    return []
def handle(event, args):
    open = tickets(kind="task", status="open", label="watch")
    actions = [create_ticket(title="Plan next step", key="replan", pool="dev", labels=["plan"], priority=1, body="filed at " + str(time.now().unix))]
    for t in open:
        actions.append(comment(ticket=t["id"], body="seen by " + event.event))
    return result(actions=actions, output={"open": len(open), "runs": len(runs(workflow="dynamic"))})
`)
	c := controltest.Start(t, root, control.Options{})
	var watched domain.Ticket
	c.Call(rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "watched", Labels: []string{"watch"}}, &watched)
	spoolPath := filepath.Join(root.Dir, "events.db")
	spool, err := ingest.OpenSpool(t.Context(), spoolPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = spool.Close() })
	endpoint := config.IngestEndpoint{Name: "tasks", Source: "generic", Map: config.IngestMap{Kind: "event"}}
	put := func(id string) {
		_, err := ingest.Put(t.Context(), spool, endpoint, id, "", []byte(`{"event":"tasks"}`), controltest.Epoch)
		require.NoError(t, err)
		c.Call(rpc.MethodIngestDrain, rpc.IngestDrainParams{Spool: spoolPath}, nil)
	}
	var records []store.Record
	done := func(n int) func() bool {
		return func() bool {
			c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Trigger: "custom"}, &records)
			finished := 0
			for _, r := range records {
				if r.State == "done" {
					finished++
				}
			}
			return finished == n
		}
	}
	put("d1")
	c.PassUntil(done(1), "handler did not finish")
	put("d2")
	c.PassUntil(done(2), "second handler did not finish")
	var planned []domain.Ticket
	c.Call(rpc.MethodTicketList, rpc.ListParams{Label: "plan"}, &planned)
	require.Len(t, planned, 1, "create_ticket must dedup on its key while the ticket is open")
	assert.Equal(t, "dev", planned[0].Pool)
	assert.Equal(t, 1, planned[0].Priority)
	assert.Contains(t, planned[0].Body, "filed at ")
	assert.Contains(t, string(planned[0].Meta), `"trigger":{"key":"custom/replan","name":"custom"}`)
	var thread rpc.ShowResult
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: watched.ID}, &thread)
	require.Len(t, thread.Comments, 2)
	assert.Equal(t, "seen by tasks", thread.Comments[0].Body)
	var trace rpc.AutomationResult
	c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: records[0].Key}, &trace)
	capabilities := []string{}
	for _, call := range trace.Calls {
		var v struct {
			Observation struct {
				Capability string `json:"capability"`
			} `json:"observation"`
		}
		require.NoError(t, json.Unmarshal(call.Value, &v))
		capabilities = append(capabilities, v.Observation.Capability)
	}
	assert.Equal(t, []string{"tickets", "runs"}, capabilities, "store reads are journaled")
}

// TestHookScanHandler: a when.scan trigger with a handler gets each
// finding as an event; its run carries the finding id, so the finding is held
// while the run is open and max_runs counts it.
func TestHookScanHandler(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, `pack: pack.yaml
projects: {js: {}}
agents: {fake: {command: "false"}}
`)
	pack := `pools:
  plan: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}
workflows:
  triage:
    project: js
    vars: [signal, subject]
    steps:
      - {id: diagnose, pool: plan, title: "Finding {{.signal}} {{.subject}}", prompt: x}
triggers:
  scan:
    when:
      scan: [{signal: blocked, after: 1h}]
    handler: hooks.star:finding
    args: {workflow: triage}
    max_runs: 1
`
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack.yaml"), []byte(strings.ReplaceAll(pack, "$ROOT", root.Dir)), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), []byte(`def finding(event, args):
    if event.source != "scan" or event.trigger != "scan":
        fail("not a finding event")
    p = event.payload
    return [start_run(workflow=args["workflow"], key="sc-" + p["subject"], inputs={"signal": p["signal"], "subject": p["subject"]})]
`), 0o600))
	_, err := config.Load(root)
	require.NoError(t, err)
	c := controltest.Start(t, root, control.Options{})
	var blocked domain.Ticket
	c.Call(rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "stuck", Labels: []string{"blocked"}}, &blocked)
	c.Clock.Advance(2 * time.Hour)
	var runs []domain.Ticket
	c.PassUntil(func() bool {
		c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
		return len(runs) == 1
	}, "scan handler did not start a run")
	rf, err := domain.DecodeFields[domain.RunFields](runs[0])
	require.NoError(t, err)
	assert.Equal(t, "sc-"+blocked.ID, rf.Key)
	assert.Equal(t, "blocked", rf.Vars["signal"])
	assert.Contains(t, string(runs[0].Meta), `"finding":"blocked:`+blocked.ID+`"`)
	// Held while the run is open, then max_runs 1 holds it for good.
	c.Clock.Advance(time.Hour)
	c.Pass(3)
	c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
	require.Len(t, runs, 1)
	c.Call(rpc.MethodRunCancel, rpc.IDParams{ID: runs[0].ID}, nil)
	c.Clock.Advance(time.Hour)
	c.Pass(3)
	c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
	require.Len(t, runs, 1, "max_runs must count handler runs")
	assert.Equal(t, domain.StatusClosed, runs[0].Status)
}

func TestSyncHandlers(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return []
def completed(run, step):
    return []
def handle(event, args):
    return []
def push(event, args):
    if not event.get("target"):
        return {"decision": "deny", "reason": "push needs a target"}
    if event["target"] in args["protected"]:
        return {"decision": "deny", "reason": "no push to " + event["target"]}
    return {}
def state(input, args):
    return {"echo": input, "open": len(tickets(kind="task", status="open")), "dir": args["dir"]}
`)
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	extra := `hooks:
  - {on: action, actions: [push], handler: hooks.star:push, args: {protected: [master, main]}, timeout: 5s}
desk_commands:
  state: {handler: hooks.star:state, args: {dir: /srv}}
`
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte(extra)...), 0o600))
	_, err = config.Load(root)
	require.NoError(t, err)
	c := controltest.Start(t, root, control.Options{})

	guard := func(target string) rpc.GuardResult {
		var r rpc.GuardResult
		c.Call(rpc.MethodGuard, rpc.GuardParams{Action: "push", Target: target}, &r)
		return r
	}
	assert.Equal(t, rpc.GuardResult{Decision: domain.GuardDeny, Reason: "push needs a target"}, guard(""))
	assert.Equal(t, rpc.GuardResult{Decision: domain.GuardDeny, Reason: "no push to master"}, guard("master"))
	assert.Equal(t, domain.GuardAllow, guard("feature").Decision)
	var other rpc.GuardResult
	c.Call(rpc.MethodGuard, rpc.GuardParams{Action: "comment"}, &other)
	assert.Equal(t, domain.GuardAllow, other.Decision, "actions limits the hook")

	var res rpc.DeskExecResult
	c.Call(rpc.MethodDeskExec, rpc.DeskExecParams{Name: "state", Input: []byte(`{"a":1}`)}, &res)
	assert.JSONEq(t, `{"echo":{"a":1},"open":0,"dir":"/srv"}`, string(res.Output))

	var records []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{State: hooks.SyncDone}, &records)
	require.Len(t, records, 4)
	desk := slices.IndexFunc(records, func(r store.Record) bool {
		return strings.Contains(string(r.Value), `"sync":"desk_commands.state"`)
	})
	require.NotEqual(t, -1, desk)
	var trace rpc.AutomationResult
	c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: records[desk].Key}, &trace)
	require.Len(t, trace.Calls, 1, "the store read is journaled")

	// Sync journals are pruned after three days; a pass runs the prune.
	c.Clock.Advance(73 * time.Hour)
	c.Pass(1)
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{State: hooks.SyncDone}, &records)
	assert.Empty(t, records)
	err = c.Admin.Call(t.Context(), rpc.MethodAutomationShow, rpc.IDParams{ID: trace.Invocation.Key}, nil)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

// A script scan signal returns subject -> detail; an issue with an open
// run in its repo is skipped, and the answer stands until the next poll.
func TestScanScriptSignal(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, `pack: pack.yaml
projects: {js: {}}
agents: {fake: {command: "false"}}
`)
	// The scan trigger is loaded by a reload after issue 8's run exists: the
	// controller's first pass runs at start, and a scan that saw 8 without
	// its run would start a run for it.
	packBase := `pools:
  plan: {project: js, agent: fake, max_sessions: 0, dir: "$ROOT"}
workflows:
  triage:
    project: js
    vars: [signal, subject, detail]
    steps:
      - {id: diagnose, pool: plan, title: "Finding {{.signal}} {{.subject}}", prompt: x}
  issue:
    project: js
    repo: acme/widget
    issue: issue
    vars: [issue]
    steps:
      - {id: work, pool: plan, title: "Work {{.issue}}", prompt: x}
`
	scan := `triggers:
  scan:
    when:
      scan: [{signal: issue_no_run, repo: acme/widget, script: signals.star:issues, args: {label: watch}, poll: 10m}]
    run: {workflow: triage, vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}}
    max_runs: 1
`
	writePack := func(pack string) {
		require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack.yaml"), []byte(strings.ReplaceAll(pack, "$ROOT", root.Dir)), 0o600))
	}
	writePack(packBase)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "signals.star"), []byte(`def issues(args):
    found = {"#7": "open, no run", "8": "open, no run"}
    for t in tickets(label=args["label"], status="open"):
        found[t["title"]] = "from the store"
    return found
`), 0o600))
	_, err := config.Load(root)
	require.NoError(t, err)
	c := controltest.Start(t, root, control.Options{})
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "issue", Vars: map[string]string{"issue": "8"}}, nil)
	writePack(packBase + scan)
	c.Call(rpc.MethodReload, struct{}{}, nil)
	var runs []domain.Ticket
	scanRuns := func() []domain.Ticket {
		var all []domain.Ticket
		c.Call(rpc.MethodRunList, rpc.RunListParams{}, &all)
		runs = slices.DeleteFunc(all, func(r domain.Ticket) bool {
			rf, err := domain.DecodeFields[domain.RunFields](r)
			return err != nil || rf.Workflow != "triage"
		})
		return runs
	}
	c.PassUntil(func() bool { return len(scanRuns()) == 1 }, "script signal did not start a run")
	rf, err := domain.DecodeFields[domain.RunFields](runs[0])
	require.NoError(t, err)
	assert.Equal(t, "7", rf.Vars["subject"], "issue 8 has an open run")
	assert.Equal(t, "open, no run", rf.Vars["detail"])

	// A new subject waits for the next poll.
	c.Call(rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "9", Labels: []string{"watch"}}, nil)
	c.Pass(3)
	require.Len(t, scanRuns(), 1)
	c.Clock.Advance(11 * time.Minute)
	c.PassUntil(func() bool { return len(scanRuns()) == 2 }, "next poll did not see the new subject")
}

// A deploy that rewrites the callbacks strands runs started before it:
// they fail with "hook manifest changed" and plain resume refuses. An admin
// resume with Repin moves the run to the live callbacks and it finishes.
func TestHookRunResumeRepin(t *testing.T) {
	t.Parallel()
	root := hookTestRoot(t, `def begin(run):
    return [start_step(id="a",pool="dev",prompt="first")]
def completed(run, step):
    if step.id == "a":
        return [start_step(id="b",pool="dev",prompt="second",needs=["a"])]
    return [finish_run(outcome="done")]
def handle(event, args):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic"}, &run)
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 1 }, "first step not created")
	first := hookRun(t, c, run.Run.ID).Steps[0]
	// The deploy: callback source changes and the controller reloads.
	path := filepath.Join(root.Dir, "hooks.star")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(data, []byte("# deployed\n")...), 0o600))
	c.Call(rpc.MethodReload, struct{}{}, nil)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: first.ID}, nil)
	runState := func() string {
		f, err := domain.DecodeFields[domain.RunFields](hookRun(t, c, run.Run.ID).Run)
		require.NoError(t, err)
		return f.State
	}
	c.PassUntil(func() bool { return runState() == domain.RunNeedsHuman }, "changed callbacks did not stop the run")

	err = c.Admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: run.Run.ID}, nil)
	require.Error(t, err, "plain resume must not adopt the changed manifest")
	assert.Equal(t, domain.RunNeedsHuman, runState())

	require.NoError(t, c.Admin.Call(t.Context(), rpc.MethodRunResume, rpc.RunResumeParams{ID: run.Run.ID, Repin: true}, nil))
	c.PassUntil(func() bool { return len(hookRun(t, c, run.Run.ID).Steps) == 2 }, "repinned run did not create its second step")
	assert.Equal(t, domain.RunRunning, runState())
	var shown rpc.ShowResult
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: run.Run.ID}, &shown)
	var repinned bool
	for _, cm := range shown.Comments {
		repinned = repinned || strings.Contains(cm.Body, "Repinned to the live callback manifest")
	}
	assert.True(t, repinned, "the repin is logged as a run comment")
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: hookRun(t, c, run.Run.ID).Steps[1].ID}, nil)
	c.PassUntil(func() bool { return hookRun(t, c, run.Run.ID).Run.Status == domain.StatusClosed }, "repinned run did not finish")
}
