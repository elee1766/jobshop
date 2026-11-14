package control

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// Housekeeping (JOB-197). Pruning, journal compression and store
// maintenance ran as the last steps of the supervisor pass, under tickMu:
// a slow prune delayed the lease sweep and scaling behind it. They run
// here, on their own loop and without tickMu, so a pass is sweep,
// triggers, placement, supervise and runs. Each step keeps its own
// interval guard, so the loop's period only bounds how late one is.
//
// What they touch is theirs alone: the *At fields on Controller are
// written only under housekeepMu, and the store takes its own locks.
// A controller with Options.Pass (tests) does not run the loop: a
// requested pass runs housekeeping after the tick, so a test drives it
// the same way it drives everything else.

// housekeepEvery is how often the loop looks for a due step.
const housekeepEvery = time.Minute

// housekeepingSteps are the step names, in run order.
var housekeepingSteps = []string{
	"prune_inbox", "prune_write_keys", "prune_sync", "prune_artifacts", "compress_journals", "maintain_store",
}

// pumpHousekeeping runs housekeeping once at start and then every
// housekeepEvery until the controller stops.
func (c *Controller) pumpHousekeeping() {
	if c.opt.Pass != nil {
		return
	}
	tick := c.clk().NewTicker(housekeepEvery)
	defer tick.Stop()
	for {
		c.housekeeping()
		select {
		case <-c.ctx.Done():
			return
		case <-tick.C():
		}
	}
}

// housekeeping runs each step that is due. Runs do not overlap.
func (c *Controller) housekeeping() {
	if c.ctx.Err() != nil {
		return
	}
	c.housekeepMu.Lock()
	defer c.housekeepMu.Unlock()
	ctx, span := c.rootSpan(c.ctx, "housekeeping")
	defer span.End()
	cfg := c.config()
	c.hkStep(ctx, cfg, "prune_inbox", c.pruneInbox)
	c.hkStep(ctx, cfg, "prune_write_keys", c.pruneWriteKeys)
	c.hkStep(ctx, cfg, "prune_sync", c.pruneSync)
	c.hkStep(ctx, cfg, "prune_artifacts", c.pruneArtifacts)
	c.hkStep(ctx, cfg, "compress_journals", c.maybeCompressJournals)
	c.hkStep(ctx, cfg, "maintain_store", c.maintainStore)
}

// hkStep runs one housekeeping step under a span and records its
// duration as jobshop.housekeeping.duration{step}. It is not a tick step:
// the tick's step histogram and status carry the pass only.
func (c *Controller) hkStep(ctx context.Context, cfg *config.Config, name string, fn func(ctx context.Context, cfg *config.Config)) {
	ctx, span := telemetry.StartSpan(ctx, "housekeeping "+name, attribute.String("jobshop.housekeeping.step", name))
	defer span.End()
	start := time.Now()
	defer c.metrics.SinceHousekeeping(ctx, name, start)
	// A slow tick lists the steps that ran during it (JOB-212).
	if c.probe != nil {
		done := c.probe.houseStart(name, start)
		defer func() { done(time.Now()) }()
	}
	fn(ctx, cfg)
}
