package config

import (
	"fmt"

	"github.com/gfx-labs/jobshop/pkg/support/redact"
)

// Redact configures secret scanning (#578): rules added to betterleaks'
// defaults and jobshop's built-ins, and ids to turn off. Host-only:
// packs cannot set it.
type Redact struct {
	Rules         []RedactRule `yaml:"rules"`
	DisabledRules []string     `yaml:"disabled_rules"`
	// scanner is the compiled rule set, built at load.
	scanner *redact.Scanner
}

// RedactRule is one operator rule, in betterleaks' rule fields.
type RedactRule struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Regex       string   `yaml:"regex"`
	Keywords    []string `yaml:"keywords"`
	ValueGroup  int      `yaml:"value_group"`
}

// compile builds the scanner; an error names the rule.
func (r *Redact) compile() error {
	rules := make([]redact.Rule, len(r.Rules))
	for i, x := range r.Rules {
		rules[i] = redact.Rule{ID: x.ID, Description: x.Description, Regex: x.Regex, Keywords: x.Keywords, ValueGroup: x.ValueGroup}
	}
	sc, err := redact.Compile(rules, r.DisabledRules)
	if err != nil {
		return fmt.Errorf("redact: %w", err)
	}
	r.scanner = sc
	return nil
}

// Scanner is the compiled rule set: the built-in one when the config set
// no rules.
func (r Redact) Scanner() *redact.Scanner {
	if r.scanner == nil {
		return redact.Builtin()
	}
	return r.scanner
}
