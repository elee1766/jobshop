package control

import (
	"context"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/control/chatlife"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Question chats (JOB-52) work like escalation chats: the operator talks
// to the chat agent, which answers the asking worker with question.answer
// (grant question-answer), then reports the worker's reaction in the
// chat. A re-ask from the same session on the same ticket threads into
// the open chat. The chat stays open while a follow-up or re-ask is
// unsettled or a Desk client watches it.

const (
	grantQuestionAnswer = "question-answer"
	questionChatPrefix  = chatlife.QuestionChatPrefix
)

// noteOperator keeps the human who wrote to an attention chat, so the
// answer the chat relays records them as from.
func (c *Controller) noteOperator(ctx context.Context, sess domain.Ticket, actor string) {
	f, err := domain.DecodeFields[domain.SessionFields](sess)
	if err != nil || !strings.HasPrefix(f.Key, attentionKeyPrefix) || f.Operator == actor || actor == "" || actor == "admin" || actor == "controller" {
		return
	}
	if err := c.updateSessionFields(ctx, sess.ID, func(f *domain.SessionFields) { f.Operator = actor }); err != nil {
		c.log.WarnContext(ctx, "record chat operator", "session", sess.ID, "err", err)
	}
}

// chatAnswerFrom is who a session's answer is from: the operator who
// wrote to the chat, else the session.
func (c *Controller) chatAnswerFrom(ctx context.Context, sc scope) string {
	if s, err := c.store.Get(ctx, sc.session); err == nil {
		if f, err := domain.DecodeFields[domain.SessionFields](s); err == nil && f.Operator != "" {
			return f.Operator
		}
	}
	return sc.actor
}

// followAnswer marks the answering chat as waiting on the asker's
// reaction to message (nil when the asker had ended).
func (c *Controller) followAnswer(ctx context.Context, chat string, qt domain.Ticket, qm questionMeta, msg *domain.Ticket) {
	if chat == "" {
		return
	}
	fl := &domain.QuestionFollow{Question: qt.ID, Ticket: qm.Ticket, Asker: qm.Session, Since: c.now().UTC()}
	if msg != nil {
		fl.Message = msg.ID
	}
	if err := c.updateSessionFields(ctx, chat, func(f *domain.SessionFields) { f.Follow = fl }); err != nil {
		c.log.WarnContext(ctx, "follow answer", "session", chat, "err", err)
	}
}

// answerDelivered is called after the turn that delivered msgs ended: a
// chat following one of them reports the reaction now.
func (c *Controller) answerDelivered(ctx context.Context, msgs []domain.Ticket) {
	for _, m := range msgs {
		if !strings.HasPrefix(m.Title, answerTitlePrefix) {
			continue
		}
		chats, err := c.followingChats(ctx)
		if err != nil {
			return
		}
		for _, s := range chats {
			if f, _ := domain.DecodeFields[domain.SessionFields](s); f.Follow != nil && f.Follow.Message == m.ID {
				c.reportReaction(ctx, s, *f.Follow)
			}
		}
	}
}

// followingChats are the open sessions waiting on an asker's reaction.
func (c *Controller) followingChats(ctx context.Context) ([]domain.Ticket, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
	if err != nil {
		return nil, err
	}
	var out []domain.Ticket
	for _, s := range open {
		if f, err := domain.DecodeFields[domain.SessionFields](s); err == nil && f.Follow != nil {
			out = append(out, s)
		}
	}
	return out, nil
}

// sweepQuestionFollows reports reactions whose delivery turn will never
// end here: the asker ended, or the answer had no session to go to.
func (c *Controller) sweepQuestionFollows(ctx context.Context) {
	chats, err := c.followingChats(ctx)
	if err != nil {
		return
	}
	for _, s := range chats {
		f, _ := domain.DecodeFields[domain.SessionFields](s)
		fl := *f.Follow
		gone := fl.Message == "" || fl.Asker == ""
		if !gone {
			if a, err := c.store.Get(ctx, fl.Asker); err != nil || a.Status == domain.StatusClosed {
				gone = true
			}
		}
		if gone {
			c.reportReaction(ctx, s, fl)
		}
	}
}

// reportReaction tells the chat what the asker did after the answer and
// ends the follow-up.
func (c *Controller) reportReaction(ctx context.Context, chat domain.Ticket, fl domain.QuestionFollow) {
	if err := c.updateSessionFields(ctx, chat.ID, func(f *domain.SessionFields) { f.Follow = nil }); err != nil {
		c.log.WarnContext(ctx, "end follow-up", "session", chat.ID, "err", err)
		return
	}
	text := c.reaction(ctx, fl)
	if err := c.sendMessage(ctx, chat, "", text); err != nil {
		c.log.WarnContext(ctx, "report reaction", "session", chat.ID, "err", err)
	}
	c.log.InfoContext(ctx, "question reaction", "question", fl.Question, "session", chat.ID)
}

// reaction describes the work ticket after the asker's turn: re-asked,
// parked or resolved.
func (c *Controller) reaction(ctx context.Context, fl domain.QuestionFollow) string {
	head := fmt.Sprintf("Reaction of the asking session %s on %s to the answer to %s: ", fl.Asker, fl.Ticket, fl.Question)
	if re, ok := c.reask(ctx, fl.Asker, fl.Ticket, fl.Question); ok {
		return head + "re-asked. New question " + re.ID + ": " + strings.TrimPrefix(re.Title, "Question: ") +
			"\nShow it to the operator and answer it here with `jobshop answer " + re.ID + " \"<text>\"` when they reply."
	}
	t, err := c.store.Get(ctx, fl.Ticket)
	if err != nil {
		return head + "unknown (" + err.Error() + ")."
	}
	switch {
	case t.Status == domain.StatusClosed:
		return head + "resolved: " + t.ID + " is closed (" + t.Resolution + "). Tell the operator, then this chat can end."
	case waitOf(t) != nil:
		return head + "parked: " + t.ID + " waits on " + waitOf(t).Reason + ". Tell the operator."
	case t.DeferUntil.After(c.now()):
		return head + "parked: " + t.ID + " is deferred until " + t.DeferUntil.UTC().Format("2006-01-02 15:04 MST") + ". Tell the operator."
	}
	if rm, ok := reportOf(t); ok && rm.State != rpc.ReportProgress {
		return head + "resolved: " + t.ID + " reported " + rm.State + noteSuffix(rm.Note) + ". Tell the operator."
	}
	return head + "resolved: " + t.ID + " is " + string(t.Status) + ". Tell the operator."
}

func noteSuffix(n string) string {
	if n == "" {
		return ""
	}
	return " (" + n + ")"
}

// reask is the newest open question from asker on ticket other than
// skip.
func (c *Controller) reask(ctx context.Context, asker, ticket, skip string) (domain.Ticket, bool) {
	if asker == "" {
		return domain.Ticket{}, false
	}
	qs, err := c.store.List(ctx, store.Filter{Kind: domain.KindMessage, Status: domain.StatusOpen, MetaPack: questionMetaPack, MetaKey: "session", MetaValue: asker, Order: store.OrderNewest})
	if err != nil {
		return domain.Ticket{}, false
	}
	for _, q := range qs {
		if qm, ok := questionOf(q); ok && q.ID != skip && qm.Ticket == ticket {
			return q, true
		}
	}
	return domain.Ticket{}, false
}

// threadQuestion puts a re-ask into the open chat of an earlier question
// from the same session on the same ticket: the new question records
// that chat's key, so it gets no chat of its own, and the chat is told.
func (c *Controller) threadQuestion(ctx context.Context, q domain.Ticket, qm questionMeta) {
	if qm.Session == "" {
		return
	}
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
	if err != nil {
		return
	}
	for _, s := range open {
		f, err := domain.DecodeFields[domain.SessionFields](s)
		if err != nil || !strings.HasPrefix(f.Key, questionChatPrefix) {
			continue
		}
		first, err := c.store.Get(ctx, strings.TrimPrefix(f.Key, questionChatPrefix))
		if err != nil {
			continue
		}
		if fm, ok := questionOf(first); !ok || fm.Session != qm.Session || fm.Ticket != qm.Ticket {
			continue
		}
		if err := c.recordAttentionChat(ctx, q.ID, f.Key); err != nil {
			c.log.WarnContext(ctx, "thread question", "question", q.ID, "err", err)
			return
		}
		if f.Follow != nil {
			// The follow-up reports it as a re-ask when the turn ends.
			return
		}
		text := fmt.Sprintf("The asking session %s asked again on %s. New question %s:\n%s\nShow it to the operator and answer it here with `jobshop answer %s \"<text>\"` when they reply.", qm.Session, qm.Ticket, q.ID, q.Body, q.ID)
		if err := c.sendMessage(ctx, s, "", text); err != nil {
			c.log.WarnContext(ctx, "thread question", "session", s.ID, "err", err)
		}
		return
	}
}

// questionChatOpen: a question chat stays open while its follow-up or a
// threaded question is unsettled, or while a Desk client watches it.
func (c *Controller) questionChatOpen(ctx context.Context, s domain.Ticket, f domain.SessionFields) bool {
	if f.Follow != nil || c.watched(s.ID) {
		return true
	}
	qs, err := c.store.List(ctx, store.Filter{Kind: domain.KindMessage, Status: domain.StatusOpen, MetaPack: deskMetaPack, MetaKey: "attention_chat", MetaValue: f.Key})
	return err == nil && len(qs) > 0
}

// watch counts a Desk client streaming session id; the returned func
// ends it.
func (c *Controller) watch(id string) func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.watchers == nil {
		c.watchers = map[string]int{}
	}
	c.watchers[id]++
	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.watchers[id]--; c.watchers[id] <= 0 {
			delete(c.watchers, id)
		}
	}
}

func (c *Controller) watched(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.watchers[id] > 0
}
