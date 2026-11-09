package cue_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/pack/cue"
)

func TestLoadLocalReplacementSources(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"cue.mod/module.cue": `module: "example.com/main@v0"
language: version: "v0.17.0"
deps: "example.com/dep@v0": v: "v0.1.0"
`,
		"cue.mod/local-module.cue": `deps: "example.com/dep@v0": replaceWith: "./local_dep"
`,
		"pack.cue": `package main
import "example.com/dep@v0:lib"
records: "with.dot": lib.#Record
items: [lib.#Record]
`,
		"sibling.cue": `package main
records: "with.dot": note: "sibling"
items: [{note: "list sibling"}]
`,
		"local_dep/cue.mod/module.cue": `module: "example.com/dep@v0"
language: version: "v0.17.0"
`,
		"local_dep/lib.cue": `package lib
#Record: {value: *"local default" | string, note: string}
`,
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(data), 0o600))
	}

	result, err := cue.Load(filepath.Join(dir, "pack.cue"), dir)
	require.NoError(t, err)
	records := result.Data["records"].(map[string]any)
	assert.Equal(t, map[string]any{"value": "local default", "note": "sibling"}, records["with.dot"])
	assert.Equal(t, []any{map[string]any{"value": "local default", "note": "list sibling"}}, result.Data["items"])

	imported := filepath.Join(dir, "local_dep/lib.cue")
	sibling := filepath.Join(dir, "sibling.cue")
	assert.Contains(t, result.Sources("records", "with.dot"), imported)
	assert.Contains(t, result.Sources("records", "with.dot"), sibling)
	assert.Contains(t, result.Sources("items", "0"), imported)
	assert.Contains(t, result.Sources("items", "0"), sibling)
	assert.Nil(t, result.Sources("records.with.dot"))
	assert.Nil(t, result.Sources("items", "1"))

	for _, name := range []string{
		"pack.cue", "sibling.cue", "cue.mod/module.cue", "cue.mod/local-module.cue",
		"local_dep/lib.cue",
	} {
		assert.Contains(t, result.Files, filepath.Join(dir, name))
	}
	assert.ElementsMatch(t, []string{dir, filepath.Join(dir, "local_dep")}, result.Dirs)
}
