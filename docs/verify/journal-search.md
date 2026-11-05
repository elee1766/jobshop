# Journal search on gasserver (#530)

2026-11-05. After deploying with `services.jobshop.journal.victorialogs`
enabled (on a port other than Gas City's VictoriaLogs).

1. `systemctl status victorialogs` is active and `jobshop config show`
   has `journal.victorialogs.url`.
2. `jobshop journal backfill` returns `{"sessions": N, "frames": M}` with
   M > 0. A second run returns zeros or only new frames.
3. `jobshop journal search --since 24h '* | stats count() n'` is close to
   the frame count of the last day's journals.
4. Pick an open session: `jobshop journal search 'session:<id> | stats max(seq) m'`
   is within a few frames of `jobshop session journal <id> --since 0 | tail -1`.
5. Failed tools:
   `jobshop journal search --since 24h 'status:failed | uniq by (session, seq, tool) | stats by (tool) count() n | sort by (n desc) | limit 5'`
   lists tool titles.
6. From a session in a pool with the `journal-search` grant,
   `jobshop journal search 'project:<another project>'` prints nothing.
7. `curl -s 127.0.0.1:9464/metrics | grep jobshop_journal_ship_lag_seconds`
   shows 0 or a few seconds per pool; `jobshop attention` has no
   `journal-lag` item.
8. Outage: `systemctl stop victorialogs` for 1m, then start it. The lag
   rises, `jobshop_journal_ship_errors_total` grows, and both return to
   normal within a minute; step 4 still matches.
