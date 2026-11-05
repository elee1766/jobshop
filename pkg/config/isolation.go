package config

import (
	"fmt"
	"math"
	"net/netip"
	"os/user"
	"strconv"
	"strings"
)

// Isolation runs a pool's agents as a mapped host user and in a cgroup
// with limits (#388). Set under defaults, on agents and on pools; the
// pool's fields win over the agent's over the default. Optional: with no
// isolation config nothing changes.
type Isolation struct {
	User   *IsolationUser `yaml:"user"`
	Limits *CgroupLimits  `yaml:"limits"`
}

// IsolationUser maps the agent's uid/gid to host ids in the controller
// user's /etc/subuid and /etc/subgid ranges. Inside the namespace the
// agent runs as UID and GID (GID defaults to UID). Groups are host groups
// (names or gids) the controller user holds, kept as supplementary
// groups so worktrees stay shared with the controller.
type IsolationUser struct {
	UID    int      `yaml:"uid"`
	GID    *int     `yaml:"gid"`
	MapUID int      `yaml:"map_uid"`
	MapGID int      `yaml:"map_gid"`
	Groups []string `yaml:"groups"`
	// PerSession gives each live session its own host uid/gid from
	// [map_uid, map_uid+range) and [map_gid, map_gid+range) (#540).
	PerSession *PerSession `yaml:"per_session"`
}

// PerSession is the size of a pool's per-session id range.
type PerSession struct {
	Range int `yaml:"range"`
}

// Slots is how many distinct sessions the user's ids allow: the
// per_session range, or 0 when every session shares map_uid.
func (u IsolationUser) Slots() int {
	if u.PerSession == nil {
		return 0
	}
	return u.PerSession.Range
}

// CgroupLimits are cgroup v2 limits for each session of a pool. CPU is a count
// of CPUs ("4", "0.5"). Memory values take K, M, G or T suffixes (powers
// of 1024) or "max". Zero values are unset.
type CgroupLimits struct {
	CPU        string `yaml:"cpu"`
	Memory     string `yaml:"memory"`
	MemoryHigh string `yaml:"memory_high"`
	// Swap is memory.swap.max. Unset means "0" when memory is set, so an
	// agent at the memory limit is OOM killed instead of swapping.
	Swap string `yaml:"swap"`
	Pids int64  `yaml:"pids"`
}

// GIDOrUID is the inside gid.
func (u IsolationUser) GIDOrUID() int {
	if u.GID != nil {
		return *u.GID
	}
	return u.UID
}

func (i *Isolation) check() error {
	if i == nil {
		return nil
	}
	if u := i.User; u != nil {
		if u.UID < 0 || u.GIDOrUID() < 0 || u.MapUID <= 0 || u.MapGID <= 0 {
			return fmt.Errorf("isolation user: uid and gid must be >= 0, map_uid and map_gid > 0")
		}
		for _, g := range u.Groups {
			if n, err := strconv.Atoi(g); (err == nil && n < 0) || g == "" {
				return fmt.Errorf("isolation user: bad group %q", g)
			}
		}
		if ps := u.PerSession; ps != nil {
			if ps.Range < 1 {
				return fmt.Errorf("isolation user per_session range %d must be at least 1", ps.Range)
			}
			if len(u.Groups) == 0 {
				return fmt.Errorf("isolation user per_session needs groups: the first owns each session's worktree with the controller")
			}
		}
	}
	if l := i.Limits; l != nil {
		if _, err := l.CPUMax(); err != nil {
			return err
		}
		for _, m := range []struct{ k, v string }{{"memory", l.Memory}, {"memory_high", l.MemoryHigh}, {"swap", l.Swap}} {
			if _, err := ParseBytes(m.v); err != nil {
				return fmt.Errorf("isolation limits %s: %w", m.k, err)
			}
		}
		if l.Pids < 0 {
			return fmt.Errorf("isolation limits pids %d is negative", l.Pids)
		}
	}
	return nil
}

// cpuPeriod is the cpu.max period in microseconds.
const cpuPeriod = 100000

// CPUMax is l.CPU as a cgroup cpu.max value, or "" when unset.
func (l CgroupLimits) CPUMax() (string, error) {
	if l.CPU == "" {
		return "", nil
	}
	f, err := strconv.ParseFloat(l.CPU, 64)
	if err != nil || f <= 0 || math.IsInf(f, 0) {
		return "", fmt.Errorf("isolation limits cpu %q: want a positive number of CPUs", l.CPU)
	}
	q := int64(math.Round(f * cpuPeriod))
	if q < 1000 {
		return "", fmt.Errorf("isolation limits cpu %q: below 0.01", l.CPU)
	}
	return fmt.Sprintf("%d %d", q, cpuPeriod), nil
}

// ParseBytes turns "8G", "512M", "1024" or "max" into a cgroup value: a
// byte count or "max". Empty stays empty.
func ParseBytes(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "max" {
		return s, nil
	}
	mult := uint64(1)
	switch strings.ToUpper(s[len(s)-1:]) {
	case "K":
		mult = 1 << 10
	case "M":
		mult = 1 << 20
	case "G":
		mult = 1 << 30
	case "T":
		mult = 1 << 40
	}
	num := s
	if mult > 1 {
		num = s[:len(s)-1]
	}
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil || n > math.MaxUint64/mult {
		return "", fmt.Errorf("%q: want a byte count with optional K, M, G or T, or max", s)
	}
	return strconv.FormatUint(n*mult, 10), nil
}

// IsolationFor merges the default, agent and pool isolation of pool. Each
// of user and limits is taken whole from the most specific level that
// sets it.
func (c *Config) IsolationFor(pool string) Isolation {
	p := c.Pools[pool]
	a := c.Agents[p.Agent]
	var out Isolation
	for _, i := range []*Isolation{c.Defaults.Isolation, a.Isolation, p.Isolation} {
		if i == nil {
			continue
		}
		if i.User != nil {
			out.User = i.User
		}
		if i.Limits != nil {
			out.Limits = i.Limits
		}
	}
	return out
}

// LookupGroup resolves a group name or numeric gid to a gid.
func LookupGroup(g string) (int, error) {
	if n, err := strconv.Atoi(g); err == nil {
		if n < 0 {
			return 0, fmt.Errorf("group %d is negative", n)
		}
		return n, nil
	}
	gr, err := user.LookupGroup(g)
	if err != nil {
		return 0, fmt.Errorf("group %q: %w", g, err)
	}
	return strconv.Atoi(gr.Gid)
}

// checkDeny validates sandbox deny CIDRs.
func checkDeny(network string, deny []string) error {
	if len(deny) > 0 && network != NetworkPasta && network != "" {
		return fmt.Errorf("sandbox deny needs network %s", NetworkPasta)
	}
	for _, d := range deny {
		if _, err := netip.ParsePrefix(d); err != nil {
			return fmt.Errorf("sandbox deny %q: %w", d, err)
		}
	}
	return nil
}
