package jev_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elee1766/jobshop/pkg/execution/automation"
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

func TestJevVerifiedWireContractAndFailures(t *testing.T) {
	var requests atomic.Int32
	var foreignRequests atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { foreignRequests.Add(1) }))
	defer foreign.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer private-key" {
			t.Error("wrong Jev request method or authorization")
		}
		var body struct {
			Model     string `json:"model"`
			State     string `json:"state"`
			Questions map[string]struct {
				Type         string            `json:"type"`
				Instructions string            `json:"instructions"`
				Criteria     map[string]string `json:"criteria"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		question := body.Questions["decision"]
		if body.Model != "typesafe/jev-1.13" || body.State != `{"data":"untrusted"}` || question.Type != "choice" || question.Criteria["a"] != "first" || len(body.Questions) != 1 || !strings.Contains(question.Instructions, "untrusted data") {
			t.Errorf("wire contract mismatch: %#v", body)
		}
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, foreign.URL, http.StatusTemporaryRedirect)
		case "/bad":
			_, _ = io.WriteString(w, `{"answers":{"decision":{"type":"choice","choice":"private-key"}}}`)
		case "/down":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "private-key")
		default:
			_, _ = io.WriteString(w, `{"answers":{"decision":{"type":"choice","choice":"b","confidence":0.8}}}`)
		}
	}))
	defer server.Close()
	source := `def handle(default="", timeout="10s"):
    decision = jev(instructions="Pick one", options={"a": "first", "b": "second"}, state={"data": "untrusted"}, default=default, timeout=timeout)
    return [start_run(workflow=decision.choice, key="decision")]
`
	engine := hook(t, source, automation.Options{Jev: automation.JevOptions{Enabled: true, Endpoint: server.URL, APIKey: "private-key"}})
	result, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 1 || result.Decisions[0].Choice != "b" || result.Decisions[0].Defaulted || result.Actions[0].Workflow != "b" {
		t.Fatalf("wrong decision/action: %#v", result)
	}
	for _, test := range []struct{ path, fallback, want string }{{"/bad", "a", "not supplied"}, {"/down", "", "HTTP 503"}, {"/down", "a", ""}, {"/redirect", "", "HTTP 307"}} {
		engine := hook(t, source, automation.Options{Jev: automation.JevOptions{Enabled: true, Endpoint: server.URL + test.path, APIKey: "private-key"}})
		result, err := engine.Invoke(t.Context(), "main.star:handle", test.fallback)
		if test.want != "" {
			if err == nil || !strings.Contains(err.Error(), test.want) || strings.Contains(err.Error(), "private-key") {
				t.Fatalf("wrong or secret-bearing Jev failure: %v", err)
			}
		} else if err != nil || len(result.Decisions) != 1 || !result.Decisions[0].Defaulted || result.Decisions[0].Choice != "a" {
			t.Fatalf("wrong explicit fallback: %#v %v", result, err)
		}
	}
	if foreignRequests.Load() != 0 {
		t.Fatal("Jev forwarded credentials through a redirect")
	}
	result, err = engine.Invoke(t.Context(), "main.star:handle", "a", "1ns")
	if err != nil || len(result.Decisions) != 1 || !result.Decisions[0].Defaulted || result.Decisions[0].Choice != "a" {
		t.Fatalf("timeout did not use explicit default: %#v %v", result, err)
	}
	disabled := hook(t, source, automation.Options{})
	before := requests.Load()
	if _, err := disabled.Invoke(t.Context(), "main.star:handle"); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("disabled Jev did not fail clearly: %v", err)
	}
	if before != requests.Load() {
		t.Fatal("disabled Jev made a request")
	}
}
