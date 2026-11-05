package storetest

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// slotSteps creates a run with one dev step per id, each needing res.
func slotSteps(t *testing.T, s store.Store, res string, ids ...string) {
	t.Helper()
	runID, err := s.NewID(proj)
	require.NoError(t, err)
	nts := []store.NewTicket{{
		ID: runID, Project: proj, Kind: store.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"w","state":"running"}`),
	}}
	for _, id := range ids {
		nts = append(nts, store.NewTicket{
			ID: id, Project: proj, Kind: store.KindStep, Pool: "dev", ParentID: runID,
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":%q,"retries":5,"resources":[%q]}`, runID, id, res)),
		})
	}
	_, err = s.CreateGraph(t.Context(), "test", nts, nil)
	require.NoError(t, err)
}

func slotState(t *testing.T, s store.Store, pool, id string) store.Slot {
	t.Helper()
	all, err := s.Slots(t.Context())
	require.NoError(t, err)
	for _, sl := range all {
		if sl.Pool == pool && sl.ID == id {
			return sl
		}
	}
	return store.Slot{Pool: pool, ID: id, State: "absent"}
}

// testProviderSlots walks the slot states (#407): no slot, not ready; a
// claim takes a slot; the claim end moves it to releasing; reset ok is
// ready, failed is gone; unhealthy while held requeues without counting
// a failure; remove while held deletes it after the claim.
func testProviderSlots(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	require.NoError(t, s.SetResourceProviders(ctx, []string{"emu"}))
	slotSteps(t, s, "emu", "js-a", "js-b")
	ready := func() []string {
		t.Helper()
		r, err := s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
		require.NoError(t, err)
		return ids(r)
	}
	assert.Empty(t, ready(), "no slot, nothing ready")
	_, _, err := s.Claim(ctx, "js-a", "s1", lease)
	require.ErrorIs(t, err, store.ErrNoWork)

	_, err = s.AddSlot(ctx, "emu", "e1", map[string]string{"ANDROID_SERIAL": "emulator-5554"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"js-a", "js-b"}, ready())

	ta, ca, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s1", Lease: lease})
	require.NoError(t, err)
	assert.Equal(t, "js-a", ta.ID)
	sl := slotState(t, s, "emu", "e1")
	assert.Equal(t, store.SlotHeld, sl.State)
	assert.Equal(t, "js-a", sl.Ticket)
	held, err := s.TicketSlots(ctx, "js-a")
	require.NoError(t, err)
	require.Len(t, held, 1)
	assert.Equal(t, "emulator-5554", held[0].Env["ANDROID_SERIAL"])
	assert.Empty(t, ready(), "the only slot is held")

	// Re-claim of the holder's own ticket keeps the slot, takes no other.
	_, _, err = s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err)

	// The claim ends: releasing, then reset ok -> ready.
	rel, err := s.ReleaseSlots(ctx)
	require.NoError(t, err)
	assert.Empty(t, rel, "a live claim keeps its slot")
	_, err = s.CloseTicket(ctx, "s1", "js-a", &ca, "done")
	require.NoError(t, err)
	assert.Empty(t, ready(), "releasing is not ready")
	rel, err = s.ReleaseSlots(ctx)
	require.NoError(t, err)
	require.Len(t, rel, 1)
	assert.Equal(t, store.SlotReleasing, rel[0].State)
	require.NoError(t, s.SlotReset(ctx, "emu", "e1", true))
	require.ErrorIs(t, s.SlotReset(ctx, "emu", "e1", true), store.ErrConflict, "reset once")
	assert.Equal(t, []string{"js-b"}, ready())

	// Unhealthy while held: requeued, failures not counted, slot gone.
	_, _, err = s.Claim(ctx, "js-b", "s2", lease)
	require.NoError(t, err)
	_, reopened, err := s.SlotUnhealthy(ctx, "provider", "emu", "e1", "adb offline")
	require.NoError(t, err)
	assert.Equal(t, "js-b", reopened)
	tb, err := s.Get(ctx, "js-b")
	require.NoError(t, err)
	assert.Equal(t, store.StatusOpen, tb.Status)
	sf, err := store.DecodeFields[store.StepFields](tb)
	require.NoError(t, err)
	assert.Zero(t, sf.Failures, "infrastructure failure is not a step failure")
	assert.Equal(t, store.SlotGone, slotState(t, s, "emu", "e1").State)
	assert.Empty(t, ready(), "gone is not ready")

	// Add brings a gone slot back; a failed reset makes it gone.
	_, err = s.AddSlot(ctx, "emu", "e1", nil)
	require.NoError(t, err)
	_, cb, err := s.Claim(ctx, "js-b", "s2", lease)
	require.NoError(t, err)
	_, err = s.Release(ctx, cb, "later", c.T.Add(time.Hour))
	require.NoError(t, err)
	_, err = s.ReleaseSlots(ctx)
	require.NoError(t, err)
	require.NoError(t, s.SlotReset(ctx, "emu", "e1", false))
	assert.Equal(t, store.SlotGone, slotState(t, s, "emu", "e1").State)

	// Lease expiry ends a claim too; remove while held deletes after it.
	_, err = s.AddSlot(ctx, "emu", "e1", nil)
	require.NoError(t, err)
	c.Advance(2 * time.Hour)
	_, _, err = s.Claim(ctx, "js-b", "s3", lease)
	require.NoError(t, err)
	_, err = s.RemoveSlot(ctx, "emu", "e1")
	require.NoError(t, err)
	assert.True(t, slotState(t, s, "emu", "e1").Removing)
	c.Advance(lease + time.Minute)
	_, err = s.SweepLeases(ctx, "controller")
	require.NoError(t, err)
	rel, err = s.ReleaseSlots(ctx)
	require.NoError(t, err)
	require.Len(t, rel, 1)
	require.NoError(t, s.SlotReset(ctx, "emu", "e1", true))
	assert.Equal(t, "absent", slotState(t, s, "emu", "e1").State, "removed slot deleted after its claim")

	// Idle remove deletes at once; demand counts.
	_, err = s.AddSlot(ctx, "emu", "e2", nil)
	require.NoError(t, err)
	d, err := s.SlotDemand(ctx, "emu")
	require.NoError(t, err)
	assert.Equal(t, 1, d.Waiting)
	assert.Equal(t, 1, d.Ready)
	_, err = s.RemoveSlot(ctx, "emu", "e2")
	require.NoError(t, err)
	assert.Equal(t, "absent", slotState(t, s, "emu", "e2").State)
	_, err = s.RemoveSlot(ctx, "emu", "e2")
	require.ErrorIs(t, err, store.ErrNotFound)

	// Not a provider resource any more: no gate.
	require.NoError(t, s.SetResourceProviders(ctx, nil))
	assert.Equal(t, []string{"js-b"}, ready())
}

// testConcurrentSlotClaims: holders race for steps sharing two slots;
// each claim takes a distinct slot and no more than two run at once.
func testConcurrentSlotClaims(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	require.NoError(t, s.SetResourceProviders(ctx, []string{"emu"}))
	var steps []string
	for i := range 6 {
		steps = append(steps, fmt.Sprintf("js-q%d", i))
	}
	slotSteps(t, s, "emu", steps...)
	for _, id := range []string{"e1", "e2"} {
		_, err := s.AddSlot(ctx, "emu", id, nil)
		require.NoError(t, err)
	}
	var mu sync.Mutex
	var got []string
	var wg sync.WaitGroup
	for i := range 8 {
		holder := fmt.Sprintf("h%d", i)
		wg.Go(func() {
			tk, _, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: holder, Lease: lease})
			if err != nil {
				return
			}
			mu.Lock()
			got = append(got, tk.ID)
			mu.Unlock()
		})
	}
	wg.Wait()
	assert.Len(t, got, 2)
	all, err := s.Slots(ctx)
	require.NoError(t, err)
	var tickets []string
	for _, sl := range all {
		assert.Equal(t, store.SlotHeld, sl.State)
		tickets = append(tickets, sl.Ticket)
	}
	assert.ElementsMatch(t, got, tickets)
}
