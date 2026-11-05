package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// Provider-owned slots (#407). See store.Slot.

// setProviders replaces the provider resource names.
func setProviders(ctx context.Context, q sqlcdb.Querier, names []string) error {
	if err := q.DeleteResourceProviders(ctx); err != nil {
		return err
	}
	for _, n := range slices.Compact(slices.Sorted(slices.Values(names))) {
		if err := q.InsertResourceProvider(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

// Slot applies one slot event.
func (s *Store) Slot(ctx context.Context, actor string, e store.SlotEvent) (store.Slot, string, error) {
	switch e.Op {
	case store.SlotAdd:
		sl, err := s.addSlot(ctx, e.Pool, e.ID, e.Env)
		return sl, "", err
	case store.SlotRemove:
		sl, err := s.removeSlot(ctx, e.Pool, e.ID)
		return sl, "", err
	case store.SlotUnhealthy:
		return s.slotUnhealthy(ctx, actor, e.Pool, e.ID, e.Reason)
	case store.SlotReset:
		sl, err := s.slotReset(ctx, e.Pool, e.ID, e.OK)
		return sl, "", err
	}
	return store.Slot{}, "", fmt.Errorf("%w: slot op %q", store.ErrInvalid, e.Op)
}

func slotOf(r sqlcdb.ResourceSlot) store.Slot {
	sl := store.Slot{
		Pool: r.Pool, ID: r.ID, State: r.State, Ticket: r.Ticket, Fence: r.Fence,
		Removing: r.Removing != 0, UpdatedAt: fromMS(r.UpdatedAt),
	}
	_ = json.Unmarshal([]byte(r.Env), &sl.Env)
	return sl
}

func checkSlot(pool, id string) error {
	if pool == "" || id == "" {
		return fmt.Errorf("%w: resource and slot id are required", store.ErrInvalid)
	}
	return nil
}

func (s *Store) getSlot(ctx context.Context, q sqlcdb.Querier, pool, id string) (sqlcdb.ResourceSlot, error) {
	r, err := q.GetSlot(ctx, sqlcdb.GetSlotParams{Pool: pool, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return r, fmt.Errorf("%w: slot %s/%s", store.ErrNotFound, pool, id)
	}
	return r, err
}

func (s *Store) writeSlot(ctx context.Context, q sqlcdb.Querier, r sqlcdb.ResourceSlot) error {
	return q.UpdateSlot(ctx, sqlcdb.UpdateSlotParams{
		State: r.State, Env: r.Env, Ticket: r.Ticket, Fence: r.Fence, Removing: r.Removing,
		UpdatedAt: ms(s.now()), Pool: r.Pool, ID: r.ID,
	})
}

// addSlot adds a ready slot or updates a known one's env. A gone slot,
// or one marked removing, is back in service.
func (s *Store) addSlot(ctx context.Context, pool, id string, env map[string]string) (out store.Slot, err error) {
	if err := checkSlot(pool, id); err != nil {
		return out, err
	}
	if env == nil {
		env = map[string]string{}
	}
	b, err := json.Marshal(env)
	if err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := s.getSlot(ctx, q, pool, id)
		switch {
		case errors.Is(err, store.ErrNotFound):
			if err := q.InsertSlot(ctx, sqlcdb.InsertSlotParams{Pool: pool, ID: id, Env: string(b), UpdatedAt: ms(s.now())}); err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			r.Env, r.Removing = string(b), 0
			if r.State == store.SlotGone {
				r.State, r.Ticket, r.Fence = store.SlotReady, "", 0
			}
			if err := s.writeSlot(ctx, q, r); err != nil {
				return err
			}
		}
		r, err = s.getSlot(ctx, q, pool, id)
		out = slotOf(r)
		return err
	})
	return out, err
}

// removeSlot deletes a slot that no claim holds; a held or releasing
// one is marked removing and deleted when its claim ends.
func (s *Store) removeSlot(ctx context.Context, pool, id string) (out store.Slot, err error) {
	if err := checkSlot(pool, id); err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := s.getSlot(ctx, q, pool, id)
		if err != nil {
			return err
		}
		if r.State == store.SlotHeld || r.State == store.SlotReleasing {
			r.Removing = 1
			out = slotOf(r)
			return s.writeSlot(ctx, q, r)
		}
		out = slotOf(r)
		out.State = store.SlotGone
		return q.DeleteSlot(ctx, sqlcdb.DeleteSlotParams{Pool: pool, ID: id})
	})
	return out, err
}

// slotUnhealthy marks a slot gone. If a live claim holds it, the ticket
// is reopened in the same transaction without counting a step failure:
// the infrastructure failed, not the work.
func (s *Store) slotUnhealthy(ctx context.Context, actor, pool, id, reason string) (out store.Slot, reopened string, err error) {
	if err := checkSlot(pool, id); err != nil {
		return out, "", err
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := s.getSlot(ctx, q, pool, id)
		if err != nil {
			return err
		}
		if r.State == store.SlotHeld && r.Ticket != "" {
			t, err := s.get(ctx, q, r.Ticket)
			if err != nil && !errors.Is(err, store.ErrNotFound) {
				return err
			}
			if err == nil && t.Status == store.StatusInProgress && t.Fence == r.Fence {
				_, err := s.mutate(ctx, q, t.ID, actor, "release", func(t *store.Ticket) (any, error) {
					reopen(t)
					return map[string]any{"reason": "resource unhealthy", "slot": pool + "/" + id}, nil
				})
				if err != nil {
					return err
				}
				note := fmt.Sprintf("Resource %s slot %s went unhealthy: %s. Requeued without counting a failed attempt.", pool, id, reason)
				if err := s.note(ctx, q, t.ID, note); err != nil {
					return err
				}
				reopened = t.ID
			}
		}
		if r.Removing != 0 {
			out = slotOf(r)
			out.State = store.SlotGone
			return q.DeleteSlot(ctx, sqlcdb.DeleteSlotParams{Pool: pool, ID: id})
		}
		r.State, r.Ticket, r.Fence = store.SlotGone, "", 0
		out = slotOf(r)
		return s.writeSlot(ctx, q, r)
	})
	return out, reopened, err
}

// slots lists every slot, or pool's when set.
func (s *Store) slots(ctx context.Context, pool string) ([]store.Slot, error) {
	rs, err := s.q.ListSlots(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]store.Slot, 0, len(rs))
	for _, r := range rs {
		if pool == "" || r.Pool == pool {
			out = append(out, slotOf(r))
		}
	}
	return out, nil
}

// ReleaseSlots moves held slots whose claim ended to releasing and
// returns all releasing slots. A claim can end on many paths (close,
// release, lease expiry, dead holder); reading the ticket here covers
// all of them without a hook in each.
func (s *Store) ReleaseSlots(ctx context.Context) (out []store.Slot, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		if err := q.MarkSlotsReleasing(ctx, ms(s.now())); err != nil {
			return err
		}
		rs, err := q.ListSlots(ctx)
		if err != nil {
			return err
		}
		for _, r := range rs {
			if r.State == store.SlotReleasing {
				out = append(out, slotOf(r))
			}
		}
		return nil
	})
	return out, err
}

// slotReset records the provider's reset of a releasing slot.
func (s *Store) slotReset(ctx context.Context, pool, id string, ok bool) (out store.Slot, err error) {
	if err := checkSlot(pool, id); err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := s.getSlot(ctx, q, pool, id)
		if err != nil {
			return err
		}
		if r.State != store.SlotReleasing {
			return fmt.Errorf("%w: slot %s/%s is %s, not releasing", store.ErrConflict, pool, id, r.State)
		}
		if r.Removing != 0 {
			out = slotOf(r)
			out.State = store.SlotGone
			return q.DeleteSlot(ctx, sqlcdb.DeleteSlotParams{Pool: pool, ID: id})
		}
		r.Ticket, r.Fence = "", 0
		r.State = store.SlotGone
		if ok {
			r.State = store.SlotReady
		}
		out = slotOf(r)
		return s.writeSlot(ctx, q, r)
	})
	return out, err
}

// slotDemand reports a provider resource's demand from its slots.
func (s *Store) slotDemand(ctx context.Context, pool string, slots []store.Slot) (store.SlotDemand, error) {
	var d store.SlotDemand
	w, err := s.q.ProviderDemand(ctx, pool)
	if err != nil {
		return d, err
	}
	d.Waiting = int(w.Waiting)
	d.OldestWait = fromMS(w.Oldest)
	for _, r := range slots {
		switch r.State {
		case store.SlotReady:
			d.Ready++
		case store.SlotHeld:
			d.Held++
		}
	}
	return d, nil
}

// takeSlots holds one ready slot of each provider resource t needs, in
// the claim transaction. No ready slot is a conflict: another claimer
// took the last one, and the transaction rolls back.
func (s *Store) takeSlots(ctx context.Context, q sqlcdb.Querier, t store.Ticket) error {
	names, err := q.TicketProviderResources(ctx, t.ID)
	if err != nil {
		return err
	}
	for _, n := range names {
		_, err := q.TakeSlot(ctx, sqlcdb.TakeSlotParams{Ticket: t.ID, Fence: t.Fence, UpdatedAt: ms(s.now()), Pool: n})
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s: no ready %s slot", store.ErrConflict, t.ID, n)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
