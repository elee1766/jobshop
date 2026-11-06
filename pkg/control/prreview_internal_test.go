package control

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestPRReview (JOB-33): packs/github renders pr-review for a plain and a
// prefixed instance, each in its own repo and pools. pr-unmerged starts
// pr-review keyed pr-<n>; a start by hand (another instance's implement
// step) with that key gets the open run, and the trigger does not double
// a run started first by hand.
func TestPRReview(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH")
	}
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	yaml := "projects:\n  js: {}\nagents:\n  fake: {command: /bin/false}\npacks:\n" +
		"  - path: " + pack + "\n    vars: {repo: acme/widget, checkout: %[1]q, agent: fake, human: wile}\n" +
		"  - path: " + pack + "\n    prefix: other-\n    vars: {repo: acme/other, checkout: %[1]q, agent: fake, human: wile}\n"
	c, _, dir := stallControllerYAML(t, yaml)

	cfg := c.config()
	for wf, want := range map[string]struct{ repo, refinery, deploy string }{
		"pr-review":       {"acme/widget", "refinery", "deploy-check"},
		"other-pr-review": {"acme/other", "other-refinery", "other-deploy-check"},
	} {
		w, ok := cfg.Workflows[wf]
		require.True(t, ok, wf)
		assert.Equal(t, []string{"pr"}, w.Vars, wf)
		assert.Equal(t, "pr", w.PR, wf)
		assert.Equal(t, want.repo, w.Repo, wf)
		require.Len(t, w.Steps, 2, wf)
		assert.Equal(t, "review-merge", w.Steps[0].ID)
		assert.Equal(t, want.refinery, w.Steps[0].Pool, wf)
		assert.Equal(t, want.deploy, w.Steps[1].Pool, wf)
		require.NotNil(t, w.Steps[1].When, wf)
		assert.Equal(t, "merged", w.Steps[1].When.Outcome, wf)
	}
	assert.Equal(t, "other-pr-review", cfg.Triggers["other-pr-unmerged"].Run.Workflow)

	bin := filepath.Join(dir, "bin")
	require.NoError(t, os.Mkdir(bin, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), fmt.Appendf(nil, fakeGH, dir), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	green := `{"conclusion": "SUCCESS"}`
	pr := func(n int) string {
		return fmt.Sprintf(`{"number": %d, "isDraft": false, "updatedAt": "2020-01-01T00:00:00Z", "author": {"login": "bot"}, "labels": [], "statusCheckRollup": [%s]}`, n, green)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "prs.json"), []byte("["+pr(8)+","+pr(9)+"]"), 0o600))

	start := func(wf, n string) rpc.RunResult {
		res, err := c.startRun(t.Context(), c.config(), scope{admin: true, actor: "admin"},
			rpc.RunStartParams{Workflow: wf, Vars: map[string]string{"pr": n}, Key: "pr-" + n}, nil)
		require.NoError(t, err)
		return res.(rpc.RunResult)
	}
	// Another instance's implement step hands PR 9 over before the
	// trigger sees it.
	hand := start("pr-review", "9")
	require.False(t, hand.Existing)

	runs := func(wf string) map[string]string {
		rs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindRun, Workflow: wf})
		require.NoError(t, err)
		out := map[string]string{}
		for _, r := range rs {
			f, err := store.DecodeFields[store.RunFields](r)
			require.NoError(t, err)
			out[f.Key] = r.ID
		}
		return out
	}
	for range 3 {
		c.runTriggers(t.Context(), c.config())
	}
	got := runs("pr-review")
	assert.Len(t, got, 2, "one run per PR over repeated ticks")
	assert.Equal(t, hand.Run.ID, got["pr-9"], "the hand-started run is kept")
	require.Contains(t, got, "pr-8")

	again := start("pr-review", "8")
	assert.True(t, again.Existing, "a repeated start returns the open run")
	assert.Equal(t, got["pr-8"], again.Run.ID)

	r, err := c.store.Get(t.Context(), got["pr-8"])
	require.NoError(t, err)
	var gh struct {
		Repo string
		PR   int
	}
	require.NoError(t, jsonField(r.Meta, "github", &gh))
	assert.Equal(t, "acme/widget", gh.Repo)
	assert.Equal(t, 8, gh.PR)

	// The prefixed instance reviews its own repo's PRs under its own
	// workflow.
	assert.Len(t, runs("other-pr-review"), 2)
}
