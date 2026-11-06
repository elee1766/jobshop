package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// intake is a stall trigger used for intake with key_prefix review
// (JOB-95); plain watches the same signal with the default prefix. intake
// sorts first, so it wins a subject both see.
const stallPrefixYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plan: {agent: fake, dir: %[1]q}
workflows:
  pr-review:
    vars: [subject]
    steps: [{id: review, pool: plan, retries: 0, title: "Review {{.subject}}", prompt: p}]
  plain-check:
    vars: [subject]
    steps: [{id: check, pool: plan, retries: 0, title: "Plain {{.subject}}", prompt: p}]
triggers:
  intake:
    when:
      stall:
        - {signal: review_requested, check: 'cat %[1]s/reviews'}
    run: {workflow: pr-review, key_prefix: review, vars: {subject: "{{.stall_subject}}"}}
    cooldown: 10m
    max_runs: 2
  plain:
    when:
      stall:
        - {signal: review_requested, check: 'cat %[1]s/plain'}
    run: {workflow: plain-check, vars: {subject: "{{.stall_subject}}"}}
`

// TestStallKeyPrefix (JOB-95): a key_prefix trigger keys and titles its
// runs review:<signal>:<subject>, they are still stall runs, and
// cooldown and max_runs apply to them.
func TestStallKeyPrefix(t *testing.T) {
	t.Parallel()
	c, clk, dir := stallControllerYAML(t, stallPrefixYAML)
	ctx := t.Context()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc review requested\n"), 0o600))

	res, err := c.triggerList(ctx, scope{admin: true}, struct{}{})
	require.NoError(t, err)
	for _, in := range res.([]rpc.TriggerInfo) {
		if in.Name == "intake" {
			assert.Equal(t, []string{"review:review_requested:5@abc"}, in.Stalls)
		}
	}

	c.runTriggers(t.Context())
	rs := triggerTickets(t, c, "intake")
	require.Len(t, rs, 1)
	assert.Equal(t, []string{"review:review_requested:5@abc"}, stallRunKeys(t, c, "intake"))
	assert.Equal(t, "Run pr-review review:review_requested:5@abc", rs[0].Title)
	assert.Equal(t, "intake/review:review_requested:5@abc", metaString(rs[0].Meta, triggerPack, "key"))
	assert.Equal(t, "review_requested:5@abc", metaString(rs[0].Meta, triggerPack, stallMetaKey))
	assert.True(t, isStallRun(rs[0]), "a prefixed run is a stall run")

	c.runTriggers(t.Context())
	assert.Len(t, triggerTickets(t, c, "intake"), 1, "open run holds the key")

	closeRun := func(run store.Ticket) {
		t.Helper()
		steps, err := store.Children(ctx, c.store, run.ID)
		require.NoError(t, err)
		_, err = store.Close(ctx, c.store, "admin", steps[0].ID, nil, "done")
		require.NoError(t, err)
		c.completeRun(ctx, run.ID)
	}
	closeRun(rs[0])
	c.runTriggers(t.Context())
	assert.Len(t, triggerTickets(t, c, "intake"), 1, "cooldown")
	clk.advance(11 * time.Minute)
	c.runTriggers(t.Context())
	rs, err = c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "name", MetaValue: "intake", Order: store.OrderNewest})
	require.NoError(t, err)
	require.Len(t, rs, 2, "past the cooldown")
	closeRun(rs[0])
	clk.advance(11 * time.Minute)
	c.runTriggers(t.Context())
	assert.Len(t, triggerTickets(t, c, "intake"), 2, "max_runs 2")
}

// TestStallKeyPrefixHold (JOB-95, #704): the cross-trigger hold is on the
// signal and subject, so a prefixed trigger and a default one share it,
// and a run from before meta.trigger.stall still holds a prefixed trigger.
func TestStallKeyPrefixHold(t *testing.T) {
	t.Parallel()
	c, _, dir := stallControllerYAML(t, stallPrefixYAML)
	ctx := t.Context()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plain"), []byte("5@abc\n9@fed\n"), 0o600))

	c.runTriggers(t.Context())
	assert.Equal(t, []string{"review:review_requested:5@abc"}, stallRunKeys(t, c, "intake"))
	assert.Equal(t, []string{"stall:review_requested:9@fed"}, stallRunKeys(t, c, "plain"),
		"5@abc is held by intake's open run under another prefix")
	res, err := c.triggerList(ctx, scope{admin: true}, struct{}{})
	require.NoError(t, err)
	for _, in := range res.([]rpc.TriggerInfo) {
		if in.Name == "plain" {
			assert.Contains(t, in.Why, "stall:review_requested:5@abc: run ")
		}
	}

	// A stall run started before meta.trigger.stall: only the default key.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc\n7@old\n"), 0o600))
	old, err := c.startRun(ctx, scope{admin: true, actor: "admin"},
		rpc.RunStartParams{Workflow: "plain-check", Vars: map[string]string{"subject": "7@old"}, Key: "stall:review_requested:7@old"},
		map[string]any{triggerPack: map[string]string{"name": "gone", "key": "gone/stall:review_requested:7@old"}})
	require.NoError(t, err)
	require.True(t, isStallRun(old.(rpc.RunResult).Run))
	c.runTriggers(t.Context())
	assert.Equal(t, []string{"review:review_requested:5@abc"}, stallRunKeys(t, c, "intake"), "the old run holds 7@old")
}
