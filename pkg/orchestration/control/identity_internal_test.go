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
	"github.com/elee1766/jobshop/pkg/integrations/identity"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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
	t.Parallel()
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

// TestLinearAppCheck: identity check mints a linear-app token, reads
// viewer with Bearer auth, fails on a bad secret, and never logs the
// token (#278).
func TestLinearAppCheck(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write := func(name, s string) string {
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, []byte(s+"\n"), 0o600))
		return p
	}
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("client_secret") != "good" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"oauth-tok-1","expires_in":86400}`))
	}))
	t.Cleanup(oauth.Close)
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer oauth-tok-1" {
			_, _ = w.Write([]byte(`{"errors":[{"message":"Authentication required"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","name":"elee-agent"}}}`))
	}))
	t.Cleanup(gql.Close)
	yaml := fmt.Sprintf(`
projects: {js: {}}
identities:
  lapp:
    kind: linear-app
    client_id_file: %s
    client_secret_file: %s
    token_url: %s
    workspace: acme
    teams: [ENG]
`, write("cid", "cid"), write("secret", "good"), oauth.URL)
	cfg, err := config.Parse([]byte(yaml))
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	var logs strings.Builder
	c := &Controller{
		live: new(atomic.Pointer[liveConfig]), log: slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
		minter: identity.NewMinter(), linearURL: gql.URL,
	}
	c.live.Store(lc)
	admin := scope{admin: true, actor: "admin"}

	r, err := c.identityCheck(t.Context(), admin, rpc.IdentityParams{Target: "ENG"})
	require.NoError(t, err)
	res := r.(rpc.IdentityCheckResult)
	assert.Equal(t, "elee-agent", res.Viewer)
	assert.Equal(t, config.IdentityLinearApp, res.Kind)
	_, err = c.identityToken(t.Context(), admin, rpc.IdentityParams{Target: "ENG"})
	require.NoError(t, err)
	assert.NotContains(t, logs.String(), "oauth-tok-1", "the token is never logged")

	bad := *cfg
	bad.Identities = map[string]config.Identity{"lapp": cfg.Identities["lapp"]}
	id := bad.Identities["lapp"]
	id.ClientSecretFile = config.SecretPath(write("bad", "nope"))
	bad.Identities["lapp"] = id
	lc, err = newLiveConfig(&bad)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter.Forget()
	_, err = c.identityCheck(t.Context(), admin, rpc.IdentityParams{Target: "ENG"})
	require.ErrorContains(t, err, "invalid_client")
}
