package linearsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
)

// Linear issue communication (#293). For a run with meta.linear.issue,
// worker questions and run escalations are asked on the Linear issue
// (the same commentCreate as `jobshop linear ask`), mentioning
// meta.linear.human. A human reply under the ask comment answers the
// question. Questions also set the Blocked label until answered. Routine
// progress stays on GitHub. All of this is opt-in with
// linear.ask_on_issue: true (JOB-114); by default questions and
// escalations stay on Desk.

const (
	// AskPack is meta.linear_ask on a question or escalation ticket: the
	// posted comment, so replies can be matched to it.
	AskPack = "linear_ask"
	// BlockedLabel marks a Linear issue waiting on a scope question;
	// linear_skip_labels should list it to keep it out of intake.
	BlockedLabel = "Blocked"
)

// askTimeout bounds one ask post: the token, the comment, the label and
// the store writes.
const askTimeout = 2 * time.Minute

// Ask is meta.linear_ask.
type Ask struct {
	Comment string `json:"comment"`
	URL     string `json:"url,omitempty"`
	Issue   string `json:"issue"`
}

// AskOf is the posted ask recorded on t.
func AskOf(t domain.Ticket) (Ask, bool) {
	var m map[string]json.RawMessage
	var a Ask
	if json.Unmarshal(t.Meta, &m) != nil || json.Unmarshal(m[AskPack], &a) != nil || a.Comment == "" {
		return a, false
	}
	return a, true
}

// issueOf is meta.linear.issue and meta.linear.human of t.
func issueOf(t domain.Ticket) (issue, human string) {
	var m struct {
		Linear struct {
			Issue string `json:"issue"`
			Human string `json:"human"`
		} `json:"linear"`
	}
	if json.Unmarshal(t.Meta, &m) != nil {
		return "", ""
	}
	return m.Linear.Issue, m.Linear.Human
}

// SweepAsks posts every open question and escalation for a Linear run
// that is not posted yet. It runs each tick, so a failed post (Linear
// down, token expiring) is retried; the write key makes a retry after an
// unknown outcome find the first comment instead of posting again.
func (s *Service) SweepAsks(ctx context.Context, cfg *config.Config) {
	if !cfg.Linear.AsksOnIssue() {
		return
	}
	for _, kind := range []domain.Kind{domain.KindMessage, domain.KindTask} {
		ts, err := s.host.OpenNeedsHuman(ctx, kind)
		if err != nil {
			s.log().WarnContext(ctx, "linear ask: list", "err", err)
			return
		}
		for _, t := range ts {
			if issue, _ := issueOf(t); issue == "" {
				continue
			}
			if _, ok := AskOf(t); ok {
				continue
			}
			// Outside the tick (JOB-103), one post per ticket at a time.
			s.host.OffGo("linear-ask\x00"+t.ID, "linear_ask", askTimeout, func(ctx context.Context) {
				// Read again: a post that ended after this pass listed
				// the ticket has set meta.linear_ask.
				cur, err := s.host.Get(ctx, t.ID)
				if err != nil || cur.Status != domain.StatusOpen {
					return
				}
				if err := s.post(ctx, cfg, cur); err != nil && s.host.Context().Err() == nil {
					s.log().WarnContext(ctx, "linear ask", "ticket", t.ID, "err", err)
				}
			})
		}
	}
}

// post posts t when it is a question or escalation on a Linear issue and
// not posted yet.
func (s *Service) post(ctx context.Context, cfg *config.Config, t domain.Ticket) error {
	key, human := issueOf(t)
	if key == "" {
		return nil
	}
	if _, ok := AskOf(t); ok {
		return nil
	}
	question := s.host.IsQuestion(t)
	if !question && !s.host.IsEscalation(t) {
		return nil
	}
	cl, err := s.host.ClientIn(cfg, key)
	if err != nil {
		return err
	}
	var mention []string
	if human != "" {
		mention = []string{human}
	}
	wkey := "linear-ask/" + t.ID
	m, err := cl.PlanAsk(ctx, key, linear.AskParams{Body: askBody(t, question), Mention: mention}, linear.KeyID(wkey))
	if err != nil {
		return err
	}
	out, err := cl.Do(ctx, s.host.Keys(t.Project), wkey, false, m)
	if err != nil {
		return err
	}
	res, _ := out.(linear.Result)
	if res.ID == "" {
		return errors.New("linear ask: no comment id")
	}
	if question {
		// Scope question: Blocked until answered.
		if err := cl.AddLabel(ctx, key, BlockedLabel); err != nil {
			s.log().WarnContext(ctx, "linear ask: blocked label", "issue", key, "err", err)
		}
	}
	if err := s.host.SetMeta(ctx, t.ID, AskPack, Ask{Comment: res.ID, URL: res.URL, Issue: key}); err != nil {
		return err
	}
	return s.host.Comment(ctx, t.ID, fmt.Sprintf("asked on Linear %s: %s", key, res.URL))
}

func askBody(t domain.Ticket, question bool) string {
	if question {
		return t.Body + "\nReply to this comment to answer; the reply is delivered to the session that asked.\n"
	}
	// The escalation body is long; the issue gets the summary and the
	// ticket id to look at.
	head, _, _ := strings.Cut(t.Body, "\n## ")
	return head + fmt.Sprintf("\nDetails: jobshop ticket %s. Replies to this comment reach that ticket.\n", t.ID)
}

// Reply is a comment on a Linear issue thread: Parent is the comment it
// replies to.
type Reply struct {
	Parent, Body, Author string
}

// openQuestions is the open questions with a posted ask.
func (s *Service) openQuestions(ctx context.Context) ([]domain.Ticket, error) {
	qs, err := s.host.OpenNeedsHuman(ctx, domain.KindMessage)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(qs, func(q domain.Ticket) bool {
		_, ok := AskOf(q)
		return !ok || !s.host.IsQuestion(q)
	}), nil
}

// Answers answers the open questions whose ask comment on issue the
// replies reply to, and reports which replies it took; the rest are
// normal comments. A failed answer is not taken, so the reply is not lost.
func (s *Service) Answers(ctx context.Context, issue string, replies []Reply) []bool {
	taken := make([]bool, len(replies))
	if issue == "" || !s.host.Config().Linear.AsksOnIssue() {
		return taken
	}
	qs, err := s.openQuestions(ctx)
	if err != nil {
		return taken
	}
	byComment := map[string]domain.Ticket{}
	for _, q := range qs {
		if a, _ := AskOf(q); a.Issue == issue {
			byComment[a.Comment] = q
		}
	}
	for i, r := range replies {
		q, ok := byComment[r.Parent]
		if !ok {
			continue
		}
		if err := s.host.AnswerQuestion(ctx, q.ID, r.Body, r.Author); err != nil {
			s.log().WarnContext(ctx, "linear answer", "question", q.ID, "err", err)
			continue
		}
		taken[i] = true
		delete(byComment, r.Parent) // a second reply is a normal comment
	}
	return taken
}

// Unblock removes the Blocked label once no other open question on the
// issue is waiting. It runs with ask_on_issue off too: a label posted
// before the switch was turned off is still cleared.
func (s *Service) Unblock(ctx context.Context, qt domain.Ticket) {
	a, ok := AskOf(qt)
	if !ok {
		return
	}
	qs, err := s.openQuestions(ctx)
	if err != nil {
		return
	}
	if slices.ContainsFunc(qs, func(o domain.Ticket) bool {
		oa, _ := AskOf(o)
		return o.ID != qt.ID && oa.Issue == a.Issue
	}) {
		return
	}
	cl, err := s.host.ClientIn(s.host.Config(), a.Issue)
	if err != nil {
		s.log().WarnContext(ctx, "linear unblock", "issue", a.Issue, "err", err)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := cl.RemoveLabel(ctx, a.Issue, BlockedLabel); err != nil {
		s.log().WarnContext(ctx, "linear unblock", "issue", a.Issue, "err", err)
	}
}
