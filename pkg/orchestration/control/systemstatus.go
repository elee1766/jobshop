package control

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func (c *Controller) finishTick(start time.Time) {
	c.statusMu.Lock()
	c.status.LastTick = c.now()
	c.status.TickSeconds = time.Since(start).Seconds()
	c.status.Steps = maps.Clone(c.tickSteps)
	c.tickSamples = append(c.tickSamples, c.status.TickSeconds)
	if len(c.tickSamples) > 60 {
		c.tickSamples = slices.Clone(c.tickSamples[len(c.tickSamples)-60:])
	}
	samples := slices.Sorted(slices.Values(c.tickSamples))
	c.status.Samples = len(samples)
	c.status.P95Seconds = samples[(len(samples)-1)*95/100]
	status := c.status
	status.Steps = maps.Clone(c.status.Steps)
	c.statusMu.Unlock()
	if c.events != nil {
		c.events.publish([]rpc.Event{{Kind: rpc.EventKindSystem, System: &status, Time: status.LastTick}})
	}
}

func (c *Controller) systemStatus(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "pool-list", "system status"); err != nil {
		return nil, err
	}
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	out := c.status
	out.Steps = maps.Clone(c.status.Steps)
	return out, nil
}
