package control

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	host "github.com/elee1766/acphost/pkg/acphost"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// Per-session isolation (#540): with isolation.user.per_session each live
// session of a pool runs as its own host uid/gid, map_uid+slot and
// map_gid+slot. The slot is chosen when the session is created (lowest
// free among the pool's open sessions) and kept in the session's fields,
// so a controller restart reuses it and two sessions never share one. A
// full range is a capacity limit: scale up stops at it.
//
// The controller is unprivileged and cannot chown to a subuid, so it
// stays the owner of the session dir and worktree and grants the slot's
// uid, itself and the pool's first group through POSIX ACLs. The agent
// gets no supplementary groups, so another session's uid matches no
// entry and falls to "other", which has none.

// allocSlot returns the lowest slot in [0, n) not used by another open
// session of pool, or ok false when all are taken. Called under
// c.slotMu by newSession, so two creates cannot pick the same slot.
func (c *Controller) allocSlot(ctx context.Context, pool string, n int) (int, bool, error) {
	used, err := c.usedSlots(ctx, pool)
	if err != nil {
		return 0, false, err
	}
	for s := range n {
		if !used[s] {
			return s, true, nil
		}
	}
	return 0, false, nil
}

// usedSlots is the slots of pool's open sessions.
func (c *Controller) usedSlots(ctx context.Context, pool string) (map[int]bool, error) {
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Pool: pool})
	if err != nil {
		return nil, err
	}
	used := map[int]bool{}
	for _, t := range open {
		if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.Slot != nil {
			used[*f.Slot] = true
		}
	}
	return used, nil
}

// slotCapacity is how many more sessions pool may start, or -1 for no
// per-session limit.
func (c *Controller) slotCapacity(ctx context.Context, cfg *config.Config, pool string) int {
	u := cfg.IsolationFor(pool).User
	if u == nil || u.Slots() == 0 {
		return -1
	}
	used, err := c.usedSlots(ctx, pool)
	if err != nil {
		return 0
	}
	n := 0
	for s := range u.Slots() {
		if !used[s] {
			n++
		}
	}
	return n
}

// errSlotsFull is a pool whose per-session range is all in use.
var errSlotsFull = errors.New("per_session range full")

// sessionSlot picks a slot for a new session of pool, or nil when the
// pool does not use per-session ids.
func (c *Controller) sessionSlot(ctx context.Context, cfg *config.Config, pool string) (*int, error) {
	u := cfg.IsolationFor(pool).User
	if u == nil || u.Slots() == 0 {
		return nil, nil
	}
	s, ok, err := c.allocSlot(ctx, pool, u.Slots())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%w: pool %s uses all %d per-session ids", errSlotsFull, pool, u.Slots())
	}
	return &s, nil
}

// applySlot moves spec.User to the session's own ids. A pool that turned
// per_session on after the session was created, or a slot past a range
// that shrank on reload, is a permanent error: that session must end,
// not run as a shared uid.
func applySlot(spec *host.Spec, u config.IsolationUser, slot *int) error {
	if u.Slots() == 0 || spec.User == nil {
		return nil
	}
	if slot == nil {
		return host.Permanentf("isolation per_session: session has no id slot (created before per_session was set)")
	}
	if *slot >= u.Slots() {
		return host.Permanentf("isolation per_session: slot %d is outside range %d", *slot, u.Slots())
	}
	spec.User.MapUID = u.MapUID + *slot
	spec.User.MapGID = u.MapGID + *slot
	// acphost checks [MapUID, MapUID+MapCount) lies in one subuid entry
	// (and the same for gids). Counting to the end of the pool's block
	// catches a range that runs past the subuid range from any slot, not
	// only once the last slot is used. Permanent, so the pool is
	// quarantined as for #388.
	spec.User.MapCount = u.Slots() - *slot
	// Shared groups would let one session reach another's files.
	spec.User.Groups = nil
	return nil
}

// restrictSession moves spec.User to the session's slot and limits the
// session dir and worktree to the slot's uid, the controller and the
// pool's first group (#540).
func (c *Controller) restrictSession(spec *host.Spec, u config.IsolationUser, slot *int, dir string) error {
	if err := applySlot(spec, u, slot); err != nil {
		return err
	}
	gid, err := config.LookupGroup(u.Groups[0])
	if err != nil {
		return host.Permanentf("isolation user: %w", err)
	}
	if err := restrictDir(dir, spec.User.MapUID, gid, false); err != nil {
		return err
	}
	return restrictDir(spec.Dir, spec.User.MapUID, gid, true)
}

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

// sessionACL grants rwx to the owner (the controller), the session's uid
// and the pool's group, and nothing to the owning group or others.
func sessionACL(uid, gid int, file bool) []byte {
	rw := uint16(7)
	if file {
		rw = 6
	}
	es := []aclEntry{
		{aclUserObj, rw, aclUndefID},
		{aclUser, rw, uint32(uid)},
		{aclGroupObj, 0, aclUndefID},
		{aclGroup, rw, uint32(gid)},
		{aclMask, rw, aclUndefID},
		{aclOther, 0, aclUndefID},
	}
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

// restrictDir sets sessionACL on dir and, with recursive, everything
// under it; directories also get it as their default ACL so files the
// agent creates inherit it. Unix sockets directly in dir are covered too.
// A filesystem without ACLs is a permanent error: the isolation would
// silently not hold.
func restrictDir(dir string, uid, gid int, recursive bool) error {
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
			return host.Permanentf("isolation per_session: %s does not support POSIX ACLs", dir)
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
