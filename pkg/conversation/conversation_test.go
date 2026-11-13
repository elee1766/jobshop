package conversation_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/conversation"
	"github.com/gfx-labs/jobshop/pkg/conversation/attention"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func session(id, key string, mods ...func(*domain.SessionFields)) domain.Ticket {
	f := domain.SessionFields{Key: key}
	for _, m := range mods {
		m(&f)
	}
	raw, _ := json.Marshal(f)
	return domain.Ticket{ID: id, Kind: domain.KindSession, Status: domain.StatusOpen, Fields: raw}
}

func TestNamesAreUniqueAcrossNamedChats(t *testing.T) {
	t.Parallel()
	name, err := conversation.NormalizeName("  Release   plan ")
	require.NoError(t, err)
	assert.Equal(t, "Release plan", name)
	_, err = conversation.NormalizeName(" \n ")
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = conversation.NormalizeName(strings.Repeat("x", conversation.MaxDisplayName+1))
	require.ErrorIs(t, err, domain.ErrInvalid)

	desk := session("a", "desk:1", func(f *domain.SessionFields) { f.DisplayName = "release plan" })
	converted := session("b", "escalation:js-9", func(f *domain.SessionFields) { f.Named, f.DisplayName = true, "Other" })
	unnamed := session("c", "escalation:js-8", func(f *domain.SessionFields) { f.DisplayName = "Third" })
	all := []domain.Ticket{desk, converted, unnamed}
	require.ErrorIs(t, conversation.CheckNameFree(all, "", name), domain.ErrConflict, "case-insensitive")
	require.NoError(t, conversation.CheckNameFree(all, "a", name), "renaming a chat to its own name")
	require.ErrorIs(t, conversation.CheckNameFree(all, "", "other"), domain.ErrConflict, "a converted chat is named")
	require.NoError(t, conversation.CheckNameFree(all, "", "Third"), "an unconverted chat's label is not a name")
}

func TestManagedAndTargets(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"desk:1", "escalation:js-1", "attention:run:js-1"} {
		assert.True(t, conversation.Managed(k), k)
	}
	assert.False(t, conversation.Managed("matrix:room"))
	_, err := conversation.Target(session("s", "matrix:room"))
	require.ErrorIs(t, err, domain.ErrInvalid)
	task := session("s", "desk:1")
	task.Kind = domain.KindTask
	_, err = conversation.Target(task)
	require.ErrorIs(t, err, domain.ErrInvalid)
	f, err := conversation.Target(session("s", "desk:1"))
	require.NoError(t, err)
	assert.Equal(t, "desk:1", f.Key)
}

func TestConvertOnlyOpenEscalationChats(t *testing.T) {
	t.Parallel()
	esc := session("e", "escalation:js-1", func(f *domain.SessionFields) { f.Prestart, f.ReadOnly = true, true })
	ef, _ := conversation.Target(esc)
	require.NoError(t, conversation.CheckConvert(esc, ef))
	conversation.Convert(&ef, "Kept")
	assert.True(t, ef.Named)
	assert.Equal(t, "Kept", ef.DisplayName)
	assert.False(t, ef.Prestart || ef.ReadOnly)
	assert.Equal(t, "escalation:js-1", ef.Key, "the routing key stays")

	closed := esc
	closed.Status = domain.StatusClosed
	cf, _ := conversation.Target(closed)
	require.ErrorIs(t, conversation.CheckConvert(closed, cf), domain.ErrInvalid)
	desk := session("d", "desk:1")
	df, _ := conversation.Target(desk)
	require.ErrorIs(t, conversation.CheckConvert(desk, df), domain.ErrInvalid)
}

// Resolve closes the ticket the escalation chat answers for, so it must be
// an unconverted chat and still the ticket's current session.
func TestResolveTargetsTheEscalatedTicket(t *testing.T) {
	t.Parallel()
	esc := session("e", "escalation:js-1")
	f, _ := conversation.Target(esc)
	id, err := conversation.ResolveTarget(f)
	require.NoError(t, err)
	assert.Equal(t, "js-1", id)

	converted := session("e", "escalation:js-1", func(f *domain.SessionFields) { f.Named = true })
	cf, _ := conversation.Target(converted)
	_, err = conversation.ResolveTarget(cf)
	require.ErrorIs(t, err, domain.ErrInvalid)
	df, _ := conversation.Target(session("d", "desk:1"))
	_, err = conversation.ResolveTarget(df)
	require.ErrorIs(t, err, domain.ErrInvalid)

	target := domain.Ticket{ID: "js-1", Meta: json.RawMessage(`{"escalation":{"key":"escalation:js-1","session":"e"}}`)}
	require.NoError(t, conversation.CheckCurrent(target, "e"))
	require.ErrorIs(t, conversation.CheckCurrent(target, "other"), domain.ErrConflict, "a newer chat took over")
	require.ErrorIs(t, conversation.CheckCurrent(domain.Ticket{ID: "js-1"}, "e"), domain.ErrConflict)
}

func TestRestoreKeepsTheConversation(t *testing.T) {
	t.Parallel()
	at := time.Now()
	f := domain.SessionFields{Key: "desk:1", ACPSession: "acp-1", Archived: &at, Ended: &domain.SessionEnded{}}
	conversation.Restore(&f)
	assert.Nil(t, f.Archived)
	assert.Nil(t, f.Ended)
	assert.True(t, f.Suspended)
	assert.Equal(t, "acp-1", f.ACPSession)
}

func TestCheckNew(t *testing.T) {
	t.Parallel()
	require.NoError(t, conversation.CheckNew(rpc.ChatManageParams{Text: "hi"}))
	require.NoError(t, conversation.CheckNew(rpc.ChatManageParams{Attachments: []rpc.AttachmentContent{{Name: "a"}}}))
	require.ErrorIs(t, conversation.CheckNew(rpc.ChatManageParams{Text: " "}), domain.ErrInvalid)
	require.ErrorIs(t, conversation.CheckNew(rpc.ChatManageParams{Text: strings.Repeat("x", conversation.MaxText+1)}), domain.ErrInvalid)
	require.ErrorIs(t, conversation.CheckNew(rpc.ChatManageParams{Text: "hi", Context: strings.Repeat("x", conversation.MaxText+1)}), domain.ErrInvalid)
}

func TestCheckView(t *testing.T) {
	t.Parallel()
	ok := func(action string, args map[string]string) rpc.DeskCommandParams {
		return rpc.DeskCommandParams{Action: action, Args: args}
	}
	require.NoError(t, conversation.CheckView(ok("navigate", map[string]string{"view": "trace"})))
	require.NoError(t, conversation.CheckView(ok("filter", map[string]string{"window": "/inflight", "repo": "a/b"})))
	require.NoError(t, conversation.CheckView(ok("select", map[string]string{"refs": "js-1"})))
	for name, p := range map[string]rpc.DeskCommandParams{
		"unknown action": ok("delete", nil),
		"unknown arg":    ok("navigate", map[string]string{"view": "home", "x": "y"}),
		"unknown view":   ok("navigate", map[string]string{"view": "nope"}),
		"unknown window": ok("expand", map[string]string{"window": "nope"}),
		"filter waiting": ok("filter", map[string]string{"window": "waiting"}),
		"no refs":        ok("attach", map[string]string{"refs": " "}),
		"long arg":       ok("navigate", map[string]string{"view": "home", "ref": strings.Repeat("x", 2049)}),
	} {
		require.ErrorIs(t, conversation.CheckView(p), domain.ErrInvalid, name)
	}
	require.NoError(t, conversation.CheckViewTarget(domain.SessionFields{Key: "desk:1"}))
	require.NoError(t, conversation.CheckViewTarget(domain.SessionFields{Key: "escalation:1", Named: true}))
	require.ErrorIs(t, conversation.CheckViewTarget(domain.SessionFields{Key: "attention:run:1"}), domain.ErrInvalid)
}

func TestReadOnly(t *testing.T) {
	t.Parallel()
	assert.True(t, conversation.ReadOnlyMethod(rpc.MethodTicketShow))
	assert.True(t, conversation.ReadOnlyMethod(rpc.MethodSessionRename))
	assert.False(t, conversation.ReadOnlyMethod(rpc.MethodSessionSend))
	assert.False(t, conversation.ReadOnlyMethod(rpc.MethodTicketClose))
	edit, execute := acp.ToolKindEdit, acp.ToolKindExecute
	assert.False(t, conversation.ReadOnlyTool(&edit))
	assert.True(t, conversation.ReadOnlyTool(&execute))
	assert.True(t, conversation.ReadOnlyTool(nil))
}

func TestSummarize(t *testing.T) {
	t.Parallel()
	id, s, ok := conversation.Summarize("attention:question:js-7", "  It asks which branch.\n\nRecommended option: 2\n")
	require.True(t, ok)
	assert.Equal(t, "js-7", id)
	assert.Equal(t, "2", s.Recommend)
	assert.Equal(t, "It asks which branch.\n\nRecommended option: 2", s.Context)
	assert.Equal(t, "It asks which branch. Recommended option: 2", s.Row)

	_, s, _ = conversation.Summarize("attention:run:js-7", "Recommended option: 9")
	assert.Empty(t, s.Recommend, "out of range")
	_, _, ok = conversation.Summarize("attention:run:js-7", " ")
	assert.False(t, ok)
	_, _, ok = conversation.Summarize("desk:1", "hello")
	assert.False(t, ok)
	_, _, ok = conversation.Summarize("attention:pool:dev", "hello")
	assert.False(t, ok, "a pool item has no ticket to record on")
}

func TestSweepClosesChatsWhoseItemLeft(t *testing.T) {
	t.Parallel()
	open := []domain.Ticket{
		session("run", "attention:run:r1", func(f *domain.SessionFields) { f.Prestart = true }),
		session("moved", "attention:question:q1"),
		session("gone", "attention:ticket:t9"),
		session("followed", "attention:question:q2", func(f *domain.SessionFields) { f.Follow = &domain.QuestionFollow{} }),
		session("desk", "desk:1"),
		session("esc", "escalation:js-1"),
	}
	items := []rpc.AttentionItem{{Kind: rpc.AttentionRun, ID: "r1"}, {Kind: rpc.AttentionWait, ID: "q1"}}
	left, prestarted := conversation.Sweep(items, open, func(_ domain.Ticket, f domain.SessionFields) bool { return f.Follow != nil })
	var ids []string
	for _, l := range left {
		ids = append(ids, l.Session.ID+" "+l.Item)
	}
	assert.Equal(t, []string{"gone ticket:t9"}, ids, "a question that became a wait and a followed question stay")
	assert.Equal(t, 1, prestarted)
}

func TestPromptNamesItemOptionsAndQuestionContext(t *testing.T) {
	t.Parallel()
	q := domain.Ticket{ID: "q1", Title: "which db?", Meta: json.RawMessage(`{"github":{"repo":"a/b","issue":7},"question":{"ticket":"t1","context":"ctx","options":[{"id":1,"title":"A"}]}}`)}
	text := conversation.AttentionPrompt{
		Item: rpc.AttentionItem{Kind: rpc.AttentionQuestion, ID: "q1"}, Why: "ask", Ticket: q,
		Actions: []attention.Action{{Option: domain.QuestionOption{ID: 1, Title: "A", Description: "do a"}}},
	}.Text()
	for _, want := range []string{
		"Desk attention item: question q1", "Title: which db?", "Why: ask", "Issue: https://github.com/a/b/issues/7",
		"Option 1: A. do a", "Recommended option: N", "jobshop answer q1", "The asker's context:\n\nctx",
	} {
		assert.Contains(t, text, want)
	}
	// Why equal to the kind adds nothing.
	assert.NotContains(t, conversation.AttentionPrompt{Item: rpc.AttentionItem{Kind: "run", ID: "r"}, Why: "run"}.Text(), "Why:")
}

func TestPromptCutsALongRunResult(t *testing.T) {
	t.Parallel()
	run := domain.Ticket{ID: "r", Kind: domain.KindRun}
	res := rpc.RunResult{Run: run, Steps: []domain.Ticket{{ID: "s", Body: strings.Repeat("x", 40<<10)}}}
	text := conversation.AttentionPrompt{Item: rpc.AttentionItem{Kind: rpc.AttentionRun, ID: "r"}, Ticket: run, Run: &res}.Text()
	assert.Contains(t, text, "[cut]")
	assert.Contains(t, text, "`jobshop run show r`")
}

func TestHold(t *testing.T) {
	t.Parallel()
	hold := domain.Ticket{Kind: domain.KindTask, Labels: []string{attention.NeedsHumanLabel}}
	assert.True(t, conversation.Hold(hold))
	step := hold
	step.Kind = domain.KindStep
	assert.True(t, conversation.Hold(step))
	msg := hold
	msg.Kind = domain.KindMessage
	assert.False(t, conversation.Hold(msg))
	asked := hold
	asked.Labels = []string{attention.NeedsHumanLabel, attention.QuestionLabel}
	assert.False(t, conversation.Hold(asked))
	filed := hold
	filed.Meta = json.RawMessage(`{"question":{"ticket":"t"}}`)
	assert.False(t, conversation.Hold(filed))
	assert.False(t, conversation.Hold(domain.Ticket{Kind: domain.KindTask}))
}

type mem struct {
	store.Store
	tickets map[string]domain.Ticket
	kids    []domain.Ticket
}

func (m mem) Get(_ context.Context, id string) (domain.Ticket, error) {
	if t, ok := m.tickets[id]; ok {
		return t, nil
	}
	return domain.Ticket{}, domain.ErrNotFound
}

func (m mem) List(context.Context, store.Filter) ([]domain.Ticket, error) { return m.kids, nil }

func (m mem) Thread(context.Context, string, store.Parts) (store.Thread, error) {
	return store.Thread{}, nil
}

func TestClassifyHold(t *testing.T) {
	t.Parallel()
	failed := domain.Ticket{
		ID: "step1", Title: "Implement", Kind: domain.KindStep, Status: domain.StatusClosed, Resolution: "failed",
		Fields: json.RawMessage(`{"step":"implement","failures":2,"retries":2}`),
		Meta:   json.RawMessage(`{"report":{"state":"failed","note":"tests red"}}`),
	}
	s := mem{kids: []domain.Ticket{failed}, tickets: map[string]domain.Ticket{"hs": {ID: "hs", Kind: domain.KindSession}, "step1": failed}}
	ctx := t.Context()

	run := conversation.ClassifyHold(ctx, s, domain.Ticket{ID: "esc", Meta: json.RawMessage(`{"run":{"id":"r1","workflow":"work","escalation":"r1/1"},"github":{"repo":"a/b","issue":3}}`)})
	assert.Equal(t, conversation.CauseRun, run.Kind)
	assert.Equal(t, "step1", run.Held)
	assert.Equal(t, "implement", run.Step)
	assert.Equal(t, "step implement failed 2/2 times: tests red", run.Reason)
	assert.Contains(t, run.IssueRef, "gh issue view 3 --repo a/b")

	pool := conversation.ClassifyHold(ctx, s, domain.Ticket{ID: "p", Meta: json.RawMessage(`{"pool":{"name":"dev","quarantine":"no dir"}}`)})
	assert.Equal(t, conversation.CausePool, pool.Kind)
	assert.Equal(t, "no dir", pool.Reason)

	guard := conversation.ClassifyHold(ctx, s, domain.Ticket{ID: "g", Meta: json.RawMessage(`{"hook":{"session":"hs","ticket":"step1","reason":"rm -rf"}}`)})
	assert.Equal(t, conversation.CauseGuard, guard.Kind)
	assert.Equal(t, "hs", guard.Worker)
	assert.Equal(t, "Implement", guard.HeldTitle)

	timeout := domain.Ticket{ID: "w", Kind: domain.KindTask, Holder: "admin", Fields: json.RawMessage(`{"wait":{"reason":"ci","owner":"ann","timed_out":true,"timeout":"2026-11-08T12:00:00Z"}}`)}
	wait := conversation.ClassifyHold(ctx, s, timeout)
	assert.Equal(t, conversation.CauseWait, wait.Kind)
	assert.Equal(t, "wait (ci) passed its timeout 2026-11-08T12:00:00Z; owner ann", wait.Reason)
	assert.Empty(t, wait.Worker, "admin is an actor, not a session")

	label := conversation.ClassifyHold(ctx, s, domain.Ticket{ID: "l", Title: "stuck", Body: "needs a call"})
	assert.Equal(t, conversation.CauseLabel, label.Kind)
	assert.Equal(t, "needs a call", label.Reason)

	// A failed callback holds the run, not a step (#1246).
	s.tickets["r2"] = domain.Ticket{ID: "r2", Kind: domain.KindRun, Fields: json.RawMessage(`{"workflow":"work","state":"needs_human","hook_error":"var \"pool_prefix\" is not set"}`)}
	hook := conversation.ClassifyHold(ctx, s, domain.Ticket{ID: "esc2", Meta: json.RawMessage(`{"run":{"id":"r2","workflow":"work","escalation":"r2/1"}}`)})
	assert.Equal(t, conversation.CauseRun, hook.Kind)
	assert.Empty(t, hook.Held)
	assert.Equal(t, `its workflow callback failed: var "pool_prefix" is not set`, hook.Reason)
}

func TestPolicy(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Pools: map[string]config.Pool{"ops": {Project: "js"}},
		Escalation: &config.EscalationChat{
			ChatPool: "ops", Prompt: "run {{.run}} {{.reason}}", PoolPrompt: "pool {{.pool}}", HoldPrompt: "hold {{.held}}",
		},
	}
	x := conversation.Cause{Kind: conversation.CauseRun, Run: "r1", Reason: "boom"}
	pool, prompt, ok, err := conversation.Policy(cfg, domain.Ticket{ID: "t"}, x)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "ops", pool)
	assert.Equal(t, "run r1 boom", prompt)

	_, prompt, ok, _ = conversation.Policy(cfg, domain.Ticket{ID: "t", Project: "js"}, conversation.Cause{Kind: conversation.CauseWait, Held: "h"})
	assert.True(t, ok)
	assert.Equal(t, "hold h", prompt)

	// No chat for a hold in a project the chat pool does not serve, or for a
	// quarantine of the chat pool itself.
	_, _, ok, _ = conversation.Policy(cfg, domain.Ticket{ID: "t", Project: "other"}, conversation.Cause{Kind: conversation.CauseWait})
	assert.False(t, ok)
	_, _, ok, _ = conversation.Policy(cfg, domain.Ticket{ID: "t"}, conversation.Cause{Kind: conversation.CausePool, Pool: "ops"})
	assert.False(t, ok)
	_, _, ok, _ = conversation.Policy(&config.Config{}, domain.Ticket{ID: "t"}, x)
	assert.False(t, ok)
	assert.False(t, conversation.Configured(&config.Config{}))
	assert.True(t, conversation.Configured(cfg))
}

func TestFinalMessage(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "jobshop: escalation js-1 closed (done). Nothing more is needed unless the operator writes again.", conversation.FinalMessage("js-1", ""))
	assert.Contains(t, conversation.FinalMessage("js-1", "cancelled"), "(cancelled)")
}

func TestEscalationOf(t *testing.T) {
	t.Parallel()
	e, ok := conversation.EscalationOf(domain.Ticket{Meta: json.RawMessage(`{"escalation":{"key":"escalation:t","pool":"ops","session":"s1","final":true}}`)})
	require.True(t, ok)
	assert.Equal(t, conversation.Escalation{Key: "escalation:t", Pool: "ops", Session: "s1", Final: true}, e)
	_, ok = conversation.EscalationOf(domain.Ticket{Meta: json.RawMessage(`{"escalation":{"pool":"ops"}}`)})
	assert.False(t, ok, "no key, no chat")
	_, ok = conversation.EscalationOf(domain.Ticket{})
	assert.False(t, ok)
}
