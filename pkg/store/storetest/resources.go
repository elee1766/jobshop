package storetest

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// testResources: a step whose resource is full is neither ready nor
// claimable, and its slot frees on close, release, lease expiry and a
// dead holder (#400). Keyed keys count separately; unknown keys are
// unlimited.
func testResources(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	require.NoError(t, s.SetResourceCaps(ctx, map[string]int{"emu": 1, "pr": 1}))
	runID, err := s.NewID(proj)
	require.NoError(t, err)
	nts := []store.NewTicket{{
		ID: runID, Project: proj, Kind: store.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"w","state":"running"}`),
	}}
	step := func(id string, res ...string) {
		b, err := json.Marshal(res)
		require.NoError(t, err)
		nts = append(nts, store.NewTicket{
			ID: id, Project: proj, Kind: store.KindStep, Pool: "dev", ParentID: runID,
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":%q,"retries":5,"resources":%s}`, runID, id, b)),
		})
	}
	step("js-a", "emu")
	step("js-b", "emu")
	step("js-c", "pr:x#1")
	step("js-d", "pr:x#1")
	step("js-e", "pr:x#2")
	step("js-f", "other")
	step("js-g", "other")
	_, err = s.CreateGraph(ctx, "test", nts, nil)
	require.NoError(t, err)

	ready := func() []string {
		t.Helper()
		r, err := s.Ready(ctx, proj, "dev", 0)
		require.NoError(t, err)
		return ids(r)
	}
	req := func(h string) store.ClaimNext {
		return store.ClaimNext{Project: proj, Pool: "dev", Holder: h, Lease: lease}
	}

	_, ca, err := s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err)
	_, cc, err := s.Claim(ctx, "js-c", "s2", lease)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"js-e", "js-f", "js-g"}, ready(), "emu and pr:x#1 full; pr:x#2 and uncapped free")

	_, _, err = s.Claim(ctx, "js-b", "s3", lease)
	require.ErrorIs(t, err, store.ErrNoWork, "claim by id respects the cap")
	_, _, err = s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err, "the holder re-claims its own ticket")

	use, err := s.Resources(ctx)
	require.NoError(t, err)
	assert.Contains(t, use, store.ResourceUse{Key: "emu", Capacity: 1, Held: 1, Waiting: 1})
	assert.Contains(t, use, store.ResourceUse{Key: "pr:x#1", Capacity: 1, Held: 1, Waiting: 1})
	assert.Contains(t, use, store.ResourceUse{Key: "other", Held: 0, Waiting: 2})

	// Close frees the slot.
	_, err = s.CloseTicket(ctx, "s1", "js-a", &ca, "done")
	require.NoError(t, err)
	got, cb, err := s.ClaimNext(ctx, req("s3"))
	require.NoError(t, err)
	assert.Equal(t, "js-b", got.ID, "the waiting emu step is claimable once the slot frees")

	// Release frees it.
	_, err = s.Release(ctx, cb, "x", time.Time{})
	require.NoError(t, err)
	assert.Contains(t, ready(), "js-b")
	_, cb, err = s.Claim(ctx, "js-b", "s3", lease)
	require.NoError(t, err)

	// A dead holder frees it.
	_, err = s.ReleaseHolder(ctx, "controller", "s3", "session died")
	require.NoError(t, err)
	assert.Contains(t, ready(), "js-b")
	_, err = s.Heartbeat(ctx, cb, lease)
	require.ErrorIs(t, err, store.ErrNotHolder)

	// Lease expiry frees it.
	assert.NotContains(t, ready(), "js-d")
	c.Advance(lease + time.Minute)
	_, err = s.SweepLeases(ctx, "controller")
	require.NoError(t, err)
	assert.Contains(t, ready(), "js-d", "the expired pr:x#1 claim frees the key")
	_, err = s.Heartbeat(ctx, cc, lease)
	require.ErrorIs(t, err, store.ErrNotHolder)

	// No cap: unlimited.
	require.NoError(t, s.SetResourceCaps(ctx, nil))
	_, _, err = s.Claim(ctx, "js-c", "s4", lease)
	require.NoError(t, err)
	assert.Contains(t, ready(), "js-d", "caps removed: no limit")

	require.ErrorIs(t, s.SetResourceCaps(ctx, map[string]int{"x": 0}), store.ErrInvalid)
}
