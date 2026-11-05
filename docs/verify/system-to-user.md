# Move a host from the NixOS module to home-manager

2026-11-05, JOB-22. One-time migration from `jobshop` (system units,
root self-deploy) to rootless user units under `jobcity`. See
[../home-manager.md](../home-manager.md).

1. **Host prerequisites.** In the host's NixOS config: the `jobcity`
   user with `linger = true`, its subuid/subgid range, nix-ld, group
   membership for shared sockets, and
   `systemd.tmpfiles.rules = [ "d /nix/var/nix/gcroots/per-user/jobcity 0755 jobcity users -" ]`.
   Keep `services.jobshop` on. Switch.
2. **Home config.** Add `homeConfigurations.jobcity` with the same
   `services.jobshop` settings and secret
   paths `jobcity` can read. Point `selfDeploy.hostFlake` at
   `<flake>#jobcity`. Do not switch it yet.
3. **Stop.** `sudo -u jobshop env JOBSHOP_ROOT=/var/lib/jobshop jobshop shutdown`,
   then `systemctl stop jobshop jobshop-self-deploy.timer 'jobshop-*'`.
   Check `systemctl list-units 'jobshop*'` shows nothing active.
4. **Copy.** `rsync -aHAX /var/lib/jobshop/ ~jobcity/.local/state/jobshop/`
   and `chown -R jobcity: ~jobcity/.local/state/jobshop`. Do the same for
   `/var/lib/jobshop-mise` and `/var/lib/jobshop-deploy` if used. Remove
   the copied `jobshop.yaml` and `jobshop.module.yaml` links: the home
   switch recreates them.
5. **Turn off the system module.** `services.jobshop.enable = false` in
   the host config. Switch.
6. **Start.** As jobcity: `home-manager switch --flake <flake>#jobcity`.
   `systemctl --user status jobshop` is active.
7. **Verify.**
   - `jobshop session list` shows the old sessions, adopted or reaped.
   - `jobshop run list` shows the runs from before the move.
   - A probe ticket in an isolated pool runs to close.
   - `systemctl --user start jobshop-self-deploy` logs `already at <sha>`
     or deploys, with no rollback.
   - Desk loads through the proxy.
8. **Clean up** after a week with no rollback: delete `/var/lib/jobshop`.

Rollback before step 8: stop the user units, set
`services.jobshop.enable = true` again, switch the host. The old data
dir is untouched.
