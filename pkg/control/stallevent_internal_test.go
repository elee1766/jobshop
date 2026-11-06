package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// review is a stall trigger woken by review requests (JOB-72). Its check
// logs the PR it was asked about and lists the reviews file, narrowed to
// that PR on an event. plain has no on: and is polled every tick.
const stallEventYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plan: {agent: fake, dir: %[1]q}
workflows:
  pr-review:
    vars: [subject]
    steps: [{id: review, pool: plan, title: "Review {{.subject}}", prompt: p}]
  plain-check:
    vars: [subject]
    steps: [{id: check, pool: plan, title: "Plain {{.subject}}", prompt: p}]
triggers:
  review:
    on:
      - github: {event: pull_request, action: review_requested}
    fallback: 15m
    when:
      stall:
        - signal: review_requested
          check: 'echo "pr={{.event.pr}} branch={{.event.branch}}" >> %[1]s/seen; grep "^{{.event.pr}}" %[1]s/reviews'
    run: {workflow: pr-review, vars: {subject: "{{.stall_subject}}"}}
  plain:
    when:
      stall:
        - {signal: plain, check: 'echo x >> %[1]s/seen-plain; cat %[1]s/plain'}
    run: {workflow: plain-check, vars: {subject: "{{.stall_subject}}"}}
`

// pullRequest is an admitted pull_request delivery.
func pullRequest(delivery, action string, n int, branch, sha string) store.InboxItem {
	p, _ := json.Marshal(map[string]any{
		"action":       action,
		"number":       n,
		"pull_request": map[string]any{"head": map[string]any{"ref": branch, "sha": sha}},
	})
	return store.InboxItem{Source: "github", DeliveryID: delivery, Kind: "pull_request", RepoOrTeam: "acme/app", Number: int64(n), Payload: p}
}

func stallRunKeys(t *testing.T, c *Controller, trigger string) []string {
	t.Helper()
	var keys []string
	for _, r := range triggerTickets(t, c, trigger) {
		f, err := store.DecodeFields[store.RunFields](r)
		require.NoError(t, err)
		keys = append(keys, f.Key)
	}
	return keys
}

func readLines(t *testing.T, p string) []string {
	t.Helper()
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

// TestStallTriggerOn (JOB-72): an event runs the stall check at once with
// the event's fields, the fallback poll does not fire the same subject
// again, and the poll alone still fires a new one.
func TestStallTriggerOn(t *testing.T) {
	t.Parallel()
	c, clk, dir := stallControllerYAML(t, stallEventYAML)
	seen := filepath.Join(dir, "seen")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc review requested\n7@def review requested\n"), 0o600))

	fireEvents(c, pullRequest("d1", "closed", 5, "feat", "abc"))
	assert.Empty(t, stallRunKeys(t, c, "review"), "other action: no check")
	assert.Empty(t, readLines(t, seen))

	fireEvents(c, pullRequest("d2", "review_requested", 5, "feat;rm -rf x", "abc"))
	assert.Equal(t, []string{"stall:review_requested:5@abc"}, stallRunKeys(t, c, "review"), "the event narrows the check to its PR")
	assert.Equal(t, []string{"pr=5 branch="}, readLines(t, seen), "event fields reach the check; unsafe ones are empty")

	c.runTriggers(t.Context()) // first poll
	assert.ElementsMatch(t, []string{"stall:review_requested:5@abc", "stall:review_requested:7@def"}, stallRunKeys(t, c, "review"),
		"the poll lists every PR; 5 already has a run")
	assert.Equal(t, []string{"pr=5 branch=", "pr= branch="}, readLines(t, seen))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc\n7@def\n9@fed\n"), 0o600))
	c.runTriggers(t.Context())
	assert.Len(t, stallRunKeys(t, c, "review"), 2, "within fallback: not polled")
	assert.Len(t, readLines(t, seen), 2, "the check did not run")

	clk.advance(16 * time.Minute)
	c.runTriggers(t.Context())
	assert.Contains(t, stallRunKeys(t, c, "review"), "stall:review_requested:9@fed", "the fallback poll fires without an event")
	assert.Len(t, stallRunKeys(t, c, "review"), 3)
}

// TestStallTriggerNoOn: without on:, a stall trigger is polled every tick
// as before, and its check is run as written.
func TestStallTriggerNoOn(t *testing.T) {
	t.Parallel()
	c, _, dir := stallControllerYAML(t, stallEventYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plain"), []byte("a\n"), 0o600))
	c.runTriggers(t.Context())
	c.runTriggers(t.Context())
	c.runTriggers(t.Context())
	assert.Len(t, readLines(t, filepath.Join(dir, "seen-plain")), 3, "checked every tick")
	assert.Equal(t, []string{"stall:plain:a"}, stallRunKeys(t, c, "plain"))
	fireEvents(c, pullRequest("d1", "review_requested", 5, "feat", "abc"))
	assert.Len(t, readLines(t, filepath.Join(dir, "seen-plain")), 3, "no on: is not woken by events")
}
