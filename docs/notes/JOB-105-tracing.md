# Tracing: tick steps, outbound HTTP, subprocesses, store, reload (JOB-105)

2026-11-06.

Traces showed `supervisor tick` spans of 20 to 65s with no children, so the
slow step was invisible. As built:

- **Tick.** `supervisor tick` has one child per step (`tick sweep`,
  `tick run_triggers`, ... `tick compress_journals`, attribute
  `jobshop.tick.step`). Each step gets the step span's ctx instead of
  `c.ctx`, so spans below it nest under it. Per trigger, `trigger <name>`
  carries `jobshop.trigger.result` (fired, held, stall, error).
- **Outbound HTTP.** Every client is `telemetry.HTTPClient(kind, timeout)`:
  an otelhttp transport that makes a client span under the caller's span
  and records `jobshop_outbound_duration_seconds{kind,target}`. The span
  shows scheme, host and a route template set with `telemetry.WithRoute`,
  never the real path or query (webhook URLs and tokens live there). The
  `http.DefaultClient` calls (hooks, steps, providers, GitHub writes) now
  use shared traced clients with a ceiling timeout.
- **Subprocesses.** `exec <program>` spans from `runIsolated` (hooks,
  rules, steps, trigger checks, Desk exec) and `telemetry.ExecContext`
  (trigger commands, provider `sh`, `gh` in waits, subscribe and
  activity). Attributes: command name, a label (trigger, hook or the `gh`
  subcommand), exit code, `jobshop.exec.timed_out`. No script, args, env
  or stdin.
- **Store.** `store.Traced` records a span for a call that takes 50ms or
  more (`store <op>`, or `store <collection>.<op>`), under the span in its
  ctx. Faster calls cost two clock reads and no span.
- **Reload.** `config reload` (child of `rpc config.reload`) has
  `wait reloadMu`, `config load`, `set resources`, one
  `preflight` per distinct check (attribute `jobshop.pools`) and the
  `report_orphans` step. Reload runs under `context.WithoutCancel`, so a
  client that gave up does not cancel it half done.
- **Locks.** `jobshop_lock_wait_seconds{lock,for}` for `tickMu`, waiters
  `tick`, `webhook`, `event`, `check` (reload stopped taking it, JOB-104;
  `check` is a trigger check replay, JOB-109). A wait over 10ms is also a
  span `wait tickMu` under the waiter's span. It names the holder it
  waited behind (`jobshop.lock.for` = tick, webhook, event or check, and
  `jobshop.lock.holder.trigger` for an event or check job), the waiter
  (`jobshop.lock.waiter`), and links the holder's span, so a slow pass
  says what blocked it (JOB-109). With no holder at the time, only the
  waiter is set.
- **Metrics.** `jobshop_tick_step_duration_seconds{step}`,
  `jobshop_outbound_duration_seconds{kind,target}`,
  `jobshop_lock_wait_seconds{lock,for}`.

Sampling is unchanged: child spans follow the root's decision, and a
`context.Background()` caller (no recording span) records nothing.

Store spans and `StartSpan` need a tracer provider in ctx: they use the
provider of the parent span, so a package that never sees a tracer adds no
dependency.
