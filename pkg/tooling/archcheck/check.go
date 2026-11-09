package archcheck

import (
	"fmt"
	"slices"
	"strings"
)

// Scope says which edges a rule looks at.
type Scope int

const (
	// Prod looks at the direct imports of non-test files.
	Prod Scope = iota
	// ProdAndTest also looks at the direct imports of test files.
	ProdAndTest
	// Closure looks at everything the production code of a package reaches,
	// third-party packages included. Use it where a direct import is not the
	// problem but the code it drags in is.
	Closure
	// TestOnly looks at the imports only test files have, so a rule can say
	// "production code never imports a test harness" while tests may.
	TestOnly
)

func (s Scope) String() string {
	return [...]string{"production", "production or test", "production closure", "test-only"}[s]
}

// Rule forbids edges from packages matching From (minus Except) to packages
// matching To. Patterns are as in Graph.Match, and "std" matches the standard
// library. A pattern with no path in the graph fails the check, so renaming
// or moving a package cannot leave a rule that silently checks nothing.
type Rule struct {
	Name   string
	Why    string
	Scope  Scope
	From   []string
	Except []string
	To     []string
}

// Only is the opposite shape: packages matching From may import only what
// matches Allow, plus the standard library when Std is set and any
// third-party package when Third is set. It is for a leaf that must stay
// small. Test files count: a leaf's tests do not reach into the module either,
// or the leaf would not be usable on its own.
type Only struct {
	Name  string
	Why   string
	From  []string
	Allow []string
	Std   bool
	Third bool
}

// Known is one accepted existing violation: a single edge of a single rule,
// with the child that removes it. An entry that no longer matches a violation
// fails the check, so the list shrinks as the migration lands and cannot
// become a blanket exemption. A Closure exception names the reached package.
type Known struct {
	Rule  string
	From  string
	To    string
	Issue string
	Why   string
}

// Violation is one broken rule.
type Violation struct {
	Rule     string
	From, To string
	// Path is the import chain for a Closure rule, From first and To last.
	Path []string
	Test bool
}

func (v Violation) String() string {
	via := ""
	if len(v.Path) > 2 {
		via = " (via " + strings.Join(v.Path[1:len(v.Path)-1], " -> ") + ")"
	}
	kind := "imports"
	if v.Test {
		kind = "imports in a test file"
	}
	return fmt.Sprintf("[%s] %s %s %s%s", v.Rule, v.From, kind, v.To, via)
}

// Config is everything one check needs.
type Config struct {
	Owners   []Owner
	Families []Family
	Rules    []Rule
	Only     []Only
	Known    []Known
}

// Result is the outcome of one check.
type Result struct {
	// Problems are violations not covered by Known, plus structural errors
	// in the tables themselves. A non-empty list fails the check.
	Problems []string
	// Accepted are the violations a Known entry covers.
	Accepted []Violation
}

// Check evaluates cfg against g.
func Check(g *Graph, cfg Config) Result {
	var res Result
	var found []Violation

	fam, errs := cfg.familyOf(g)
	res.Problems = append(res.Problems, errs...)
	res.Problems = append(res.Problems, cfg.checkTables(g)...)

	for _, r := range cfg.Rules {
		vs, errs := evalRule(g, r)
		res.Problems = append(res.Problems, errs...)
		found = append(found, vs...)
	}
	for _, o := range cfg.Only {
		vs, errs := evalOnly(g, o)
		res.Problems = append(res.Problems, errs...)
		found = append(found, vs...)
	}
	found = append(found, cfg.familyEdges(g, fam)...)

	used := make([]bool, len(cfg.Known))
	for _, v := range found {
		i := slices.IndexFunc(cfg.Known, func(k Known) bool {
			return k.Rule == v.Rule && g.Match(k.From, v.From) && g.Match(k.To, v.To)
		})
		if i >= 0 {
			used[i] = true
			res.Accepted = append(res.Accepted, v)
			continue
		}
		res.Problems = append(res.Problems, v.String())
	}
	for i, k := range cfg.Known {
		if k.Issue == "" || k.Why == "" {
			res.Problems = append(res.Problems, fmt.Sprintf("known exception %s: %s -> %s needs an issue and a reason", k.Rule, k.From, k.To))
		}
		if !used[i] {
			res.Problems = append(res.Problems, fmt.Sprintf("stale exception [%s] %s -> %s (%s): the edge is gone, delete the entry", k.Rule, k.From, k.To, k.Issue))
		}
	}
	slices.Sort(res.Problems)
	return res
}

// resolve expands patterns to local package paths, or to third-party and
// standard library paths too when wide is set. It reports patterns that
// match nothing.
func resolve(g *Graph, patterns []string, wide bool) ([]string, []string) {
	var out, missing []string
	for _, pat := range patterns {
		var hit bool
		for path, p := range g.Pkgs {
			if !wide && !p.Main {
				continue
			}
			if matchTarget(g, pat, p, path) {
				out = append(out, path)
				hit = true
			}
		}
		if !hit {
			missing = append(missing, pat)
		}
	}
	slices.Sort(out)
	return slices.Compact(out), missing
}

// matchTarget is Graph.Match plus the "std" pattern.
func matchTarget(g *Graph, pat string, p *Package, path string) bool {
	if pat == "std" {
		return p.Std
	}
	return g.Match(pat, path)
}

func evalRule(g *Graph, r Rule) ([]Violation, []string) {
	var problems []string
	from, missing := resolve(g, r.From, false)
	for _, m := range missing {
		problems = append(problems, fmt.Sprintf("rule %s: From %q matches no package", r.Name, m))
	}
	to, missing := resolve(g, r.To, true)
	for _, m := range missing {
		problems = append(problems, fmt.Sprintf("rule %s: To %q matches no package", r.Name, m))
	}
	except := map[string]bool{}
	if len(r.Except) > 0 {
		ex, missing := resolve(g, r.Except, false)
		for _, m := range missing {
			problems = append(problems, fmt.Sprintf("rule %s: Except %q matches no package", r.Name, m))
		}
		for _, e := range ex {
			except[e] = true
		}
	}
	target := map[string]bool{}
	for _, t := range to {
		target[t] = true
	}

	var out []Violation
	for _, f := range from {
		if except[f] {
			continue
		}
		p := g.Pkgs[f]
		switch r.Scope {
		case Prod, ProdAndTest:
			for _, imp := range p.Imports {
				if target[imp] && imp != f {
					out = append(out, Violation{Rule: r.Name, From: f, To: imp})
				}
			}
			if r.Scope == ProdAndTest {
				for _, imp := range p.TestOnly() {
					// An external test package (x_test) importing x is not an edge.
					if target[imp] && imp != f {
						out = append(out, Violation{Rule: r.Name, From: f, To: imp, Test: true})
					}
				}
			}
		case TestOnly:
			for _, imp := range p.TestOnly() {
				if target[imp] {
					out = append(out, Violation{Rule: r.Name, From: f, To: imp, Test: true})
				}
			}
		case Closure:
			paths := reach(g, f, target)
			for _, t := range slices.Sorted(keys(paths)) {
				out = append(out, Violation{Rule: r.Name, From: f, To: t, Path: paths[t]})
			}
		}
	}
	return out, problems
}

func keys[V any](m map[string]V) func(func(string) bool) {
	return func(yield func(string) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

// reach walks the production imports from start and returns, for every
// target package reached, the shortest chain to it. The walk does not
// continue past a target: what a forbidden package pulls in is its own
// finding. When a package is reachable through a forbidden target only, the
// target is what a developer must cut.
func reach(g *Graph, start string, target map[string]bool) map[string][]string {
	prev := map[string]string{start: ""}
	queue := []string{start}
	found := map[string][]string{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		p := g.Pkgs[cur]
		if p == nil {
			continue
		}
		for _, imp := range p.Imports {
			if _, seen := prev[imp]; seen {
				continue
			}
			prev[imp] = cur
			if target[imp] {
				found[imp] = chain(prev, imp)
				continue
			}
			queue = append(queue, imp)
		}
	}
	return found
}

func chain(prev map[string]string, end string) []string {
	var out []string
	for at := end; at != ""; at = prev[at] {
		out = append(out, at)
	}
	slices.Reverse(out)
	return out
}

// Why returns the shortest production import chain from one package to
// another, or nil.
func Why(g *Graph, from, to string) []string {
	return reach(g, from, map[string]bool{to: true})[to]
}

func evalOnly(g *Graph, o Only) ([]Violation, []string) {
	var problems []string
	from, missing := resolve(g, o.From, false)
	for _, m := range missing {
		problems = append(problems, fmt.Sprintf("rule %s: From %q matches no package", o.Name, m))
	}
	var out []Violation
	for _, f := range from {
		p := g.Pkgs[f]
		for _, imp := range slices.Concat(p.Imports, p.TestOnly()) {
			q := g.Pkgs[imp]
			switch {
			case q == nil, imp == f:
			case q.Std && o.Std:
			case !q.Std && !q.Main && o.Third:
			case g.MatchAny(o.Allow, imp):
			default:
				_, prod := slices.BinarySearch(p.Imports, imp)
				out = append(out, Violation{Rule: o.Name, From: f, To: imp, Test: !prod})
			}
		}
	}
	return out, problems
}
