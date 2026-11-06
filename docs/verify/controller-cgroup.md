# Controller CPU reserve (JOB-91)

On gasserver with `isolation.enable` and a build of at least the release
with JOB-91.

1. After the switch, `systemctl status jobshop` shows `controller` and
   `sessions/hosts` under its cgroup. Journal: `isolation: sessions cgroup
   ready`, and no `tuning not applied` warning.
2. With the unit's cgroup dir `C` (`/sys/fs/cgroup/.../jobshop.service`):
   `cat $C/controller/cpu.weight $C/sessions/cpu.weight $C/sessions/cpu.max`
   reads `1000`, `100` and `500000 100000` on the default quota.
3. `$C/sessions/hosts/cgroup.procs` lists the live session hosts, and
   `$C/controller/cgroup.procs` only the controller (plus short-lived
   hooks).
4. Start a session that runs `yes > /dev/null` six times. `top` shows the
   slice at five CPUs, and `jobshop ingest` or an admission hook still
   answers in well under its timeout.
5. Change `sessionsCpu` in the host config, switch (reload only): step 2
   shows the new cap without restarting sessions.
