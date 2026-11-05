package storetest

import (
	"context"

	"github.com/elee1766/jobshop/pkg/store"
)

// Short forms of Store.Slot and Store.Resources for the slot tests.

func addSlot(ctx context.Context, s store.Store, pool, id string, env map[string]string) (store.Slot, error) {
	sl, _, err := s.Slot(ctx, "", store.SlotEvent{Op: store.SlotAdd, Pool: pool, ID: id, Env: env})
	return sl, err
}

func removeSlot(ctx context.Context, s store.Store, pool, id string) (store.Slot, error) {
	sl, _, err := s.Slot(ctx, "", store.SlotEvent{Op: store.SlotRemove, Pool: pool, ID: id})
	return sl, err
}

func slotReset(ctx context.Context, s store.Store, pool, id string, ok bool) error {
	_, _, err := s.Slot(ctx, "", store.SlotEvent{Op: store.SlotReset, Pool: pool, ID: id, OK: ok})
	return err
}

func slotUnhealthy(ctx context.Context, s store.Store, actor, pool, id, reason string) (store.Slot, string, error) {
	return s.Slot(ctx, actor, store.SlotEvent{Op: store.SlotUnhealthy, Pool: pool, ID: id, Reason: reason})
}

func allSlots(ctx context.Context, s store.Store) ([]store.Slot, error) {
	r, err := s.Resources(ctx, "")
	return r.Slots, err
}

func ticketSlots(ctx context.Context, s store.Store, id string) ([]store.Slot, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	r, err := s.Resources(ctx, "")
	return store.TicketSlots(r, t), err
}

func slotDemand(ctx context.Context, s store.Store, pool string) (store.SlotDemand, error) {
	r, err := s.Resources(ctx, pool)
	return r.Demand, err
}

func resourceUse(ctx context.Context, s store.Store) ([]store.ResourceUse, error) {
	r, err := s.Resources(ctx, "")
	return r.Use, err
}

func resourceWaits(ctx context.Context, s store.Store) ([]store.ResourceWait, error) {
	r, err := s.Resources(ctx, "")
	return r.Waits, err
}

func setCaps(ctx context.Context, s store.Store, caps map[string]int) error {
	return s.SetResources(ctx, store.ResourceConfig{Caps: caps})
}

func setProviders(ctx context.Context, s store.Store, names []string) error {
	return s.SetResources(ctx, store.ResourceConfig{Providers: names})
}
