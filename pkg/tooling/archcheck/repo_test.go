package archcheck_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/tooling/archcheck"
)

func repoGraph(t *testing.T) *archcheck.Graph {
	t.Helper()
	g, err := archcheck.Load(context.Background(), "../../..", "victorialogs")
	require.NoError(t, err)
	return g
}

// TestRepository is the check CI runs, as a test: the real tree must satisfy
// the real rules, with only the dated exceptions in known.go.
func TestRepository(t *testing.T) {
	res := archcheck.Check(repoGraph(t), archcheck.Jobshop())
	assert.Empty(t, res.Problems)
}

// TestRepositoryCatchesDeliberateEdges is the JOB-165 acceptance check. It
// adds prohibited edges to the loaded graph, the way a stray import would
// appear in `go list`, and expects the real rules to name each one.
func TestRepositoryCatchesDeliberateEdges(t *testing.T) {
	const m = "github.com/elee1766/jobshop/"
	cases := []struct {
		name     string
		from, to string
		test     bool
		want     string
	}{
		{"store imports control", "pkg/store", "pkg/control", false, "[control-is-top] " + m + "pkg/store imports " + m + "pkg/control"},
		{"journal test imports control", "pkg/journal", "pkg/control", true, "[control-is-top] " + m + "pkg/journal imports in a test file " + m + "pkg/control"},
		{"hooks imports sessionrt", "pkg/hooks", "pkg/control/sessionrt", false, "[command-paths-no-session] " + m + "pkg/hooks imports " + m + "pkg/control/sessionrt"},
		{"hooks imports hostspec", "pkg/hooks", "pkg/hostspec", false, "[command-paths-no-session] " + m + "pkg/hooks imports " + m + "pkg/hostspec"},
		{"automation imports chatlife", "pkg/automation/command", "pkg/control/chatlife", false, "[command-paths-no-session] " + m + "pkg/automation/command imports " + m + "pkg/control/chatlife"},
		{"memory imports store", "pkg/memory", "pkg/store", false, "[independent-leaves] " + m + "pkg/memory imports " + m + "pkg/store"},
		{"linear imports hooks", "pkg/linear", "pkg/hooks", false, "[family-integrations] " + m + "pkg/linear imports " + m + "pkg/hooks"},
		{"sqlite imports postgres", "pkg/store/sqlite", "pkg/store/postgres", false, "[store-backends-are-siblings] " + m + "pkg/store/sqlite imports " + m + "pkg/store/postgres"},
		{"cli production imports controltest", "pkg/cli", "pkg/control/controltest", false, "[no-harness-in-production] " + m + "pkg/cli imports " + m + "pkg/control/controltest"},
		{"gateway imports starlark", "pkg/gateway", "go.starlark.net/starlark", false, "[starlark-only-in-automation] " + m + "pkg/gateway imports go.starlark.net/starlark"},
		{"clock imports store", "pkg/clock", "pkg/store", false, "[independent-leaves] " + m + "pkg/clock imports " + m + "pkg/store"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := repoGraph(t)
			from := g.Pkgs[m+tc.from]
			require.NotNil(t, from)
			to := tc.to
			if !strings.HasPrefix(to, "go.") {
				to = m + to
			}
			if g.Pkgs[to] == nil {
				g.Pkgs[to] = &archcheck.Package{Path: to}
			}
			if tc.test {
				from.TestImports = append(slices.Clone(from.TestImports), to)
				slices.Sort(from.TestImports)
			} else {
				from.Imports = append(slices.Clone(from.Imports), to)
				slices.Sort(from.Imports)
			}
			res := archcheck.Check(g, archcheck.Jobshop())
			require.NotEmpty(t, res.Problems, "the deliberate edge passed the check")
			assert.Contains(t, res.Problems, tc.want)
		})
	}
}

// TestRepositoryLegitimateTestHarnessEdgesPass pins the test-only side of the
// graph: production code never imports a harness, test files do, and that is
// allowed. If this fails the rules started banning legitimate test setups.
func TestRepositoryLegitimateTestHarnessEdgesPass(t *testing.T) {
	g := repoGraph(t)
	const m = "github.com/elee1766/jobshop/"
	for _, h := range []string{"pkg/store/storetest", "pkg/control/controltest"} {
		var testImporters int
		for _, p := range g.Pkgs {
			if !p.Main {
				continue
			}
			assert.NotContains(t, p.Imports, m+h, "%s imports harness %s in production code", p.Path, h)
			if slices.Contains(p.TestImports, m+h) {
				testImporters++
			}
		}
		assert.Positive(t, testImporters, "%s is used by tests, so the graph has test edges to check", h)
	}
	assert.Empty(t, archcheck.Check(g, archcheck.Jobshop()).Problems)
}

// TestInventoryCoversEveryPackage keeps the ownership table complete:
// every package with Go files has a row, and the inventory lists them all.
func TestInventoryCoversEveryPackage(t *testing.T) {
	g := repoGraph(t)
	cfg := archcheck.Jobshop()
	inv := cfg.Inventory(g)
	var want int
	for _, path := range g.Local() {
		if p := g.Pkgs[path]; p.Files+p.TestFiles > 0 {
			want++
		}
	}
	assert.Len(t, inv, want)
	for _, i := range inv {
		assert.NotEmpty(t, i.Family, i.Pkg)
		assert.NotEmpty(t, i.Child, i.Pkg)
	}
}
