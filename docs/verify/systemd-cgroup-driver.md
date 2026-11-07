# Restart with live limited sessions (systemd cgroup driver, #944)

On a home-manager host with `isolation.enable` (cgroupDriver defaults to
systemd) and a pool with `isolation.limits`, on a build with #944.

1. `systemctl --user show jobshop -p Delegate` is `Delegate=no`.
   `jobshop config show` (or the generated jobshop.yaml) has
   `limits.cgroup_driver: systemd`.
2. Start a session in the limited pool. Then:

   ```sh
   s=<session id>
   iso=$(jq -c .cgroup ~/.local/state/jobshop/sessions/$s/isolation.json)
   echo "$iso"   # driver systemd, unit acphost-...scope, dir under jobshop-sessions.slice
   d=$(echo "$iso" | jq -r .dir)
   cat $d/memory.max $d/pids.max $d/cpu.max | tee /tmp/before
   systemctl --user status "$(echo "$iso" | jq -r .unit)"
   ```

3. `systemctl --user restart jobshop`, then `systemctl --user is-active
   jobshop` is `active` (no `Result=resources` in `journalctl --user -u
   jobshop`).
4. `cat $d/memory.max $d/pids.max $d/cpu.max | diff /tmp/before -` prints
   nothing, and `jobshop session list` shows the session with the same
   host pid.
5. `cat /sys/fs/cgroup/.../jobshop.service/cgroup.subtree_control` is
   empty: the unit enables nothing.

The same check runs in `go test ./pkg/control -run TestSystemdScopeRestart`
on a host with a user manager.
