package pack

import (
	"fmt"
	"regexp"
	"sort"
)

// Pack templates are Go text/template with missingkey=error and the
// functions in pack/tmpl (#383). The older ${name} syntax was removed in
// #566.

// templateGo is the only accepted value of a pack's template: key, kept
// as a no-op for one release so packs that still set it load.
const templateGo = "go"

func checkSyntax(s string) error {
	if s != "" && s != templateGo {
		return fmt.Errorf("%q is not supported: packs use Go templates ({{.name}}); the ${name} syntax was removed in #566", s)
	}
	return nil
}

// Placeholder is the load-time value of a var filled later (check
// output, scan vars): a reference to itself, left for the later render.
func Placeholder(name string) string {
	return "{{." + name + "}}"
}

// dollarRef is a leftover ${name} from the removed syntax.
var dollarRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// dollarAllowed are ${name} forms that are not pack syntax: the
// controller expands ${SESSION_DIR} in agent args at session start.
var dollarAllowed = map[string]bool{"SESSION_DIR": true}

// leftoverVars lists ${name} left in a pack. Since #566 that is a load
// error: it would render as literal text.
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
			for _, m := range dollarRef.FindAllStringSubmatch(x, -1) {
				if !dollarAllowed[m[1]] {
					keys = append(keys, fmt.Sprintf("%s: %s has ${%s}; the ${name} syntax was removed in #566, use {{.%s}}", file, key, m[1], m[1]))
					break
				}
			}
		}
	}
	walk(m, "")
	sort.Strings(keys)
	return keys
}
