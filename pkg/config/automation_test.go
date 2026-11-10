package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/pack"
)

func TestAutomationConfig(t *testing.T) {
	_, err := Parse([]byte(`automation:
  http: true
  shell: true
  timeout: 30s
  credentials:
    github:
      origins: [https://api.github.com]
      header: Authorization
      prefix: 'Bearer '
      secret: /run/secrets/github
jev:
  enabled: true
  endpoint: https://jev.example.com/v1/chat/completions
  api_key: /run/secrets/jev
`))
	require.NoError(t, err, "validation must not read secret files or invoke external capabilities")

	for _, tc := range []struct {
		name, input, message string
	}{
		{"timeout", "automation: {timeout: eventually}", "automation.timeout"},
		{"negative output", "automation: {output_bytes: -1}", "automation.output_bytes"},
		{"missing Jev endpoint", "jev: {enabled: true, api_key: /not/read}", "jev.endpoint"},
		{"missing Jev key", "jev: {enabled: true, endpoint: https://jev.example.com}", "jev.api_key"},
		{"credential URL userinfo", "automation: {credentials: {github: {origins: ['https://user:password@api.github.com'], header: Authorization, secret: /not/read}}}", "origins"},
		{"credential URL path", "automation: {credentials: {github: {origins: ['https://api.github.com/repos'], header: Authorization, secret: /not/read}}}", "origins"},
		{"invalid header", "automation: {credentials: {github: {origins: ['https://api.github.com'], header: 'bad header', secret: /not/read}}}", "header"},
		{"missing secret", "automation: {credentials: {github: {origins: ['https://api.github.com'], header: Authorization}}}", "secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.input))
			require.ErrorContains(t, err, tc.message)
		})
	}
}

func TestAutomationCapabilitiesAreHostOnly(t *testing.T) {
	for _, key := range []string{"automation", "jev"} {
		t.Run(key, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte("pack: pack.yaml\n"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, pack.FileName), []byte(key+": {enabled: true}\n"), 0o600))
			_, err := LoadFile(filepath.Join(dir, FileName))
			require.ErrorContains(t, err, key)
		})
	}
}
