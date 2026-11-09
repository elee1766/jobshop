package pack

// prefixLayers renames the pools, agents, triggers and workflows the
// layers declare to prefix+name, and rewrites the references between
// them (pool agent, step pool, escalate_pool, trigger when.pool,
// ticket.pool and run.workflow). Names the pack does not declare (host
// pools and agents) are left alone. This lets one pack, such as
// packs/github, be included once per repo (#184).
func prefixLayers(layers []layer, prefix string) {
	declared := map[string]map[string]bool{}
	for _, s := range mergedSections {
		declared[s] = map[string]bool{}
	}
	for _, l := range layers {
		for _, s := range mergedSections {
			sm, _ := l.m[s].(map[string]any)
			for n := range sm {
				declared[s][n] = true
			}
		}
	}
	ren := func(m map[string]any, key, section string) {
		if v, ok := m[key].(string); ok && declared[section][v] {
			m[key] = prefix + v
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
							for i, v := range list {
								if p, ok := v.(string); ok && declared[section][p] {
									list[i] = prefix + p
								}
							}
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
				}
				out[prefix+n] = e
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
				for i, v := range list {
					if s, ok := v.(string); ok && declared[section][s] {
						list[i] = prefix + s
					}
				}
			}
		}
		// Hooks name the pools they guard; the pack's own get the prefix.
		hooks, _ := l.m["hooks"].([]any)
		for _, h := range hooks {
			hm, _ := h.(map[string]any)
			list, _ := hm["pools"].([]any)
			for i, v := range list {
				if s, ok := v.(string); ok && declared["pools"][s] {
					list[i] = prefix + s
				}
			}
		}
	}
}
