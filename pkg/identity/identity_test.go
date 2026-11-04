package identity_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
)

// FakeGitHub serves POST /app/installations/<id>/access_tokens with a
// token naming the installation and a count, and checks the JWT shape.
func fakeGitHub(t *testing.T, ttl time.Duration, now func() time.Time) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if r.Method != http.MethodPost || !strings.HasPrefix(auth, "Bearer ") || strings.Count(auth, ".") != 2 {
			http.Error(w, "bad request", http.StatusUnauthorized)
			return
		}
		inst := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/app/installations/"), "/access_tokens")
		c := n.Add(1)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"tok-` + inst + `-` + strconv.Itoa(int(c)) + `","expires_at":"` + now().Add(ttl).UTC().Format(time.RFC3339) + `"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

// WriteKey writes a fresh RSA key in PKCS1 PEM and returns its path.
func WriteKey(t *testing.T) string {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	p := filepath.Join(t.TempDir(), "app.pem")
	require.NoError(t, os.WriteFile(p, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}), 0o600))
	return p
}

// TestMinterCachesAndRefreshes: a token is reused until 10m before
// expiry, then minted again; installations are cached separately.
func TestMinterCachesAndRefreshes(t *testing.T) {
	now := time.Date(2026, 11, 3, 6, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	srv, calls := fakeGitHub(t, time.Hour, clock)
	id := config.Identity{
		Kind: config.IdentityGitHubApp, AppID: 1, PrivateKeyFile: WriteKey(t), APIURL: srv.URL,
		Installations: map[string]int64{"a": 11, "b": 22}, Repos: []string{"a/*", "b/x"},
	}
	m := identity.NewMinter()
	m.Now = clock
	ctx := t.Context()

	t1, err := m.Token(ctx, "app", id, "a/one")
	require.NoError(t, err)
	assert.Equal(t, "tok-11-1", t1.Token)
	t2, err := m.Token(ctx, "app", id, "a/two")
	require.NoError(t, err)
	assert.Equal(t, t1.Token, t2.Token, "same owner reuses the installation token")
	tb, err := m.Token(ctx, "app", id, "b/x")
	require.NoError(t, err)
	assert.Equal(t, "tok-22-2", tb.Token)

	now = now.Add(49 * time.Minute)
	t3, err := m.Token(ctx, "app", id, "a/one")
	require.NoError(t, err)
	assert.Equal(t, "tok-11-1", t3.Token, "11m left: still cached")
	now = now.Add(2 * time.Minute)
	t4, err := m.Token(ctx, "app", id, "a/one")
	require.NoError(t, err)
	assert.Equal(t, "tok-11-3", t4.Token, "under 10m left: refreshed")
	assert.Equal(t, int32(3), calls.Load())
}

func TestMinterErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	id := config.Identity{Kind: config.IdentityGitHubApp, AppID: 1, PrivateKeyFile: WriteKey(t), APIURL: srv.URL, Installations: map[string]int64{"a": 1}, Repos: []string{"a/*"}}
	_, err := identity.NewMinter().Token(t.Context(), "app", id, "a/b")
	require.ErrorContains(t, err, "Bad credentials")

	lin := filepath.Join(t.TempDir(), "lin")
	require.NoError(t, os.WriteFile(lin, []byte("lin_api_x\n"), 0o600))
	tok, err := identity.NewMinter().Token(t.Context(), "lin", config.Identity{Kind: config.IdentityLinear, TokenFile: lin}, "ENG")
	require.NoError(t, err)
	assert.Equal(t, "lin_api_x", tok.Token)
}

// TestGitHubTokenLogin: the PAT is read from its file, sent as the
// bearer token to GET /user, and the login is cached.
func TestGitHubTokenLogin(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/user" || r.Header.Get("Authorization") != "Bearer ghp_x" {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"login":"alice"}`))
	}))
	t.Cleanup(srv.Close)
	pat := filepath.Join(t.TempDir(), "pat")
	require.NoError(t, os.WriteFile(pat, []byte("ghp_x\n"), 0o600))
	id := config.Identity{Kind: config.IdentityGitHubToken, TokenFile: pat, Repos: []string{"a/*"}, APIURL: srv.URL}
	m := identity.NewMinter()
	tok, err := m.Token(t.Context(), "me", id, "a/b")
	require.NoError(t, err)
	assert.Equal(t, "ghp_x", tok.Token)
	assert.Equal(t, config.IdentityGitHubToken, tok.Kind)
	for range 2 {
		l, err := m.Login(t.Context(), "me", id)
		require.NoError(t, err)
		assert.Equal(t, "alice", l)
	}
	assert.Equal(t, int32(1), calls.Load())

	require.NoError(t, os.WriteFile(pat, []byte("wrong"), 0o600))
	m.Forget()
	_, err = m.Login(t.Context(), "me", id)
	require.ErrorContains(t, err, "Bad credentials")
}

// fakeLinearOAuth serves Linear's client_credentials token endpoint. The
// client secret must be "good".
func fakeLinearOAuth(t *testing.T, ttl time.Duration) (*httptest.Server, *atomic.Int32, *atomic.Value) {
	t.Helper()
	var n atomic.Int32
	var scope atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.ParseForm() != nil || r.PostForm.Get("grant_type") != "client_credentials" {
			http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
			return
		}
		if r.PostForm.Get("client_id") != "cid" || r.PostForm.Get("client_secret") != "good" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"Client authentication failed"}`))
			return
		}
		scope.Store(r.PostForm.Get("scope"))
		c := n.Add(1)
		_, _ = w.Write([]byte(`{"access_token":"lin_oauth_` + strconv.Itoa(int(c)) + `","token_type":"Bearer","expires_in":` + strconv.Itoa(int(ttl.Seconds())) + `}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &n, &scope
}

func writeFile(t *testing.T, name, s string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(p, []byte(s+"\n"), 0o600))
	return p
}

// TestLinearApp: client_credentials mint, cache hit, refresh under 10m
// left, and a bad secret error that names neither secret nor token.
func TestLinearApp(t *testing.T) {
	now := time.Date(2026, 11, 4, 6, 0, 0, 0, time.UTC)
	srv, calls, scope := fakeLinearOAuth(t, 24*time.Hour)
	id := config.Identity{
		Kind: config.IdentityLinearApp, ClientIDFile: writeFile(t, "id", "cid"),
		ClientSecretFile: writeFile(t, "secret", "good"), TokenURL: srv.URL, Workspace: "acme", Teams: []string{"ENG"},
	}
	m := identity.NewMinter()
	m.Now = func() time.Time { return now }
	ctx := t.Context()

	t1, err := m.Token(ctx, "lapp", id, "ENG")
	require.NoError(t, err)
	assert.Equal(t, "lin_oauth_1", t1.Token)
	assert.Equal(t, now.Add(24*time.Hour), t1.Expires)
	assert.Equal(t, "read,write,issues:create,comments:create", scope.Load(), "default scopes")

	now = now.Add(23*time.Hour + 49*time.Minute)
	t2, err := m.Token(ctx, "lapp", id, "ENG")
	require.NoError(t, err)
	assert.Equal(t, "lin_oauth_1", t2.Token, "11m left: cached")
	now = now.Add(2 * time.Minute)
	t3, err := m.Token(ctx, "lapp", id, "ENG")
	require.NoError(t, err)
	assert.Equal(t, "lin_oauth_2", t3.Token, "under 10m left: refreshed")

	// Another scope set is a separate cache entry.
	id.Scopes = []string{"read"}
	t4, err := m.Token(ctx, "lapp", id, "ENG")
	require.NoError(t, err)
	assert.Equal(t, "lin_oauth_3", t4.Token)
	assert.Equal(t, "read", scope.Load())
	assert.Equal(t, int32(3), calls.Load())

	bad := id
	bad.ClientSecretFile = writeFile(t, "bad", "s3cret-wrong")
	_, err = identity.NewMinter().Token(ctx, "lapp", bad, "ENG")
	require.ErrorContains(t, err, "invalid_client")
	assert.NotContains(t, err.Error(), "s3cret-wrong")
	assert.NotContains(t, err.Error(), "lin_oauth")
}

// TestGitHubAppLogin: an App's login is its slug plus [bot], from GET
// /app with the App JWT (#402).
func TestGitHubAppLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app" || strings.Count(r.Header.Get("Authorization"), ".") != 2 {
			http.Error(w, "bad", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"slug":"my-app"}`))
	}))
	t.Cleanup(srv.Close)
	id := config.Identity{Kind: config.IdentityGitHubApp, AppID: 7, PrivateKeyFile: WriteKey(t), APIURL: srv.URL}
	l, err := identity.NewMinter().Login(t.Context(), "app", id)
	require.NoError(t, err)
	assert.Equal(t, "my-app[bot]", l)
}
