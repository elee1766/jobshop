package control

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// TestConfigWatchRedactRules (#578): a redact.rules change in the watched
// config takes effect without a restart, in redaction and in the CLI's
// secret.scan; a bad rule keeps the last good scanner.
func TestConfigWatchRedactRules(t *testing.T) {
	t.Parallel()
	c, clk, dir, res := watchController(t)
	tok := "acme" + "_" + strings.Repeat("Zq8", 10) + "x9"
	scan := func() []rpc.SecretFinding {
		t.Helper()
		r, err := c.secretScan(context.Background(), scope{}, rpc.SecretScanParams{Texts: []rpc.SecretText{{From: "body", Text: "x " + tok}}})
		require.NoError(t, err)
		return r.(rpc.SecretScanResult).Findings
	}
	require.Empty(t, scan())
	assert.Contains(t, c.redactor(nil).Text("x "+tok), tok)

	cfgFile := filepath.Join(dir, "jobshop.yaml")
	base := "projects:\n  js: {}\nagents:\n  a: {command: x}\nreload: {watch: true, debounce: 2s}\npacks:\n  - path: p\n"
	rule := "redact:\n  rules:\n    - {id: acme-token, regex: '\\bacme_[A-Za-z0-9]{32}\\b', keywords: [acme_]}\n"
	require.NoError(t, os.WriteFile(cfgFile, []byte(base+rule), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	got := scan()
	require.Len(t, got, 1)
	assert.Equal(t, "acme-token", got[0].Rule)
	assert.Equal(t, "body", got[0].From)
	assert.NotContains(t, c.redactor(nil).Text("x "+tok), tok)

	bad := "redact:\n  rules:\n    - {id: acme-token, regex: '(', keywords: [acme_]}\n"
	require.NoError(t, os.WriteFile(cfgFile, []byte(base+bad), 0o600))
	require.Error(t, fireDebounce(t, clk, res).err)
	assert.Len(t, scan(), 1, "the last good scanner stays")
}
