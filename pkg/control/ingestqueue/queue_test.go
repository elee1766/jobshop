package ingestqueue_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/ingestqueue"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type fakeHost struct {
	clk   *clock.Fake
	woken [][]store.InboxItem
}

func (h *fakeHost) Clock() clock.Clock              { return h.clk }
func (h *fakeHost) Store() ingest.InboxStore        { return nil }
func (h *fakeHost) SpoolPath() string               { return "" }
func (h *fakeHost) WakeWaits(its []store.InboxItem) { h.woken = append(h.woken, its) }
func (h *fakeHost) Drained(_ []store.InboxItem)     {}

func newService() (*ingestqueue.Service, *fakeHost) {
	h := &fakeHost{clk: clock.NewFake(time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC))}
	return ingestqueue.New(h, slog.New(slog.DiscardHandler)), h
}

func TestAppendedAdvancesAndWakes(t *testing.T) {
	t.Parallel()
	s, h := newService()
	ch := s.Changed()
	s.Appended(nil)
	select {
	case <-ch:
		t.Fatal("an empty append is not a change")
	default:
	}
	s.Appended([]store.InboxItem{{DeliveryID: "d"}})
	<-ch
	assert.Len(t, h.woken, 1)
	res := s.Next(t.Context(), 0, time.Second)
	assert.True(t, res.Changed)
	assert.EqualValues(t, 1, res.Seq)
}

func TestNextReturnsOnTimeout(t *testing.T) {
	t.Parallel()
	s, h := newService()
	done := make(chan rpc.IngestNextResult)
	go func() { done <- s.Next(t.Context(), 0, time.Second) }()
	require.True(t, h.clk.WaitFor(1, t.Context().Done()))
	h.clk.Advance(2 * time.Second)
	res := <-done
	assert.False(t, res.Changed)
}

func TestHeartbeats(t *testing.T) {
	t.Parallel()
	s, h := newService()
	start := h.clk.Now()
	assert.Equal(t, start, s.LastBeat(start), "no worker: the given start")
	s.Beat(rpc.IngestClaimParams{Worker: "b", Host: "hb"}, false, false)
	h.clk.Advance(time.Minute)
	s.Beat(rpc.IngestClaimParams{Worker: "a"}, true, false)
	s.Beat(rpc.IngestClaimParams{Worker: "a"}, false, true)
	ws := s.Workers()
	require.Len(t, ws, 2)
	assert.Equal(t, "a", ws[0].ID, "sorted by id")
	assert.EqualValues(t, 1, ws[0].Admitted)
	assert.False(t, ws[0].LastClaim.IsZero())
	assert.True(t, ws[1].LastClaim.IsZero())
	assert.Equal(t, h.clk.Now(), s.LastBeat(start))
	assert.EqualValues(t, 1, s.Live(start.Add(30*time.Second)), "b is older than the cutoff")
}

func TestSettle(t *testing.T) {
	t.Parallel()
	run := ingestqueue.Run{Event: ingestqueue.Event{Payload: []byte(`{"a":1}`)}}
	st, payload, _ := ingestqueue.Settle(run, rpc.GuardResult{}, false)
	assert.Equal(t, store.InboxAdmitted, st)
	assert.JSONEq(t, `{"a":1}`, string(payload))

	run.Asks = []string{"v: unsure"}
	st, payload, _ = ingestqueue.Settle(run, rpc.GuardResult{}, false)
	assert.Equal(t, store.InboxHeld, st)
	assert.Nil(t, payload)

	st, _, rs := ingestqueue.Settle(run, rpc.GuardResult{Decision: rpc.GuardAllow}, true)
	assert.Equal(t, store.InboxAdmitted, st, "a human's allow stands over the hold")
	assert.Equal(t, "human", rs[len(rs)-1].Name)

	run.Denied = true
	st, _, _ = ingestqueue.Settle(run, rpc.GuardResult{}, false)
	assert.Equal(t, store.InboxDenied, st)
	st, _, _ = ingestqueue.Settle(run, rpc.GuardResult{Decision: rpc.GuardDeny}, true)
	assert.Equal(t, store.InboxDenied, st)
}

func TestHooksAddsIntake(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.Ingest.Admission = []config.AdmissionHook{{Name: "m", Type: config.AdmissionMutating}}
	cfg.Hooks = []config.Hook{{On: config.HookIntake, Command: "true"}, {On: "guard", Command: "x"}}
	hs := ingestqueue.Hooks(cfg)
	require.Len(t, hs, 2)
	assert.Equal(t, "on-intake-1", hs[1].Name)
	assert.True(t, hs[1].Intake)
	assert.Greater(t, ingestqueue.Lease(hs), time.Minute)
}
