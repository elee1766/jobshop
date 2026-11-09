package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestIngestPut: `ingest put` spools a generic delivery, prints
// duplicate for a repeat id (exit 0 both times), refuses unknown
// endpoints, and fails when the spool cannot be written (#399).
func TestIngestPut(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(`projects:
  js: {}
ingest:
  endpoints:
    - {name: hints, source: generic, secret_file: s, signature_header: X-Sig, map: {kind: event}}
    - {name: gh, source: github, secret_file: s}
`), 0o600))
	var buf bytes.Buffer
	oldOut, oldIn := stdout, stdin
	stdout = &buf
	t.Cleanup(func() { stdout, stdin = oldOut, oldIn })
	g := &Globals{Root: dir}
	put := func(c ingestPutCmd) error {
		stdin = strings.NewReader(`{"event":"push"}`)
		return c.Run(g)
	}

	require.NoError(t, put(ingestPutCmd{Endpoint: "hints", DeliveryID: "d1"}))
	require.NoError(t, put(ingestPutCmd{Endpoint: "hints", DeliveryID: "d1"}))
	assert.Equal(t, "spooled\nduplicate\n", buf.String())
	require.ErrorContains(t, put(ingestPutCmd{Endpoint: "nope"}), "no endpoint")
	require.ErrorContains(t, put(ingestPutCmd{Endpoint: "gh"}), "generic endpoints")

	// A spool path that is a directory cannot be opened.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bad.db"), 0o700))
	require.Error(t, put(ingestPutCmd{Endpoint: "hints", Spool: "bad.db"}))
}
