package control

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	host "github.com/elee1766/acphost/pkg/acphost"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// A pool with isolation.user runs its agent as a mapped host uid that is
// not the controller user, and the controller's home is 0700, so the
// sandbox's "~/" rw paths (build caches, agent state) cannot be bound
// from it (JOB-19). Such a session gets a home of its own under
// <root>/homes, granted to the mapped uid by ACL: per pool, or per slot
// with per_session ids. "~/" rw paths bind from there to the same path
// inside the sandbox, and the agent's credentials files are copied in.

// mappedUser is the host uid and gids a session's agent runs as, or ok
// false when its pool has no isolation.user.
func (c *Controller) mappedUser(pool string, slot *int) (uid int, gids []int, ok bool, err error) {
	u := c.config().IsolationFor(pool).User
	if u == nil {
		return 0, nil, false, nil
	}
	uid, gid := u.MapUID, u.MapGID
	if u.Slots() > 0 && slot != nil {
		uid, gid = uid+*slot, gid+*slot
	}
	gids = []int{gid}
	if u.Slots() > 0 {
		// Per-session agents get no supplementary groups (applySlot).
		return uid, gids, true, nil
	}
	for _, g := range u.Groups {
		n, err := config.LookupGroup(g)
		if err != nil {
			return 0, nil, false, host.Permanentf("isolation user: %w", err)
		}
		gids = append(gids, n)
	}
	return uid, gids, true, nil
}

// mappedHomeDir is the home for pool's mapped user: one per pool, or one
// per slot with per_session ids so sessions do not share files.
func (c *Controller) mappedHomeDir(pool string, slot *int) string {
	name := pool
	if u := c.config().IsolationFor(pool).User; u != nil && u.Slots() > 0 && slot != nil {
		name += "-" + strconv.Itoa(*slot)
	}
	return filepath.Join(c.root.Dir, "homes", name)
}

// prepareMappedHome creates home, grants it to uid (and the first gid, so
// the controller's group can still clean up) and copies the agent's
// credentials files into it.
func (c *Controller) prepareMappedHome(home string, uid int, gids []int, creds []string) error {
	if err := os.MkdirAll(home, 0o700); err != nil {
		return host.Permanentf("isolation home %s: %w", home, err)
	}
	if err := restrictDir(home, uid, gids[0], false); err != nil {
		return err
	}
	for _, p := range creds {
		rel, ok := strings.CutPrefix(p, "~/")
		if !ok {
			continue
		}
		dst := filepath.Join(home, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o770); err != nil {
			return host.Permanentf("isolation credentials %s: %w", p, err)
		}
		if err := copyCredential(expandHome(p), dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				c.log.Warn("isolation: agent credentials file missing", "path", expandHome(p))
				continue
			}
			return host.Permanentf("isolation credentials %s: %w", p, err)
		}
	}
	// Directories made above inherit the home's default ACL; files copied
	// in get it too. Redo the tree so the copy is readable by uid.
	return restrictDir(home, uid, gids[0], true)
}

// copyCredential copies src to dst through a temp file, so the agent never
// reads a half-written login. 0660: with a default ACL the group bits
// are the mask, so 0600 would shut out the mapped uid.
func copyCredential(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".cred-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o660); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

// sessionMappedHome returns the mapped user's home for t, prepared, or ""
// when the pool has no isolation.user.
func (c *Controller) sessionMappedHome(t store.Ticket, f store.SessionFields) (string, error) {
	uid, gids, ok, err := c.mappedUser(t.Pool, f.Slot)
	if err != nil || !ok {
		return "", err
	}
	cfg := c.config()
	agent := cfg.Agents[cfg.Pools[t.Pool].Agent]
	home := c.mappedHomeDir(t.Pool, f.Slot)
	if err := c.prepareMappedHome(home, uid, gids, agent.Credentials); err != nil {
		return "", err
	}
	return home, nil
}

// removeSlotHome deletes a closed per_session session's home (JOB-28), so
// the next session given the same slot starts empty and cannot read what
// this one left. Per-pool homes are shared by design and kept. The
// controller can delete agent files through the pool group's default
// ACL entry; a failure is logged, and the next spawn in the slot still
// re-grants the dir.
func (c *Controller) removeSlotHome(t store.Ticket) {
	u := c.config().IsolationFor(t.Pool).User
	if t.Kind != store.KindSession || t.Status != store.StatusClosed || u == nil || u.Slots() == 0 {
		return
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil || f.Slot == nil {
		return
	}
	if host.Alive(c.root.SessionDir(t.ID)) {
		return
	}
	home := c.mappedHomeDir(t.Pool, f.Slot)
	if err := os.RemoveAll(home); err != nil {
		c.log.Warn("remove per-session home", "session", t.ID, "home", home, "err", err)
	}
}

// checkMounts fails with the first sandbox mount source (or the cwd) the
// mapped uid cannot reach, naming the path and its mode. Without it bwrap
// fails at spawn ("Can't find source path ... Permission denied") and the
// pool crash-loops into quarantine.
func checkMounts(s *host.Sandbox, cwd string, uid int, gids []int) error {
	type need struct {
		path string
		mode uint32
	}
	var ns []need
	if s != nil {
		for _, m := range s.RO {
			if _, err := os.Stat(m.Src); m.SkipMissing && err != nil {
				continue
			}
			ns = append(ns, need{m.Src, unix.R_OK})
		}
		for _, m := range s.RW {
			ns = append(ns, need{m.Src, unix.R_OK | unix.W_OK})
		}
	}
	ns = append(ns, need{cwd, unix.R_OK | unix.W_OK})
	for _, n := range ns {
		if err := reachable(n.path, n.mode, uid, gids); err != nil {
			return host.Permanentf("isolation: mapped uid %d cannot use %w", uid, err)
		}
	}
	return nil
}

// reachable checks uid can search every directory above p and has want
// on p itself.
func reachable(p string, want uint32, uid int, gids []int) error {
	p = filepath.Clean(p)
	var dirs []string
	for d := filepath.Dir(p); ; d = filepath.Dir(d) {
		dirs = append(dirs, d)
		if d == filepath.Dir(d) {
			break
		}
	}
	slices.Reverse(dirs)
	for _, d := range dirs {
		if err := allowed(d, unix.X_OK, uid, gids); err != nil {
			return err
		}
	}
	st, err := os.Stat(p)
	if err != nil {
		return err
	}
	if st.IsDir() {
		want |= unix.X_OK
	}
	return allowed(p, want, uid, gids)
}

// allowed evaluates p's mode bits and POSIX ACL (access check order of
// acl(5)) for uid and gids.
func allowed(p string, want uint32, uid int, gids []int) error {
	var st unix.Stat_t
	if err := unix.Stat(p, &st); err != nil {
		return err
	}
	deny := fmt.Errorf("%s (mode %04o, owner %d:%d)", p, st.Mode&0o7777, st.Uid, st.Gid)
	ok := func(perm uint32) error {
		if perm&want == want {
			return nil
		}
		return deny
	}
	if int(st.Uid) == uid {
		return ok(st.Mode >> 6 & 7)
	}
	inGroup := func(g uint32) bool { return slices.Contains(gids, int(g)) }
	buf := make([]byte, 1024)
	n, err := unix.Getxattr(p, aclAccessKey, buf)
	if err != nil || n < 4 {
		// No ACL: plain mode bits.
		if inGroup(st.Gid) {
			return ok(st.Mode >> 3 & 7)
		}
		return ok(st.Mode & 7)
	}
	mask, other := uint32(7), uint32(0)
	var user *uint32
	var groups []uint32
	matched := false
	for i := 4; i+8 <= n; i += 8 {
		tag := binary.LittleEndian.Uint16(buf[i:])
		perm := uint32(binary.LittleEndian.Uint16(buf[i+2:]))
		id := binary.LittleEndian.Uint32(buf[i+4:])
		switch tag {
		case aclUser:
			if int(id) == uid {
				user = &perm
			}
		case aclGroupObj:
			if inGroup(st.Gid) {
				groups, matched = append(groups, perm), true
			}
		case aclGroup:
			if inGroup(id) {
				groups, matched = append(groups, perm), true
			}
		case aclMask:
			mask = perm
		case aclOther:
			other = perm
		}
	}
	if user != nil {
		return ok(*user & mask)
	}
	if matched {
		for _, g := range groups {
			if g&mask&want == want {
				return nil
			}
		}
		return deny
	}
	return ok(other)
}

// grantSearch adds search (x) for uid to dir's ACL, keeping its other
// entries. The jobshop root and its sessions and homes dirs are 0700;
// the mapped uid needs to pass through them, not list them.
func grantSearch(dir string, uid int) error {
	var st unix.Stat_t
	if err := unix.Stat(dir, &st); err != nil {
		return err
	}
	es := []aclEntry{}
	buf := make([]byte, 4096)
	if n, err := unix.Getxattr(dir, aclAccessKey, buf); err == nil && n >= 4 {
		for i := 4; i+8 <= n; i += 8 {
			es = append(es, aclEntry{
				tag:  binary.LittleEndian.Uint16(buf[i:]),
				perm: binary.LittleEndian.Uint16(buf[i+2:]),
				id:   binary.LittleEndian.Uint32(buf[i+4:]),
			})
		}
	} else {
		m := uint16(st.Mode)
		es = []aclEntry{
			{aclUserObj, m >> 6 & 7, aclUndefID},
			{aclGroupObj, m >> 3 & 7, aclUndefID},
			{aclOther, m & 7, aclUndefID},
		}
	}
	found := false
	for i, e := range es {
		if e.tag == aclUser && int(e.id) == uid {
			if e.perm&1 != 0 {
				return nil
			}
			es[i].perm |= 1
			found = true
		}
	}
	if !found {
		es = append(es, aclEntry{aclUser, 1, uint32(uid)})
	}
	// The mask bounds named entries and the owning group: their union.
	var mask uint16
	es = slices.DeleteFunc(es, func(e aclEntry) bool { return e.tag == aclMask })
	for _, e := range es {
		if e.tag == aclUser || e.tag == aclGroup || e.tag == aclGroupObj {
			mask |= e.perm
		}
	}
	es = append(es, aclEntry{aclMask, mask, aclUndefID})
	slices.SortFunc(es, func(a, b aclEntry) int {
		if a.tag != b.tag {
			return int(a.tag) - int(b.tag)
		}
		return int(a.id) - int(b.id)
	})
	b := binary.LittleEndian.AppendUint32(nil, aclVersion)
	for _, e := range es {
		b = binary.LittleEndian.AppendUint16(b, e.tag)
		b = binary.LittleEndian.AppendUint16(b, e.perm)
		b = binary.LittleEndian.AppendUint32(b, e.id)
	}
	if err := unix.Setxattr(dir, aclAccessKey, b, 0); err != nil {
		if errors.Is(err, unix.EOPNOTSUPP) {
			return host.Permanentf("isolation: %s does not support POSIX ACLs", dir)
		}
		return err
	}
	return nil
}

// checkMappedAccess lets the session's mapped uid pass through the root
// to its session dir and home, then checks every mount source and the
// cwd are usable by it, so a wrong mode is one clear permanent error.
func (c *Controller) checkMappedAccess(spec host.Spec, slot *int, pool string) error {
	uid, gids, ok, err := c.mappedUser(pool, slot)
	if err != nil || !ok {
		return err
	}
	for _, d := range []string{c.root.Dir, filepath.Join(c.root.Dir, "sessions"), filepath.Join(c.root.Dir, "homes")} {
		if err := grantSearch(d, uid); err != nil {
			return host.Permanentf("isolation: let uid %d through %s: %w", uid, d, err)
		}
	}
	return checkMounts(spec.Sandbox, spec.Dir, uid, gids)
}
