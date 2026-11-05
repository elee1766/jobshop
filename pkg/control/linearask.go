package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/writekey"
)

// Linear issue communication (#293). For a run with meta.linear.issue,
// worker questions and run escalations are asked on the Linear issue
// (the same commentCreate as `jobshop linear ask`), mentioning
// meta.linear.human. A human reply under the ask comment answers the
// question. Questions also set the Blocked label until answered. Routine
// progress stays on GitHub.

const (
	// linearAskPack is meta.linear_ask on a question or escalation
	// ticket: the posted comment, so replies can be matched to it.
	linearAskPack = "linear_ask"
	// linearBlockedLabel marks a Linear issue waiting on a scope
	// question; linear_skip_labels should list it to keep it out of
	// intake.
	linearBlockedLabel = "Blocked"
)

// linearAsk is meta.linear_ask.
type linearAsk struct {
	Comment string `json:"comment"`
	URL     string `json:"url,omitempty"`
	Issue   string `json:"issue"`
}

func linearAskOf(t store.Ticket) (linearAsk, bool) {
	var m map[string]json.RawMessage
	var a linearAsk
	if json.Unmarshal(t.Meta, &m) != nil || json.Unmarshal(m[linearAskPack], &a) != nil || a.Comment == "" {
		return a, false
	}
	return a, true
}

// sweepLinearAsks posts every open question and escalation for a Linear
// run that is not posted yet. It runs each tick, so a failed post (Linear
// down, token expiring) is retried; the write key makes a retry after an
// unknown outcome find the first comment instead of posting again.
func (c *Controller) sweepLinearAsks(ctx context.Context) {
	for _, f := range []store.Filter{
		{Kind: store.KindMessage, Status: store.StatusOpen, Label: needsHumanLabel, Limit: 200},
		{Kind: store.KindTask, Status: store.StatusOpen, Label: needsHumanLabel, Limit: 200},
	} {
		ts, err := c.store.List(ctx, f)
		if err != nil {
			c.log.Warn("linear ask: list", "err", err)
			return
		}
		for _, t := range ts {
			if err := c.postLinearAsk(ctx, t); err != nil {
				c.log.Warn("linear ask", "ticket", t.ID, "err", err)
			}
		}
	}
}

// postLinearAsk posts t when it is a question or escalation on a Linear
// issue and not posted yet.
func (c *Controller) postLinearAsk(ctx context.Context, t store.Ticket) error {
	key := metaString(t.Meta, "linear", "issue")
	if key == "" {
		return nil
	}
	if _, ok := linearAskOf(t); ok {
		return nil
	}
	_, question := questionOf(t)
	escalation := metaString(t.Meta, runMetaPack, "escalation") != ""
	if !question && !escalation {
		return nil
	}
	cl, err := c.linearClient(ctx, key)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	var mention []string
	if h := metaString(t.Meta, "linear", "human"); h != "" {
		mention = []string{h}
	}
	wkey := "linear-ask/" + t.ID
	m, err := cl.PlanAsk(ctx, key, linear.AskParams{Body: linearAskBody(t, question), Mention: mention}, linear.KeyID(wkey))
	if err != nil {
		return err
	}
	out, err := cl.Do(ctx, storeKeys{c: c, project: t.Project}, wkey, false, m)
	if err != nil {
		return err
	}
	res, _ := out.(linear.Result)
	if res.ID == "" {
		return errors.New("linear ask: no comment id")
	}
	if question {
		// Scope question: Blocked until answered.
		if err := cl.AddLabel(ctx, key, linearBlockedLabel); err != nil {
			c.log.Warn("linear ask: blocked label", "issue", key, "err", err)
		}
	}
	if err := c.setMetaPack(ctx, t.ID, linearAskPack, linearAsk{Comment: res.ID, URL: res.URL, Issue: key}); err != nil {
		return err
	}
	_, err = c.store.AddComment(ctx, "controller", t.ID, fmt.Sprintf("asked on Linear %s: %s", key, res.URL))
	return err
}

func linearAskBody(t store.Ticket, question bool) string {
	if question {
		return t.Body + "\nReply to this comment to answer; the reply is delivered to the session that asked.\n"
	}
	// The escalation body is long; the issue gets the summary and the
	// ticket id to look at.
	head, _, _ := strings.Cut(t.Body, "\n## ")
	return head + fmt.Sprintf("\nDetails: jobshop ticket %s. Replies to this comment reach that ticket.\n", t.ID)
}

// setMetaPack replaces meta[pack] on id, conditional on the revision read.
func (c *Controller) setMetaPack(ctx context.Context, id, pack string, v any) error {
	for attempt := 0; ; attempt++ {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		m := map[string]json.RawMessage{}
		if len(t.Meta) > 0 {
			_ = json.Unmarshal(t.Meta, &m)
		}
		if m[pack], err = json.Marshal(v); err != nil {
			return err
		}
		mb, err := json.Marshal(m)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, "controller", id, store.Update{Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, store.ErrConflict) && attempt < 5 {
			continue
		}
		return err
	}
}

// storeKeys is linear.KeyStore on the store, named like a session's keys
// of the ticket's project.
type storeKeys struct {
	c       *Controller
	project string
}

func (k storeKeys) Claim(ctx context.Context, key, op string) (string, bool, bool, error) {
	w, err := writekey.Claim(ctx, k.c.store, writeKeyName(scope{project: k.project}, key), op, "controller")
	if err != nil {
		return "", false, false, err
	}
	return w.Result, w.Done, !w.Done && !w.New, nil
}

func (k storeKeys) Finish(ctx context.Context, key, result string) (string, error) {
	w, err := writekey.Finish(ctx, k.c.store, writeKeyName(scope{project: k.project}, key), result)
	if err != nil {
		return "", err
	}
	return w.Result, nil
}

// linearAnswers takes replies to open questions' ask comments out of cs
// and answers those questions with them; the rest is returned for normal
// delivery. A failed answer leaves the reply in cs so it is not lost.
func (c *Controller) linearAnswers(ctx context.Context, th thread, cs []ghComment) []ghComment {
	if th.Linear == "" {
		return cs
	}
	qs, err := c.store.List(ctx, store.Filter{Kind: store.KindMessage, Status: store.StatusOpen, Label: needsHumanLabel, Limit: 200})
	if err != nil {
		return cs
	}
	byComment := map[string]store.Ticket{}
	for _, q := range qs {
		if a, ok := linearAskOf(q); ok && a.Issue == th.Linear {
			if _, isQ := questionOf(q); isQ {
				byComment[a.Comment] = q
			}
		}
	}
	if len(byComment) == 0 {
		return cs
	}
	var rest []ghComment
	for _, cm := range cs {
		q, ok := byComment[cm.Parent]
		if !ok {
			rest = append(rest, cm)
			continue
		}
		_, err := c.answerQuestion(ctx, scope{admin: true, actor: "controller"}, rpc.AnswerParams{ID: q.ID, Text: cm.Body, From: linearAuthorPrefix + cm.Author})
		if err != nil {
			c.log.Warn("linear answer", "question", q.ID, "err", err)
			rest = append(rest, cm)
			continue
		}
		delete(byComment, cm.Parent) // a second reply is a normal comment
	}
	return rest
}

// unblockLinear removes the Blocked label once no other open question on
// the issue is waiting.
func (c *Controller) unblockLinear(ctx context.Context, qt store.Ticket) {
	a, ok := linearAskOf(qt)
	if !ok {
		return
	}
	qs, err := c.store.List(ctx, store.Filter{Kind: store.KindMessage, Status: store.StatusOpen, Label: needsHumanLabel, Limit: 200})
	if err != nil {
		return
	}
	if slices.ContainsFunc(qs, func(o store.Ticket) bool {
		oa, ok := linearAskOf(o)
		_, isQ := questionOf(o)
		return ok && isQ && o.ID != qt.ID && oa.Issue == a.Issue
	}) {
		return
	}
	cl, err := c.linearClient(ctx, a.Issue)
	if err != nil {
		c.log.Warn("linear unblock", "issue", a.Issue, "err", err)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := cl.RemoveLabel(ctx, a.Issue, linearBlockedLabel); err != nil {
		c.log.Warn("linear unblock", "issue", a.Issue, "err", err)
	}
}
