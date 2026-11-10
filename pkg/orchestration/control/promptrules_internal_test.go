package control

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// rulesController is a Controller with a store and the given config, for
// prompt assembly without sessions.
func rulesController(t *testing.T, cfgYAML string) (*Controller, *bytesLog) {
	t.Helper()
	dir := t.TempDir()
	root := config.Root{Dir: dir}
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(cfgYAML), 0o600))
	cfg, err := config.Load(root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	live := new(atomic.Pointer[liveConfig])
	live.Store(lc)
	st, err := sqlite.Open(t.Context(), filepath.Join(dir, "db"), sqlite.Options{Prefix: func(p string) string { return p }})
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	logs := &bytesLog{}
	c := &Controller{root: root, live: live, store: st, log: slog.New(slog.NewTextHandler(logs, nil)), runners: map[string]*runner{}}
	c.ctx = t.Context()
	return c, logs
}

type bytesLog struct {
	b strings.Builder
}

func (l *bytesLog) Write(p []byte) (int, error) { return l.b.Write(p) }

const rulesBase = "projects:\n  js: {}\nagents:\n  fake:\n    command: x\npools:\n" +
	"  dev:\n    agent: fake\n    dir: /tmp\n    prompt: POOL-DEV\n" +
	"  plan:\n    agent: fake\n    dir: /tmp\n    prompt: POOL-PLAN\n" +
	"  chat:\n    agent: fake\n    dir: /tmp\n    type: chat\n    prompt: POOL-CHAT\n"

func sources(parts []PromptPart) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p.Source)
	}
	return out
}

// TestAssembleOrder: system, before, pool, after, turn, body; rule order
// within a position; system only on a host's first turn; before/after only
// where the pool prompt goes; match by pool, step and chat.
func TestAssembleOrder(t *testing.T) {
	t.Parallel()
	c, _ := rulesController(t, rulesBase+`prompt_rules:
  - {name: t1, text: T1}
  - {name: sys, text: SYS, at: system}
  - {name: aft, text: AFT, at: after}
  - {name: bef, text: BEF, at: before}
  - {name: t2, text: T2, match: {pools: [dev]}}
  - {name: impl, text: IMPL, match: {steps: [implement], workflows: [pr-*]}}
  - {name: chatonly, text: CHAT, match: {chat: true}}
`)
	ctx := t.Context()
	run, err := store.Create(ctx, c.store, "t", store.NewTicket{
		Project: "js", Kind: domain.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"pr-work","state":"running"}`),
	})

	require.NoError(t, err)
	step, err := store.Create(ctx, c.store, "t", store.NewTicket{
		Project: "js", Kind: domain.KindStep, Title: "s", Pool: "dev", ParentID: run.ID,
		Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"implement"}`, run.ID)),
	})

	require.NoError(t, err)
	body := PromptPart{Source: "ticket", Text: "BODY"}

	first := c.assemble(ctx, turnSpec{pool: "dev", kind: turnTicket, ticket: &step, system: true, poolPrompt: true}, body)
	assert.Equal(t, []string{"rule:sys", "rule:bef", "pool", "rule:aft", "rule:t1", "rule:t2", "rule:impl", "ticket"}, sources(first))
	assert.Equal(t, "SYS\n\nBEF\n\nPOOL-DEV\n\nAFT\n\nT1\n\nT2\n\nIMPL\n\nBODY", joinParts(first))

	later := c.assemble(ctx, turnSpec{pool: "dev", kind: turnTicket, ticket: &step, poolPrompt: true}, body)
	assert.NotContains(t, sources(later), "rule:sys", "system rules once per host")

	plan := c.assemble(ctx, turnSpec{pool: "plan", kind: turnTicket, ticket: &step, poolPrompt: true}, body)
	assert.Equal(t, []string{"rule:bef", "pool", "rule:aft", "rule:t1", "rule:impl", "ticket"}, sources(plan), "t2 is dev only")

	// A chat message turn after the first: no pool prompt, no before/after,
	// no ticket-only rules.
	msg := c.assemble(ctx, turnSpec{pool: "chat", kind: turnMessage}, PromptPart{Source: "messages", Text: "M"})
	assert.Equal(t, []string{"rule:t1", "rule:chatonly", "messages"}, sources(msg))
	firstMsg := c.assemble(ctx, turnSpec{pool: "chat", kind: turnMessage, system: true, poolPrompt: true}, PromptPart{Source: "messages", Text: "M"})
	assert.Equal(t, []string{"rule:sys", "rule:bef", "pool", "rule:aft", "rule:t1", "rule:chatonly", "messages"}, sources(firstMsg))
	assert.Equal(t, []string{"sys", "bef", "aft", "t1", "chatonly"}, ruleNames(firstMsg))
}

// TestRuleCommand: the hook gets the event on stdin and its stdout follows
// the rule's text; a failure or timeout adds nothing and logs WARN; output
// is capped.
func TestRuleCommand(t *testing.T) {
	t.Parallel()
	c, logs := rulesController(t, rulesBase+fmt.Sprintf(`prompt_rules:
  - name: echo
    text: STATIC
    command: "cat > event.json; echo HOOK-OUT"
  - {name: fails, command: "echo partial; exit 3"}
  - {name: slow, command: "sleep 5; echo late", timeout: 200ms}
  - {name: big, command: "head -c %d /dev/zero | tr '\\0' x"}
`, config.MaxRuleOutput+100))
	ctx := t.Context()
	tk, err := store.Create(ctx, c.store, "t", store.NewTicket{
		Project: "js", Kind: domain.KindTask, Title: "x", Pool: "dev",
		Meta: json.RawMessage(`{"github":{"repo":"gfx-labs/jobshop","issue":205}}`),
	})

	require.NoError(t, err)
	start := time.Now()
	parts := c.assemble(ctx, turnSpec{session: "js-s", pool: "dev", agent: "fake", project: "js", kind: turnTicket, ticket: &tk}, PromptPart{Source: "ticket"})
	assert.Less(t, time.Since(start), 4*time.Second, "slow hook cut at its timeout")
	assert.Equal(t, []string{"rule:echo", "rule:big", "ticket"}, sources(parts), "failed and timed-out hooks add nothing")
	assert.Equal(t, "STATIC\n\nHOOK-OUT", parts[0].Text)
	assert.Len(t, parts[1].Text, config.MaxRuleOutput)

	var ev RuleEvent
	b, err := os.ReadFile(filepath.Join(c.root.Dir, "event.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &ev))
	assert.Equal(t, RuleEvent{
		Rule: "echo", At: config.RuleTurn, Session: "js-s", Pool: "dev", Agent: "fake", Project: "js",
		Turn: turnTicket, Ticket: tk.ID, Repo: "gfx-labs/jobshop",
	}, ev)

	out := logs.b.String()
	assert.Contains(t, out, `msg="prompt rule command failed" rule=fails`)
	assert.Contains(t, out, `msg="prompt rule command timed out" rule=slow`)
	assert.Contains(t, out, `msg="prompt rule command output truncated" rule=big`)
}

// TestRuleMemory: a memory rule injects the global, repo and pool files at
// system with the skills index; an over-cap file is cut with a WARN; an
// empty or missing dir adds nothing (#266).
func TestRuleMemory(t *testing.T) {
	t.Parallel()
	mem := t.TempDir()
	for p, s := range map[string]string{
		"memory/global.md":                 "GLOBAL-MEM",
		"memory/repo/gfx-labs__jobshop.md": "REPO-MEM",
		"memory/pool/dev.md":               "DEV-MEM " + strings.Repeat("x", 50),
		"skills/logs/SKILL.md":             "---\nname: logs\ndescription: Read host logs\n---\n",
	} {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(mem, p)), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(mem, p), []byte(s), 0o600))
	}
	c, logs := rulesController(t, rulesBase+fmt.Sprintf(`prompt_rules:
  - {name: mem, at: system, memory: {dir: %q, cap: 20}}
  - {name: none, at: system, memory: {dir: ""}}
  - {name: gone, at: system, memory: {dir: %q}}
`, mem, filepath.Join(mem, "absent")))
	ctx := t.Context()
	tk, err := store.Create(ctx, c.store, "t", store.NewTicket{
		Project: "js", Kind: domain.KindTask, Title: "x", Pool: "dev",
		Meta: json.RawMessage(`{"github":{"repo":"gfx-labs/jobshop","issue":266}}`),
	})

	require.NoError(t, err)
	parts := c.assemble(ctx, turnSpec{pool: "dev", kind: turnTicket, ticket: &tk, system: true}, PromptPart{Source: "ticket"})
	assert.Equal(t, []string{"rule:mem", "ticket"}, sources(parts))
	text := parts[0].Text
	for _, s := range []string{"GLOBAL-MEM", "REPO-MEM", "DEV-MEM xxxxxxxxxxxx\n", "- logs: Read host logs (" + filepath.Join(mem, "skills/logs/SKILL.md") + ")"} {
		assert.Contains(t, text, s)
	}
	assert.NotContains(t, text, strings.Repeat("x", 13), "cut at the cap")
	assert.Contains(t, logs.b.String(), `msg="memory file over cap, truncated" rule=mem file=memory/pool/dev.md chars=58`)

	plan := c.assemble(ctx, turnSpec{pool: "plan", kind: turnMessage, system: true}, PromptPart{Source: "messages"})
	assert.Contains(t, plan[0].Text, "GLOBAL-MEM")
	assert.NotContains(t, plan[0].Text, "REPO-MEM", "no ticket, no repo memory")
	assert.NotContains(t, plan[0].Text, "DEV-MEM")

	later := c.assemble(ctx, turnSpec{pool: "dev", kind: turnTicket, ticket: &tk}, PromptPart{Source: "ticket"})
	assert.Equal(t, []string{"ticket"}, sources(later), "system rules only on a host's first turn")
}

// TestLearnNudge (#271): an after rule with no match (as packs/github's
// learn-nudge) is in the after position of dev, refinery and operator
// ticket turns, and in a chat session's first message turn, but not in
// later chat turns. Inline config, not packs/ (JOB-2).
func TestLearnNudge(t *testing.T) {
	t.Parallel()
	c, _ := rulesController(t, `projects:
  js: {}
agents:
  a: {command: x}
pools:
  dev: {agent: a, dir: /tmp, prompt: POOL-DEV}
  refinery: {agent: a, dir: /tmp, prompt: POOL-REF}
  operator: {agent: a, dir: /tmp, type: chat, prompt: POOL-OP}
prompt_rules:
  - name: learn-nudge
    at: after
    text: "run jobshop learn. Do not record a learning for routine success."
`)
	ctx := t.Context()
	task, err := store.Create(ctx, c.store, "t", store.NewTicket{Project: "js", Kind: domain.KindTask, Title: "t"})
	require.NoError(t, err)
	body := PromptPart{Source: "ticket", Text: "BODY"}
	for _, pool := range []string{"dev", "refinery", "operator"} {
		parts := c.assemble(ctx, turnSpec{pool: pool, kind: turnTicket, ticket: &task, poolPrompt: true}, body)
		src := sources(parts)
		require.Contains(t, src, "rule:learn-nudge", pool)
		assert.Contains(t, joinParts(parts), "jobshop learn", pool)
		assert.Contains(t, joinParts(parts), "Do not record a learning for routine success", pool)
		assert.Less(t, slices.Index(src, "pool"), slices.Index(src, "rule:learn-nudge"), pool)
	}
	first := c.assemble(ctx, turnSpec{pool: "operator", kind: turnMessage, poolPrompt: true}, PromptPart{Source: "messages", Text: "M"})
	assert.Contains(t, sources(first), "rule:learn-nudge")
	later := c.assemble(ctx, turnSpec{pool: "operator", kind: turnMessage}, PromptPart{Source: "messages", Text: "M"})
	assert.NotContains(t, sources(later), "rule:learn-nudge")
}
