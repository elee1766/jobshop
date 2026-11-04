# Ingest end-to-end tests (#453, 2026-11-04)

Part of #318. `pkg/control/ingest_e2e_internal_test.go` drives one
GitHub issue comment through the whole pipeline: the `ingest.Handler`
(signed request, 202), the controller's spool, `drainOnce`, the inbox,
admission and dispatch, ending at the interrupt message the session's
runner sends as its prompt (`runner.deliver`).

- A GitHub webhook row has no consumer of its own (#442): the poller
  row for the same comment is the one delivered. Each test therefore
  also runs a subscription poll (fake thread source) for the comment.
- A restart is a second controller on the same root dir, DB, spool and
  fake clock, opened after the first is closed without draining.
- Cases: restart before and after drain, webhook and poll in either
  order, mutating redaction, validating deny and ask_human, hook
  timeout under both failure policies, url: hook signature check, and
  the receiver refusing an unsigned delivery.
- Hook timeouts use the hook's own context (50ms), not the fake clock:
  `callAdmission` times out with `context.WithTimeout`, which the
  controller clock does not drive.
