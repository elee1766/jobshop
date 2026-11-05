package control

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
// supervisor tick after reload returns, prompts and turn timeouts from
// the next turn, agent command, args and env from the next spawned host (adopted hosts go
// through the env refresh path). Running sessions are not stopped.
// Telemetry settings need a restart.
func (c *Controller) reload() error { return c.reloadWith("", nil) }

// reloadWith is reload tagged with its cause: "watch" for the config
// watcher (#533), which also names the changed files. A failed watch
// reload raises a config-invalid attention item; any good reload clears
// it.
func (c *Controller) reloadWith(by string, files []string) (err error) {
	defer func() {
		if err == nil {
			// Non-blocking: the watcher re-syncs its paths once per burst.
			select {
			case c.configSwapped <- struct{}{}:
			default:
			}
		}
	}()
	c.reloadMu.Lock()
	defer c.reloadMu.Unlock()
	// Swap between passes: a pass that read the old pools must finish
	// before reload returns (#102).
	c.tickMu.Lock()
	defer c.tickMu.Unlock()
	cfg, err := config.Load(c.root)
	if err == nil {
		var lc *liveConfig
		if lc, err = newLiveConfig(cfg); err == nil {
			// Caps first: a failed write keeps the old config and caps.
			err = c.store.SetResources(context.Background(), store.ResourceConfig{Caps: cfg.ResourceCaps(), Providers: cfg.ResourceProviders()})
		}
		if err == nil {
			old := c.live.Swap(lc).cfg
			if !reflect.DeepEqual(old.Telemetry, cfg.Telemetry) {
				c.log.Warn("config reload: telemetry changes need a controller restart")
			}
			if !reflect.DeepEqual(old.Identities, cfg.Identities) {
				// A rotated key or moved installation must not serve old tokens.
				c.minter.Forget()
			}
			// The fix for a permanent spawn error is a config change.
			if cleared := c.unquarantinePools(""); len(cleared) > 0 {
				c.log.Info("config reload: pools unquarantined", "pools", cleared)
			}
			msg := "config reloaded"
			if by != "" {
				msg += " (" + by + ")"
			}
			c.log.Info(msg, "files", files, "pools", len(cfg.Pools), "agents", len(cfg.Agents), "triggers", len(cfg.Triggers), "workflows", len(cfg.Workflows))
			c.saveLastGood(cfg)
			c.mu.Lock()
			c.configErr, c.watchErr = "", ""
			c.mu.Unlock()
			// A removed pool orphans its queue; say so now, not next tick.
			c.reportOrphans()
			return nil
		}
	}
	c.log.Error("config reload failed, keeping old config", "err", err, "by", by, "files", files)
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

func (c *Controller) reloadConfig(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "config-reload", "config reload"); err != nil {
		return nil, err
	}
	if err := c.reload(); err != nil {
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
	out := rpc.ConfigShowResult{Config: map[string]any{}, Sources: cfg.Sources, Instances: cfg.Instances}
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
