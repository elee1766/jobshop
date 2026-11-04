package control_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// escChatRoot writes a config with a pack whose workflow pr escalates to
// pool ops. chat is the escalate.chat_pool line ("" for none).
func escChatRoot(t *testing.T, chat string) (config.Root, func() func()) {
	root, start := poolRoot(t, "1m")
	pack := filepath.Join(root.Dir, "pk")
	require.NoError(t, os.MkdirAll(filepath.Join(pack, "prompts"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "esc.md"), []byte("ESC-RUN ${ticket} run ${run} wf ${workflow} site ${site}; read ${issue_ref}.\n${body}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "pool.md"), []byte("ESC-POOL ${ticket} pool ${pool} site ${site}: ${quarantine}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName), fmt.Appendf(nil, `workflows:
  pr:
    vars: [issue]
    issue: issue
    repo: acme/widget
    escalate_pool: ops
%s
    steps:
      - {id: implement, pool: dev, retries: 1, prompt: "Implement #${issue}", title: "Implement #${issue}"}
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
pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}
  ops: {project: js, agent: fake, max_sessions: 0, dir: %q}
  bad: {project: js, agent: fake, max_sessions: 1, idle_timeout: 1m, repo: %q}
  escalations: {project: js, agent: fake, type: chat, dir: %q}
`, os.Args[0], root.Dir, root.Dir, missing, root.Dir)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	return root, start
}

const escChatBlock = `    escalate:
      chat_pool: escalations
      prompt_file: prompts/esc.md
      pool_prompt_file: prompts/pool.md`

func chatSessions(t *testing.T, admin *rpc.Client) []store.Ticket {
	t.Helper()
	var all []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &all))
	var out []store.Ticket
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

func metaEsc(t *testing.T, tk store.Ticket) map[string]any {
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
	f, err := store.DecodeFields[store.SessionFields](sessions[0])
	require.NoError(t, err)
	assert.Equal(t, key, f.Key)

	first := "ESC-RUN " + id + " run " + r.Run.ID + " wf pr site acme; read its GitHub issue (`gh issue view 7 --repo acme/widget --comments`)."
	require.Eventually(t, func() bool { return strings.Contains(promptsFile(t, root), first) }, 20*time.Second, 50*time.Millisecond)
	assert.Contains(t, promptsFile(t, root), "Run "+r.Run.ID+" (workflow pr) needs a human", "the ticket body is in the prompt")

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
	require.Eventually(t, func() bool { return strings.Contains(promptsFile(t, root), "Ticket "+id+" is closed (done)") }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, 1, strings.Count(promptsFile(t, root), "Ticket "+id+" is closed"))
	assert.Equal(t, true, metaEsc(t, getTicket(t, admin, id).Ticket)["final"])
}

// TestEscalationChatPool: a pool quarantine escalation gets its chat with
// the pool prompt.
func TestEscalationChatPool(t *testing.T) {
	root, start := escChatRoot(t, escChatBlock)
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "bad"}, &task))

	var esc store.Ticket
	require.Eventually(t, func() bool {
		var all []store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Label: "needs-human"}, &all))
		if len(all) != 1 {
			return false
		}
		esc = getTicket(t, admin, all[0].ID).Ticket
		return metaEsc(t, esc) != nil
	}, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, "escalation:"+esc.ID, metaEsc(t, esc)["key"])
	require.Eventually(t, func() bool {
		p := promptsFile(t, root)
		return strings.Contains(p, "ESC-POOL "+esc.ID+" pool bad site acme: spawn: pool bad repo ")
	}, 20*time.Second, 50*time.Millisecond)
	pass(t, root.Dir, 3)
	assert.Len(t, chatSessions(t, admin), 1, "one chat per ticket")
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
