package config

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// maxLabelLen bounds one pool label.
const maxLabelLen = 64

// checkLabels rejects labels that cannot be a filter key or value: empty,
// padded, or containing whitespace or a comma (the CLI and Desk filters
// split on it).
func checkLabels(labels []string) error {
	for _, l := range labels {
		switch {
		case l == "":
			return fmt.Errorf("label must not be empty")
		case len(l) > maxLabelLen:
			return fmt.Errorf("label %q is longer than %d", l, maxLabelLen)
		case strings.ContainsAny(l, " \t\r\n,"):
			return fmt.Errorf("label %q must not contain whitespace or a comma", l)
		}
	}
	return nil
}

// InstanceLabels are the labels a packs entry gives every pool it
// declares: instance:<name> from the entry's prefix (or the pack's name
// without one) and pack:<pack> (JOB-81).
func InstanceLabels(in PackInstance) []string {
	pack := filepath.Base(in.Path)
	inst := strings.Trim(in.Prefix, "-_")
	if inst == "" {
		inst = pack
	}
	return []string{"instance:" + inst, "pack:" + pack}
}

// addInstanceLabels adds each packs entry's labels to its pools, after the
// pack and host files are merged, and sorts and dedupes every pool's
// labels so the RPC order is stable.
func (c *Config) addInstanceLabels() {
	for _, in := range c.Instances {
		add := InstanceLabels(in)
		for _, name := range in.Pools {
			p, ok := c.Pools[name]
			if !ok {
				continue // disabled or dropped
			}
			p.Labels = append(slices.Clone(p.Labels), add...)
			c.Pools[name] = p
		}
	}
	for name, p := range c.Pools {
		if len(p.Labels) == 0 {
			continue
		}
		labels := slices.Clone(p.Labels)
		slices.Sort(labels)
		p.Labels = slices.Compact(labels)
		c.Pools[name] = p
	}
}

// checkRawPoolLabels rejects a pool label that is not a YAML string. The
// typed decode would accept 1 as "1"; a label is a name, so a number is
// most likely a mistake.
func checkRawPoolLabels(b []byte, src func(section, name string) string) error {
	top, err := decodeMap(b)
	if err != nil {
		return nil // the typed decode reports the syntax error
	}
	pools, _ := top["pools"].(map[string]any)
	for name, e := range pools {
		em, _ := e.(map[string]any)
		raw, ok := em["labels"]
		if !ok || raw == nil {
			continue
		}
		list, ok := raw.([]any)
		if !ok {
			return fmt.Errorf("%s: pool %s: labels must be a list of strings", src("pools", name), name)
		}
		for _, l := range list {
			if _, ok := l.(string); !ok {
				return fmt.Errorf("%s: pool %s: labels must be strings, got %v", src("pools", name), name, l)
			}
		}
	}
	return nil
}
