package sqlite

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite/sqldb"
)

// SetResources replaces the resource capacities and provider names in
// one transaction.
func (s *Store) SetResources(ctx context.Context, r domain.ResourceConfig) error {
	caps := r.Caps
	return s.tx(ctx, func(q sqldb.Querier) error {
		if err := setProviders(ctx, q, r.Providers); err != nil {
			return err
		}
		if err := q.DeleteResourceCaps(ctx); err != nil {
			return err
		}
		for _, name := range slices.Sorted(maps.Keys(caps)) {
			n := caps[name]
			if n <= 0 {
				return fmt.Errorf("%w: resource %s: capacity must be > 0", domain.ErrInvalid, name)
			}
			if err := q.InsertResourceCap(ctx, sqldb.InsertResourceCapParams{Name: name, Capacity: int64(n)}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Resources reports key use, resource waits and slots (pool's only when
// pool is set, with its demand).
func (s *Store) Resources(ctx context.Context, pool string) (domain.ResourceReport, error) {
	var r domain.ResourceReport
	use, err := s.resourceUse(ctx)
	if err != nil {
		return r, err
	}
	r.Use = use
	if r.Waits, err = s.resourceWaits(ctx, use); err != nil {
		return r, err
	}
	if r.Slots, err = s.slots(ctx, pool); err != nil {
		return r, err
	}
	if pool != "" {
		if r.Demand, err = s.slotDemand(ctx, pool, r.Slots); err != nil {
			return r, err
		}
	}
	return r, nil
}

// resourceUse reports every key held or waited on, with its capacity.
func (s *Store) resourceUse(ctx context.Context) ([]domain.ResourceUse, error) {
	caps, err := s.rq.ResourceCaps(ctx)
	if err != nil {
		return nil, err
	}
	capOf := map[string]int{}
	for _, c := range caps {
		capOf[c.Name] = int(c.Capacity)
	}
	use, err := s.rq.ResourceUse(ctx)
	if err != nil {
		return nil, err
	}
	wait, err := s.rq.ResourceWaiting(ctx)
	if err != nil {
		return nil, err
	}
	by := map[string]*domain.ResourceUse{}
	get := func(k string) *domain.ResourceUse {
		u, ok := by[k]
		if !ok {
			c, exact := capOf[k]
			if !exact {
				c = capOf[domain.ResourceName(k)]
			}
			u = &domain.ResourceUse{Key: k, Capacity: c}
			by[k] = u
		}
		return u
	}
	for _, r := range use {
		get(r.Rkey).Held = int(r.Held)
	}
	for _, r := range wait {
		get(r.Rkey).Waiting = int(r.Waiting)
	}
	out := make([]domain.ResourceUse, 0, len(by))
	for _, k := range slices.Sorted(maps.Keys(by)) {
		out = append(out, *by[k])
	}
	return out, nil
}

// resourceWaits lists tickets waiting on a full resource key, in claim
// order per key (#408). A ticket naming several full keys is listed under
// the first.
func (s *Store) resourceWaits(ctx context.Context, use []domain.ResourceUse) ([]domain.ResourceWait, error) {
	full := map[string]domain.ResourceUse{}
	for _, u := range use {
		if u.Capacity > 0 && u.Held >= u.Capacity {
			full[u.Key] = u
		}
	}
	if len(full) == 0 {
		return nil, nil
	}
	rows, err := s.rq.ResourceWaiters(ctx, ns(s.now()))
	if err != nil {
		return nil, err
	}
	var out []domain.ResourceWait
	seen := map[string]bool{}
	ahead := map[string]int{}
	for _, r := range rows {
		u, ok := full[r.Rkey]
		if !ok || seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		out = append(out, domain.ResourceWait{Ticket: r.ID, Key: r.Rkey, Ahead: ahead[r.Rkey], Held: u.Held, Capacity: u.Capacity})
		ahead[r.Rkey]++
	}
	return out, nil
}
