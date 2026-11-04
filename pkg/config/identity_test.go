package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

func TestIdentityConfig(t *testing.T) {
	ok := `
projects:
  js: {identities: [gh]}
identities:
  gh: {kind: github-app, app_id: 1, private_key_file: /k, installations: {elee1766: 5}, repos: ["elee1766/*"]}
  lin: {kind: linear, token_file: /t, workspace: w, teams: [ENG]}
`
	c, err := config.Parse([]byte(ok))
	require.NoError(t, err)
	n, found := c.IdentityFor([]string{"gh", "lin"}, "elee1766/jobshop")
	assert.True(t, found)
	assert.Equal(t, "gh", n)
	n, _ = c.IdentityFor([]string{"gh", "lin"}, "ENG")
	assert.Equal(t, "lin", n)
	_, found = c.IdentityFor([]string{"gh"}, "other/jobshop")
	assert.False(t, found)

	for name, bad := range map[string]string{
		"unknown kind":       "identities: {x: {kind: gitlab}}",
		"no installations":   "identities: {x: {kind: github-app, app_id: 1, private_key_file: /k, repos: [a/b]}}",
		"inline token":       "identities: {x: {kind: linear, token: abc}}",
		"mixed fields":       "identities: {x: {kind: linear, token_file: /t, workspace: w, teams: [A], app_id: 3}}",
		"unknown in project": "projects: {js: {identities: [nope]}}",
	} {
		_, err := config.Parse([]byte(bad))
		assert.Error(t, err, name)
	}
}

func TestLinearAppConfig(t *testing.T) {
	c, err := config.Parse([]byte(`identities: {la: {kind: linear-app, client_id_file: /i, client_secret_file: /s, workspace: w, teams: [ENG]}}`))
	require.NoError(t, err)
	n, ok := c.IdentityFor(c.IdentityNames(), "ENG")
	assert.True(t, ok)
	assert.Equal(t, "la", n)
	assert.Equal(t, config.DefaultLinearAppScopes, c.Identities["la"].LinearScopes())
	assert.Equal(t, []string{"ENG"}, c.ActivityTeams())
	for name, bad := range map[string]string{
		"no secret":    "identities: {x: {kind: linear-app, client_id_file: /i, workspace: w, teams: [A]}}",
		"token_file":   "identities: {x: {kind: linear-app, client_id_file: /i, client_secret_file: /s, token_file: /t, workspace: w, teams: [A]}}",
		"comma scopes": "identities: {x: {kind: linear-app, client_id_file: /i, client_secret_file: /s, workspace: w, teams: [A], scopes: [\"read,write\"]}}",
		"app on key":   "identities: {x: {kind: linear, token_file: /t, workspace: w, teams: [A], client_id_file: /i}}",
	} {
		_, err := config.Parse([]byte(bad))
		assert.Error(t, err, name)
	}
}
