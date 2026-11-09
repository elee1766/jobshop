package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func TestClassifyGH(t *testing.T) {
	cases := []struct {
		args              []string
		action, repo, num string
		body              string
	}{
		{args: []string{"pr", "view", "3"}},
		{args: []string{"pr", "list", "-R", "o/n"}, repo: "o/n"},
		{args: []string{"pr", "create", "--title", "x", "--body", "hi"}, action: "pr-create", body: "hi"},
		{args: []string{"pr", "merge", "-m", "12", "--repo=o/n"}, action: "merge", repo: "o/n", num: "12"},
		{args: []string{"-R", "o/n", "pr", "merge", "12", "--squash"}, action: "merge", repo: "o/n", num: "12"},
		{args: []string{"pr", "comment", "https://github.com/a/b/pull/7", "-b", "x"}, action: "comment", repo: "a/b", num: "7", body: "x"},
		{args: []string{"issue", "close", "#4", "-Rgithub.com/o/n"}, action: "issue-close", repo: "o/n", num: "4"},
		{args: []string{"repo", "deploy-key", "add", "k.pub"}, action: "repo-write"},
		{args: []string{"api", "repos/o/n/pulls"}, repo: "o/n"},
		{args: []string{"api", "repos/o/n/issues/1/comments", "-f", "body=x"}, action: "api-post", repo: "o/n"},
		{args: []string{"api", "-X", "DELETE", "/repos/o/n/git/refs/heads/x"}, action: "api-delete", repo: "o/n"},
		{args: []string{"api", "--method=GET", "user"}},
		{args: []string{"--help"}},
	}
	for _, c := range cases {
		got := classifyGH(c.args)
		name := strings.Join(c.args, " ")
		assert.Equal(t, c.action, got.action, name)
		assert.Equal(t, c.repo, got.repo, name)
		assert.Equal(t, c.num, got.number, name)
		assert.Equal(t, c.body, got.body, name)
	}
}

func TestNormRepo(t *testing.T) {
	for in, want := range map[string]string{
		"o/n":                           "o/n",
		"git@github.com:o/n.git\n":      "o/n",
		"https://github.com/o/n":        "o/n",
		"https://github.com/o/n.git":    "o/n",
		"https://github.com/o/n/pull/3": "o/n",
		"github.com/o/n":                "o/n",
		"nope":                          "",
	} {
		assert.Equal(t, want, normRepo(in), in)
	}
}

// fakeGH writes a gh script into dir that records its args in a file.
func fakeGH(t *testing.T, dir string) (gh, marker string) {
	t.Helper()
	marker = filepath.Join(dir, "ran")
	gh = filepath.Join(dir, "gh")
	require.NoError(t, os.WriteFile(gh, []byte("#!/bin/sh\necho \"$@\" > "+marker+"\nexit 3\n"), 0o700))
	return gh, marker
}

// TestRealGHSkipsSelf: a gh on PATH that links to this binary is skipped,
// so a gh -> jobshop symlink does not exec itself forever.
func TestRealGHSkipsSelf(t *testing.T) {
	self, err := os.Executable()
	require.NoError(t, err)
	a, b := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(self, filepath.Join(a, "gh")))
	gh, _ := fakeGH(t, b)
	t.Setenv("JOBSHOP_GH", "")
	t.Setenv("PATH", a+string(os.PathListSeparator)+b)
	got, err := realGH()
	require.NoError(t, err)
	assert.Equal(t, gh, got)

	t.Setenv("JOBSHOP_GH", "/x/gh")
	got, err = realGH()
	require.NoError(t, err)
	assert.Equal(t, "/x/gh", got)
}

// TestGHReadPassesThrough: a read runs gh with the args untouched and
// returns its exit code, even with no controller to ask.
func TestGHReadPassesThrough(t *testing.T) {
	dir := t.TempDir()
	gh, marker := fakeGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "set")
	g := &Globals{Socket: filepath.Join(dir, "none.sock")}
	var out, errb bytes.Buffer
	code := runGH(g, []string{"pr", "view", "1", "-R", "o/n", "--json", "title"}, nil, &out, &errb)
	assert.Equal(t, 3, code)
	b, err := os.ReadFile(marker)
	require.NoError(t, err)
	assert.Equal(t, "pr view 1 -R o/n --json title\n", string(b))
}

// TestGHWriteFailsClosed: a write with no controller to guard it does
// not run gh.
func TestGHWriteFailsClosed(t *testing.T) {
	dir := t.TempDir()
	gh, marker := fakeGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "set")
	g := &Globals{Socket: filepath.Join(dir, "none.sock")}
	var out, errb bytes.Buffer
	code := runGH(g, []string{"pr", "merge", "1", "-R", "o/n"}, nil, &out, &errb)
	assert.Equal(t, 1, code)
	assert.Contains(t, errb.String(), "guard merge")
	assert.NoFileExists(t, marker)
}

// fakeController serves guard, identity.token and activity.record on a
// unix socket: guard answers decision, identity.token gives "tok-<repo>".
type fakeController struct {
	mu       sync.Mutex
	decision string
	guards   []rpc.GuardParams
	records  []rpc.ActivityRecordParams
	// keys is the write key table: key -> result ("" pending).
	keys map[string]string
	// intake answers linear.intake; nil is method not found.
	intake *rpc.LinearIntakeResult
}

func (f *fakeController) serve(t *testing.T) string {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "s")
	ln, err := net.Listen("unix", sock)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			nc, err := ln.Accept()
			if err != nil {
				return
			}
			gate := rpc.NewGate(nc)
			conn := acp.NewConnection(func(_ context.Context, method string, raw json.RawMessage) (any, *acp.RequestError) {
				raw, _ = rpc.SplitMeta(raw)
				f.mu.Lock()
				defer f.mu.Unlock()
				switch method {
				case rpc.MethodGuard:
					var p rpc.GuardParams
					_ = json.Unmarshal(raw, &p)
					f.guards = append(f.guards, p)
					return rpc.GuardResult{Decision: f.decision, Reason: "because"}, nil
				case rpc.MethodIdentityToken:
					var p rpc.IdentityParams
					_ = json.Unmarshal(raw, &p)
					return rpc.IdentityTokenResult{Token: "tok-" + p.Target}, nil
				case rpc.MethodActivityRecord:
					var p rpc.ActivityRecordParams
					_ = json.Unmarshal(raw, &p)
					f.records = append(f.records, p)
					return struct{}{}, nil
				case rpc.MethodWriteKeyClaim:
					var p rpc.WriteKeyParams
					_ = json.Unmarshal(raw, &p)
					res, ok := f.keys[p.Key]
					if !ok {
						f.keys[p.Key] = ""
					}
					return store.WriteKey{Key: p.Key, Scoped: "js/" + p.Key, Op: p.Op, Result: res, Done: res != "", New: !ok}, nil
				case rpc.MethodWriteKeyFinish:
					var p rpc.WriteKeyParams
					_ = json.Unmarshal(raw, &p)
					if f.keys[p.Key] == "" {
						f.keys[p.Key] = p.Result
					}
					return store.WriteKey{Key: p.Key, Result: f.keys[p.Key], Done: true}, nil
				case rpc.MethodLinearIntake:
					if f.intake != nil {
						return *f.intake, nil
					}
				case rpc.MethodSecretScan:
					// As the controller, with the built-in rules.
					var p rpc.SecretScanParams
					_ = json.Unmarshal(raw, &p)
					res := rpc.SecretScanResult{Findings: []rpc.SecretFinding{}}
					for _, t := range p.Texts {
						for _, x := range redact.Find(t.Text) {
							res.Findings = append(res.Findings, rpc.SecretFinding{From: t.From, Rule: x.Rule, Line: x.Line})
						}
					}
					return res, nil
				}
				return nil, acp.NewMethodNotFound(method)
			}, nc, gate)
			conn.SetLogger(slog.New(slog.DiscardHandler))
			gate.Open()
		}
	}()
	return sock
}

// tokenGH writes a fake gh that saves its GH_TOKEN and exits 5.
func tokenGH(t *testing.T, dir string) (gh, marker string) {
	t.Helper()
	marker = filepath.Join(dir, "token")
	gh = filepath.Join(dir, "gh")
	require.NoError(t, os.WriteFile(gh, []byte("#!/bin/sh\nprintf %s \"$GH_TOKEN\" > "+marker+"\nexit 5\n"), 0o700))
	return gh, marker
}

func TestGHGuard(t *testing.T) {
	for _, tc := range []struct {
		decision string
		code     int
		ran      bool
	}{
		{domain.GuardAllow, 5, true},
		{domain.GuardDeny, 1, false},
		{domain.GuardAskHuman, 1, false},
	} {
		t.Run(tc.decision, func(t *testing.T) {
			dir := t.TempDir()
			gh, marker := tokenGH(t, dir)
			t.Setenv("JOBSHOP_GH", gh)
			t.Setenv("GH_TOKEN", "")
			f := &fakeController{decision: tc.decision}
			g := &Globals{Socket: f.serve(t)}
			var out, errb bytes.Buffer
			code := runGH(g, []string{"pr", "comment", "9", "-R", "o/n", "--body", "hello"}, nil, &out, &errb)
			assert.Equal(t, tc.code, code)
			if tc.ran {
				b, err := os.ReadFile(marker)
				require.NoError(t, err)
				assert.Equal(t, "tok-o/n", string(b), "identity token set")
			} else {
				assert.NoFileExists(t, marker)
				assert.Contains(t, errb.String(), "because")
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			require.Len(t, f.guards, 1)
			want := rpc.GuardParams{
				Action: "comment", Target: "o/n#9", Detail: "hello",
				Extra: json.RawMessage(`{"argv":["pr","comment","9","-R","o/n","--body","hello"],"command":"pr comment"}`),
			}
			assert.Equal(t, want, f.guards[0])
			require.Len(t, f.records, 1)
			assert.Equal(t, tc.decision, f.records[0].Decision)
			assert.Equal(t, tc.code, f.records[0].ExitCode)
			assert.Equal(t, "gh pr comment 9", f.records[0].Command)
		})
	}
}

// TestGHExplicitToken: a set GH_TOKEN is kept and no identity is asked.
func TestGHExplicitToken(t *testing.T) {
	dir := t.TempDir()
	gh, marker := tokenGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "mine")
	f := &fakeController{decision: domain.GuardAllow}
	g := &Globals{Socket: f.serve(t)}
	var out, errb bytes.Buffer
	assert.Equal(t, 5, runGH(g, []string{"pr", "view", "-R", "o/n"}, nil, &out, &errb))
	b, err := os.ReadFile(marker)
	require.NoError(t, err)
	assert.Equal(t, "mine", string(b))
}

// TestGHOnPathAsGH: jobshop installed as gh on PATH runs the real gh
// further down PATH, not itself.
func TestGHOnPathAsGH(t *testing.T) {
	self, err := os.Executable()
	require.NoError(t, err)
	a, b := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(self, filepath.Join(a, "gh")))
	_, marker := fakeGH(t, b)
	t.Setenv("JOBSHOP_GH", "")
	t.Setenv("GH_TOKEN", "set")
	t.Setenv("PATH", a+string(os.PathListSeparator)+b+string(os.PathListSeparator)+os.Getenv("PATH"))
	var out, errb bytes.Buffer
	code := runGH(&Globals{Socket: filepath.Join(a, "none")}, []string{"repo", "view", "-R", "o/n"}, nil, &out, &errb)
	assert.Equal(t, 3, code)
	assert.FileExists(t, marker)
}
