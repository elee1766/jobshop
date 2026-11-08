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

	"github.com/elee1766/jobshop/pkg/control/sessionrt"
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
	assert.NoError(t, sessionrt.Allowed(filepath.Join(home, ".jcode", "auth.json"), unix.R_OK, uid, nil))
	assert.NoError(t, sessionrt.Allowed(filepath.Join(home, ".jcode"), unix.R_OK|unix.W_OK|unix.X_OK, uid, nil))
}
