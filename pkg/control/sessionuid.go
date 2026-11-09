package control

import (
	"context"
	"errors"
	"fmt"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/execution/session"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Per-session isolation (#540): see execution/session/slot.go. The slot is part of
// the session's durable fields, so choosing it stays here: newSession
// holds c.slotMu from the pick until the row is created.

// slotStore reads the slots of open sessions from the store. It
// implements session.SlotSource.
type slotStore struct{ store store.Store }

// UsedSlots is the slots of pool's open sessions.
func (s slotStore) UsedSlots(ctx context.Context, pool string) (map[int]bool, error) {
	open, err := s.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: pool})
	if err != nil {
		return nil, err
	}
	used := map[int]bool{}
	for _, t := range open {
		if f, err := domain.DecodeFields[domain.SessionFields](t); err == nil && f.Slot != nil {
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
	n, err := session.FreeSlots(ctx, slotStore{c.store}, pool, u.Slots())
	if err != nil {
		return 0
	}
	return n
}

// errSlotsFull is a pool whose per-session range is all in use.
var errSlotsFull = errors.New("per_session range full")

// sessionSlot picks a slot for a new session of pool, or nil when the
// pool does not use per-session ids. Called under c.slotMu by newSession,
// so two creates cannot pick the same slot.
func (c *Controller) sessionSlot(ctx context.Context, cfg *config.Config, pool string) (*int, error) {
	u := cfg.IsolationFor(pool).User
	if u == nil || u.Slots() == 0 {
		return nil, nil
	}
	s, ok, err := session.AllocSlot(ctx, slotStore{c.store}, pool, u.Slots())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%w: pool %s uses all %d per-session ids", errSlotsFull, pool, u.Slots())
	}
	return &s, nil
}
