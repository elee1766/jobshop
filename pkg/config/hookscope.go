package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/store"
)

// HookScope records every pool and workflow reference owner's callbacks
// can resolve now, as a run pins it at start (see store.RunScope).
func (c *Config) HookScope(owner string) *store.RunScope {
	scope := &store.RunScope{Pools: map[string]store.ScopeRef{}, Workflows: map[string]store.ScopeRef{}}
	var prefixes []string
	for _, in := range c.Instances {
		if in.Prefix != "" {
			prefixes = append(prefixes, in.Prefix)
		}
	}
	add := func(kind, name, project string, into map[string]store.ScopeRef) {
		inputs := []string{name}
		for _, prefix := range prefixes {
			if local, ok := strings.CutPrefix(name, prefix); ok && local != "" {
				inputs = append(inputs, local)
			}
		}
		for _, input := range inputs {
			if resolved, err := c.ResolveHookReference(owner, kind, input); err == nil && resolved == name {
				into[input] = store.ScopeRef{Name: name, Project: project}
			}
		}
	}
	for name, pool := range c.Pools {
		add("pool", name, pool.Project, scope.Pools)
	}
	for name, w := range c.Workflows {
		add("workflow", name, w.Project, scope.Workflows)
	}
	return scope
}

// ResolvePinnedHookReference resolves name now and requires the run's
// pinned scope to have accepted the same target in the same project.
func (c *Config) ResolvePinnedHookReference(owner, kind, name string, scope *store.RunScope) (string, error) {
	resolved, err := c.ResolveHookReference(owner, kind, name)
	if err != nil || scope == nil {
		return resolved, err
	}
	pinned, project := scope.Pools, c.Pools[resolved].Project
	if kind == "workflow" {
		pinned, project = scope.Workflows, c.Workflows[resolved].Project
	}
	if ref, ok := pinned[name]; !ok || ref != (store.ScopeRef{Name: resolved, Project: project}) {
		return "", fmt.Errorf("%s %q is outside the scope %s accepted when the run started", kind, name, owner)
	}
	return resolved, nil
}

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
