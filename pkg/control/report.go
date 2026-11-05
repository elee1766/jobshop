package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Structured worker reports (#186). A session reports the state of the
// ticket it works on; the controller records it in meta.report, so
// stall-check and the planner read states instead of comment text. A
// report may carry a question for a human, filed as a needs-human task
// with meta.question; the gateway opens a thread for it and a reply is
// delivered to the asking session through question.answer.

const (
	reportMetaPack   = "report"
	questionMetaPack = "question"
	// resolutionAnswered closes a question ticket that got its answer.
	resolutionAnswered = "answered"
)

// reportMeta is meta.report on a reported ticket.
type reportMeta struct {
	State    string    `json:"state"`
	Note     string    `json:"note,omitempty"`
	Question string    `json:"question,omitempty"` // open question ticket id
	Until    string    `json:"until,omitempty"`
	By       string    `json:"by"`
	At       time.Time `json:"at"`
}

// questionMeta is meta.question on a question ticket.
type questionMeta struct {
	Ticket  string `json:"ticket"`            // the ticket the question is about
	Session string `json:"session,omitempty"` // who asked; the answer goes there
}

func (c *Controller) report(ctx context.Context, sc scope, p rpc.ReportParams) (any, error) {
	if !slices.Contains(rpc.ReportStates, p.State) {
		return nil, fmt.Errorf("%w: report state %q is not one of %s", store.ErrInvalid, p.State, strings.Join(rpc.ReportStates, ", "))
	}
	if p.Until != nil && p.State != rpc.ReportWaiting {
		return nil, fmt.Errorf("%w: --until is only for waiting", store.ErrInvalid)
	}
	if p.Until != nil && !p.Until.After(c.now()) {
		return nil, fmt.Errorf("%w: until must be in the future", store.ErrInvalid)
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if !sc.admin {
		if _, err := c.heldClaim(ctx, sc, id); err != nil {
			return nil, err
		}
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	var res rpc.ReportResult
	rm := reportMeta{State: p.State, Note: strings.TrimSpace(p.Note), By: sc.actor, At: c.now().UTC()}
	if p.Until != nil {
		rm.Until = p.Until.UTC().Format(time.RFC3339)
	}
	if q := strings.TrimSpace(p.Question); q != "" {
		qt, err := c.fileQuestion(ctx, sc, t, q)
		if err != nil {
			return nil, err
		}
		res.Question = &qt
		rm.Question = qt.ID
	}
	if t, err = c.setReport(ctx, sc.actor, id, rm); err != nil {
		return nil, err
	}
	if _, err := c.store.AddComment(ctx, sc.actor, id, reportComment(rm, p.Question)); err != nil {
		return nil, err
	}
	// The blocked label is what the stall signal and the planner see.
	has := slices.Contains(t.Labels, blockedLabel)
	switch {
	case p.State == rpc.ReportBlocked && !has:
		err = c.store.AddLabel(ctx, sc.actor, id, blockedLabel)
	case p.State != rpc.ReportBlocked && has:
		err = c.store.RemoveLabel(ctx, sc.actor, id, blockedLabel)
	}
	if err != nil {
		return nil, err
	}
	if p.Until != nil && !sc.admin {
		// Waiting until a time hands the ticket back deferred, as release --until.
		cl, err := c.heldClaim(ctx, sc, id)
		if err != nil {
			return nil, err
		}
		if _, err := c.store.Release(ctx, cl, "waiting: "+rm.Note, *p.Until); err != nil {
			return nil, err
		}
	}
	if res.Ticket, err = c.store.Get(ctx, id); err != nil {
		return nil, err
	}
	return res, nil
}

func reportComment(rm reportMeta, question string) string {
	s := "report " + rm.State
	if rm.Note != "" {
		s += ": " + rm.Note
	}
	if rm.Until != "" {
		s += " (until " + rm.Until + ")"
	}
	if rm.Question != "" {
		s += "\nquestion " + rm.Question + ": " + strings.TrimSpace(question)
	}
	return s
}

// setReport replaces meta.report, keeping the other packs, conditional on
// the revision read.
func (c *Controller) setReport(ctx context.Context, actor, id string, rm reportMeta) (store.Ticket, error) {
	for attempt := 0; ; attempt++ {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return t, err
		}
		m := map[string]json.RawMessage{}
		if len(t.Meta) > 0 {
			_ = json.Unmarshal(t.Meta, &m)
		}
		m[reportMetaPack], _ = json.Marshal(rm)
		mb, err := json.Marshal(m)
		if err != nil {
			return t, err
		}
		t, err = c.store.Update(ctx, actor, id, store.Update{Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, store.ErrConflict) && attempt < 5 {
			continue
		}
		return t, err
	}
}

// reportOf is the ticket's last report, if any.
func reportOf(t store.Ticket) (reportMeta, bool) {
	var m map[string]json.RawMessage
	var rm reportMeta
	if json.Unmarshal(t.Meta, &m) != nil || json.Unmarshal(m[reportMetaPack], &rm) != nil || rm.State == "" {
		return rm, false
	}
	return rm, true
}

// questionTo addresses a question. No session has it as id, key or pool,
// so the message is never delivered to an agent.
const questionTo = "human"

// fileQuestion creates the question as a needs-human message to a human.
// A message is not claimable and has no pool, so no agent picks it up and
// it is not an orphan; a human answers through the gateway or
// `jobshop answer`.
func (c *Controller) fileQuestion(ctx context.Context, sc scope, t store.Ticket, q string) (store.Ticket, error) {
	meta := map[string]json.RawMessage{}
	if len(t.Meta) > 0 {
		_ = json.Unmarshal(t.Meta, &meta)
	}
	delete(meta, reportMetaPack)
	delete(meta, "gateway")
	delete(meta, runMetaPack) // a step's run link would make it look like a run escalation
	meta[questionMetaPack], _ = json.Marshal(questionMeta{Ticket: t.ID, Session: sc.session})
	mb, err := json.Marshal(meta)
	if err != nil {
		return store.Ticket{}, err
	}
	title, _, _ := strings.Cut(q, "\n")
	if len(title) > 120 {
		title = title[:120] + "..."
	}
	b := fmt.Sprintf("Question from %s on %s (%s):\n\n%s\n", sc.actor, t.ID, t.Title, q)
	b += fmt.Sprintf("\nAnswer with `jobshop answer <this ticket> \"<text>\"` or in its gateway thread; the answer is delivered to the asking session and commented on %s.\n", t.ID)
	fields, err := json.Marshal(store.MessageFields{To: questionTo, From: sc.actor})
	if err != nil {
		return store.Ticket{}, err
	}
	prio := 1
	return c.store.Create(ctx, sc.actor, store.NewTicket{
		Project: t.Project, Kind: store.KindMessage, Title: "Question: " + title, Body: b,
		Priority: &prio, Labels: []string{needsHumanLabel}, Meta: mb, Fields: fields,
	})
}

// questionOf is meta.question of a question ticket.
func questionOf(t store.Ticket) (questionMeta, bool) {
	var m map[string]json.RawMessage
	var q questionMeta
	if json.Unmarshal(t.Meta, &m) != nil || json.Unmarshal(m[questionMetaPack], &q) != nil || q.Ticket == "" {
		return q, false
	}
	return q, true
}

// answerQuestion delivers a human's answer: a message to the asking
// session when it is still open, a comment on the asked ticket (so a new
// session sees it), a waiting defer lifted, and the question closed.
func (c *Controller) answerQuestion(ctx context.Context, sc scope, p rpc.AnswerParams) (any, error) {
	if err := requireAdmin(sc, "answer"); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(p.Text)
	if text == "" {
		return nil, fmt.Errorf("%w: answer text is required", store.ErrInvalid)
	}
	qt, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	qm, ok := questionOf(qt)
	if !ok {
		return nil, fmt.Errorf("%w: %s is not a question", store.ErrInvalid, qt.ID)
	}
	if qt.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: question %s is already closed (%s)", store.ErrInvalid, qt.ID, qt.Resolution)
	}
	from := sc.actor
	if p.From != "" {
		from = p.From
	}
	res := rpc.AnswerResult{}
	c.recordAnswer(ctx, qt, qm.Ticket, from, text)
	if _, err := c.store.AddComment(ctx, sc.actor, qt.ID, "answer from "+from+": "+text); err != nil {
		return nil, err
	}
	if _, err := c.store.AddComment(ctx, sc.actor, qm.Ticket, fmt.Sprintf("answer to question %s from %s: %s", qt.ID, from, text)); err != nil {
		return nil, err
	}
	if qm.Session != "" {
		if s, err := c.store.Get(ctx, qm.Session); err == nil && s.Status != store.StatusClosed {
			fields, err := json.Marshal(store.MessageFields{To: s.ID, From: from})
			if err != nil {
				return nil, err
			}
			m, err := c.store.Create(ctx, sc.actor, store.NewTicket{
				Project: s.Project, Kind: store.KindMessage,
				Title: fmt.Sprintf("Answer to your question %s on %s", qt.ID, qm.Ticket), Body: text, Fields: fields,
			})
			if err != nil {
				return nil, err
			}
			c.wakeRunner(s.ID)
			res.Message = &m
		}
	}
	if t, err := c.store.Get(ctx, qm.Ticket); err == nil && t.Status != store.StatusClosed {
		if rm, ok := reportOf(t); ok && rm.Question == qt.ID {
			rm.State, rm.Note, rm.Question, rm.Until, rm.By, rm.At = rpc.ReportProgress, "question "+qt.ID+" answered", "", "", sc.actor, c.now().UTC()
			if _, err := c.setReport(ctx, sc.actor, t.ID, rm); err != nil {
				return nil, err
			}
			if slices.Contains(t.Labels, blockedLabel) {
				if err := c.store.RemoveLabel(ctx, sc.actor, t.ID, blockedLabel); err != nil {
					return nil, err
				}
			}
		}
		if t.Status == store.StatusOpen && t.DeferUntil.After(c.now()) {
			now := c.now()
			if _, err := c.store.Update(ctx, sc.actor, t.ID, store.Update{DeferUntil: &now}); err != nil {
				return nil, err
			}
		}
	}
	if res.Question, err = c.store.CloseTicket(ctx, sc.actor, qt.ID, nil, resolutionAnswered); err != nil {
		return nil, err
	}
	c.unblockLinear(ctx, qt)
	return res, nil
}
