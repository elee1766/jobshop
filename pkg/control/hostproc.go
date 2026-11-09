package control

import (
	"github.com/elee1766/acphost/pkg/proc"

	"github.com/elee1766/jobshop/pkg/execution/session"
)

// stopHost asks a host to stop (it forwards SIGTERM to the agent's group),
// waits for it on its lock, and kills it if it is still running after
// session.HostStopWait. Then it stops what is left of the session's
// processes. In-process hosts (Options.Hosts, tests) stop through their
// runner.
func (c *Controller) stopHost(dir string) {
	if c.opt.Hosts != nil && c.opt.Hosts.Stop(dir) {
		return
	}
	if session.StopDetached(dir, session.HostStopWait) {
		session.StopProcs(dir)
	}
}

// startDetachedHost starts the session host for dir where the config
// places it (JOB-91, #992), see session.StartDetached.
func (c *Controller) startDetachedHost(id, dir string) (*proc.Detached, error) {
	return session.StartDetached(id, dir, hostInheritedEnv(), c.hostPlacement(), c.log)
}
