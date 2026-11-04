package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// textFiles reads every `*_file` prompt-like field of one config file
// into its inline field (#382). Paths are relative to dir, the declaring
// file's directory. In a pack, vars is non-nil and the content gets pack
// var substitution like inline text; host files stay literal.
type textFiles struct {
	dir, file string
	vars      map[string]any
	syn       string // the pack's template syntax
	// sources records "<key>.<field>" -> the file read, for config show.
	sources map[string][]string
}

// load replaces m[fileField] with the file's content as m[field]. key
// names m in errors. subst false keeps the content literal even in a pack.
func (t textFiles) load(m map[string]any, field, fileField, key string, subst bool) error {
	pf, ok := m[fileField]
	if !ok {
		return nil
	}
	if _, both := m[field]; both {
		return fmt.Errorf("%s: %s: set %s or %s, not both", t.file, key, field, fileField)
	}
	s, ok := pf.(string)
	if !ok || s == "" {
		return fmt.Errorf("%s: %s.%s must be a path", t.file, key, fileField)
	}
	if !filepath.IsAbs(s) {
		s = filepath.Join(t.dir, s)
	}
	b, err := os.ReadFile(s)
	if err != nil {
		return fmt.Errorf("%s: %s.%s: %w", t.file, key, fileField, err)
	}
	delete(m, fileField)
	var text any = string(b)
	if subst && t.vars != nil {
		if text, err = substitute(text, t.vars, key+"."+fileField, t.syn); err != nil {
			return fmt.Errorf("%s: %w", t.file, err)
		}
	}
	m[field] = text
	if t.sources != nil {
		t.sources[key+"."+field] = []string{s}
	}
	return nil
}

// resolveTextFiles loads pools.*.prompt_file, workflows.*.steps[].prompt_file,
// triggers.*.ticket.body_file and prompt_rules[].file.
func resolveTextFiles(m map[string]any, t textFiles) error {
	pools, _ := m["pools"].(map[string]any)
	for _, n := range sortedKeys(pools) {
		if p, ok := pools[n].(map[string]any); ok {
			if err := t.load(p, "prompt", "prompt_file", "pools."+n, true); err != nil {
				return err
			}
		}
	}
	wfs, _ := m["workflows"].(map[string]any)
	for _, n := range sortedKeys(wfs) {
		wm, _ := wfs[n].(map[string]any)
		// Escalation chat prompts keep their vars until the ticket is
		// filed (#379).
		if em, ok := wm["escalate"].(map[string]any); ok {
			key := "workflows." + n + ".escalate"
			if err := t.load(em, "prompt", "prompt_file", key, false); err != nil {
				return err
			}
			if err := t.load(em, "pool_prompt", "pool_prompt_file", key, false); err != nil {
				return err
			}
		}
		steps, _ := wm["steps"].([]any)
		for i, s := range steps {
			// Step prompts keep ${run var} references for run start.
			if sm, ok := s.(map[string]any); ok {
				if err := t.load(sm, "prompt", "prompt_file", fmt.Sprintf("workflows.%s.steps[%d]", n, i), false); err != nil {
					return err
				}
			}
		}
	}
	trs, _ := m["triggers"].(map[string]any)
	for _, n := range sortedKeys(trs) {
		tm, _ := trs[n].(map[string]any)
		if tk, ok := tm["ticket"].(map[string]any); ok {
			if err := t.load(tk, "body", "body_file", "triggers."+n+".ticket", true); err != nil {
				return err
			}
		}
	}
	rules, _ := m["prompt_rules"].([]any)
	for i, r := range rules {
		rm, ok := r.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: prompt_rules[%d] must be a mapping", t.file, i)
		}
		key := fmt.Sprintf("prompt_rules[%d]", i)
		if n, ok := rm["name"].(string); ok && n != "" {
			key = "prompt_rules." + n
		}
		// Rule files stay literal, as before #382.
		if err := t.load(rm, "text", "file", key, false); err != nil {
			return err
		}
	}
	return nil
}

// prefixSources renames a pack layer's text file source keys after
// prefixLayers renamed its entries.
func prefixSources(src map[string][]string, prefix string) map[string][]string {
	out := make(map[string][]string, len(src))
	for k, v := range src {
		section, rest, ok := strings.Cut(k, ".")
		if ok && (section == "prompt_rules" || section == "pools" || section == "workflows" || section == "triggers") {
			k = section + "." + prefix + rest
		}
		out[k] = v
	}
	return out
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
