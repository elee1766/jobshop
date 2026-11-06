package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestGHExtra (JOB-113): the guard event carries the parsed write
// arguments, not only the body.
func TestGHExtra(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.json")
	require.NoError(t, os.WriteFile(in, []byte(`{"name":"x"}`), 0o600))
	cases := []struct {
		args  []string
		stdin []byte
		want  ghExtra
	}{
		{
			args: []string{"pr", "edit", "5", "--add-reviewer", "elee1766,bob", "--add-reviewer=carol", "--remove-label", "wip", "--add-label", "ready", "--add-assignee", "@me", "-t", "new"},
			want: ghExtra{
				Command: "pr edit", Title: "new",
				ReviewersAdded: []string{"elee1766", "bob", "carol"},
				LabelsAdded:    []string{"ready"}, LabelsRemoved: []string{"wip"},
				AssigneesAdded: []string{"@me"},
			},
		},
		{
			args: []string{"pr", "create", "-r", "elee1766", "-rteam/x", "--reviewer", "dan", "-l", "a,b", "-a", "me", "--title", "t", "-f"},
			want: ghExtra{
				Command: "pr create", Title: "t",
				ReviewersAdded: []string{"dan", "elee1766", "team/x"},
				LabelsAdded:    []string{"a", "b"}, AssigneesAdded: []string{"me"},
			},
		},
		{
			// -r is --rebase in pr merge: not a reviewer.
			args: []string{"pr", "merge", "-r", "7"},
			want: ghExtra{Command: "pr merge"},
		},
		{
			args: []string{"issue", "edit", "3", "--remove-assignee", "x", "--remove-reviewer", "y"},
			want: ghExtra{Command: "issue edit", AssigneesRemoved: []string{"x"}, ReviewersRemoved: []string{"y"}},
		},
		{
			args:  []string{"api", "-f", "reviewers[]=elee1766", "repos/o/n/pulls/5/requested_reviewers", "--raw-field=a=b", "-Fn=@-", "--field", "c=1"},
			stdin: []byte("from stdin"),
			want: ghExtra{
				Command: "api", Method: "POST", Path: "repos/o/n/pulls/5/requested_reviewers",
				Fields: []ghField{{"reviewers[]", "elee1766"}, {"a", "b"}, {"n", "from stdin"}, {"c", "1"}},
			},
		},
		{
			args: []string{"api", "-X", "patch", "repos/o/n", "--input", in},
			want: ghExtra{Command: "api", Method: "PATCH", Path: "repos/o/n", Input: `{"name":"x"}`},
		},
	}
	for _, tc := range cases {
		name := strings.Join(tc.args, " ")
		c := classifyGH(tc.args)
		require.NotEmpty(t, c.action, name)
		var got ghExtra
		require.NoError(t, json.Unmarshal(buildGHExtra(c, tc.args, tc.stdin), &got), name)
		tc.want.Argv = tc.args
		assert.Equal(t, tc.want, got, name)
	}
}

// TestClassifyGHAPIPath: a -f before the endpoint is not the endpoint.
func TestClassifyGHAPIPath(t *testing.T) {
	c := classifyGH([]string{"api", "-f", "body=x", "repos/o/n/issues/1/comments"})
	assert.Equal(t, "repos/o/n/issues/1/comments", c.apiPath)
	assert.Equal(t, "o/n", c.repo)
	assert.Equal(t, "api-post", c.action)
}

// TestGHGuardExtra: runGH sends the extra on the guard call, so a hook can
// refuse a review request to a given person.
func TestGHGuardExtra(t *testing.T) {
	dir := t.TempDir()
	gh, _ := tokenGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "set")
	f := &fakeController{decision: rpc.GuardAllow}
	g := &Globals{Socket: f.serve(t)}
	var out, errb bytes.Buffer
	args := []string{"pr", "edit", "5", "-R", "o/n", "--add-reviewer", "elee1766"}
	assert.Equal(t, 5, runGH(g, args, nil, &out, &errb), errb.String())
	f.mu.Lock()
	defer f.mu.Unlock()
	require.Len(t, f.guards, 1)
	assert.Equal(t, "pr-edit", f.guards[0].Action)
	assert.Equal(t, "o/n#5", f.guards[0].Target)
	var e ghExtra
	require.NoError(t, json.Unmarshal(f.guards[0].Extra, &e))
	assert.Equal(t, args, e.Argv)
	assert.Equal(t, []string{"elee1766"}, e.ReviewersAdded)
}
