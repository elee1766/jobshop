package storerules_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/storerules"
)

var now = time.Date(2026, 11, 6, 1, 2, 3, 456789000, time.UTC)

func TestClaimFence(t *testing.T) {
	tk := domain.Ticket{ID: "t1", Status: domain.StatusOpen, Fence: 3}
	c, d, fresh := storerules.Claim(&tk, "h", time.Minute, now)
	assert.True(t, fresh)
	assert.Equal(t, int64(4), c.Fence)
	assert.Equal(t, now.Add(time.Minute).Round(0), c.LeaseUntil)
	assert.Equal(t, false, d["resumed"])

	// A resume by the same holder keeps the fence.
	c2, d2, fresh := storerules.Claim(&tk, "h", time.Minute, now)
	assert.False(t, fresh)
	assert.Equal(t, c.Fence, c2.Fence)
	assert.Equal(t, true, d2["resumed"])
	require.NoError(t, storerules.Holds(tk, c2))
	assert.ErrorIs(t, storerules.Holds(tk, store.Claim{Holder: "h", Fence: 3}), domain.ErrNotHolder)
}

func TestCanClaim(t *testing.T) {
	open := domain.Ticket{ID: "t", Status: domain.StatusOpen}
	require.NoError(t, storerules.CanClaim(open, "h", false, false))
	assert.ErrorIs(t, storerules.CanClaim(open, "h", true, false), domain.ErrNoWork)
	assert.ErrorIs(t, storerules.CanClaim(open, "h", false, true), domain.ErrNoWork)
	assigned := domain.Ticket{ID: "t", Status: domain.StatusOpen, Holder: "other"}
	assert.ErrorIs(t, storerules.CanClaim(assigned, "h", false, false), domain.ErrNotHolder)
	held := domain.Ticket{ID: "t", Status: domain.StatusInProgress, Holder: "h"}
	require.NoError(t, storerules.CanClaim(held, "h", false, true), "a resume does not wait for a slot")
	assert.ErrorIs(t, storerules.CanClaim(domain.Ticket{Status: domain.StatusClosed}, "h", false, false), domain.ErrClosed)
	assert.ErrorIs(t, storerules.StillCandidate(assigned, "h"), domain.ErrConflict)
}

func TestHandoffBumpsFence(t *testing.T) {
	tk := domain.Ticket{ID: "t", Status: domain.StatusInProgress, Holder: "h", Fence: 2, Pool: "a"}
	c := store.Claim{Holder: "h", Fence: 2}
	d, err := storerules.Handoff(&tk, &c, "b", "why")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, tk.Status)
	assert.Equal(t, int64(3), tk.Fence)
	assert.Equal(t, "b", tk.Pool)
	assert.Equal(t, map[string]any{"from_pool": "a", "pool": "b", "holder": "h", "note": "why"}, d)
	_, err = storerules.Handoff(&tk, nil, "", "")
	assert.ErrorIs(t, err, domain.ErrInvalid)
}

func TestReleaseDeferIsNotFailure(t *testing.T) {
	tk := domain.Ticket{ID: "t", Kind: domain.KindStep, Status: domain.StatusInProgress, Holder: "h", Fence: 1, Fields: json.RawMessage(`{}`)}
	until := now.Add(time.Hour)
	d, note, err := storerules.Release(&tk, store.Claim{Holder: "h", Fence: 1}, "ci", until, now)
	require.NoError(t, err)
	assert.NotContains(t, d, "failures")
	assert.Contains(t, note, "Released until")
	assert.Equal(t, domain.StatusOpen, tk.Status)
}

func TestCloseRebaseReopensStep(t *testing.T) {
	tk := domain.Ticket{ID: "t", Kind: domain.KindStep, Status: domain.StatusInProgress, Holder: "h"}
	_, note, err := storerules.Close(&tk, nil, domain.ResolutionRebase, now)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, tk.Status)
	assert.Empty(t, tk.Resolution)
	assert.True(t, tk.ClosedAt.IsZero())
	assert.Contains(t, note, "rebase")
	_, _, err = storerules.Close(&domain.Ticket{Status: domain.StatusClosed}, nil, "done", now)
	assert.ErrorIs(t, err, domain.ErrClosed)
}

func TestExpireUnclaimedIsNotFailure(t *testing.T) {
	tk := domain.Ticket{ID: "t", Kind: domain.KindStep, Status: domain.StatusOpen, Holder: "h"}
	d, note, err := storerules.Expire(&tk, "release", map[string]any{"reason": "gone"}, now)
	require.NoError(t, err)
	assert.Empty(t, note)
	assert.Equal(t, "gone", d["reason"])
	assert.Empty(t, tk.Holder)
}

func TestApplyUpdateDiff(t *testing.T) {
	title, same := "new", "p"
	tk := domain.Ticket{ID: "t", Title: "old", Pool: "p", Status: domain.StatusInProgress, Revision: 4}
	d, err := storerules.ApplyUpdate(&tk, store.Update{Title: &title, Pool: &same})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"title": "new"}, d)
	h := "x"
	_, err = storerules.ApplyUpdate(&tk, store.Update{Holder: &h})
	assert.ErrorIs(t, err, domain.ErrInvalid)
	_, err = storerules.ApplyUpdate(&tk, store.Update{IfRevision: 3})
	assert.ErrorIs(t, err, domain.ErrConflict)

	prev := storerules.Bump(&tk, now)
	assert.Equal(t, int64(4), prev)
	assert.Equal(t, int64(5), tk.Revision)
	assert.Equal(t, now.Round(0), tk.UpdatedAt)
}

func TestFieldColumns(t *testing.T) {
	c := storerules.FieldColumns(json.RawMessage(`{"key":"k","state":3,"index":2,"wait":{"a":1}}`))
	require.NotNil(t, c.Key)
	assert.Equal(t, "k", *c.Key)
	assert.Nil(t, c.State, "a non-string is NULL")
	assert.Nil(t, c.To)
	require.NotNil(t, c.Index)
	assert.Equal(t, int64(2), *c.Index)
	assert.True(t, c.Waiting)
	assert.False(t, storerules.FieldColumns(json.RawMessage(`{"wait":null}`)).Waiting)
	assert.Equal(t, storerules.Columns{}, storerules.FieldColumns(json.RawMessage(`bad`)))
}
