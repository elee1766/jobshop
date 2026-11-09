package desk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func embedCtl(embeds ...map[string]any) *fakeCtl {
	return &fakeCtl{results: map[string]any{rpc.MethodConfigShow: showWith(map[string]any{"embeds": embeds})}}
}

func indexReq(h http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "127.0.0.1:7480"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestEmbedsAPI(t *testing.T) {
	ctl := embedCtl(
		map[string]any{"id": "grafana", "label": "Grafana", "url": "https://grafana.example/d/1", "sandbox": []string{"allow-forms", "allow-same-origin"}, "allow": []string{"fullscreen", "clipboard-write"}},
		map[string]any{"id": "local", "label": "local", "url": "http://127.0.0.1:3000/"},
		// Same host as Desk (any port): never gets allow-same-origin.
		map[string]any{"id": "mine", "label": "mine", "url": "http://127.0.0.1:9000/", "sandbox": []string{"allow-same-origin"}},
		// A config from another version that fails today's checks is dropped.
		map[string]any{"id": "bad", "label": "bad", "url": "javascript:alert(1)"},
	)
	h := New(Config{}, ctl, nil).Handler()
	var list EmbedList
	get(t, h, "/api/embeds", &list)
	assert.Equal(t, []EmbedView{
		{ID: "grafana", Label: "Grafana", URL: "https://grafana.example/d/1", Sandbox: "allow-scripts allow-forms allow-same-origin", Allow: "fullscreen; clipboard-write"},
		{ID: "local", Label: "local", URL: "http://127.0.0.1:3000/", Sandbox: "allow-scripts"},
		{ID: "mine", Label: "mine", URL: "http://127.0.0.1:9000/", Sandbox: "allow-scripts"},
	}, list.Embeds)
}

func TestEmbedsAPIEmpty(t *testing.T) {
	h := New(Config{}, &fakeCtl{results: map[string]any{rpc.MethodConfigShow: showWith(nil)}}, nil).Handler()
	w := get(t, h, "/api/embeds", nil)
	assert.JSONEq(t, `{"embeds":[]}`, w.Body.String(), "an empty list, not null")
}

func TestEmbedsCSP(t *testing.T) {
	// No embeds: the policy and page are exactly what they were.
	none := New(Config{}, &fakeCtl{results: map[string]any{rpc.MethodConfigShow: showWith(nil)}}, nil).Handler()
	w := indexReq(none)
	assert.Equal(t, baseCSP, w.Header().Get("Content-Security-Policy"))
	assert.NotContains(t, w.Body.String(), embedOriginsMeta)

	ctl := embedCtl(
		map[string]any{"id": "a", "label": "A", "url": "https://a.example/x"},
		map[string]any{"id": "b", "label": "B", "url": "https://a.example/y"},
		map[string]any{"id": "c", "label": "C", "url": "http://127.0.0.1:3000/"},
	)
	h := New(Config{}, ctl, nil).Handler()
	w = indexReq(h)
	csp := w.Header().Get("Content-Security-Policy")
	assert.Equal(t, baseCSP+"; frame-src https://a.example http://127.0.0.1:3000", csp, "distinct origins, config order")
	for _, kept := range []string{"default-src 'self'", "script-src 'self'", "frame-ancestors 'none'", "form-action 'none'", "connect-src 'self'"} {
		assert.Contains(t, csp, kept)
	}
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"), "Desk itself is still not frameable")
	assert.Contains(t, w.Body.String(), `<meta name="`+embedOriginsMeta+`" content="https://a.example http://127.0.0.1:3000">`)

	// The API and assets keep the base policy: nothing there frames.
	assert.Equal(t, baseCSP, get(t, h, "/api/embeds", nil).Header().Get("Content-Security-Policy"))
}

func TestEmbedsCSPConfigUnavailable(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodConfigShow: assert.AnError}}
	w := indexReq(New(Config{}, ctl, nil).Handler())
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, baseCSP, w.Header().Get("Content-Security-Policy"), "no frame-src without a readable config")
}

func TestEmbedsRefuseCrossSiteWrite(t *testing.T) {
	// There is no write endpoint for embeds: a POST is the CSRF-guarded 403 or 404.
	h := New(Config{}, embedCtl(), nil).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/embeds", nil)
	req.Host = "127.0.0.1:7480"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}
