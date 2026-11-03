# Prompt rules

Add to the host `jobshop.yaml` on gasserver, then `jobshop reload`:

```yaml
prompt_rules:
  - name: no-master
    text: Never push to master; open a PR.
    at: system
  - name: dev-only
    text: Run go build, go vet and go test before pushing.
    match: {pools: [dev]}
```

```sh
JOBSHOP_SOCKET= jobshop config show | grep -A3 prompt_rules
JOBSHOP_SOCKET= jobshop session list            # pick a dev and a plan session
JOBSHOP_SOCKET= jobshop session prompt <dev-session>
JOBSHOP_SOCKET= jobshop session prompt <plan-session>
```

- The dev session shows `=== rule:dev-only ===` before its ticket. The
  plan session does not.
- `=== rule:no-master ===` shows only for a session whose host has not
  had a turn yet (for example right after `session restart`).
- After the next dev turn, `journalctl -u jobshop | grep "prompt rules
  applied"` lists `rules=dev-only` (and `no-master` on a host's first
  turn).
- Remove `dev-only`, run `jobshop reload`, and `session prompt` on the dev
  session no longer shows it.
