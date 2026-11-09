package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// escYAML has a chat pool and a workflow whose escalation chat goes
// there; plain has no escalate block. escController drops the chat agent
// after loading, so a chat session is a ticket with a socket and no host
// process.
const escYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
  chatagent: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
  chat: {agent: chatagent, type: chat, dir: %[1]q}
workflows:
  work:
    vars: [n]
    issue: n
    repo: acme/widget
    escalate_pool: plan
    escalate:
      chat_pool: chat
      prompt: "Escalation {{.ticket}} for run {{.run}} ({{.workflow}}). Read {{.issue_ref}}.\n\n{{.body}}"
      pool_prompt: "Pool {{.pool}} is quarantined in {{.ticket}}: {{.error}}"
    steps:
      - {id: a, pool: dev, retries: 1, title: "A {{.n}}", prompt: a}
  plain:
    vars: [n]
    escalate_pool: plan
    steps:
      - {id: a, pool: dev, retries: 1, title: "P {{.n}}", prompt: a}
`

// escController builds a controller over dir without serving. Calling it
// again on the same dir is a controller restart over the same database.
func escController(t *testing.T, dir string) *Controller {
	t.Helper()
	root := config.Root{Dir: dir}
	cfg, err := config.Load(root)
	require.NoError(t, err)
	delete(cfg.Agents, "chatagent")
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	live := new(atomic.Pointer[liveConfig])
	live.Store(lc)
	clk := newScanClock(time.Date(2026, 11, 4, 8, 0, 0, 0, time.UTC))
	st, err := sqlite.Open(t.Context(), filepath.Join(dir, "db"), sqlite.Options{Now: clk.now, Prefix: func(p string) string { return p }})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	c := &Controller{
		root: root, live: live, store: st, opt: Options{Clock: clk}, log: slog.New(slog.DiscardHandler),
		sessions: map[string]context.CancelFunc{}, runners: map[string]*runner{},
		drained: map[string]bool{}, poolCrashLog: map[string][]time.Time{}, poolPausedUntil: map[string]time.Time{},
		poolQuarantined: map[string]string{},
	}
	c.ctx = ctx
	c.tracer = noop.NewTracerProvider().Tracer("test")
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	t.Cleanup(func() {
		for _, id := range c.liveSessions() {
			c.closeSession(id)
		}
		cancel()
		_ = st.Close()
	})
	return c
}

// liveSessions lists sessions with an open socket.
func (c *Controller) liveSessions() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, len(c.sessions))
	for id := range c.sessions {
		out = append(out, id)
	}
	return out
}

func escDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "js") // short: unix socket paths
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, escYAML, dir), 0o600))
	return dir
}

// escalateRunNow starts a run of workflow and fails its only step at the
// retry limit, returning the escalation ticket.
func escalateRunNow(t *testing.T, c *Controller, workflow string) domain.Ticket {
	t.Helper()
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	res, err := c.startRun(ctx, c.config(), admin, rpc.RunStartParams{Workflow: workflow, Vars: map[string]string{"n": "7"}}, nil)
	require.NoError(t, err)
	run := res.(rpc.RunResult)
	_, err = c.closeTicket(ctx, admin, rpc.CloseParams{ID: run.Steps[0].ID, Resolution: "failed"})
	require.NoError(t, err)
	esc, err := c.store.List(ctx, store.Filter{MetaPack: runMetaPack, MetaKey: "id", MetaValue: run.Run.ID})
	require.NoError(t, err)
	require.Len(t, esc, 1)
	return esc[0]
}

// chatSessions lists the sessions keyed escalation:<id>.
func chatSessions(t *testing.T, c *Controller, id string) []domain.Ticket {
	t.Helper()
	all, err := c.store.List(t.Context(), store.Filter{Kind: domain.KindSession})
	require.NoError(t, err)
	var out []domain.Ticket
	for _, s := range all {
		f, err := domain.DecodeFields[domain.SessionFields](s)
		if err == nil && f.Key == escalationKeyPrefix+id {
			out = append(out, s)
		}
	}
	return out
}

// TestEscalationChatRun: a run escalation gets exactly one chat keyed
// escalation:<ticket> whose first message is the pack prompt with vars
// filled, across a controller restart; closing the ticket sends one final
// message.
func TestEscalationChatRun(t *testing.T) {
	t.Parallel()
	dir := escDir(t)
	c := escController(t, dir)
	esc := escalateRunNow(t, c, "work")

	got, err := c.store.Get(t.Context(), esc.ID)
	require.NoError(t, err)
	e, ok := escalationMeta(got)
	require.True(t, ok)
	assert.Equal(t, escalationKeyPrefix+esc.ID, e.Key)
	assert.Equal(t, "chat", e.Pool)
	sess := chatSessions(t, c, esc.ID)
	require.Len(t, sess, 1)
	assert.Equal(t, sess[0].ID, e.Session)
	assert.Equal(t, "chat", sess[0].Pool)

	msgs := messagesTo(t, c, e.Session)
	require.Len(t, msgs, 1)
	// The pack prompt is model-only context, not a visible message.
	assert.Empty(t, msgs[0].Title)
	first := messageContext(t, msgs[0])
	var run struct{ ID string }
	require.NoError(t, jsonField(got.Meta, runMetaPack, &run))
	assert.True(t, strings.HasPrefix(first, fmt.Sprintf("Escalation %s for run %s (work). Read its GitHub issue (`gh issue view 7 --repo acme/widget --comments`).", esc.ID, run.ID)), first)
	assert.Contains(t, first, "needs a human: a step failed at its retry limit", "${body} is the ticket body")

	// Sweeps and a restart over the same database open no second chat
	// and resend nothing.
	c.sweepEscalationChats(t.Context(), c.config())
	c2 := escController(t, dir)
	c2.sweepEscalationChats(t.Context(), c2.config())
	c2.sweepEscalationChats(t.Context(), c2.config())
	assert.Len(t, chatSessions(t, c2, esc.ID), 1)
	assert.Len(t, messagesTo(t, c2, e.Session), 1)

	// Closing the ticket sends one final message, once.
	_, err = c2.closeTicket(t.Context(), scope{admin: true, actor: "admin"}, rpc.CloseParams{ID: esc.ID})
	require.NoError(t, err)
	c2.sweepEscalationChats(t.Context(), c2.config())
	msgs = messagesTo(t, c2, e.Session)
	require.Len(t, msgs, 2)
	assert.Contains(t, msgs[1].Title, "escalation "+esc.ID+" closed (done)")
	got, err = c2.store.Get(t.Context(), esc.ID)
	require.NoError(t, err)
	e, _ = escalationMeta(got)
	assert.True(t, e.Final)
}

// TestEscalationChatRepair: a chat recorded on the ticket but not started
// (a crash between the two writes) is started by the next sweep with the
// prompt, and only once.
func TestEscalationChatRepair(t *testing.T) {
	t.Parallel()
	dir := escDir(t)
	c := escController(t, dir)
	esc := escalateRunNow(t, c, "work")
	got, err := c.store.Get(t.Context(), esc.ID)
	require.NoError(t, err)
	e, _ := escalationMeta(got)
	// Undo the session record, close the session: as if the controller
	// crashed after recording the key.
	_, err = c.closeTicket(t.Context(), scope{admin: true, actor: "admin"}, rpc.CloseParams{ID: e.Session})
	require.NoError(t, err)
	_, err = c.updateEscalationMeta(t.Context(), esc.ID, func(_ domain.Ticket, m *EscalationMeta, _ bool) error {
		m.Session = ""
		return nil
	})
	require.NoError(t, err)

	c2 := escController(t, dir)
	c2.sweepEscalationChats(t.Context(), c2.config())
	c2.sweepEscalationChats(t.Context(), c2.config())
	got, err = c2.store.Get(t.Context(), esc.ID)
	require.NoError(t, err)
	e2, _ := escalationMeta(got)
	require.NotEmpty(t, e2.Session)
	assert.NotEqual(t, e.Session, e2.Session, "a new session for the same key")
	assert.Len(t, messagesTo(t, c2, e2.Session), 1, "the prompt, once")
}

// TestEscalationChatNoConfig: a workflow without escalate (and no
// top-level escalation block) files the ticket only.
func TestEscalationChatNoConfig(t *testing.T) {
	t.Parallel()
	c := escController(t, escDir(t))
	esc := escalateRunNow(t, c, "plain")
	c.sweepEscalationChats(t.Context(), c.config())
	got, err := c.store.Get(t.Context(), esc.ID)
	require.NoError(t, err)
	_, ok := escalationMeta(got)
	assert.False(t, ok)
	assert.Empty(t, chatSessions(t, c, esc.ID))
}

// TestEscalationChatPool: a pool quarantine escalation filed in plan gets
// the chat with the pool prompt of the workflow escalating to plan.
func TestEscalationChatPool(t *testing.T) {
	t.Parallel()
	c := escController(t, escDir(t))
	c.quarantinePool(t.Context(), "dev", "spawn: repo /x: no such file")
	esc, err := c.store.List(t.Context(), store.Filter{MetaPack: poolMetaPack, MetaKey: "name", MetaValue: "dev"})
	require.NoError(t, err)
	require.Len(t, esc, 1)
	got, err := c.store.Get(t.Context(), esc[0].ID)
	require.NoError(t, err)
	e, ok := escalationMeta(got)
	require.True(t, ok)
	msgs := messagesTo(t, c, e.Session)
	require.Len(t, msgs, 1)
	assert.Equal(t, fmt.Sprintf("Pool dev is quarantined in %s: spawn: repo /x: no such file", esc[0].ID), messageContext(t, msgs[0]))
	c.quarantinePool(t.Context(), "dev", "again")
	c.sweepEscalationChats(t.Context(), c.config())
	assert.Len(t, chatSessions(t, c, esc[0].ID), 1)
}

// messageContext is the model-only context of a message ticket.
func messageContext(t *testing.T, m domain.Ticket) string {
	t.Helper()
	f, err := domain.DecodeFields[domain.MessageFields](m)
	require.NoError(t, err)
	return f.Context
}

// jsonField decodes meta[pack] into v.
func jsonField(meta []byte, pack string, v any) error {
	raw, ok := metaMap(meta)[pack]
	if !ok {
		return fmt.Errorf("no %s", pack)
	}
	return json.Unmarshal(raw, v)
}
