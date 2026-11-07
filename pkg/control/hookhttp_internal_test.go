package control

import (
	"crypto/hmac"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
)

func TestHookURLControllerConfiguration(t *testing.T) {
	t.Parallel()
	c, _, _ := subController(t)
	secret := []byte("topsecret")
	require.NoError(t, os.WriteFile(filepath.Join(c.root.Dir, "sec"), secret, 0o600))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !hmac.Equal([]byte(r.Header.Get("X-Jobshop-Signature-256")), []byte(hookSignature(secret, body))) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"decision":"deny","reason":"remote says no"}`)
	}))
	t.Cleanup(srv.Close)
	cfg := *c.config()
	cfg.Hooks = []config.Hook{{On: config.HookAction, URL: srv.URL, SecretFile: config.SecretPath("sec")}}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	got := c.runHooks(t.Context(), c.config(), hooks.GuardEvent{On: config.HookAction, Action: "push"})
	assert.Equal(t, hooks.Answer{Decision: rpc.GuardDeny, Reason: "remote says no"}, got)
}
