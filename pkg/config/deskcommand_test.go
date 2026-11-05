package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// desk_commands merge per name from packs; with a prefix the names are
// prefixed like pools (#542).
func TestDeskCommandsFromPack(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml": "desk_commands:\n  report:\n    command: ./report.sh\n    timeout: 5s\n",
		"jobshop.yaml": hostBase + `packs:
  - path: p
  - path: p
    prefix: b-
desk_commands:
  report:
    timeout: 10s
`,
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, DeskCommand{Command: "./report.sh", Timeout: "10s"}, c.DeskCommands["report"], "host overrides one field")
	assert.Equal(t, DeskCommand{Command: "./report.sh", Timeout: "5s"}, c.DeskCommands["b-report"])
}

func TestDeskCommandsInvalid(t *testing.T) {
	for _, bad := range []string{
		"desk_commands: {x: {}}",
		"desk_commands: {x: {command: a, timeout: soon}}",
		"desk_commands: {Bad: {command: a}}",
	} {
		_, err := Parse([]byte(hostBase + bad))
		assert.Error(t, err, bad)
	}
}
