package sandbox

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/elee1766/acphost/pkg/userns"
	"github.com/elee1766/acphost/pkg/xerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
)

func TestApplySlot(t *testing.T) {
	u := config.IsolationUser{UID: 1000, MapUID: 100000, MapGID: 200000, Groups: []string{"100"}, PerSession: &config.PerSession{Range: 4}}
	user := &userns.User{UID: 1000, GID: 1000, MapUID: 100000, MapGID: 200000, Groups: []int{100}}
	slot := 3
	require.NoError(t, ApplySlot(user, u, &slot))
	assert.Equal(t, 100003, user.MapUID)
	assert.Equal(t, 200003, user.MapGID)
	assert.Empty(t, user.Groups, "no shared groups across sessions")

	var p *xerr.PermanentError
	assert.ErrorAs(t, ApplySlot(user, u, nil), &p, "no slot is permanent")
	big := 4
	assert.ErrorAs(t, ApplySlot(user, u, &big), &p, "slot outside a shrunk range")
}

func TestSessionACL(t *testing.T) {
	b := sessionACL(100003, 100, false)
	require.Len(t, b, 4+6*8)
	assert.Equal(t, uint32(aclVersion), binary.LittleEndian.Uint32(b))
	var tags []uint16
	for i := 4; i < len(b); i += 8 {
		tags = append(tags, binary.LittleEndian.Uint16(b[i:]))
	}
	assert.Equal(t, []uint16{aclUserObj, aclUser, aclGroupObj, aclGroup, aclMask, aclOther}, tags)
	assert.Equal(t, uint32(100003), binary.LittleEndian.Uint32(b[4+8+4:]))
	assert.Equal(t, uint16(0), binary.LittleEndian.Uint16(b[4+5*8+2:]), "other gets nothing")
}

// TestRestrictDir writes the ACLs on a real directory tree and checks
// the kernel accepts them; skipped where the filesystem has no ACLs.
func TestRestrictDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a", "f"), []byte("x"), 0o644))
	err := RestrictDir(dir, os.Getuid()+1, os.Getgid(), true)
	if p := new(xerr.PermanentError); errors.As(err, &p) {
		t.Skip(err)
	}
	// JOB-18: inside a user namespace that maps only the caller (agent
	// sandboxes), the kernel rejects an ACL naming an unmapped uid with
	// EINVAL. The controller runs unnamespaced, so only the test skips.
	if errors.Is(err, unix.EINVAL) {
		t.Skipf("uid %d is not mapped in this user namespace: %v", os.Getuid()+1, err)
	}
	require.NoError(t, err)
	buf := make([]byte, 256)
	n, err := unix.Lgetxattr(filepath.Join(dir, "a", "f"), aclAccessKey, buf)
	require.NoError(t, err)
	assert.Equal(t, sessionACL(os.Getuid()+1, os.Getgid(), true), buf[:n])
	n, err = unix.Lgetxattr(filepath.Join(dir, "a"), aclDefault, buf)
	require.NoError(t, err)
	assert.Equal(t, sessionACL(os.Getuid()+1, os.Getgid(), false), buf[:n])
	st, err := os.Stat(filepath.Join(dir, "a", "f"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), st.Mode().Perm()&0o607, "group shows the mask, other none")
}
