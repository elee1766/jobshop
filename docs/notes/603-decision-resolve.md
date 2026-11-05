# Decisions resolve the proposal issue; chat delivery wakes the runner (#603)

2026-11-05. Desk's decision buttons relabeled or closed the Jobshop ticket,
not the GitHub proposal issue, approval did not wake the planner (parked
with `release --until 24h`), and chat messages waited for a runner poll.

- `decision.resolve {id, action, note}` (CLI `jobshop decision
  approve|reject|dismiss <id> [-n note]`): one implementation for Desk,
  gateways and sessions. The ticket must be labeled `proposed`; its
  `meta.github` names the issue. GitHub calls use the identity covering
  the repo, else the host's gh token.
  - approve: add `target`, remove `proposed`, comment who approved; close
    the ticket `approved`.
  - reject: comment, close the issue as not planned; ticket `rejected`.
  - dismiss: ticket `dismissed`, issue untouched. It leaves attention
    (closed). The planner prompt treats it as "not now" for 7 days.
  - Each clears `defer_until` on open tasks and steps in the project
    that carry the same `meta.github` issue or name it (`owner/repo#n`
    or `#n`, not `#n0`) in title or body, and wakes the wait pump. A
    second call on a closed ticket fails, so a double click acts once.
- `ticket.undefer` (`jobshop ticket undefer <id>`): admin, or a session
  whose pool has the `ticket-undefer` grant, in its own project.
  `decision-resolve` is grantable too; the operator pool (Desk's
  attention chats) has both.
- Wait type `github.label` (`issue=o/r#n,label=target`): met when the
  issue has the label or is closed. The planner parks a proposal ticket
  it holds on it instead of `release --until 24h`, and otherwise closes
  its planning ticket after filing.
- Desk: Approve as target, Reject, Dismiss call `decision.resolve`;
  errors show as a toast. Why the old buttons did nothing: the shell's
  confirm dialog cleared its state before calling resolve, and the
  `{@const}` it read from was already null, so OK threw and every
  confirmed action in Desk was dropped. It now takes resolve first.
- Chat latency: the measured 4s from message to prompt was the runner's
  idle poll (`Poll`, 5s). Creating a message for a session now wakes its
  runner (session send, controller messages, question answers,
  subscription news). Desk shows a sent message at once with its state
  ("sent, waiting for the session" or "queued for after this turn")
  until the turn that delivers it starts. Tool rows and text already
  stream over `session.subscribe`; the 16s was the agent's turn.
- Tests: `TestDecisionApprove`, `TestDecisionReject`,
  `TestDecisionDismiss`, `TestDecisionGrant`, `TestWaitLabel` (GitHub
  responses recorded from acphost#56 in
  `pkg/control/testdata/github_decision.json`); Playwright "a decision
  item" and the pending check in "composer". Live check:
  `docs/verify/desk-chat-latency.md`.
