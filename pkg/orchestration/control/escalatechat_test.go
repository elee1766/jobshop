package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	packpkg "github.com/elee1766/jobshop/pkg/pack"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// escChatRoot writes a config with a pack whose workflow pr escalates to
// pool ops. chat is the escalate.chat_pool line ("" for none).
func escChatRoot(t *testing.T, chat string) (config.Root, func() func()) {
	root, start := poolRoot(t, "1m")
	pack := filepath.Join(root.Dir, "pk")
	require.NoError(t, os.MkdirAll(filepath.Join(pack, "prompts"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "esc.md"), []byte("ESC-RUN {{.ticket}} run {{.run}} wf {{.workflow}} step {{.step}} held {{.held}} ({{.held_title}}) cause {{.cause}}: {{.reason}} site {{.site}}; read {{.issue_ref}}.\n{{.body}}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "hold.md"), []byte("ESC-HOLD {{.ticket}} ({{.title}}) cause {{.cause}} held {{.held}} ({{.held_title}}) worker {{.worker}} site {{.site}}: {{.reason}}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "pool.md"), []byte("ESC-POOL {{.ticket}} pool {{.pool}} site {{.site}}: {{.quarantine}}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, packpkg.FileName), fmt.Appendf(nil, `workflows:
  pr:
    vars: [issue]
    issue: issue
    repo: acme/widget
    escalate_pool: ops
%s
    steps:
      - {id: implement, pool: dev, retries: 1, prompt: "Implement #{{.issue}}", title: "Implement #{{.issue}}"}
`, chat), 0o600))
	missing := filepath.Join(root.Dir, "no-such-repo")
	cfg := fmt.Sprintf(`packs:
  - path: pk
    vars: {site: acme}
projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
hooks:
  - on: action
    command: 'echo "{\"decision\":\"ask_human\",\"reason\":\"merges need a person\"}"'
pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}
  ops: {project: js, agent: fake, max_sessions: 0, dir: %q}
  bad: {project: js, agent: fake, max_sessions: 1, idle_timeout: 1m, repo: %q}
  escalations: {project: js, agent: fake, type: chat, dir: %q, admin: [ticket-release]}
`, os.Args[0], root.Dir, root.Dir, missing, root.Dir)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	return root, start
}

const escChatBlock = `    escalate:
      chat_pool: escalations
      prompt_file: prompts/esc.md
      pool_prompt_file: prompts/pool.md
      hold_prompt_file: prompts/hold.md`

func chatSessions(t *testing.T, admin *rpc.Client) []domain.Ticket {
	t.Helper()
	var all []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession}, &all))
	var out []domain.Ticket
	for _, s := range all {
		if s.Pool == "escalations" {
			out = append(out, s)
		}
	}
	return out
}

func promptsFile(t *testing.T, root config.Root) string {
	t.Helper()
	b, _ := os.ReadFile(filepath.Join(root.Dir, "prompts"))
	return string(b)
}

func metaEsc(t *testing.T, tk domain.Ticket) map[string]any {
	t.Helper()
	var m map[string]map[string]any
	require.NoError(t, json.Unmarshal(tk.Meta, &m))
	return m["escalation"]
}

// TestEscalationChat (#379): a run escalation starts exactly one chat
// session keyed escalation:<ticket> whose first message is the pack's
// prompt with vars filled, also across a controller restart, and closing
// the ticket sends one final message.
func TestEscalationChat(t *testing.T) {
	root, start := escChatRoot(t, escChatBlock)
	stop := start()
	admin := dial(t, root.AdminSocket())
	ctx := t.Context()

	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "7"}})
	closeStep(t, admin, r.Steps[0].ID, "failed")
	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	id := esc[0].ID
	key := "escalation:" + id

	tk := getTicket(t, admin, id).Ticket
	m := metaEsc(t, tk)
	require.NotNil(t, m, "meta.escalation is recorded")
	assert.Equal(t, key, m["key"])
	assert.NotEmpty(t, m["session"])
	assert.Equal(t, "escalations", m["pool"])
	sessions := chatSessions(t, admin)
	require.Len(t, sessions, 1)
	assert.Equal(t, m["session"], sessions[0].ID)
	f, err := domain.DecodeFields[domain.SessionFields](sessions[0])
	require.NoError(t, err)
	assert.Equal(t, key, f.Key)

	first := "ESC-RUN " + id + " run " + r.Run.ID + " wf pr step implement held " + r.Steps[0].ID + " (Implement #7) cause run: step implement failed 1/1 times"
	require.Eventually(t, func() bool { return strings.Contains(promptsFile(t, root), first) }, 20*time.Second, 50*time.Millisecond)
	assert.Contains(t, promptsFile(t, root), "site acme; read its GitHub issue (`gh issue view 7 --repo acme/widget --comments`).")
	assert.Equal(t, "run", m["cause"])
	assert.Equal(t, r.Steps[0].ID, m["held"])
	assert.Contains(t, promptsFile(t, root), "Run "+r.Run.ID+" (workflow pr) needs a human", "the ticket body is in the prompt")

	// The pack prompt is model-only context, not a visible message: the
	// session's journal shows the turn with empty text and the prompt as
	// its context.
	var firstTurn atomic.Pointer[rpc.SessionEvent]
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub := dial(t, root.AdminSocket())
	go func() {
		_, _ = sub.Subscribe(subCtx, rpc.SessionSubscribeParams{ID: sessions[0].ID}, func(ev rpc.SessionEvent) {
			if ev.Type == rpc.EventTurnStart {
				firstTurn.CompareAndSwap(nil, &ev)
			}
		})
	}()
	require.Eventually(t, func() bool { return firstTurn.Load() != nil }, 20*time.Second, 20*time.Millisecond)
	turn := firstTurn.Load()
	assert.Empty(t, turn.Text, "no visible operator message")
	assert.Contains(t, turn.Context, "ESC-RUN "+id)

	// A restart opens no second chat and sends no second first message.
	stop()
	stop = start()
	defer stop()
	admin = dial(t, root.AdminSocket())
	pass(t, root.Dir, 3)
	assert.Len(t, chatSessions(t, admin), 1)
	assert.Equal(t, 1, strings.Count(promptsFile(t, root), "ESC-RUN "+id))

	// Closing the ticket sends one final message.
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: id}, nil))
	pass(t, root.Dir, 3)
	pass(t, root.Dir, 3)
	require.Eventually(t, func() bool { return strings.Contains(promptsFile(t, root), "escalation "+id+" closed (done)") }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, 1, strings.Count(promptsFile(t, root), "escalation "+id+" closed"))
	assert.Equal(t, true, metaEsc(t, getTicket(t, admin, id).Ticket)["final"])
}

// TestEscalationChatPool: a pool quarantine escalation gets its chat with
// the pool prompt.
func TestEscalationChatPool(t *testing.T) {
	root, start := escChatRoot(t, escChatBlock)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "bad"}, &task))

	var esc domain.Ticket
	// Test controllers tick hourly (#433): a pass spawns the session that
	// quarantines the pool.
	passUntil(t, root.Dir, func() bool {
		var all []domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Label: "needs-human"}, &all))
		if len(all) != 1 {
			return false
		}
		esc = getTicket(t, admin, all[0].ID).Ticket
		return metaEsc(t, esc) != nil
	}, "the quarantine opens a chat")
	assert.Equal(t, "escalation:"+esc.ID, metaEsc(t, esc)["key"])
	require.Eventually(t, func() bool {
		p := promptsFile(t, root)
		return strings.Contains(p, "ESC-POOL "+esc.ID+" pool bad site acme: spawn: pool bad repo ")
	}, 20*time.Second, 50*time.Millisecond)
	pass(t, root.Dir, 3)
	assert.Len(t, chatSessions(t, admin), 1, "one chat per ticket")
}

// holdTask files a task as session sess (a worker) and returns its id.
func holdTask(t *testing.T, sess *rpc.Client, title string) string {
	t.Helper()
	var tk domain.Ticket
	require.NoError(t, sess.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: title}, &tk))
	return tk.ID
}

func newWorker(t *testing.T, admin *rpc.Client) (rpc.SessionNewResult, *rpc.Client) {
	t.Helper()
	var s rpc.SessionNewResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &s))
	return s, dial(t, s.Socket)
}

// TestEscalationChatHolds (#379): every path that holds work for a
// human mints one escalation chat through the pack's hold prompt, with
// the hold's cause in the template vars. A worker's needs-human label
// names the worker and its last comment; a guard ask_human names the
// session and the reason; a question gets none.
func TestEscalationChatHolds(t *testing.T) {
	root, start := escChatRoot(t, escChatBlock)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	ctx := t.Context()
	worker, w := newWorker(t, admin)

	// 1. A worker labels its own ticket.
	id := holdTask(t, w, "Needs the staging key")
	require.NoError(t, w.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: id, Body: "blocked: no access to the staging key"}, nil))
	require.NoError(t, w.Call(ctx, rpc.MethodLabelAdd, rpc.LabelParams{ID: id, Label: "needs-human"}, nil))
	m := metaEsc(t, getTicket(t, admin, id).Ticket)
	require.NotNil(t, m, "the label opened a chat")
	assert.Equal(t, "label", m["cause"])
	assert.Equal(t, id, m["held"])
	want := "ESC-HOLD " + id + " (Needs the staging key) cause label held " + id + " (Needs the staging key) worker " + worker.Ticket.ID + " site acme: blocked: no access to the staging key"
	require.Eventually(t, func() bool { return strings.Contains(promptsFile(t, root), want) }, 20*time.Second, 50*time.Millisecond)

	// 2. A guard ask_human files an approval ticket that gets a chat.
	var g rpc.GuardResult
	require.NoError(t, w.Call(ctx, rpc.MethodGuard, rpc.GuardParams{Action: "merge", Target: "PR 5"}, &g))
	require.Equal(t, domain.GuardAskHuman, g.Decision)
	gm := metaEsc(t, getTicket(t, admin, g.Ticket).Ticket)
	require.NotNil(t, gm, "the guard ask opened a chat")
	assert.Equal(t, "guard", gm["cause"])
	require.Eventually(t, func() bool {
		return strings.Contains(promptsFile(t, root), "ESC-HOLD "+g.Ticket+" (Approve merge by "+worker.Ticket.ID+") cause guard held  () worker "+worker.Ticket.ID+" site acme: merges need a person")
	}, 20*time.Second, 50*time.Millisecond)

	// 3. A question gets no chat; neither does a ticket labeled twice.
	require.NoError(t, w.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: id}, nil))
	require.NoError(t, w.Call(ctx, rpc.MethodReport, rpc.ReportParams{State: rpc.ReportWaiting, Note: "waiting", Question: "which key?"}, nil))
	require.NoError(t, w.Call(ctx, rpc.MethodLabelAdd, rpc.LabelParams{ID: id, Label: "needs-human"}, nil))
	pass(t, root.Dir, 3)
	assert.Len(t, chatSessions(t, admin), 2, "one chat per hold, none for the question")
	assert.Equal(t, 1, strings.Count(promptsFile(t, root), "ESC-HOLD "+id+" "))
}

// TestEscalationChatPowers: the escalation session, with the
// ticket-release grant only, can hand back a ticket another session
// holds, message the worker and comment; a plain worker cannot release
// another's ticket.
func TestEscalationChatPowers(t *testing.T) {
	root, start := escChatRoot(t, escChatBlock)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	ctx := t.Context()
	worker, w := newWorker(t, admin)
	other, o := newWorker(t, admin)
	id := holdTask(t, w, "held work")
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketUpdate, rpc.UpdateParams{ID: id, Pool: new("dev")}, nil))
	claimAs(t, admin, id, worker.Ticket.ID)
	require.NoError(t, w.Call(ctx, rpc.MethodLabelAdd, rpc.LabelParams{ID: id, Label: "needs-human"}, nil))
	m := metaEsc(t, getTicket(t, admin, id).Ticket)
	require.NotNil(t, m)
	chat := dial(t, filepath.Join(root.SessionDir(m["session"].(string)), "cli.sock"))

	assert.ErrorIs(t, o.Call(ctx, rpc.MethodRelease, rpc.ReleaseParams{ID: id}, nil), domain.ErrNotHolder, "no grant, not the holder")
	_ = other
	require.NoError(t, chat.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: id, Body: "operator says use the sandbox key"}, nil))
	require.NoError(t, chat.Call(ctx, rpc.MethodRelease, rpc.ReleaseParams{ID: id, Reason: "operator unblocked"}, nil))
	got := getTicket(t, admin, id).Ticket
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.Empty(t, got.Holder)
	// Other admin methods stay closed to the chat.
	assert.ErrorIs(t, chat.Call(ctx, rpc.MethodTicketReopen, rpc.IDParams{ID: id}, nil), rpc.ErrForbidden)
	assert.ErrorIs(t, chat.Call(ctx, rpc.MethodDrain, struct{}{}, nil), rpc.ErrForbidden)
	// It reaches the worker session with send (interrupt allowed).
	var msg domain.Ticket
	require.NoError(t, chat.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{
		Kind: domain.KindMessage, Title: "use the sandbox key",
		Fields: json.RawMessage(`{"to":"` + worker.Ticket.ID + `","interrupt":true}`),
	}, &msg))
	assert.Equal(t, domain.KindMessage, msg.Kind)
}

// TestEscalationChatSweepsLabel: a ticket labeled needs-human by the
// admin (no event path) gets its chat from the sweep.
func TestEscalationChatSweepsLabel(t *testing.T) {
	root, start := escChatRoot(t, escChatBlock)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	var tk domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "stuck", Pool: "ops", Labels: []string{"needs-human"}}, &tk))
	passUntil(t, root.Dir, func() bool { return metaEsc(t, getTicket(t, admin, tk.ID).Ticket) != nil }, "the sweep opens a chat")
	assert.Equal(t, "label", metaEsc(t, getTicket(t, admin, tk.ID).Ticket)["cause"])
}

// TestEscalationNoChat: a workflow without an escalate chat pool files the
// ticket and starts no session.
func TestEscalationNoChat(t *testing.T) {
	root, start := escChatRoot(t, "")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "7"}})
	closeStep(t, admin, r.Steps[0].ID, "failed")
	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	pass(t, root.Dir, 3)
	assert.Empty(t, chatSessions(t, admin))
	assert.Nil(t, metaEsc(t, getTicket(t, admin, esc[0].ID).Ticket))
}
