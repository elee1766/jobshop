package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// labelLinear is a fake Linear that keeps one issue's labels, enough for
// update's plan, mutation and read back.
type labelLinear struct {
	mu        sync.Mutex
	labels    []string
	mutations int
}

var labelIDs = map[string]string{"lb-blocked": "Blocked", "lb-tracking": "tracking"}

func (f *labelLinear) serve(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req linearReq
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &req)
		f.mu.Lock()
		defer f.mu.Unlock()
		q := req.Query
		switch {
		case strings.Contains(q, "issueUpdate"):
			f.mutations++
			in, _ := req.Variables["input"].(map[string]any)
			for _, id := range strs(in["addedLabelIds"]) {
				if !slices.Contains(f.labels, labelIDs[id]) {
					f.labels = append(f.labels, labelIDs[id])
				}
			}
			for _, id := range strs(in["removedLabelIds"]) {
				f.labels = slices.DeleteFunc(f.labels, func(l string) bool { return l == labelIDs[id] })
			}
			_, _ = io.WriteString(w, `{"data":{"issueUpdate":{"success":true,"issue":{"id":"iss-1","identifier":"JOB-7","url":"u"}}}}`)
		case strings.Contains(q, "team(id"):
			_, _ = io.WriteString(w, `{"data":{"team":{"id":"team-1","key":"JOB",
 "states":{"nodes":[{"id":"st-todo","name":"Todo","type":"unstarted"}]},
 "labels":{"nodes":[{"id":"lb-blocked","name":"Blocked"},{"id":"lb-tracking","name":"tracking"}]},
 "members":{"nodes":[]}}}}`)
		case strings.Contains(q, "updatedAt team"):
			_, _ = io.WriteString(w, `{"data":{"issue":{"id":"iss-1","updatedAt":"2026-11-05T00:00:00.000Z","team":{"key":"JOB"}}}}`)
		case strings.Contains(q, "archivedAt"):
			nodes := make([]map[string]string, len(f.labels))
			for i, l := range f.labels {
				nodes[i] = map[string]string{"name": l}
			}
			nb, _ := json.Marshal(nodes)
			_, _ = io.WriteString(w, `{"data":{"issue":{"id":"iss-1","identifier":"JOB-7","title":"t","url":"u","state":{"id":"x","name":"Todo","type":"unstarted"},"team":{"id":"team-1","key":"JOB"},"labels":{"nodes":`+string(nb)+`},"children":{"nodes":[]},"relations":{"nodes":[]},"inverseRelations":{"nodes":[]},"comments":{"nodes":[]}}}}`)
		default:
			_, _ = io.WriteString(w, `{"errors":[{"message":"unexpected query"}]}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func strs(v any) []string {
	var out []string
	l, _ := v.([]any)
	for _, x := range l {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

// TestLinearLabel (JOB-9): add and remove a label, an unknown label
// names the team's labels, and a repeat with the same key sends one
// mutation.
func TestLinearLabel(t *testing.T) {
	f := &fakeController{decision: rpc.GuardAllow, keys: map[string]string{}}
	g := &Globals{Socket: f.serve(t)}
	lin := &labelLinear{}
	l := &linearCmd{APIURL: lin.serve(t)}
	run := func(c linearLabelCmd) (string, error) {
		var out bytes.Buffer
		old := stdout
		stdout = &out
		t.Cleanup(func() { stdout = old })
		err := c.Run(g, l)
		stdout = old
		return out.String(), err
	}

	_, err := run(linearLabelCmd{IssueKey: "JOB-7", Op: "add", Labels: []string{"blocked"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"Blocked"}, lin.labels)

	_, err = run(linearLabelCmd{IssueKey: "JOB-7", Op: "remove", Labels: []string{"Blocked"}})
	require.NoError(t, err)
	assert.Empty(t, lin.labels)

	_, err = run(linearLabelCmd{IssueKey: "JOB-7", Op: "add", Labels: []string{"nope"}})
	require.ErrorContains(t, err, `no label "nope" (have: Blocked, tracking)`)
	assert.Equal(t, 2, lin.mutations, "unknown label sends nothing")

	c := linearLabelCmd{IssueKey: "JOB-7", Op: "add", Labels: []string{"tracking"}}
	c.Key = "k1"
	_, err = run(c)
	require.NoError(t, err)
	out, err := run(c)
	require.NoError(t, err)
	assert.Equal(t, 3, lin.mutations, "same key: one mutation")
	assert.Contains(t, out, `"replayed": true`)
}
