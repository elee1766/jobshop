package memory_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/knowledge/memory"
)

func TestTarget(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"global":                 "memory/global.md",
		"repo/gfx-labs__jobshop": "memory/repo/gfx-labs__jobshop.md",
		"pool/dev.md":            "memory/pool/dev.md",
		"skills/rebase":          "skills/rebase/SKILL.md",
	} {
		got, err := memory.Target(in)
		require.NoError(t, err, in)
		assert.Equal(t, filepath.FromSlash(want), got, in)
	}
	for _, bad := range []string{"", "global/x", "repo/", "repo/a/b", "pool/..", "skills/../x", "x"} {
		_, err := memory.Target(bad)
		assert.Error(t, err, bad)
	}
}

func TestWriteRestore(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := filepath.Join("memory", "pool", "dev.md")
	require.NoError(t, memory.Write(dir, rel, "one\n", 0))
	got, ok, err := memory.ReadRaw(dir, rel)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "one\n", got)

	err = memory.Write(dir, rel, strings.Repeat("x", 11), 10)
	require.ErrorIs(t, err, memory.ErrOverCap)
	got, _, _ = memory.ReadRaw(dir, rel)
	assert.Equal(t, "one\n", got, "a refused write leaves the file")

	require.NoError(t, memory.Write(dir, rel, "two\n", 0))
	require.NoError(t, memory.Restore(dir, rel, "one\n", true))
	got, _, _ = memory.ReadRaw(dir, rel)
	assert.Equal(t, "one\n", got)
	require.NoError(t, memory.Restore(dir, rel, "", false))
	_, ok, err = memory.ReadRaw(dir, rel)
	require.NoError(t, err)
	assert.False(t, ok)

	// A skill is not capped; restoring a new skill removes its dir.
	sk := filepath.Join("skills", "s", "SKILL.md")
	require.NoError(t, memory.Write(dir, sk, strings.Repeat("y", 10), 5))
	files, err := memory.List(dir)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, sk, files[0].Path)
	require.NoError(t, memory.Restore(dir, sk, "", false))
	_, err = os.Stat(filepath.Join(dir, "skills", "s"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}
