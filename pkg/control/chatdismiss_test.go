package control_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

func TestManagedChatDismissal(t *testing.T) {
	t.Parallel()
	cfg := fmt.Sprintf(`default_project: js
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
    max_sessions: 4
    dir: $ROOT
escalation:
  chat_pool: operator
  hold_prompt: "Discuss {{.ticket}}: {{.body}}"
attention:
  prestart_max: 1
`, os.Args[0])
	for _, kind := range []string{"escalation", "attention", "converted"} {
		t.Run(kind, func(t *testing.T) {
			root := controltest.NewRoot(t, cfg)
			c := controltest.Start(t, root, control.Options{})
			label := "needs-human"
			if kind == "attention" {
				label = "question"
			}
			var source domain.Ticket
			c.Call(rpc.MethodTicketCreate, rpc.CreateParams{Title: "Operator decision", Labels: []string{label}}, &source)
			key := "escalation:" + source.ID
			if kind == "attention" {
				key = "attention:question:" + source.ID
			}
			var chat domain.Ticket
			c.PassUntil(func() bool {
				var sessions []domain.Ticket
				c.Call(rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession}, &sessions)
				for _, s := range sessions {
					f, err := domain.DecodeFields[domain.SessionFields](s)
					require.NoError(t, err)
					if f.Key == key {
						chat = s
						return true
					}
				}
				return false
			}, "managed chat was not started")
			if kind == "converted" {
				var converted rpc.SessionInfo
				c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "convert", ID: chat.ID, Name: "ongoing discussion"}, &converted)
				assert.Equal(t, chat.ID, converted.ID)
				var sourceView rpc.ShowResult
				c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: source.ID}, &sourceView)
				assert.Equal(t, domain.StatusOpen, sourceView.Ticket.Status)
				fields, err := domain.DecodeFields[domain.SessionFields](*converted.Ticket)
				require.NoError(t, err)
				assert.True(t, fields.Named)
				assert.Equal(t, key, fields.Key)
				require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodChatManage, rpc.ChatManageParams{Action: "resolve", ID: chat.ID}, nil), domain.ErrInvalid)
				c.Call(rpc.MethodSessionSend, rpc.SessionSendParams{ID: chat.ID, Text: "Continue independently"}, nil)
				c.PassUntil(func() bool {
					var journal rpc.SessionJournalResult
					c.Call(rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: chat.ID}, &journal)
					var decoder rpc.EventDecoder
					for _, frame := range journal.Frames {
						if event, ok := decoder.Decode(frame.Seq, frame.T, frame.From, frame.Msg); ok && event.Type == rpc.EventTurnStart && event.Text == "Continue independently" {
							return true
						}
					}
					return false
				}, "converted conversation did not accept a message")
				c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: source.ID}, nil)
				c.Pass(2)
				var shown rpc.ShowResult
				c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: chat.ID}, &shown)
				assert.Equal(t, domain.StatusOpen, shown.Ticket.Status)
				c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "archive", ID: chat.ID}, nil)
				return
			}
			c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "archive", ID: chat.ID}, nil)
			c.Pass(2)
			var shown rpc.ShowResult
			c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: source.ID}, &shown)
			assert.Equal(t, domain.StatusOpen, shown.Ticket.Status, "closing the chat does not resolve its source")
			c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: chat.ID}, &shown)
			assert.Equal(t, domain.StatusClosed, shown.Ticket.Status)
			fields, err := domain.DecodeFields[domain.SessionFields](shown.Ticket)
			require.NoError(t, err)
			assert.False(t, fields.Named, "managed chats keep their original type")
			require.NotNil(t, fields.Archived)
			var sessions []domain.Ticket
			c.Call(rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession, Status: domain.StatusOpen}, &sessions)
			for _, s := range sessions {
				f, err := domain.DecodeFields[domain.SessionFields](s)
				require.NoError(t, err)
				assert.NotEqual(t, key, f.Key, "a supervisor pass does not undo explicit dismissal")
			}
			c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "restore", ID: chat.ID}, nil)
			if kind == "escalation" {
				c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "resolve", ID: chat.ID}, nil)
				c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "resolve", ID: chat.ID}, nil)
				c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: source.ID}, &shown)
				assert.Equal(t, domain.StatusClosed, shown.Ticket.Status)
				c.Call(rpc.MethodTicketShow, rpc.IDParams{ID: chat.ID}, &shown)
				assert.Equal(t, domain.StatusClosed, shown.Ticket.Status)
			} else {
				require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodChatManage, rpc.ChatManageParams{Action: "resolve", ID: chat.ID}, nil), domain.ErrInvalid)
				c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "archive", ID: chat.ID}, nil)
				c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: source.ID}, nil)
			}
		})
	}
}
