package storetest

import (
	"context"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Short forms of Store.Slot and Store.Resources for the slot tests.

func addSlot(ctx context.Context, s store.Store, pool, id string, env map[string]string) (domain.Slot, error) {
	sl, _, err := s.Slot(ctx, "", domain.SlotEvent{Op: domain.SlotAdd, Pool: pool, ID: id, Env: env})
	return sl, err
}

func removeSlot(ctx context.Context, s store.Store, pool, id string) (domain.Slot, error) {
	sl, _, err := s.Slot(ctx, "", domain.SlotEvent{Op: domain.SlotRemove, Pool: pool, ID: id})
	return sl, err
}

func slotReset(ctx context.Context, s store.Store, pool, id string, ok bool) error {
	_, _, err := s.Slot(ctx, "", domain.SlotEvent{Op: domain.SlotReset, Pool: pool, ID: id, OK: ok})
	return err
}

func slotUnhealthy(ctx context.Context, s store.Store, actor, pool, id, reason string) (domain.Slot, string, error) {
	return s.Slot(ctx, actor, domain.SlotEvent{Op: domain.SlotUnhealthy, Pool: pool, ID: id, Reason: reason})
}

func slotHealthFailed(ctx context.Context, s store.Store, pool, id string) (domain.Slot, error) {
	sl, _, err := s.Slot(ctx, "", domain.SlotEvent{Op: domain.SlotHealth, Pool: pool, ID: id})
	return sl, err
}

func allSlots(ctx context.Context, s store.Store) ([]domain.Slot, error) {
	r, err := s.Resources(ctx, "")
	return r.Slots, err
}

func ticketSlots(ctx context.Context, s store.Store, id string) ([]domain.Slot, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	r, err := s.Resources(ctx, "")
	return domain.TicketSlots(r, t), err
}

func slotDemand(ctx context.Context, s store.Store, pool string) (domain.SlotDemand, error) {
	r, err := s.Resources(ctx, pool)
	return r.Demand, err
}

func resourceUse(ctx context.Context, s store.Store) ([]domain.ResourceUse, error) {
	r, err := s.Resources(ctx, "")
	return r.Use, err
}

func resourceWaits(ctx context.Context, s store.Store) ([]domain.ResourceWait, error) {
	r, err := s.Resources(ctx, "")
	return r.Waits, err
}

func setCaps(ctx context.Context, s store.Store, caps map[string]int) error {
	return s.SetResources(ctx, domain.ResourceConfig{Caps: caps})
}

func setProviders(ctx context.Context, s store.Store, names []string) error {
	return s.SetResources(ctx, domain.ResourceConfig{Providers: names})
}
