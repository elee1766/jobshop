package linear_test

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

	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/execution/automation"
	linclient "github.com/elee1766/jobshop/pkg/integrations/linear"
)

const token = "lin_oauth_minted-per-call"

// The linear built-in posts GraphQL with the host's token as a Bearer,
// returns data, raises Linear's errors, redacts the token from
// observations, replays without calling Linear and needs the grant.
func TestLinearGraphQL(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("missing host token: %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "bad") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"errors":[{"message":"Cannot query field \"bad\""}]}`)
			return
		}
		if !strings.Contains(string(body), `"team":"ENG"`) {
			t.Errorf("variables not sent: %s", body)
		}
		_, _ = io.WriteString(w, `{"data":{"team":{"key":"ENG","echo":"`+token+`"}}}`)
	}))
	defer server.Close()

	var teams []string
	records := map[int]automation.Observation{}
	options := automation.Options{
		Linear: func(_ context.Context, team string) (*linclient.Client, error) {
			teams = append(teams, team)
			return linclient.New(server.URL, apicall.Static(token)), nil
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
    data = linear.graphql("query($team: String!) { team(id: $team) { key } }", variables={"team": "ENG"}, team="ENG")
    return {"key": data["team"]["key"]}

def bad():
    return linear.graphql("query { bad }", team="ENG")
`
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.star"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := []string{"main.star:handle", "main.star:bad"}
	engine, err := automation.Load(t.Context(), root, entries, options)
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(res.Output)
	if string(got) != `{"key":"ENG"}` {
		t.Fatalf("output %s", got)
	}
	if strings.Join(teams, ",") != "ENG" {
		t.Fatalf("identity selection used teams %v", teams)
	}
	encoded, _ := json.Marshal(records)
	if strings.Contains(string(encoded), token) || !strings.Contains(string(encoded), "[redacted]") {
		t.Fatalf("token not redacted from observations: %s", encoded)
	}
	if _, err := engine.Invoke(t.Context(), "main.star:bad"); err == nil || !strings.Contains(err.Error(), `Cannot query field "bad"`) {
		t.Fatalf("graphql errors not raised: %v", err)
	}

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
		t.Fatal("replay called Linear")
	}
	if replayed, _ := json.Marshal(again.Output); string(replayed) != string(got) {
		t.Fatalf("replayed %s, want %s", replayed, got)
	}

	disabled, err := automation.Load(t.Context(), root, entries[:1], automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := disabled.Invoke(t.Context(), "main.star:handle"); err == nil || !strings.Contains(err.Error(), "disabled by host") {
		t.Fatalf("linear ran without a grant: %v", err)
	}
}
