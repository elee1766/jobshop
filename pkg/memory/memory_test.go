package memory_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/memory"
)

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for p, s := range files {
		full := filepath.Join(dir, p)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(s), 0o600))
	}
}

func TestScopes(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"memory/global.md":                 "G",
		"memory/repo/elee1766__jobshop.md": "R",
		"memory/repo/other__x.md":          "OTHER",
		"memory/pool/dev.md":               "P",
		"memory/pool/plan.md":              "PLAN",
	})
	files, err := memory.Files(dir, memory.Scope{Repo: "elee1766/jobshop", Pool: "dev"}, 0)
	require.NoError(t, err)
	var got []string
	for _, f := range files {
		got = append(got, f.Scope+"="+f.Text)
	}
	assert.Equal(t, []string{"global=G", "repo=R", "pool=P"}, got)

	files, err = memory.Files(dir, memory.Scope{Pool: "nope"}, 0)
	require.NoError(t, err)
	require.Len(t, files, 1, "no repo, unknown pool: global only")

	files, err = memory.Files(dir, memory.Scope{Pool: "../memory/global"}, 0)
	require.NoError(t, err)
	assert.Len(t, files, 1, "a pool name cannot escape the pool dir")
}

func TestCap(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"memory/global.md":   strings.Repeat("é", 30),
		"memory/pool/dev.md": "short",
	})
	out, over, err := memory.Render(dir, memory.Scope{Pool: "dev"}, 10)
	require.NoError(t, err)
	require.Len(t, over, 1)
	assert.Equal(t, 30, over[0].Chars)
	assert.Contains(t, out, "Memory (global, memory/global.md):\n"+strings.Repeat("é", 10)+"\n\n")
	assert.NotContains(t, out, strings.Repeat("é", 11))
	assert.Contains(t, out, "short")

	bad, err := memory.Check(dir, 10)
	require.NoError(t, err)
	require.Len(t, bad, 1)
	assert.Equal(t, filepath.Join("memory", "global.md"), bad[0].Path)
	bad, err = memory.Check(dir, 0)
	require.NoError(t, err)
	assert.Empty(t, bad, "default cap 2200")
}

func TestSkillsIndex(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"skills/zeta/SKILL.md":   "---\nname: zeta-skill\ndescription: \"Do zeta things\"\n---\nbody",
		"skills/alpha/SKILL.md":  "---\nname: alpha\ndescription: Read logs\n---\n",
		"skills/noname/SKILL.md": "no frontmatter",
		"skills/empty/README":    "not a skill",
	})
	out, _, err := memory.Render(dir, memory.Scope{}, 0)
	require.NoError(t, err)
	assert.Equal(t, "Skills (read the file when one applies):\n"+
		"- alpha: Read logs ("+filepath.Join(dir, "skills/alpha/SKILL.md")+")\n"+
		"- noname ("+filepath.Join(dir, "skills/noname/SKILL.md")+")\n"+
		"- zeta-skill: Do zeta things ("+filepath.Join(dir, "skills/zeta/SKILL.md")+")", out)
}

func TestMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "absent")
	out, over, err := memory.Render(dir, memory.Scope{Repo: "a/b", Pool: "dev"}, 0)
	require.NoError(t, err)
	assert.Empty(t, out)
	assert.Empty(t, over)
	bad, err := memory.Check(dir, 0)
	require.NoError(t, err)
	assert.Empty(t, bad)
}
