package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// markerGH is a fake gh: `api` prints the URL of a stored post containing
// the marker (it greps the posts file for the marker in the -jq arg), any
// other call appends its args to the posts file and prints a URL.
func markerGH(t *testing.T, dir string) (gh, posts string) {
	t.Helper()
	posts = filepath.Join(dir, "posts")
	gh = filepath.Join(dir, "gh")
	script := `#!/bin/sh
posts=` + posts + `
if [ "$1" = api ]; then
  for a; do m=$(printf %s "$a" | sed -n 's/.*contains("\(<!-- jobshop-key: [^>]*-->\)").*/\1/p'); [ -n "$m" ] && break; done
  if [ -f "$posts" ] && grep -qF "$m" "$posts"; then echo https://github.com/o/n/issues/9#issuecomment-1; fi
  exit 0
fi
printf '%s\n' "$*" >> "$posts"
echo https://github.com/o/n/issues/9#issuecomment-1
`
	require.NoError(t, os.WriteFile(gh, []byte(script), 0o700))
	return gh, posts
}

func TestCutGHKey(t *testing.T) {
	k, rest := cutGHKey([]string{"pr", "comment", "1", "--jobshop-key", "k1", "-b", "x"})
	assert.Equal(t, "k1", k)
	assert.Equal(t, []string{"pr", "comment", "1", "-b", "x"}, rest)
	k, rest = cutGHKey([]string{"--jobshop-key=k2", "issue", "comment", "2"})
	assert.Equal(t, "k2", k)
	assert.Equal(t, []string{"issue", "comment", "2"}, rest)
}

// TestGHKeyedCommentAfterLocalKeyGone: the first keyed comment carries the
// hidden marker; with the local key gone (#415, pruned) a retry finds the
// marker on the thread and does not post again.
func TestGHKeyedCommentAfterLocalKeyGone(t *testing.T) {
	dir := t.TempDir()
	gh, posts := markerGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "set")
	f := &fakeController{decision: rpc.GuardAllow, keys: map[string]string{}}
	g := &Globals{Socket: f.serve(t)}
	args := []string{"issue", "comment", "9", "-R", "o/n", "--body", "hello", "--jobshop-key", "c1"}

	var out, errb bytes.Buffer
	require.Equal(t, 0, runGH(g, args, nil, &out, &errb), errb.String())
	b, err := os.ReadFile(posts)
	require.NoError(t, err)
	assert.Contains(t, string(b), "hello\n\n<!-- jobshop-key: js/c1 -->")
	assert.NotContains(t, string(b), "jobshop-key c1", "flag is not passed to gh")

	f.mu.Lock()
	clear(f.keys) // the 48h sweep removed the local row
	f.mu.Unlock()
	out.Reset()
	require.Equal(t, 0, runGH(g, args, nil, &out, &errb), errb.String())
	b, err = os.ReadFile(posts)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(b), "jobshop-key"), "posted once")
	assert.Contains(t, out.String(), "issuecomment-1")
}

func TestGHKeyOnlyForWrites(t *testing.T) {
	dir := t.TempDir()
	gh, posts := markerGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "set")
	var out, errb bytes.Buffer
	code := runGH(&Globals{Socket: filepath.Join(dir, "none")}, []string{"pr", "view", "1", "--jobshop-key", "k"}, nil, &out, &errb)
	assert.Equal(t, 1, code)
	assert.NoFileExists(t, posts)
}
