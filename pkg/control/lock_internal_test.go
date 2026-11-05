package control

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// TestLockReleasedWhileFdShared: unlock frees the controller lock even when
// another descriptor for the same open file is still open, as in a child
// forked between lock and unlock that has not exec'd yet (#601).
func TestLockReleasedWhileFdShared(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "controller.lock")
	unlock, err := lock(path)
	require.NoError(t, err)

	// Dup the lock's fd, standing in for the forked child's copy.
	ents, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	fd := -1
	for _, e := range ents {
		if l, err := os.Readlink("/proc/self/fd/" + e.Name()); err == nil && l == path {
			fd, err = strconv.Atoi(e.Name())
			require.NoError(t, err)
		}
	}
	require.GreaterOrEqual(t, fd, 0, "lock fd not found")
	shared, err := unix.Dup(fd)
	require.NoError(t, err)
	defer func() { _ = unix.Close(shared) }()

	unlock()
	again, err := lock(path)
	require.NoError(t, err, "lock still held through the shared descriptor")
	again()
}
