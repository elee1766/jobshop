package control

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// workerController is an admissionController in ingest.workers external
// mode, started now on the fake clock.
func workerController(t *testing.T, hooks []config.AdmissionHook, edit func(*config.Ingest)) (*Controller, *scanClock) {
	t.Helper()
	c, clk := admissionController(t, hooks)
	cfg := *c.config()
	cfg.Ingest.Workers = config.IngestWorkersExternal
	if edit != nil {
		edit(&cfg.Ingest)
	}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.started = c.now()
	return c, clk
}

// inProcCall routes a worker's RPCs to the controller's admin handlers,
// through the same JSON the socket carries.
func inProcCall(c *Controller) func(ctx context.Context, method string, params, out any) error {
	return func(ctx context.Context, method string, params, out any) error {
		raw, err := json.Marshal(params)
		if err != nil {
			return err
		}
		res, err := methods[method].call(c, ctx, scope{admin: true, actor: "admin"}, raw)
		if err != nil {
			return err
		}
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, out)
	}
}

func worker(c *Controller, id string) IngestWorker {
	return IngestWorker{Call: inProcCall(c), Dir: c.root.Dir, Worker: id, Host: "h"}
}

func TestIngestWorkerAdmitsControllerDoesNot(t *testing.T) {
	t.Parallel()
	c, _ := workerController(t, []config.AdmissionHook{
		mut("upper", `echo '{"patch":[{"op":"replace","path":"/payload/body","value":"HELLO"}]}'`),
		val("deny-bad", `grep -q '"actor":"bad"' && echo '{"decision":"deny","reason":"bad actor"}' || true`),
	}, nil)
	ok1 := putInbox(t, c, store.InboxItem{DeliveryID: "a"})
	bad := putInbox(t, c, store.InboxItem{DeliveryID: "b", Actor: "bad"})

	c.ingestPass(t.Context())
	got, err := ingest.InboxGet(t.Context(), c.store, ok1.ID)
	require.NoError(t, err)
	assert.Equal(t, store.InboxPending, got.State, "the controller does not admit in external mode")

	n, err := worker(c, "w1").Once(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	got, err = ingest.InboxGet(t.Context(), c.store, ok1.ID)
	require.NoError(t, err)
	assert.Equal(t, store.InboxAdmitted, got.State)
	assert.JSONEq(t, `{"body":"HELLO","token":"s3cret"}`, string(got.Admitted))
	got, err = ingest.InboxGet(t.Context(), c.store, bad.ID)
	require.NoError(t, err)
	assert.Equal(t, store.InboxDenied, got.State)

	ws := c.workerList()
	require.Len(t, ws, 1)
	assert.Equal(t, "w1", ws[0].ID)
	assert.EqualValues(t, 2, ws[0].Admitted)
	assert.False(t, ws[0].LastClaim.IsZero())
}

func TestIngestWorkerClaimRefusedInternal(t *testing.T) {
	t.Parallel()
	c, _ := admissionController(t, nil)
	putInbox(t, c, store.InboxItem{})
	_, err := worker(c, "w1").Once(t.Context())
	require.ErrorIs(t, err, store.ErrConflict)
}

// TestIngestWorkerDeadLeaseExpires: a worker that claims and dies leaves
// the item leased; after the lease another worker admits it, and the
// dead worker's late report is refused.
func TestIngestWorkerDeadLeaseExpires(t *testing.T) {
	t.Parallel()
	c, clk := workerController(t, nil, nil)
	it := putInbox(t, c, store.InboxItem{})
	var cl rpc.IngestClaimResult
	require.NoError(t, inProcCall(c)(t.Context(), rpc.MethodIngestClaim, rpc.IngestClaimParams{Worker: "dead"}, &cl))
	require.NotNil(t, cl.Item)

	n, err := worker(c, "w2").Once(t.Context())
	require.NoError(t, err)
	assert.Zero(t, n, "leased by the dead worker")

	clk.advance(admissionLease(nil) + time.Second)
	n, err = worker(c, "w2").Once(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	got, err := ingest.InboxGet(t.Context(), c.store, it.ID)
	require.NoError(t, err)
	assert.Equal(t, store.InboxAdmitted, got.State)
	assert.EqualValues(t, 2, got.Attempts)

	run, _ := json.Marshal(admissionRun{})
	err = inProcCall(c)(t.Context(), rpc.MethodIngestAdmit, rpc.IngestAdmitParams{Worker: "dead", ID: it.ID, Run: run}, &json.RawMessage{})
	require.ErrorIs(t, err, store.ErrConflict)
}

func attentionIDs(t *testing.T, c *Controller) []string {
	t.Helper()
	items, err := c.attentionItems(t.Context(), c.config())
	require.NoError(t, err)
	var ids []string
	for _, it := range items {
		if it.Kind == rpc.AttentionIngest {
			ids = append(ids, it.ID)
		}
	}
	return ids
}

func TestIngestMissingWorkerAttention(t *testing.T) {
	t.Parallel()
	c, clk := workerController(t, nil, func(i *config.Ingest) { i.WorkerMissing = "5m" })
	assert.Empty(t, attentionIDs(t, c))
	clk.advance(6 * time.Minute)
	assert.Equal(t, []string{"ingest-worker"}, attentionIDs(t, c), "no worker since start")
	s, err := c.signalIngest(t.Context(), c.config(), config.SignalIngestWorker)
	require.NoError(t, err)
	require.Len(t, s, 1)

	_, err = worker(c, "w1").Once(t.Context())
	require.NoError(t, err)
	assert.Empty(t, attentionIDs(t, c), "a heartbeat clears it")
	clk.advance(6 * time.Minute)
	assert.Equal(t, []string{"ingest-worker"}, attentionIDs(t, c), "the worker went silent")

	// Internal mode never lists a missing worker.
	c2, clk2 := admissionController(t, nil)
	clk2.advance(time.Hour)
	assert.Empty(t, attentionIDs(t, c2))
}

func TestIngestBacklogAttention(t *testing.T) {
	t.Parallel()
	c, clk := workerController(t, nil, func(i *config.Ingest) {
		i.BacklogAfter, i.BacklogMax, i.WorkerMissing = "10m", 1, "1h"
	})
	putInbox(t, c, store.InboxItem{DeliveryID: "a"})
	putInbox(t, c, store.InboxItem{DeliveryID: "b"})
	clk.advance(5 * time.Minute)
	assert.Empty(t, attentionIDs(t, c), "not old enough")
	clk.advance(6 * time.Minute)
	assert.Equal(t, []string{"ingest-backlog"}, attentionIDs(t, c))
	s, err := c.signalIngest(t.Context(), c.config(), config.SignalIngestBacklog)
	require.NoError(t, err)
	require.Len(t, s, 1)
	assert.Contains(t, s[0].detail, "2 pending")

	_, err = worker(c, "w1").Once(t.Context())
	require.NoError(t, err)
	assert.Empty(t, attentionIDs(t, c), "admitted")
	counts, err := ingest.InboxCounts(t.Context(), c.store)
	require.NoError(t, err)
	assert.EqualValues(t, 2, counts[store.InboxAdmitted])
	assert.Zero(t, counts[store.InboxPending])
}
