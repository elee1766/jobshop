package desk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

func proxyReq(peer, user string, set bool) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Host = "127.0.0.1:7480"
	req.RemoteAddr = peer
	if set {
		req.Header.Set("X-Forwarded-User", user)
	}
	return req
}

func TestProxyUser(t *testing.T) {
	proxies, err := ParseTrustedProxies([]string{"127.0.0.1", "10.1.0.0/16"})
	require.NoError(t, err)
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionList: []rpc.SessionInfo{}}}
	h := New(Config{UserHeader: "X-Forwarded-User", TrustedProxies: proxies}, ctl, nil).Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, proxyReq("127.0.0.1:5555", "alice@example.com", true))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, []string{"desk:alice@example.com"}, ctl.actors, "the user reaches the controller as the actor")

	w = httptest.NewRecorder()
	h.ServeHTTP(w, proxyReq("10.1.2.3:5555", "bob", true))
	assert.Equal(t, http.StatusOK, w.Code)

	for name, req := range map[string]*http.Request{
		"untrusted peer":  proxyReq("192.168.1.9:5555", "alice", true),
		"missing header":  proxyReq("127.0.0.1:5555", "", false),
		"empty header":    proxyReq("127.0.0.1:5555", "", true),
		"control char":    proxyReq("127.0.0.1:5555", "al\x01ice", true),
		"too long":        proxyReq("127.0.0.1:5555", strings.Repeat("a", 300), true),
		"no peer address": proxyReq("pipe", "alice", true),
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code, name)
	}

	var who Whoami
	req := proxyReq("127.0.0.1:5555", "alice", true)
	req.URL.Path = "/api/whoami"
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &who))
	assert.Equal(t, "alice", who.User)
}

func TestNoProxyUserUnchanged(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{}}
	h := New(Config{}, ctl, nil).Handler()
	// Any peer, header ignored: today's behavior.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, proxyReq("192.168.1.9:5555", "mallory", true))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{""}, ctl.actors, "no actor: the admin socket's own")
	var who Whoami
	get(t, h, "/api/whoami", &who)
	assert.Empty(t, who.User)
}

func TestParseTrustedProxies(t *testing.T) {
	_, err := ParseTrustedProxies([]string{"nope"})
	require.Error(t, err)
	p, err := ParseTrustedProxies([]string{"::1", "100.64.0.0/10"})
	require.NoError(t, err)
	assert.Equal(t, "::1/128", p[0].String())
}
