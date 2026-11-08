package control

import (
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/sessionrt"
	"github.com/elee1766/jobshop/pkg/store"

	host "github.com/elee1766/acphost/pkg/acphost"
)

// The mapped uid, home and ACL work lives in sessionrt (see
// sessionrt/home.go for why a mapped user gets a home of its own). The
// controller keeps what needs config and the durable session fields: which
// pool user applies, the slot, and whether the ticket is closed.

// mappedUser is the host uid and gids a session's agent runs as, or ok
// false when its pool has no isolation.user.
func (c *Controller) mappedUser(pool string, slot *int) (uid int, gids []int, ok bool, err error) {
	return sessionrt.MappedUser(c.config().IsolationFor(pool).User, slot)
}

// homes is the mapped users' homes under the controller root.
func (c *Controller) homes() sessionrt.Homes {
	return sessionrt.Homes{Root: c.root.Dir, Log: c.log}
}

// mappedHomeDir is the home for pool's mapped user: one per pool, or one
// per slot with per_session ids so sessions do not share files.
func (c *Controller) mappedHomeDir(pool string, slot *int) string {
	return c.mappedHomeDirIn(c.config(), pool, slot)
}

// mappedHomeDirIn is mappedHomeDir under cfg.
func (c *Controller) mappedHomeDirIn(cfg *config.Config, pool string, slot *int) string {
	return c.homes().Dir(pool, cfg.IsolationFor(pool).User, slot)
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
	if err := c.homes().Prepare(home, uid, gids, agent.Credentials); err != nil {
		return "", err
	}
	return home, nil
}

// removeSlotHome deletes a closed per_session session's home (JOB-28).
// Per-pool homes are shared by design and kept. The decision (closed
// session, per_session pool, a slot on the row) is the controller's; the
// deletion is sessionrt's.
func (c *Controller) removeSlotHome(cfg *config.Config, t store.Ticket) {
	u := cfg.IsolationFor(t.Pool).User
	if t.Kind != store.KindSession || t.Status != store.StatusClosed || u == nil || u.Slots() == 0 {
		return
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil || f.Slot == nil {
		return
	}
	c.homes().RemoveSlot(t.ID, c.root.SessionDir(t.ID), c.mappedHomeDirIn(cfg, t.Pool, f.Slot))
}

// checkMappedAccess checks every mount source and the cwd are usable by
// the session's mapped uid, so a wrong mode is one clear permanent error.
func (c *Controller) checkMappedAccess(spec host.Spec, slot *int, pool string) error {
	uid, gids, ok, err := c.mappedUser(pool, slot)
	if err != nil || !ok {
		return err
	}
	return c.homes().CheckAccess(spec, uid, gids)
}
