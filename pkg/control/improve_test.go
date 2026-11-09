package control_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestImprovementSession (#249): a send keyed improve:<slug> starts a
// session in a worktree of the checkout whose first turn carries the
// improvement rule; a matrix thread does not get it; the push hook denies
// master and allows a branch. Uses a fixture pack, not packs/operator.
func TestImprovementSession(t *testing.T) {
	root, start := poolRoot(t, "1m")
	repo := filepath.Join(root.Dir, "repo")
	for _, args := range [][]string{
		{"init", "-q", "-b", "master", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
		{"-C", repo, "update-ref", "refs/remotes/origin/master", "HEAD"},
	} {
		out, err := exec.Command("git", args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	// A fixture shaped like packs/operator (JOB-2): a chat pool, an
	// after rule for improve:* keys and a push hook denying master.
	pack := filepath.Join(root.Dir, "oppack")
	for name, body := range map[string]string{
		"pack.yaml": `pools:
  operator:
    type: chat
    agent: "{{.agent}}"
    repo: "{{.checkout}}"
    ref: origin/master
    max_sessions: 0
    idle_timeout: 12h
    prompt: "You are the operator's chat assistant."
prompt_rules:
  - name: operator-improve
    file: improve.md
    at: after
    match: {pools: [operator], keys: ["improve:*"]}
hooks:
  - on: action
    actions: [push]
    pools: [operator]
    command: |
      ev=$(cat)
      if ! printf '%s' "$ev" | grep -q '"target":"[^"]'; then
        echo '{"decision":"deny","reason":"name the branch"}'
      elif printf '%s' "$ev" | grep -Eq '"target":"([^"]*[:/])?(master|main)"'; then
        echo '{"decision":"deny","reason":"never push to master"}'
      fi
`,
		"improve.md": "This is an improvement session. Run jobshop guard push --target <branch>.\n",
	} {
		require.NoError(t, os.MkdirAll(pack, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(pack, name), []byte(body), 0o600))
	}
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npacks:\n  - path: %q\n    vars: {checkout: %q, agent: fake}\n",
		os.Args[0], pack, repo)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())

	send := func(key string) rpc.SessionSendResult {
		var res rpc.SessionSendResult
		require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "operator", Key: key, Text: "fix the dev prompt"}, &res))
		return res
	}
	firstPrompt := func(session string) string {
		path := filepath.Join(root.SessionDir(session), "work", "prompts")
		var p string
		require.Eventually(t, func() bool {
			b, err := os.ReadFile(path)
			if err != nil {
				return false
			}
			line, _, _ := strings.Cut(string(b), "\n")
			return json.Unmarshal([]byte(line), &p) == nil
		}, dialWait, 50*time.Millisecond, "the agent ran in the session worktree")
		return p
	}

	imp := send("improve:dev-prompt")
	p := firstPrompt(imp.Session)
	assert.Contains(t, p, "This is an improvement session")
	assert.Contains(t, p, "jobshop guard push --target")
	assert.Less(t, strings.Index(p, "operator's chat assistant"), strings.Index(p, "This is an improvement session"), "after the pool prompt")
	out, err := exec.Command("git", "-C", filepath.Join(root.SessionDir(imp.Session), "work"), "rev-parse", "--show-toplevel", "--git-common-dir").CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), filepath.Join(repo, ".git"), "a worktree of the checkout")

	chat := send("matrix:!room:thread")
	assert.NotContains(t, firstPrompt(chat.Session), "This is an improvement session")

	s := dial(t, filepath.Join(root.SessionDir(imp.Session), "cli.sock"))
	guard := func(target string) rpc.GuardResult {
		var r rpc.GuardResult
		require.NoError(t, s.Call(ctx, rpc.MethodGuard, rpc.GuardParams{Action: "push", Target: target}, &r))
		return r
	}
	for _, target := range []string{"master", "origin/master", "HEAD:master", "refs/heads/main", ""} {
		r := guard(target)
		assert.Equal(t, domain.GuardDeny, r.Decision, "push %q", target)
		assert.NotEmpty(t, r.Reason)
	}
	assert.Equal(t, domain.GuardAllow, guard("issue-249").Decision)
	assert.Equal(t, domain.GuardAllow, guard("fix-master-red").Decision, "a branch merely naming master")
}
