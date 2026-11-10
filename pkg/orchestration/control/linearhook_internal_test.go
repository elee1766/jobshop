package control

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/integrations/identity"
)

// LinearClient picks the Linear identity covering the team, shares one
// client per identity, reads its key per request, and needs team= when
// more than one Linear identity exists.
func TestHookLinearClient(t *testing.T) {
	t.Parallel()
	c, _, dir := scanControllerYAML(t, `projects: {js: {}}
agents: {fake: {command: /bin/false}}
identities:
  eng: {kind: linear, token_file: %[1]s/eng, workspace: acme, teams: [ENG]}
  ops: {kind: linear, token_file: %[1]s/ops, workspace: acme, teams: [OPS]}
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "eng"), []byte("lin_api_eng\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ops"), []byte("lin_api_ops\n"), 0o600))
	var auth []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = append(auth, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u","name":"bot"}}}`))
	}))
	t.Cleanup(srv.Close)
	c.minter, c.linearURL = identity.NewMinter(), srv.URL
	host := hookHost{c}

	ops, err := host.LinearClient(t.Context(), "OPS")
	require.NoError(t, err)
	_, err = ops.Viewer(t.Context())
	require.NoError(t, err)
	eng, err := host.LinearClient(t.Context(), "ENG")
	require.NoError(t, err)
	assert.NotSame(t, ops, eng)
	again, err := host.LinearClient(t.Context(), "ENG")
	require.NoError(t, err)
	assert.Same(t, eng, again, "one client per identity")
	assert.Same(t, eng, c.linearFor("eng"), "the controller's own callers share it")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "eng"), []byte("lin_api_rotated\n"), 0o600))
	_, err = eng.Viewer(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"lin_api_ops", "lin_api_rotated"}, auth, "key read per request")

	_, err = host.LinearClient(t.Context(), "")
	require.ErrorContains(t, err, "pass team=")
	_, err = host.LinearClient(t.Context(), "SALES")
	require.ErrorContains(t, err, "no Linear identity covers team SALES")

	// An identity's api_url selects its endpoint over the controller default.
	var own int
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		own++
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u","name":"bot"}}}`))
	}))
	t.Cleanup(fake.Close)
	cfg := *c.config()
	cfg.Identities = maps.Clone(cfg.Identities)
	id := cfg.Identities["ops"]
	id.APIURL = fake.URL
	cfg.Identities["ops"] = id
	lc := *c.live.Load()
	lc.cfg = &cfg
	c.live.Store(&lc)
	moved, err := host.LinearClient(t.Context(), "OPS")
	require.NoError(t, err)
	assert.NotSame(t, ops, moved, "a changed endpoint replaces the cached client")
	_, err = moved.Viewer(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, own)
	assert.Len(t, auth, 2, "the default endpoint saw no further call")
}
