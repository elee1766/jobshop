package control

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// drainAll is the key in Controller.drained that drains every pool.
const drainAll = "*"

const resolutionDrained = "drained"

// drainPools marks pools drained: the supervisor starts no sessions for them and
// their sessions claim no new tickets, ending once they hold nothing. The
// state is in memory only; a controller restart undrains everything.
func (c *Controller) drainPools(_ context.Context, sc scope, p rpc.DrainParams) (any, error) {
	if err := requireAdmin(sc, "drain"); err != nil {
		return nil, err
	}
	if p.Pool != "" {
		if _, ok := c.config().Pools[p.Pool]; !ok {
			return nil, fmt.Errorf("%w: unknown pool %q", store.ErrInvalid, p.Pool)
		}
	}
	key := p.Pool
	if key == "" {
		key = drainAll
	}
	c.mu.Lock()
	switch {
	case !p.Undrain:
		c.drained[key] = true
	case p.Pool == "":
		clear(c.drained)
	default:
		delete(c.drained, key)
	}
	out := slices.Sorted(maps.Keys(c.drained))
	c.mu.Unlock()
	c.log.Info("drain changed", "pool", key, "undrain", p.Undrain, "drained", out)
	return rpc.DrainResult{Drained: out}, nil
}

// poolDrained reports whether sessions of pool should take no new work.
func (c *Controller) poolDrained(pool string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.drained[drainAll] || (pool != "" && c.drained[pool])
}
