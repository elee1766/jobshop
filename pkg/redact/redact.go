// Package redact removes secrets from text and JSON before it leaves the
// controller: session.journal pages, journal shipping (#530), command
// step output, session events and stored writes (#574). Secret scanning
// is betterleaks' default rules, jobshop's built-ins (rules.toml) and the
// operator's config rules (#578).
package redact

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	blconfig "github.com/betterleaks/betterleaks/v2/config"
	"github.com/betterleaks/betterleaks/v2/scan"
)

// Mask replaces each redacted value.
const Mask = "[REDACTED]"

//go:embed rules.toml
var builtinRules string

// Rule is an operator rule from config (redact.rules), in betterleaks'
// rule fields.
type Rule struct {
	ID          string
	Description string
	Regex       string
	Keywords    []string
	// ValueGroup is the capture group holding the secret; 0 is the first
	// non-empty group, else the whole match.
	ValueGroup int
}

// Scanner is a compiled rule set. It is safe for concurrent use.
type Scanner struct{ s *scan.Scanner }

var (
	builtinOnce sync.Once
	builtinScan *Scanner
	builtinErr  error
)

// Builtin is the scanner with betterleaks' defaults and jobshop's
// built-in rules, compiled once.
func Builtin() *Scanner {
	builtinOnce.Do(func() {
		var cfg *blconfig.Config
		if cfg, builtinErr = builtinConfig(); builtinErr == nil {
			builtinScan, builtinErr = build(cfg)
		}
	})
	if builtinErr != nil {
		// The embedded rules are tested; this only fails on a broken build.
		panic(fmt.Sprintf("redact: built-in rules: %v", builtinErr))
	}
	return builtinScan
}

func builtinConfig() (*blconfig.Config, error) {
	return blconfig.ParseTOMLString(builtinRules, "rules.toml")
}

func build(cfg *blconfig.Config) (*Scanner, error) {
	s, err := scan.New(cfg, scan.WithPrecompile())
	if err != nil {
		return nil, err
	}
	return &Scanner{s: s}, nil
}

// Compile builds the scanner for config: defaults, then built-ins, then
// extra, without the disabled ids. A duplicate id, a bad regex or an
// unknown disabled id is an error naming the rule. With no extra and
// nothing disabled it is Builtin.
func Compile(extra []Rule, disabled []string) (*Scanner, error) {
	if len(extra) == 0 && len(disabled) == 0 {
		return Builtin(), nil
	}
	cfg, err := builtinConfig()
	if err != nil {
		return nil, err
	}
	for _, id := range disabled {
		i := slices.IndexFunc(cfg.Rules, func(r blconfig.Rule) bool { return r.ID == id })
		if i < 0 {
			return nil, fmt.Errorf("disabled_rules: no rule %q", id)
		}
		cfg.Rules = slices.Delete(cfg.Rules, i, i+1)
	}
	for _, r := range extra {
		if _, dup := cfg.Rule(r.ID); dup {
			return nil, fmt.Errorf("rule %q: id already used by a default or built-in rule", r.ID)
		}
		if len(r.Keywords) == 0 {
			return nil, fmt.Errorf("rule %q: keywords are required", r.ID)
		}
		kw := make([]string, len(r.Keywords))
		for i, k := range r.Keywords {
			kw[i] = strings.ToLower(k)
		}
		cfg.Rules = append(cfg.Rules, blconfig.Rule{
			ID: r.ID, Description: r.Description, Regex: r.Regex, Keywords: kw, ValueGroup: r.ValueGroup, Tags: []string{},
		})
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}
	s, err := build(cfg)
	if err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}
	return s, nil
}

// Finding is a match without the secret: what the CLI prints when it
// refuses an outbound write (#574).
type Finding struct {
	Rule string
	Line int // 1-based line in the scanned text
}

// Find returns the findings in s, without the secrets.
func (sc *Scanner) Find(s string) []Finding {
	if sc == nil || s == "" {
		return nil
	}
	var out []Finding
	for _, f := range sc.s.ScanString(s) {
		out = append(out, Finding{Rule: f.RuleID, Line: max(f.Location.StartLine, 1)})
	}
	return out
}

// mask replaces each finding's secret in s: the rule's value group when
// it has one (a URI's password, not the whole URI), else the match.
func (sc *Scanner) mask(s string) string {
	if sc == nil || s == "" {
		return s
	}
	for _, f := range sc.s.ScanString(s) {
		v := f.Match.Value
		if v == "" {
			v = f.Match.Full
		}
		if v != "" {
			s = strings.ReplaceAll(s, v, Mask)
		}
	}
	return s
}

// Redactor holds the secret values of one env and a scanner. Build it
// once per session and reuse it: the env scan is per call otherwise.
type Redactor struct {
	env []string // secret values, longest first
	sc  *Scanner
}

// Redactor returns a redactor for env (KEY=VALUE entries) with sc's
// rules: values of secret-named keys are replaced wherever they appear.
func (sc *Scanner) Redactor(env []string) *Redactor {
	return &Redactor{env: SecretEnvValues(env), sc: sc}
}

// New is Builtin().Redactor(env).
func New(env []string) *Redactor { return Builtin().Redactor(env) }

// Text replaces env secret values and scanner findings in s.
func (r *Redactor) Text(s string) string {
	for _, v := range r.env {
		s = strings.ReplaceAll(s, v, Mask)
	}
	return r.sc.mask(s)
}

// JSON redacts values under secret-named keys, then env secret values and
// findings in the encoded text. The result is valid JSON when raw is:
// replacements only touch string contents.
func (r *Redactor) JSON(raw json.RawMessage) json.RawMessage {
	s := r.Text(keys(raw))
	if !json.Valid([]byte(s)) {
		// A finding spanning an escape broke the encoding; fall back to
		// a string so nothing unredacted leaks.
		b, _ := json.Marshal(s)
		return b
	}
	return json.RawMessage(s)
}

// Text is New(env).Text(s).
func Text(s string, env []string) string { return New(env).Text(s) }

// JSON is New(env).JSON(raw).
func JSON(raw json.RawMessage, env []string) json.RawMessage { return New(env).JSON(raw) }

// Find is Builtin().Find(s).
func Find(s string) []Finding { return Builtin().Find(s) }

var secretKey = regexp.MustCompile(`(?i)(secret|token|passw(or)?d|api[_-]?key|auth|credential|private[_-]?key|session[_-]?key|cookie)`)

// SecretEnvValues returns values of secret-named env vars, longest first
// so a value containing another is replaced whole. Short values are
// skipped: replacing "1" everywhere would wreck the text.
func SecretEnvValues(env []string) []string {
	var out []string
	for _, kv := range env {
		k, v, ok := strings.Cut(kv, "=")
		if ok && len(v) >= 8 && secretKey.MatchString(k) {
			out = append(out, v)
		}
	}
	slices.SortFunc(out, func(a, b string) int { return len(b) - len(a) })
	return out
}

// keys re-encodes raw JSON with values under secret-named keys replaced.
// Non-JSON input is returned as is.
func keys(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return string(raw)
	}
	b, err := json.Marshal(value(v))
	if err != nil {
		return string(raw)
	}
	return string(b)
}

func value(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			if secretKey.MatchString(k) {
				t[k] = Mask
			} else {
				t[k] = value(x)
			}
		}
	case []any:
		for i, x := range t {
			t[i] = value(x)
		}
	}
	return v
}
