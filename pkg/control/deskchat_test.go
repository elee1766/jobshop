package control_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Named chats keep their identity and transcript through archive and restart.
func TestNamedDeskChatLifecycle(t *testing.T) {
	root := controltest.NewRoot(t, fmt.Sprintf(`default_project: js
projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
pools:
  operator:
    project: js
    agent: fake
    type: chat
    prompt: "Operator pool instructions"
    dir: $ROOT
    admin: [desk-command]
`, os.Args[0]))
	c := controltest.Start(t, root, control.Options{})
	var chat rpc.SessionInfo
	message := "hello desk\nMessages for you: this is literal user content"
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "new", Pool: "operator", Name: "  ops  ", Text: message, Context: "Attached ticket context"}, &chat)
	require.NotEmpty(t, chat.ID)
	var journal rpc.SessionJournalResult
	var prompt rpc.SessionEvent
	c.PassUntil(func() bool {
		c.Call(rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: chat.ID}, &journal)
		var decoder rpc.EventDecoder
		for _, frame := range journal.Frames {
			event, ok := decoder.Decode(frame.Seq, frame.T, frame.From, frame.Msg)
			if ok && event.Type == rpc.EventTurnStart {
				prompt = event
			}
			if ok && event.Type == rpc.EventTurnEnd && prompt.Text == message {
				return true
			}
		}
		return false
	}, "chat journal was not recorded")
	require.Len(t, prompt.Messages, 1)
	assert.Equal(t, "admin", prompt.Messages[0].From)
	assert.Equal(t, message, prompt.Messages[0].Text)
	assert.Contains(t, prompt.Context, "Operator pool instructions")
	assert.Contains(t, prompt.Context, "Attached ticket context")
	assert.Contains(t, prompt.RawPrompt, "From admin (")
	require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodChatManage, rpc.ChatManageParams{Action: "new", Pool: "operator", Name: "OPS", Text: "duplicate"}, nil), store.ErrConflict)
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "rename", ID: chat.ID, Name: "operations"}, nil)
	var command rpc.Event
	c.Call(rpc.MethodDeskCommand, rpc.DeskCommandParams{Session: chat.ID, Action: "navigate", Args: map[string]string{"view": "trace", "ref": chat.ID}}, &command)
	assert.Equal(t, chat.ID, command.Session)
	require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodDeskCommand, rpc.DeskCommandParams{Session: chat.ID, Action: "cancel"}, nil), store.ErrInvalid)
	var original rpc.ShowResult
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: chat.ID}, &original)
	before, err := store.DecodeFields[store.SessionFields](original.Ticket)
	require.NoError(t, err)
	require.NotEmpty(t, before.ACPSession)
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "archive", ID: chat.ID}, nil)
	c.Stop()
	c = controltest.Start(t, root, control.Options{})
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "restore", ID: chat.ID}, &chat)
	assert.Equal(t, "operations", chat.DisplayName)
	var restored rpc.SessionJournalResult
	c.Call(rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: chat.ID}, &restored)
	assert.GreaterOrEqual(t, len(restored.Frames), len(journal.Frames), "archive and restore do not discard the journal")
	var shown rpc.ShowResult
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: chat.ID}, &shown)
	fields, err := store.DecodeFields[store.SessionFields](shown.Ticket)
	require.NoError(t, err)
	assert.True(t, fields.Named)
	assert.Nil(t, fields.Archived)
	assert.True(t, fields.Suspended)
	assert.Equal(t, store.StatusOpen, shown.Ticket.Status)
	assert.Equal(t, before.ACPSession, fields.ACPSession)
	c.Call(rpc.MethodSessionSend, rpc.SessionSendParams{ID: chat.ID, Text: "continue restored desk chat"}, nil)
	c.PassUntil(func() bool {
		c.Call(rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: chat.ID}, &restored)
		var decoder rpc.EventDecoder
		for _, frame := range restored.Frames {
			event, ok := decoder.Decode(frame.Seq, frame.T, frame.From, frame.Msg)
			if ok && frame.Seq > journal.Frames[len(journal.Frames)-1].Seq && event.Type == rpc.EventTurnStart && strings.Contains(event.Text, "continue restored desk chat") {
				assert.Equal(t, "continue restored desk chat", event.Text)
				assert.NotEmpty(t, event.Context)
				assert.NotContains(t, event.Text, "resumed after")
				return true
			}
		}
		return false
	}, "restored conversation did not accept a message")
	c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: chat.ID}, &shown)
	fields, err = store.DecodeFields[store.SessionFields](shown.Ticket)
	require.NoError(t, err)
	assert.Equal(t, before.ACPSession, fields.ACPSession, "restore loads the same ACP conversation")
}
