package control

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// drainAll is the key in Controller.drained that drains every pool.
const drainAll = "*"

const resolutionDrained = "drained"

// drainPools marks pools drained: desired becomes 0, the supervisor starts
// no sessions for them and their sessions claim no new tickets, ending once
// they hold nothing. The state is saved in scale.json, so it survives a
// controller restart (#281).
func (c *Controller) drainPools(ctx context.Context, sc scope, p rpc.DrainParams) (any, error) {
	if err := c.requireGrant(sc, "drain", "drain"); err != nil {
		return nil, err
	}
	if p.Pool != "" {
		if _, ok := c.config().Pools[p.Pool]; !ok {
			return nil, fmt.Errorf("%w: unknown pool %q", domain.ErrInvalid, p.Pool)
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
	err := c.saveScaleLocked()
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("save %s: %w", scaleFile, err)
	}
	c.log.InfoContext(ctx, "drain changed", "pool", key, "undrain", p.Undrain, "drained", out)
	return rpc.DrainResult{Drained: out}, nil
}

// poolDrained reports whether sessions of pool should take no new work.
func (c *Controller) poolDrained(pool string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.drained[drainAll] || (pool != "" && c.drained[pool])
}
