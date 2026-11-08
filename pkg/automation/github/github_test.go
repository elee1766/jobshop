package github_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/automation/api"
)

const token = "ghs_minted-per-call"

func TestGitHubCalls(t *testing.T) {
	var requests atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("%s: missing host token", r.URL.Path)
		}
		switch r.URL.Path {
		case "/repos/o/r/pulls":
			if r.URL.Query().Get("page") == "2" {
				_, _ = io.WriteString(w, `[{"number":2}]`)
				return
			}
			if r.URL.Query().Get("state") != "open" {
				t.Errorf("params not sent: %s", r.URL.RawQuery)
			}
			w.Header().Set("Link", "<"+server.URL+"/repos/o/r/pulls?state=open&page=2>; rel=\"next\"")
			_, _ = io.WriteString(w, `[{"number":1}]`)
		case "/repos/o/r/issues/7/comments":
			body, _ := io.ReadAll(r.Body)
			if r.Method != http.MethodPost || !strings.Contains(string(body), `"body":"hi"`) {
				t.Errorf("comment: %s %s", r.Method, body)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"id":9,"echo":"`+token+`"}`)
		case "/graphql":
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "bad") {
				_, _ = io.WriteString(w, `{"errors":[{"message":"Field 'bad' doesn't exist"}]}`)
				return
			}
			_, _ = io.WriteString(w, `{"data":{"repository":{"name":"r"}}}`)
		case "/redirect":
			http.Redirect(w, r, "https://elsewhere.example/", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var repos []string
	records := map[int]automation.Observation{}
	options := automation.Options{
		GitHub: func(_ context.Context, repo string) (api.Token, error) {
			repos = append(repos, repo)
			return api.Token{APIURL: server.URL, Token: token}, nil
		},
		Observe: func(_ context.Context, o automation.Observation) error {
			if o.Phase == "completed" {
				records[o.Index] = o
			}
			return nil
		},
	}
	source := `
def handle():
    pulls = github.get("/repos/o/r/pulls", params={"state": "open"}, pages=3).json()
    posted = github.post("/repos/o/r/issues/7/comments", json={"body": "hi"})
    data = github.graphql("query { repository(owner: \"o\", name: \"r\") { name } }", repo="o/r")
    return {"pulls": [p["number"] for p in pulls], "posted": posted.status_code, "name": data["repository"]["name"]}

def bad():
    return github.graphql("query { bad }", repo="o/r")

def redirect():
    return github.get("/redirect", repo="o/r")

def no_auth_header():
    return github.get("/repos/o/r/pulls", headers={"Authorization": "x"})
`
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.star"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := []string{"main.star:handle", "main.star:bad", "main.star:redirect", "main.star:no_auth_header"}
	engine, err := automation.Load(t.Context(), root, entries, options)
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(res.Output)
	if string(got) != `{"name":"r","posted":201,"pulls":[1,2]}` {
		t.Fatalf("output %s", got)
	}
	if strings.Join(repos, ",") != "o/r,o/r,o/r" {
		t.Fatalf("identity selection used repos %v", repos)
	}
	encoded, _ := json.Marshal(records)
	if strings.Contains(string(encoded), token) || !strings.Contains(string(encoded), "[redacted]") {
		t.Fatalf("token not redacted from observations: %s", encoded)
	}
	if _, err := engine.Invoke(t.Context(), "main.star:bad"); err == nil || !strings.Contains(err.Error(), "Field 'bad' doesn't exist") {
		t.Fatalf("graphql errors not raised: %v", err)
	}
	if _, err := engine.Invoke(t.Context(), "main.star:redirect"); err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("redirect off the API followed: %v", err)
	}
	if _, err := engine.Invoke(t.Context(), "main.star:no_auth_header"); err == nil || !strings.Contains(err.Error(), "host sets Authorization") {
		t.Fatalf("script set Authorization: %v", err)
	}

	// Replay restores recorded results without calling GitHub.
	before := requests.Load()
	options.Observe = nil
	options.Replay = func(_ context.Context, o automation.Observation) (any, bool, error) {
		r, ok := records[o.Index]
		return r.Result, ok, nil
	}
	replay, err := automation.Load(t.Context(), root, entries[:1], options)
	if err != nil {
		t.Fatal(err)
	}
	again, err := replay.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() != before {
		t.Fatal("replay called GitHub")
	}
	if replayed, _ := json.Marshal(again.Output); string(replayed) != string(got) {
		t.Fatalf("replayed %s, want %s", replayed, got)
	}

	disabled, err := automation.Load(t.Context(), root, entries[:1], automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := disabled.Invoke(t.Context(), "main.star:handle"); err == nil || !strings.Contains(err.Error(), "disabled by host") {
		t.Fatalf("github ran without a grant: %v", err)
	}
}
