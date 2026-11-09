package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// createLinear is a fake Linear for create: issues land in state unless
// the input sets a stateId.
func createLinear(t *testing.T, state string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req linearReq
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &req)
		switch q := req.Query; {
		case strings.Contains(q, "issueCreate"):
			s := state
			if in, _ := req.Variables["input"].(map[string]any); in["stateId"] == "st-todo" {
				s = "Todo"
			}
			_, _ = io.WriteString(w, `{"data":{"issueCreate":{"success":true,"issue":{"id":"iss-9","identifier":"JOB-9","url":"u","state":{"name":"`+s+`"}}}}}`)
		case strings.Contains(q, "states("):
			_, _ = io.WriteString(w, `{"data":{"team":{"states":{"nodes":[{"id":"st-todo","name":"Todo","type":"unstarted"},{"id":"st-backlog","name":"Backlog","type":"backlog"}],"pageInfo":{"hasNextPage":false}}}}}`)
		case strings.Contains(q, "team(id"):
			_, _ = io.WriteString(w, `{"data":{"team":{"id":"team-1","key":"JOB"}}}`)
		default:
			_, _ = io.WriteString(w, `{"errors":[{"message":"unexpected query"}]}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestLinearCreateIntakeNote (JOB-121): create without --state notes on
// stderr when the issue lands outside the team's intake states; it says
// nothing when the issue is in one, when --state is set, or when no
// intake signal watches the team.
func TestLinearCreateIntakeNote(t *testing.T) {
	todo := &rpc.LinearIntakeResult{Watched: true, States: []string{"Todo"}}
	for _, tc := range []struct {
		name   string
		landed string
		state  string
		intake *rpc.LinearIntakeResult
		note   bool
	}{
		{name: "backlog", landed: "Backlog", intake: todo, note: true},
		{name: "todo default", landed: "Todo", intake: todo},
		{name: "state set", landed: "Backlog", state: "Todo", intake: todo},
		{name: "not watched", landed: "Backlog", intake: &rpc.LinearIntakeResult{}},
		{name: "any state", landed: "Backlog", intake: &rpc.LinearIntakeResult{Watched: true}},
		{name: "old controller", landed: "Backlog"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeController{decision: domain.GuardAllow, keys: map[string]string{}, intake: tc.intake}
			g := &Globals{Socket: f.serve(t)}
			l := &linearCmd{APIURL: createLinear(t, tc.landed)}
			var out, errb bytes.Buffer
			oldOut, oldErr := stdout, stderr
			stdout, stderr = &out, &errb
			t.Cleanup(func() { stdout, stderr = oldOut, oldErr })

			err := linearCreateCmd{Team: "JOB", Title: "t", State: tc.state}.Run(g, l)
			require.NoError(t, err)
			assert.Contains(t, out.String(), `"identifier": "JOB-9"`)
			if !tc.note {
				assert.Empty(t, errb.String())
				return
			}
			assert.Equal(t, "note: JOB-9 is in Backlog; Linear intake only takes Todo (linear_intake_states), so no run starts until it moves there (jobshop linear state JOB-9 \"Todo\")\n", errb.String())
			assert.Contains(t, out.String(), `"state": "Backlog"`)
		})
	}
}
