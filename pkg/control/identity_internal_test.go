package control

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
)

const identityYAML = `
projects:
  js: {identities: [elee]}
  other: {}
agents:
  a: {command: x}
pools:
  dev: {project: js, agent: a, dir: /tmp}
  linear: {project: js, agent: a, dir: /tmp, identities: [lin]}
identities:
  elee:
    kind: github-app
    app_id: 1
    private_key_file: %[1]s/key.pem
    api_url: %[2]s
    installations: {elee1766: 11}
    repos: ["elee1766/*"]
  second:
    kind: github-app
    app_id: 2
    private_key_file: %[1]s/key.pem
    api_url: %[2]s
    installations: {acme: 22}
    repos: [acme/site]
  lin:
    kind: linear
    token_file: %[1]s/linear
    workspace: acme
    teams: [ENG]
`

// TestIdentityToken: sessions get tokens only through their project and
// pool identities, each repo gets its own App's installation token, and
// an uncovered target is refused as forbidden (#200).
func TestIdentityToken(t *testing.T) {
	dir := t.TempDir()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "linear"), []byte("lin_api_1\n"), 0o600))
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		inst := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/app/installations/"), "/access_tokens")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"token":"ghs-%s","expires_at":"2099-01-01T00:00:00Z"}`, inst)
	}))
	t.Cleanup(srv.Close)
	cfg, err := config.Parse(fmt.Appendf(nil, identityYAML, dir, srv.URL))
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	c := &Controller{live: new(atomic.Pointer[liveConfig]), log: slog.New(slog.DiscardHandler), minter: identity.NewMinter()}
	c.live.Store(lc)
	ctx := t.Context()

	dev := scope{session: "s1", project: "js", pool: "dev"}
	r, err := c.identityToken(ctx, dev, rpc.IdentityParams{Target: "elee1766/jobshop"})
	require.NoError(t, err)
	assert.Equal(t, rpc.IdentityTokenResult{Identity: "elee", Kind: config.IdentityGitHubApp, Token: "ghs-11", Expires: r.(rpc.IdentityTokenResult).Expires}, r)

	for _, tc := range []struct {
		sc     scope
		target string
	}{
		{dev, "acme/site"}, // covered, but not this project's identity
		{dev, "ENG"},       // linear only for the linear pool
		{scope{session: "s2", project: "other"}, "elee1766/jobshop"}, // project with no identities
		{dev, "someone/else"},
	} {
		_, err := c.identityToken(ctx, tc.sc, rpc.IdentityParams{Target: tc.target})
		require.Error(t, err, tc.target)
		assert.True(t, errors.Is(err, rpc.ErrForbidden), tc.target)
		assert.Contains(t, err.Error(), "no identity covers "+tc.target)
	}

	lin := scope{session: "s3", project: "js", pool: "linear"}
	r, err = c.identityToken(ctx, lin, rpc.IdentityParams{Target: "ENG"})
	require.NoError(t, err)
	assert.Equal(t, "lin_api_1", r.(rpc.IdentityTokenResult).Token)

	admin := scope{admin: true}
	r, err = c.identityCheck(ctx, admin, rpc.IdentityParams{Target: "acme/site"})
	require.NoError(t, err)
	assert.Equal(t, "second", r.(rpc.IdentityCheckResult).Identity)
	_, err = c.identityCheck(ctx, dev, rpc.IdentityParams{Target: "acme/site"})
	assert.True(t, errors.Is(err, rpc.ErrForbidden), "check is admin only")

	l, err := c.identityList(ctx, admin, struct{}{})
	require.NoError(t, err)
	infos := l.([]rpc.IdentityInfo)
	require.Len(t, infos, 3)
	assert.Equal(t, []string{"js"}, infos[0].Projects)
	assert.Equal(t, []string{"linear"}, infos[1].Pools)
	assert.Equal(t, int32(2), calls.Load(), "one mint per installation")
}
