package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func attentionKinds(t *testing.T, c *Controller) []string {
	t.Helper()
	items, err := c.attentionItems(t.Context())
	require.NoError(t, err)
	out := []string{}
	for _, it := range items {
		out = append(out, it.Kind+" "+it.ID)
	}
	return out
}

// TestAttention: empty when healthy; lists needs-human tickets,
// questions, unescalated needs_human runs, quarantined pools and a
// rollback, oldest first, with each escalated run or pool once (#185).
func TestAttention(t *testing.T) {
	t.Parallel()
	c, clk, dir := stallController(t)
	// Healthy: a running run and an ordinary ticket are not listed.
	mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "ok", Fields: json.RawMessage(`{"workflow":"work","state":"running"}`)})
	mkTicket(t, c, store.NewTicket{Title: "task"})
	assert.Empty(t, attentionKinds(t, c))

	run := mkTicket(t, c, store.NewTicket{
		Kind: store.KindRun, Title: "stuck", Project: "js",
		Fields: json.RawMessage(`{"workflow":"work","state":"needs_human"}`), Meta: json.RawMessage(`{"github":{"repo":"acme/widget","issue":7}}`),
	})
	clk.advance(time.Minute)
	bare := mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "bare", Fields: json.RawMessage(`{"workflow":"work","state":"needs_human"}`)})
	clk.advance(time.Minute)
	esc := mkTicket(t, c, store.NewTicket{
		Title: "Escalate run " + run.ID, Labels: []string{needsHumanLabel},
		Meta: json.RawMessage(`{"github":{"repo":"acme/widget","issue":7},"run":{"id":"` + run.ID + `","workflow":"work","escalation":"x"}}`),
	})
	clk.advance(time.Minute)
	q := mkTicket(t, c, store.NewTicket{Title: "which db?", Labels: []string{questionLabel}})
	clk.advance(time.Minute)
	prop := mkTicket(t, c, store.NewTicket{Title: "Proposed target: x (#9)", Labels: []string{proposedLabel}})
	clk.advance(time.Minute)
	c.poolQuarantined = map[string]string{"dev": "spawn: no such dir", "plan": "spawn: bad agent"}
	poolEsc := mkTicket(t, c, store.NewTicket{
		Title: "Pool dev quarantined", Labels: []string{needsHumanLabel},
		Meta: json.RawMessage(`{"pool":{"name":"dev","quarantine":"spawn: no such dir"}}`),
	})

	got := attentionKinds(t, c)
	// Pool dev has an escalation ticket, so it is listed once, as the
	// ticket; the escalated run likewise. plan (no escalation) is listed
	// as a pool, since the controller started (clock start).
	assert.Equal(t, []string{"pool plan", "run " + bare.ID, "ticket " + esc.ID, "question " + q.ID, "decision " + prop.ID, "ticket " + poolEsc.ID}, got)

	items, err := c.attentionItems(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "acme/widget", items[2].Repo)
	assert.Equal(t, "#7", items[2].Ref)

	// Rollback: listed while "failed" is newer than "deployed".
	state := filepath.Join(dir, "deploy")
	require.NoError(t, os.MkdirAll(state, 0o750))
	_, ok := rollbackItem(filepath.Join(state, "failed"))
	assert.False(t, ok, "no failed file")
	require.NoError(t, os.WriteFile(filepath.Join(state, "deployed"), []byte("old\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(state, "failed"), []byte("abc123\n"), 0o600))
	past := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(filepath.Join(state, "deployed"), past, past))
	it, ok := rollbackItem(filepath.Join(state, "failed"))
	require.True(t, ok)
	assert.Equal(t, rpc.AttentionItem{Kind: rpc.AttentionRollback, ID: "abc123", Title: "self-deploy rolled back abc123", Since: it.Since}, it)
	// A failed pre-switch config check names the error (#261).
	require.NoError(t, os.WriteFile(filepath.Join(state, "failed.reason"), []byte(`pools.operator.prompt_file: var "jobshop_repo" is not set`+"\n"), 0o600))
	it, ok = rollbackItem(filepath.Join(state, "failed"))
	require.True(t, ok)
	assert.Equal(t, `self-deploy skipped abc123, config check failed: pools.operator.prompt_file: var "jobshop_repo" is not set`, it.Title)
	require.NoError(t, os.WriteFile(filepath.Join(state, "deployed"), []byte("new\n"), 0o600))
	_, ok = rollbackItem(filepath.Join(state, "failed"))
	assert.False(t, ok, "a later deploy clears it")

	// Config repo (#532): listed while the failed file exists.
	repoFile := filepath.Join(dir, "config-repo.failed")
	_, ok = configRepoItem(repoFile)
	assert.False(t, ok, "no failed file")
	require.NoError(t, os.WriteFile(repoFile, []byte("abc123\npacks[0]: pack.yaml: no such file\n"), 0o600))
	it, ok = configRepoItem(repoFile)
	require.True(t, ok)
	assert.Equal(t, rpc.AttentionItem{Kind: rpc.AttentionConfigRepo, ID: "abc123", Title: "config repo commit abc123 failed its check, running the last good config: packs[0]: pack.yaml: no such file", Since: it.Since}, it)
}
