// Package sandbox is process isolation and resource mechanics (JOB-167):
// the bubblewrap sandbox spec built from config, the mapped user and the
// POSIX ACLs and mount checks it needs, delegated cgroups and pool
// limits, and one isolated command run with cancellation and output
// streams. Sessions and command hooks both use it. It takes a config and
// a process spec, and knows nothing of session hosts, adoption, ACP,
// leases or scaling.
package sandbox

import (
	"encoding/binary"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/elee1766/acphost/pkg/userns"
	"github.com/elee1766/acphost/pkg/xerr"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
)

// The controller is unprivileged and cannot chown to a subuid, so for a
// pool with isolation.user it stays the owner of the session dir and
// worktree and grants the mapped uid access through POSIX ACLs (#540,
// JOB-19). The agent gets no supplementary groups with per_session, so
// another session's uid matches no entry and falls to "other", which has
// none.

// POSIX ACL xattr format (linux/posix_acl_xattr.h): a 4-byte version,
// then 8-byte entries of tag, perm and id, sorted by tag then id.
const (
	aclVersion   = 2
	aclUserObj   = 0x01
	aclUser      = 0x02
	aclGroupObj  = 0x04
	aclGroup     = 0x08
	aclMask      = 0x10
	aclOther     = 0x20
	aclUndefID   = 0xffffffff
	aclAccessKey = "system.posix_acl_access"
	aclDefault   = "system.posix_acl_default"
)

type aclEntry struct {
	tag  uint16
	perm uint16
	id   uint32
}

// encodeACL sorts es by tag then id and encodes them as an ACL xattr.
func encodeACL(es []aclEntry) []byte {
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
	return b
}

// sessionACL grants rwx to the owner (the controller), the session's uid
// and the pool's group, and nothing to the owning group or others.
func sessionACL(uid, gid int, file bool) []byte {
	rw := uint16(7)
	if file {
		rw = 6
	}
	return encodeACL([]aclEntry{
		{aclUserObj, rw, aclUndefID},
		{aclUser, rw, uint32(uid)},
		{aclGroupObj, 0, aclUndefID},
		{aclGroup, rw, uint32(gid)},
		{aclMask, rw, aclUndefID},
		{aclOther, 0, aclUndefID},
	})
}

// RestrictDir sets the session ACL on dir and, with recursive, everything
// under it; directories also get it as their default ACL so files the
// agent creates inherit it. Unix sockets directly in dir are covered too.
// A filesystem without ACLs is a permanent error: the isolation would
// silently not hold.
func RestrictDir(dir string, uid, gid int, recursive bool) error {
	dirACL, fileACL := sessionACL(uid, gid, false), sessionACL(uid, gid, true)
	fix := func(p string, d fs.DirEntry) error {
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if !d.IsDir() {
			return unix.Lsetxattr(p, aclAccessKey, fileACL, 0)
		}
		if err := unix.Lsetxattr(p, aclAccessKey, dirACL, 0); err != nil {
			return err
		}
		return unix.Lsetxattr(p, aclDefault, dirACL, 0)
	}
	wrap := func(err error) error {
		if errors.Is(err, unix.EOPNOTSUPP) {
			return xerr.Permanentf("isolation per_session: %s does not support POSIX ACLs", dir)
		}
		return err
	}
	if recursive {
		return wrap(filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return fix(p, d)
		}))
	}
	st, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if err := fix(dir, fs.FileInfoToDirEntry(st)); err != nil {
		return wrap(err)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if e.Type()&fs.ModeSocket != 0 {
			if err := fix(filepath.Join(dir, e.Name()), e); err != nil {
				return wrap(err)
			}
		}
	}
	return nil
}

// GrantSearch adds search (x) for uid to dir's ACL, keeping its other
// entries. The jobshop root and its sessions and homes dirs are 0700;
// the mapped uid needs to pass through them, not list them.
func GrantSearch(dir string, uid int) error {
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
	if err := unix.Setxattr(dir, aclAccessKey, encodeACL(es), 0); err != nil {
		if errors.Is(err, unix.EOPNOTSUPP) {
			return xerr.Permanentf("isolation: %s does not support POSIX ACLs", dir)
		}
		return err
	}
	return nil
}

// ShareDir gives gid group ownership of dir, group rw (rwx and setgid on
// directories, so new files keep the group), and with recursive the same
// for everything under it. Unix sockets in dir are made group writable
// too. Used for pools with isolation.user (#388).
func ShareDir(dir string, gid int, recursive bool) error {
	fix := func(p string, d fs.DirEntry) error {
		// Files the mapped user made (a respawn after it wrote the
		// worktree) are not ours to chown; the setgid dir gave them the
		// group already.
		var st syscall.Stat_t
		if err := syscall.Lstat(p, &st); err != nil {
			return err
		}
		if int(st.Uid) != os.Geteuid() {
			return nil
		}
		if err := os.Lchown(p, -1, gid); err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		mode := info.Mode()
		perm := mode.Perm() | 0o060
		if mode.IsDir() {
			perm |= 0o010
			return os.Chmod(p, perm|fs.ModeSetgid)
		}
		return os.Chmod(p, perm)
	}
	if recursive {
		return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return fix(p, d)
		})
	}
	st, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if err := fix(dir, fs.FileInfoToDirEntry(st)); err != nil {
		return err
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if e.Type()&fs.ModeSocket != 0 {
			if err := fix(filepath.Join(dir, e.Name()), e); err != nil {
				return err
			}
		}
	}
	return nil
}

// ApplySlot moves user to the session's own ids. A pool that turned
// per_session on after the session was created, or a slot past a range
// that shrank on reload, is a permanent error: that session must end,
// not run as a shared uid.
func ApplySlot(user *userns.User, u config.IsolationUser, slot *int) error {
	if u.Slots() == 0 || user == nil {
		return nil
	}
	if slot == nil {
		return xerr.Permanentf("isolation per_session: session has no id slot (created before per_session was set)")
	}
	if *slot >= u.Slots() {
		return xerr.Permanentf("isolation per_session: slot %d is outside range %d", *slot, u.Slots())
	}
	user.MapUID = u.MapUID + *slot
	user.MapGID = u.MapGID + *slot
	// acphost checks [MapUID, MapUID+MapCount) lies in one subuid entry
	// (and the same for gids). Counting to the end of the pool's block
	// catches a range that runs past the subuid range from any slot, not
	// only once the last slot is used. Permanent, so the pool is
	// quarantined as for #388.
	user.MapCount = u.Slots() - *slot
	// Shared groups would let one session reach another's files.
	user.Groups = nil
	return nil
}

// RestrictSession moves user to the session's slot and limits the
// session dir and the worktree cwd to the slot's uid, the controller and the
// pool's first group (#540).
func RestrictSession(user *userns.User, cwd string, u config.IsolationUser, slot *int, dir string) error {
	if err := ApplySlot(user, u, slot); err != nil {
		return err
	}
	gid, err := config.LookupGroup(u.Groups[0])
	if err != nil {
		return xerr.Permanentf("isolation user: %w", err)
	}
	if err := RestrictDir(dir, user.MapUID, gid, false); err != nil {
		return err
	}
	return RestrictDir(cwd, user.MapUID, gid, true)
}
