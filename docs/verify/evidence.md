# Evidence

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
