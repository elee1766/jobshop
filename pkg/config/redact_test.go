package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/pack"
)

// TestRedactRules (#578): a config rule is caught, disabled_rules turns
// one off, and bad rules fail the load naming the rule.
func TestRedactRules(t *testing.T) {
	tok := "acme" + "_" + strings.Repeat("Zq8", 10) + "x9"
	c, err := Parse([]byte(`redact:
  rules:
    - id: acme-token
      description: ACME internal token
      regex: '\bacme_[A-Za-z0-9]{32}\b'
      keywords: [acme_]
  disabled_rules: [generic-api-key]
`))
	require.NoError(t, err)
	sc := c.Redact.Scanner()
	got := sc.Find("x " + tok)
	require.Len(t, got, 1)
	assert.Equal(t, "acme-token", got[0].Rule)
	assert.Empty(t, sc.Find("WEBHOOK_SECRET=Zq8vK2mN7pR4tY6wB1cD3fG5hJ9kL0x"), "generic-api-key disabled")

	plain, err := Parse(nil)
	require.NoError(t, err)
	assert.Empty(t, plain.Redact.Scanner().Find("x "+tok), "default has no acme rule")

	for _, tt := range []struct{ yaml, want string }{
		{"redact:\n  rules:\n    - {id: bad-re, regex: '(', keywords: [x]}\n", "bad-re"},
		{"redact:\n  rules:\n    - {id: github-pat, regex: 'x', keywords: [x]}\n", "github-pat"},
		{"redact:\n  rules:\n    - {id: twice, regex: 'x', keywords: [x]}\n    - {id: twice, regex: 'y', keywords: [y]}\n", "twice"},
		{"redact:\n  disabled_rules: [no-such-rule]\n", "no-such-rule"},
	} {
		_, err := Parse([]byte(tt.yaml))
		require.Error(t, err, tt.yaml)
		assert.Contains(t, err.Error(), tt.want)
	}
}

// TestRedactNotInPacks: redact is host-only, like identities.
func TestRedactNotInPacks(t *testing.T) {
	dir := t.TempDir()
	pk := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pk, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pk, pack.FileName), []byte("redact:\n  disabled_rules: [generic-api-key]\n"), 0o600))
	cfgPath := filepath.Join(dir, FileName)
	require.NoError(t, os.WriteFile(cfgPath, []byte("packs:\n  - path: pk\n"), 0o600))
	_, err := LoadFile(cfgPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redact")
}
