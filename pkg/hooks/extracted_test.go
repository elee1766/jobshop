package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
)

func TestCheckStepOutput(t *testing.T) {
	t.Parallel()
	schema := map[string]any{"type": "object", "properties": map[string]any{"n": map[string]any{"type": "integer"}}, "required": []any{"n"}}
	require.NoError(t, CheckStepOutput(0, schema, map[string]any{"n": 1}))
	err := CheckStepOutput(0, schema, map[string]any{"n": "x"})
	require.ErrorContains(t, err, "step output")
	require.Error(t, CheckStepOutput(0, schema, map[string]any{}), "required key missing")
	require.NoError(t, CheckStepOutput(0, nil, map[string]any{"anything": true}), "no schema accepts any output")

	// The size limit applies with or without a schema; 0 is the default.
	big := map[string]any{"s": "0123456789"}
	require.ErrorContains(t, CheckStepOutput(8, nil, big), "exceeds 8 bytes")
	require.NoError(t, CheckStepOutput(0, nil, big))
}

func TestCheckSchemaCompileOnly(t *testing.T) {
	t.Parallel()
	require.NoError(t, CheckSchema(map[string]any{"type": "object"}, nil, false))
	err := CheckSchema(map[string]any{"type": "nonsense"}, nil, false)
	require.ErrorContains(t, err, "output_schema")
	// Compile-only ignores the output.
	require.NoError(t, CheckSchema(map[string]any{"type": "object"}, map[string]any{"x": 1}, false))
}

func step(id, status, resolution string, rev int64, fields domain.StepFields) domain.Ticket {
	f, _ := json.Marshal(fields)
	return domain.Ticket{ID: id, Kind: domain.KindStep, Pool: "dev", Status: domain.Status(status), Resolution: resolution, Revision: rev, Fields: f}
}

func runTicket(t *testing.T, rf domain.RunFields) (domain.Ticket, domain.RunFields) {
	t.Helper()
	f, err := json.Marshal(rf)
	require.NoError(t, err)
	return domain.Ticket{ID: "run1", Kind: domain.KindRun, Project: "js", Revision: 3, Fields: f}, rf
}

func TestStepInput(t *testing.T) {
	t.Parallel()
	tk := step("s1", "closed", "done", 2, domain.StepFields{Step: "a", Outcome: "ok", Failures: 1, Retries: 2, StructuredOutput: map[string]any{"k": "v"}})
	got, err := StepInput(tk)
	require.NoError(t, err)
	assert.Equal(t, Step{ID: "a", TicketID: "s1", Pool: "dev", Status: "closed", Resolution: "done", Outcome: "ok", Output: map[string]any{"k": "v"}, Failures: 1, Retries: 2}, got)
}

// TestRunDue: on_start while no batch is applied, then the first finished
// step whose revision was not completed, with a key that changes with the
// step revision and the automation version.
func TestRunDue(t *testing.T) {
	t.Parallel()
	auto := &domain.RunAutomation{OnStart: "start", OnStepCompleted: "done", Digest: "d"}
	run, rf := runTicket(t, domain.RunFields{Workflow: "w", Key: "k", State: domain.RunRunning, Automation: auto})
	a := step("sa", "closed", "done", 5, domain.StepFields{Step: "a"})
	b := step("sb", "open", "", 1, domain.StepFields{Step: "b"})
	children := []domain.Ticket{a, b, {ID: "note", Kind: domain.KindTask, Revision: 9}}

	key, in, ok, err := RunDue(run, rf, children, "parent")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "run1/start", key)
	assert.Equal(t, "start", in.Entry)
	assert.Equal(t, "parent", in.ParentInvocation)
	assert.Equal(t, "d", in.Digest)
	assert.Equal(t, map[string]int64{"run1": 3, "sa": 5, "sb": 1, "note": 9}, in.ReadSet, "every child is in the read set, steps or not")
	require.Len(t, in.RunInput.Steps, 2, "only steps are in the snapshot")
	assert.Empty(t, in.Step)

	// After the first batch: the finished, uncompleted step is due.
	auto.Version = 1
	key, in, ok, err = RunDue(run, rf, children, "")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "run1/step/sa/5/1", key)
	assert.Equal(t, "done", in.Entry)
	assert.Equal(t, "sa", in.Step)
	assert.Equal(t, int64(5), in.StepRevision)
	require.NotNil(t, in.StepInput)
	assert.Equal(t, "a", in.StepInput.ID)

	// Completed at this revision: nothing is due. A new revision is.
	auto.Completed = map[string]int64{"sa": 5}
	_, _, ok, err = RunDue(run, rf, children, "")
	require.NoError(t, err)
	assert.False(t, ok)
	children[0].Revision = 6
	key, _, ok, err = RunDue(run, rf, children, "")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "run1/step/sa/6/1", key)

	// A skipped step counts as finished; an open one never does.
	children[0] = step("sa", "closed", domain.StepSkipped, 7, domain.StepFields{Step: "a"})
	_, _, ok, _ = RunDue(run, rf, children, "")
	assert.True(t, ok)
	children[0] = step("sa", "open", "", 7, domain.StepFields{Step: "a"})
	_, _, ok, _ = RunDue(run, rf, children, "")
	assert.False(t, ok)

	// A workflow with no on_start has nothing due at version 0.
	auto.Version, auto.OnStart = 0, ""
	_, _, ok, _ = RunDue(run, rf, children, "")
	assert.False(t, ok)
}

func TestRetryAfter(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	assert.Zero(t, RetryAfter("", now))
	assert.Equal(t, 7*time.Second, RetryAfter(" 7 ", now))
	assert.Zero(t, RetryAfter("-3", now))
	assert.Equal(t, 90*time.Second, RetryAfter(now.Add(90*time.Second).Format(http.TimeFormat), now))
	assert.Zero(t, RetryAfter(now.Add(-time.Minute).Format(http.TimeFormat), now), "a past date")
	assert.Zero(t, RetryAfter("soon", now))
}

// newWorker runs the host's goroutines on wg and counts enqueue and
// process calls. process answers from results in order, then false.
func newWorker(ctx context.Context, wg *sync.WaitGroup, enqueue func(context.Context) error, process func(context.Context) (bool, error)) *Worker {
	return &Worker{
		Go:      func(f func()) { wg.Go(f) },
		Context: func() context.Context { return ctx },
		Enqueue: enqueue, Process: process, Log: slog.New(slog.DiscardHandler),
	}
}

// TestWorkerDrain: one Kick enqueues, processes until nothing is left,
// and enqueues again after each processed invocation.
func TestWorkerDrain(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	var enqueued, processed int
	left := 3
	w := newWorker(t.Context(), &wg,
		func(context.Context) error { enqueued++; return nil },
		func(context.Context) (bool, error) {
			if left == 0 {
				return false, nil
			}
			left--
			processed++
			return true, nil
		})
	w.Kick()
	wg.Wait()
	assert.Equal(t, 3, processed)
	assert.Equal(t, 4, enqueued, "once at the start, then after each processed invocation")
}

// TestWorkerBudget: a queue longer than the budget keeps draining in a
// follow-up run instead of stopping.
func TestWorkerBudget(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	left := 2*drainBudget + 5
	var processed int
	w := newWorker(t.Context(), &wg,
		func(context.Context) error { return nil },
		func(context.Context) (bool, error) {
			if left == 0 {
				return false, nil
			}
			left--
			processed++
			return true, nil
		})
	w.Kick()
	wg.Wait()
	assert.Equal(t, 2*drainBudget+5, processed)
}

// TestWorkerKickWhileRunning: a Kick during a drain makes it go round
// again, and never runs two drains at once.
func TestWorkerKickWhileRunning(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	var mu sync.Mutex
	var active, maxActive, passes int
	inside := make(chan struct{})
	release := make(chan struct{})
	w := newWorker(t.Context(), &wg,
		func(context.Context) error {
			mu.Lock()
			active++
			maxActive = max(maxActive, active)
			passes++
			first := passes == 1
			mu.Unlock()
			if first {
				close(inside)
				<-release
			}
			return nil
		},
		func(context.Context) (bool, error) {
			mu.Lock()
			active--
			mu.Unlock()
			return false, nil
		})
	w.Kick()
	<-inside
	w.Kick() // while the first drain is inside Enqueue
	w.Kick()
	close(release)
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, maxActive)
	assert.Equal(t, 2, passes, "the kicks during the drain collapse into one more pass")
}

func TestWorkerErrorsAndStop(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	calls := 0
	w := newWorker(t.Context(), &wg,
		func(context.Context) error { return nil },
		func(context.Context) (bool, error) { calls++; return false, errors.New("boom") })
	w.Kick()
	wg.Wait()
	assert.Equal(t, 1, calls, "an error ends the drain")

	// An ended context makes Kick a no-op.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	w = newWorker(ctx, &wg, func(context.Context) error { called = true; return nil }, nil)
	w.Kick()
	wg.Wait()
	assert.False(t, called)

	// So does a host with no context yet.
	w = newWorker(t.Context(), &wg, func(context.Context) error { called = true; return nil }, nil)
	w.Context = func() context.Context { return nil }
	w.Kick()
	wg.Wait()
	assert.False(t, called)
}
