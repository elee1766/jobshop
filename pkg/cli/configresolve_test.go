package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestConfigResolveMigration(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
		return path
	}
	write("pack.yaml", "projects: {js: {}}\nagents: {a: {command: echo}}\npools: {dev: {agent: a, dir: /tmp, max_sessions: 2}}\n")
	write("pack.cue", `package migration
projects: js: {}
agents: a: command: "echo"
pools: dev: {agent: "a", dir: "/tmp", max_sessions: 2}
`)
	legacy := write("legacy.yaml", "pack: pack.yaml\n")
	native := write("native.yaml", "pack: pack.cue\n")
	old := stdout
	t.Cleanup(func() { stdout = old })
	resolve := func(file, format string) map[string]any {
		var output bytes.Buffer
		stdout = &output
		require.Equal(t, 0, Main([]string{"config", "resolve", "--file", file, "--format", format}))
		var result map[string]any
		if format == "json" {
			require.NoError(t, json.Unmarshal(output.Bytes(), &result))
		} else {
			// Normalize numeric types through JSON for a format-independent comparison.
			var parsed map[string]any
			require.NoError(t, yaml.Unmarshal(output.Bytes(), &parsed))
			b, err := json.Marshal(parsed)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(b, &result))
		}
		require.NotContains(t, result, "pack")
		require.NotContains(t, result, "packs")
		return result
	}
	before := resolve(legacy, "json")
	require.Equal(t, before, resolve(native, "json"))
	require.Equal(t, before, resolve(native, "yaml"))
	write("pack.cue", `package migration
projects: js: {}
agents: a: command: "echo"
pools: dev: {agent: "a", dir: "/tmp", max_sessions: 3}
`)
	require.NotEqual(t, before, resolve(native, "json"), "behavior differences must remain visible in migration comparisons")
}
