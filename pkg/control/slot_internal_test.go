package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

const slotYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev:
    agent: fake
    dir: %[1]q
    isolation:
      user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: ["100"], per_session: {range: 2}}
  plain: {agent: fake, dir: %[1]q}
`

// createSlotSession does what newSession does for the slot (pick under
// slotMu, store it on the row) without starting a runner.
func createSlotSession(t *testing.T, c *Controller, pool string) (store.Ticket, error) {
	t.Helper()
	c.slotMu.Lock()
	defer c.slotMu.Unlock()
	slot, err := c.sessionSlot(t.Context(), pool)
	if err != nil {
		return store.Ticket{}, err
	}
	fields, err := json.Marshal(store.SessionFields{Agent: "fake", Slot: slot})
	require.NoError(t, err)
	return store.Create(t.Context(), c.store, "test", store.NewTicket{Project: "js", Kind: store.KindSession, Title: "s", Pool: pool, Fields: fields})
}

func slotOf(t *testing.T, tk store.Ticket) int {
	t.Helper()
	f, err := store.DecodeFields[store.SessionFields](tk)
	require.NoError(t, err)
	require.NotNil(t, f.Slot)
	return *f.Slot
}

// TestSlotAllocation (#540): lowest free slot, reuse after a session
// ends, a full range blocks both create and scale up, and pools without
// per_session get no slot and no limit.
func TestSlotAllocation(t *testing.T) {
	c, _, _ := stallControllerYAML(t, slotYAML)
	ctx := t.Context()

	assert.Equal(t, 2, c.slotCapacity(ctx, "dev"))
	assert.Equal(t, -1, c.slotCapacity(ctx, "plain"), "no per_session: no limit")
	s, err := c.sessionSlot(ctx, "plain")
	require.NoError(t, err)
	assert.Nil(t, s)

	a, err := createSlotSession(t, c, "dev")
	require.NoError(t, err)
	b, err := createSlotSession(t, c, "dev")
	require.NoError(t, err)
	assert.Equal(t, 0, slotOf(t, a))
	assert.Equal(t, 1, slotOf(t, b))
	assert.Equal(t, 0, c.slotCapacity(ctx, "dev"), "full range: scale up gets no starts")

	_, err = createSlotSession(t, c, "dev")
	require.ErrorIs(t, err, errSlotsFull)

	// Ending the slot-0 session frees slot 0, the lowest free.
	_, err = store.Close(ctx, c.store, "test", a.ID, nil, "done")
	require.NoError(t, err)
	assert.Equal(t, 1, c.slotCapacity(ctx, "dev"))
	d, err := createSlotSession(t, c, "dev")
	require.NoError(t, err)
	assert.Equal(t, 0, slotOf(t, d))
}

// TestRemoveSlotHome (JOB-28): closing a per_session session deletes its
// slot's home and leaves the other slot's; an open session keeps its own.
func TestRemoveSlotHome(t *testing.T) {
	c, _, _ := stallControllerYAML(t, slotYAML)
	ctx := t.Context()
	a, err := createSlotSession(t, c, "dev")
	require.NoError(t, err)
	b, err := createSlotSession(t, c, "dev")
	require.NoError(t, err)
	ha, hb := c.mappedHomeDir("dev", new(slotOf(t, a))), c.mappedHomeDir("dev", new(slotOf(t, b)))
	for _, h := range []string{ha, hb} {
		require.NoError(t, os.MkdirAll(filepath.Join(h, ".cache"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(h, ".cache", "f"), []byte("x"), 0o600))
	}

	c.removeSlotHome(a)
	assert.DirExists(t, ha, "open session keeps its home")

	closed, err := store.Close(ctx, c.store, "test", a.ID, nil, "done")
	require.NoError(t, err)
	c.removeSlotHome(closed)
	assert.NoDirExists(t, ha)
	assert.DirExists(t, hb)
}

// TestSlotRestart: a new controller over the same store sees the
// persisted slots, so it neither reuses a live session's slot nor moves
// a session to another one.
func TestSlotRestart(t *testing.T) {
	c, _, _ := stallControllerYAML(t, slotYAML)
	a, err := createSlotSession(t, c, "dev")
	require.NoError(t, err)

	c2 := &Controller{root: c.root, live: c.live, store: c.store, opt: c.opt, log: c.log}
	got, err := c2.store.Get(t.Context(), a.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, slotOf(t, got), "slot read back from the row")
	b, err := createSlotSession(t, c2, "dev")
	require.NoError(t, err)
	assert.Equal(t, 1, slotOf(t, b), "slot 0 is still taken after restart")
}

// TestApplySlotMapCount: the session reserves the ids from its slot to
// the end of the pool's block, so acphost rejects a range that runs past
// the subuid entry at slot 0, not only at the last slot (#722).
func TestApplySlotMapCount(t *testing.T) {
	u := config.IsolationUser{MapUID: 100000, MapGID: 200000, Groups: []string{"100"}, PerSession: &config.PerSession{Range: 4}}
	for slot, want := range []int{4, 3, 2, 1} {
		spec := host.Spec{User: &host.User{MapUID: 100000, MapGID: 200000}}
		require.NoError(t, applySlot(&spec, u, &slot))
		assert.Equal(t, want, spec.User.MapCount, "slot %d", slot)
		assert.Equal(t, 100000+spec.User.MapCount+slot, 100004, "block end")
	}
}
