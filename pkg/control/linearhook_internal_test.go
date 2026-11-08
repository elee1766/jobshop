package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/identity"
)

// LinearToken picks the Linear identity covering the team, reads its key
// per call, and needs team= when more than one Linear identity exists.
func TestHookLinearToken(t *testing.T) {
	c, _, dir := scanControllerYAML(t, `projects: {js: {}}
agents: {fake: {command: /bin/false}}
identities:
  eng: {kind: linear, token_file: %[1]s/eng, workspace: acme, teams: [ENG]}
  ops: {kind: linear, token_file: %[1]s/ops, workspace: acme, teams: [OPS]}
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "eng"), []byte("lin_api_eng\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ops"), []byte("lin_api_ops\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), "http://linear.test/graphql"
	host := hookHost{c}

	tok, err := host.LinearToken(t.Context(), "OPS")
	require.NoError(t, err)
	assert.Equal(t, "lin_api_ops", tok.Token)
	assert.Equal(t, "http://linear.test/graphql", tok.APIURL)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "eng"), []byte("lin_api_rotated\n"), 0o600))
	tok, err = host.LinearToken(t.Context(), "ENG")
	require.NoError(t, err)
	assert.Equal(t, "lin_api_rotated", tok.Token, "key read per call")

	_, err = host.LinearToken(t.Context(), "")
	require.ErrorContains(t, err, "pass team=")
	_, err = host.LinearToken(t.Context(), "SALES")
	require.ErrorContains(t, err, "no Linear identity covers team SALES")
}
