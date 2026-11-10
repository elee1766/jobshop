package desk

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postRaw(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(h, path, body, nil)
}

func showWith(desk map[string]any) rpc.ConfigShowResult {
	return rpc.ConfigShowResult{Config: map[string]any{"desk": desk}}
}

func TestRPCCall(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodConfigShow: showWith(map[string]any{"rpc": map[string]any{"deny": []string{"learn.*"}}}),
		rpc.MethodTicketShow: map[string]any{"ticket": map[string]any{"id": "js-1"}},
	}}
	h := New(Config{}, ctl, nil).Handler()
	w := postRaw(t, h, "/api/rpc/ticket.show", `{"id":"js-1","_meta":{"actor":"evil"}}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.JSONEq(t, `{"ticket":{"id":"js-1"}}`, w.Body.String())
	assert.Contains(t, ctl.calls[len(ctl.calls)-1], `ticket.show {"id":"js-1"}`)
	assert.Equal(t, http.StatusForbidden, postRaw(t, h, "/api/rpc/claim", `{}`).Code)
	assert.Equal(t, http.StatusForbidden, postRaw(t, h, "/api/rpc/learn.add", `{}`).Code)
	assert.Equal(t, http.StatusBadRequest, postRaw(t, h, "/api/rpc/ticket.show", `[1]`).Code)
	req := httptest.NewRequest(http.MethodPost, "/api/rpc/ticket.show", strings.NewReader(`{}`))
	req.Host = "127.0.0.1:7480"
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusForbidden, rw.Code)
	var list RPCList
	get(t, h, "/api/rpc", &list)
	assert.Contains(t, list.Methods, rpc.MethodTicketList)
	assert.NotContains(t, list.Methods, rpc.MethodClaim)
	assert.NotContains(t, list.Methods, rpc.MethodLearnAdd)
	assert.NotContains(t, list.Methods, rpc.MethodIdentityToken)
	assert.Equal(t, "#/$defs/rpc.ListParams", list.Schemas[rpc.MethodTicketList].Params["$ref"])
	assert.Contains(t, list.Defs, "domain.Ticket")
}

func TestRPCActorIsProxyUser(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodConfigShow: showWith(nil)}}
	h := New(Config{UserHeader: "X-User", TrustedProxies: mustProxies(t, "192.0.2.1")}, ctl, nil).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/rpc/ticket.list", strings.NewReader(`{}`))
	req.Host = "127.0.0.1:7480"
	req.RemoteAddr = "192.0.2.1:5000"
	req.Header.Set(CSRFHeader, "1")
	req.Header.Set("X-User", "alice")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, "desk:alice", ctl.actors[len(ctl.actors)-1])
}

func mustProxies(t *testing.T, v ...string) []netip.Prefix {
	t.Helper()
	p, err := ParseTrustedProxies(v)
	require.NoError(t, err)
	return p
}
