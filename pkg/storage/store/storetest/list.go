package storetest

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

func testListQuery(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	a := task(t, s, "Review wallet", "dev", func(n *store.NewTicket) {
		n.Body = "receipt-repair"
		n.Meta = json.RawMessage(`{"github":{"repo":"acme/widget","pr":42}}`)
	})
	c.Advance(time.Second)
	b := task(t, s, "WALLET 100%_!", "dev")
	fields, err := json.Marshal(domain.RunFields{Workflow: "trade-review", State: domain.RunRunning, Vars: map[string]string{"subject": "4520@abc"}})
	require.NoError(t, err)
	run, err := store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: domain.KindRun, Title: "Review", Fields: fields})
	require.NoError(t, err)
	_, err = store.Close(ctx, s, "test", a.ID, nil, "done")
	require.NoError(t, err)
	for query, want := range map[string][]string{
		"wallet review": {a.ID}, "receipt-repair": {a.ID}, a.ID: {a.ID},
		"ACME/widget#42": {a.ID}, "4520 trade-review": {run.ID},
		"100%_!": {b.ID}, "%": {b.ID}, "does-not-exist": {}, "%' OR 1=1": {},
	} {
		t.Run(query, func(t *testing.T) {
			got, err := s.List(ctx, store.Filter{Query: query})
			require.NoError(t, err)
			assert.Equal(t, want, ids(got))
		})
	}
	first, err := s.List(ctx, store.Filter{Query: "wallet", Order: store.OrderNewest, Limit: 1})
	require.NoError(t, err)
	require.Equal(t, []string{b.ID}, ids(first))
	next, err := s.List(ctx, store.Filter{Query: "wallet", Order: store.OrderNewest, Limit: 1, After: first[0].ID})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ids(next), "closed matches remain searchable across pages")
	filtered, err := s.List(ctx, store.Filter{Query: "wallet", Status: domain.StatusOpen})
	require.NoError(t, err)
	assert.Equal(t, []string{b.ID}, ids(filtered))
	filtered, err = s.List(ctx, store.Filter{Query: "wallet", Project: "another"})
	require.NoError(t, err)
	assert.Empty(t, filtered)
	for _, query := range []string{strings.Repeat("x", 513), strings.Repeat("x ", 17)} {
		_, err := s.List(ctx, store.Filter{Query: query})
		require.ErrorIs(t, err, domain.ErrInvalid)
	}
}

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
	rf, _ := json.Marshal(domain.RunFields{Workflow: "w", Key: "k", State: domain.RunRunning})
	msg, err := store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: domain.KindRun, Title: "r", Fields: rf})
	require.NoError(t, err)
	_, err = store.Close(ctx, s, "test", a.ID, nil, "done")
	require.NoError(t, err)

	assert.Equal(t, []string{b.ID, held.ID, msg.ID}, list(store.Filter{Statuses: domain.OpenStatuses}))
	assert.Equal(t, []string{b.ID, held.ID, msg.ID}, list(store.Filter{Statuses: []domain.Status{domain.StatusOpen}}))
	assert.Equal(t, []string{a.ID}, list(store.Filter{Status: domain.StatusClosed}))
	assert.Equal(t, []string{msg.ID}, list(store.Filter{Kinds: []domain.Kind{domain.KindMessage, domain.KindRun}}))
	assert.Equal(t, []string{msg.ID}, list(store.Filter{Key: "k", Workflow: "w", Statuses: domain.OpenStatuses}))
	assert.Empty(t, list(store.Filter{Key: "other"}))
	assert.Equal(t, []string{b.ID, msg.ID}, list(store.Filter{Status: domain.StatusOpen, Unheld: true}))

	// Orders, and the oldest order's cursor.
	assert.Equal(t, []string{a.ID, b.ID, held.ID, msg.ID}, list(store.Filter{Order: store.OrderOldest}))
	assert.Equal(t, []string{held.ID, msg.ID}, list(store.Filter{Order: store.OrderOldest, After: b.ID}))
	assert.Equal(t, []string{msg.ID, held.ID, b.ID, a.ID}, list(store.Filter{Order: store.OrderNewest}))
	_, err = s.List(ctx, store.Filter{Order: store.OrderIndex, After: a.ID})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = s.List(ctx, store.Filter{Order: "sideways"})
	require.ErrorIs(t, err, domain.ErrInvalid)

	// Created is [Since, Until) on created_at, not updated_at.
	assert.Equal(t, []string{b.ID, held.ID}, list(store.Filter{
		Created: &store.TimeRange{Since: t0.Add(time.Second), Until: t0.Add(3 * time.Second)}, Order: store.OrderOldest,
	}))
	assert.Len(t, list(store.Filter{Limit: store.NoLimit}), 4)
	assert.Len(t, list(store.Filter{Limit: 2}), 2)

	// Ready is open, unblocked, claimable work, held or not; with a
	// kind it is that kind.
	assert.Equal(t, []string{b.ID, held.ID}, list(store.Filter{Project: proj, Ready: true}))
	assert.Equal(t, []string{msg.ID}, list(store.Filter{Ready: true, Kind: domain.KindRun}))
}
