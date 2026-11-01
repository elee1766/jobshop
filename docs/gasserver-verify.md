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
jobshop session list --state crashed --since 24h   # expect []
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
- No `session ended` line with `resolution=crashed`.

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
- No `resolution=crashed`.

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
- No `resolution=crashed`.

Clean up: close the probe ticket, remove the pool, reload.

## 4. Evidence

One command per check; paste its output on #36:

```sh
jobshop session list --state crashed --since 24h
```

`--state` keeps closed sessions with that resolution, `--since` keeps
sessions closed within the duration; either one lists closed sessions only
and skips orphan hosts. `[]` is a pass. For the 24h run, run it at the end
of the window, together with:

```sh
jobshop session list --since 24h | jq -r '.[].ticket.resolution' | sort | uniq -c
# before #54 is deployed (no --state/--since):
jobshop session list --all | jq '[.[] | select(.ticket.resolution=="crashed")]'
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
