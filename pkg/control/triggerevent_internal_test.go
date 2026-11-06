package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

const eventTriggerYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q}
triggers:
  green:
    on:
      - github: {event: workflow_run, action: completed, workflow: ci, branch: master, conclusion: success}
    command:
      run: echo deploy {{.event.head_sha}} >> deployed
  red:
    on:
      - github: {event: workflow_run, action: completed, workflow: ci, conclusion: failure}
    fallback: 10m
    when:
      check: cat red 2>/dev/null | grep .
    ticket: {title: "master red", pool: dev}
  broken:
    on:
      - github: {event: push, branch: broken}
    command:
      run: echo nope; exit 3
`

// workflowRun is an admitted workflow_run delivery.
func workflowRun(delivery, branch, conclusion, sha string) store.InboxItem {
	p, _ := json.Marshal(map[string]any{
		"action": "completed",
		"workflow_run": map[string]any{
			"name": "CI", "path": ".github/workflows/ci.yml", "head_branch": branch,
			"head_sha": sha, "conclusion": conclusion,
		},
		"repository": map[string]any{"full_name": "acme/app"},
	})
	return store.InboxItem{Source: "github", DeliveryID: delivery, Kind: "workflow_run", RepoOrTeam: "acme/app", Payload: p}
}

func triggerTickets(t *testing.T, c *Controller, name string) []store.Ticket {
	t.Helper()
	ts, err := c.store.List(t.Context(), store.Filter{MetaPack: triggerPack, MetaKey: "name", MetaValue: name})
	require.NoError(t, err)
	return ts
}

// fireEvents runs what triggerEvents runs in the background, in line.
func fireEvents(c *Controller, items ...store.InboxItem) {
	trs := c.config().Triggers
	for _, it := range items {
		ev := eventOf(it)
		for name, tr := range trs {
			if len(tr.On) > 0 && triggerOn(tr, ev) {
				c.fireOnEvent(c.ctx, name, tr, ev)
			}
		}
	}
}

// TestEventTrigger (JOB-62): a matching event fires once per commit,
// a repeat does nothing, a non-matching one does nothing, and the
// command sees the event fields.
func TestEventTrigger(t *testing.T) {
	t.Parallel()
	c, _, dir := stallControllerYAML(t, eventTriggerYAML)

	fireEvents(c, workflowRun("d1", "feature", "success", "aaa"))
	assert.Empty(t, triggerTickets(t, c, "green"), "other branch")

	fireEvents(c, workflowRun("d2", "master", "success", "bbb"), workflowRun("d3", "master", "success", "bbb"))
	ts := triggerTickets(t, c, "green")
	require.Len(t, ts, 1, "a redelivery of the same commit fires once")
	assert.Equal(t, store.StatusClosed, ts[0].Status)
	b, err := os.ReadFile(filepath.Join(dir, "deployed"))
	require.NoError(t, err)
	assert.Equal(t, "deploy bbb\n", string(b))

	fireEvents(c, workflowRun("d4", "master", "success", "ccc"))
	assert.Len(t, triggerTickets(t, c, "green"), 2, "a new commit fires again")
	assert.Empty(t, triggerTickets(t, c, "red"), "success does not match red")
}

// TestEventTriggerFallback: an event and the fallback poll of the same
// thing fire once (the check key dedups); the poll runs only every
// fallback; with no event the poll still fires.
func TestEventTriggerFallback(t *testing.T) {
	t.Parallel()
	c, clk, dir := stallControllerYAML(t, eventTriggerYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "red"), []byte("sha1\n"), 0o600))
	c.root = config.Root{Dir: dir}

	fireEvents(c, workflowRun("d1", "master", "failure", "sha1"))
	require.Len(t, triggerTickets(t, c, "red"), 1, "event fires")
	c.runTriggers()
	assert.Len(t, triggerTickets(t, c, "red"), 1, "the poll of the same key does not fire again")

	// Close it; a new failure arrives with no event. The poll waits for
	// the fallback interval.
	red := triggerTickets(t, c, "red")[0]
	_, err := store.Close(t.Context(), c.store, "admin", red.ID, nil, "done")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "red"), []byte("sha2\n"), 0o600))
	c.runTriggers()
	assert.Len(t, triggerTickets(t, c, "red"), 1, "within fallback: no poll")
	clk.advance(11 * time.Minute)
	c.runTriggers()
	assert.Len(t, triggerTickets(t, c, "red"), 2, "fallback poll fires without an event")
}

// TestEventTriggerCommandFails: a failing command leaves one open
// needs-human record that holds the trigger.
func TestEventTriggerCommandFails(t *testing.T) {
	t.Parallel()
	c, _, _ := stallControllerYAML(t, eventTriggerYAML)
	push := func(id, sha string) store.InboxItem {
		p, _ := json.Marshal(map[string]any{"ref": "refs/heads/broken", "after": sha})
		return store.InboxItem{Source: "github", DeliveryID: id, Kind: "push", RepoOrTeam: "acme/app", Payload: p}
	}
	fireEvents(c, push("p1", "s1"))
	ts := triggerTickets(t, c, "broken")
	require.Len(t, ts, 1)
	assert.Equal(t, store.StatusOpen, ts[0].Status)
	assert.Contains(t, ts[0].Labels, needsHumanLabel)
	fireEvents(c, push("p2", "s2"))
	assert.Len(t, triggerTickets(t, c, "broken"), 1, "held while the failure is open")
	c.runTriggers()
	assert.Len(t, triggerTickets(t, c, "broken"), 1, "no when: never polled")
}

func TestEventMatch(t *testing.T) {
	t.Parallel()
	ev := eventOf(workflowRun("d", "master", "success", "abc"))
	assert.Equal(t, "abc", ev["head_sha"])
	m := config.EventMatch{Event: "workflow_run", Workflow: "ci", Branch: "Master", Conclusion: "success", Repo: "acme/app"}
	assert.True(t, ev.matches("github", m), "workflow by file name, case-insensitive")
	m.Workflow = "CI"
	assert.True(t, ev.matches("github", m), "workflow by name")
	assert.False(t, ev.matches("linear", m))
	m.Label = "x"
	assert.False(t, ev.matches("github", m))
}
