package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/support/clock"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// A2A escalation gateway (#547). It mirrors an escalation chat to another
// agent over A2A instead of a chat room: the conversation's contextId is
// the ticket id, each session turn is a SendMessage, and each peer reply
// is a session.send on the escalation key. The peer only talks to the
// escalation session, which keeps the resolve tools; Desk and humans can
// still answer, and whoever resolves first wins.

// A2AState is meta.gateway.a2a.<gateway> on an escalation ticket. It is
// what a restarted gateway resumes from.
type A2AState struct {
	Context string `json:"context"`
	// Seq is the last session journal seq sent to the peer.
	Seq int64 `json:"seq,omitempty"`
	// Task and State are the peer's open task, while it works on it.
	Task  string `json:"task,omitempty"`
	State string `json:"state,omitempty"`
	Turns int    `json:"turns,omitempty"` // messages sent
	// Error is set once a failure was commented, so retries add none.
	Error string `json:"error,omitempty"`
	Done  bool   `json:"done,omitempty"` // stopped mirroring
}

// A2AConfig is one A2A gateway.
type A2AConfig struct {
	Name  string
	URL   string // the agent card is under it
	Token string
	Pool  string // chat pool when meta.escalation has none
	// Escalations are the causes sent (meta.escalation.cause), or "all".
	Escalations []string
	Poll        time.Duration // GetTask interval, default 30s
	// MaxTurns caps messages per escalation, so two agents cannot talk
	// forever. Default 20.
	MaxTurns int
	DeskURL  string
	Clock    clock.Clock
	HTTP     *http.Client
}

// A2A is a running A2A gateway.
type A2A struct {
	cfg A2AConfig
	ctl Controller
	log *slog.Logger
	cli *a2aClient

	mu     sync.Mutex
	active map[string]chan string // ticket id -> close notice ("" while open)
	wg     sync.WaitGroup
}

// NewA2A returns an A2A gateway. log may be nil.
func NewA2A(cfg A2AConfig, ctl Controller, log *slog.Logger) *A2A {
	if cfg.Poll <= 0 {
		cfg.Poll = 30 * time.Second
	}
	if cfg.MaxTurns <= 0 {
		cfg.MaxTurns = 20
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.Real{}
	}
	if cfg.HTTP == nil {
		cfg.HTTP = telemetry.HTTPClient("a2a", time.Minute)
	}
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &A2A{
		cfg: cfg, ctl: ctl, log: log, active: map[string]chan string{},
		cli: &a2aClient{base: cfg.URL, token: cfg.Token, http: cfg.HTTP},
	}
}

// Run follows escalation tickets until ctx ends.
func (a *A2A) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		_, err := a.ctl.Events(ctx, rpc.EventsSubscribeParams{Labels: []string{EscalationLabel}}, func(ev rpc.Event) {
			a.event(ctx, ev)
		})
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			a.log.WarnContext(ctx, "a2a: events", "err", err)
		}
		select {
		case <-ctx.Done():
		case <-a.cfg.Clock.After(2 * time.Second):
		}
	}
	a.wg.Wait()
	return nil
}

func (a *A2A) event(ctx context.Context, ev rpc.Event) {
	switch {
	case ev.Kind == rpc.EventKindReady:
		a.Resync(ctx)
	case ev.Kind != rpc.EventKindTicket:
	case ev.Op == "close":
		a.mu.Lock()
		ch, ok := a.active[ev.ID]
		a.mu.Unlock()
		if ok {
			select {
			case ch <- ev.Resolution + "\x00" + ev.Actor:
			default:
			}
		}
	default:
		if ev.Status == domain.StatusClosed {
			return
		}
		r, err := a.ctl.Show(ctx, ev.ID)
		if err != nil {
			a.log.WarnContext(ctx, "a2a: show", "ticket", ev.ID, "err", err)
			return
		}
		a.start(ctx, r.Ticket)
	}
}

// Resync starts a conversation for every open escalation this gateway
// routes, resuming from meta.gateway after a restart.
func (a *A2A) Resync(ctx context.Context) {
	for _, st := range []domain.Status{domain.StatusOpen, domain.StatusInProgress} {
		ts, err := a.ctl.List(ctx, rpc.ListParams{Label: EscalationLabel, Status: st})
		if err != nil {
			a.log.WarnContext(ctx, "a2a: list escalations", "err", err)
			return
		}
		for _, t := range ts {
			a.start(ctx, t)
		}
	}
}

// Routes reports whether cause is sent to the peer.
func (a *A2A) Routes(cause string) bool {
	return slices.Contains(a.cfg.Escalations, "all") || slices.Contains(a.cfg.Escalations, cause)
}

// start begins one ticket's conversation, once.
func (a *A2A) start(ctx context.Context, t domain.Ticket) {
	if t.Status == domain.StatusClosed || !slices.Contains(t.Labels, EscalationLabel) {
		return
	}
	chat, ok := metaChat(t)
	if !ok || !a.Routes(chatCause(t)) {
		return
	}
	if e, _ := metaEscalationAny(t); e.A2A[a.cfg.Name] != nil && e.A2A[a.cfg.Name].Done {
		return
	}
	a.mu.Lock()
	if _, ok := a.active[t.ID]; ok {
		a.mu.Unlock()
		return
	}
	closed := make(chan string, 1)
	a.active[t.ID] = closed
	a.mu.Unlock()
	a.wg.Go(func() {
		defer func() {
			a.mu.Lock()
			delete(a.active, t.ID)
			a.mu.Unlock()
		}()
		c := &a2aConv{a: a, ticket: t, chat: chat, closed: closed}
		c.run(ctx)
	})
}

// chatCause is meta.escalation.cause, the escalation kind.
func chatCause(t domain.Ticket) string {
	var c struct {
		Cause string `json:"cause"`
	}
	if raw, ok := ticketMeta(t)["escalation"]; ok {
		_ = json.Unmarshal(raw, &c)
	}
	return c.Cause
}

// metaEscalationAny is meta.gateway even without a chat thread.
func metaEscalationAny(t domain.Ticket) (Escalation, bool) {
	var e Escalation
	raw, ok := ticketMeta(t)[MetaPack]
	if !ok || json.Unmarshal(raw, &e) != nil {
		return e, false
	}
	return e, true
}

// a2aConv is one escalation's conversation with the peer.
type a2aConv struct {
	a      *A2A
	ticket domain.Ticket
	chat   chatMeta
	closed chan string
	st     A2AState
	// backoff is the retry delay after a failed call; 0 when healthy.
	backoff time.Duration
}

type a2aTurn struct {
	seq  int64
	text string
}

func (c *a2aConv) run(ctx context.Context) {
	a := c.a
	if e, _ := metaEscalationAny(c.ticket); e.A2A[a.cfg.Name] != nil {
		c.st = *e.A2A[a.cfg.Name]
	}
	if c.st.Context == "" {
		c.st.Context = c.ticket.ID
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	turns := make(chan a2aTurn, 16)
	go c.follow(ctx, turns)
	var pending []a2aTurn
	// One timer at a time, rebuilt only when what it waits for changes,
	// so loop turns do not pile up timers.
	var wait <-chan time.Time
	var waitKind string
	for {
		kind, d := "", time.Duration(0)
		switch {
		case c.backoff > 0:
			kind, d = "backoff "+c.backoff.String(), c.backoff
		case c.waitingOnTask():
			kind, d = "poll", a.cfg.Poll
		}
		if kind != waitKind {
			wait, waitKind = nil, kind
			if kind != "" {
				wait = a.cfg.Clock.After(d)
			}
		}
		select {
		case <-ctx.Done():
			return
		case note := <-c.closed:
			res, by, _ := strings.Cut(note, "\x00")
			c.finish(ctx, res, by)
			return
		case tr := <-turns:
			pending = append(pending, tr)
			if c.backoff == 0 && !c.waitingOnTask() {
				pending = c.flush(ctx, pending)
			}
		case <-wait:
			waitKind = "" // fired: the next turn arms a new one
			if c.waitingOnTask() {
				c.poll(ctx)
			}
			// A failed poll keeps its backoff; a failed send doubles it.
			if !c.waitingOnTask() {
				pending = c.flush(ctx, pending)
			}
		}
		if c.st.Done {
			return
		}
	}
}

// waitingOnTask: the peer holds an open task that is not asking for input.
func (c *a2aConv) waitingOnTask() bool {
	return c.st.Task != "" && !a2aTerminal(c.st.State) && c.st.State != A2AInputRequired
}

// follow reads the escalation session's turns after st.Seq.
func (c *a2aConv) follow(ctx context.Context, out chan<- a2aTurn) {
	since := c.st.Seq
	var buf strings.Builder
	for ctx.Err() == nil {
		res, err := c.a.ctl.Subscribe(ctx, rpc.SessionSubscribeParams{ID: c.chat.Session, SinceSeq: since}, func(ev rpc.SessionEvent) {
			switch ev.Type {
			case rpc.EventTurnStart:
				buf.Reset()
			case rpc.EventTool:
				if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
					buf.WriteString("\n\n")
				}
			case rpc.EventText:
				buf.WriteString(ev.Text)
			case rpc.EventTurnEnd:
				text := strings.TrimSpace(buf.String())
				buf.Reset()
				if text != "" && ev.Seq > c.st.Seq {
					select {
					case out <- a2aTurn{seq: ev.Seq, text: text}:
					case <-ctx.Done():
					}
				}
			}
		})
		since = max(since, res.Seq)
		if err == nil && res.Closed {
			return
		}
		if err != nil && ctx.Err() == nil {
			c.a.log.WarnContext(ctx, "a2a: subscribe", "session", c.chat.Session, "err", err)
		}
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
		case <-c.a.cfg.Clock.After(2 * time.Second):
		}
	}
}

// flush sends pending turns in order, stopping at the first failure or
// when the peer starts a task; what is left is returned.
func (c *a2aConv) flush(ctx context.Context, pending []a2aTurn) []a2aTurn {
	if len(pending) == 0 {
		c.backoff = 0 // nothing left to retry
	}
	for len(pending) > 0 && !c.st.Done {
		if c.st.Turns >= c.a.cfg.MaxTurns {
			c.stop(ctx, fmt.Sprintf("a2a %s: stopped after %d messages to the peer; the escalation stays open for a human.", c.a.cfg.Name, c.st.Turns))
			return nil
		}
		tr := pending[0]
		text := tr.text
		if c.st.Turns == 0 {
			text = c.opening() + "\n\n" + text
		}
		task := ""
		if c.st.State == A2AInputRequired {
			task = c.st.Task
		}
		reply, err := c.a.cli.send(ctx, c.st.Context, task, text)
		if err != nil {
			c.failed(ctx, err)
			return pending
		}
		c.backoff = 0
		c.st.Seq, c.st.Turns = tr.seq, c.st.Turns+1
		c.st.Error = ""
		pending = pending[1:]
		c.handle(ctx, reply)
		if c.waitingOnTask() {
			break
		}
	}
	return pending
}

// opening introduces the escalation to the peer on its first message.
func (c *a2aConv) opening() string {
	t := c.ticket
	s := fmt.Sprintf("Jobshop escalation %s (%s): %s", t.ID, chatCause(t), t.Title)
	if d := c.a.cfg.DeskURL; d != "" {
		s += "\nDesk: " + strings.TrimRight(d, "/") + "/?ticket=" + t.ID
	}
	return s + "\nReply in this conversation; your replies go to the escalation session, which can resolve it."
}

// handle routes a peer reply: text to the session, task state recorded.
func (c *a2aConv) handle(ctx context.Context, r a2aReply) {
	if r.Task != nil {
		c.st.Task, c.st.State = r.Task.ID, a2aState(r.Task.Status.State)
		if r.Task.ContextID != "" {
			c.st.Context = r.Task.ContextID
		}
	} else {
		c.st.Task, c.st.State = "", ""
	}
	c.save(ctx)
	switch c.st.State {
	case A2AFailed, A2ARejected, A2ACanceled:
		msg := fmt.Sprintf("a2a %s: peer task %s ended %s", c.a.cfg.Name, c.st.Task, c.st.State)
		if t := r.text(); t != "" {
			msg += ": " + t
		}
		c.stop(ctx, msg+". The escalation stays open for a human.")
		return
	case A2ASubmitted, A2AWorking:
		return // polled until it answers
	}
	text := r.text()
	if text == "" {
		return
	}
	pool := c.chat.Pool
	if pool == "" {
		pool = c.a.cfg.Pool
	}
	if _, err := c.a.ctl.Send(ctx, rpc.SessionSendParams{Pool: pool, Key: c.chat.Key, Text: text}); err != nil {
		c.a.log.WarnContext(ctx, "a2a: session send", "ticket", c.ticket.ID, "err", err)
	}
	if r.Task != nil && c.st.State == A2ACompleted {
		// A finished task: the next session turn opens a new one.
		c.st.Task, c.st.State = "", ""
		c.save(ctx)
	}
}

func (c *a2aConv) poll(ctx context.Context) {
	t, err := c.a.cli.task(ctx, c.st.Task)
	if err != nil {
		c.failed(ctx, err)
		return
	}
	c.backoff = 0
	c.handle(ctx, a2aReply{Task: &t})
}

// failed backs off and comments the first error of a streak.
func (c *a2aConv) failed(ctx context.Context, err error) {
	c.backoff = min(max(2*c.backoff, c.a.cfg.Poll), 10*time.Minute)
	c.a.log.WarnContext(ctx, "a2a: peer", "ticket", c.ticket.ID, "err", err, "retry", c.backoff)
	msg := err.Error()
	if c.st.Error == msg {
		return
	}
	c.st.Error = msg
	c.comment(ctx, fmt.Sprintf("a2a %s: peer unreachable, retrying: %s", c.a.cfg.Name, msg))
	c.save(ctx)
}

// finish tells the peer the escalation was resolved, then stops.
func (c *a2aConv) finish(ctx context.Context, res, by string) {
	if res == "" {
		res = "done"
	}
	if c.st.Turns > 0 {
		text := fmt.Sprintf("Jobshop escalation %s was resolved (%s)", c.ticket.ID, res)
		if by != "" {
			text += " by " + by
		}
		if _, err := c.a.cli.send(ctx, c.st.Context, "", text+". No further reply is needed."); err != nil {
			c.a.log.WarnContext(ctx, "a2a: final message", "ticket", c.ticket.ID, "err", err)
		}
	}
	c.st.Done, c.st.Task, c.st.State = true, "", ""
	c.save(ctx)
}

func (c *a2aConv) stop(ctx context.Context, comment string) {
	c.comment(ctx, comment)
	c.st.Done = true
	c.save(ctx)
}

func (c *a2aConv) comment(ctx context.Context, body string) {
	if err := c.a.ctl.Comment(ctx, rpc.CommentParams{ID: c.ticket.ID, Body: body}); err != nil {
		c.a.log.WarnContext(ctx, "a2a: comment", "ticket", c.ticket.ID, "err", err)
	}
}

// save writes st to meta.gateway.a2a.<name>.
func (c *a2aConv) save(ctx context.Context) {
	st := c.st
	_, err := updateGatewayMeta(ctx, c.a.ctl, c.ticket.ID, func(_ domain.Ticket, e *Escalation, _ bool) error {
		if e.A2A == nil {
			e.A2A = map[string]*A2AState{}
		}
		e.A2A[c.a.cfg.Name] = &st
		return nil
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		c.a.log.WarnContext(ctx, "a2a: record state", "ticket", c.ticket.ID, "err", err)
	}
}
