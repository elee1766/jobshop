# Controller restart

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
