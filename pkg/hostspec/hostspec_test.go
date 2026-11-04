package hostspec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/elee1766/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oldSpec is a spec.json as pkg/host wrote it before acphost.
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
    "headers": {"Authorization": "Bearer x"},
    "pool": "dev",
    "project": "js",
    "traceparent": "00-aa-bb-01"
  }
}`

func TestReadOldFormat(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, acphost.SpecFile), []byte(oldSpec), 0o600))

	s, err := Read(dir)
	require.NoError(t, err)
	assert.Equal(t, "js-1", s.Session)
	assert.Equal(t, []string{"acp"}, s.Args)
	assert.Equal(t, "h", s.EnvHash)
	assert.Empty(t, s.MethodPrefix, "default prefix")
	require.NotNil(t, s.ConvExport)
	assert.Equal(t, "dev", s.ConvExport.Pool)
	assert.Equal(t, "Bearer x", s.ConvExport.Headers["Authorization"])

	// acphost reads the same file, ignoring conv_export.
	base, err := acphost.ReadSpec(dir)
	require.NoError(t, err)
	assert.Equal(t, s.Spec, base)

	// Writing it back keeps the old layout.
	require.NoError(t, Write(dir, s))
	var got, want map[string]any
	b, err := os.ReadFile(filepath.Join(dir, acphost.SpecFile))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &got))
	require.NoError(t, json.Unmarshal([]byte(oldSpec), &want))
	assert.Equal(t, want, got)

	assert.Len(t, HostOptions(dir), 1, "exporter configured")
}

func TestHostOptionsNoExport(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Write(dir, Spec{Spec: acphost.Spec{Protocol: acphost.Protocol, Session: "s"}}))
	assert.Empty(t, HostOptions(dir))
	assert.Empty(t, HostOptions(t.TempDir()), "missing spec")
}
