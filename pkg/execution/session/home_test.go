package session

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/elee1766/acphost/pkg/xerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/execution/sandbox"
)

// skipNoACL skips where the temp filesystem has no POSIX ACLs, or the
// test runs in a user namespace that does not map the test's uids
// (EINVAL from setxattr).
func skipNoACL(t *testing.T, err error) {
	t.Helper()
	if p := new(xerr.PermanentError); errors.As(err, &p) || errors.Is(err, unix.EINVAL) {
		t.Skip(err)
	}
}

// TestPrepareHome copies a login file into the home and grants the tree
// to the mapped uid; a missing file is skipped.
func TestPrepareHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".jcode"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(os.Getenv("HOME"), ".jcode", "auth.json"), []byte(`{"k":1}`), 0o600))
	h := Homes{Root: t.TempDir(), Log: slog.New(slog.DiscardHandler)}
	home := h.Dir("dev", nil, nil)
	assert.Equal(t, filepath.Join(h.Root, "homes", "dev"), home)
	uid, gid := os.Getuid()+100000, os.Getgid()
	err := h.Prepare(home, uid, []int{gid}, []string{"~/.jcode/auth.json", "~/.missing"})
	skipNoACL(t, err)
	require.NoError(t, err)
	b, err := os.ReadFile(filepath.Join(home, ".jcode", "auth.json"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"k":1}`, string(b))
	assert.NoError(t, sandbox.Allowed(filepath.Join(home, ".jcode", "auth.json"), unix.R_OK, uid, nil))
	assert.NoError(t, sandbox.Allowed(filepath.Join(home, ".jcode"), unix.R_OK|unix.W_OK|unix.X_OK, uid, nil))
}

// TestHomeDirSlot: per_session ids give each slot its own home, a pool
// without them one shared home.
func TestHomeDirSlot(t *testing.T) {
	h := Homes{Root: "/r"}
	per := &config.IsolationUser{PerSession: &config.PerSession{Range: 2}}
	assert.Equal(t, "/r/homes/dev-1", h.Dir("dev", per, new(1)))
	assert.Equal(t, "/r/homes/dev", h.Dir("dev", per, nil))
	assert.Equal(t, "/r/homes/dev", h.Dir("dev", &config.IsolationUser{}, new(1)))
}

// TestMappedUser: slots offset the ids and drop supplementary groups;
// no user means not mapped.
func TestMappedUser(t *testing.T) {
	_, _, ok, err := MappedUser(nil, nil)
	require.NoError(t, err)
	assert.False(t, ok)

	u := &config.IsolationUser{MapUID: 100000, MapGID: 200000, Groups: []string{"nosuchgroup-js153"}, PerSession: &config.PerSession{Range: 4}}
	uid, gids, ok, err := MappedUser(u, new(3))
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 100003, uid)
	assert.Equal(t, []int{200003}, gids)

	u.PerSession = nil
	_, _, _, err = MappedUser(u, nil)
	require.Error(t, err, "an unknown shared group is a permanent error")
}

// TestRemoveSlot deletes the home unless a host still runs in the
// session dir.
func TestRemoveSlot(t *testing.T) {
	h := Homes{Root: t.TempDir(), Log: slog.New(slog.DiscardHandler)}
	home := filepath.Join(h.Root, "homes", "dev-0")
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".cache"), 0o700))
	h.RemoveSlot("s1", filepath.Join(h.Root, "sessions", "s1"), home)
	assert.NoDirExists(t, home)
}
