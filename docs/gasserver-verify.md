# Gasserver verification runbook

Written 2026-11-01 (#52), for the #36 box "controller restart, env refresh
and rate-limit backoff each verified on gasserver with no session ending
crashed", and for the 24h run. Behavior being checked is in
`docs/notes/session-files.md`, `11-env-refresh.md` and
`10-rate-limits-drain-quarantine.md`.

Agent sessions already run as `jobshop` (granted 2026-11-01 in
gasserver.nix `modules/jobshop.nix`):

- `journalctl -u jobshop` works (group `systemd-journal`).
- `sudo -n /run/current-system/sw/bin/systemctl restart jobshop.service`
  and `... reload jobshop.service` work without a password. Nothing else
  runs under sudo. A restart cannot kill the session running it: hosts
  survive (`KillMode=process`), only CLI calls fail briefly.
- The admin socket is not reachable from a session socket. Use
  `JOBSHOP_SOCKET= jobshop ...` (empty) so the CLI falls back to
  `$JOBSHOP_ROOT/admin.sock`, which `jobshop` owns.
- A Nix deploy (`verify/env-refresh.md`) and editing `jobshop.yaml` (`verify/rate-limit-backoff.md`, it is
  a Nix store link) need the operator. Ask on the issue with the exact
  change; `jobshop session restart <id>` covers the env refresh path
  meanwhile.
- A host started before a grant keeps its old PATH and groups (found
  2026-11-01, #56). Then `sudo` and `curl` are not found and the journal
  is unreadable. Use full paths: `/run/wrappers/bin/sudo -n ...`,
  `/run/current-system/sw/bin/curl`, and
  `/run/wrappers/bin/sg systemd-journal -c "journalctl -u jobshop ..."`.
  `jobshop session restart <id>` on that session also picks up the new env.

An operator shell:

```sh
sudo -u jobshop -i          # HOME and JOBSHOP_ROOT are /var/lib/jobshop
export JOBSHOP_ROOT=/var/lib/jobshop
```

Journal lines below are from `journalctl -u jobshop`. The controller logs
with slog, so match on the `msg` text.

## Checks

One file per check in `docs/verify/<slug>.md` (2026-11-02, #131). The
directory listing is the index: a new drill is a new file, so parallel
PRs do not collide on a section number. Do not add sections here. The
checks moved from the old numbered sections, in their old order:

- [Baseline](verify/baseline.md)
- [Controller restart](verify/controller-restart.md)
- [Env refresh via deploy](verify/env-refresh.md)
- [Forced rate-limit backoff](verify/rate-limit-backoff.md)
- [Evidence](verify/evidence.md)
- [Telemetry](verify/telemetry.md)
- [Refinery handoff dry run](verify/refinery-dry-run.md)
- [Closed-loop evidence](verify/closed-loop.md)
- [Escalation drill](verify/escalation-drill.md)
- [CI repair drill](verify/ci-repair-drill.md)
- [Stale-blocker drill](verify/stale-blocker-drill.md)

## Self-deploy

Promotion is unattended (`services.jobshop.selfDeploy`, `docs/plan.md`
section 6). Every 10 minutes `jobshop-self-deploy` picks the newest master
commit with green CI and rebuilds with that jobshop rev. Check it with:

```sh
journalctl -u jobshop-self-deploy -o cat --since -1h   # "deployed <rev>" or "already at"
cat /var/lib/jobshop-deploy/{deployed,failed} 2>/dev/null
```

A rev that failed to build or left the controller unhealthy is recorded in
`failed` and skipped until a newer green commit lands; the host was switched
back to the previous generation.
