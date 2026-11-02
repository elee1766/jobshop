# Env refresh via deploy

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
