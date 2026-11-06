package desk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
)

func writePlugin(t *testing.T, dir, name, element string) {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(p, 0o755))
	m := PluginManifest{Name: name, Entry: "index.js", Contributes: []Contribution{{Slot: "view", Element: element, ID: name, Title: name}}}
	b, _ := json.Marshal(m)
	require.NoError(t, os.WriteFile(filepath.Join(p, "plugin.json"), b, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(p, "index.js"), []byte("// "+element), 0o600))
}

func showWith(desk map[string]any, packs ...string) rpc.ConfigShowResult {
	cs := rpc.ConfigShowResult{Config: map[string]any{}}
	if desk != nil {
		cs.Config["desk"] = desk
	}
	for _, p := range packs {
		cs.Instances = append(cs.Instances, config.PackInstance{Path: p})
	}
	return cs
}

func TestPluginsResolve(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, filepath.Join(root, "pk", "desk"), "shared", "pack-view")
	writePlugin(t, filepath.Join(root, "pk", "desk"), "packonly", "pack-only")
	writePlugin(t, filepath.Join(root, "desk", "plugins"), "shared", "root-view")
	writePlugin(t, filepath.Join(root, "desk", "plugins"), "off", "off-view")
	// Broken plugins are skipped, not fatal.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "desk", "plugins", "bad"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "desk", "plugins", "bad", "plugin.json"), []byte(`{"name":"bad","entry":"../x.js"}`), 0o600))

	ctl := &fakeCtl{results: map[string]any{rpc.MethodConfigShow: showWith(map[string]any{
		"plugins": map[string]any{"off": map[string]any{"enabled": false}, "shared": map[string]any{"config": map[string]any{"x": 1}}},
	}, "pk")}}
	h := New(Config{Root: root}, ctl, nil).Handler()
	var list []Plugin
	get(t, h, "/api/plugins", &list)
	byName := map[string]Plugin{}
	for _, p := range list {
		byName[p.Name] = p
	}
	assert.NotContains(t, byName, "off")
	assert.NotContains(t, byName, "bad")
	assert.Equal(t, "pack:pk", byName["packonly"].Source)
	sh := byName["shared"]
	assert.Equal(t, "root", sh.Source)
	assert.Equal(t, "pack:pk", sh.Overrides)
	assert.Equal(t, "root-view", sh.Contributes[0].Element)
	assert.EqualValues(t, 1, sh.Config["x"])
	assert.Equal(t, "plugins/shared/", sh.URL)

	// Files come from the winning source.
	w := get(t, h, "/plugins/shared/index.js", nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "// root-view", w.Body.String())
	assert.Equal(t, "text/javascript; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, http.StatusNotFound, get(t, h, "/plugins/off/index.js", nil).Code)
	// The mux redirects a literal "..", so try the escaped form.
	assert.Equal(t, http.StatusNotFound, get(t, h, "/plugins/shared/..%2fpackonly%2findex.js", nil).Code)
	assert.Equal(t, http.StatusNotFound, get(t, h, "/plugins/shared/nope.js", nil).Code)
}

func TestPluginSlotNames(t *testing.T) {
	for slot, ok := range map[string]bool{
		"view":                  true,
		"session":               true,
		"work:toolbar":          true,
		"chat:composer-actions": true,
		"my-plugin:a1":          true,
		"toolbar":               false,
		"Work:toolbar":          false,
		"work:":                 false,
		":toolbar":              false,
		"work:tool:bar":         false,
		"work:tool bar":         false,
	} {
		m := PluginManifest{Name: "p", Entry: "index.js", Contributes: []Contribution{{Slot: slot, Element: "p-x", ID: "p"}}}
		if ok {
			assert.NoError(t, m.validate(), slot)
		} else {
			assert.ErrorContains(t, m.validate(), "<plugin>:<name>", slot)
		}
	}
}

func TestPluginFileNoSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, filepath.Join(root, "desk", "plugins"), "p", "p-view")
	secret := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(secret, []byte("s"), 0o600))
	require.NoError(t, os.Symlink(secret, filepath.Join(root, "desk", "plugins", "p", "leak.js")))
	ctl := &fakeCtl{results: map[string]any{rpc.MethodConfigShow: showWith(nil)}}
	h := New(Config{Root: root}, ctl, nil).Handler()
	assert.Equal(t, http.StatusNotFound, get(t, h, "/plugins/p/leak.js", nil).Code)
}

func postRaw(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Host = "127.0.0.1:7480"
	req.Header.Set(CSRFHeader, "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestRPCPolicy(t *testing.T) {
	for _, tc := range []struct {
		m    string
		p    config.DeskRPC
		want bool
	}{
		{rpc.MethodTicketList, config.DeskRPC{}, true},
		{rpc.MethodClaim, config.DeskRPC{Allow: []string{"*"}}, false},
		{rpc.MethodEventsSubscribe, config.DeskRPC{}, false},
		{rpc.MethodIdentityToken, config.DeskRPC{}, false},
		{rpc.MethodIdentityToken, config.DeskRPC{Allow: []string{"identity.*"}}, true},
		{rpc.MethodShutdown, config.DeskRPC{Allow: []string{"*"}}, true},
		{rpc.MethodTicketCreate, config.DeskRPC{Deny: []string{"ticket.*"}}, false},
		{rpc.MethodTicketList, config.DeskRPC{Allow: []string{"run.*"}}, false},
		{"nope.method", config.DeskRPC{Allow: []string{"*"}}, false},
	} {
		assert.Equal(t, tc.want, rpcAllowed(tc.m, tc.p), "%s %+v", tc.m, tc.p)
	}
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

	w = postRaw(t, h, "/api/rpc/claim", `{}`)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"forbidden"`)
	assert.Equal(t, http.StatusForbidden, postRaw(t, h, "/api/rpc/learn.add", `{}`).Code)
	assert.Equal(t, http.StatusBadRequest, postRaw(t, h, "/api/rpc/ticket.show", `[1]`).Code)

	// Writes need the CSRF header like every other POST.
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
	assert.Contains(t, list.Schemas, rpc.MethodTicketList)
	assert.NotContains(t, list.Schemas, rpc.MethodClaim)
	assert.Equal(t, "#/$defs/rpc.ListParams", list.Schemas[rpc.MethodTicketList].Params["$ref"])
	assert.Contains(t, list.Defs, "store.Ticket")
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
