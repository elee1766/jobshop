package control

import (
	"context"
	"errors"
	"fmt"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// Per-session isolation (#540): with isolation.user.per_session each live
// session of a pool runs as its own host uid/gid, map_uid+slot and
// map_gid+slot. The slot is chosen when the session is created (lowest
// free among the pool's open sessions) and kept in the session's fields,
// so a controller restart reuses it and two sessions never share one. A
// full range is a capacity limit: scale up stops at it.
//
// The controller is unprivileged and cannot chown to a subuid, so it
// stays the owner of the session dir and worktree and grants the slot's
// uid, itself and the pool's first group through POSIX ACLs. The agent
// gets no supplementary groups, so another session's uid matches no
// entry and falls to "other", which has none.

// allocSlot returns the lowest slot in [0, n) not used by another open
// session of pool, or ok false when all are taken. Called under
// c.slotMu by newSession, so two creates cannot pick the same slot.
func (c *Controller) allocSlot(ctx context.Context, pool string, n int) (int, bool, error) {
	used, err := c.usedSlots(ctx, pool)
	if err != nil {
		return 0, false, err
	}
	for s := range n {
		if !used[s] {
			return s, true, nil
		}
	}
	return 0, false, nil
}

// usedSlots is the slots of pool's open sessions.
func (c *Controller) usedSlots(ctx context.Context, pool string) (map[int]bool, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Pool: pool})
	if err != nil {
		return nil, err
	}
	used := map[int]bool{}
	for _, t := range open {
		if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.Slot != nil {
			used[*f.Slot] = true
		}
	}
	return used, nil
}

// slotCapacity is how many more sessions pool may start, or -1 for no
// per-session limit.
func (c *Controller) slotCapacity(ctx context.Context, cfg *config.Config, pool string) int {
	u := cfg.IsolationFor(pool).User
	if u == nil || u.Slots() == 0 {
		return -1
	}
	used, err := c.usedSlots(ctx, pool)
	if err != nil {
		return 0
	}
	n := 0
	for s := range u.Slots() {
		if !used[s] {
			n++
		}
	}
	return n
}

// errSlotsFull is a pool whose per-session range is all in use.
var errSlotsFull = errors.New("per_session range full")

// sessionSlot picks a slot for a new session of pool, or nil when the
// pool does not use per-session ids.
func (c *Controller) sessionSlot(ctx context.Context, cfg *config.Config, pool string) (*int, error) {
	u := cfg.IsolationFor(pool).User
	if u == nil || u.Slots() == 0 {
		return nil, nil
	}
	s, ok, err := c.allocSlot(ctx, pool, u.Slots())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%w: pool %s uses all %d per-session ids", errSlotsFull, pool, u.Slots())
	}
	return &s, nil
}
