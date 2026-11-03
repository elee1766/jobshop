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
