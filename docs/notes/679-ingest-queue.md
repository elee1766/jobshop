# Ingest: one queue, long-poll consumers, targeted wait wake (#679)

2026-11-05.

Every source now takes one path, and waits react to the event that
lands instead of a timer.

- **One append path.** Receivers (`ingest serve`, the controller's
  Linear listener, `ingest put`) append to the spool; pollers append to
  the inbox; the drain moves the spool into the inbox. Every new inbox
  item, from any source, goes through `inboxAppended`
  (`pkg/control/ingestdrain.go`). There are no source-specific wake
  hooks: the Linear listener's `OnLinear` is gone. The shared
  `ingest.Handler.OnSpool` runs after any new spool row, and the Linear
  consumer (`linearWebhookEvent`) now runs from the drained inbox item,
  with the identity taken from the item's endpoint.
- **Receivers notify.** The controller's listener wakes its drain
  directly. `ingest serve` calls the new admin method `ingest.notify`
  after each new row (coalesced, one call at a time). The 5s drain timer
  stays as the backstop when the controller was unreachable.
- **Long poll.** `ingest.next {after, timeout_ms}` blocks until the inbox
  changed (an append, or an admission for dispatch-only workers), up to
  25s. `jobshop ingest work` loops pass, `ingest.next`, pass; if the
  call fails (older controller, restart) it falls back to `--every`.
  The controller's admission pump waits on the same notifier, with its
  1s idle kept for leases and answered holds. The seq counts changes
  since controller start; a cursor from an older controller returns at
  once.
- **Targeted wait wake** (`pkg/control/waitwake.go`). An inbox item
  names subjects: its repo, repo#number, or a Linear key. GitHub
  `check_suite` and `check_run` items now carry the PR number from
  `pull_requests`. A wake evaluates only the waits with an unmet
  condition on those subjects (one PR read per PR, cached per pass).
  `github.checks` on a PR also matches repo-wide events, since status
  events name only the sha. An item with no subject wakes a full pass.
- **Internal events.** A run state change wakes waits on `run:<id>`.
  The pump sleeps until the earliest `time` condition or wait timeout,
  capped at the subscription interval. `deploy` is checked by the
  startup pass (a deploy restarts the controller) and the interval.
- **Unchanged.** At-least-once delivery, dedupe on delivery id and
  natural key, fenced leases. The interval full pass stays as reconcile.

Test: `TestWaitWebhookResume` (pkg/control) posts a signed
`check_suite` webhook while the wait and drain pumps run on a fake
clock that never moves; the `--resume` step is released to its session
within the 2s test bound. `TestWaitWakeTargeted`, `TestWaitNextTime`,
`TestNotifier`, `TestCheckEventSubject` cover the parts.
