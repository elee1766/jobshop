# towa provider unreachable backs off (JOB-215)

2026-11-14. On a scratch root, not the live controller: the
`isolated-controller` skill's rules apply (root under /tmp, `env -u
JOBSHOP_SOCKET -u JOBSHOP_SESSION -u JOBSHOP_ROOT` on every command).
Needs `towa` on PATH and one account whose provider URL is not reachable
through the proxy below.

1. `jobshop init <scratch root>`, then in its `jobshop.yaml` define one
   pool on towa:

   ```yaml
   projects:
     js: {}
   agents:
     towa:
       command: towa
       args: [acp]    # use the invocation your towa build takes
   pools:
     dev:
       project: js
       agent: towa
       max_sessions: 1
       dir: <scratch root>
   ```

2. Start the controller with a dead proxy, so every provider request
   fails to connect:

   ```sh
   HTTPS_PROXY=http://127.0.0.1:1 jobshop serve --root <scratch root> --tick=1s 2>&1 | tee /tmp/towa-verify.log
   ```

3. In another shell, create one ticket for the pool:

   ```sh
   jobshop --root <scratch root> ticket create "towa unreachable probe" --pool dev
   ```

Pass, in the log after the session claims the ticket:

- `provider unreachable, backing off` with `attempt=1`, then 2 and on,
  each with a longer `wait`, up to 6 attempts. The `err` holds
  `"kind":"network"`.
- No `session interrupted, restarting` line and no `resolution=crashed`.
- `jobshop --root <scratch root> session list` shows `backoff_until` set;
  the ticket stays `in_progress` and no second session starts.

After the sixth attempt the turn fails as for any other prompt error:
that is the bound, not a failure of the check. Before this change the log
had no backoff line and the session ended `crashed` after three restarts.

Clean up: stop the controller, remove the scratch root.
