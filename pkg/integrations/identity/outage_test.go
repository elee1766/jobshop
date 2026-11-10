package identity_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
)

func TestOutageClassification(t *testing.T) {
	t.Parallel()
	http401 := &identity.HTTPError{Op: "installation 1 token", Code: 401, Status: "401 Unauthorized", Body: "Bad credentials"}
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"401", http401, true},
		{"wrapped 401", fmt.Errorf("identity app: %w", http401), true},
		{"403 bad credentials", &identity.HTTPError{Code: 403, Body: "Bad credentials"}, true},
		{"403 rate limit", &identity.HTTPError{Code: 403, Body: "API rate limit exceeded"}, false},
		{"422 from the token provider", &identity.HTTPError{Code: 422}, false},
		{"404", &identity.HTTPError{Code: 404}, false},
		{"502", &identity.HTTPError{Code: 502}, true},
		{"secret command", fmt.Errorf("identity gh: %w", &config.SecretError{}), true},
		{"missing token file", fmt.Errorf("identity gh: %w", os.ErrNotExist), true},
		{"unreachable", &net.OpError{Op: "dial", Err: errors.New("connection refused")}, true},
		{"cancelled", context.Canceled, false},
		{"plain", errors.New("identity app: no installation for acme"), false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, identity.Outage(c.err), c.name)
	}
}

// TestProbe: a probe forgets the cache, mints again and reads
// /rate_limit with the new token, so it fails while GitHub refuses it.
func TestProbe(t *testing.T) {
	var status atomic.Int32
	status.Store(http.StatusUnauthorized)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rate_limit" || r.Header.Get("Authorization") != "Bearer ghp_x" {
			http.Error(w, "bad", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(int(status.Load()))
	}))
	t.Cleanup(srv.Close)
	tok := filepath.Join(t.TempDir(), "tok")
	id := config.Identity{Kind: config.IdentityGitHubToken, TokenFile: config.SecretPath(tok), APIURL: srv.URL, Repos: []string{"a/*"}}
	m := identity.NewMinter()

	err := m.Probe(t.Context(), "gh", id, "a/b")
	assert.True(t, identity.Outage(err), "no token file: %v", err)

	require.NoError(t, os.WriteFile(tok, []byte("ghp_x\n"), 0o600))
	err = m.Probe(t.Context(), "gh", id, "a/b")
	require.ErrorContains(t, err, "401")
	assert.True(t, identity.Outage(err))

	status.Store(http.StatusOK)
	require.NoError(t, m.Probe(t.Context(), "gh", id, "a/b"))
}
