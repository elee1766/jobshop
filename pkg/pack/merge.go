package pack

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/pack/tmpl"
)

// mergedSections are the top-level keys packs contribute to. Their
// entries merge per key, and inside an entry per field.
var mergedSections = []string{"agents", "pools", "triggers", "workflows", "desk_commands", "sandbox_profiles", "isolation_profiles", "projects", "resources"}

func isMergedSection(k string) bool {
	return slices.Contains(mergedSections, k)
}

// DecodeMap decodes a YAML mapping. An empty document is an empty map.
func DecodeMap(b []byte) (map[string]any, error) {
	m := map[string]any{}
	if len(bytes.TrimSpace(b)) == 0 {
		return m, nil
	}
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// cloneValue deep-copies maps and lists so substitution does not edit
// the caller's config.
func cloneValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = cloneValue(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = cloneValue(e)
		}
		return out
	}
	return v
}

// substitute renders Go templates in string values. A string that is
// exactly one reference takes the var's value with its type, so numbers
// stay numbers.
func substitute(v any, vars map[string]any, key string) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			r, err := substitute(e, vars, joinKey(key, k))
			if err != nil {
				return nil, err
			}
			x[k] = r
		}
		return x, nil
	case []any:
		for i, e := range x {
			r, err := substitute(e, vars, fmt.Sprintf("%s[%d]", key, i))
			if err != nil {
				return nil, err
			}
			x[i] = r
		}
		return x, nil
	case string:
		r, err := tmpl.Value(x, vars)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		return r, nil
	}
	return v, nil
}

func joinKey(a, b string) string {
	if a == "" {
		return b
	}
	return a + "." + b
}

func mergeLayer(dst map[string]any, l layer, sources map[string][]string) {
	declared := sources
	if len(l.entrySources) > 0 {
		declared = map[string][]string{}
	}
	mergeInto(dst, l.m, l.file, declared)
	if len(l.entrySources) > 0 {
		for key, files := range declared {
			if cueFiles := l.entrySources[key]; len(cueFiles) > 0 {
				files = cueFiles
			}
			sources[key] = append(sources[key], files...)
		}
	}
	for key, files := range l.textSources {
		sources[key] = files
	}
}

// mergeInto applies src over dst. In the merged sections each entry merges
// recursively: maps merge per key, other values (lists, scalars) replace.
// Other top-level keys replace. sources records which files set each
// section entry, as "section.name".
func mergeInto(dst, src map[string]any, file string, sources map[string][]string) {
	for k, v := range src {
		if k == "hooks" {
			// Appended like prompt_rules: a pack adds guardrails and cannot
			// drop the host's or another pack's (#249).
			hooks, _ := v.([]any)
			old, _ := dst[k].([]any)
			dst[k] = append(slices.Clone(old), hooks...)
			continue
		}
		if k == "learnings" {
			// learnings.kinds is appended like prompt_rules: a pack adds
			// the kinds it records (#524). Other keys replace.
			sm, _ := v.(map[string]any)
			dm, _ := dst[k].(map[string]any)
			out := maps.Clone(dm)
			if out == nil {
				out = map[string]any{}
			}
			for sk, sv := range sm {
				if sk == "kinds" {
					add, _ := sv.([]any)
					old, _ := out[sk].([]any)
					out[sk] = append(slices.Clone(old), add...)
					continue
				}
				out[sk] = sv
			}
			dst[k] = out
			continue
		}
		if k == "prompt_rules" {
			// Appended, not replaced: a later file adds rules and cannot
			// drop earlier ones. Duplicate names fail validation.
			rules, _ := v.([]any)
			old, _ := dst[k].([]any)
			dst[k] = append(slices.Clone(old), rules...)
			for _, r := range rules {
				if rm, ok := r.(map[string]any); ok {
					if n, ok := rm["name"].(string); ok && n != "" {
						key := "prompt_rules." + n
						sources[key] = append(sources[key], file)
					}
				}
			}
			continue
		}
		if !isMergedSection(k) {
			dst[k] = v
			if k == "default_project" {
				sources[k+"."] = append(sources[k+"."], file)
			}
			continue
		}
		sm, ok := v.(map[string]any)
		if !ok {
			dst[k] = v
			continue
		}
		dm, _ := dst[k].(map[string]any)
		if dm == nil {
			dm = map[string]any{}
			dst[k] = dm
		}
		for name, entry := range sm {
			merged := MergeValue(dm[name], entry)
			if k == "pools" || k == "agents" {
				unionProfiles(merged, dm[name], entry)
			}
			if k == "pools" {
				unionLabels(merged, dm[name], entry)
			}
			dm[name] = merged
			key := k + "." + name
			sources[key] = append(sources[key], file)
		}
	}
}

// unionLabels sets out's labels to the union of old's and v's, so a host
// adds to a pack pool's labels instead of replacing them (JOB-81). The
// other fields merge per field in mergeValue.
func unionLabels(out, old, v any) {
	om, ok := out.(map[string]any)
	if !ok {
		return
	}
	var all []any
	seen := map[string]bool{}
	for _, e := range []any{old, v} {
		em, _ := e.(map[string]any)
		list, _ := em["labels"].([]any)
		for _, l := range list {
			// A non-string stays for the typed decode to reject.
			if s, ok := l.(string); ok {
				if seen[s] {
					continue
				}
				seen[s] = true
			}
			all = append(all, l)
		}
	}
	if len(all) > 0 {
		om["labels"] = all
	}
}

// unionProfiles sets the sandbox and isolation profiles of out (a merged
// pool or agent entry) to the union of old's and v's, so a host adds
// profiles to a pack pool instead of replacing the pack's (JOB-83). The
// other fields merge per field in mergeValue.
func unionProfiles(out, old, v any) {
	om, ok := out.(map[string]any)
	if !ok {
		return
	}
	for _, key := range []string{"sandbox", "isolation"} {
		var all []any
		seen := map[string]bool{}
		for _, e := range []any{old, v} {
			em, _ := e.(map[string]any)
			lm, _ := em[key].(map[string]any)
			list, _ := lm["profiles"].([]any)
			for _, p := range list {
				// A non-string stays for the typed decode to reject.
				if s, ok := p.(string); ok {
					if seen[s] {
						continue
					}
					seen[s] = true
				}
				all = append(all, p)
			}
		}
		if len(all) == 0 {
			continue
		}
		if sm, ok := om[key].(map[string]any); ok {
			sm["profiles"] = all
		}
	}
}

// MergeValue merges two decoded YAML values: maps merge per key and
// recursively, anything else is replaced by v. The service file uses it for
// its includes, the same rule packs use for an entry.
func MergeValue(old, v any) any {
	om, ok1 := old.(map[string]any)
	nm, ok2 := v.(map[string]any)
	if !ok1 || !ok2 {
		return v
	}
	out := make(map[string]any, len(om)+len(nm))
	for k, e := range om {
		out[k] = e
	}
	for k, e := range nm {
		out[k] = MergeValue(om[k], e)
	}
	return out
}
