package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLastGood: a merged config round-trips through the last good file,
// and only pack errors fall back to it (#227).
func TestLastGood(t *testing.T) {
	dir := t.TempDir()
	pack, err := filepath.Abs("../../packs/selfdev")
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":  "pools:\n  dev: {agent: a, dir: /w, prompt: \"hi ${who}\"}\n",
		"jobshop.yaml": hostBase + "packs:\n  - path: " + pack + "\n    vars: {repo: /src, agent: a}\n",
	})
	r := Root{Dir: dir}
	c, err := Load(r)
	require.NoError(t, err)
	require.NoError(t, SaveLastGood(r.LastGood(), c))
	lg, err := LoadLastGood(r.LastGood())
	require.NoError(t, err)
	assert.Len(t, lg.Pools, len(c.Pools))
	assert.Equal(t, c.Workflows["pr-work"].Steps[0].Prompt, lg.Workflows["pr-work"].Steps[0].Prompt)
	assert.Equal(t, c.Workflows["pr-work"].PackVars, lg.Workflows["pr-work"].PackVars)

	// A pack with a var this binary does not set: last good, plus the error.
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: p\n"})
	got, err := LoadOrLastGood(r)
	require.ErrorIs(t, err, ErrPack)
	assert.ErrorContains(t, err, `var "who" is not set`)
	require.NotNil(t, got)
	assert.Contains(t, got.Pools, "plan", "the saved config")

	// A pack entry the validator rejects is the pack's too.
	writeFiles(t, dir, map[string]string{"p/pack.yaml": "pools:\n  dev: {agent: nosuch, dir: /w}\n"})
	got, err = LoadOrLastGood(r)
	require.ErrorIs(t, err, ErrPack)
	require.NotNil(t, got)

	// Host errors still fail.
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "bogus: 1\n"})
	got, err = LoadOrLastGood(r)
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrPack))
	assert.Nil(t, got)

	// No last good: the pack error fails the start.
	require.NoError(t, os.Remove(r.LastGood()))
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: p\n"})
	got, err = LoadOrLastGood(r)
	require.ErrorIs(t, err, ErrPack)
	assert.Nil(t, got)
}
