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

// intake is a scan trigger used for intake with key_prefix review
// (JOB-95); plain watches the same signal with the default prefix. intake
// sorts first, so it wins a subject both see.
const scanPrefixYAML = `projects:
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
      scan:
        - {signal: review_requested, check: 'cat %[1]s/reviews'}
    run: {workflow: pr-review, key_prefix: review, vars: {subject: "{{.subject}}"}}
    cooldown: 10m
    max_runs: 2
  plain:
    when:
      scan:
        - {signal: review_requested, check: 'cat %[1]s/plain'}
    run: {workflow: plain-check, vars: {subject: "{{.subject}}"}}
`

// TestScanKeyPrefix (JOB-95): a key_prefix trigger keys and titles its
// runs review:<signal>:<subject>, they are still scan runs, and
// cooldown and max_runs apply to them.
func TestScanKeyPrefix(t *testing.T) {
	t.Parallel()
	c, clk, dir := scanControllerYAML(t, scanPrefixYAML)
	ctx := t.Context()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc review requested\n"), 0o600))

	res, err := c.triggerList(ctx, scope{admin: true}, struct{}{})
	require.NoError(t, err)
	for _, in := range res.([]rpc.TriggerInfo) {
		if in.Name == "intake" {
			assert.Equal(t, []string{"review:review_requested:5@abc"}, in.Findings)
		}
	}

	c.runTriggers(t.Context(), c.config())
	rs := triggerTickets(t, c, "intake")
	require.Len(t, rs, 1)
	assert.Equal(t, []string{"review:review_requested:5@abc"}, scanRunKeys(t, c, "intake"))
	assert.Equal(t, "Run pr-review review:review_requested:5@abc", rs[0].Title)
	assert.Equal(t, "intake/review:review_requested:5@abc", metaString(rs[0].Meta, triggerPack, "key"))
	assert.Equal(t, "review_requested:5@abc", metaString(rs[0].Meta, triggerPack, findingMetaKey))
	assert.True(t, isScanRun(rs[0]), "a prefixed run is a scan run")

	c.runTriggers(t.Context(), c.config())
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
	c.runTriggers(t.Context(), c.config())
	assert.Len(t, triggerTickets(t, c, "intake"), 1, "cooldown")
	clk.advance(11 * time.Minute)
	c.runTriggers(t.Context(), c.config())
	rs, err = c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "name", MetaValue: "intake", Order: store.OrderNewest})
	require.NoError(t, err)
	require.Len(t, rs, 2, "past the cooldown")
	closeRun(rs[0])
	clk.advance(11 * time.Minute)
	c.runTriggers(t.Context(), c.config())
	assert.Len(t, triggerTickets(t, c, "intake"), 2, "max_runs 2")
}

// TestScanKeyPrefixHold (JOB-95, #704): the cross-trigger hold is on the
// signal and subject, so a prefixed trigger and a default one share it.
func TestScanKeyPrefixHold(t *testing.T) {
	t.Parallel()
	c, _, dir := scanControllerYAML(t, scanPrefixYAML)
	ctx := t.Context()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviews"), []byte("5@abc\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plain"), []byte("5@abc\n9@fed\n"), 0o600))

	c.runTriggers(t.Context(), c.config())
	assert.Equal(t, []string{"review:review_requested:5@abc"}, scanRunKeys(t, c, "intake"))
	assert.Equal(t, []string{"scan:review_requested:9@fed"}, scanRunKeys(t, c, "plain"),
		"5@abc is held by intake's open run under another prefix")
	res, err := c.triggerList(ctx, scope{admin: true}, struct{}{})
	require.NoError(t, err)
	for _, in := range res.([]rpc.TriggerInfo) {
		if in.Name == "plain" {
			assert.Contains(t, in.Why, "scan:review_requested:5@abc: run ")
		}
	}
}
