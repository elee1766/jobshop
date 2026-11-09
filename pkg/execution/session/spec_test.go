package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oldSpec is a spec.json from a controller that still set conv_export.
const oldSpec = `{
  "protocol": 1,
  "session": "js-1",
  "command": "agent",
  "args": ["acp"],
  "env": ["A=1"],
  "env_hash": "h",
  "dir": "/work",
  "lease_ms": 60000,
  "conv_export": {
    "logs_url": "http://otel/v1/logs",
    "pool": "dev",
    "project": "js"
  }
}`

// TestReadOldFormat: conv_export (removed by #530) is ignored and the
// host gets no exporter.
func TestReadOldFormat(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, acphost.SpecFile), []byte(oldSpec), 0o600))

	s, err := Read(dir)
	require.NoError(t, err)
	assert.Equal(t, "js-1", s.Session)
	assert.Equal(t, []string{"acp"}, s.Args)
	assert.Equal(t, "h", s.EnvHash)

	base, err := acphost.ReadSpec(dir)
	require.NoError(t, err)
	assert.Equal(t, s.Spec, base)

	require.NoError(t, Write(dir, s))
	b, err := os.ReadFile(filepath.Join(dir, acphost.SpecFile))
	require.NoError(t, err)
	assert.NotContains(t, string(b), "conv_export")
	assert.Empty(t, HostOptions(dir))
}
