# Gasserver verification runbook

Written 2026-11-01 (#52), for the #36 box "controller restart, env refresh
and rate-limit backoff each verified on gasserver with no session ending
crashed", and for the 24h run. Behavior being checked is in
`docs/plan.md` sections 7.1 to 7.3.

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
- A Nix deploy (section 2) and editing `jobshop.yaml` (section 3, it is
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

## 0. Baseline

Before each scenario, with at least one session busy on a ticket:

```sh
jobshop session list                       # note ids, host_pid, current
JOBSHOP_SOCKET= deploy/verify-sessions.sh  # crashed or stuck in 24h fails
date -u +%FT%TZ > /tmp/verify-start        # start time for journalctl
```

## 1. Controller restart

```sh
sudo -n /run/current-system/sw/bin/systemctl restart jobshop.service
journalctl -u jobshop --since "$(cat /tmp/verify-start)" -o cat \
  | grep -E 'controller stopped|controller serving|resumed conversation|new\+summary|session ended'
jobshop session list
```

Pass:
- `controller stopped`, then `controller serving ... sessions=N` with N the
  number of open sessions from the baseline.
- `session list` shows the same session ids with the same `host_pid`
  (hosts were adopted, not respawned; `KillMode=process` keeps them alive).
- A session that was mid-turn finishes that turn (its ticket moves on).
- No `session ended` line with `resolution=crashed` or `resolution=stuck`.

## 2. Env refresh via deploy

Change something in the agent env through Nix (for example add a dummy
`services.jobshop.environment.VERIFY = "1";`), then deploy:

```sh
sudo nixos-rebuild switch --flake .#gasserver
journalctl -u jobshop --since "$(cat /tmp/verify-start)" -o cat \
  | grep -E 'refreshing host env|resumed conversation|new\+summary|session ended'
jobshop session list
```

Pass:
- One `refreshing host env host_pid=<old>` per session: idle sessions at
  once, busy ones after their current turn.
- Followed by `resumed conversation ... path=load` for the same session
  (or `started a new conversation ... path=new+summary` if the agent cannot
  load; both are fine, a crash is not).
- `session list` shows a new `host_pid` for each session, same ids,
  same `current` ticket.
- `tr '\0' '\n' < /proc/<new host_pid>/environ | grep VERIFY` shows the
  new value.
- No `resolution=crashed` or `resolution=stuck`.

`jobshop session restart <id>` exercises the same path on one session and
logs `restarting host on request`.

## 3. Forced rate-limit backoff

The controller treats a prompt error containing `429`, `rate_limit`,
`rate limit` or `is unavailable` as a rate limit. To force one without
burning the real account, give a scratch pool an agent whose provider URL
points at a stub that always answers 429:

```sh
python3 -c '
import http.server
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(429); self.send_header("content-type","application/json"); self.end_headers()
        self.wfile.write(b"{\"type\":\"error\",\"error\":{\"type\":\"rate_limit_error\",\"message\":\"429\"}}")
http.server.HTTPServer(("127.0.0.1", 8429), H).serve_forever()' &
```

In `jobshop.yaml`, add a pool `verify-rl` using the normal agent with
`env: {ANTHROPIC_BASE_URL: "http://127.0.0.1:8429"}` (use the variable your
agent's provider reads), then `sudo systemctl reload jobshop` and:

```sh
jobshop ticket create "rate limit probe" --pool verify-rl
journalctl -u jobshop -f -o cat | grep -E 'provider rate limited|session ended'
jobshop session list                       # backoff_until is set
```

Pass:
- `provider rate limited, backing off wait=1m0s`, then `2m0s`, `4m0s` ...
  capped at `15m0s`.
- `session list` shows `backoff_until` for the session; its ticket stays
  claimed; no second session starts for `verify-rl`.
- Kill the stub. On the next retry the prompt goes through or fails as a
  normal error, `backoff_until` clears.
- No `resolution=crashed` or `resolution=stuck`.

Clean up: close the probe ticket, remove the pool, reload.

## 4. Evidence

One command per check; paste its output on #36. A session that ends
`stuck` (agent stopped closing tickets, #62) counts as a failure just like
`crashed` (#71, 2026-11-01): both must be `[]` for the evidence window.

```sh
jobshop session list --state crashed --since 24h
jobshop session list --state stuck --since 24h
```

`--state` keeps closed sessions with that resolution, `--since` keeps
sessions closed within the duration; either one lists closed sessions only
and skips orphan hosts. `[]` is a pass. For the 24h run, run it at the end
of the window, together with:

```sh
jobshop session list --since 24h | jq -r '.[].ticket.resolution' | sort | uniq -c
# before #54 is deployed (no --state/--since):
jobshop session list --all | jq '[.[] | select(.ticket.resolution=="crashed" or .ticket.resolution=="stuck")]'
journalctl -u jobshop --since -24h -o cat | grep -c 'session ended'
```

## 5. Telemetry

Wired 2026-11-01 (gasserver.nix `modules/jobshop.nix`, gascity-city
`ops/observability/native`):

- Controller: `telemetry.metrics_listen = 127.0.0.1:9464`,
  `otlp_endpoint = http://127.0.0.1:14318` (the agent-run OTel collector,
  which forwards traces to VictoriaTraces and logs to VictoriaLogs).
- VictoriaMetrics (127.0.0.1:18428) scrapes `/metrics` as job `jobshop`.
- Grafana (127.0.0.1:3300) provisions the deployed
  `deploy/grafana/jobshop.json` from `/etc/jobshop/grafana`, folder
  `Jobshop`, uid `jobshop`, datasource pinned to `gc-vm`. A dashboard
  change ships with the next Jobshop deploy.

Checks, from a session (all endpoints are on loopback):

```sh
curl -s 127.0.0.1:9464/metrics | grep -c '^jobshop_'
curl -s 'http://127.0.0.1:18428/api/v1/targets' \
  | jq -r '.data.activeTargets[] | select(.labels.job=="jobshop") | .health'   # up
curl -s 'http://127.0.0.1:18428/api/v1/query?query=jobshop_sessions' | jq '.data.result | length'
curl -s 'http://127.0.0.1:20428/select/jaeger/api/services' | jq .data           # has "jobshop"
# crashed or stuck session ends in the last 24h, by pool; expect []
curl -s 'http://127.0.0.1:18428/api/v1/query' --data-urlencode \
  'query=sum by (jobshop_pool, jobshop_resolution) (increase(jobshop_sessions_ended_total{jobshop_resolution=~"crashed|stuck"}[24h])) > 0' \
  | jq .data.result
```

Counters such as `jobshop_session_restarts_total` and
`jobshop_lease_expiries_total` appear only after their first increment.
Grafana needs a login; panels are checked by an operator or with the
queries above.

### Script (#55)

`deploy/verify-telemetry.sh` runs the checks above and exits non-zero if
the target is down, `jobshop_sessions` or `jobshop_tickets_ready` is
missing for a pool, or VictoriaTraces has no `session` trace for service
`jobshop` in the last hour. Counters that have not incremented yet print
as `info` lines and do not fail. Defaults are the gasserver loopback
URLs; `--help` lists the flags.

```sh
deploy/verify-telemetry.sh                       # paste the output on #36
CURL=/run/current-system/sw/bin/curl deploy/verify-telemetry.sh   # if curl is not on PATH
```

Dashboard panels and the query each one needs data from (an operator with
a Grafana login can confirm visually; otherwise query VictoriaMetrics):

| Panel | Has data once |
| --- | --- |
| Sessions by pool and state, Sessions in backoff | supervisor ticks (always) |
| Tickets ready | supervisor ticks (always) |
| Claims and closes per hour | a ticket is claimed or closed |
| Restarts per hour | a session restarts |
| RPC p95 latency | any CLI call |
| Turns per hour by stop reason | the first agent turn after a deploy with #55 |

## 6. Refinery handoff dry run

Added 2026-11-01 (#35, acceptance item from #28). The pool moves of a
handed-off ticket show the route:

```sh
jobshop ticket show <id> | jq -r '.history[]|select(.diff.pool!=null)|"\(.at) \(.diff.pool) \(.actor)"'
```

- Merge path: one move to `refinery`, then closed `done` after the PR is
  merged and master CI is green.
- Route-back path: a review comment on the PR, left before the handoff,
  should give `refinery`, `dev`, `refinery`, then closed `done`.

The dry run fails if any of these happen:

- The PR is merged while a review comment is still unresolved.
- The ticket is closed without passing through `dev` on the route-back
  path, or closed `done` with the PR still open.
- The ticket moves to `refinery` while PR CI is red.

## 7. Closed-loop evidence

Added 2026-11-01 (#76) for the #36 item "Planner sees a closed loop". One
cycle: a trigger files a plan ticket, the planner files dev tickets while
holding it, one of them goes through the refinery and its PR is merged.
Needs a binary with `ticket trace` (#53).

Pick the plan ticket: the newest closed `replan` ticket whose dev work has
had time to merge.

```sh
JOBSHOP_SOCKET= jobshop ticket list --pool plan --status closed \
  | jq -r '.[] | select(.meta.trigger.name == "replan") | "\(.closed_at) \(.id) \(.title)"' \
  | sort -r | head
```

Run the check and paste its full output on #36, plus the trace if it fails:

```sh
JOBSHOP_SOCKET= deploy/verify-closed-loop.sh <plan-ticket-id>
JOBSHOP_SOCKET= jobshop ticket trace <plan-ticket-id>   # on failure
```

Pass: every line `ok`, ending `closed-loop check passed`. On failure the
`FAIL` line names the missing link: no trigger, plan ticket not closed, no
dev ticket from the planner session, none reached `refinery`, or no
merged PR linked. A dev ticket that is still in `dev` or `refinery` means
the cycle has not finished yet. Pick an older plan ticket or wait for it.
`--trace FILE` checks saved trace JSON offline.

## 8. Escalation drill

Added 2026-11-02 (#114) for the #82 boxes "Retry limits" and
"Escalation". File a throwaway issue, then start the drill against it:

```sh
JOBSHOP_SOCKET= jobshop run start drill --var issue=<n> --key drill-$(date +%F)
```

A dev session claims the step, comments and closes it `failed`. Check:

```sh
JOBSHOP_SOCKET= jobshop run show <run>          # state needs_human, run still open
JOBSHOP_SOCKET= jobshop ticket list --label needs-human \
  | jq -r '.[] | select(.title == "Escalate run <run>") | "\(.id) \(.pool)"'
gh issue view <n> --comments                    # step-history comment
```

Pass: the run is `needs_human`, exactly one "Escalate run <run>" ticket
is in `plan`, and the issue has the step-history comment. Post the run
id and links on #82, then close the escalation ticket and the throwaway
issue. The same key returns the existing run, so use a new date to
repeat.

## 9. CI repair drill

Added 2026-11-02 (#117) for the #82 box "CI repair". It uses a red
throwaway branch, so master stays green.

1. Push a branch `drill/red-ci` from master with one failing test (e.g.
   `func TestDrillRed(t *testing.T) { t.Fatal("ci drill") }` in
   `pkg/ids`). Do not open a PR.
2. Wait for its CI run to fail and take its SHA:

   ```sh
   gh run list --branch drill/red-ci --workflow ci --status completed --limit 1 \
     --json conclusion,headSha --jq '.[] | select(.conclusion == "failure") | .headSha'
   ```

3. Start the run by hand (or point a drill workspace's `ci_branch` at
   `drill/red-ci` and let `master-red` start it):

   ```sh
   JOBSHOP_SOCKET= jobshop run start ci-repair --var sha=<red sha> --key drill-ci-$(date +%F)
   ```

Check:

- The `fix` step is claimed by a dev session and comments on its ticket
  (`jobshop run show <run>`).
- At most 2 attempts on `fix` for that SHA: `failures` never exceeds 2
  and a repeated start with the same key returns the same run.
- The failing test is not on master, so the fix cannot land there. The
  run ends `needs_human` with one "Escalate run <run>" ticket in `plan`.

Post the run id and outcome on #82. Then close the escalation ticket,
close any PR the drill opened, and delete the branch:
`git push origin --delete drill/red-ci`.

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
