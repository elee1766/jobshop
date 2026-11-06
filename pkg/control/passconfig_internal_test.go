package control

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// passRoots are the functions that run a supervisor pass or part of one
// under tickMu: tick, and the trigger event path (JOB-104).
var passRoots = []string{"tick", "onEvent"}

// passLoads may read the live config: tick loads it once per pass.
var passLoads = map[string]bool{"tick": true}

// liveReads are the Controller members that read the live config. Pass
// code takes the pass's cfg instead, so a reload mid-pass cannot mix two
// configs in one pass.
var liveReads = map[string]bool{"config": true, "live": true, "leaseDur": true, "scanner": true, "redactor": true}

// TestPassReadsConfigOnce walks the calls from the pass roots through
// Controller methods and package functions, and fails on any that reads
// the live config (JOB-104). Calls through other receivers (a runner's
// r.c) and goroutines the pass starts (go, wg.Go) are not followed: they
// run beside the pass, not in it.
func TestPassReadsConfigOnce(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	// calls maps a Controller method or package func to the names it
	// reaches through c.<name> or a plain call; reads are its live reads.
	calls := map[string][]string{}
	reads := map[string][]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range af.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			name, recv := fd.Name.Name, ""
			if fd.Recv != nil {
				if !isController(fd.Recv.List[0].Type) || len(fd.Recv.List[0].Names) == 0 {
					continue
				}
				recv = fd.Recv.List[0].Names[0].Name
			} else {
				name = "func " + name
			}
			var cs []string
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.GoStmt:
					return false
				case *ast.CallExpr:
					if s, ok := x.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "Go" {
						return false
					}
					if id, ok := x.Fun.(*ast.Ident); ok {
						cs = append(cs, "func "+id.Name)
					}
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && recv != "" && id.Name == recv {
						if liveReads[x.Sel.Name] {
							reads[name] = append(reads[name], fset.Position(x.Pos()).String())
						} else {
							cs = append(cs, x.Sel.Name)
						}
					}
				}
				return true
			})
			calls[name] = cs
		}
	}
	for _, r := range passRoots {
		if _, ok := calls[r]; !ok {
			t.Fatalf("pass root %s not found", r)
		}
	}
	from := map[string]string{}
	q := slices.Clone(passRoots)
	for _, r := range passRoots {
		from[r] = ""
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		for _, m := range calls[n] {
			if _, ok := calls[m]; !ok {
				continue
			}
			if _, ok := from[m]; ok {
				continue
			}
			from[m] = n
			q = append(q, m)
		}
	}
	var bad []string
	for name := range from {
		if passLoads[name] {
			continue
		}
		for _, at := range reads[name] {
			path := []string{name}
			for p := from[name]; p != ""; p = from[p] {
				path = append(path, p)
			}
			slices.Reverse(path)
			bad = append(bad, at+": "+strings.Join(path, " -> "))
		}
	}
	slices.Sort(bad)
	for _, b := range bad {
		t.Errorf("pass code reads the live config; take the pass cfg instead: %s", b)
	}
}

func isController(e ast.Expr) bool {
	if s, ok := e.(*ast.StarExpr); ok {
		e = s.X
	}
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "Controller"
}
