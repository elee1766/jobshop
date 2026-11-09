package control_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/cli"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// guideBlocks returns the fenced blocks of the "A minimal pack" section of
// docs/packs.md, in order: child pack, prompts/work.md, jobshop.yaml,
// root pack.yaml, and the shell lines to try it.
func guideBlocks(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile("../../docs/packs.md")
	require.NoError(t, err)
	_, sec, ok := strings.Cut(string(b), "\n## A minimal pack\n")
	require.True(t, ok, "docs/packs.md has no minimal pack section")
	sec, _, _ = strings.Cut(sec, "\n## ")
	var blocks []string
	var cur []string
	in := false
	for line := range strings.SplitSeq(sec, "\n") {
		switch {
		case !in && strings.HasPrefix(line, "```"):
			in, cur = true, nil
		case in && line == "```":
			in = false
			blocks = append(blocks, strings.Join(cur, "\n")+"\n")
		case in:
			cur = append(cur, line)
		}
	}
	require.Len(t, blocks, 5, "child pack, work.md, jobshop.yaml, root pack, try-it")
	return blocks
}

// TestNewUserPath follows docs/packs.md for a new user (#494): init a
// root, write the guide's packs and jobshop.yaml, serve, start the
// guide's run, and wait for it to finish done. The pack and config come
// from the doc, so a guide that drifts from the code fails here. Only the
// agent command (the fake agent) and the checkout (a temp repo) differ.
func TestNewUserPath(t *testing.T) {
	root, start := poolRoot(t, "1m")
	blocks := guideBlocks(t)

	// 1. jobshop init.
	require.NoError(t, os.Remove(root.Config()))
	require.Equal(t, 0, cli.Main([]string{"init", root.Dir}))

	// 2. The guide's pack in <root>/packs/main.
	pack := filepath.Join(root.Dir, "packs", "main")
	require.NoError(t, os.MkdirAll(filepath.Join(pack, "prompts"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "pack.yaml"), []byte(blocks[0]), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pack, "prompts", "work.md"), []byte(blocks[1]), 0o600))

	// The guide's pool works a checkout at origin/main.
	repo := filepath.Join(root.Dir, "widget")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
		{"-C", repo, "update-ref", "refs/remotes/origin/main", "HEAD"},
	} {
		out, err := exec.Command("git", args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}

	require.NoError(t, os.WriteFile(root.Config(), []byte(blocks[2]), 0o600))
	// The guide's root pack, with the agent and checkout swapped.
	var cfg map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(blocks[3]), &cfg))
	agents, ok := cfg["agents"].(map[string]any)
	require.True(t, ok, "guide root pack has agents")
	require.Len(t, agents, 1)
	for name := range agents {
		agents[name] = map[string]any{"command": os.Args[0], "args": []string{"fake-agent"}}
	}
	packs, ok := cfg["packs"].([]any)
	require.True(t, ok, "guide root pack has packs")
	require.Len(t, packs, 1)
	entry, ok := packs[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "packs/main", entry["path"])
	vars, ok := entry["vars"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, vars, "checkout")
	vars["checkout"] = repo
	out, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.PackFileName), out, 0o600))
	_, err = config.LoadFile(root.Config())
	require.NoError(t, err, "the guide's config loads")

	// 3. jobshop serve, with in-process hosts and the fake agent.
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())
	runs, _ := subscribe(ctx, t, root.AdminSocket(), rpc.EventsSubscribeParams{Kinds: []string{rpc.EventKindRun}})

	// 4. The guide's `jobshop run start` line, through the CLI.
	var line string
	for l := range strings.SplitSeq(blocks[4], "\n") {
		if strings.HasPrefix(l, "jobshop run start ") {
			line = l
		}
	}
	require.NotEmpty(t, line, "guide has a run start line")
	args := []string{"--root", root.Dir, "--socket", root.AdminSocket()}
	args = append(args, shellWords(t, strings.TrimPrefix(line, "jobshop "))...)
	require.Equal(t, 0, cli.Main(args), "jobshop %s", line)

	// The same start with the same key returns the run it made.
	var p rpc.RunStartParams
	p.Vars = map[string]string{}
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "start" && i+1 < len(args):
			p.Workflow = args[i+1]
		case args[i] == "--var" && i+1 < len(args):
			k, v, _ := strings.Cut(args[i+1], "=")
			p.Vars[k] = v
		case args[i] == "--key" && i+1 < len(args):
			p.Key = args[i+1]
		}
	}
	r := startRun(t, admin, p)
	require.True(t, r.Existing)
	require.Len(t, r.Steps, 2)

	// 5. The run finishes done: wait for its close event.
	done := func() bool {
		for _, ev := range runs.get() {
			if ev.ID == r.Run.ID && ev.Status == domain.StatusClosed {
				return true
			}
		}
		return false
	}
	passEventually(t, root.Dir, done, "the guide's run did not finish")
	got := showRun(t, admin, r.Run.ID)
	require.Equal(t, "done", got.Run.Resolution)
	for _, s := range got.Steps {
		require.Equal(t, domain.StatusClosed, s.Status, s.Title)
	}
}

// shellWords splits a command line on spaces, honoring double quotes.
func shellWords(t *testing.T, s string) []string {
	t.Helper()
	var out []string
	var cur strings.Builder
	quoted, has := false, false
	for _, c := range s {
		switch {
		case c == '"':
			quoted, has = !quoted, true
		case c == ' ' && !quoted:
			if has {
				out = append(out, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(c)
			has = true
		}
	}
	require.False(t, quoted, "unbalanced quotes in %q", s)
	if has {
		out = append(out, cur.String())
	}
	return out
}
