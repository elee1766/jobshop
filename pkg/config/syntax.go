package config

import (
	"fmt"
	"sort"

	"github.com/elee1766/jobshop/pkg/config/tmpl"
)

// Template syntaxes (#383). SyntaxDollar is ${name} with $$ for a literal
// $; SyntaxGo is Go text/template with missingkey=error and tmpl.Funcs.
// ${name} goes away once hosts' own packs have moved.
const (
	SyntaxDollar = ""
	SyntaxGo     = "go"
)

func checkSyntax(s string) error {
	if s != SyntaxDollar && s != SyntaxGo {
		return fmt.Errorf("%q must be empty (${name}) or %q", s, SyntaxGo)
	}
	return nil
}

// placeholder is the load-time value of a var filled later (check
// output, stall vars): a reference to itself, left for the later render.
func placeholder(name, syn string) string {
	if syn == SyntaxGo {
		return "{{." + name + "}}"
	}
	return "${" + name + "}"
}

// setSyntax records the pack's syntax on the workflows and trigger runs
// it declares, so run start and trigger firing render them the same way.
func setSyntax(m map[string]any, syn string) {
	if syn == SyntaxDollar {
		return
	}
	wfs, _ := m["workflows"].(map[string]any)
	for _, w := range wfs {
		if wm, ok := w.(map[string]any); ok {
			wm["template"] = syn
			if e, ok := wm["escalate"].(map[string]any); ok {
				e["template"] = syn
			}
		}
	}
	if e, ok := m["escalation"].(map[string]any); ok {
		e["template"] = syn
	}
	trs, _ := m["triggers"].(map[string]any)
	for _, t := range trs {
		tm, _ := t.(map[string]any)
		if rm, ok := tm["run"].(map[string]any); ok {
			rm["template"] = syn
		}
	}
}

// leftoverVars warns about ${name} left in a Go-template pack, which
// is a missed migration more often than a literal.
func leftoverVars(m map[string]any, file string) []string {
	var keys []string
	var walk func(v any, key string)
	walk = func(v any, key string) {
		switch x := v.(type) {
		case map[string]any:
			for k, e := range x {
				walk(e, joinKey(key, k))
			}
		case []any:
			for i, e := range x {
				walk(e, fmt.Sprintf("%s[%d]", key, i))
			}
		case string:
			if m := varPattern.FindStringSubmatch(x); m != nil && m[1] != "" {
				keys = append(keys, fmt.Sprintf("%s: %s has ${%s} in a template: go pack; use {{.%s}}", file, key, m[1], m[1]))
			}
		}
	}
	walk(m, "")
	sort.Strings(keys)
	return keys
}

// Fill renders s in syntax syn from vars. A missing var is an error.
func Fill(s, syn string, vars map[string]any) (string, error) {
	if syn == SyntaxGo {
		return tmpl.Render(s, vars)
	}
	return FillTemplate(s, vars)
}

// refs returns the var names s references in syntax syn.
func refs(s, syn string) ([]string, error) {
	if syn == SyntaxGo {
		return tmpl.Vars(s)
	}
	return templateVars(s), nil
}
