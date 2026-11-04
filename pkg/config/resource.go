package config

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"
)

// Resource is a named slot pool steps hold while in progress (#400). A
// step lists resource keys in needs_resources; it starts only when each
// key has a free slot, and the slot frees when its claim ends.
type Resource struct {
	// Max is how many steps may hold the resource at once.
	Max int `yaml:"max"`
	// Keyed makes slots per key: a step needs "name:<key>" (for example
	// "pr:${repo}#${pr}"), and each key has CapacityPerKey slots.
	Keyed bool `yaml:"keyed"`
	// CapacityPerKey is the slots per key of a keyed resource. Default 1.
	CapacityPerKey int `yaml:"capacity_per_key"`
	// Provider owns concrete slots (#407); the slots are the capacity, so
	// max and keyed are not set with it.
	Provider *Provider `yaml:"provider"`
}

// Capacity is the slots per key of a keyed resource, or the total.
func (r Resource) Capacity() int {
	if r.Keyed {
		if r.CapacityPerKey == 0 {
			return 1
		}
		return r.CapacityPerKey
	}
	return r.Max
}

// ResourceCaps maps each resource name to its slot count, for the store.
func (c *Config) ResourceCaps() map[string]int {
	out := make(map[string]int, len(c.Resources))
	for name, r := range c.Resources {
		if r.Provider != nil {
			continue // its slots gate claims, not a count
		}
		out[name] = r.Capacity()
	}
	return out
}

func (c *Config) checkResources() error {
	for _, name := range slices.Sorted(maps.Keys(c.Resources)) {
		r := c.Resources[name]
		if !namePattern.MatchString(name) {
			return fmt.Errorf("resource %q: name must match %s", name, namePattern)
		}
		if r.Provider != nil {
			if r.Keyed || r.Max != 0 || r.CapacityPerKey != 0 {
				return fmt.Errorf("resource %s: a provider's slots are its capacity: drop max and keyed", name)
			}
			if err := r.Provider.check(); err != nil {
				return fmt.Errorf("resource %s: %w", name, err)
			}
			continue
		}
		switch {
		case r.Keyed && r.Max != 0:
			return fmt.Errorf("resource %s: keyed takes capacity_per_key, not max", name)
		case r.Keyed && r.CapacityPerKey < 0:
			return fmt.Errorf("resource %s: capacity_per_key must be >= 1", name)
		case !r.Keyed && r.CapacityPerKey != 0:
			return fmt.Errorf("resource %s: capacity_per_key needs keyed: true", name)
		case !r.Keyed && r.Max <= 0:
			return fmt.Errorf("resource %s: max must be >= 1", name)
		}
	}
	return nil
}

// CheckResourceKey checks a filled needs_resources entry against the
// configured resources: a keyed resource needs "name:<key>", a counted
// one the bare name.
func (c *Config) CheckResourceKey(key string) error {
	name, sub, keyed := strings.Cut(key, ":")
	r, ok := c.Resources[name]
	switch {
	case !ok:
		return fmt.Errorf("unknown resource %q", name)
	case r.Keyed && (!keyed || sub == ""):
		return fmt.Errorf("resource %s is keyed: use %s:<key>", name, name)
	case !r.Keyed && keyed:
		return fmt.Errorf("resource %s is not keyed: use %s", name, name)
	}
	return nil
}

// ResolveResources fills a step's needs_resources templates from vars
// and checks each key. Duplicates are dropped.
func (c *Config) ResolveResources(w *Workflow, needs []string, vars map[string]any) ([]string, error) {
	var out []string
	for _, n := range needs {
		key, err := w.Render(n, vars)
		if err != nil {
			return nil, fmt.Errorf("needs_resources: %w", err)
		}
		if err := c.CheckResourceKey(key); err != nil {
			return nil, fmt.Errorf("needs_resources: %w", err)
		}
		if !slices.Contains(out, key) {
			out = append(out, key)
		}
	}
	return out, nil
}

// checkStepResources checks needs_resources at load: every var is known,
// and the resource name (the text before ':') is configured.
func (c *Config) checkStepResources(w *Workflow, needs []string) error {
	for _, n := range needs {
		if n == "" {
			return errors.New("needs_resources: empty entry")
		}
		vs, err := refs(n, w.Template)
		if err != nil {
			return fmt.Errorf("needs_resources: %w", err)
		}
		for _, v := range vs {
			if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) {
				return fmt.Errorf("needs_resources uses %s, not in workflow vars or pack vars", placeholder(v, w.Template))
			}
		}
		name, _, _ := strings.Cut(n, ":")
		if _, ok := c.Resources[name]; !ok {
			return fmt.Errorf("needs_resources: unknown resource %q", name)
		}
	}
	return nil
}

// Provider owns a resource's slots (#407): it adds and removes them
// (`jobshop resource add|remove|unhealthy`), and the controller calls it
// to reset a slot after each claim, to report demand, and to remove idle
// slots. Set exactly one of Command and URL.
type Provider struct {
	// Command is run as `sh -c '<command> "$@"' <op> [id]`, with the call
	// as JSON on stdin. Exit 0 is ok.
	Command string `yaml:"command"`
	// URL gets the call as a JSON POST. A 2xx status is ok.
	URL string `yaml:"url"`
	// Timeout bounds one call. Default 2m.
	Timeout string `yaml:"timeout"`
	// IdleRelease asks the provider to remove a slot that has been ready
	// this long while nothing waits for the resource. Empty: never.
	IdleRelease string `yaml:"idle_release"`
	// DemandEvery is how often demand is sent when it has not changed.
	// Default 1m.
	DemandEvery string `yaml:"demand_every"`
}

// Provider defaults.
const (
	DefaultProviderTimeout = 2 * time.Minute
	DefaultDemandEvery     = time.Minute
)

func durOr(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

// TimeoutDuration is Timeout or its default.
func (p *Provider) TimeoutDuration() time.Duration { return durOr(p.Timeout, DefaultProviderTimeout) }

// DemandEveryDuration is DemandEvery or its default.
func (p *Provider) DemandEveryDuration() time.Duration {
	return durOr(p.DemandEvery, DefaultDemandEvery)
}

// IdleReleaseDuration is IdleRelease, 0 when unset.
func (p *Provider) IdleReleaseDuration() time.Duration { return durOr(p.IdleRelease, 0) }

func (p *Provider) check() error {
	if (p.Command == "") == (p.URL == "") {
		return errors.New("provider: set exactly one of command and url")
	}
	if p.URL != "" {
		u, err := url.Parse(p.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("provider: url %q must be an http or https URL", p.URL)
		}
	}
	for k, v := range map[string]string{"timeout": p.Timeout, "idle_release": p.IdleRelease, "demand_every": p.DemandEvery} {
		if v == "" {
			continue
		}
		if d, err := time.ParseDuration(v); err != nil || d <= 0 {
			return fmt.Errorf("provider: %s %q must be a positive duration", k, v)
		}
	}
	return nil
}

// ResourceProviders lists the resources a provider owns, sorted.
func (c *Config) ResourceProviders() []string {
	var out []string
	for _, name := range slices.Sorted(maps.Keys(c.Resources)) {
		if c.Resources[name].Provider != nil {
			out = append(out, name)
		}
	}
	return out
}
