package control_test

import (
	"fmt"
	"os"
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
    dir: $ROOT
    admin: [desk-command]
`, os.Args[0]))
	c := controltest.Start(t, root, control.Options{})
	var chat rpc.SessionInfo
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "new", Pool: "operator", Name: "  ops  ", Text: "hello desk"}, &chat)
	require.NotEmpty(t, chat.ID)
	var journal rpc.SessionJournalResult
	c.PassUntil(func() bool {
		c.Call(rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: chat.ID}, &journal)
		for _, frame := range journal.Frames {
			if string(frame.Msg) != "" && frame.From == "agent" {
				return true
			}
		}
		return false
	}, "chat journal was not recorded")
	require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodChatManage, rpc.ChatManageParams{Action: "new", Pool: "operator", Name: "OPS", Text: "duplicate"}, nil), store.ErrConflict)
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "rename", ID: chat.ID, Name: "operations"}, nil)
	var command rpc.Event
	c.Call(rpc.MethodDeskCommand, rpc.DeskCommandParams{Session: chat.ID, Action: "navigate", Args: map[string]string{"view": "trace", "ref": chat.ID}}, &command)
	assert.Equal(t, chat.ID, command.Session)
	require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodDeskCommand, rpc.DeskCommandParams{Session: chat.ID, Action: "cancel"}, nil), store.ErrInvalid)
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
}
