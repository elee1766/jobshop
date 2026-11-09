package github_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/integrations/github"
)

type fakePulls struct {
	prs     map[int]github.Pull
	commits map[string][]int
}

func (f fakePulls) PullRequest(_ context.Context, _ string, n int) (github.Pull, error) {
	pr, ok := f.prs[n]
	if !ok {
		return github.Pull{}, fmt.Errorf("no PR %d", n)
	}
	return pr, nil
}

func (f fakePulls) CommitPulls(_ context.Context, _, sha string) ([]int, error) {
	return f.commits[sha], nil
}

func TestPullReads(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/o/r/pulls/7":
			_, _ = io.WriteString(w, `{"number":7,"title":"t","merged":true,"merge_commit_sha":"abc","mergeable_state":"dirty","html_url":"u"}`)
		case "/repos/o/r/commits/abc/pulls":
			_, _ = io.WriteString(w, `[{"number":7},{"number":9}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	cl := github.New(srv.URL, apicall.Static("tok"))

	pr, err := cl.PullRequest(t.Context(), "o/r", 7)
	require.NoError(t, err)
	assert.Equal(t, github.Pull{Number: 7, Title: "t", Merged: true, MergeCommitSHA: "abc", MergeableState: "dirty", HTMLURL: "u"}, pr)

	nums, err := cl.CommitPulls(t.Context(), "o/r", "abc")
	require.NoError(t, err)
	assert.Equal(t, []int{7, 9}, nums)

	_, err = cl.PullRequest(t.Context(), "o/r", 8)
	assert.True(t, github.IsNotFound(err))
}

func TestIsRevert(t *testing.T) {
	t.Parallel()
	assert.True(t, github.IsRevert(github.Pull{Title: `Revert "x"`}))
	assert.True(t, github.IsRevert(github.Pull{Body: "This reverts commit abcdef1."}))
	assert.True(t, github.IsRevert(github.Pull{Body: "Reverts o/r#4"}))
	assert.False(t, github.IsRevert(github.Pull{Title: "fix", Body: "plain"}))
}

func TestRevertedPRByNumber(t *testing.T) {
	t.Parallel()
	const repo = "elee1766/jobshop"
	src := fakePulls{prs: map[int]github.Pull{
		5: {Number: 5, Merged: true, MergeCommitSHA: "x"},
		6: {Number: 6, Merged: false},
	}}
	got, err := github.RevertedPR(t.Context(), src, repo, github.Pull{Number: 9, Body: "Reverts elee1766/jobshop#5"})
	require.NoError(t, err)
	assert.Equal(t, 5, got.Number)
	got, err = github.RevertedPR(t.Context(), src, repo, github.Pull{Number: 9, Body: "Reverts #6"})
	require.NoError(t, err)
	assert.Zero(t, got.Number, "an unmerged PR was not reverted")
	got, err = github.RevertedPR(t.Context(), src, repo, github.Pull{Number: 9, Body: "Reverts other/repo#5"})
	require.NoError(t, err)
	assert.Zero(t, got.Number, "another repo's PR")
}

func TestRevertedPRByCommit(t *testing.T) {
	t.Parallel()
	src := fakePulls{
		prs: map[int]github.Pull{
			3: {Number: 3, Merged: true, MergeCommitSHA: "AAAA1111bbbb"},
			9: {Number: 9, Merged: true, MergeCommitSHA: "other"},
		},
		commits: map[string][]int{"aaaa111": {9, 3}},
	}
	got, err := github.RevertedPR(t.Context(), src, "o/r", github.Pull{Number: 9, Body: "This reverts commit AAAA111."})
	require.NoError(t, err)
	assert.Equal(t, 3, got.Number, "skips the reverting PR and matches the merge commit prefix")
}
