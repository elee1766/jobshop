package control

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// skipNoACL skips where the temp filesystem has no POSIX ACLs, or the
// test runs in a user namespace that does not map the test's uids
// (EINVAL from setxattr).
func skipNoACL(t *testing.T, err error) {
	t.Helper()
	if p := new(host.PermanentError); errors.As(err, &p) || errors.Is(err, unix.EINVAL) {
		t.Skip(err)
	}
}

// TestCheckMountsModes: a 0700 dir owned by someone else is refused with
// its path and mode; a granted ACL lets the mapped uid through.
func TestCheckMountsModes(t *testing.T) {
	other := os.Getuid() + 100000
	gid := os.Getgid() + 100000
	root := t.TempDir()
	cache := filepath.Join(root, "go")
	require.NoError(t, os.Mkdir(cache, 0o755))
	require.NoError(t, os.Chmod(root, 0o700))

	sb := &host.Sandbox{RW: []host.Mount{{Src: cache}}}
	err := checkMounts(sb, cache, other, []int{gid})
	var p *host.PermanentError
	require.ErrorAs(t, err, &p)
	// root, or a 0700 dir above it on this host, is named with its mode.
	assert.Regexp(t, `cannot use /\S* \(mode 0700, owner \d+:\d+\)`, err.Error())

	skipNoACL(t, grantSearch(root, other))
	if reachable(filepath.Dir(root), unix.X_OK, other, nil) != nil {
		// Only the leaf checks below hold on this host.
		require.NoError(t, restrictDir(cache, other, gid, false))
		assert.NoError(t, allowed(root, unix.X_OK, other, nil))
		assert.NoError(t, allowed(cache, unix.R_OK|unix.W_OK|unix.X_OK, other, nil))
		t.Skip("a dir above the temp dir is closed to others here")
	}
	require.NoError(t, restrictDir(cache, other, gid, false))
	assert.NoError(t, checkMounts(sb, cache, other, []int{gid}))

	// A missing ro mount marked SkipMissing is not checked.
	sb.RO = []host.Mount{{Src: filepath.Join(root, "nope"), SkipMissing: true}}
	assert.NoError(t, checkMounts(sb, cache, other, []int{gid}))
}

// TestGrantSearchKeepsEntries adds x for a uid without dropping the
// owner's bits or an earlier named entry.
func TestGrantSearchKeepsEntries(t *testing.T) {
	d := t.TempDir()
	require.NoError(t, os.Chmod(d, 0o700))
	skipNoACL(t, grantSearch(d, 4242))
	require.NoError(t, grantSearch(d, 4343))
	require.NoError(t, grantSearch(d, 4343), "idempotent")
	for _, uid := range []int{4242, 4343} {
		assert.NoError(t, allowed(d, unix.X_OK, uid, nil))
		assert.Error(t, allowed(d, unix.R_OK, uid, nil), "search only")
	}
	assert.NoError(t, allowed(d, unix.R_OK|unix.W_OK|unix.X_OK, os.Getuid(), nil), "owner keeps rwx")
	assert.Error(t, allowed(d, unix.X_OK, 4444, nil), "others still out")
}

// TestPrepareMappedHome copies a login file into the home and grants
// the tree to the mapped uid; a missing file is skipped.
func TestPrepareMappedHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".jcode"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(os.Getenv("HOME"), ".jcode", "auth.json"), []byte(`{"k":1}`), 0o600))
	c := &Controller{log: slog.New(slog.DiscardHandler)}
	home := filepath.Join(t.TempDir(), "homes", "dev")
	uid, gid := os.Getuid()+100000, os.Getgid()
	err := c.prepareMappedHome(home, uid, []int{gid}, []string{"~/.jcode/auth.json", "~/.missing"})
	skipNoACL(t, err)
	require.NoError(t, err)
	b, err := os.ReadFile(filepath.Join(home, ".jcode", "auth.json"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"k":1}`, string(b))
	assert.NoError(t, allowed(filepath.Join(home, ".jcode", "auth.json"), unix.R_OK, uid, nil))
	assert.NoError(t, allowed(filepath.Join(home, ".jcode"), unix.R_OK|unix.W_OK|unix.X_OK, uid, nil))
}
