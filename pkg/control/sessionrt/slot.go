package sessionrt

import "context"

// Per-session isolation (#540): with isolation.user.per_session each live
// session of a pool runs as its own host uid/gid, map_uid+slot and
// map_gid+slot. The slot is chosen when the session is created (lowest
// free among the pool's open sessions) and kept in the session's fields,
// so a controller restart reuses it and two sessions never share one. A
// full range is a capacity limit: scale up stops at it.
//
// The controller is unprivileged and cannot chown to a subuid, so it
// stays the owner of the session dir and worktree and grants the slot's
// uid, itself and the pool's first group through POSIX ACLs (ApplySlot).
// The agent gets no supplementary groups, so another session's uid
// matches no entry and falls to "other", which has none.

// SlotSource reports which slots the open sessions of a pool hold. The
// controller implements it over the store, where the slot lives in the
// session's durable fields. sessionrt reads nothing else of a session.
type SlotSource interface {
	UsedSlots(ctx context.Context, pool string) (map[int]bool, error)
}

// AllocSlot returns the lowest slot in [0, n) not used by another open
// session of pool, or ok false when all are taken. The caller holds the
// controller's slotMu from this call until the session row carries the
// slot, so two creates cannot pick the same one.
func AllocSlot(ctx context.Context, src SlotSource, pool string, n int) (slot int, ok bool, err error) {
	used, err := src.UsedSlots(ctx, pool)
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

// FreeSlots is how many slots in [0, n) pool has left.
func FreeSlots(ctx context.Context, src SlotSource, pool string, n int) (int, error) {
	used, err := src.UsedSlots(ctx, pool)
	if err != nil {
		return 0, err
	}
	free := 0
	for s := range n {
		if !used[s] {
			free++
		}
	}
	return free, nil
}
