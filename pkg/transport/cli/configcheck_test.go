package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathRewrite(t *testing.T) {
	rw, err := pathRewrite([]string{"/src/jobshop=/tmp/new"})
	require.NoError(t, err)
	assert.Equal(t, "/tmp/new/packs/operator", rw("/src/jobshop/packs/operator"))
	assert.Equal(t, "/tmp/new", rw("/src/jobshop"))
	assert.Equal(t, "/src/jobshop2/packs/x", rw("/src/jobshop2/packs/x"), "prefix is a path, not a string")
	assert.Equal(t, "packs/x", rw("packs/x"))
	_, err = pathRewrite([]string{"nope"})
	require.Error(t, err)
}

// A pack var removed on the new commit fails the check while the old
// checkout still loads (#261).
func TestConfigCheck(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	}
	write("old/op/pack.yaml", "pools:\n  op:\n    agent: a\n    repo: \"{{.repo}}\"\n")
	write("new/op/pack.yaml", "pools:\n  op:\n    agent: a\n    repo: \"{{.jobshop_repo}}\"\n")
	cfg := filepath.Join(dir, "jobshop.yaml")
	write("jobshop.yaml", "projects:\n  js: {}\nagents:\n  a:\n    command: x\npacks:\n  - path: "+filepath.Join(dir, "old/op")+"\n    vars: {repo: /src}\n")

	assert.Equal(t, 0, Main([]string{"config", "check", "--file", cfg}, nil))
	assert.Equal(t, 1, Main([]string{"config", "check", "--file", cfg, "--replace", filepath.Join(dir, "old") + "=" + filepath.Join(dir, "new")}, nil))
}
