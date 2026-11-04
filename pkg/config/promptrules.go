package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Prompt rule positions (PromptRule.At).
const (
	// RuleSystem is sent once per host: the first turn, and again on the
	// first turn after a respawn.
	RuleSystem = "system"
	// RuleBefore and RuleAfter wrap the pool prompt: every ticket turn,
	// and a chat session's first message turn.
	RuleBefore = "before"
	RuleAfter  = "after"
	// RuleTurn is sent on every ticket and message turn.
	RuleTurn = "turn"
)

// Prompt rule hook limits.
const (
	DefaultRuleTimeout = 10 * time.Second
	MaxRuleOutput      = 16 << 10
)

// PromptRule adds operator text to the prompts of sessions it matches
// (#205). Rules from packs come first in pack order, then the host file's;
// names are unique, so a pack cannot replace a host rule.
type PromptRule struct {
	Name  string    `yaml:"name"`
	Match RuleMatch `yaml:"match"`
	// Text is the guidance. File is read into Text at load, relative to
	// the declaring file; a reload reads it again.
	Text string `yaml:"text"`
	File string `yaml:"file,omitempty"`
	// At is system, before, after or turn. Default turn.
	At string `yaml:"at"`
	// Command runs with sh -c in the root dir, a RuleEvent as JSON on
	// stdin; its stdout is added after Text. A non-zero exit or timeout
	// adds nothing.
	Command string `yaml:"command"`
	// Timeout bounds Command (Go duration). Default 10s.
	Timeout string `yaml:"timeout"`
}

// RuleMatch selects sessions. Every set field must match; a list matches
// when any of its globs (path.Match) does. An empty match is every session.
// Repos, workflows, steps and labels describe the turn's ticket and never
// match a turn without one.
type RuleMatch struct {
	Pools     []string `yaml:"pools,omitempty"`
	Agents    []string `yaml:"agents,omitempty"`
	Packs     []string `yaml:"packs,omitempty"`
	Projects  []string `yaml:"projects,omitempty"`
	Repos     []string `yaml:"repos,omitempty"`
	Workflows []string `yaml:"workflows,omitempty"`
	Steps     []string `yaml:"steps,omitempty"`
	Labels    []string `yaml:"labels,omitempty"`
	Chat      *bool    `yaml:"chat,omitempty"`
	// Keys match the session key (a chat thread such as matrix:<room>, or
	// improve:<slug> for an improvement session, #249).
	Keys []string `yaml:"keys,omitempty"`
}

// RuleSubject is what a rule matches against for one turn.
type RuleSubject struct {
	Pool, Agent, Project string
	Packs                []string // pack dir names that declared the pool
	Chat                 bool
	Key                  string // the session key, "" for none
	// Ticket is false for system and message turns.
	Ticket               bool
	Repo, Workflow, Step string
	Labels               []string
}

// Matches reports whether the rule applies to s.
func (m RuleMatch) Matches(s RuleSubject) bool {
	if m.Chat != nil && *m.Chat != s.Chat {
		return false
	}
	one := func(globs []string, v string) bool { return len(globs) == 0 || globAny(globs, v) }
	if !one(m.Pools, s.Pool) || !one(m.Agents, s.Agent) || !one(m.Projects, s.Project) {
		return false
	}
	if len(m.Keys) > 0 && !globAny(m.Keys, s.Key) {
		return false
	}
	if len(m.Packs) > 0 && !anyOf(m.Packs, s.Packs) {
		return false
	}
	ticketOnly := len(m.Repos) > 0 || len(m.Workflows) > 0 || len(m.Steps) > 0 || len(m.Labels) > 0
	if !ticketOnly {
		return true
	}
	if !s.Ticket {
		return false
	}
	if !one(m.Repos, s.Repo) || !one(m.Workflows, s.Workflow) || !one(m.Steps, s.Step) {
		return false
	}
	return len(m.Labels) == 0 || anyOf(m.Labels, s.Labels)
}

func globAny(globs []string, v string) bool {
	if v == "" {
		return false
	}
	for _, g := range globs {
		if ok, err := path.Match(g, v); err == nil && ok {
			return true
		}
	}
	return false
}

func anyOf(globs, vs []string) bool {
	for _, v := range vs {
		if globAny(globs, v) {
			return true
		}
	}
	return false
}

// resolveRuleFiles reads prompt_rules[].file into text, relative to dir
// (the declaring file's directory), like a pool's prompt_file.
func resolveRuleFiles(m map[string]any, dir, file string) error {
	rules, _ := m["prompt_rules"].([]any)
	for i, r := range rules {
		rm, _ := r.(map[string]any)
		if rm == nil {
			return fmt.Errorf("%s: prompt_rules[%d] must be a mapping", file, i)
		}
		if f, ok := rm["file"]; ok {
			if _, both := rm["text"]; both {
				return fmt.Errorf("%s: prompt_rules[%d]: set text or file, not both", file, i)
			}
			s, _ := f.(string)
			if s == "" {
				return fmt.Errorf("%s: prompt_rules[%d].file must be a path", file, i)
			}
			if !filepath.IsAbs(s) {
				s = filepath.Join(dir, s)
			}
			b, err := os.ReadFile(s)
			if err != nil {
				return fmt.Errorf("%s: prompt_rules[%d].file: %w", file, i, err)
			}
			delete(rm, "file")
			rm["text"] = string(b)
		}
	}
	return nil
}

// checkRules validates c.PromptRules: names unique and valid, a known
// position, text or a command, and a valid timeout.
func (c *Config) checkRules() error {
	seen := map[string]bool{}
	for i, r := range c.PromptRules {
		where := fmt.Sprintf("prompt_rules[%d]", i)
		if r.Name != "" {
			where = "prompt_rule " + r.Name
			if s := c.Sources["prompt_rules."+r.Name]; len(s) > 0 {
				where = fmt.Sprintf("%s: %s", s[len(s)-1], where)
			}
		}
		if !namePattern.MatchString(r.Name) {
			return fmt.Errorf("%s: name %q must match %s", where, r.Name, namePattern)
		}
		if seen[r.Name] {
			return fmt.Errorf("%s: duplicate name (declared in %v); a pack cannot replace another file's rule", where, c.Sources["prompt_rules."+r.Name])
		}
		seen[r.Name] = true
		if r.File != "" {
			return fmt.Errorf("%s: file needs a config file location", where)
		}
		switch r.At {
		case "":
			c.PromptRules[i].At = RuleTurn
		case RuleSystem, RuleBefore, RuleAfter, RuleTurn:
		default:
			return fmt.Errorf("%s: at %q must be system, before, after or turn", where, r.At)
		}
		if r.Text == "" && r.Command == "" {
			return fmt.Errorf("%s: set text, file or command", where)
		}
		if r.Timeout != "" {
			if r.Command == "" {
				return fmt.Errorf("%s: timeout needs a command", where)
			}
			if d, err := time.ParseDuration(r.Timeout); err != nil || d <= 0 {
				return fmt.Errorf("%s: timeout %q must be a positive duration", where, r.Timeout)
			}
		}
		for _, g := range ruleGlobs(r.Match) {
			if _, err := path.Match(g, ""); err != nil {
				return fmt.Errorf("%s: bad glob %q", where, g)
			}
		}
	}
	return nil
}

func ruleGlobs(m RuleMatch) []string {
	var out []string
	for _, l := range [][]string{m.Pools, m.Agents, m.Packs, m.Projects, m.Repos, m.Workflows, m.Steps, m.Labels, m.Keys} {
		out = append(out, l...)
	}
	return out
}

// PoolPacks returns the directory names of the packs that set pool, in
// merge order: what a rule's match.packs compares against.
func (c *Config) PoolPacks(pool string) []string {
	var out []string
	for _, f := range c.Sources["pools."+pool] {
		if filepath.Base(f) == PackFileName {
			out = append(out, filepath.Base(filepath.Dir(f)))
		}
	}
	return out
}
