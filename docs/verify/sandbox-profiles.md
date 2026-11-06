# Sandbox and isolation profiles on gasserver (JOB-83)

Needs a pool that is sandboxed and isolated inline today (the `isolated`
pool) and one open session of it.

1. Save the before: `jobshop session sandbox <id> > /tmp/before.txt`
   and `jobshop pool show isolated --sandbox > /tmp/show-before.txt`.
2. Move the pool's inline `sandbox:` and `isolation:` into
   `sandbox_profiles` (`host`, `egress-deny`) and `isolation_profiles`
   (`mapped`), and write the pool as `sandbox: {profiles: [host,
   egress-deny]}` and `isolation: {profiles: [mapped]}`.
3. `jobshop config check`: no error, and no "applies to nothing" warning.
   `jobshop reload`, then restart the session (`jobshop session restart
   <id> --now`) and save `jobshop session sandbox <new id> >
   /tmp/after.txt`.
4. `diff /tmp/before.txt /tmp/after.txt` shows only the session id (and
   the slot, if per-session ids are on): same mounts, env, deny list, user
   and cgroup, same bwrap argv.
5. `jobshop pool show isolated --sandbox` names `profile host (pool)` and
   the other profiles as the source of each entry.
6. A label rule: add `defaults.sandbox_by_label: [{labels:
   [instance:<name>], profiles: [host]}]`, `config check`, and
   `jobshop pool show <a pool of that instance> --sandbox` names
   `label rule 1`. A rule with a label no pool has warns.

Roll back: restore the inline config and reload.
