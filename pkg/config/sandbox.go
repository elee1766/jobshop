package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Sandbox backends (#262).
const (
	SandboxNone       = "none"
	SandboxBubblewrap = "bubblewrap"
	// NetworkPasta gives the agent a private network namespace through
	// passt's pasta, filtered by Deny (#388).
	NetworkPasta = "pasta"
)

// Sandbox configures how a session's agent is confined (#262). It is set
// under defaults, on agents and on pools. Backend and Network: the pool's
// wins over the agent's over the default. RO, RW and Hide are appended:
// the default's, then the agent's, then the pool's. Paths may start with
// "~/" for the controller's HOME.
type Sandbox struct {
	// Backend is "none" or "bubblewrap". Empty inherits.
	Backend string `yaml:"backend"`
	// RO lists extra read-only binds; missing ones are skipped.
	RO []string `yaml:"ro"`
	// RW lists extra read-write binds, created if missing.
	RW []string `yaml:"rw"`
	// Hide lists paths covered by an empty tmpfs.
	Hide []string `yaml:"hide"`
	// Network is "host", "none" or "pasta". Empty inherits; default host.
	Network string `yaml:"network"`
	// Deny lists CIDRs the agent cannot reach; needs network pasta.
	// Appended like RO.
	Deny []string `yaml:"deny"`
}

// Defaults are host-wide settings agents and pools inherit.
type Defaults struct {
	Sandbox   *Sandbox   `yaml:"sandbox"`
	Isolation *Isolation `yaml:"isolation"`
}

func (s *Sandbox) check() error {
	if s == nil {
		return nil
	}
	switch s.Backend {
	case "", SandboxNone, SandboxBubblewrap:
	default:
		return fmt.Errorf("sandbox backend %q must be %s or %s", s.Backend, SandboxNone, SandboxBubblewrap)
	}
	switch s.Network {
	case "", "host", "none", NetworkPasta:
	default:
		return fmt.Errorf("sandbox network %q must be host, none or %s", s.Network, NetworkPasta)
	}
	if err := checkDeny(s.Network, s.Deny); err != nil {
		return err
	}
	for _, l := range [][]string{s.RO, s.RW, s.Hide} {
		for _, p := range l {
			if !strings.HasPrefix(p, "~/") && !filepath.IsAbs(p) {
				return fmt.Errorf("sandbox path %q must be absolute or start with ~/", p)
			}
		}
	}
	return nil
}

// SandboxFor merges the default, agent and pool sandbox settings of pool.
// An empty Backend means the session is not sandboxed.
func (c *Config) SandboxFor(pool string) Sandbox {
	p := c.Pools[pool]
	a := c.Agents[p.Agent]
	var out Sandbox
	for _, s := range []*Sandbox{c.Defaults.Sandbox, a.Sandbox, p.Sandbox} {
		if s == nil {
			continue
		}
		if s.Backend != "" {
			out.Backend = s.Backend
		}
		if s.Network != "" {
			out.Network = s.Network
		}
		out.RO = append(out.RO, s.RO...)
		out.RW = append(out.RW, s.RW...)
		out.Hide = append(out.Hide, s.Hide...)
		out.Deny = append(out.Deny, s.Deny...)
	}
	if out.Backend == SandboxNone {
		out.Backend = ""
	}
	if out.Network != NetworkPasta {
		// Deny set on a pasta default and a host pool has nothing to filter.
		out.Deny = nil
	}
	return out
}
