# Telemetry

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
