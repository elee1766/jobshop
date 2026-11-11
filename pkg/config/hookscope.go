package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/pack"
)

// HookScope records every pool and workflow reference owner's callbacks
// can resolve now, as a run pins it at start (see store.RunScope).
func (c *Config) HookScope(owner string) *domain.RunScope {
	scope := &domain.RunScope{Pools: map[string]domain.ScopeRef{}, Workflows: map[string]domain.ScopeRef{}}
	var prefixes []string
	for _, in := range c.Instances {
		if in.Prefix != "" {
			prefixes = append(prefixes, in.Prefix)
		}
	}
	add := func(kind, name, project string, into map[string]domain.ScopeRef) {
		inputs := []string{name}
		for _, prefix := range prefixes {
			if local, ok := strings.CutPrefix(name, prefix); ok && local != "" {
				inputs = append(inputs, local)
			}
		}
		for _, input := range inputs {
			if resolved, err := c.ResolveHookReference(owner, kind, input); err == nil && resolved == name {
				into[input] = domain.ScopeRef{Name: name, Project: project}
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
func (c *Config) ResolvePinnedHookReference(owner, kind, name string, scope *domain.RunScope) (string, error) {
	resolved, err := c.ResolveHookReference(owner, kind, name)
	if err != nil || scope == nil {
		return resolved, err
	}
	pinned, project := scope.Pools, c.Pools[resolved].Project
	if kind == "workflow" {
		pinned, project = scope.Workflows, c.Workflows[resolved].Project
	}
	if ref, ok := pinned[name]; !ok || ref != (domain.ScopeRef{Name: resolved, Project: project}) {
		return "", fmt.Errorf("%s %q is outside the scope %s accepted when the run started", kind, name, owner)
	}
	return resolved, nil
}

// hookInstance is the pack instance that declares owner ("workflows.x" or
// "triggers.x"), or nil.
func (c *Config) hookInstance(owner string) *pack.Instance {
	section, ownerName, _ := strings.Cut(owner, ".")
	for i := range c.Instances {
		in := &c.Instances[i]
		if section == "workflows" && slices.Contains(in.Workflows, ownerName) || section == "triggers" && slices.Contains(in.Triggers, ownerName) {
			return in
		}
	}
	return nil
}

func hookCandidates(in pack.Instance, kind string) []string {
	if kind == "workflow" {
		return in.Workflows
	}
	return in.Pools
}

// HookReadName maps a pool or workflow name a hook reads with tickets(),
// runs() or sessions() to its own instance's prefixed name when that
// exists, the way ResolveHookReference does for actions (#1221). Any
// other name is returned unchanged: reads see the whole store.
func (c *Config) HookReadName(owner, kind, name string) string {
	own := c.hookInstance(owner)
	if name != "" && own != nil && own.Prefix != "" && slices.Contains(hookCandidates(*own, kind), own.Prefix+name) {
		return own.Prefix + name
	}
	return name
}

// ResolveHookReference keeps dynamic references inside the declaring instance.
func (c *Config) ResolveHookReference(owner, kind, name string) (string, error) {
	if !strings.Contains(owner, ".") {
		return "", fmt.Errorf("hook owner must name its configuration section")
	}
	own := c.hookInstance(owner)
	candidates := func(in pack.Instance) []string { return hookCandidates(in, kind) }
	name = c.HookReadName(owner, kind, name)
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
