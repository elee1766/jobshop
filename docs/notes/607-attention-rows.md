# #607 Attention rows: summary and inline resolution

2026-11-05.

- Each attention item carries `summary` (rpc and Desk API): the
  ticket's `meta.desk.attention_summary` when set, else the first two
  non-empty lines of its body (question text, escalation reason,
  proposal), or for a wait past its timeout the wait reason. Clipped to
  240 bytes. The pre-started chat (#606) writes
  `meta.desk.attention_summary`; Desk's `recordChat` keeps it when it
  rewrites `meta.desk`. The list never reads journals.
- A row shows title, summary (2 lines), the kind's canned buttons
  (`ui/kit/attention.js`, shared with the chat panel) and Open chat. A
  canned button confirms, resolves, drops the row; no session opens.
  The row click still opens the chat.
- In the chat panel, a canned button that succeeds closes the panel. A
  failure keeps it open with the error toast.
- Mobile: the buttons wrap under the summary, 40px tall.
- Tests: `ui/e2e/attention-chat.spec.js`, `TestAttentionSummary`.
