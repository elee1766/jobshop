package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// fakeLinear stands in for Linear: it records each request and answers
// with resp.
type fakeLinear struct {
	mu   sync.Mutex
	reqs []linearReq
	resp string
}

type linearReq struct {
	Auth      string
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func (f *fakeLinear) serve(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req linearReq
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &req)
		req.Auth = r.Header.Get("Authorization")
		f.mu.Lock()
		f.reqs = append(f.reqs, req)
		resp := f.resp
		f.mu.Unlock()
		if resp == "" {
			resp = `{"data":{"viewer":{"id":"u1"}}}`
		}
		_, _ = io.WriteString(w, resp)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// runLinearAPI runs c with stdin in and returns stdout, stderr and the
// error.
func runLinearAPI(t *testing.T, g *Globals, url string, c linearAPICmd, in string) (string, string, error) {
	t.Helper()
	var out, errb bytes.Buffer
	oldOut, oldIn, oldErr := stdout, stdin, stderr
	stdout, stdin, stderr = &out, strings.NewReader(in), &errb
	t.Cleanup(func() { stdout, stdin, stderr = oldOut, oldIn, oldErr })
	err := c.Run(g, &linearCmd{APIURL: url})
	return out.String(), errb.String(), err
}

func TestLinearAPIStdinTeam(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")
	lin := &fakeLinear{}
	f := &fakeController{decision: rpc.GuardAllow}
	g := &Globals{Socket: f.serve(t)}
	out, _, err := runLinearAPI(t, g, lin.serve(t), linearAPICmd{Team: "ENG", Input: "-"}, "{ viewer { id name } }\n")
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"viewer":{"id":"u1"}}}`, out)
	require.Len(t, lin.reqs, 1)
	assert.Equal(t, "{ viewer { id name } }\n", lin.reqs[0].Query)
	assert.Equal(t, "Bearer tok-ENG", lin.reqs[0].Auth, "team identity token")
	assert.Empty(t, f.guards, "a query asks no guard")
	require.Len(t, f.records, 1)
	assert.Equal(t, rpc.ActivityRecordParams{Action: "linear-api-query", Target: "ENG", Decision: rpc.GuardAllow, Command: "linear api query"}, f.records[0])
}

func TestLinearAPIJSONInputAndVariables(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")
	lin := &fakeLinear{}
	f := &fakeController{decision: rpc.GuardAllow}
	g := &Globals{Socket: f.serve(t)}
	dir := t.TempDir()
	file := filepath.Join(dir, "v.json")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v.json"), []byte(`{"a":[1,2]}`+"\n"), 0o600))
	in := `  {"query": "query Get($id: String!) { issue(id: $id) { title } }", "variables": {"x": 1}}`
	c := linearAPICmd{Input: "-", Field: []string{"id=OPS-7"}, Typed: []string{"n=3", "on=true", "obj=@" + file}}
	_, _, err := runLinearAPI(t, g, lin.serve(t), c, in)
	require.NoError(t, err)
	require.Len(t, lin.reqs, 1)
	assert.Equal(t, map[string]any{"x": 1.0, "id": "OPS-7", "n": 3.0, "on": true, "obj": map[string]any{"a": []any{1.0, 2.0}}}, lin.reqs[0].Variables)
	assert.Equal(t, "Bearer tok-OPS", lin.reqs[0].Auth, "team from the issue key in the variables")
	assert.Equal(t, "linear api query Get", f.records[0].Command)
}

// TestLinearAPIOwnKey: LINEAR_API_KEY is used and no controller is asked
// for a token; the call works with the controller down.
func TestLinearAPIOwnKey(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "lin_api_mine")
	lin := &fakeLinear{}
	g := &Globals{Socket: filepath.Join(t.TempDir(), "down")}
	out, errOut, err := runLinearAPI(t, g, lin.serve(t), linearAPICmd{Query: "{ viewer { id } }"}, "")
	require.NoError(t, err)
	assert.Contains(t, out, "u1")
	require.Len(t, lin.reqs, 1)
	assert.Equal(t, "lin_api_mine", lin.reqs[0].Auth)
	assert.Contains(t, errOut, "activity record", "the record failure only warns")
}

func TestLinearAPINoTeam(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")
	lin := &fakeLinear{}
	f := &fakeController{decision: rpc.GuardAllow}
	_, _, err := runLinearAPI(t, &Globals{Socket: f.serve(t)}, lin.serve(t), linearAPICmd{Query: "{ viewer { id } }"}, "")
	require.EqualError(t, err, "linear api: no LINEAR_API_KEY and no --team (or issue key in variables) to pick an identity")
	assert.Empty(t, lin.reqs)
}

// TestLinearAPIMutationGuard: a denied mutation is not sent; an allowed
// one is, and the guard sees names only, never variable values or the
// token. The guard runs with an own key too.
func TestLinearAPIMutationGuard(t *testing.T) {
	q := `# make one
mutation Make($t: String!) { made: issueCreate(input: {title: $t, note: "x { y"}) { success } commentCreate(input: {}) { success } }`
	for _, tc := range []struct {
		decision, key string
		sent          bool
	}{
		{rpc.GuardDeny, "", false},
		{rpc.GuardAskHuman, "lin_api_mine", false},
		{rpc.GuardAllow, "", true},
	} {
		t.Run(tc.decision, func(t *testing.T) {
			t.Setenv("LINEAR_API_KEY", tc.key)
			lin := &fakeLinear{}
			f := &fakeController{decision: tc.decision}
			c := linearAPICmd{Team: "ENG", Query: q, Field: []string{"t=secret title"}}
			_, _, err := runLinearAPI(t, &Globals{Socket: f.serve(t)}, lin.serve(t), c, "")
			if tc.sent {
				require.NoError(t, err)
				assert.Len(t, lin.reqs, 1)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not sent")
				assert.Empty(t, lin.reqs, "no HTTP request")
			}
			require.Len(t, f.guards, 1)
			assert.Equal(t, rpc.GuardParams{Action: config.ActionLinearAPIMutation, Target: "ENG", Detail: "Make issueCreate commentCreate"}, f.guards[0])
			require.Len(t, f.records, 1)
			rec, _ := json.Marshal(f.records[0])
			assert.NotContains(t, string(rec), "tok-")
			assert.NotContains(t, string(rec), "lin_api")
			assert.NotContains(t, string(rec), "secret")
			assert.Equal(t, "linear api mutation Make", f.records[0].Command)
			assert.Equal(t, tc.decision, f.records[0].Decision)
		})
	}
}

func TestLinearAPIErrorsExit1(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")
	lin := &fakeLinear{resp: `{"data":null,"errors":[{"message":"nope"}]}`}
	f := &fakeController{decision: rpc.GuardAllow}
	out, errOut, err := runLinearAPI(t, &Globals{Socket: f.serve(t)}, lin.serve(t), linearAPICmd{Team: "ENG", Query: "{ viewer { id } }"}, "")
	require.ErrorIs(t, err, errGraphQL)
	assert.Equal(t, 1, exitCode(err))
	assert.Contains(t, out, `"nope"`, "full body on stdout")
	assert.Contains(t, errOut, "nope")
	assert.Equal(t, 1, f.records[0].ExitCode)
}

func TestParseOperation(t *testing.T) {
	for doc, want := range map[string]operation{
		"{ viewer { id } }":           {kind: "query", fields: []string{"viewer"}},
		"query Q { a b(x: 1) { c } }": {kind: "query", name: "Q", fields: []string{"a", "b"}},
		"fragment F on Issue { id }\nmutation { x: issueUpdate { success } }": {kind: "mutation", fields: []string{"issueUpdate"}},
		`mutation M($a: [String!] = ["{"]) { a { ...F } }`:                    {kind: "mutation", name: "M", fields: []string{"a"}},
	} {
		got, err := parseOperation(doc)
		require.NoError(t, err, doc)
		assert.Equal(t, want, got, doc)
	}
	_, err := parseOperation("subscription { x }")
	assert.ErrorContains(t, err, "subscription")
	_, err = parseOperation("# only a comment")
	assert.Error(t, err)
}
