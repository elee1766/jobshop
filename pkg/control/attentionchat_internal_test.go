package control

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// TestAttentionSummaryFromChat: an attention chat's first reply becomes
// the item's row summary (#606, #607), keeping meta.desk.attention_chat.
func TestAttentionSummaryFromChat(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	q := mkTicket(t, c, store.NewTicket{Title: "which branch?", Body: "the body", Labels: []string{questionLabel}})
	key := attentionKey("question", q.ID)
	require.NoError(t, c.recordAttentionChat(ctx, q.ID, key))
	sess := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Title: "s", Fields: json.RawMessage(`{"agent":"fake","key":"` + key + `"}`)})

	c.recordAttentionSummary(ctx, sess.ID, "  It asks which branch.\n\nRun jobshop question answer.  ")
	got, err := c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, "It asks which branch. Run jobshop question answer.", metaString(got.Meta, deskMetaPack, "attention_summary"))
	assert.Equal(t, key, metaString(got.Meta, deskMetaPack, "attention_chat"))
	items, err := c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "It asks which branch. Run jobshop question answer.", items[0].Summary)

	linked := strings.Repeat("Context before the link. ", 15) + "\n\n[Ticket](jobshop://ticket/" + q.ID + ") and [Issue](https://github.com/acme/widget/issues/9)."
	c.recordAttentionSummary(ctx, sess.ID, linked)
	items, err = c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	assert.Equal(t, linked, items[0].Context)
	assert.LessOrEqual(t, len(items[0].Summary), maxSummary+3)
	got, err = c.store.Get(ctx, q.ID)
	require.NoError(t, err)

	// Not an attention chat: nothing is written.
	other := mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Title: "s2", Fields: json.RawMessage(`{"agent":"fake","key":"desk:abc"}`)})
	c.recordAttentionSummary(ctx, other.ID, "hello")
	got2, err := c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, got.Revision, got2.Revision)
}
