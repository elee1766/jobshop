package config

import (
	"fmt"
	"net/netip"
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
	// DNS is pasta's DNS forwarding (JOB-20, acphost#124): queries the agent
	// sends to Forward inside the namespace go to Host on the host side.
	// Needs network pasta. Nil inherits; unset keeps pasta's default.
	DNS *SandboxDNS `yaml:"dns"`
	// ResolvConf is a host file bound read-only over /etc/resolv.conf in
	// the sandbox, usually naming DNS.Forward. Empty inherits.
	ResolvConf string `yaml:"resolv_conf"`
	// DNSCheck is a host name a pasta sandbox must resolve before the
	// pool's first session spawns; on failure the pool is quarantined.
	// Empty inherits; unset skips the check.
	DNSCheck string `yaml:"dns_check"`
}

// SandboxDNS is passed to pasta as --dns-forward and --dns-host. Both are
// IP addresses; jobshop picks no defaults (host config sets them).
type SandboxDNS struct {
	Forward string `yaml:"forward"`
	Host    string `yaml:"host"`
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
	if err := checkDNS(s); err != nil {
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

// checkDNS validates the pasta DNS settings (JOB-20).
func checkDNS(s *Sandbox) error {
	if s.DNS != nil {
		if s.Network != "" && s.Network != NetworkPasta {
			return fmt.Errorf("sandbox dns needs network %s", NetworkPasta)
		}
		if _, err := netip.ParseAddr(s.DNS.Forward); err != nil {
			return fmt.Errorf("sandbox dns.forward %q must be an IP address", s.DNS.Forward)
		}
		if s.DNS.Host != "" {
			if _, err := netip.ParseAddr(s.DNS.Host); err != nil {
				return fmt.Errorf("sandbox dns.host %q must be an IP address", s.DNS.Host)
			}
		}
	}
	if p := s.ResolvConf; p != "" && !strings.HasPrefix(p, "~/") && !filepath.IsAbs(p) {
		return fmt.Errorf("sandbox resolv_conf %q must be absolute or start with ~/", p)
	}
	if s.DNSCheck != "" && strings.ContainsAny(s.DNSCheck, " /") {
		return fmt.Errorf("sandbox dns_check %q must be a host name", s.DNSCheck)
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
		if s.DNS != nil {
			out.DNS = s.DNS
		}
		if s.ResolvConf != "" {
			out.ResolvConf = s.ResolvConf
		}
		if s.DNSCheck != "" {
			out.DNSCheck = s.DNSCheck
		}
	}
	if out.Backend == SandboxNone {
		out.Backend = ""
	}
	if out.Network != NetworkPasta {
		// Deny set on a pasta default and a host pool has nothing to filter,
		// and pasta's DNS settings have no pasta to go to.
		out.Deny = nil
		out.DNS = nil
		out.DNSCheck = ""
	}
	return out
}
