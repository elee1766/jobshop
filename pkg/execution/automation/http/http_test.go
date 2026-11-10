package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gfx-labs/jobshop/pkg/execution/automation"
)

func hook(t *testing.T, source string, options automation.Options) *automation.Engine {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.star"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	engine, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, options)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func TestHTTPAuthJSONAndRedirects(t *testing.T) {
	var foreignRequests atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		foreignRequests.Add(1)
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-Credential") != "" {
			t.Error("credential leaked across origin")
		}
	}))
	defer foreign.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/public-redirect" {
			http.Redirect(w, r, foreign.URL, http.StatusFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing named credential")
		}
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, foreign.URL, http.StatusFound)
			return
		}
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Error("wrong JSON request contract")
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if string(data) != "{\"nested\":[1,true]}" {
			t.Errorf("wrong JSON body %s", data)
		}
		_, _ = io.WriteString(w, `{"value":9007199254740993}`)
	}))
	defer server.Close()
	options := automation.Options{AllowHTTP: true, Auth: map[string]automation.Credential{"service": {Origins: []string{server.URL}, Headers: map[string]string{"Authorization": "Bearer secret"}}}}
	engine := hook(t, `def handle(url):
    response = http.post(url, auth="service", json={"nested": [1, True]})
    return {"status": response.status_code, "value": response.json()["value"]}
`, options)
	result, err := engine.Invoke(t.Context(), "main.star:handle", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if result.Output["value"] != int64(9007199254740993) {
		t.Fatalf("JSON conversion lost data: %#v", result.Output)
	}
	for _, url := range []string{foreign.URL, server.URL + "/redirect"} {
		if _, err := engine.Invoke(t.Context(), "main.star:handle", url); err == nil {
			t.Fatal("allowed credential origin violation")
		}
	}
	if foreignRequests.Load() != 0 {
		t.Fatal("authenticated redirect reached another origin")
	}
	public := hook(t, "def handle(url):\n    return {\"status\": http.get(url, headers={\"Authorization\": \"caller-secret\", \"X-Credential\": \"custom-secret\"}).status_code}\n", automation.Options{AllowHTTP: true})
	if _, err := public.Invoke(t.Context(), "main.star:handle", server.URL+"/public-redirect"); err != nil {
		t.Fatal(err)
	}
	if foreignRequests.Load() != 1 {
		t.Fatal("unauthenticated redirect did not reach destination")
	}
}

func TestHTTPResponseLimitAndCancellation(t *testing.T) {
	started := make(chan struct{})
	var smallRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/small" {
			smallRequests.Add(1)
			_, _ = io.WriteString(w, "{}")
			return
		}
		if r.URL.Path == "/large" {
			_, _ = io.WriteString(w, strings.Repeat("x", 100))
			return
		}
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	engine := hook(t, "def handle(url):\n    return http.get(url).text\n", automation.Options{AllowHTTP: true, Limits: automation.Limits{OutputBytes: 10}})
	if _, err := engine.Invoke(t.Context(), "main.star:handle", server.URL+"/large"); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("unbounded HTTP output: %v", err)
	}
	quota := hook(t, "def handle(url):\n    http.get(url)\n    return http.get(url).text\n", automation.Options{AllowHTTP: true, Limits: automation.Limits{Calls: 1}})
	if _, err := quota.Invoke(t.Context(), "main.star:handle", server.URL+"/small"); err == nil || !strings.Contains(err.Error(), "call limit") {
		t.Fatalf("unbounded capability calls: %v", err)
	}
	if smallRequests.Load() != 1 {
		t.Fatalf("call budget allowed %d requests", smallRequests.Load())
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { _, err := engine.Invoke(ctx, "main.star:handle", server.URL); done <- err }()
	<-started
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("HTTP cancellation succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP cancellation blocked")
	}
}
