package storetest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// testListShapes covers the Filter fields that replaced the per-shape
// reads (#579): Kinds, Statuses, Unheld, Key, Created, the orders and
// NoLimit.
func testListShapes(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	list := func(f store.Filter) []string {
		t.Helper()
		l, err := s.List(ctx, f)
		require.NoError(t, err)
		return ids(l)
	}
	t0 := c.Now()
	a := task(t, s, "a", "dev", prio(3))
	c.Advance(time.Second)
	b := task(t, s, "b", "dev", prio(1))
	c.Advance(time.Second)
	held := task(t, s, "held", "dev", func(n *store.NewTicket) { n.Holder = "s1" })
	c.Advance(time.Second)
	rf, _ := json.Marshal(store.RunFields{Workflow: "w", Key: "k", State: store.RunRunning})
	msg, err := s.Create(ctx, "test", store.NewTicket{Project: proj, Kind: store.KindRun, Title: "r", Fields: rf})
	require.NoError(t, err)
	_, err = store.Close(ctx, s, "test", a.ID, nil, "done")
	require.NoError(t, err)

	assert.Equal(t, []string{b.ID, held.ID, msg.ID}, list(store.Filter{Statuses: store.OpenStatuses}))
	assert.Equal(t, []string{b.ID, held.ID, msg.ID}, list(store.Filter{Statuses: []store.Status{store.StatusOpen}}))
	assert.Equal(t, []string{a.ID}, list(store.Filter{Status: store.StatusClosed}))
	assert.Equal(t, []string{msg.ID}, list(store.Filter{Kinds: []store.Kind{store.KindMessage, store.KindRun}}))
	assert.Equal(t, []string{msg.ID}, list(store.Filter{Key: "k", Workflow: "w", Statuses: store.OpenStatuses}))
	assert.Empty(t, list(store.Filter{Key: "other"}))
	assert.Equal(t, []string{b.ID, msg.ID}, list(store.Filter{Status: store.StatusOpen, Unheld: true}))

	// Orders, and the oldest order's cursor.
	assert.Equal(t, []string{a.ID, b.ID, held.ID, msg.ID}, list(store.Filter{Order: store.OrderOldest}))
	assert.Equal(t, []string{held.ID, msg.ID}, list(store.Filter{Order: store.OrderOldest, After: b.ID}))
	assert.Equal(t, []string{msg.ID, held.ID, b.ID, a.ID}, list(store.Filter{Order: store.OrderNewest}))
	_, err = s.List(ctx, store.Filter{Order: store.OrderIndex, After: a.ID})
	require.ErrorIs(t, err, store.ErrInvalid)
	_, err = s.List(ctx, store.Filter{Order: "sideways"})
	require.ErrorIs(t, err, store.ErrInvalid)

	// Created is [Since, Until) on created_at, not updated_at.
	assert.Equal(t, []string{b.ID, held.ID}, list(store.Filter{
		Created: &store.TimeRange{Since: t0.Add(time.Second), Until: t0.Add(3 * time.Second)}, Order: store.OrderOldest,
	}))
	assert.Len(t, list(store.Filter{Limit: store.NoLimit}), 4)
	assert.Len(t, list(store.Filter{Limit: 2}), 2)

	// Ready is open, unblocked, claimable work, held or not; with a
	// kind it is that kind.
	assert.Equal(t, []string{b.ID, held.ID}, list(store.Filter{Project: proj, Ready: true}))
	assert.Equal(t, []string{msg.ID}, list(store.Filter{Ready: true, Kind: store.KindRun}))
}
