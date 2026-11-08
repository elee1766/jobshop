package control

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/sessionrt"
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
	return c.mappedHomeDirIn(c.config(), pool, slot)
}

// mappedHomeDirIn is mappedHomeDir under cfg.
func (c *Controller) mappedHomeDirIn(cfg *config.Config, pool string, slot *int) string {
	name := pool
	if u := cfg.IsolationFor(pool).User; u != nil && u.Slots() > 0 && slot != nil {
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
	if err := sessionrt.RestrictDir(home, uid, gids[0], false); err != nil {
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
		if err := sessionrt.CopyCredential(expandHome(p), dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				c.log.Warn("isolation: agent credentials file missing", "path", expandHome(p))
				continue
			}
			return host.Permanentf("isolation credentials %s: %w", p, err)
		}
	}
	// Directories made above inherit the home's default ACL; files copied
	// in get it too. Redo the tree so the copy is readable by uid.
	return sessionrt.RestrictDir(home, uid, gids[0], true)
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
func (c *Controller) removeSlotHome(cfg *config.Config, t store.Ticket) {
	u := cfg.IsolationFor(t.Pool).User
	if t.Kind != store.KindSession || t.Status != store.StatusClosed || u == nil || u.Slots() == 0 {
		return
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil || f.Slot == nil {
		return
	}
	if sessionrt.HostAlive(c.root.SessionDir(t.ID)) {
		return
	}
	home := c.mappedHomeDirIn(cfg, t.Pool, f.Slot)
	if err := os.RemoveAll(home); err != nil {
		c.log.Warn("remove per-session home", "session", t.ID, "home", home, "err", err)
	}
}

// checkMappedAccess checks every mount source and the cwd are usable by
// the session's mapped uid, so a wrong mode is one clear permanent
// error. Unsandboxed, the agent opens its session dir and home through
// the 0700 root, so it is first let through the root, sessions and
// homes dirs; sandboxed, bwrap binds them as the controller and no
// grant is needed (JOB-53).
func (c *Controller) checkMappedAccess(spec host.Spec, slot *int, pool string) error {
	uid, gids, ok, err := c.mappedUser(pool, slot)
	if err != nil || !ok {
		return err
	}
	if spec.Sandbox == nil {
		for _, d := range []string{c.root.Dir, filepath.Join(c.root.Dir, "sessions"), filepath.Join(c.root.Dir, "homes")} {
			if err := sessionrt.GrantSearch(d, uid); err != nil {
				return host.Permanentf("isolation: let uid %d through %s: %w", uid, d, err)
			}
		}
	}
	return sessionrt.CheckMounts(spec.Sandbox, spec.Dir, uid, gids)
}
