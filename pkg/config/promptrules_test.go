package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/pack"
)

func TestRuleMatch(t *testing.T) {
	yes, no := true, false
	dev := RuleSubject{Pool: "dev", Agent: "jcode", Project: "js", Packs: []string{"github", "selfdev"}}
	step := dev
	step.Ticket, step.Repo, step.Workflow, step.Step, step.Labels = true, "gfx-labs/jobshop", "pr-work", "implement", []string{"x", "urgent"}
	chat := RuleSubject{Pool: "operator", Agent: "jcode", Project: "js", Chat: true}
	for _, tc := range []struct {
		name string
		m    RuleMatch
		s    RuleSubject
		want bool
	}{
		{"empty matches all", RuleMatch{}, chat, true},
		{"pool", RuleMatch{Pools: []string{"dev"}}, dev, true},
		{"pool other", RuleMatch{Pools: []string{"plan"}}, dev, false},
		{"pool glob", RuleMatch{Pools: []string{"d*"}}, dev, true},
		{"pack", RuleMatch{Packs: []string{"selfdev"}}, dev, true},
		{"pack other", RuleMatch{Packs: []string{"operator"}}, dev, false},
		{"repo glob", RuleMatch{Repos: []string{"gfx-labs/*"}}, step, true},
		{"repo other owner", RuleMatch{Repos: []string{"other/*"}}, step, false},
		{"repo needs a ticket", RuleMatch{Repos: []string{"*"}}, dev, false},
		{"step", RuleMatch{Steps: []string{"implement"}}, step, true},
		{"step other", RuleMatch{Steps: []string{"review-merge"}}, step, false},
		{"workflow", RuleMatch{Workflows: []string{"pr-*"}}, step, true},
		{"label any", RuleMatch{Labels: []string{"urgent", "y"}}, step, true},
		{"label none", RuleMatch{Labels: []string{"y"}}, step, false},
		{"chat true", RuleMatch{Chat: &yes}, chat, true},
		{"chat true on ticket pool", RuleMatch{Chat: &yes}, dev, false},
		{"chat false", RuleMatch{Chat: &no}, dev, true},
		{"all fields AND", RuleMatch{Pools: []string{"dev"}, Steps: []string{"implement"}}, step, true},
		{"all fields AND fails", RuleMatch{Pools: []string{"plan"}, Steps: []string{"implement"}}, step, false},
	} {
		assert.Equal(t, tc.want, tc.m.Matches(tc.s), tc.name)
	}
}

// TestPromptRulesMerge: pack rules come first in pack order, host rules
// last; a file rule is read relative to its pack; a pack cannot reuse a
// host rule's name.
func TestPromptRulesMerge(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p1/pack.yaml":   "pools:\n  dev:\n    agent: a\n    dir: /tmp\nprompt_rules:\n  - name: p1-rule\n    file: rules/p1.md\n    at: system\n",
		"p1/rules/p1.md": "from p1 ${x}\n",
		"p2/pack.yaml":   "prompt_rules:\n  - name: p2-rule\n    text: from p2\n    match: {pools: [dev]}\n",
		"rules/host.md":  "host file\n",
		FileName:         hostBase + "packs:\n  - path: p1\n  - path: p2\nprompt_rules:\n  - name: global\n    text: never push to master\n  - name: from-file\n    file: rules/host.md\n    at: before\n",
	})
	c, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	var names []string
	for _, r := range c.PromptRules {
		names = append(names, r.Name)
	}
	assert.Equal(t, []string{"p1-rule", "p2-rule", "global", "from-file"}, names)
	assert.Equal(t, "from p1 ${x}\n", c.PromptRules[0].Text, "rule files are literal")
	assert.Equal(t, RuleSystem, c.PromptRules[0].At)
	assert.Equal(t, RuleTurn, c.PromptRules[1].At, "default at")
	assert.Equal(t, "host file\n", c.PromptRules[3].Text)
	assert.Equal(t, []string{filepath.Join(dir, "p2", pack.FileName)}, c.Sources["prompt_rules.p2-rule"])
	assert.Equal(t, []string{"p1"}, c.PoolPacks("dev"))

	// A pack rule cannot take a host rule's name (and so cannot replace it).
	writeFiles(t, dir, map[string]string{
		"p2/pack.yaml": "prompt_rules:\n  - name: global\n    text: overridden\n",
	})
	_, err = LoadFile(filepath.Join(dir, FileName))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate name")
}

// TestPromptRulesPrefix: a pack included twice with prefixes (#184) gets
// its rule names and its own pool names in matches prefixed, so the two
// copies do not collide and each matches its own pools.
func TestPromptRulesPrefix(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"gh/pack.yaml": "pools:\n  dev:\n    agent: a\n    dir: /tmp\nprompt_rules:\n  - name: dev-rule\n    text: x\n    match: {pools: [dev, other]}\n",
		FileName:       hostBase + "packs:\n  - path: gh\n  - path: gh\n    prefix: gz-\n",
	})
	c, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	require.Len(t, c.PromptRules, 2)
	assert.Equal(t, "dev-rule", c.PromptRules[0].Name)
	assert.Equal(t, []string{"dev", "other"}, c.PromptRules[0].Match.Pools)
	assert.Equal(t, "gz-dev-rule", c.PromptRules[1].Name)
	assert.Equal(t, []string{"gz-dev", "other"}, c.PromptRules[1].Match.Pools, "host pools stay")
}

func TestPromptRuleErrors(t *testing.T) {
	for _, tc := range []struct{ rules, want string }{
		{"  - text: x\n", "name"},
		{"  - name: a\n", "set text, file, command or memory"},
		{"  - {name: a, memory: {dir: x, cap: -1}}\n", "memory.cap must not be negative"},
		{"  - name: a\n    text: x\n    at: middle\n", "at \"middle\""},
		{"  - name: a\n    text: x\n    timeout: 1s\n", "timeout needs a command"},
		{"  - name: a\n    command: echo\n    timeout: soon\n", "positive duration"},
		{"  - name: a\n    text: x\n    match: {pools: [\"[\"]}\n", "bad glob"},
		{"  - name: a\n    text: x\n    file: f\n", "not both"},
		{"  - name: a\n    text: x\n  - name: a\n    text: y\n", "duplicate name"},
	} {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{FileName: hostBase + "prompt_rules:\n" + tc.rules})
		_, err := LoadFile(filepath.Join(dir, FileName))
		if assert.Error(t, err, tc.rules) {
			assert.Contains(t, err.Error(), tc.want, tc.rules)
		}
	}
}
