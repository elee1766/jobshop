# Rollback drill

Added 2026-11-02 (#155) for #145: the self-deploy timer rolls back when
the controller is unhealthy after a switch. Run on gasserver once the
#155 deploy is live (`jobshop version` at or after the merge commit) and
master has a newer green commit than the running one.

Set the drill (as root). The marker applies to the next deploy only;
the deploy unit deletes it before switching:

```sh
touch /var/lib/jobshop-deploy/drill
systemctl start jobshop-self-deploy.service   # or wait for the timer
```

Expect, in `journalctl -u jobshop-self-deploy -n 50`:

- `rollback drill: <rev> will be reported unhealthy`
- `controller unhealthy after <rev>, rolling back to /nix/store/...`
- the unit failed (exit 1).

Check:

```sh
cat /var/lib/jobshop-deploy/failed       # <rev>
test ! -e /var/lib/jobshop-deploy/drill  # marker cleared
jobshop version                          # the commit from before the drill
systemctl start jobshop-self-deploy.service
journalctl -u jobshop-self-deploy -n 5   # "<rev> failed before, waiting ..."
```

Clear: delete `drill` if the deploy never ran. To deploy the drilled
rev without waiting for a newer green commit, delete
`/var/lib/jobshop-deploy/failed` and start the unit again.

Evidence: the journal lines and `failed` contents, posted on #145.
