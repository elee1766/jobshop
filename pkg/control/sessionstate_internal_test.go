package control

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/proc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestSessionListState (JOB-32): a closed session is listed from the
// state saved on its ticket when it ended, and an orphaned host appears
// once a supervisor pass has scanned the session directories.
func TestSessionListState(t *testing.T) {
	t.Parallel()
	c, _, _ := subController(t)
	ctx := t.Context()
	c.ctx = ctx
	list := func() map[string]rpc.SessionInfo {
		v, err := c.sessionList(ctx, scope{admin: true}, rpc.SessionListParams{All: true})
		require.NoError(t, err)
		out := map[string]rpc.SessionInfo{}
		for _, i := range v.([]rpc.SessionInfo) {
			out[i.ID] = i
		}
		return out
	}

	// A closed session whose host wrote an exit file and a journal.
	sess := mkSession(t, c)
	dir := c.root.SessionDir(sess.ID)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.ExitFile), []byte(`{"reason":"agent exited"}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.JournalFile), []byte(`{"seq":7,"t":1762387200000,"dir":"out"}`+"\n"), 0o600))
	_, err := store.Close(ctx, c.store, "controller", sess.ID, nil, "done")
	require.NoError(t, err)
	c.endSession(ctx, sess.ID, "test")

	// The directory goes; the listing does not read it.
	require.NoError(t, os.RemoveAll(dir))
	got := list()[sess.ID]
	assert.Equal(t, "agent exited", got.Exit)
	assert.Equal(t, int64(7), got.Seq)
	assert.False(t, got.Alive)
	assert.False(t, got.Orphan)

	// An orphan: a locked session directory no ticket names, made after
	// the first pass, so it is listed only after the next one.
	c.scanSessions(t.Context())
	odir := c.root.SessionDir("js-orphan")
	require.NoError(t, os.MkdirAll(odir, 0o700))
	lf, err := os.Create(filepath.Join(odir, proc.LockFile))
	require.NoError(t, err)
	t.Cleanup(func() { _ = lf.Close() })
	require.NoError(t, syscall.Flock(int(lf.Fd()), syscall.LOCK_EX))
	require.NoError(t, os.WriteFile(filepath.Join(odir, proc.SupervisorPidFile), strconv.AppendInt(nil, int64(os.Getpid()), 10), 0o600))
	assert.NotContains(t, list(), "js-orphan", "no pass has seen it")

	c.scanSessions(t.Context())
	o, ok := list()["js-orphan"]
	require.True(t, ok, "the pass found the orphan")
	assert.True(t, o.Orphan)
	assert.True(t, o.Alive)
	assert.Equal(t, os.Getpid(), o.HostPid)

	// The host stops: the next request drops it before any pass. Closing
	// the file does not drop the flock at once: a process another parallel
	// test forks in that moment holds a copy of the descriptor until it
	// execs, and the lock lives as long as any copy (JOB-71). Wait for the
	// lock to be gone, as a real host's exit would have made it, so the
	// listing cannot see it flip between its two liveness checks.
	require.NoError(t, lf.Close())
	require.Eventually(t, func() bool { return !hostAlive(odir) }, time.Minute, time.Millisecond, "flock released")
	assert.NotContains(t, list(), "js-orphan")
}

// TestBackfillEnded (JOB-32): sessions closed before Ended existed get it
// from their directory on a supervisor pass.
func TestBackfillEnded(t *testing.T) {
	t.Parallel()
	c, _, _ := subController(t)
	ctx := t.Context()
	c.ctx = ctx
	sess := mkSession(t, c)
	dir := c.root.SessionDir(sess.ID)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.ExitFile), []byte(`{"reason":"old exit"}`), 0o600))
	_, err := store.Close(ctx, c.store, "controller", sess.ID, nil, "done")
	require.NoError(t, err)

	c.scanSessions(t.Context())
	got, err := c.store.Get(ctx, sess.ID)
	require.NoError(t, err)
	f, err := store.DecodeFields[store.SessionFields](got)
	require.NoError(t, err)
	require.NotNil(t, f.Ended)
	assert.Equal(t, "old exit", f.Ended.Exit)
	assert.True(t, c.endedFilled)
}
