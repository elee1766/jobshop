package control

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
// questions, unescalated needs_human runs and quarantined pools, oldest
// first, with each escalated run or pool once (#185).
func TestAttention(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
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
	q := mkTicket(t, c, store.NewTicket{Title: "which db?", Labels: []string{questionLabel},
		Meta: json.RawMessage(`{"linear":{"workspace":"acme","issue":"ENG-5"},"github":{"repo":"acme/widget"}}`)})
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
	assert.Equal(t, "https://github.com/acme/widget/issues/7", items[2].URL)
	assert.Equal(t, "ENG-5", items[3].Ref)
	assert.Equal(t, "https://linear.app/acme/issue/ENG-5", items[3].URL)
}

// TestAttentionSummary: a row's summary is meta.desk.attention_summary
// when set, else the body's first lines, clipped (#607).
func TestAttentionSummary(t *testing.T) {
	t.Parallel()
	c, clk, _ := stallController(t)
	q := mkTicket(t, c, store.NewTicket{Title: "which db?", Body: "\nPostgres or SQLite?\n\nmore context\nand more", Labels: []string{questionLabel}})
	clk.advance(time.Minute)
	d := mkTicket(t, c, store.NewTicket{
		Title: "Proposed target", Body: "long proposal", Labels: []string{proposedLabel},
		Meta: json.RawMessage(`{"desk":{"attention_summary":"Approve v0.5: it closes 3 bugs."}}`),
	})
	clk.advance(time.Minute)
	long := mkTicket(t, c, store.NewTicket{Title: "long", Body: strings.Repeat("é", 300), Labels: []string{needsHumanLabel}})
	items, err := c.attentionItems(t.Context())
	require.NoError(t, err)
	got := map[string]string{}
	for _, it := range items {
		got[it.ID] = it.Summary
	}
	assert.Equal(t, "Postgres or SQLite? more context", got[q.ID])
	assert.Equal(t, "Approve v0.5: it closes 3 bugs.", got[d.ID])
	assert.True(t, strings.HasSuffix(got[long.ID], "..."))
	assert.LessOrEqual(t, len(got[long.ID]), maxSummary+3)
	assert.True(t, utf8.ValidString(got[long.ID]))
}
