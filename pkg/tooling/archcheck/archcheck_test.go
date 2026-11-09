package archcheck_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/tooling/archcheck"
)

const mod = "example.test/m"

// writeModule makes a small module on disk and loads its graph with the real
// `go list`, so these tests see what the compiler sees, not a hand-built map.
func writeModule(t *testing.T, files map[string]string) *archcheck.Graph {
	t.Helper()
	dir := t.TempDir()
	files["go.mod"] = "module " + mod + "\n\ngo 1.22\n"
	for name, body := range files {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o644))
	}
	g, err := archcheck.Load(context.Background(), dir, "")
	require.NoError(t, err)
	return g
}

func src(pkg string, imports ...string) string {
	var b strings.Builder
	b.WriteString("package " + pkg + "\n")
	for _, i := range imports {
		b.WriteString("import _ \"" + i + "\"\n")
	}
	return b.String()
}

// layout is a tiny tree shaped like the real one: a store, a top-level
// controller, a hook engine, a session runtime and a test harness.
func layout() map[string]string {
	return map[string]string{
		"store/store.go":     src("store"),
		"control/control.go": src("control", mod+"/store", mod+"/hooks"),
		"hooks/hooks.go":     src("hooks", mod+"/store"),
		"session/session.go": src("session", mod+"/store"),
		"ctest/ctest.go":     src("ctest", mod+"/control"),
	}
}

func owners() []archcheck.Owner {
	return []archcheck.Owner{
		{Pkg: "./store", Target: "stay", Family: "low", Child: "-"},
		{Pkg: "./control", Target: "stay", Family: "top", Child: "-"},
		{Pkg: "./hooks", Target: "stay", Family: "mid", Child: "-"},
		{Pkg: "./session", Target: "stay", Family: "mid", Child: "-"},
		{Pkg: "./ctest", Target: "stay", Family: "top", Child: "-", Harness: true},
	}
}

func families() []archcheck.Family {
	return []archcheck.Family{
		{Name: "low"},
		{Name: "mid", May: []string{"low"}},
		{Name: "top", May: []string{"low", "mid"}},
	}
}

func config(rules ...archcheck.Rule) archcheck.Config {
	return archcheck.Config{Owners: owners(), Families: families(), Rules: rules}
}

var controlIsTop = archcheck.Rule{
	Name:   "control-is-top",
	Scope:  archcheck.ProdAndTest,
	From:   []string{"./..."},
	Except: []string{"./control", "./ctest"},
	To:     []string{"./control"},
}

func TestCleanTreePasses(t *testing.T) {
	g := writeModule(t, layout())
	res := archcheck.Check(g, config(controlIsTop))
	assert.Empty(t, res.Problems)
}

func TestProductionEdgeIsCaught(t *testing.T) {
	files := layout()
	files["session/session.go"] = src("session", mod+"/store", mod+"/control")
	res := archcheck.Check(writeModule(t, files), config(controlIsTop))
	require.NotEmpty(t, res.Problems)
	assert.Contains(t, only(res, "control-is-top"), "[control-is-top] "+mod+"/session imports "+mod+"/control")
}

// only keeps the problems of one rule, so a test about one rule does not
// break when a fixture also trips the family check.
func only(res archcheck.Result, rule string) []string {
	var out []string
	for _, p := range res.Problems {
		if strings.HasPrefix(p, "["+rule+"]") {
			out = append(out, p)
		}
	}
	return out
}

func TestTestFileEdgeIsSeparate(t *testing.T) {
	files := layout()
	files["store/store_test.go"] = src("store", mod+"/control")

	// A rule that looks at production code only does not see the test file.
	prodOnly := controlIsTop
	prodOnly.Scope = archcheck.Prod
	assert.Empty(t, archcheck.Check(writeModule(t, files), config(prodOnly)).Problems)

	// A rule that includes tests sees it and says it is a test edge.
	res := archcheck.Check(writeModule(t, files), config(controlIsTop))
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], "imports in a test file")
}

func TestExternalTestPackageIsAnEdge(t *testing.T) {
	files := layout()
	files["store/x_test.go"] = "package store_test\nimport _ \"" + mod + "/control\"\n"
	res := archcheck.Check(writeModule(t, files), config(controlIsTop))
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], mod+"/store imports in a test file "+mod+"/control")
}

func TestHarnessMayImportWhatTestsNeed(t *testing.T) {
	// ctest imports control in layout() and the rule excepts it.
	g := writeModule(t, layout())
	harness := archcheck.Rule{Name: "no-harness", Scope: archcheck.Prod, From: []string{"./..."}, Except: []string{"./ctest"}, To: []string{"./ctest"}}
	assert.Empty(t, archcheck.Check(g, config(controlIsTop, harness)).Problems)

	files := layout()
	files["session/session.go"] = src("session", mod+"/store", mod+"/ctest")
	res := archcheck.Check(writeModule(t, files), config(controlIsTop, harness))
	require.Len(t, only(res, "no-harness"), 1)
	assert.Contains(t, only(res, "no-harness")[0], "[no-harness] "+mod+"/session imports "+mod+"/ctest")
}

func TestTestOnlyScopeIgnoresProductionImports(t *testing.T) {
	files := layout()
	files["session/session.go"] = src("session", mod+"/store", mod+"/ctest")
	files["hooks/hooks_test.go"] = src("hooks", mod+"/ctest")
	rule := archcheck.Rule{Name: "t", Scope: archcheck.TestOnly, From: []string{"./..."}, To: []string{"./ctest"}}
	res := archcheck.Check(writeModule(t, files), config(rule))
	require.Len(t, only(res, "t"), 1)
	assert.Contains(t, only(res, "t")[0], mod+"/hooks imports in a test file")
}

func TestClosureFindsTransitiveEdgeAndNamesTheChain(t *testing.T) {
	files := layout()
	// hooks -> util -> session: no direct import, still a dependency.
	files["util/util.go"] = src("util", mod+"/session")
	files["hooks/hooks.go"] = src("hooks", mod+"/store", mod+"/util")
	cfg := config(archcheck.Rule{Name: "no-session", Scope: archcheck.Closure, From: []string{"./hooks"}, To: []string{"./session"}})
	cfg.Owners = append(cfg.Owners, archcheck.Owner{Pkg: "./util", Target: "stay", Family: "mid", Child: "-"})
	res := archcheck.Check(writeModule(t, files), cfg)
	require.Len(t, res.Problems, 1)
	assert.Equal(t, "[no-session] "+mod+"/hooks imports "+mod+"/session (via "+mod+"/util)", res.Problems[0])
}

func TestClosureReachesStandardLibrary(t *testing.T) {
	files := layout()
	files["store/store.go"] = src("store", "net/http")
	rule := archcheck.Rule{Name: "no-std", Scope: archcheck.Closure, From: []string{"./hooks"}, To: []string{"std"}}
	res := archcheck.Check(writeModule(t, files), config(rule))
	require.NotEmpty(t, only(res, "no-std"))
	assert.Contains(t, strings.Join(only(res, "no-std"), "\n"), "[no-std] "+mod+"/hooks imports net/http (via "+mod+"/store)")
}

func TestFamilyDirection(t *testing.T) {
	files := layout()
	// util is a low-family package that reaches up to a mid-family one.
	files["util/util.go"] = src("util", mod+"/session")
	cfg := config()
	cfg.Owners = append(cfg.Owners, archcheck.Owner{Pkg: "./util", Target: "stay", Family: "low", Child: "-"})
	res := archcheck.Check(writeModule(t, files), cfg)
	require.Len(t, res.Problems, 1)
	assert.Equal(t, "[family-low] "+mod+"/util imports "+mod+"/session", res.Problems[0])
}

func TestKnownExceptionCoversOneEdgeOnly(t *testing.T) {
	files := layout()
	files["session/session.go"] = src("session", mod+"/store", mod+"/control")
	files["util/util.go"] = src("util", mod+"/control")
	cfg := config(controlIsTop)
	cfg.Owners = append(cfg.Owners, archcheck.Owner{Pkg: "./util", Target: "stay", Family: "top", Child: "-"})
	cfg.Known = []archcheck.Known{{Rule: "control-is-top", From: "./session", To: "./control", Issue: "JOB-1", Why: "temporary"}}
	res := archcheck.Check(writeModule(t, files), cfg)
	assert.Len(t, res.Accepted, 1)
	require.Len(t, only(res, "control-is-top"), 1, "the util edge is a new violation and stays an error")
	assert.Contains(t, res.Problems[0], mod+"/util imports")
}

func TestStaleExceptionFails(t *testing.T) {
	cfg := config(controlIsTop)
	cfg.Known = []archcheck.Known{{Rule: "control-is-top", From: "./session", To: "./control", Issue: "JOB-1", Why: "temporary"}}
	res := archcheck.Check(writeModule(t, layout()), cfg)
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], "stale exception")
}

func TestExceptionNeedsIssueAndReason(t *testing.T) {
	files := layout()
	files["util/util.go"] = src("util", mod+"/control")
	cfg := config(controlIsTop)
	cfg.Owners = append(cfg.Owners, archcheck.Owner{Pkg: "./util", Target: "stay", Family: "top", Child: "-"})
	cfg.Known = []archcheck.Known{{Rule: "control-is-top", From: "./util", To: "./control"}}
	res := archcheck.Check(writeModule(t, files), cfg)
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], "needs an issue and a reason")
}

func TestRuleWithDeadPatternFails(t *testing.T) {
	rule := archcheck.Rule{Name: "r", Scope: archcheck.Prod, From: []string{"./nothing"}, To: []string{"./control"}}
	res := archcheck.Check(writeModule(t, layout()), config(rule))
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], `From "./nothing" matches no package`)
}

func TestNewPackageNeedsAnOwner(t *testing.T) {
	files := layout()
	files["extra/extra.go"] = src("extra")
	res := archcheck.Check(writeModule(t, files), config(controlIsTop))
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], "extra has no row in the ownership table")
}

func TestOwnerRowForMissingPackageFails(t *testing.T) {
	cfg := config(controlIsTop)
	cfg.Owners = append(cfg.Owners, archcheck.Owner{Pkg: "./gone", Target: "stay", Family: "low", Child: "-"})
	res := archcheck.Check(writeModule(t, layout()), cfg)
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], "./gone matches no package")
}

func TestOnlyRuleCountsTestImports(t *testing.T) {
	files := layout()
	files["store/store_test.go"] = src("store", mod+"/session")
	cfg := config()
	cfg.Only = []archcheck.Only{{Name: "leaf", From: []string{"./store"}, Std: true}}
	res := archcheck.Check(writeModule(t, files), cfg)
	require.Len(t, res.Problems, 1)
	assert.Contains(t, res.Problems[0], "[leaf] "+mod+"/store imports in a test file "+mod+"/session")
}

func TestMatch(t *testing.T) {
	g := &archcheck.Graph{Module: mod}
	assert.True(t, g.Match("./a", mod+"/a"))
	assert.False(t, g.Match("./a", mod+"/a/b"))
	assert.True(t, g.Match("./a/...", mod+"/a/b"))
	assert.False(t, g.Match("./a/...", mod+"/a"), "x/... does not match x")
	assert.False(t, g.Match("./a/...", mod+"/ab"))
	assert.True(t, g.Match("example.org/x/...", "example.org/x/y"))
}

func TestBrokenBuildIsNotSilentlyIgnored(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+mod+"\n\ngo 1.22\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte(src("a", "example.test/missing/pkg")), 0o644))
	_, err := archcheck.Load(context.Background(), dir, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broken packages")
}
