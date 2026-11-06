package control_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// prestartRoot is poolRoot with an operator chat pool and
// attention.prestart_max 1.
func prestartRoot(t *testing.T) (config.Root, func() func()) {
	root, start := poolRoot(t, "1m")
	b, err := os.ReadFile(filepath.Join(root.Dir, config.FileName))
	require.NoError(t, err)
	cfg := string(b) + "  operator:\n    project: js\n    agent: fake\n    type: chat\n    max_sessions: 3\n    dir: " + root.Dir + "\nattention:\n  prestart_max: 1\n"
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	return root, start
}

func attentionSessions(t *testing.T, admin *rpc.Client) map[string]store.Ticket {
	t.Helper()
	var all []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &all))
	out := map[string]store.Ticket{}
	for _, s := range all {
		f, err := store.DecodeFields[store.SessionFields](s)
		require.NoError(t, err)
		if strings.HasPrefix(f.Key, "attention:") {
			out[f.Key] = s
		}
	}
	return out
}

// TestAttentionPrestart (#606): an item entering attention gets one
// pre-started, read-only chat with the context prompt, up to
// prestart_max; a click attaches without a new turn; an item past the
// cap starts on click; a human message lifts read-only; resolving the
// item closes its chat.
func TestAttentionPrestart(t *testing.T) {
	root, start := prestartRoot(t)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	ctx := t.Context()
	pass(t, root.Dir, 1)

	var q1, q2 store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "which branch?", Labels: []string{"question"}}, &q1))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "which tag?", Labels: []string{"question"}}, &q2))
	k1, k2 := "attention:question:"+q1.ID, "attention:question:"+q2.ID

	pass(t, root.Dir, 1)
	sess := attentionSessions(t, admin)
	require.Len(t, sess, 1, "prestart_max 1")
	s1, ok := sess[k1]
	require.True(t, ok, "the oldest item first")
	f, err := store.DecodeFields[store.SessionFields](s1)
	require.NoError(t, err)
	assert.True(t, f.Prestart)
	assert.True(t, f.ReadOnly)
	assert.Equal(t, "operator", s1.Pool)
	assert.Equal(t, k1, metaDesk(t, getTicket(t, admin, q1.ID).Ticket))
	passEventually(t, root.Dir, func() bool {
		return strings.Contains(promptsFile(t, root), "Desk attention item: question "+q1.ID)
	}, "the first turn gets the context prompt")
	assert.Contains(t, promptsFile(t, root), "never a bare option number or letter", "a question chat relays clear answer text")

	// Another pass starts nothing new.
	pass(t, root.Dir, 2)
	assert.Len(t, attentionSessions(t, admin), 1)
	assert.Equal(t, 1, strings.Count(promptsFile(t, root), "Desk attention item:"))

	// A read-only session cannot write through its socket.
	sock := dial(t, filepath.Join(root.SessionDir(s1.ID), "cli.sock"))
	err = sock.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "x"}, nil)
	require.ErrorIs(t, err, rpc.ErrForbidden)
	require.NoError(t, sock.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: q1.ID}, &rpc.ShowResult{}))

	// A click attaches: same session, nothing sent.
	var res rpc.AttentionChatResult
	require.NoError(t, admin.Call(ctx, rpc.MethodAttentionChat, rpc.AttentionChatParams{Kind: rpc.AttentionQuestion, ID: q1.ID}, &res))
	assert.Equal(t, rpc.AttentionChatResult{Kind: "question", ID: q1.ID, Key: k1, Session: s1.ID, Pool: "operator"}, res)
	pass(t, root.Dir, 1)
	assert.Equal(t, 1, strings.Count(promptsFile(t, root), "Desk attention item:"), "no new turn on click")

	// Past the cap, the click starts the chat.
	require.NoError(t, admin.Call(ctx, rpc.MethodAttentionChat, rpc.AttentionChatParams{Kind: rpc.AttentionQuestion, ID: q2.ID}, &res))
	assert.True(t, res.Started)
	assert.Equal(t, k2, res.Key)
	s2 := attentionSessions(t, admin)[k2]
	f2, err := store.DecodeFields[store.SessionFields](s2)
	require.NoError(t, err)
	assert.False(t, f2.Prestart, "started on click")

	// A human message lifts read-only.
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: s1.ID, Text: "go ahead"}, &rpc.SessionSendResult{}))
	f, err = store.DecodeFields[store.SessionFields](getTicket(t, admin, s1.ID).Ticket)
	require.NoError(t, err)
	assert.False(t, f.ReadOnly)
	var made store.Ticket
	require.NoError(t, sock.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "follow-up"}, &made))

	// Resolving the item closes its chat once nothing keeps it (JOB-52).
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: q1.ID}, nil))
	pass(t, root.Dir, 1)
	assert.Equal(t, store.StatusClosed, getTicket(t, admin, s1.ID).Ticket.Status)
	assert.Equal(t, store.StatusOpen, getTicket(t, admin, s2.ID).Ticket.Status)
}

func metaDesk(t *testing.T, tk store.Ticket) string {
	t.Helper()
	var m struct {
		Desk struct {
			AttentionChat string `json:"attention_chat"`
		} `json:"desk"`
	}
	if len(tk.Meta) > 0 {
		require.NoError(t, json.Unmarshal(tk.Meta, &m))
	}
	return m.Desk.AttentionChat
}
