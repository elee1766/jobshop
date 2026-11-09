# Restart with live limited sessions (systemd cgroup driver, #944)

On a home-manager host with `isolation.enable` (cgroupDriver defaults to
systemd) and a pool with `isolation.limits`, on a build with #944 and
#992 (hosts in their own scopes).

1. `systemctl --user show jobshop -p Delegate` is `Delegate=no`.
   `jobshop config show` (or the generated jobshop.yaml) has
   `limits.cgroup_driver: systemd`.
2. Start a session in the limited pool. Then:

   ```sh
   s=<session id>
   iso=$(jq -c .cgroup ~/.local/state/jobshop/sessions/$s/isolation.json)
   echo "$iso"   # driver systemd, unit acphost-...scope, dir under jobshop-sessions.slice
   d=$(echo "$iso" | jq -r .dir)
   u=$(echo "$iso" | jq -r .unit)
   snap() {
     cat $d/memory.max $d/pids.max $d/cpu.max
     systemctl --user show "$u" -p MemoryMax -p TasksMax
   }
   snap | tee /tmp/before
   systemctl --user status "$u"
   ```

3. `systemctl --user restart jobshop`, then `systemctl --user is-active
   jobshop` is `active` (no `Result=resources` in `journalctl --user -u
   jobshop`).
4. `snap | diff /tmp/before -` prints nothing, and `jobshop session
   list` shows the session with the same host pid.
5. Restart a second time (step 3 again), then step 4 again (JOB-99: the
   EBUSY of JOB-97 showed on the restart after the first).
6. Kill the controller hard. `KillMode=process` leaves the session
   hosts, and `Restart=on-failure` brings the unit back after
   `RestartSec`:

   ```sh
   n=$(systemctl --user show jobshop -p NRestarts --value)
   systemctl --user kill --kill-whom=main -s KILL jobshop
   until systemctl --user is-active -q jobshop; do sleep 1; done
   systemctl --user show jobshop -p NRestarts --value   # n + 1
   ```

   The unit is `active` with no `Result=resources` in the journal, and
   step 4 holds again.
7. `cat /sys/fs/cgroup/.../jobshop.service/cgroup.subtree_control` is
   empty: the unit enables nothing.

The same check (two restarts) runs in `go test ./pkg/orchestration/control -run
TestSystemdScopeRestart` on a host with a user manager. The test stops
the controller in-process, so the kill -9 in step 6 is only checked
here.
