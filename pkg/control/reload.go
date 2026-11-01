package control

import (
	"context"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
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

// Reload re-reads jobshop.yaml and its packs and swaps them in. On any
// error the old config stays active. Pool limits apply from the next
// supervisor tick, prompts and turn timeouts from the next turn, agent
// command, args and env from the next spawned host (adopted hosts go
// through the env refresh path). Running sessions are not stopped.
// Telemetry settings need a restart.
func (c *Controller) reload() error {
	c.reloadMu.Lock()
	defer c.reloadMu.Unlock()
	cfg, err := config.Load(c.root)
	if err == nil {
		var lc *liveConfig
		if lc, err = newLiveConfig(cfg); err == nil {
			old := c.live.Swap(lc).cfg
			if old.Telemetry != cfg.Telemetry {
				c.log.Warn("config reload: telemetry changes need a controller restart")
			}
			c.log.Info("config reloaded", "pools", len(cfg.Pools), "agents", len(cfg.Agents), "triggers", len(cfg.Triggers))
			return nil
		}
	}
	c.log.Error("config reload failed, keeping old config", "err", err)
	return err
}

func (c *Controller) reloadConfig(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := requireAdmin(sc, "config reload"); err != nil {
		return nil, err
	}
	if err := c.reload(); err != nil {
		return nil, err
	}
	cfg := c.config()
	return rpc.ReloadResult{Pools: len(cfg.Pools), Agents: len(cfg.Agents), Triggers: len(cfg.Triggers)}, nil
}
