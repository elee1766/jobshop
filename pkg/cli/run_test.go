package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func parse(t *testing.T, args ...string) *root {
	t.Helper()
	var r root
	p, err := kong.New(&r, kong.Exit(func(int) { t.Fatal("exit") }))
	require.NoError(t, err)
	_, err = p.Parse(args)
	require.NoError(t, err)
	return &r
}

func TestRunStartParse(t *testing.T) {
	r := parse(t, "run", "start", "pr-work", "--var", "issue=42", "-v", "x=y", "--key", "sha1")
	assert.Equal(t, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "42", "x": "y"}, Key: "sha1"},
		r.Run.Start.params())
	r = parse(t, "run", "resume", "js-1", "--step", "implement")
	assert.Equal(t, rpc.RunResumeParams{ID: "js-1", Step: "implement"}, r.Run.Resume.params())
	r = parse(t, "run", "list", "--state", "needs_human")
	assert.Equal(t, "needs_human", r.Run.List.State)
}

func TestPrintRun(t *testing.T) {
	rf, _ := json.Marshal(store.RunFields{Workflow: "pr", Vars: map[string]string{"issue": "42", "a": "b"}, Key: "k", State: store.RunRunning})
	s0, _ := json.Marshal(store.StepFields{Run: "js-r", Step: "implement", Index: 0, Retries: 3})
	s1, _ := json.Marshal(store.StepFields{Run: "js-r", Step: "review", Index: 1, Retries: 2, Failures: 1})
	res := rpc.RunResult{
		Run: store.Ticket{ID: "js-r", Kind: store.KindRun, Status: store.StatusOpen, Fields: rf},
		Steps: []store.Ticket{
			{ID: "js-s0", Kind: store.KindStep, Pool: "dev", Status: store.StatusClosed, Resolution: "done", Fields: s0},
			{ID: "js-s1", Kind: store.KindStep, Pool: "refinery", Status: store.StatusInProgress, Holder: "js-sess", Fields: s1},
		},
		Existing: true,
	}
	var b bytes.Buffer
	require.NoError(t, printRun(&b, res))
	assert.Equal(t, `run js-r  workflow pr  state running  (open)  key k  [existing run for key]
vars a=b issue=42
0  implement    js-s0  pool dev  closed  done  holder -  failures 0/3
1  review       js-s1  pool refinery  in_progress  -  holder js-sess  failures 1/2
`, b.String())
}

func TestBatchStartParse(t *testing.T) {
	r := parse(t, "batch", "start", "--key", "chat", "--issues", "122,123,125", "--order", "122>123>125", "--order", "122>125")
	assert.Equal(t, rpc.BatchStartParams{Key: "chat", Issues: []string{"122", "123", "125"}, Order: []string{"122>123>125", "122>125"}},
		r.Batch.Start.params())
}

func TestPrintBatch(t *testing.T) {
	bf, _ := json.Marshal(store.BatchFields{Key: "chat", Issues: []string{"1", "2"}, Order: [][2]string{{"1", "2"}}, State: store.BatchOpen})
	rf := func(k string) json.RawMessage {
		b, _ := json.Marshal(store.RunFields{Workflow: "pr-work", Key: k, State: store.RunRunning})
		return b
	}
	res := rpc.BatchResult{
		Batch: store.Ticket{ID: "js-b", Kind: store.KindBatch, Status: store.StatusOpen, Fields: bf},
		Runs: []rpc.RunResult{
			{Run: store.Ticket{ID: "js-r1", Kind: store.KindRun, Fields: rf("issue-1")}},
			{Run: store.Ticket{ID: "js-r2", Kind: store.KindRun, Fields: rf("issue-2")}},
		},
		Adopted: []string{"js-r2"},
	}
	var b bytes.Buffer
	require.NoError(t, printBatch(&b, res))
	assert.Equal(t, `batch js-b  key chat  state open  (open)
order 1>2
issue-1  run js-r1  state running
issue-2  run js-r2  state running  [adopted]
`, b.String())
}

// TestPrintBatchShow: batch show prints the current step and which issue
// blocks each member.
func TestPrintBatchShow(t *testing.T) {
	bf, _ := json.Marshal(store.BatchFields{Key: "chat", Issues: []string{"1", "2"}, State: store.BatchRunning})
	rf := func(k string) json.RawMessage {
		b, _ := json.Marshal(store.RunFields{Workflow: "pr-work", Key: k, State: store.RunRunning})
		return b
	}
	step := func(id, name string, status store.Status, res string) store.Ticket {
		b, _ := json.Marshal(store.StepFields{Step: name})
		return store.Ticket{ID: id, Kind: store.KindStep, Status: status, Resolution: res, Fields: b}
	}
	res := rpc.BatchResult{
		Batch: store.Ticket{ID: "js-b", Kind: store.KindBatch, Status: store.StatusOpen, Fields: bf},
		Runs: []rpc.RunResult{
			{Run: store.Ticket{ID: "js-r1", Fields: rf("issue-1")}, Steps: []store.Ticket{
				step("s1", "implement", store.StatusClosed, "done"), step("s2", "review-merge", store.StatusInProgress, ""),
			}},
			{Run: store.Ticket{ID: "js-r2", Fields: rf("issue-2")}, Steps: []store.Ticket{step("s3", "implement", store.StatusOpen, "")}},
		},
		BlockedBy: map[string][]string{"js-r2": {"1"}},
	}
	var b bytes.Buffer
	require.NoError(t, printBatch(&b, res))
	assert.Equal(t, `batch js-b  key chat  state running  (open)
issue-1  run js-r1  state running  step review-merge
issue-2  run js-r2  state running  step implement  blocked by 1
`, b.String())
}

func TestBatchShowListParse(t *testing.T) {
	r := parse(t, "batch", "show", "chat")
	assert.Equal(t, "chat", r.Batch.Show.ID)
	r = parse(t, "batch", "list", "--project", "js")
	assert.Equal(t, "js", r.Batch.List.Project)
}
