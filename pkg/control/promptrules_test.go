package control_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// setRules rewrites the root's prompt_rules section.
func setRules(t *testing.T, root config.Root, base []byte, rules string) {
	t.Helper()
	require.NoError(t, os.WriteFile(root.Config(), append(append([]byte{}, base...), []byte("prompt_rules:\n"+rules)...), 0o600))
}

// TestPromptRules runs a dev session with a global system rule, a dev-only
// turn rule and a plan-only rule: the ticket prompt carries the matching
// ones in order and journals their names; session prompt shows the next
// turn labeled; a respawn sends the system rule again; a reload that
// removes a rule drops it from the next turn.
func TestPromptRules(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	base, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	base = []byte(strings.Replace(string(base), "    max_sessions: 1\n", "    max_sessions: 1\n    prompt: POOL-DEV\n", 1))
	setRules(t, root, base, `  - {name: global, text: "never push to master", at: system}
  - {name: dev-only, text: DEV-RULE, match: {pools: [dev]}}
  - {name: plan-only, text: PLAN-RULE, match: {pools: [plan]}}
  - {name: hook, command: "echo HOOK-$(cat | tr -dc a-z | head -c 0)OK"}
`)
	stop := start()
	defer func() { stop() }()
	admin := dial(t, root.AdminSocket())
	show := func(id string) store.Ticket {
		var got rpc.ShowResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
		return got.Ticket
	}
	prompts := func() []string {
		b, _ := os.ReadFile(filepath.Join(root.Dir, "prompts"))
		var out []string
		for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			var s string
			if json.Unmarshal([]byte(l), &s) == nil {
				out = append(out, s)
			}
		}
		return out
	}
	runTicket := func(title string) {
		t.Helper()
		var tk store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: title, Pool: "dev"}, &tk))
		passEventually(t, root.Dir, func() bool { return show(tk.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	}

	runTicket("first")
	p := prompts()
	require.Len(t, p, 1)
	assert.True(t, strings.HasPrefix(p[0], "never push to master\n\nPOOL-DEV\n\nDEV-RULE\n\nHOOK-OK\n\nYou have claimed"), p[0])
	assert.NotContains(t, p[0], "PLAN-RULE")

	runTicket("second")
	p = prompts()
	require.Len(t, p, 2)
	assert.True(t, strings.HasPrefix(p[1], "POOL-DEV\n\nYou have claimed"), "unchanged rules not resent: %s", p[1])

	sessions := openSessions(t, admin)
	require.Len(t, sessions, 1)
	id := sessions[0].ID
	sdir := root.SessionDir(id)

	// The host journal records which rules each prompt carried.
	frames, _, err := host.ReadJournal(filepath.Join(sdir, host.JournalFile), 0)
	require.NoError(t, err)
	var applied [][]string
	for _, f := range frames {
		var m struct {
			Method string `json:"method"`
			Params struct {
				Meta map[string]any `json:"_meta"`
			} `json:"params"`
		}
		if f.From != "controller" || json.Unmarshal(f.Msg, &m) != nil || m.Method != "session/prompt" {
			continue
		}
		var names []string
		rules, _ := m.Params.Meta["jobshop/rules"].([]any)
		for _, r := range rules {
			names = append(names, r.(string))
		}
		applied = append(applied, names)
	}
	assert.Equal(t, [][]string{{"global", "dev-only", "hook"}, nil}, applied, "unchanged rules are not resent (#397)")

	// session prompt: the next ready ticket, labeled, nothing claimed.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var next store.Ticket
	stop()
	stop = start()
	admin = dial(t, root.AdminSocket())
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "third", Pool: "dev"}, &next))
	var pr rpc.SessionPromptResult
	// Right after the restart the session can briefly show neither a held
	// nor a ready ticket (a message turn); wait for the ticket turn (#230).
	passEventually(t, root.Dir, func() bool {
		pr = rpc.SessionPromptResult{}
		return admin.Call(t.Context(), rpc.MethodSessionPrompt, rpc.SessionPromptParams{ID: id}, &pr) == nil && pr.Turn == "ticket"
	}, "ticket turn after restart")
	assert.Equal(t, next.ID, pr.Ticket)
	var srcs []string
	for _, part := range pr.Parts {
		srcs = append(srcs, part.Source)
	}
	assert.Equal(t, []string{"pool", "rule:dev-only", "rule:hook", "ticket"}, srcs)
	// The new controller may already have claimed "third" and sent its
	// turn (held open by block) before this call; the rules are then
	// injected and show unchanged (#486). Otherwise it has no record of
	// what the session holds, so they are new.
	if pr.Parts[1].Status == "unchanged" {
		assert.Empty(t, pr.Rules)
		assert.Contains(t, pr.Text, "=== rule:dev-only (unchanged, not sent) ===\n")
		assert.Equal(t, store.StatusInProgress, show(next.ID).Status, "rules unchanged only after the turn was sent")
	} else {
		assert.Equal(t, []string{"dev-only", "hook"}, pr.Rules)
		assert.Contains(t, pr.Text, "=== rule:dev-only (new) ===\nDEV-RULE\n")
	}
	assert.Contains(t, pr.Text, "=== ticket ===\nYou have claimed Jobshop ticket "+next.ID)
	sess := dial(t, filepath.Join(sdir, "cli.sock"))
	assert.ErrorIs(t, sess.Call(t.Context(), rpc.MethodSessionPrompt, rpc.SessionPromptParams{ID: id}, nil), rpc.ErrForbidden)

	// Respawn: the next turn carries the system rule again.
	pid := hostPidOf(t, sdir)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id, Now: true}, nil))
	passEventually(t, root.Dir, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, "condition not met after supervisor passes")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(next.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	p = prompts()
	last := p[len(p)-1]
	assert.Contains(t, last, next.ID)
	assert.Contains(t, last, "never push to master", "system rule after a respawn")

	// Reload without dev-only: the next turn no longer has it.
	require.NoError(t, os.Remove(filepath.Join(root.Dir, "block")))
	setRules(t, root, base, "  - {name: global, text: \"never push to master\", at: system}\n")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	runTicket("fourth")
	p = prompts()
	last = p[len(p)-1]
	assert.NotContains(t, last, "DEV-RULE")
	assert.NotContains(t, last, "HOOK-OK")
	assert.True(t, strings.HasPrefix(last, "POOL-DEV\n\nPrompt rule \"dev-only\" no longer applies.\n\n"+
		"Prompt rule \"hook\" no longer applies.\n\nYou have claimed"), last)

	// config show lists the rules.
	var cs rpc.ConfigShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodConfigShow, struct{}{}, &cs))
	rules, _ := cs.Config["prompt_rules"].([]any)
	require.Len(t, rules, 1)
	assert.Equal(t, "global", rules[0].(map[string]any)["name"])
	assert.Equal(t, []string{root.Config()}, cs.Sources["prompt_rules.global"])
}
