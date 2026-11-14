package pack

import (
	"fmt"
	"slices"
	"strings"
)

// prefixLayers renames the pools, agents, triggers and workflows the
// layers declare to prefix+name, and rewrites the references between
// them (pool agent, step pool, escalate_pool, trigger when.pool,
// ticket.pool and run.workflow). Names the pack does not declare (host
// pools and agents) are left alone. This lets one pack, such as
// packs/github, be included once per repo (#184).
//
// pools, when not nil, maps pool names to another instance's pools
// (pools_from, JOB-207): the layers' pool entries and every reference
// to a name in it take the mapped name instead of prefix+name.
func prefixLayers(layers []layer, prefix string, pools map[string]string) {
	rename := map[string]map[string]string{}
	for _, s := range mergedSections {
		rename[s] = map[string]string{}
	}
	for _, l := range layers {
		for _, s := range mergedSections {
			sm, _ := l.m[s].(map[string]any)
			for n := range sm {
				rename[s][n] = prefix + n
			}
		}
	}
	if pools != nil {
		rename["pools"] = pools
	}
	ren := func(m map[string]any, key, section string) {
		if v, ok := m[key].(string); ok {
			if to, ok := rename[section][v]; ok {
				m[key] = to
			}
		}
	}
	renList := func(list []any, section string) {
		for i, v := range list {
			if s, ok := v.(string); ok {
				if to, ok := rename[section][s]; ok {
					list[i] = to
				}
			}
		}
	}
	for _, l := range layers {
		for _, s := range mergedSections {
			sm, ok := l.m[s].(map[string]any)
			if !ok {
				continue
			}
			out := make(map[string]any, len(sm))
			for n, e := range sm {
				em, _ := e.(map[string]any)
				if s == "pools" || s == "agents" {
					for key, section := range map[string]string{"sandbox": "sandbox_profiles", "isolation": "isolation_profiles"} {
						if sb, ok := em[key].(map[string]any); ok {
							list, _ := sb["profiles"].([]any)
							renList(list, section)
						}
					}
				}
				switch {
				case em == nil:
				case s == "pools":
					ren(em, "agent", "agents")
				case s == "workflows":
					ren(em, "escalate_pool", "pools")
					if e, ok := em["escalate"].(map[string]any); ok {
						ren(e, "chat_pool", "pools")
					}
					steps, _ := em["steps"].([]any)
					for _, st := range steps {
						if sm, ok := st.(map[string]any); ok {
							ren(sm, "pool", "pools")
							ren(sm, "retry_pool", "pools")
						}
					}
				case s == "triggers":
					for _, k := range []string{"when", "ticket"} {
						if m, ok := em[k].(map[string]any); ok {
							ren(m, "pool", "pools")
						}
					}
					if m, ok := em["run"].(map[string]any); ok {
						ren(m, "workflow", "workflows")
					}
					if w, ok := em["when"].(map[string]any); ok {
						sigs, _ := w["scan"].([]any)
						for _, sig := range sigs {
							if sm, ok := sig.(map[string]any); ok {
								ren(sm, "start", "workflows")
							}
						}
					}
				}
				to, ok := rename[s][n]
				if !ok {
					to = prefix + n
				}
				out[to] = e
			}
			l.m[s] = out
		}
		// Prompt rules are a list: prefix their names (unique across
		// files) and the pack's own pools, agents and workflows they match.
		rules, _ := l.m["prompt_rules"].([]any)
		for _, r := range rules {
			rm, _ := r.(map[string]any)
			if rm == nil {
				continue
			}
			if n, ok := rm["name"].(string); ok && n != "" {
				rm["name"] = prefix + n
			}
			match, _ := rm["match"].(map[string]any)
			for key, section := range map[string]string{"pools": "pools", "agents": "agents", "workflows": "workflows"} {
				list, _ := match[key].([]any)
				renList(list, section)
			}
		}
		// Hooks name the pools they guard; the pack's own get the prefix.
		hooks, _ := l.m["hooks"].([]any)
		for _, h := range hooks {
			hm, _ := h.(map[string]any)
			list, _ := hm["pools"].([]any)
			renList(list, "pools")
		}
	}
}

// renameSources is prefixSources with pool entries renamed by pools
// when it is not nil (pools_from).
func renameSources(src map[string][]string, prefix string, pools map[string]string) map[string][]string {
	out := prefixSources(src, prefix)
	if pools == nil {
		return out
	}
	for k, v := range src {
		if local, ok := cutSection(k, "pools"); ok {
			delete(out, "pools."+prefix+local)
			if to, ok := pools[local]; ok {
				out["pools."+to] = v
			}
		}
	}
	return out
}

func cutSection(key, section string) (string, bool) {
	return strings.CutPrefix(key, section+".")
}

// sharedPools maps the pools of owner (the pools_from entry, whose pool
// names carry prefix) back from their unprefixed names, the names a
// sharing entry of the same pack declares and references.
func sharedPools(owner Instance, prefix string) map[string]string {
	out := map[string]string{}
	for _, p := range owner.Pools {
		if local, ok := strings.CutPrefix(p, prefix); ok && local != "" {
			out[local] = p
		}
	}
	return out
}

// checkPoolsFrom requires the entry a pools_from names not to share pools
// itself, and every pool the sharing entry uses to be one of that
// entry's: a pool only the sharing entry declares would be a
// half-defined pool under the other entry's prefix. index maps entry
// names to instances.
func checkPoolsFrom(instances []Instance, index map[string]int) error {
	for _, in := range instances {
		if in.PoolsFrom == "" {
			continue
		}
		owner := instances[index[in.PoolsFrom]]
		if owner.PoolsFrom != "" {
			return fmt.Errorf("%s: pools_from %s: that entry takes its pools from %s", instanceName(in), in.PoolsFrom, owner.PoolsFrom)
		}
		for _, p := range in.Pools {
			if !slices.Contains(owner.Pools, p) {
				return fmt.Errorf("%s: pool %s is not declared by pools_from %s", instanceName(in), p, in.PoolsFrom)
			}
		}
	}
	return nil
}
