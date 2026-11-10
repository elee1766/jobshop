package session

import (
	"errors"
	"io/fs"
	"log/slog"

	host "github.com/gfx-labs/acphost/pkg/acphost"
	"github.com/gfx-labs/acphost/pkg/proc"
)

// Placement is where a new session host starts. With the systemd driver
// (#992) that is a transient scope in the sessions slice
// (acphost.InScope): the manager owns it, so the host and an agent
// without a pool scope are outside the controller's unit and a restart
// finds the unit empty. With cgroupfs it is the hosts cgroup
// (acphost.InCgroup), or nothing: the host stays in the controller's.
type Placement struct {
	Systemd bool   // start in a scope in Slice
	Slice   string // the sessions slice, with systemd
	Cgroup  string // the hosts cgroup dir, with cgroupfs
}

// StartDetached starts the session host for dir as this binary re-exec'd
// in its own session (acphost.StartDetached, #904), with env as its
// environment. It returns once the agent started; a setup failure is a
// *host.SpawnError, and host.Permanent sees through it.
//
// The host goes in the sessions slice or cgroup, so its load counts
// against the sessions cap and not the controller's share (JOB-91), and
// with systemd outside the controller's unit (#992). A placement that
// cannot be used leaves the host in the controller's cgroup, as before,
// and logs it.
func StartDetached(id, dir string, env []string, p Placement, log *slog.Logger) (*proc.Detached, error) {
	switch {
	case p.Systemd:
		d, err := host.StartDetached(dir, env, host.InScope(p.Slice))
		var se *host.SpawnError
		if err == nil || errors.As(err, &se) {
			// A SpawnError is the agent's setup, not the scope: a start
			// without the scope would fail the same way.
			return d, err
		}
		// The scope could not be started (no manager, or this process is
		// outside its manager's tree).
		log.Warn("session host stays in the controller cgroup", "session", id, "slice", p.Slice, "err", err)
	case p.Cgroup != "":
		d, err := host.StartDetached(dir, env, host.InCgroup(p.Cgroup))
		var pe *fs.PathError
		if err == nil || !errors.As(err, &pe) || pe.Path != p.Cgroup {
			return d, err
		}
		// InCgroup fails the start when the cgroup cannot be opened.
		log.Warn("session host stays in the controller cgroup", "session", id, "err", err)
	}
	return host.StartDetached(dir, env)
}
