package control

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"time"

	"github.com/goccy/go-yaml"
	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// liveConfig is the config and the lease parsed from it, swapped as one
// value so readers never see a lease from one file and pools from another.
type liveConfig struct {
	cfg   *config.Config
	lease time.Duration
}

func newLiveConfig(cfg *config.Config) (*liveConfig, error) {
	lease := DefaultLease
	if cfg.Lease != "" {
		var err error
		if lease, err = time.ParseDuration(cfg.Lease); err != nil || lease <= 0 {
			return nil, fmt.Errorf("%s: lease %q: must be a positive duration", config.FileName, cfg.Lease)
		}
	}
	return &liveConfig{cfg: cfg, lease: lease}, nil
}

// config is the active config. A reload between two reads only means the
// second read sees the new one; each value is never mutated after load.
func (c *Controller) config() *config.Config { return c.live.Load().cfg }

func (c *Controller) leaseDur() time.Duration { return c.live.Load().lease }

// scanner is the active secret rule set (#578). A reload swaps it with
// the config; a reload with bad rules fails and keeps the old one.
func (c *Controller) scanner() *redact.Scanner { return c.config().Redact.Scanner() }

// redactor is the active rule set with env's secret values.
func (c *Controller) redactor(env []string) *redact.Redactor { return c.scanner().Redactor(env) }

// Reload re-reads jobshop.yaml and its packs and swaps them in. On any
// error the old config stays active. Pool limits apply to every
// supervisor pass that starts after reload returns; a pass in flight
// finishes on the config it started with (JOB-104). Prompts and turn
// timeouts apply from the next turn, agent command, args and env from the
// next spawned host (adopted hosts go through the env refresh path).
// Running sessions are not stopped. Telemetry settings need a restart.
func (c *Controller) reload() error { return c.reloadWith(c.ctx, "", nil) }

// reloadWith is reload tagged with its cause: "watch" for the config
// watcher (#533), which also names the changed files. A failed watch
// reload raises a config-invalid attention item; any good reload clears
// it. ctx carries the caller's span (the config.reload RPC, the watcher):
// lock waits, the config load, the resource write and each pool's
// preflight are its children, so a slow reload names its slow part
// (JOB-105).
//
// It does not wait for a pass in flight: each pass reads the config once
// at its start (JOB-104). What a pass reads outside the config, pool
// quarantines, changes under c.mu, conservatively: a pool the new
// config's preflight fails is quarantined before the swap, and old
// quarantines are cleared after it.
func (c *Controller) reloadWith(ctx context.Context, by string, files []string) (err error) {
	// A client that gives up (the CLI deadline) must not cancel a reload
	// half done: keep the span and the metrics, drop the cancellation.
	ctx, span := telemetry.StartSpan(context.WithoutCancel(ctx), "config reload", attribute.String("jobshop.reload.by", by))
	defer func() { telemetry.End(span, err) }()
	defer func() {
		if err == nil {
			// Non-blocking: the watcher re-syncs its paths once per burst.
			select {
			case c.configSwapped <- struct{}{}:
			default:
			}
		}
	}()
	start := time.Now()
	c.reloadMu.Lock()
	defer c.reloadMu.Unlock()
	telemetry.RecordSlow(ctx, "wait reloadMu", start, lockSpanThreshold, nil, telemetry.AttrLock.String("reloadMu"))
	// Where a slow reload spent its time (JOB-102).
	locked := time.Now()
	_, loadSpan := telemetry.StartSpan(ctx, "config load")
	cfg, err := config.Load(c.root)
	telemetry.End(loadSpan, err)
	loaded := time.Now()
	if err == nil {
		var lc *liveConfig
		if lc, err = newLiveConfig(cfg); err == nil {
			// Caps first: a failed write keeps the old config and caps.
			resCtx, resSpan := telemetry.StartSpan(ctx, "set resources")
			err = c.store.SetResources(resCtx, store.ResourceConfig{Caps: cfg.ResourceCaps(), Providers: cfg.ResourceProviders()})
			telemetry.End(resSpan, err)
		}
		if err == nil {
			// A host still missing a prerequisite quarantines the pool
			// again now, not at its next spawn (#728). Checked on the new
			// config before the swap, so no pass sees the pool open.
			preflight := time.Now()
			failed := c.preflightFailures(ctx, cfg)
			preflightTook := time.Since(preflight)
			already := map[string]bool{}
			for pool, reason := range failed {
				already[pool] = c.setQuarantine(pool, reason)
			}
			old := c.live.Swap(lc).cfg
			if !reflect.DeepEqual(old.Telemetry, cfg.Telemetry) {
				c.log.WarnContext(ctx, "config reload: telemetry changes need a controller restart")
			}
			if !reflect.DeepEqual(old.Identities, cfg.Identities) {
				// A rotated key or moved installation must not serve old tokens.
				c.minter.Forget()
			}
			// The fix for a permanent spawn error is a config change.
			if cleared := c.unquarantineExcept("", failed); len(cleared) > 0 {
				c.log.InfoContext(ctx, "config reload: pools unquarantined", "pools", cleared)
			}
			for _, pool := range slices.Sorted(maps.Keys(failed)) {
				if !already[pool] {
					c.log.WarnContext(ctx, "pool quarantined: permanent spawn error", "pool", pool, "err", failed[pool])
				}
				c.ensurePoolEscalation(ctx, pool, failed[pool])
			}
			// New weights and cap take effect now; hosts moved on first use.
			c.tuneCgroup(cfg)
			msg := "config reloaded"
			if by != "" {
				msg += " (" + by + ")"
			}
			c.log.InfoContext(ctx, msg, "files", files, "pools", len(cfg.Pools), "agents", len(cfg.Agents), "triggers", len(cfg.Triggers), "workflows", len(cfg.Workflows),
				"took", time.Since(start), "lock_wait", locked.Sub(start), "load", loaded.Sub(locked), "preflight", preflightTook)
			c.saveLastGood(cfg)
			c.mu.Lock()
			c.configErr, c.watchErr = "", ""
			c.mu.Unlock()
			// A removed pool orphans its queue; say so now, not next tick.
			c.passStep(ctx, cfg, "report_orphans", c.reportOrphans)
			return nil
		}
	}
	c.log.ErrorContext(ctx, "config reload failed, keeping old config", "err", err, "by", by, "files", files)
	if by == "watch" {
		c.mu.Lock()
		c.watchErr, c.watchErrAt = err.Error(), c.now()
		c.mu.Unlock()
	}
	return err
}

// saveLastGood records cfg for a later start whose live packs fail (#227).
// It runs after every good load, so it also logs the load's warnings.
func (c *Controller) saveLastGood(cfg *config.Config) {
	for _, w := range cfg.Warnings {
		c.log.Warn("config: " + w)
	}
	if err := config.SaveLastGood(c.root.LastGood(), cfg); err != nil {
		c.log.Warn("config: save last good", "err", err)
	}
}

func (c *Controller) reloadConfig(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "config-reload", "config reload"); err != nil {
		return nil, err
	}
	if err := c.reloadWith(ctx, "", nil); err != nil {
		return nil, err
	}
	cfg := c.config()
	return rpc.ReloadResult{Pools: len(cfg.Pools), Agents: len(cfg.Agents), Triggers: len(cfg.Triggers), Workflows: len(cfg.Workflows)}, nil
}

// configShow returns the active config, not a fresh read of the file, so
// it shows what the controller runs with after the last good reload.
func (c *Controller) configShow(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "config-show", "config show"); err != nil {
		return nil, err
	}
	cfg := c.config()
	// Round trip through YAML so keys match jobshop.yaml, not Go names.
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	out := rpc.ConfigShowResult{Config: map[string]any{}, Sources: cfg.Sources, Instances: cfg.Instances, Disabled: cfg.Disabled}
	if cfg.Reload.Watch {
		out.Watch = &rpc.ConfigWatch{On: true, Files: len(cfg.Watch.Files), Dirs: len(cfg.Watch.Dirs)}
	}
	if err := yaml.Unmarshal(b, &out.Config); err != nil {
		return nil, err
	}
	// The resolved kind patterns, defaults included (#524).
	out.Config["learnings"] = map[string]any{"kinds": cfg.LearningKinds()}
	if out.Sources == nil {
		out.Sources = map[string][]string{}
	}
	return out, nil
}
