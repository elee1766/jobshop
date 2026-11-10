package control

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// revokingLinear is a Linear app that keeps one active token: every mint,
// the controller's or "someone else's", revokes the one before it. Its
// GraphQL side answers 401 to a revoked token.
type revokingLinear struct {
	mu     sync.Mutex
	active string
	mints  int
}

func (l *revokingLinear) mint() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mints++
	l.active = "lin_oauth_" + strconv.Itoa(l.mints)
	return l.active
}

func (l *revokingLinear) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.mints
}

func (l *revokingLinear) valid(auth string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return auth == "Bearer "+l.active
}

func (l *revokingLinear) serve(t *testing.T) (oauthURL, gqlURL string) {
	t.Helper()
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"` + l.mint() + `","expires_in":2592000}`))
	}))
	t.Cleanup(oauth.Close)
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if !l.valid(r.Header.Get("Authorization")) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Authentication required, not authenticated","extensions":{"code":"AUTHENTICATION_ERROR"}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","name":"elee-agent"}}}`))
	}))
	t.Cleanup(gql.Close)
	return oauth.URL, gql.URL
}

// linearAppController is a controller with one linear-app identity for
// team ENG.
func linearAppController(t *testing.T, oauthURL, gqlURL string) *Controller {
	t.Helper()
	dir := t.TempDir()
	write := func(name, s string) string {
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, []byte(s+"\n"), 0o600))
		return p
	}
	cfg, err := config.Parse([]byte("projects: {js: {identities: [lapp]}}\nidentities:\n  lapp:\n    kind: linear-app\n    client_id_file: " +
		write("cid", "cid") + "\n    client_secret_file: " + write("secret", "good") + "\n    token_url: " + oauthURL +
		"\n    workspace: acme\n    teams: [ENG]\n"))
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	c := &Controller{
		live: new(atomic.Pointer[liveConfig]), log: slog.New(slog.DiscardHandler),
		minter: identity.NewMinter(), linearURL: gqlURL,
	}
	c.live.Store(lc)
	return c
}

// TestSessionTokenRevokedByAnotherMinter (#866): another process mints a
// token for the same Linear app, which revokes the one the controller
// cached and handed to a session, weeks before its expires_in. A session
// that asks again naming the refused token gets one that works, and
// sessions holding the same refused token share one mint.
func TestSessionTokenRevokedByAnotherMinter(t *testing.T) {
	t.Parallel()
	lin := &revokingLinear{}
	oauth, gql := lin.serve(t)
	c := linearAppController(t, oauth, gql)
	ctx := t.Context()
	session := func(id string) scope { return scope{actor: id, project: "js"} }

	r, err := c.identityToken(ctx, session("js-s-1"), rpc.IdentityParams{Target: "ENG"})
	require.NoError(t, err)
	held := r.(rpc.IdentityTokenResult).Token
	assert.Equal(t, "lin_oauth_1", held)

	// Without Stale the cache answers, revoked or not: that is the bug.
	assert.Equal(t, "lin_oauth_2", lin.mint(), "the other minter")
	r, err = c.identityToken(ctx, session("js-s-1"), rpc.IdentityParams{Target: "ENG"})
	require.NoError(t, err)
	assert.Equal(t, held, r.(rpc.IdentityTokenResult).Token)
	require.False(t, lin.valid("Bearer "+held))

	// With Stale the controller renews once, and a second session with the
	// same refused token gets the renewed token, not another mint.
	r, err = c.identityToken(ctx, session("js-s-1"), rpc.IdentityParams{Target: "ENG", Stale: held})
	require.NoError(t, err)
	fresh := r.(rpc.IdentityTokenResult).Token
	assert.Equal(t, "lin_oauth_3", fresh)
	assert.True(t, lin.valid("Bearer "+fresh))
	r, err = c.identityToken(ctx, session("js-s-2"), rpc.IdentityParams{Target: "ENG", Stale: held})
	require.NoError(t, err)
	assert.Equal(t, fresh, r.(rpc.IdentityTokenResult).Token)
	assert.Equal(t, 3, lin.count())

	// A stale token the cache has moved past is not renewed again.
	r, err = c.identityToken(ctx, session("js-s-3"), rpc.IdentityParams{Target: "ENG", Stale: "lin_oauth_2"})
	require.NoError(t, err)
	assert.Equal(t, fresh, r.(rpc.IdentityTokenResult).Token)
	assert.Equal(t, 3, lin.count())
}

// TestStaleIgnoredForOtherKinds: a static Linear key does not change by
// asking again, and a stale token never reaches a GitHub identity.
func TestStaleIgnoredForOtherKinds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := filepath.Join(dir, "key")
	require.NoError(t, os.WriteFile(key, []byte("lin_api_static\n"), 0o600))
	cfg, err := config.Parse([]byte("projects: {js: {identities: [lin]}}\nidentities:\n  lin:\n    kind: linear\n    token_file: " + key +
		"\n    workspace: acme\n    teams: [ENG]\n"))
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	c := &Controller{live: new(atomic.Pointer[liveConfig]), log: slog.New(slog.DiscardHandler), minter: identity.NewMinter()}
	c.live.Store(lc)
	r, err := c.identityToken(t.Context(), scope{actor: "js-s-1", project: "js"}, rpc.IdentityParams{Target: "ENG", Stale: "lin_api_static"})
	require.NoError(t, err)
	assert.Equal(t, "lin_api_static", r.(rpc.IdentityTokenResult).Token)
}
