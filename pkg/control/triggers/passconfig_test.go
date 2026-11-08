package triggers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestPassConfigInEngine: pass code takes the config the pass holds
// (JOB-104). Only the entry points that begin a new pass, an event job
// (Events) and a check replay (checkLoop), read the host's live config. It
// is the engine's half of pkg/control's TestPassReadsConfigOnce, which
// cannot follow calls through the Host interface.
func TestPassConfigInEngine(t *testing.T) {
	t.Parallel()
	allowed := map[string]bool{"Events": true, "checkLoop": true}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
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
			if !ok || fd.Body == nil || allowed[fd.Name.Name] {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Config" {
					return true
				}
				if inner, ok := sel.X.(*ast.SelectorExpr); ok && inner.Sel.Name == "host" {
					t.Errorf("%s: %s reads the live config; take the pass cfg instead", fset.Position(call.Pos()), fd.Name.Name)
				}
				return true
			})
		}
	}
}
