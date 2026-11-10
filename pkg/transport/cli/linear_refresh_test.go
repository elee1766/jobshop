package cli

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

const authRefused = `{"errors":[{"message":"Authentication required, not authenticated","extensions":{"code":"AUTHENTICATION_ERROR"}}]}`

// linearApp is a Linear app with one active token, as Linear keeps per
// scope set: the controller and "someone else" both mint, and a mint
// revokes the token before it.
type linearApp struct {
	mu     sync.Mutex
	active string
	mints  int
	reqs   int // GraphQL requests
	stales []string
	// sticky: the controller hands back the token it has even when the
	// caller names it stale (a controller that cannot refresh).
	sticky bool
	held   string // the controller's cached token
}

func (a *linearApp) mintLocked() string {
	a.mints++
	a.active = "lin_oauth_" + strconv.Itoa(a.mints)
	return a.active
}

// otherMinter is another process minting for the same app.
func (a *linearApp) otherMinter() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.mintLocked()
}

func (a *linearApp) serveController(t *testing.T) string {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "s")
	ln, err := net.Listen("unix", sock)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			nc, err := ln.Accept()
			if err != nil {
				return
			}
			gate := rpc.NewGate(nc)
			conn := acp.NewConnection(func(_ context.Context, method string, raw json.RawMessage) (any, *acp.RequestError) {
				raw, _ = rpc.SplitMeta(raw)
				a.mu.Lock()
				defer a.mu.Unlock()
				switch method {
				case rpc.MethodIdentityToken:
					var p rpc.IdentityParams
					_ = json.Unmarshal(raw, &p)
					if p.Stale != "" {
						a.stales = append(a.stales, p.Stale)
					}
					if a.held == "" || (p.Stale == a.held && !a.sticky) {
						a.held = a.mintLocked()
					}
					return rpc.IdentityTokenResult{Token: a.held}, nil
				case rpc.MethodActivityRecord:
					return struct{}{}, nil
				}
				return nil, acp.NewMethodNotFound(method)
			}, nc, gate)
			conn.SetLogger(slog.New(slog.DiscardHandler))
			gate.Open()
		}
	}()
	return sock
}

func (a *linearApp) serveGraphQL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		a.mu.Lock()
		defer a.mu.Unlock()
		a.reqs++
		if r.Header.Get("Authorization") != "Bearer "+a.active {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, authRefused)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"issue":{"id":"i1","identifier":"JOB-7","title":"t","description":"","url":"u","state":{"id":"s","name":"Todo","type":"unstarted"},"team":{"id":"t1","key":"JOB"},"comments":{"nodes":[]}}}}`)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func runLinearIssue(t *testing.T, g *Globals, url string) (string, error) {
	t.Helper()
	var out strings.Builder
	old := stdout
	stdout = &out
	t.Cleanup(func() { stdout = old })
	err := linearIssueCmd{Key: "JOB-7"}.Run(g, &linearCmd{APIURL: url})
	return out.String(), err
}

// TestLinearSessionTokenRevoked (#866): the token a session got from the
// controller was revoked when another process minted for the same Linear
// app. The command asks the controller again naming the dead token, and
// succeeds, instead of failing "Authentication required".
func TestLinearSessionTokenRevoked(t *testing.T) {
	a := &linearApp{}
	g := &Globals{Socket: a.serveController(t)}
	url := a.serveGraphQL(t)

	// A first call caches the controller's token and works.
	_, err := runLinearIssue(t, g, url)
	require.NoError(t, err)
	assert.Empty(t, a.stales)

	// Another minter revokes it. The next call recovers in one retry.
	a.otherMinter()
	out, err := runLinearIssue(t, g, url)
	require.NoError(t, err)
	assert.Contains(t, out, `"identifier": "JOB-7"`)
	a.mu.Lock()
	defer a.mu.Unlock()
	assert.Equal(t, []string{"lin_oauth_1"}, a.stales, "the refused token is named")
	assert.Equal(t, 3, a.mints, "the controller's, the other's, one refresh")
}

// TestLinearSessionAuthFailureIsAnError: when the controller can only hand
// back the same dead token, the command fails with Linear's message without
// resending it and prints nothing, as it would for a missing issue.
func TestLinearSessionAuthFailureIsAnError(t *testing.T) {
	a := &linearApp{sticky: true}
	g := &Globals{Socket: a.serveController(t)}
	url := a.serveGraphQL(t)

	_, err := runLinearIssue(t, g, url)
	require.NoError(t, err)
	a.otherMinter()
	a.mu.Lock()
	before := a.reqs
	a.mu.Unlock()

	out, err := runLinearIssue(t, g, url)
	require.Error(t, err)
	assert.ErrorContains(t, err, "Authentication required, not authenticated")
	assert.Empty(t, out, "no output as if the read had worked")
	a.mu.Lock()
	defer a.mu.Unlock()
	assert.Equal(t, before+1, a.reqs, "the refused request only: the same token is not resent")
	assert.Len(t, a.stales, 1)
}
