package sqlstore

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// SetResourceCaps replaces the resource capacities in one transaction.
func (s *Store) SetResourceCaps(ctx context.Context, caps map[string]int) error {
	return s.tx(ctx, func(q sqlcdb.Querier) error {
		if err := q.DeleteResourceCaps(ctx); err != nil {
			return err
		}
		for _, name := range slices.Sorted(maps.Keys(caps)) {
			n := caps[name]
			if n <= 0 {
				return fmt.Errorf("%w: resource %s: capacity must be > 0", store.ErrInvalid, name)
			}
			if err := q.InsertResourceCap(ctx, sqlcdb.InsertResourceCapParams{Name: name, Capacity: int64(n)}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Resources reports every key held or waited on, with its capacity.
func (s *Store) Resources(ctx context.Context) ([]store.ResourceUse, error) {
	caps, err := s.q.ResourceCaps(ctx)
	if err != nil {
		return nil, err
	}
	capOf := map[string]int{}
	for _, c := range caps {
		capOf[c.Name] = int(c.Capacity)
	}
	use, err := s.q.ResourceUse(ctx)
	if err != nil {
		return nil, err
	}
	wait, err := s.q.ResourceWaiting(ctx)
	if err != nil {
		return nil, err
	}
	by := map[string]*store.ResourceUse{}
	get := func(k string) *store.ResourceUse {
		u, ok := by[k]
		if !ok {
			c, exact := capOf[k]
			if !exact {
				c = capOf[store.ResourceName(k)]
			}
			u = &store.ResourceUse{Key: k, Capacity: c}
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
	out := make([]store.ResourceUse, 0, len(by))
	for _, k := range slices.Sorted(maps.Keys(by)) {
		out = append(out, *by[k])
	}
	return out, nil
}
