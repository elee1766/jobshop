// Package archcheck checks the module's real import graph against the
// boundary rules in rules.go (JOB-165). It reads `go list -json`, so it sees
// what the compiler sees: build tags, generated files and the imports of test
// files, with production edges kept apart from test-only edges. It does not
// scan source text.
//
// Run it with `go run ./pkg/tooling/archcheck/cmd/archcheck`. CI does the same
// after `go vet`.
package archcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
)

// Package is one Go package of the graph.
type Package struct {
	Path string
	// Std is a standard library package. Main is a package of the module
	// under check. Anything else is a third-party dependency.
	Std, Main bool
	// Files and TestFiles count the Go files of the package and of its
	// internal and external tests.
	Files, TestFiles int
	// Imports are the production imports, sorted. TestImports are the
	// imports of the internal and external test files, sorted. A path may be
	// in both.
	Imports, TestImports []string
}

// Graph is the import graph of one module and everything it depends on.
type Graph struct {
	Module string
	Pkgs   map[string]*Package
}

// listPackage is the part of a `go list -json` record that is read.
type listPackage struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
	GoFiles      []string
	CgoFiles     []string
	TestGoFiles  []string
	XTestGoFiles []string
	Standard     bool
	Module       *struct {
		Path string
		Main bool
	}
	Error *struct{ Err string }
}

// listFields keeps `go list` from computing what is not used.
const listFields = "ImportPath,Imports,TestImports,XTestImports,GoFiles,CgoFiles,TestGoFiles,XTestGoFiles,Standard,Module,Error"

// Load runs `go list` in dir for every package of the module with its
// dependencies. tags are extra build tags, such as "victorialogs", so tagged
// test files are part of the graph.
func Load(ctx context.Context, dir, tags string) (*Graph, error) {
	args := []string{"list", "-e", "-deps", "-json=" + listFields}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	args = append(args, "./...")
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return Parse(&out)
}

// Parse builds a graph from the stream `go list -e -deps -json` writes.
func Parse(r io.Reader) (*Graph, error) {
	g := &Graph{Pkgs: map[string]*Package{}}
	dec := json.NewDecoder(r)
	var broken []string
	for {
		var lp listPackage
		if err := dec.Decode(&lp); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode go list: %w", err)
		}
		if lp.ImportPath == "unsafe" || lp.ImportPath == "C" {
			continue
		}
		if lp.Error != nil {
			broken = append(broken, lp.ImportPath+": "+strings.TrimSpace(lp.Error.Err))
		}
		p := &Package{
			Path:        lp.ImportPath,
			Std:         lp.Standard,
			Main:        lp.Module != nil && lp.Module.Main,
			Files:       len(lp.GoFiles) + len(lp.CgoFiles),
			TestFiles:   len(lp.TestGoFiles) + len(lp.XTestGoFiles),
			Imports:     clean(lp.Imports),
			TestImports: clean(slices.Concat(lp.TestImports, lp.XTestImports)),
		}
		if p.Main && g.Module == "" {
			g.Module = lp.Module.Path
		}
		g.Pkgs[p.Path] = p
	}
	if len(broken) > 0 {
		slices.Sort(broken)
		return nil, fmt.Errorf("go list reports broken packages (fix the build first):\n  %s", strings.Join(broken, "\n  "))
	}
	if g.Module == "" {
		return nil, errors.New("no package of the main module in the go list output")
	}
	return g, nil
}

// clean sorts and deduplicates import paths and drops cgo's pseudo import.
func clean(paths []string) []string {
	out := slices.DeleteFunc(slices.Clone(paths), func(p string) bool { return p == "C" || p == "unsafe" })
	slices.Sort(out)
	return slices.Compact(out)
}

// Local lists the paths of the module's packages, sorted.
func (g *Graph) Local() []string {
	var out []string
	for path, p := range g.Pkgs {
		if p.Main {
			out = append(out, path)
		}
	}
	slices.Sort(out)
	return out
}

// Short is path without the module prefix, for display.
func (g *Graph) Short(path string) string {
	if path == g.Module {
		return "."
	}
	if rest, ok := strings.CutPrefix(path, g.Module+"/"); ok {
		return rest
	}
	return path
}

// Match reports whether path matches pattern. A pattern starting with "./" is
// relative to the module. "x/..." matches the packages below x but not x
// itself, so a rule names a parent and its children separately.
func (g *Graph) Match(pattern, path string) bool {
	if rest, ok := strings.CutPrefix(pattern, "./"); ok {
		pattern = g.Module + "/" + rest
	}
	if base, ok := strings.CutSuffix(pattern, "/..."); ok {
		return strings.HasPrefix(path, base+"/")
	}
	return pattern == path
}

// MatchAny reports whether path matches one of patterns.
func (g *Graph) MatchAny(patterns []string, path string) bool {
	return slices.ContainsFunc(patterns, func(p string) bool { return g.Match(p, path) })
}

// Select returns the packages of the module that match any of patterns.
func (g *Graph) Select(patterns []string) []string {
	return slices.DeleteFunc(g.Local(), func(path string) bool { return !g.MatchAny(patterns, path) })
}

// TestOnly are the test imports of p that its production code does not have.
func (p *Package) TestOnly() []string {
	return slices.DeleteFunc(slices.Clone(p.TestImports), func(path string) bool {
		_, found := slices.BinarySearch(p.Imports, path)
		return found
	})
}
