package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParsePoolRepos (JOB-207): repos keys a pool's sessions by GitHub
// repo, each with its own checkout.
func TestParsePoolRepos(t *testing.T) {
	base := "projects:\n  js: {}\nagents:\n  a:\n    command: x\npools:\n  p:\n    agent: a\n"
	c, err := Parse([]byte(base + "    repo: /src/a\n    ref: origin/master\n    repos:\n      o/b: {checkout: /src/b}\n      o/c: {checkout: /src/c, ref: origin/main}\n"))
	require.NoError(t, err)
	p := c.Pools["p"]
	assert.Equal(t, []string{"o/b", "o/c"}, p.RepoKeys())
	assert.Equal(t, "o/b", p.RepoKey("O/B"), "GitHub names are not case sensitive")
	assert.Empty(t, p.RepoKey("o/a"))
	repo, ref := p.Checkout("o/b")
	assert.Equal(t, [2]string{"/src/b", "origin/master"}, [2]string{repo, ref}, "ref defaults to the pool's")
	repo, ref = p.Checkout("o/c")
	assert.Equal(t, [2]string{"/src/c", "origin/main"}, [2]string{repo, ref})
	repo, ref = p.Checkout("")
	assert.Equal(t, [2]string{"/src/a", "origin/master"}, [2]string{repo, ref})
	repo, _ = p.Checkout("o/gone")
	assert.Equal(t, "/src/a", repo, "a dropped key falls back to the default checkout")

	for _, tc := range []struct{ name, yaml, want string }{
		{"no default repo", "    dir: /w\n    repos:\n      o/b: {checkout: /src/b}\n", "repos needs repo"},
		{"bad key", "    repo: /src/a\n    repos:\n      b: {checkout: /src/b}\n", `key "b" must be owner/name`},
		{"no checkout", "    repo: /src/a\n    repos:\n      o/b: {}\n", "o/b: checkout is required"},
		{"case duplicate", "    repo: /src/a\n    repos:\n      o/b: {checkout: /x}\n      O/B: {checkout: /y}\n", "differ only in case"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(base + tc.yaml))
			require.ErrorContains(t, err, tc.want)
		})
	}
}
