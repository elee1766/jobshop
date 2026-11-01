# Gasserver verification runbook

Written 2026-11-01 (#52), for the #36 box "controller restart, env refresh
and rate-limit backoff each verified on gasserver with no session ending
crashed", and for the 24h run. Behavior being checked is in
`docs/plan.md` sections 7.1 to 7.3.

Run as the jobshop user so the CLI finds the admin socket:

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
sudo systemctl restart jobshop
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
journalctl -u jobshop --since -24h -o cat | grep -c 'session ended'
```
