package control

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// TestBackfillGitHubRepo: a run, its step and its escalation from before
// #183 get meta.github.repo from the workflow config; other meta stays,
// and a run of an unknown workflow is left alone.
func TestBackfillGitHubRepo(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	run := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"work","state":"running"}`), Meta: json.RawMessage(`{"github":{"issue":7}}`),
	})
	step := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindStep, Title: "s", Pool: "dev", ParentID: run.ID,
		Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"a","index":0}`), Meta: json.RawMessage(`{"github":{"issue":7}}`),
	})
	esc := mkTicket(t, c, store.NewTicket{
		Title: "e",
		Meta:  json.RawMessage(`{"github":{"issue":7},"run":{"id":"` + run.ID + `","workflow":"work","escalation":"x"}}`),
	})
	gone := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindRun, Title: "g",
		Fields: json.RawMessage(`{"workflow":"removed","state":"running"}`), Meta: json.RawMessage(`{"github":{"issue":8}}`),
	})

	require.NoError(t, c.backfillGitHubRepo(ctx))
	require.NoError(t, c.backfillGitHubRepo(ctx), "idempotent")

	get := func(id string) domain.Ticket {
		tk, err := c.store.Get(ctx, id)
		require.NoError(t, err)
		return tk
	}
	assert.JSONEq(t, `{"github":{"repo":"acme/widget","issue":7}}`, string(get(run.ID).Meta))
	assert.JSONEq(t, `{"github":{"repo":"acme/widget","issue":7}}`, string(get(step.ID).Meta))
	assert.JSONEq(t, `{"github":{"repo":"acme/widget","issue":7},"run":{"id":"`+run.ID+`","workflow":"work","escalation":"x"}}`, string(get(esc.ID).Meta))
	assert.Equal(t, run.Revision+1, get(run.ID).Revision, "second pass writes nothing")
	assert.JSONEq(t, `{"github":{"issue":8}}`, string(get(gone.ID).Meta))
}
