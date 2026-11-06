package control

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// fakeGH answers `gh pr list` from prs.json in the config dir, filtered by
// --label, with --jq applied by jq as gh would. Everything else gets [].
const fakeGH = `#!/bin/sh
dir=%q
sub="$1 $2"
label= q=.
while [ $# -gt 0 ]; do
  case "$1" in
    --label) label="$2"; shift ;;
    --jq) q="$2"; shift ;;
  esac
  shift
done
if [ "$sub" = "pr list" ]; then
  jq --arg l "$label" '[.[] | select($l == "" or any(.labels[].name; . == $l))]' "$dir/prs.json" | jq -r "$q"
else
  echo '[]' | jq -r "$q"
fi
`

// TestReleasePRTrigger (#698): packs/github's release-pr trigger starts
// exactly one release-merge run for a green PR labeled release_pr_label,
// however many ticks see it, and pr_unmerged flags only the unlabeled PR.
func TestReleasePRTrigger(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH")
	}
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	yaml := "projects:\n  js: {}\nagents:\n  fake: {command: /bin/false}\npacks:\n  - path: " + pack +
		"\n    vars: {repo: acme/widget, checkout: %[1]q, agent: fake, human: wile}\n"
	c, _, dir := stallControllerYAML(t, yaml)

	bin := filepath.Join(dir, "bin")
	require.NoError(t, os.Mkdir(bin, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), fmt.Appendf(nil, fakeGH, dir), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	green := `{"conclusion": "SUCCESS"}`
	prs := `[
  {"number": 7, "isDraft": false, "updatedAt": "2020-01-01T00:00:00Z", "author": {"login": "app/release-please"},
   "labels": [{"name": "autorelease: pending"}], "statusCheckRollup": [` + green + `]},
  {"number": 8, "isDraft": false, "updatedAt": "2020-01-01T00:00:00Z", "author": {"login": "bot"},
   "labels": [], "statusCheckRollup": [` + green + `]}
]`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "prs.json"), []byte(prs), 0o600))

	found, err := c.detectStalls(t.Context(), c.config().Triggers["pr-unmerged"])
	require.NoError(t, err)
	var keys []string
	for _, s := range found {
		keys = append(keys, s.key())
	}
	assert.Contains(t, keys, "stall:pr_unmerged:8")
	assert.NotContains(t, keys, "stall:pr_unmerged:7", "the release PR is the trigger's, not a stall")

	releaseRuns := func() []store.Ticket {
		runs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindRun, Workflow: "release-merge"})
		require.NoError(t, err)
		return runs
	}
	for range 3 {
		c.runTriggers()
	}
	runs := releaseRuns()
	require.Len(t, runs, 1, "one run over repeated ticks")
	var gh struct {
		Repo string
		PR   int
	}
	require.NoError(t, jsonField(runs[0].Meta, "github", &gh))
	assert.Equal(t, "acme/widget", gh.Repo)
	assert.Equal(t, 7, gh.PR)
	steps, err := c.store.List(t.Context(), store.Filter{ParentID: runs[0].ID})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "refinery", steps[0].Pool)

	// Closed, the key still dedups: a PR left open after its run does
	// not start a second one.
	for _, s := range append(steps, runs[0]) {
		_, err = store.Close(t.Context(), c.store, "admin", s.ID, nil, "done")
		require.NoError(t, err)
	}
	c.runTriggers()
	assert.Len(t, releaseRuns(), 1, "same key after close")
}
