// Package tmpl renders pack and config templates with Go text/template
// (#383): missing vars are errors, and a small FuncMap with no shell or file
// access. It is the one template language of the pack subsystem and of the
// config values rendered after load (JOB-172).
package tmpl

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/goccy/go-yaml"
)

// funcs are the functions templates may call.
var funcs = template.FuncMap{
	// default returns d when v is empty ("", 0, false, nil, empty list).
	"default": func(d, v any) any {
		if empty(v) {
			return d
		}
		return v
	},
	"join":   join,
	"quote":  func(v any) string { return strconv.Quote(fmt.Sprint(v)) },
	"lower":  func(v any) string { return strings.ToLower(fmt.Sprint(v)) },
	"indent": indent,
	"toYAML": toYAML,
}

func empty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case bool:
		return !x
	case int:
		return x == 0
	case uint64:
		return x == 0
	case int64:
		return x == 0
	case float64:
		return x == 0
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}

// join joins a list with sep. A string is returned as is, so a
// comma-separated var and a list var both work.
func join(sep string, v any) string {
	switch x := v.(type) {
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = fmt.Sprint(e)
		}
		return strings.Join(parts, sep)
	case []string:
		return strings.Join(x, sep)
	}
	return fmt.Sprint(v)
}

// indent prefixes every line but the first with n spaces, for a value
// placed under a YAML or Markdown block.
func indent(n int, v any) string {
	pad := strings.Repeat(" ", n)
	return strings.ReplaceAll(fmt.Sprint(v), "\n", "\n"+pad)
}

func toYAML(v any) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(b), "\n"), nil
}

// compile parses s with missingkey=error and funcs.
func compile(s string) (*template.Template, error) {
	t, err := template.New("").Option("missingkey=error").Funcs(funcs).Parse(s)
	if err != nil {
		return nil, clean(err)
	}
	return t, nil
}

// Render parses and executes s over vars.
func Render(s string, vars map[string]any) (string, error) {
	if !strings.Contains(s, "{{") {
		return s, nil
	}
	t, err := compile(s)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, vars); err != nil {
		return "", clean(err)
	}
	return b.String(), nil
}

var whole = regexp.MustCompile(`^\{\{-?\s*\.([A-Za-z_][A-Za-z0-9_]*)\s*-?\}\}$`)

// Value renders s, except that a string that is exactly one {{.name}}
// takes the var's value with its type, so numbers and lists stay typed.
func Value(s string, vars map[string]any) (any, error) {
	if m := whole.FindStringSubmatch(s); m != nil {
		v, ok := vars[m[1]]
		if !ok {
			return nil, fmt.Errorf("var %q is not set", m[1])
		}
		return v, nil
	}
	return Render(s, vars)
}

var noEntry = regexp.MustCompile(`map has no entry for key "([^"]*)"`)

// clean turns text/template's errors into the loader's wording: the
// template name is empty, and a missing key reads "var x is not set".
func clean(err error) error {
	msg := strings.TrimPrefix(err.Error(), "template: ")
	if m := noEntry.FindStringSubmatch(msg); m != nil {
		return fmt.Errorf("var %q is not set (%s)", m[1], msg)
	}
	return fmt.Errorf("template %s", msg)
}

// Vars returns the top-level var names s references (.name), sorted.
// Fields inside range and with bodies, where dot is not the vars map,
// are skipped.
func Vars(s string) ([]string, error) {
	if !strings.Contains(s, "{{") {
		return nil, nil
	}
	t, err := compile(s)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	walk(t.Root, set)
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}

func walk(n parse.Node, set map[string]bool) {
	switch x := n.(type) {
	case *parse.ListNode:
		if x == nil {
			return
		}
		for _, c := range x.Nodes {
			walk(c, set)
		}
	case *parse.ActionNode:
		walk(x.Pipe, set)
	case *parse.PipeNode:
		if x == nil {
			return
		}
		for _, c := range x.Cmds {
			walk(c, set)
		}
	case *parse.CommandNode:
		for _, a := range x.Args {
			walk(a, set)
		}
	case *parse.FieldNode:
		set[x.Ident[0]] = true
	case *parse.ChainNode:
		walk(x.Node, set)
	case *parse.IfNode:
		walk(x.Pipe, set)
		walk(x.List, set)
		walk(x.ElseList, set)
	case *parse.RangeNode:
		walk(x.Pipe, set)
		walk(x.ElseList, set)
	case *parse.WithNode:
		walk(x.Pipe, set)
		walk(x.ElseList, set)
	case *parse.TemplateNode:
		walk(x.Pipe, set)
	}
}
