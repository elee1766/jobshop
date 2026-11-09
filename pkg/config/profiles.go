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
