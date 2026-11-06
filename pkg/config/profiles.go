package config

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// checkLabelRules validates defaults.sandbox_by_label and
// defaults.isolation_by_label (JOB-83): labels are required (a rule with
// none is a default), and each rule's fields follow the rules of its
// level.
func (d Defaults) checkLabelRules(c *Config) error {
	for i, r := range d.SandboxByLabel {
		if err := checkRuleLabels(r.Labels); err != nil {
			return fmt.Errorf("%s: defaults.sandbox_by_label[%d]: %w", FileName, i, err)
		}
		if err := r.check(c.SandboxProfiles); err != nil {
			return fmt.Errorf("%s: defaults.sandbox_by_label[%d]: %w", FileName, i, err)
		}
	}
	for i, r := range d.IsolationByLabel {
		if err := checkRuleLabels(r.Labels); err != nil {
			return fmt.Errorf("%s: defaults.isolation_by_label[%d]: %w", FileName, i, err)
		}
		if err := r.check(c.IsolationProfiles); err != nil {
			return fmt.Errorf("%s: defaults.isolation_by_label[%d]: %w", FileName, i, err)
		}
	}
	return nil
}

func checkRuleLabels(labels []string) error {
	if len(labels) == 0 {
		return fmt.Errorf("labels is required: a rule with none would match every pool, set it under defaults instead")
	}
	return checkLabels(labels)
}

// checkSandboxes checks what only the merged config shows, after pack
// instance labels are added: the merged overlay, and sandboxes with no
// system mounts. A label rule that matches no pool is a warning: a typo
// in a label silently drops the profile otherwise.
func (c *Config) checkSandboxes() error {
	src := func(name string) string {
		if s := c.Sources["pools."+name]; len(s) > 0 {
			return strings.Join(s, ", ")
		}
		return FileName
	}
	for _, name := range slices.Sorted(maps.Keys(c.Pools)) {
		sb := c.SandboxFor(name)
		if err := checkMergedOverlay(sb); err != nil {
			return fmt.Errorf("%s: pool %s: %w", src(name), name, err)
		}
		// Without system mounts a sandboxed agent cannot exec (#625): the
		// paths moved from code to config, so say so instead of a bare
		// spawn failure.
		if sb.Backend != "" && len(sb.RO) == 0 && len(c.SharedTools) == 0 {
			c.Warnings = append(c.Warnings, fmt.Sprintf("%s: pool %s: sandboxed with no ro mounts; set sandbox ro or a sandbox profile with the host's system paths (#625)", src(name), name))
		}
	}
	matches := func(labels []string) bool {
		for _, p := range c.Pools {
			if p.HasLabels(labels) {
				return true
			}
		}
		return false
	}
	for i, r := range c.Defaults.SandboxByLabel {
		if !matches(r.Labels) {
			c.Warnings = append(c.Warnings, fmt.Sprintf("%s: defaults.sandbox_by_label[%d]: no pool has labels %v, the rule applies to nothing", FileName, i, r.Labels))
		}
	}
	for i, r := range c.Defaults.IsolationByLabel {
		if !matches(r.Labels) {
			c.Warnings = append(c.Warnings, fmt.Sprintf("%s: defaults.isolation_by_label[%d]: no pool has labels %v, the rule applies to nothing", FileName, i, r.Labels))
		}
	}
	return nil
}

// checkRemovedVars fails when the includer sets a var the pack removed.
func checkRemovedVars(given map[string]any, removed map[string]string, file string) error {
	for _, k := range slices.Sorted(maps.Keys(removed)) {
		if _, ok := given[k]; ok {
			return fmt.Errorf("%s: var %q is removed: %s", file, k, removed[k])
		}
	}
	return nil
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
