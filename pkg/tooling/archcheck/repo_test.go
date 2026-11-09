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
		{"store imports control", "pkg/storage/store", "pkg/control", false, "[control-is-top] " + m + "pkg/storage/store imports " + m + "pkg/control"},
		{"journal test imports control", "pkg/storage/journal", "pkg/control", true, "[control-is-top] " + m + "pkg/storage/journal imports in a test file " + m + "pkg/control"},
		{"hooks imports sessionrt", "pkg/hooks", "pkg/control/sessionrt", false, "[command-paths-no-session] " + m + "pkg/hooks imports " + m + "pkg/control/sessionrt"},
		{"hooks imports hostspec", "pkg/hooks", "pkg/hostspec", false, "[command-paths-no-session] " + m + "pkg/hooks imports " + m + "pkg/hostspec"},
		{"automation imports conversation", "pkg/automation/command", "pkg/conversation", false, "[command-paths-no-session] " + m + "pkg/automation/command imports " + m + "pkg/conversation"},
		{"memory imports store", "pkg/knowledge/memory", "pkg/storage/store", false, "[independent-leaves] " + m + "pkg/knowledge/memory imports " + m + "pkg/storage/store"},
		{"memory imports knowledge", "pkg/knowledge/memory", "pkg/knowledge", false, "[independent-leaves] " + m + "pkg/knowledge/memory imports " + m + "pkg/knowledge"},
		{"linear imports hooks", "pkg/integrations/linear", "pkg/hooks", false, "[family-integrations] " + m + "pkg/integrations/linear imports " + m + "pkg/hooks"},
		{"linear imports github", "pkg/integrations/linear", "pkg/integrations/github", false, "[provider-clients-are-siblings-reverse] " + m + "pkg/integrations/linear imports " + m + "pkg/integrations/github"},
		{"github imports identity", "pkg/integrations/github", "pkg/integrations/identity", false, "[provider-clients-stay-small] " + m + "pkg/integrations/github imports " + m + "pkg/integrations/identity"},
		{"hooks imports identity", "pkg/hooks", "pkg/integrations/identity", false, "[identity-only-in-control] " + m + "pkg/hooks imports " + m + "pkg/integrations/identity"},
		{"sqlite imports postgres", "pkg/storage/store/sqlite", "pkg/storage/store/postgres", false, "[store-backends-are-siblings] " + m + "pkg/storage/store/sqlite imports " + m + "pkg/storage/store/postgres"},
		{"cli production imports controltest", "pkg/cli", "pkg/control/controltest", false, "[no-harness-in-production] " + m + "pkg/cli imports " + m + "pkg/control/controltest"},
		{"gateway imports starlark", "pkg/gateway", "go.starlark.net/starlark", false, "[starlark-only-in-automation] " + m + "pkg/gateway imports go.starlark.net/starlark"},
		{"clock imports store", "pkg/clock", "pkg/storage/store", false, "[independent-leaves] " + m + "pkg/clock imports " + m + "pkg/storage/store"},
		{"domain imports store", "pkg/domain", "pkg/storage/store", false, "[domain-is-a-leaf] " + m + "pkg/domain imports " + m + "pkg/storage/store"},
		{"domain test imports rpc", "pkg/domain", "pkg/rpc", true, "[domain-is-a-leaf] " + m + "pkg/domain imports in a test file " + m + "pkg/rpc"},
		{"domain imports a third party", "pkg/domain", "github.com/stretchr/testify/assert", false, "[family-domain] " + m + "pkg/domain imports github.com/stretchr/testify/assert"},
		{"config imports rpc", "pkg/config", "pkg/rpc", false, "[types-not-from-store-or-rpc] " + m + "pkg/config imports " + m + "pkg/rpc"},
		{"question imports store", "pkg/conversation/question", "pkg/storage/store", false, "[types-not-from-store-or-rpc] " + m + "pkg/conversation/question imports " + m + "pkg/storage/store"},
		{"hooks imports rpc", "pkg/hooks", "pkg/rpc", false, "[hooks-not-from-rpc] " + m + "pkg/hooks imports " + m + "pkg/rpc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := repoGraph(t)
			from := g.Pkgs[m+tc.from]
			require.NotNil(t, from)
			to := tc.to
			if !strings.HasPrefix(to, "go.") && !strings.HasPrefix(to, "github.com/stretchr") {
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
	for _, h := range []string{"pkg/storage/store/storetest", "pkg/control/controltest"} {
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
