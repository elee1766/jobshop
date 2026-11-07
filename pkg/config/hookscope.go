package config

import (
	"fmt"
	"slices"
	"strings"
)

// ResolveHookReference keeps dynamic references inside the declaring instance.
func (c *Config) ResolveHookReference(owner, kind, name string) (string, error) {
	var own *PackInstance
	section, ownerName, ok := strings.Cut(owner, ".")
	if !ok {
		return "", fmt.Errorf("hook owner must name its configuration section")
	}
	for i := range c.Instances {
		in := &c.Instances[i]
		if section == "workflows" && slices.Contains(in.Workflows, ownerName) || section == "triggers" && slices.Contains(in.Triggers, ownerName) {
			own = in
			break
		}
	}
	candidates := func(in PackInstance) []string {
		if kind == "workflow" {
			return in.Workflows
		}
		return in.Pools
	}
	if own != nil && own.Prefix != "" && slices.Contains(candidates(*own), own.Prefix+name) {
		name = own.Prefix + name
	}
	if own != nil {
		for _, in := range c.Instances {
			if slices.Contains(candidates(in), name) && !slices.Contains(candidates(*own), name) {
				return "", fmt.Errorf("%s %q is outside %s's pack instance", kind, name, owner)
			}
		}
	}
	if kind == "workflow" {
		if _, ok := c.Workflows[name]; !ok {
			return "", fmt.Errorf("unknown workflow %q", name)
		}
	} else if _, ok := c.Pools[name]; !ok {
		return "", fmt.Errorf("unknown pool %q", name)
	}
	return name, nil
}
