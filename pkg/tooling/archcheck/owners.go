package archcheck

import (
	"fmt"
	"slices"
	"strings"
)

// Owner is one row of the ownership table: where a package lives now, where
// it goes, which family owns it and which JOB child moves it. Every package of
// the module that has Go files needs exactly one row, so a new package cannot
// appear without a decision about where it belongs.
type Owner struct {
	// Pkg is the current path, relative to the module ("./pkg/x"), or an
	// "x/..." pattern for a subtree that moves as a whole.
	Pkg string
	// Target is where the package ends up, or "stay". A package that is
	// split lists its parts with "+".
	Target string
	// Family is a name from Config.Families.
	Family string
	// Child is the JOB issue that moves it, or "-" when it stays.
	Child string
	// Harness marks a test support package. It may import what its tests
	// need, but production code never imports it (rule harness-test-only).
	Harness bool
}

// Family is one ownership family of the agreed layout (JOB-164). The layout's
// namespace directories (orchestration, execution...) are families, and so are
// "wire" (the RPC protocol types, which most families speak) and "transport"
// (the adapters that start processes and serve users).
type Family struct {
	Name string
	// May are the other families this family's production code may import.
	// Importing inside the family is always fine.
	May []string
	// StdOnly forbids any third-party import. The shared domain package is
	// the user.
	StdOnly bool
}

// ownerFor finds the row for a package. The most specific row wins, so a
// subtree can have its own row.
func (c Config) ownerFor(g *Graph, path string) (Owner, bool) {
	var best Owner
	var bestLen int
	var ok bool
	for _, o := range c.Owners {
		if !g.Match(o.Pkg, path) {
			continue
		}
		if n := len(o.Pkg); !ok || (!strings.HasSuffix(o.Pkg, "...") && strings.HasSuffix(best.Pkg, "...")) || n > bestLen {
			best, bestLen, ok = o, n, true
		}
	}
	return best, ok
}

// familyOf maps every local package to its family.
func (c Config) familyOf(g *Graph) (map[string]string, []string) {
	var problems []string
	fam := map[string]string{}
	for _, path := range g.Local() {
		p := g.Pkgs[path]
		if p.Files+p.TestFiles == 0 {
			continue
		}
		o, ok := c.ownerFor(g, path)
		if !ok {
			problems = append(problems, fmt.Sprintf("package %s has no row in the ownership table (add one in rules.go)", g.Short(path)))
			continue
		}
		fam[path] = o.Family
	}
	return fam, problems
}

// checkTables reports rows that do not match a package, rows with a missing
// field, and unknown family names.
func (c Config) checkTables(g *Graph) []string {
	var problems []string
	families := map[string]bool{}
	for _, f := range c.Families {
		if families[f.Name] {
			problems = append(problems, "family "+f.Name+" is defined twice")
		}
		families[f.Name] = true
	}
	for _, f := range c.Families {
		for _, m := range f.May {
			if !families[m] {
				problems = append(problems, fmt.Sprintf("family %s may import unknown family %s", f.Name, m))
			}
		}
	}
	seen := map[string]bool{}
	for _, o := range c.Owners {
		switch {
		case o.Target == "" || o.Family == "" || o.Child == "":
			problems = append(problems, fmt.Sprintf("ownership row %s needs a target, a family and a child (or \"-\")", o.Pkg))
		case !families[o.Family]:
			problems = append(problems, fmt.Sprintf("ownership row %s names unknown family %s", o.Pkg, o.Family))
		}
		if seen[o.Pkg] {
			problems = append(problems, fmt.Sprintf("ownership row %s appears twice", o.Pkg))
		}
		seen[o.Pkg] = true
		if len(g.Select([]string{o.Pkg})) == 0 {
			problems = append(problems, fmt.Sprintf("ownership row %s matches no package: it moved or went away, update the row", o.Pkg))
		}
	}
	return problems
}

// familyEdges reports production imports between local packages whose
// families may not import each other. Test support packages are not
// checked as importers: harness-test-only keeps production code from using
// them, and they exist to reach across families.
func (c Config) familyEdges(g *Graph, fam map[string]string) []Violation {
	may := map[string][]string{}
	std := map[string]bool{}
	for _, f := range c.Families {
		may[f.Name] = f.May
		std[f.Name] = f.StdOnly
	}
	harness := map[string]bool{}
	for _, o := range c.Owners {
		if o.Harness {
			for _, p := range g.Select([]string{o.Pkg}) {
				harness[p] = true
			}
		}
	}
	var out []Violation
	for _, from := range g.Local() {
		ff, ok := fam[from]
		if !ok || harness[from] {
			continue
		}
		for _, imp := range g.Pkgs[from].Imports {
			q := g.Pkgs[imp]
			if q == nil || q.Std {
				continue
			}
			if !q.Main {
				if std[ff] {
					out = append(out, Violation{Rule: "family-" + ff, From: from, To: imp})
				}
				continue
			}
			ft, ok := fam[imp]
			if !ok || ft == ff || slices.Contains(may[ff], ft) {
				continue
			}
			out = append(out, Violation{Rule: "family-" + ff, From: from, To: imp})
		}
	}
	return out
}

// Inventory is one line per package for the reproducible inventory.
type Inventory struct {
	Pkg, Family, Target, Child string
	Files, TestFiles           int
	Prod, TestOnly             []string
	Third                      []string
}

// Inventory lists every package with Go files, its row and its edges.
func (c Config) Inventory(g *Graph) []Inventory {
	var out []Inventory
	for _, path := range g.Local() {
		p := g.Pkgs[path]
		if p.Files+p.TestFiles == 0 {
			continue
		}
		o, _ := c.ownerFor(g, path)
		inv := Inventory{Pkg: g.Short(path), Family: o.Family, Target: o.Target, Child: o.Child, Files: p.Files, TestFiles: p.TestFiles}
		for _, imp := range p.Imports {
			inv.Prod = append(inv.Prod, g.Short(imp))
		}
		for _, imp := range p.TestOnly() {
			inv.TestOnly = append(inv.TestOnly, g.Short(imp))
		}
		inv.Prod = slices.DeleteFunc(inv.Prod, func(s string) bool { return !isLocal(g, s) })
		inv.TestOnly = slices.DeleteFunc(inv.TestOnly, func(s string) bool { return !isLocal(g, s) })
		inv.Third = thirdParty(g, p.Imports)
		out = append(out, inv)
	}
	return out
}

func isLocal(g *Graph, short string) bool {
	q := g.Pkgs[g.Module+"/"+short]
	return q != nil && q.Main
}

// thirdParty lists the third-party modules a package imports directly, as
// the first three path elements, which is enough to tell modules apart.
func thirdParty(g *Graph, imports []string) []string {
	var out []string
	for _, imp := range imports {
		q := g.Pkgs[imp]
		if q == nil || q.Std || q.Main {
			continue
		}
		parts := strings.Split(imp, "/")
		if len(parts) > 3 {
			parts = parts[:3]
		}
		out = append(out, strings.Join(parts, "/"))
	}
	slices.Sort(out)
	return slices.Compact(out)
}
