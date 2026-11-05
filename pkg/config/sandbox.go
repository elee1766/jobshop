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
// wins over the agent's over the default. RO, RW, Hide, Deny and EnvKeep
// are appended: the default's, then the agent's, then the pool's. Paths
// may start with "~/" for the controller's HOME.
//
// Jobshop's code holds no host paths or env names (#625): the system
// mounts an agent needs (/nix/store, /usr/lib, ...) and the env it keeps
// come from config, directly or through named Profiles.
type Sandbox struct {
	// Profiles names sandbox_profiles entries applied before this level's
	// own fields, in order.
	Profiles []string `yaml:"profiles"`
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
	// EnvKeep lists controller env names passed into the sandbox on top
	// of the spec env; a trailing * keeps every name with that prefix.
	// bwrap clears the rest.
	EnvKeep []string `yaml:"env_keep"`
}

// Defaults are host-wide settings agents and pools inherit.
type Defaults struct {
	Sandbox   *Sandbox   `yaml:"sandbox"`
	Isolation *Isolation `yaml:"isolation"`
}

func (s *Sandbox) check(profiles map[string]Sandbox) error {
	if s == nil {
		return nil
	}
	for _, p := range s.Profiles {
		if _, ok := profiles[p]; !ok {
			return fmt.Errorf("sandbox profile %q is not in sandbox_profiles", p)
		}
	}
	for _, k := range s.EnvKeep {
		if k == "" || k == "*" || strings.ContainsAny(strings.TrimSuffix(k, "*"), "=*") {
			return fmt.Errorf("sandbox env_keep %q must be a name or a prefix ending in *", k)
		}
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
	var levels []*Sandbox
	for _, s := range []*Sandbox{c.Defaults.Sandbox, a.Sandbox, p.Sandbox} {
		if s == nil {
			continue
		}
		for _, name := range s.Profiles {
			if pr, ok := c.SandboxProfiles[name]; ok {
				levels = append(levels, &pr)
			}
		}
		levels = append(levels, s)
	}
	var out Sandbox
	for _, s := range levels {
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
		out.EnvKeep = append(out.EnvKeep, s.EnvKeep...)
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
