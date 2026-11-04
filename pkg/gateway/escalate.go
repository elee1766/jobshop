package gateway

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

// Escalation threads (#125). The controller files an escalation ticket
// (label needs-human, meta.run.escalation) and does not post anywhere.
// A gateway with an escalation room polls for those tickets, opens one
// thread per ticket, records it in the ticket's meta.gateway, and sends
// the first prompt to its chat pool keyed by that thread. Replies in the
// thread then reach the same session through Handle.

// EscalationLabel marks escalation tickets.
const EscalationLabel = "needs-human"

// MetaPack is the ticket meta pack the gateway writes.
const MetaPack = "gateway"

// Escalation is meta.gateway on an escalation ticket.
type Escalation struct {
	Gateway string `json:"gateway"` // gateway name that opened the thread
	Room    string `json:"room"`
	Root    string `json:"root"`
	Key     string `json:"key"`
	Session string `json:"session,omitempty"` // set once the first prompt is sent
	Final   bool   `json:"final,omitempty"`   // the closed message was posted
}

// runMeta is meta.run on an escalation ticket, written by the controller.
type runMeta struct {
	ID         string `json:"id"`
	Workflow   string `json:"workflow"`
	Escalation string `json:"escalation"`
}

func ticketMeta(t store.Ticket) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	if len(t.Meta) > 0 {
		_ = json.Unmarshal(t.Meta, &m)
	}
	return m
}

func metaRun(t store.Ticket) (runMeta, bool) {
	var r runMeta
	raw, ok := ticketMeta(t)["run"]
	if !ok || json.Unmarshal(raw, &r) != nil || r.Escalation == "" {
		return r, false
	}
	return r, true
}

// githubMeta is meta.github, the work reference with its repo (#183).
type githubMeta struct {
	Repo  string `json:"repo"`
	Issue any    `json:"issue"`
	PR    any    `json:"pr"`
}

// githubRef tells the session which issue or PR to read, with its repo.
func githubRef(t store.Ticket) string {
	var g githubMeta
	if raw, ok := ticketMeta(t)["github"]; ok {
		_ = json.Unmarshal(raw, &g)
	}
	repo := ""
	if g.Repo != "" {
		repo = " --repo " + g.Repo
	}
	switch {
	case g.Issue != nil:
		return fmt.Sprintf("its GitHub issue (`gh issue view %v%s --comments`)", g.Issue, repo)
	case g.PR != nil:
		return fmt.Sprintf("its GitHub PR (`gh pr view %v%s --comments`)", g.PR, repo)
	}
	return "any GitHub issue it names (the run's vars)"
}

// poolMeta is meta.pool on a pool quarantine escalation (#146).
type poolMeta struct {
	Name       string `json:"name"`
	Quarantine string `json:"quarantine"`
}

func metaPool(t store.Ticket) (poolMeta, bool) {
	var p poolMeta
	raw, ok := ticketMeta(t)["pool"]
	if !ok || json.Unmarshal(raw, &p) != nil || p.Name == "" {
		return p, false
	}
	return p, true
}

// isEscalation reports whether the controller filed t: a run escalation,
// a pool quarantine or a worker question (#186).
func isEscalation(t store.Ticket) bool {
	if _, ok := metaRun(t); ok {
		return true
	}
	if _, ok := metaQuestion(t); ok {
		return true
	}
	_, ok := metaPool(t)
	return ok
}

// questionMeta is meta.question on a worker question, written by the
// controller on `jobshop report --question`.
type questionMeta struct {
	Ticket  string `json:"ticket"`
	Session string `json:"session"`
}

func metaQuestion(t store.Ticket) (questionMeta, bool) {
	var q questionMeta
	raw, ok := ticketMeta(t)["question"]
	if !ok || json.Unmarshal(raw, &q) != nil || q.Ticket == "" {
		return q, false
	}
	return q, true
}

func metaEscalation(t store.Ticket) (Escalation, bool) {
	var e Escalation
	raw, ok := ticketMeta(t)[MetaPack]
	if !ok || json.Unmarshal(raw, &e) != nil || e.Root == "" {
		return e, false
	}
	return e, true
}

// errSkip ends an updateEscalation without writing.
var errSkip = errors.New("skip")

// updateEscalation reads the ticket, lets fn change its meta.gateway, and
// writes it conditional on the revision read, retrying on conflicts (the
// session claiming or commenting moves the revision). This is the dedup:
// two gateways or two passes cannot both record a thread.
func (g *Gateway) updateEscalation(ctx context.Context, id string, fn func(t store.Ticket, e *Escalation, ok bool) error) (store.Ticket, error) {
	for attempt := 0; ; attempt++ {
		r, err := g.ctl.Show(ctx, id)
		if err != nil {
			return store.Ticket{}, err
		}
		t := r.Ticket
		e, ok := metaEscalation(t)
		if err := fn(t, &e, ok); err != nil {
			return t, err
		}
		m := ticketMeta(t)
		m[MetaPack], _ = json.Marshal(e)
		mb, err := json.Marshal(m)
		if err != nil {
			return t, err
		}
		t, err = g.ctl.Update(ctx, rpc.UpdateParams{ID: id, Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, store.ErrConflict) && attempt < 5 {
			continue
		}
		return t, err
	}
}

// runEscalations follows events.subscribe until ctx ends. Each connect
// sends a ready event, which triggers one resync pass over the open
// escalation tickets; after that only change events drive threads. A
// lagged stream or an error reconnects and resyncs.
func (g *Gateway) runEscalations(ctx context.Context) {
	for ctx.Err() == nil {
		res, err := g.ctl.Events(ctx, rpc.EventsSubscribeParams{Labels: []string{EscalationLabel}}, func(ev rpc.Event) {
			g.escalationEvent(ctx, ev)
		})
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			g.log.Warn("gateway: events", "err", err)
		} else if res.Lagged {
			g.log.Info("gateway: events lagged, resyncing")
		}
		select {
		case <-ctx.Done():
		case <-time.After(g.retry):
		}
	}
}

// escalationEvent handles one change to a needs-human ticket.
func (g *Gateway) escalationEvent(ctx context.Context, ev rpc.Event) {
	switch {
	case ev.Kind == rpc.EventKindReady:
		g.Escalations(ctx)
	case ev.Kind != rpc.EventKindTicket:
	case ev.Op == "close":
		g.mu.Lock()
		_, ok := g.escalating[ev.ID]
		g.mu.Unlock()
		if !ok {
			return
		}
		if err := g.resolved(ctx, ev.ID); err != nil {
			g.log.Warn("gateway: escalation closed", "ticket", ev.ID, "err", err)
			return
		}
		g.mu.Lock()
		delete(g.escalating, ev.ID)
		g.forgetQuestion(ev.ID)
		g.mu.Unlock()
	case ev.Op == "create" || ev.Op == "reopen" || (ev.Op == "label_add" && ev.Label == EscalationLabel):
		if ev.Status == store.StatusClosed {
			return
		}
		r, err := g.ctl.Show(ctx, ev.ID)
		if err != nil {
			g.log.Warn("gateway: escalation", "ticket", ev.ID, "err", err)
			return
		}
		t := r.Ticket
		if !isEscalation(t) || t.Status == store.StatusClosed || !slices.Contains(t.Labels, EscalationLabel) {
			return
		}
		if err := g.escalation(ctx, t); err != nil {
			g.log.Warn("gateway: escalation", "ticket", t.ID, "err", err)
		}
	}
}

// Escalations does one resync pass: opens threads for new escalation tickets,
// resends a missing first prompt, and posts a final message in threads
// whose ticket closed since it was last seen open.
func (g *Gateway) Escalations(ctx context.Context) {
	// Open and in_progress (the session claimed it) separately, so closed
	// tickets do not count against the list limit.
	var open []store.Ticket
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		ts, err := g.ctl.List(ctx, rpc.ListParams{Label: EscalationLabel, Status: st})
		if err != nil {
			g.log.Warn("gateway: list escalations", "err", err)
			return
		}
		open = append(open, ts...)
	}
	seen := map[string]bool{}
	for _, t := range open {
		if !isEscalation(t) {
			continue
		}
		seen[t.ID] = true
		if err := g.escalation(ctx, t); err != nil {
			g.log.Warn("gateway: escalation", "ticket", t.ID, "err", err)
		}
	}
	g.mu.Lock()
	var gone []string
	for id := range g.escalating {
		if !seen[id] {
			gone = append(gone, id)
		}
	}
	g.mu.Unlock()
	for _, id := range gone {
		if err := g.resolved(ctx, id); err != nil {
			g.log.Warn("gateway: escalation closed", "ticket", id, "err", err)
			continue
		}
		g.mu.Lock()
		delete(g.escalating, id)
		g.forgetQuestion(id)
		g.mu.Unlock()
	}
}

// escalation makes sure the open ticket t has a thread and a session.
func (g *Gateway) escalation(ctx context.Context, t store.Ticket) error {
	e, ok := metaEscalation(t)
	if ok && e.Gateway != g.cfg.Name {
		return nil // another gateway owns it
	}
	if !ok {
		th, err := g.ad.Open(ctx, g.settings().EscalationRoom, openingText(t))
		if err != nil {
			return fmt.Errorf("open thread: %w", err)
		}
		e = Escalation{Gateway: g.cfg.Name, Room: th.Room, Root: th.Root, Key: g.Key(th)}
		_, err = g.updateEscalation(ctx, t.ID, func(_ store.Ticket, cur *Escalation, ok bool) error {
			if ok {
				return errors.New("another thread was recorded first")
			}
			*cur = e
			return nil
		})
		if err != nil {
			g.post(ctx, th, "jobshop: this thread was not recorded on "+t.ID+" ("+err.Error()+"); it is not used.")
			return fmt.Errorf("record thread: %w", err)
		}
		g.log.Info("gateway: escalation thread", "ticket", t.ID, "key", e.Key)
	}
	th := Thread{Room: e.Room, Root: e.Root}
	g.mu.Lock()
	g.escalating[t.ID] = th
	if e.Key != g.Key(th) {
		g.keys[th] = e.Key // a replacement thread (#197)
	}
	g.mu.Unlock()
	if _, ok := metaQuestion(t); ok {
		// A question needs no chat session: a reply in the thread is the
		// answer, delivered to the asking session by Handle.
		g.mu.Lock()
		g.questions[e.Key] = t.ID
		g.mu.Unlock()
		return nil
	}
	if e.Session != "" {
		return nil
	}
	sent := time.Now()
	res, err := g.ctl.Send(ctx, rpc.SessionSendParams{Pool: g.cfg.Pool, Key: e.Key, Text: escalationPrompt(t, g.cfg.Name)})
	if err != nil {
		g.post(ctx, th, "jobshop: could not start the session: "+err.Error())
		return fmt.Errorf("send: %w", err)
	}
	g.watch(ctx, res.Session, sent, th)
	_, err = g.updateEscalation(ctx, t.ID, func(_ store.Ticket, cur *Escalation, _ bool) error {
		cur.Session = res.Session
		return nil
	})
	if err != nil {
		// The next pass resends the prompt to the same keyed session.
		return fmt.Errorf("record session: %w", err)
	}
	return nil
}

// rethread replaces an escalation thread whose root the network rejects
// (#197): it opens a new thread, records its room and root on the ticket
// (the session key stays, so it is the same session), moves the watches
// to it, and posts text there. Posts to other threads are only logged.
func (g *Gateway) rethread(ctx context.Context, old Thread, text string) {
	g.mu.Lock()
	var id string
	for tid, th := range g.escalating {
		if th == old {
			id = tid
			break
		}
	}
	g.mu.Unlock()
	if id == "" {
		return
	}
	th, err := g.ad.Open(ctx, g.settings().EscalationRoom, fmt.Sprintf("Escalation %s continues here: the previous thread (root %s) was not found.", id, old.Root))
	if err != nil {
		g.log.Warn("gateway: rethread", "ticket", id, "err", err)
		return
	}
	var key string
	_, err = g.updateEscalation(ctx, id, func(_ store.Ticket, cur *Escalation, ok bool) error {
		if !ok || cur.Gateway != g.cfg.Name || cur.Room != old.Room || cur.Root != old.Root {
			return errSkip
		}
		cur.Room, cur.Root = th.Room, th.Root
		key = cur.Key
		return nil
	})
	if err != nil {
		g.log.Warn("gateway: rethread", "ticket", id, "err", err)
		return
	}
	g.log.Info("gateway: escalation rethreaded", "ticket", id, "old_root", old.Root, "root", th.Root)
	g.mu.Lock()
	g.escalating[id] = th
	g.keys[th] = key
	for _, w := range g.watched {
		w.mu.Lock()
		if w.thread == old {
			w.thread = th
		}
		w.mu.Unlock()
	}
	g.mu.Unlock()
	if err := g.ad.Post(ctx, th, text); err != nil {
		g.log.Warn("gateway: post", "room", th.Room, "root", th.Root, "err", fmt.Sprint(err))
	}
}

// resolved posts the final message once the escalation ticket closed.
// A ticket that only lost its label keeps its thread and gets no message.
func (g *Gateway) resolved(ctx context.Context, id string) error {
	var e Escalation
	t, err := g.updateEscalation(ctx, id, func(t store.Ticket, cur *Escalation, ok bool) error {
		if !ok || cur.Gateway != g.cfg.Name || cur.Final || t.Status != store.StatusClosed {
			return errSkip
		}
		cur.Final = true
		e = *cur
		return nil
	})
	if errors.Is(err, errSkip) {
		return nil
	}
	if err != nil {
		return err
	}
	res := t.Resolution
	if res == "" {
		res = "done"
	}
	msg := fmt.Sprintf("jobshop: escalation %s closed (%s). Replies here still reach the same session.", t.ID, res)
	if q, ok := metaQuestion(t); ok {
		msg = fmt.Sprintf("jobshop: question %s closed (%s); the answer went to %s and is a comment on %s.", t.ID, res, q.Session, q.Ticket)
	}
	g.post(ctx, Thread{Room: e.Room, Root: e.Root}, msg)
	return nil
}

func openingText(t store.Ticket) string {
	if q, ok := metaQuestion(t); ok {
		return fmt.Sprintf("Question %s from %s on %s. Reply in this thread; the reply is the answer and goes to the asking session.\n\n%s", t.ID, q.Session, q.Ticket, t.Body)
	}
	if p, ok := metaPool(t); ok {
		return fmt.Sprintf("Escalation %s: pool %s is quarantined (%s). Reply in this thread.", t.ID, p.Name, p.Quarantine)
	}
	r, _ := metaRun(t)
	return fmt.Sprintf("Escalation %s: run %s (workflow %s) needs a human. Reply in this thread.", t.ID, r.ID, r.Workflow)
}

// escalationPrompt is the session's first message. Message prompts do
// not carry the pool prompt, so it says what the session is for.
func escalationPrompt(t store.Ticket, gateway string) string {
	if p, ok := metaPool(t); ok {
		return poolPrompt(t, p, gateway)
	}
	r, _ := metaRun(t)
	var b strings.Builder
	fmt.Fprintf(&b, "You are handling escalation ticket %s for run %s (workflow %s), opened by gateway %s. ", t.ID, r.ID, r.Workflow, gateway)
	b.WriteString("This conversation is a chat thread with the operator; each of your turns is posted there, so keep replies short.\n\n")
	fmt.Fprintf(&b, "Now: read the run (`jobshop run show %s`), its steps' comments and %s. ", r.ID, githubRef(t))
	b.WriteString("Reply with a short summary: what failed, why it is stuck, and the options (resume, resume at a step, reroute, file a follow-up, cancel). Then wait for the operator.\n\n")
	b.WriteString("When the operator answers, act on it: comment on the issue, ")
	fmt.Fprintf(&b, "`jobshop run resume %s [--step <name>]` or start another run, file follow-ups. ", r.ID)
	fmt.Fprintf(&b, "Then `jobshop claim %s` and `jobshop close %s` (closing takes it off the needs-human list), and say in the thread what you did. ", t.ID, t.ID)
	b.WriteString("Ask before cancelling a run. If a command needs the admin socket, give the operator the command instead.\n\n")
	b.WriteString("Escalation ticket body:\n\n")
	b.WriteString(t.Body)
	return b.String()
}

// poolPrompt is the first message for a pool quarantine escalation.
func poolPrompt(t store.Ticket, p poolMeta, gateway string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are handling escalation ticket %s, opened by gateway %s: pool %s starts no sessions because a spawn failed with an error that will repeat. ", t.ID, gateway, p.Name)
	b.WriteString("This conversation is a chat thread with the operator; each of your turns is posted there, so keep replies short.\n\n")
	b.WriteString("Now: reply with the error and the config it points at (pool repo or dir, agent command), and what the operator should change. ")
	fmt.Fprintf(&b, "Once the config is fixed, `jobshop reload` (or `jobshop pool unquarantine %s`) lets the pool start sessions again. ", p.Name)
	fmt.Fprintf(&b, "Then `jobshop claim %s` and `jobshop close %s`, and say in the thread what you did. ", t.ID, t.ID)
	b.WriteString("If a command needs the admin socket, give the operator the command instead.\n\n")
	b.WriteString("Escalation ticket body:\n\n")
	b.WriteString(t.Body)
	return b.String()
}

// forgetQuestion drops the thread of question id. g.mu must be held.
func (g *Gateway) forgetQuestion(id string) {
	for k, q := range g.questions {
		if q == id {
			delete(g.questions, k)
		}
	}
}

// answer sends a reply in a question thread as the answer. The first
// reply closes the question; later replies start a chat session as in
// any thread, since the map entry is gone once the close event arrives.
func (g *Gateway) answer(ctx context.Context, m Message, id, text string) {
	res, err := g.ctl.Answer(ctx, rpc.AnswerParams{ID: id, Text: text, From: m.Sender})
	if err != nil {
		g.log.Warn("gateway: answer", "question", id, "err", err)
		g.post(ctx, m.Thread, "jobshop: answer not delivered: "+err.Error())
		return
	}
	g.mu.Lock()
	g.forgetQuestion(id)
	g.mu.Unlock()
	if err := g.ad.React(ctx, m.Thread, m.ID, ReactQueued); err != nil {
		g.log.Debug("gateway: react", "err", err)
	}
	if res.Message == nil {
		g.post(ctx, m.Thread, "jobshop: the asking session has ended; the answer is a comment on the ticket for the next one.")
	}
}
