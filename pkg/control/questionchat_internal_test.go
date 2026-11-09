package control

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// grantPool gives pool the admin methods grants.
func grantPool(t *testing.T, c *Controller, pool string, grants ...string) {
	t.Helper()
	cfg := *c.config()
	pools := map[string]config.Pool{}
	for k, v := range cfg.Pools {
		pools[k] = v
	}
	p := pools[pool]
	p.Admin = grants
	pools[pool] = p
	cfg.Pools = pools
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
}

// TestQuestionAnswerGrant (JOB-52): a chat pool with question-answer
// answers, recorded as from the operator who wrote to the chat, and then
// follows the asker; a pool without it cannot. A re-ask threads into
// the same chat and the reaction reports it.
func TestQuestionAnswerGrant(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	step := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev"})
	asker := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, _, err := c.store.Claim(ctx, step.ID, asker.ID, time.Hour)
	require.NoError(t, err)
	ask := scope{actor: asker.ID, session: asker.ID, project: "js", pool: "dev"}
	out, err := c.report(ctx, ask, rpc.ReportParams{State: rpc.ReportWaiting, Question: "which branch?"})
	require.NoError(t, err)
	q := *out.(rpc.ReportResult).Question

	key := attentionKey(rpc.AttentionQuestion, q.ID)
	chat := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "plan", Fields: json.RawMessage(`{"agent":"fake","key":"` + key + `"}`)})
	csc := scope{actor: chat.ID, session: chat.ID, project: "js", pool: "plan"}
	_, err = c.answerQuestion(ctx, csc, rpc.AnswerParams{ID: q.ID, Text: "use main"})
	require.ErrorIs(t, err, rpc.ErrForbidden, "no grant")

	grantPool(t, c, "plan", grantQuestionAnswer)
	_, err = c.answerQuestion(ctx, scope{actor: "x", session: chat.ID, project: "other", pool: "plan"}, rpc.AnswerParams{ID: q.ID, Text: "use main"})
	require.ErrorIs(t, err, rpc.ErrForbidden, "other project")

	_, err = c.sessionSend(ctx, scope{admin: true, actor: "desk:alice"}, rpc.SessionSendParams{ID: chat.ID, Text: "tell it main"})
	require.NoError(t, err)
	res, err := c.answerQuestion(ctx, csc, rpc.AnswerParams{ID: q.ID, Text: "use main", From: "@spoof"})
	require.NoError(t, err)
	ar := res.(rpc.AnswerResult)
	require.NotNil(t, ar.Message)
	mf, err := domain.DecodeFields[domain.MessageFields](*ar.Message)
	require.NoError(t, err)
	assert.Equal(t, "desk:alice", mf.From, "the operator, not the session or a spoofed from")

	got, err := c.store.Get(ctx, chat.ID)
	require.NoError(t, err)
	f, err := domain.DecodeFields[domain.SessionFields](got)
	require.NoError(t, err)
	require.NotNil(t, f.Follow)
	assert.Equal(t, ar.Message.ID, f.Follow.Message)
	assert.True(t, c.questionChatOpen(ctx, got, f), "following keeps the chat open after the question closed")

	// The asker re-asks during its reaction turn: the question threads
	// into this chat, and the reaction reports the re-ask.
	re, err := c.report(ctx, ask, rpc.ReportParams{State: rpc.ReportWaiting, Question: "main or release?"})
	require.NoError(t, err)
	rq := *re.(rpc.ReportResult).Question
	rqt, err := c.store.Get(ctx, rq.ID)
	require.NoError(t, err)
	assert.Equal(t, key, metaString(rqt.Meta, deskMetaPack, "attention_chat"), "re-ask reuses the chat")
	c.answerDelivered(ctx, []domain.Ticket{*ar.Message})
	assert.Contains(t, chatContext(t, c, chat.ID), "re-asked. New question "+rq.ID)
	got, _ = c.store.Get(ctx, chat.ID)
	f, _ = domain.DecodeFields[domain.SessionFields](got)
	assert.Nil(t, f.Follow)
	assert.True(t, c.questionChatOpen(ctx, got, f), "the open re-ask keeps the chat")

	// Answering the re-ask, then the asker parks: reported as parked, and
	// the chat may close once nobody watches.
	res, err = c.answerQuestion(ctx, csc, rpc.AnswerParams{ID: rq.ID, Text: "release branch"})
	require.NoError(t, err)
	_, err = c.waitAdd(ctx, admin, rpc.WaitAddParams{ID: step.ID, Reason: "CI on PR 7", Conds: []domain.WaitCond{{Type: domain.WaitRun, Run: step.ID, State: "done"}}})
	require.NoError(t, err)
	c.answerDelivered(ctx, []domain.Ticket{*res.(rpc.AnswerResult).Message})
	assert.Contains(t, chatContext(t, c, chat.ID), "parked: "+step.ID+" waits on CI on PR 7")
	got, _ = c.store.Get(ctx, chat.ID)
	f, _ = domain.DecodeFields[domain.SessionFields](got)
	assert.False(t, c.questionChatOpen(ctx, got, f), "settled")
	done := c.watch(chat.ID)
	assert.True(t, c.questionChatOpen(ctx, got, f), "a Desk client is attached")
	done()
}

// chatContext is the text of the messages waiting for session id.
func chatContext(t *testing.T, c *Controller, id string) string {
	t.Helper()
	var b strings.Builder
	for _, m := range runnerInbox(t, c, id) {
		f, err := domain.DecodeFields[domain.MessageFields](m)
		require.NoError(t, err)
		b.WriteString(m.Title + "\n" + m.Body + "\n" + f.Context + "\n")
	}
	return b.String()
}
