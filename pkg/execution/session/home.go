package session

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/execution/sandbox"
)

// A pool with isolation.user runs its agent as a mapped host uid that is
// not the controller user, and the controller's home is 0700, so the
// sandbox's "~/" rw paths (build caches, agent state) cannot be bound
// from it (JOB-19). Such a session gets a home of its own under
// <root>/homes, granted to the mapped uid by ACL: per pool, or per slot
// with per_session ids. "~/" rw paths bind from there to the same path
// inside the sandbox, and the agent's credentials files are copied in.
//
// This file holds the filesystem side. Which slot a session has, and
// whether its ticket is closed, are durable state the controller keeps:
// it passes the answers in.

// MappedUser is the host uid and gids a session's agent runs as under u,
// or ok false when u is nil (the pool has no isolation.user).
func MappedUser(u *config.IsolationUser, slot *int) (uid int, gids []int, ok bool, err error) {
	if u == nil {
		return 0, nil, false, nil
	}
	uid, gid := u.MapUID, u.MapGID
	if u.Slots() > 0 && slot != nil {
		uid, gid = uid+*slot, gid+*slot
	}
	gids = []int{gid}
	if u.Slots() > 0 {
		// Per-session agents get no supplementary groups (ApplySlot).
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

// Homes is the mapped users' home directories under one controller root.
type Homes struct {
	// Root is the controller root dir; homes live in <Root>/homes.
	Root string
	Log  *slog.Logger
}

// Dir is the home for pool's mapped user: one per pool, or one per slot
// with per_session ids so sessions do not share files.
func (h Homes) Dir(pool string, u *config.IsolationUser, slot *int) string {
	name := pool
	if u != nil && u.Slots() > 0 && slot != nil {
		name += "-" + strconv.Itoa(*slot)
	}
	return filepath.Join(h.Root, "homes", name)
}

// Prepare creates home, grants it to uid (and the first gid, so the
// controller's group can still clean up) and copies the agent's
// credentials files into it.
func (h Homes) Prepare(home string, uid int, gids []int, creds []string) error {
	if err := os.MkdirAll(home, 0o700); err != nil {
		return host.Permanentf("isolation home %s: %w", home, err)
	}
	if err := sandbox.RestrictDir(home, uid, gids[0], false); err != nil {
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
		if err := sandbox.CopyCredential(sandbox.ExpandHome(p), dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				h.Log.Warn("isolation: agent credentials file missing", "path", sandbox.ExpandHome(p))
				continue
			}
			return host.Permanentf("isolation credentials %s: %w", p, err)
		}
	}
	// Directories made above inherit the home's default ACL; files copied
	// in get it too. Redo the tree so the copy is readable by uid.
	return sandbox.RestrictDir(home, uid, gids[0], true)
}

// RemoveSlot deletes a closed per_session session's home (JOB-28), so
// the next session given the same slot starts empty and cannot read what
// this one left. The caller has checked the session is closed and its
// pool has per_session ids; a live host in sessionDir keeps the home.
// The controller can delete agent files through the pool group's default
// ACL entry; a failure is logged, and the next spawn in the slot still
// re-grants the dir.
func (h Homes) RemoveSlot(session, sessionDir, home string) {
	if HostAlive(sessionDir) {
		return
	}
	if err := os.RemoveAll(home); err != nil {
		h.Log.Warn("remove per-session home", "session", session, "home", home, "err", err)
	}
}

// CheckAccess checks every mount source and the cwd of spec are usable
// by the mapped uid, so a wrong mode is one clear permanent error.
// Unsandboxed, the agent opens its session dir and home through the 0700
// root, so it is first let through the root, sessions and homes dirs;
// sandboxed, bwrap binds them as the controller and no grant is needed
// (JOB-53).
func (h Homes) CheckAccess(spec host.Spec, uid int, gids []int) error {
	if spec.Sandbox == nil {
		for _, d := range []string{h.Root, filepath.Join(h.Root, "sessions"), filepath.Join(h.Root, "homes")} {
			if err := sandbox.GrantSearch(d, uid); err != nil {
				return host.Permanentf("isolation: let uid %d through %s: %w", uid, d, err)
			}
		}
	}
	return sandbox.CheckMounts(spec.Sandbox, spec.Dir, uid, gids)
}
